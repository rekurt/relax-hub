package antifraud_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func newTestEngine() (antifraud.FraudEngine, *mock.FraudFlagRepo) {
	repo := mock.NewFraudFlagRepo().(*mock.FraudFlagRepo)
	log := logger.New(logger.LevelError)
	engine := antifraud.NewFraudEngine(repo, log)
	return engine, repo
}

func TestCheckWalletTopUp_NoTrigger(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	err := engine.CheckWalletTopUp(ctx, userID, 100000, 2) // 2 cards, below threshold
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 0 {
		t.Errorf("expected 0 flags, got %d", len(result.Items))
	}
}

func TestCheckWalletTopUp_MultiCardTrigger(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	err := engine.CheckWalletTopUp(ctx, userID, 100000, 5) // 5 cards, above threshold of 3
	if err != nil {
		t.Fatalf("expected no error (action is freeze_wallet, not block), got %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(result.Items))
	}

	flag := result.Items[0]
	if flag.Rule != domain.FraudRuleMultiCardTopUp {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleMultiCardTopUp, flag.Rule)
	}
	if flag.Severity != domain.FraudSeverityHigh {
		t.Errorf("expected severity high, got %s", flag.Severity)
	}
	if flag.Action != domain.FraudActionFreezeWallet {
		t.Errorf("expected action freeze_wallet, got %s", flag.Action)
	}
}

func TestCheckBookingCreate_RapidBookings(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()
	ownerID := uuid.New()

	// 6 bookings in 1h - should trigger and block
	err := engine.CheckBookingCreate(ctx, userID, ownerID, 6)
	if err == nil {
		t.Fatal("expected ErrFraudDetected, got nil")
	}
	if err != domain.ErrFraudDetected {
		t.Fatalf("expected ErrFraudDetected, got %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleRapidBookings {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleRapidBookings, result.Items[0].Rule)
	}
}

func TestCheckBookingCreate_SelfBooking(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	// User books their own bathhouse - should trigger and block
	err := engine.CheckBookingCreate(ctx, userID, userID, 1)
	if err == nil {
		t.Fatal("expected ErrFraudDetected, got nil")
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) < 1 {
		t.Fatalf("expected at least 1 flag, got %d", len(result.Items))
	}

	found := false
	for _, f := range result.Items {
		if f.Rule == domain.FraudRuleSelfBooking {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected RULE_SELF_BOOKING flag")
	}
}

func TestCheckBookingCreate_NormalBooking(t *testing.T) {
	engine, _ := newTestEngine()
	ctx := context.Background()

	// Normal booking - different owner, low count
	err := engine.CheckBookingCreate(ctx, uuid.New(), uuid.New(), 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckBookingCancel_TopUpCancelCycle(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	// 3 top-up + cancel cycles in 7 days - should block
	err := engine.CheckBookingCancel(ctx, userID, 3)
	if err == nil {
		t.Fatal("expected ErrFraudDetected, got nil")
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleTopUpCancelCycle {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleTopUpCancelCycle, result.Items[0].Rule)
	}
}

func TestCheckBookingCancel_BelowThreshold(t *testing.T) {
	engine, _ := newTestEngine()
	ctx := context.Background()

	err := engine.CheckBookingCancel(ctx, uuid.New(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckPayoutRequest_Structuring(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	// 4 small withdrawals - should trigger freeze
	err := engine.CheckPayoutRequest(ctx, userID, 100000, 4)
	if err != nil {
		t.Fatalf("expected no error (action is freeze_wallet, not block), got %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleStructuring {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleStructuring, result.Items[0].Rule)
	}
	if result.Items[0].Action != domain.FraudActionFreezeWallet {
		t.Errorf("expected action freeze_wallet, got %s", result.Items[0].Action)
	}
}

func TestCheckPayoutRequest_Normal(t *testing.T) {
	engine, _ := newTestEngine()
	ctx := context.Background()

	err := engine.CheckPayoutRequest(ctx, uuid.New(), 500000, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckDormantBalance_Triggered(t *testing.T) {
	engine, repo := newTestEngine()
	ctx := context.Background()
	userID := uuid.New()

	// 60k balance, 45 days since last booking - should trigger notification
	err := engine.CheckDormantBalance(ctx, userID, 6000000, 45)
	if err != nil {
		t.Fatalf("expected no error (action is notify_admin, not block), got %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleDormantBalance {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleDormantBalance, result.Items[0].Rule)
	}
	if result.Items[0].Action != domain.FraudActionNotifyAdmin {
		t.Errorf("expected action notify_admin, got %s", result.Items[0].Action)
	}
}

func TestCheckDormantBalance_LowBalance(t *testing.T) {
	engine, _ := newTestEngine()
	ctx := context.Background()

	// Balance below 50k - should not trigger
	err := engine.CheckDormantBalance(ctx, uuid.New(), 4000000, 45)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckDormantBalance_RecentBooking(t *testing.T) {
	engine, _ := newTestEngine()
	ctx := context.Background()

	// High balance but recent booking - should not trigger
	err := engine.CheckDormantBalance(ctx, uuid.New(), 6000000, 15)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
