package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AccessChecker struct {
	repRepo repository.RepresentativeRepository
	bhRepo  repository.BathhouseRepository
}

func NewAccessChecker(repRepo repository.RepresentativeRepository, bhRepo repository.BathhouseRepository) *AccessChecker {
	return &AccessChecker{repRepo: repRepo, bhRepo: bhRepo}
}

// CanManageBathhouse checks if the user can manage (update) a bathhouse.
// admin - always allowed
// owner - if they own the bathhouse
// representative with manager role - if assigned to the bathhouse
// observer representatives are denied (read-only)
func (a *AccessChecker) CanManageBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if userRole == domain.RoleAdmin {
		return nil
	}

	bh, err := a.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if userRole == domain.RoleOwner && bh.OwnerID == userID {
		return nil
	}

	if userRole == domain.RoleRepresentative {
		rep, err := a.repRepo.GetByUserAndBathhouse(ctx, userID, bathhouseID)
		if err == nil && rep.Role.CanWrite() {
			return nil
		}
	}

	return domain.ErrForbidden
}

// CanViewBathhouse checks if the user can view a bathhouse's management data.
// Both manager and observer representatives are allowed.
func (a *AccessChecker) CanViewBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if userRole == domain.RoleAdmin {
		return nil
	}

	bh, err := a.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return err
	}

	if userRole == domain.RoleOwner && bh.OwnerID == userID {
		return nil
	}

	if userRole == domain.RoleRepresentative {
		_, err := a.repRepo.GetByUserAndBathhouse(ctx, userID, bathhouseID)
		if err == nil {
			return nil
		}
	}

	return domain.ErrForbidden
}

// GetManagedBathhouseIDs returns bathhouse IDs the representative is assigned to.
func (a *AccessChecker) GetManagedBathhouseIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return a.repRepo.ListBathhouseIDsByUser(ctx, userID)
}
