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
// representative - if assigned to the bathhouse
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
		_, err := a.repRepo.GetByUserAndBathhouse(ctx, userID, bathhouseID)
		if err == nil {
			return nil
		}
	}

	return domain.ErrForbidden
}

// CanViewBathhouseBookings checks if the user can view bookings for a bathhouse.
// admin - always allowed
// owner - if they own the bathhouse
// representative - if assigned to the bathhouse
func (a *AccessChecker) CanViewBathhouseBookings(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	return a.CanManageBathhouse(ctx, userID, userRole, bathhouseID)
}
