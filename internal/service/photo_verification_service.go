package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type UploadPhotoInput struct {
	BathhouseID  uuid.UUID
	URL          string
	ThumbnailURL string
}

type PhotoVerificationService interface {
	UploadPhoto(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, input UploadPhotoInput) (*domain.BathhousePhoto, error)
	DeletePhoto(ctx context.Context, photoID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error
	ReorderPhotos(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole, photoIDs []uuid.UUID) error
	VerifyPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID) (*domain.BathhousePhoto, error)
	RejectPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID, reason string) (*domain.BathhousePhoto, error)
	GetPendingPhotos(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	ListVerifiedByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
}

type photoVerificationService struct {
	photoRepo   repository.BathhousePhotoRepository
	bhRepo      repository.BathhouseRepository
	accessCheck *AccessChecker
}

func NewPhotoVerificationService(
	photoRepo repository.BathhousePhotoRepository,
	bhRepo repository.BathhouseRepository,
	accessCheck *AccessChecker,
) PhotoVerificationService {
	return &photoVerificationService{
		photoRepo:   photoRepo,
		bhRepo:      bhRepo,
		accessCheck: accessCheck,
	}
}

const maxPhotosPerBathhouse = 30

func (s *photoVerificationService) UploadPhoto(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, input UploadPhotoInput) (*domain.BathhousePhoto, error) {
	if err := s.accessCheck.CanManageBathhouse(ctx, userID, userRole, input.BathhouseID); err != nil {
		return nil, err
	}

	existing, err := s.photoRepo.ListByBathhouse(ctx, input.BathhouseID)
	if err != nil {
		return nil, err
	}

	if len(existing) >= maxPhotosPerBathhouse {
		return nil, domain.ErrInvalidInput
	}

	// Calculate next position as max(existing positions) + 1 to avoid collisions after deletions.
	nextPosition := 0
	for _, p := range existing {
		if p.Position >= nextPosition {
			nextPosition = p.Position + 1
		}
	}

	photo := &domain.BathhousePhoto{
		ID:           uuid.New(),
		BathhouseID:  input.BathhouseID,
		URL:          input.URL,
		ThumbnailURL: input.ThumbnailURL,
		Position:     nextPosition,
		Status:       domain.PhotoStatusPending,
	}

	if err := photo.Validate(); err != nil {
		return nil, err
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, err
	}

	// New photo resets verification badge
	if err := s.bhRepo.UpdatePhotoVerified(ctx, input.BathhouseID, false); err != nil {
		return nil, fmt.Errorf("reset photo verification: %w", err)
	}

	return photo, nil
}

func (s *photoVerificationService) DeletePhoto(ctx context.Context, photoID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error {
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return err
	}

	// Admin can always delete, owner/representative can manage their bathhouse
	if userRole != domain.RoleAdmin {
		if err := s.accessCheck.CanManageBathhouse(ctx, userID, userRole, photo.BathhouseID); err != nil {
			return err
		}
	}

	if err := s.photoRepo.Delete(ctx, photoID); err != nil {
		return err
	}

	if err := s.recalcVerificationStatus(ctx, photo.BathhouseID); err != nil {
		return fmt.Errorf("recalculate verification status: %w", err)
	}

	return nil
}

func (s *photoVerificationService) ReorderPhotos(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole, photoIDs []uuid.UUID) error {
	if err := s.accessCheck.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}

	if len(photoIDs) == 0 {
		return domain.ErrInvalidInput
	}

	// Validate that the provided photo IDs are an exact permutation of the bathhouse's photos.
	existing, err := s.photoRepo.ListByBathhouse(ctx, bathhouseID)
	if err != nil {
		return err
	}
	if len(photoIDs) != len(existing) {
		return domain.ErrInvalidInput
	}
	existingSet := make(map[uuid.UUID]struct{}, len(existing))
	for _, p := range existing {
		existingSet[p.ID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(photoIDs))
	for _, id := range photoIDs {
		if _, ok := existingSet[id]; !ok {
			return domain.ErrInvalidInput
		}
		if _, dup := seen[id]; dup {
			return domain.ErrInvalidInput
		}
		seen[id] = struct{}{}
	}

	return s.photoRepo.Reorder(ctx, bathhouseID, photoIDs)
}

func (s *photoVerificationService) VerifyPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID) (*domain.BathhousePhoto, error) {
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return nil, err
	}

	if photo.Status != domain.PhotoStatusPending {
		return nil, domain.ErrInvalidInput
	}

	if err := s.photoRepo.UpdateStatus(ctx, photoID, domain.PhotoStatusVerified, &adminID, ""); err != nil {
		return nil, err
	}

	if err := s.recalcVerificationStatus(ctx, photo.BathhouseID); err != nil {
		return nil, fmt.Errorf("recalculate verification status: %w", err)
	}

	return s.photoRepo.GetByID(ctx, photoID)
}

func (s *photoVerificationService) RejectPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID, reason string) (*domain.BathhousePhoto, error) {
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return nil, err
	}

	if photo.Status != domain.PhotoStatusPending {
		return nil, domain.ErrInvalidInput
	}

	if reason == "" || len(reason) > 1000 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.photoRepo.UpdateStatus(ctx, photoID, domain.PhotoStatusRejected, &adminID, reason); err != nil {
		return nil, err
	}

	if err := s.recalcVerificationStatus(ctx, photo.BathhouseID); err != nil {
		return nil, fmt.Errorf("recalculate verification status: %w", err)
	}

	return s.photoRepo.GetByID(ctx, photoID)
}

func (s *photoVerificationService) GetPendingPhotos(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
	return s.photoRepo.ListPending(ctx, page, pageSize)
}

func (s *photoVerificationService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	return s.photoRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *photoVerificationService) ListVerifiedByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	return s.photoRepo.ListVerifiedByBathhouse(ctx, bathhouseID)
}

func (s *photoVerificationService) recalcVerificationStatus(ctx context.Context, bathhouseID uuid.UUID) error {
	photos, err := s.photoRepo.ListByBathhouse(ctx, bathhouseID)
	if err != nil {
		return fmt.Errorf("list photos for verification check: %w", err)
	}

	// Only consider non-rejected photos; a bathhouse is verified when it has
	// at least one photo and all non-rejected photos are verified (no pending).
	var nonRejected int
	allVerified := true
	for _, p := range photos {
		if p.Status == domain.PhotoStatusRejected {
			continue
		}
		nonRejected++
		if p.Status != domain.PhotoStatusVerified {
			allVerified = false
			break
		}
	}

	verified := nonRejected > 0 && allVerified
	if err := s.bhRepo.UpdatePhotoVerified(ctx, bathhouseID, verified); err != nil {
		return fmt.Errorf("update photo verification flag: %w", err)
	}
	return nil
}
