package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
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
}

type photoVerificationService struct {
	photoRepo    repository.BathhousePhotoRepository
	bhRepo       repository.BathhouseRepository
	accessCheck  *AccessChecker
	logger       *logger.Logger
}

func NewPhotoVerificationService(
	photoRepo repository.BathhousePhotoRepository,
	bhRepo repository.BathhouseRepository,
	accessCheck *AccessChecker,
	log *logger.Logger,
) PhotoVerificationService {
	return &photoVerificationService{
		photoRepo:   photoRepo,
		bhRepo:      bhRepo,
		accessCheck: accessCheck,
		logger:      log,
	}
}

func (s *photoVerificationService) UploadPhoto(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, input UploadPhotoInput) (*domain.BathhousePhoto, error) {
	if err := s.accessCheck.CanManageBathhouse(ctx, userID, userRole, input.BathhouseID); err != nil {
		return nil, err
	}

	existing, err := s.photoRepo.ListByBathhouse(ctx, input.BathhouseID)
	if err != nil {
		return nil, err
	}

	photo := &domain.BathhousePhoto{
		ID:           uuid.New(),
		BathhouseID:  input.BathhouseID,
		URL:          input.URL,
		ThumbnailURL: input.ThumbnailURL,
		Position:     len(existing),
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
		s.logger.Error("failed to reset photo verification", "bathhouse_id", input.BathhouseID, "error", err)
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

	// Recalculate verification status
	s.recalcVerificationStatus(ctx, photo.BathhouseID)

	return nil
}

func (s *photoVerificationService) ReorderPhotos(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole, photoIDs []uuid.UUID) error {
	if err := s.accessCheck.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
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

	// Check if all photos of this bathhouse are verified
	s.recalcVerificationStatus(ctx, photo.BathhouseID)

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

	if err := s.photoRepo.UpdateStatus(ctx, photoID, domain.PhotoStatusRejected, &adminID, reason); err != nil {
		return nil, err
	}

	return s.photoRepo.GetByID(ctx, photoID)
}

func (s *photoVerificationService) GetPendingPhotos(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
	return s.photoRepo.ListPending(ctx, page, pageSize)
}

func (s *photoVerificationService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	return s.photoRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *photoVerificationService) recalcVerificationStatus(ctx context.Context, bathhouseID uuid.UUID) {
	photos, err := s.photoRepo.ListByBathhouse(ctx, bathhouseID)
	if err != nil {
		s.logger.Error("failed to list photos for verification check", "bathhouse_id", bathhouseID, "error", err)
		return
	}

	if len(photos) == 0 {
		if err := s.bhRepo.UpdatePhotoVerified(ctx, bathhouseID, false); err != nil {
			s.logger.Error("failed to update photo verification", "bathhouse_id", bathhouseID, "error", err)
		}
		return
	}

	allVerified := true
	for _, p := range photos {
		if p.Status != domain.PhotoStatusVerified {
			allVerified = false
			break
		}
	}

	if err := s.bhRepo.UpdatePhotoVerified(ctx, bathhouseID, allVerified); err != nil {
		s.logger.Error("failed to update photo verification", "bathhouse_id", bathhouseID, "error", err)
	}
}
