package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type referralTestEnv struct {
	svc      service.ReferralService
	userRepo *mock.UserRepo
}

func newReferralTestEnv() *referralTestEnv {
	referralRepo := mock.NewReferralRepo()
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewReferralService(referralRepo, userRepo, log)
	return &referralTestEnv{
		svc:      svc,
		userRepo: userRepo,
	}
}

func createTestUser(t *testing.T, env *referralTestEnv) *domain.User {
	t.Helper()
	user := &domain.User{
		ID:       uuid.New(),
		Email:    uuid.New().String() + "@test.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	if err := env.userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return user
}

func TestReferralService_GenerateCode_Success(t *testing.T) {
	env := newReferralTestEnv()
	user := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty code")
	}
	if len(code) != 8 {
		t.Errorf("code length = %d, want 8", len(code))
	}
}

func TestReferralService_GenerateCode_Idempotent(t *testing.T) {
	env := newReferralTestEnv()
	user := createTestUser(t, env)

	code1, err := env.svc.GenerateCode(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	code2, err := env.svc.GenerateCode(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if code1 != code2 {
		t.Errorf("codes differ: %q != %q", code1, code2)
	}
}

func TestReferralService_GenerateCode_UserNotFound(t *testing.T) {
	env := newReferralTestEnv()

	_, err := env.svc.GenerateCode(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestReferralService_RegisterReferral_Success(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	err = env.svc.RegisterReferral(context.Background(), code, referee.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReferralService_RegisterReferral_EmptyCode(t *testing.T) {
	env := newReferralTestEnv()
	referee := createTestUser(t, env)

	err := env.svc.RegisterReferral(context.Background(), "", referee.ID)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestReferralService_RegisterReferral_InvalidCode(t *testing.T) {
	env := newReferralTestEnv()
	referee := createTestUser(t, env)

	err := env.svc.RegisterReferral(context.Background(), "nonexistent", referee.ID)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestReferralService_RegisterReferral_SelfReferral(t *testing.T) {
	env := newReferralTestEnv()
	user := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	err = env.svc.RegisterReferral(context.Background(), code, user.ID)
	if err != domain.ErrSelfReferral {
		t.Errorf("err = %v, want ErrSelfReferral", err)
	}
}

func TestReferralService_RegisterReferral_AlreadyReferred(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("register referral: %v", err)
	}

	err = env.svc.RegisterReferral(context.Background(), code, referee.ID)
	if err != domain.ErrAlreadyReferred {
		t.Errorf("err = %v, want ErrAlreadyReferred", err)
	}
}

func TestReferralService_CompleteReferral_Success(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("register referral: %v", err)
	}

	_, err = env.svc.CompleteReferral(context.Background(), referee.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify referrer got bonus
	referrerBalance, err := env.svc.GetBalance(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("get referrer balance: %v", err)
	}
	if referrerBalance.Balance != 50000 {
		t.Errorf("referrer balance = %d, want 50000", referrerBalance.Balance)
	}

	// Verify referee got bonus
	refereeBalance, err := env.svc.GetBalance(context.Background(), referee.ID)
	if err != nil {
		t.Fatalf("get referee balance: %v", err)
	}
	if refereeBalance.Balance != 50000 {
		t.Errorf("referee balance = %d, want 50000", refereeBalance.Balance)
	}
}

func TestReferralService_CompleteReferral_NotReferred(t *testing.T) {
	env := newReferralTestEnv()
	user := createTestUser(t, env)

	// Should not error for non-referred users
	_, err := env.svc.CompleteReferral(context.Background(), user.ID)
	if err != nil {
		t.Errorf("unexpected error for non-referred user: %v", err)
	}
}

func TestReferralService_CompleteReferral_Idempotent(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("register referral: %v", err)
	}

	if _, err := env.svc.CompleteReferral(context.Background(), referee.ID); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	_, err = env.svc.CompleteReferral(context.Background(), referee.ID)
	if err != nil {
		t.Errorf("second complete should not error: %v", err)
	}

	// Balance should not be doubled
	referrerBalance, err := env.svc.GetBalance(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("get referrer balance: %v", err)
	}
	if referrerBalance.Balance != 50000 {
		t.Errorf("referrer balance = %d, want 50000 (should not double)", referrerBalance.Balance)
	}
}

func TestReferralService_GetBalance_NoBalance(t *testing.T) {
	env := newReferralTestEnv()

	balance, err := env.svc.GetBalance(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Balance != 0 {
		t.Errorf("balance = %d, want 0", balance.Balance)
	}
	if balance.TotalEarned != 0 {
		t.Errorf("total_earned = %d, want 0", balance.TotalEarned)
	}
}

func TestReferralService_UseBalance_Success(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("register referral: %v", err)
	}
	if _, err := env.svc.CompleteReferral(context.Background(), referee.ID); err != nil {
		t.Fatalf("complete referral: %v", err)
	}

	bookingID := uuid.New()
	err = env.svc.UseBalance(context.Background(), referrer.ID, 20000, bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	balance, err := env.svc.GetBalance(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Balance != 30000 {
		t.Errorf("balance = %d, want 30000", balance.Balance)
	}
}

func TestReferralService_UseBalance_Insufficient(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)
	referee := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("register referral: %v", err)
	}
	if _, err := env.svc.CompleteReferral(context.Background(), referee.ID); err != nil {
		t.Fatalf("complete referral: %v", err)
	}

	err = env.svc.UseBalance(context.Background(), referrer.ID, 100000, uuid.New())
	if err != domain.ErrInsufficientReferralBalance {
		t.Errorf("err = %v, want ErrInsufficientReferralBalance", err)
	}
}

func TestReferralService_UseBalance_InvalidAmount(t *testing.T) {
	env := newReferralTestEnv()

	err := env.svc.UseBalance(context.Background(), uuid.New(), 0, uuid.New())
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for zero amount", err)
	}

	err = env.svc.UseBalance(context.Background(), uuid.New(), -100, uuid.New())
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for negative amount", err)
	}
}

func TestReferralService_GetStats(t *testing.T) {
	env := newReferralTestEnv()
	referrer := createTestUser(t, env)

	// Invite 2 users, complete 1
	referee1 := createTestUser(t, env)
	referee2 := createTestUser(t, env)

	code, err := env.svc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee1.ID); err != nil {
		t.Fatalf("register referral 1: %v", err)
	}
	if err := env.svc.RegisterReferral(context.Background(), code, referee2.ID); err != nil {
		t.Fatalf("register referral 2: %v", err)
	}
	if _, err := env.svc.CompleteReferral(context.Background(), referee1.ID); err != nil {
		t.Fatalf("complete referral: %v", err)
	}

	stats, err := env.svc.GetStats(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalInvited != 2 {
		t.Errorf("total_invited = %d, want 2", stats.TotalInvited)
	}
	if stats.TotalCompleted != 1 {
		t.Errorf("total_completed = %d, want 1", stats.TotalCompleted)
	}
	if stats.TotalEarned != 50000 {
		t.Errorf("total_earned = %d, want 50000", stats.TotalEarned)
	}
}

func TestReferralService_GetStats_NoReferrals(t *testing.T) {
	env := newReferralTestEnv()

	stats, err := env.svc.GetStats(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalInvited != 0 {
		t.Errorf("total_invited = %d, want 0", stats.TotalInvited)
	}
	if stats.TotalCompleted != 0 {
		t.Errorf("total_completed = %d, want 0", stats.TotalCompleted)
	}
	if stats.TotalEarned != 0 {
		t.Errorf("total_earned = %d, want 0", stats.TotalEarned)
	}
}
