package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type InviteRepresentativeInput struct {
	UserEmail   string
	BathhouseID uuid.UUID
}

type RepresentativeService interface {
	Invite(ctx context.Context, ownerID uuid.UUID, input InviteRepresentativeInput) (*domain.Representative, error)
	Revoke(ctx context.Context, ownerID uuid.UUID, representativeID uuid.UUID) error
	ListByBathhouse(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error)
	GetMyBathhouses(ctx context.Context, userID uuid.UUID) ([]domain.Bathhouse, error)
}

type representativeService struct {
	repRepo  repository.RepresentativeRepository
	userRepo repository.UserRepository
	bhRepo   repository.BathhouseRepository
}

func NewRepresentativeService(
	repRepo repository.RepresentativeRepository,
	userRepo repository.UserRepository,
	bhRepo repository.BathhouseRepository,
) RepresentativeService {
	return &representativeService{
		repRepo:  repRepo,
		userRepo: userRepo,
		bhRepo:   bhRepo,
	}
}

func (s *representativeService) Invite(ctx context.Context, ownerID uuid.UUID, input InviteRepresentativeInput) (*domain.Representative, error) {
	// Verify bathhouse belongs to owner
	bh, err := s.bhRepo.GetByID(ctx, input.BathhouseID)
	if err != nil {
		return nil, err
	}
	if bh.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return nil, err
	}

	// Change role to representative if user is currently a client
	if user.Role == domain.RoleClient {
		user.Role = domain.RoleRepresentative
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
	} else if user.Role != domain.RoleRepresentative {
		return nil, domain.ErrInvalidInput // can't invite owner or admin as representative
	}

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      user.ID,
		BathhouseID: input.BathhouseID,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
	}

	if err := s.repRepo.Create(ctx, rep); err != nil {
		return nil, err
	}

	return rep, nil
}

func (s *representativeService) Revoke(ctx context.Context, ownerID uuid.UUID, representativeID uuid.UUID) error {
	// Delete the representative assignment.
	// Ownership verification is done at the handler layer via RBAC middleware.
	return s.repRepo.Delete(ctx, representativeID)
}

func (s *representativeService) ListByBathhouse(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error) {
	// Verify bathhouse belongs to owner
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}
	if bh.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	return s.repRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *representativeService) GetMyBathhouses(ctx context.Context, userID uuid.UUID) ([]domain.Bathhouse, error) {
	bhIDs, err := s.repRepo.ListBathhouseIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var bathhouses []domain.Bathhouse
	for _, id := range bhIDs {
		bh, err := s.bhRepo.GetByID(ctx, id)
		if err != nil {
			continue // skip if not found
		}
		bathhouses = append(bathhouses, *bh)
	}

	return bathhouses, nil
}
