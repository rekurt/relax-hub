package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func createBathhouse(t *testing.T, bhRepo *mock.BathhouseRepo, ownerID uuid.UUID) *domain.Bathhouse {
	t.Helper()
	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Test Bathhouse",
		Address:      "123 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
		Status:       domain.BathhouseStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatal(err)
	}
	return bh
}

func TestAccessChecker_CanManageBathhouse_AdminAlwaysAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	adminID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), adminID, domain.RoleAdmin, bh.ID)
	if err != nil {
		t.Errorf("admin should always be allowed, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_OwnerAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Errorf("owner should be allowed for own bathhouse, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_OwnerForbiddenForOthersBathhouse(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), otherOwnerID, domain.RoleOwner, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_RepresentativeAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      repUserID,
		BathhouseID: bh.ID,
		OwnerID:     ownerID,
	}
	if err := repRepo.Create(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	err := checker.CanManageBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != nil {
		t.Errorf("representative assigned to bathhouse should be allowed, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_RepresentativeForbiddenForOtherBathhouse(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("representative not assigned to bathhouse should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_ClientForbidden(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), clientID, domain.RoleClient, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("client should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_NotFound(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	err := checker.CanManageBathhouse(context.Background(), uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("should return not found for non-existent bathhouse, got: %v", err)
	}
}
