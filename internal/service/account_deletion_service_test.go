package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newAccountDeletionService(userRepo *mock.UserRepo) service.AccountDeletionService {
	log := logger.New(logger.LevelWarn)
	sessionSvc := service.NewSessionService(mock.NewSessionRepo(), log)
	notifSvc := &noopNotifService{}
	return service.NewAccountDeletionService(userRepo, sessionSvc, notifSvc, log)
}

func createDeletionTestUser(t *testing.T, userRepo *mock.UserRepo) *domain.User {
	t.Helper()
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestAccountDeletionService_RequestDeletion(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)
	user := createDeletionTestUser(t, userRepo)

	err := svc.RequestDeletion(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion fields are set
	updated, err := userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if updated.DeletionRequestedAt == nil {
		t.Error("expected DeletionRequestedAt to be set")
	}
	if updated.DeletionScheduledAt == nil {
		t.Error("expected DeletionScheduledAt to be set")
	}
	if updated.DeletionScheduledAt != nil {
		expectedSchedule := updated.DeletionRequestedAt.AddDate(0, 0, 30)
		diff := updated.DeletionScheduledAt.Sub(expectedSchedule)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("expected scheduled_at ~30 days from requested_at, got diff %v", diff)
		}
	}
}

func TestAccountDeletionService_RequestDeletion_AlreadyPending(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)
	user := createDeletionTestUser(t, userRepo)

	_ = svc.RequestDeletion(context.Background(), user.ID)

	// Second request should fail
	err := svc.RequestDeletion(context.Background(), user.ID)
	if err != domain.ErrAccountDeletionPending {
		t.Errorf("expected ErrAccountDeletionPending, got %v", err)
	}
}

func TestAccountDeletionService_RestoreAccount(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)
	user := createDeletionTestUser(t, userRepo)

	_ = svc.RequestDeletion(context.Background(), user.ID)

	err := svc.RestoreAccount(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion fields are cleared
	updated, err := userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if updated.DeletionRequestedAt != nil {
		t.Error("expected DeletionRequestedAt to be nil after restore")
	}
	if updated.DeletionScheduledAt != nil {
		t.Error("expected DeletionScheduledAt to be nil after restore")
	}
}

func TestAccountDeletionService_RestoreAccount_NotPending(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)
	user := createDeletionTestUser(t, userRepo)

	err := svc.RestoreAccount(context.Background(), user.ID)
	if err != domain.ErrAccountDeletionNotPending {
		t.Errorf("expected ErrAccountDeletionNotPending, got %v", err)
	}
}

func TestAccountDeletionService_ExecuteDeletion(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)
	user := createDeletionTestUser(t, userRepo)

	_ = svc.RequestDeletion(context.Background(), user.ID)

	err := svc.ExecuteDeletion(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user is anonymized
	updated, err := userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if updated.Name != "Deleted User" {
		t.Errorf("expected name 'Deleted User', got %q", updated.Name)
	}
	if updated.Phone != "" {
		t.Errorf("expected phone to be empty, got %q", updated.Phone)
	}
	if updated.IsActive {
		t.Error("expected user to be inactive")
	}
	if updated.DeletionRequestedAt != nil {
		t.Error("expected DeletionRequestedAt to be nil after execution")
	}
	if updated.DeletionScheduledAt != nil {
		t.Error("expected DeletionScheduledAt to be nil after execution")
	}
	if updated.Bio != "" {
		t.Error("expected bio to be empty")
	}
	if updated.AvatarURL != "" {
		t.Error("expected avatar to be empty")
	}
}

func TestAccountDeletionService_CheckPendingDeletions(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)

	// Create two users, both with past scheduled deletions
	user1 := &domain.User{
		ID: uuid.New(), Email: "u1@test.com", Name: "User1", Role: domain.RoleClient, IsActive: true,
	}
	user2 := &domain.User{
		ID: uuid.New(), Email: "u2@test.com", Name: "User2", Role: domain.RoleClient, IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user1)
	_ = userRepo.Create(context.Background(), user2)

	// Set past deletion schedule (should be executed)
	pastTime := time.Now().Add(-24 * time.Hour)
	_ = userRepo.SetDeletionSchedule(context.Background(), user1.ID, &pastTime, &pastTime)
	_ = userRepo.SetDeletionSchedule(context.Background(), user2.ID, &pastTime, &pastTime)

	executed, err := svc.CheckPendingDeletions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executed != 2 {
		t.Errorf("expected 2 executed, got %d", executed)
	}

	// Verify both anonymized
	u1, _ := userRepo.GetByID(context.Background(), user1.ID)
	if u1.Name != "Deleted User" {
		t.Errorf("expected user1 to be anonymized, got name %q", u1.Name)
	}
	u2, _ := userRepo.GetByID(context.Background(), user2.ID)
	if u2.Name != "Deleted User" {
		t.Errorf("expected user2 to be anonymized, got name %q", u2.Name)
	}
}

func TestAccountDeletionService_RequestDeletion_UserNotFound(t *testing.T) {
	userRepo := mock.NewUserRepo()
	svc := newAccountDeletionService(userRepo)

	err := svc.RequestDeletion(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
