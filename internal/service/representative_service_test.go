package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newRepresentativeService() (service.RepresentativeService, *mock.BathhouseRepo, *mock.UserRepo, *mock.RepresentativeRepo) {
	bhRepo := mock.NewBathhouseRepo()
	userRepo := mock.NewUserRepo()
	repRepo := mock.NewRepresentativeRepo()
	svc := service.NewRepresentativeService(repRepo, userRepo, bhRepo, nil)
	return svc, bhRepo, userRepo, repRepo
}

func TestRepresentativeService_Invite_Success(t *testing.T) {
	svc, bhRepo, userRepo, _ := newRepresentativeService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	targetUser := &domain.User{
		ID: uuid.New(), Email: "rep@example.com", Name: "Rep User",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), targetUser)

	rep, err := svc.Invite(context.Background(), ownerID, service.InviteRepresentativeInput{
		UserEmail:   "rep@example.com",
		BathhouseID: bh.ID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.UserID != targetUser.ID {
		t.Errorf("rep userID = %v, want %v", rep.UserID, targetUser.ID)
	}

	// Verify user role changed to representative
	updated, _ := userRepo.GetByID(context.Background(), targetUser.ID)
	if updated.Role != domain.RoleRepresentative {
		t.Errorf("user role should change to representative, got: %v", updated.Role)
	}
}

func TestRepresentativeService_Invite_NotOwnerForbidden(t *testing.T) {
	svc, bhRepo, userRepo, _ := newRepresentativeService()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	targetUser := &domain.User{
		ID: uuid.New(), Email: "rep@example.com", Name: "Rep User",
		Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), targetUser)

	_, err := svc.Invite(context.Background(), otherOwnerID, service.InviteRepresentativeInput{
		UserEmail:   "rep@example.com",
		BathhouseID: bh.ID,
	})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("non-owner should be forbidden from inviting, got: %v", err)
	}
}

func TestRepresentativeService_Invite_CantInviteOwnerOrAdmin(t *testing.T) {
	svc, bhRepo, userRepo, _ := newRepresentativeService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	adminUser := &domain.User{
		ID: uuid.New(), Email: "admin@example.com", Name: "Admin",
		Role: domain.RoleAdmin, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), adminUser)

	_, err := svc.Invite(context.Background(), ownerID, service.InviteRepresentativeInput{
		UserEmail:   "admin@example.com",
		BathhouseID: bh.ID,
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for admin user, got: %v", err)
	}
}

func TestRepresentativeService_ListByBathhouse_OwnerAllowed(t *testing.T) {
	svc, bhRepo, _, _ := newRepresentativeService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.ListByBathhouse(context.Background(), ownerID, bh.ID)
	if err != nil {
		t.Errorf("owner should list representatives: %v", err)
	}
}

func TestRepresentativeService_ListByBathhouse_OtherOwnerForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newRepresentativeService()
	ownerID := uuid.New()
	otherID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.ListByBathhouse(context.Background(), otherID, bh.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("non-owner should be forbidden, got: %v", err)
	}
}

func TestRepresentativeService_Revoke(t *testing.T) {
	svc, bhRepo, userRepo, repRepo := newRepresentativeService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	repUserID := uuid.New()
	repUser := &domain.User{
		ID: repUserID, Email: "rep@example.com", Name: "Rep",
		Role: domain.RoleRepresentative, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), repUser)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
		Role:        domain.RepRoleManager,
	}
	_ = repRepo.Create(context.Background(), rep)

	err := svc.Revoke(context.Background(), ownerID, rep.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user role reverted to client since no remaining assignments
	updated, _ := userRepo.GetByID(context.Background(), repUserID)
	if updated.Role != domain.RoleClient {
		t.Errorf("user role should revert to client after last revoke, got: %v", updated.Role)
	}
}

func TestRepresentativeService_GetMyBathhouses(t *testing.T) {
	svc, bhRepo, _, repRepo := newRepresentativeService()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
		Role:        domain.RepRoleManager,
	}
	_ = repRepo.Create(context.Background(), rep)

	bathhouses, err := svc.GetMyBathhouses(context.Background(), repUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bathhouses) != 1 {
		t.Errorf("expected 1 bathhouse, got %d", len(bathhouses))
	}
	if len(bathhouses) > 0 && bathhouses[0].ID != bh.ID {
		t.Errorf("expected bathhouse ID %v, got %v", bh.ID, bathhouses[0].ID)
	}
}

func TestRepresentativeService_Invite_AlreadyRepresentative(t *testing.T) {
	svc, bhRepo, userRepo, _ := newRepresentativeService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	targetUser := &domain.User{
		ID: uuid.New(), Email: "rep@example.com", Name: "Rep User",
		Role: domain.RoleRepresentative, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), targetUser)

	rep, err := svc.Invite(context.Background(), ownerID, service.InviteRepresentativeInput{
		UserEmail:   "rep@example.com",
		BathhouseID: bh.ID,
	})
	if err != nil {
		t.Fatalf("should allow inviting existing representative: %v", err)
	}
	if rep == nil {
		t.Fatal("representative should not be nil")
	}
}
