package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type InviteRepresentativeInput struct {
	UserEmail   string
	BathhouseID uuid.UUID
	Role        domain.RepresentativeRole
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
	auditSvc AuditLogService
}

func NewRepresentativeService(
	repRepo repository.RepresentativeRepository,
	userRepo repository.UserRepository,
	bhRepo repository.BathhouseRepository,
	auditSvc AuditLogService,
) RepresentativeService {
	return &representativeService{
		repRepo:  repRepo,
		userRepo: userRepo,
		bhRepo:   bhRepo,
		auditSvc: auditSvc,
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

	role := input.Role
	if !role.IsValid() {
		role = domain.RepRoleManager
	}

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      user.ID,
		BathhouseID: input.BathhouseID,
		OwnerID:     ownerID,
		Role:        role,
		CreatedAt:   time.Now(),
	}

	if err := s.repRepo.Create(ctx, rep); err != nil {
		return nil, err
	}

	// Audit log: representative invited
	if s.auditSvc != nil {
		_ = s.auditSvc.LogChange(ctx, "representative", rep.ID, ownerID, domain.AuditActionCreate, map[string]interface{}{
			"user_email":   input.UserEmail,
			"user_id":      user.ID.String(),
			"bathhouse_id": input.BathhouseID.String(),
			"role":         string(role),
		})
	}

	return rep, nil
}

func (s *representativeService) Revoke(ctx context.Context, ownerID uuid.UUID, representativeID uuid.UUID) error {
	rep, err := s.repRepo.GetByID(ctx, representativeID)
	if err != nil {
		return err
	}

	if rep.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	if err := s.repRepo.Delete(ctx, representativeID); err != nil {
		return err
	}

	// Audit log: representative revoked
	if s.auditSvc != nil {
		_ = s.auditSvc.LogChange(ctx, "representative", representativeID, ownerID, domain.AuditActionDelete, map[string]interface{}{
			"user_id":      rep.UserID.String(),
			"bathhouse_id": rep.BathhouseID.String(),
			"role":         string(rep.Role),
		})
	}

	// If the user has no remaining representative assignments, revert role to client
	remaining, err := s.repRepo.ListByUser(ctx, rep.UserID)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		user, err := s.userRepo.GetByID(ctx, rep.UserID)
		if err != nil {
			return err
		}
		if user.Role == domain.RoleRepresentative {
			user.Role = domain.RoleClient
			if err := s.userRepo.Update(ctx, user); err != nil {
				return err
			}
		}
	}

	return nil
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
			if errors.Is(err, domain.ErrNotFound) {
				continue // skip deleted bathhouses
			}
			return nil, err
		}
		bathhouses = append(bathhouses, *bh)
	}

	return bathhouses, nil
}
