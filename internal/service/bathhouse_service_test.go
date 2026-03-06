package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newBathhouseService() (service.BathhouseService, *mock.BathhouseRepo, *mock.RepresentativeRepo, *mock.BookingRepo) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	bookingRepo := mock.NewBookingRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	svc := service.NewBathhouseService(bhRepo, bookingRepo, access)
	return svc, bhRepo, repRepo, bookingRepo
}

func TestBathhouseService_Create_PendingStatus(t *testing.T) {
	svc, _, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh, err := svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "My Bathhouse",
		Address:      "123 Street",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bh.Status != domain.BathhouseStatusPending {
		t.Errorf("status = %q, want %q", bh.Status, domain.BathhouseStatusPending)
	}
	if bh.OwnerID != ownerID {
		t.Errorf("ownerID = %v, want %v", bh.OwnerID, ownerID)
	}
}

func TestBathhouseService_Create_InvalidInput(t *testing.T) {
	svc, _, _, _ := newBathhouseService()

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBathhouseInput{
		Name: "", // empty name
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should return ErrInvalidInput, got: %v", err)
	}
}

func TestBathhouseService_Update_OwnerAllowed(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	newName := "Updated Name"
	updated, err := svc.Update(context.Background(), ownerID, domain.RoleOwner, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}
}

func TestBathhouseService_Update_RepresentativeAllowed(t *testing.T) {
	svc, bhRepo, repRepo, _ := newBathhouseService()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
	}
	_ = repRepo.Create(context.Background(), rep)

	newName := "Updated by Rep"
	_, err := svc.Update(context.Background(), repUserID, domain.RoleRepresentative, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if err != nil {
		t.Errorf("representative should be allowed to update assigned bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Update_ClientForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	newName := "Hacked"
	_, err := svc.Update(context.Background(), clientID, domain.RoleClient, bh.ID, service.UpdateBathhouseInput{
		Name: &newName,
	})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from updating bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Delete_OwnerOnly(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := svc.Delete(context.Background(), ownerID, bh.ID)
	if err != nil {
		t.Fatalf("owner should be able to delete own bathhouse, got: %v", err)
	}

	_, err = svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("bathhouse should be deleted, got: %v", err)
	}
}

func TestBathhouseService_Delete_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := svc.Delete(context.Background(), otherOwnerID, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden from deleting bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Approve_AdminOnly(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	err := svc.Approve(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("approve should work: %v", err)
	}

	updated, _ := svc.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusActive {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusActive)
	}
}

func TestBathhouseService_Approve_NotPendingFails(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	bh := createBathhouse(t, bhRepo, uuid.New()) // active status

	err := svc.Approve(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for non-pending bathhouse, got: %v", err)
	}
}

func TestBathhouseService_Reject(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	err := svc.Reject(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("reject should work: %v", err)
	}

	// After rejection, public GetByID should not find the bathhouse
	_, err = svc.GetByID(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("rejected bathhouse should not be visible via GetByID, got: %v", err)
	}

	// Verify the status was actually changed in the repo
	updated, _ := bhRepo.GetByID(context.Background(), bh.ID)
	if updated.Status != domain.BathhouseStatusRejected {
		t.Errorf("status = %q, want %q", updated.Status, domain.BathhouseStatusRejected)
	}
}

func TestBathhouseService_GetWidgetKey_Success(t *testing.T) {
	svc, _, _, _ := newBathhouseService()
	ownerID := uuid.New()

	bh, err := svc.Create(context.Background(), ownerID, service.CreateBathhouseInput{
		Name:         "Test Bath",
		Address:      "123 St",
		CityID:       1,
		PricePerHour: 5000,
		MinDuration:  1,
		MaxGuests:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	apiKey, err := svc.GetWidgetKey(context.Background(), ownerID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey == "" {
		t.Errorf("expected non-empty API key")
	}
	if apiKey != bh.ApiKey {
		t.Errorf("apiKey = %q, want %q", apiKey, bh.ApiKey)
	}
}

func TestBathhouseService_GetWidgetKey_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.GetWidgetKey(context.Background(), otherOwnerID, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}

func TestBathhouseService_RegenerateWidgetKey_Success(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	originalKey := bh.ApiKey

	newKey, err := svc.RegenerateWidgetKey(context.Background(), ownerID, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newKey == "" {
		t.Errorf("expected non-empty API key")
	}
	if newKey == originalKey {
		t.Errorf("new key should be different from old key")
	}

	// Verify the key was updated in the repository
	updated, _ := bhRepo.GetByID(context.Background(), bh.ID)
	if updated.ApiKey != newKey {
		t.Errorf("stored apiKey = %q, want %q", updated.ApiKey, newKey)
	}
}

func TestBathhouseService_RegenerateWidgetKey_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBathhouseService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.RegenerateWidgetKey(context.Background(), otherOwnerID, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}
