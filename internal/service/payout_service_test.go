package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type payoutTestEnv struct {
	payoutSvc service.PayoutService
	walletSvc service.WalletService
}

func newPayoutTestEnv() *payoutTestEnv {
	walletRepo := mock.NewWalletRepo()
	payoutRepo := mock.NewPayoutRepo()
	log := logger.New(logger.LevelWarn)
	walletSvc := service.NewWalletService(walletRepo, log)
	payoutSvc := service.NewPayoutService(payoutRepo, walletRepo, log)
	return &payoutTestEnv{payoutSvc: payoutSvc, walletSvc: walletSvc}
}

func setupOwnerWithBalance(t *testing.T, env *payoutTestEnv, balance int64) (*domain.Wallet, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	wallet, err := env.walletSvc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	// Top up in increments within limits
	for balance > 0 {
		amount := int64(3_000_000) // max per top-up
		if amount > balance {
			amount = balance
		}
		if amount < domain.WalletTopUpMinRUB {
			amount = domain.WalletTopUpMinRUB
		}
		_, err := env.walletSvc.TopUp(context.Background(), userID, amount)
		if err != nil {
			t.Fatalf("top up: %v", err)
		}
		balance -= amount
		if balance < domain.WalletTopUpMinRUB && balance > 0 {
			break
		}
	}
	return wallet, userID
}

func TestPayoutService_RequestPayout_Success(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000) // 10,000 RUB

	bankDetails := json.RawMessage(`{"bik":"044525225","account":"40817810000000000001"}`)
	payout, err := env.payoutSvc.RequestPayout(context.Background(), userID, 500_000, bankDetails) // 5000 RUB
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payout.Amount != 500_000 {
		t.Errorf("amount = %d, want 500000", payout.Amount)
	}
	if payout.Status != domain.PayoutStatusPending {
		t.Errorf("status = %v, want pending", payout.Status)
	}
	if payout.UserID != userID {
		t.Errorf("user_id = %v, want %v", payout.UserID, userID)
	}
}

func TestPayoutService_RequestPayout_BelowMinimum(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000)

	_, err := env.payoutSvc.RequestPayout(context.Background(), userID, 10_000, nil) // 100 RUB < 500 min
	if !errors.Is(err, domain.ErrPayoutBelowMinimum) {
		t.Errorf("err = %v, want ErrPayoutBelowMinimum", err)
	}
}

func TestPayoutService_RequestPayout_InsufficientBalance(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 500_000) // 5000 RUB

	_, err := env.payoutSvc.RequestPayout(context.Background(), userID, 600_000, nil) // 6000 RUB > 5000 available
	if !errors.Is(err, domain.ErrInsufficientWalletBalance) {
		t.Errorf("err = %v, want ErrInsufficientWalletBalance", err)
	}
}

func TestPayoutService_RequestPayout_WalletNotFound(t *testing.T) {
	env := newPayoutTestEnv()

	_, err := env.payoutSvc.RequestPayout(context.Background(), uuid.New(), 100_000, nil)
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestPayoutService_RequestPayout_DeductsFromAvailable(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000) // 10,000 RUB

	// First payout of 500 RUB
	_, err := env.payoutSvc.RequestPayout(context.Background(), userID, 50_000, nil)
	if err != nil {
		t.Fatalf("first payout: %v", err)
	}

	// Available for payout should now be reduced by pending amount
	available, err := env.payoutSvc.CalculateAvailableBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("calc available: %v", err)
	}
	if available != 950_000 {
		t.Errorf("available = %d, want 950000", available)
	}
}

func TestPayoutService_SetAutoPayoutThreshold_Success(t *testing.T) {
	env := newPayoutTestEnv()
	userID := uuid.New()

	if err := env.payoutSvc.SetAutoPayoutThreshold(context.Background(), userID, 500_000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, err := env.payoutSvc.GetAutoPayoutSettings(context.Background(), userID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.Threshold != 500_000 {
		t.Errorf("threshold = %d, want 500000", settings.Threshold)
	}
}

func TestPayoutService_SetAutoPayoutThreshold_Disable(t *testing.T) {
	env := newPayoutTestEnv()
	userID := uuid.New()

	// Set threshold
	if err := env.payoutSvc.SetAutoPayoutThreshold(context.Background(), userID, 500_000); err != nil {
		t.Fatalf("set threshold: %v", err)
	}

	// Disable by setting to 0
	if err := env.payoutSvc.SetAutoPayoutThreshold(context.Background(), userID, 0); err != nil {
		t.Fatalf("disable: %v", err)
	}

	settings, err := env.payoutSvc.GetAutoPayoutSettings(context.Background(), userID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.Threshold != 0 {
		t.Errorf("threshold = %d, want 0", settings.Threshold)
	}
}

func TestPayoutService_SetAutoPayoutThreshold_BelowMinimum(t *testing.T) {
	env := newPayoutTestEnv()
	userID := uuid.New()

	err := env.payoutSvc.SetAutoPayoutThreshold(context.Background(), userID, 10_000) // 100 RUB < 500 min
	if !errors.Is(err, domain.ErrPayoutBelowMinimum) {
		t.Errorf("err = %v, want ErrPayoutBelowMinimum", err)
	}
}

func TestPayoutService_SetAutoPayoutThreshold_Negative(t *testing.T) {
	env := newPayoutTestEnv()
	userID := uuid.New()

	err := env.payoutSvc.SetAutoPayoutThreshold(context.Background(), userID, -100)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPayoutService_ProcessPayout_Success(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000)

	payout, err := env.payoutSvc.RequestPayout(context.Background(), userID, 200_000, nil)
	if err != nil {
		t.Fatalf("request payout: %v", err)
	}

	if err := env.payoutSvc.ProcessPayout(context.Background(), payout.ID); err != nil {
		t.Fatalf("process payout: %v", err)
	}

	// Check wallet balance decreased
	balance, err := env.walletSvc.GetBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Balance != 800_000 {
		t.Errorf("balance = %d, want 800000", balance.Balance)
	}
}

func TestPayoutService_ProcessPayout_AlreadyProcessed(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000)

	payout, err := env.payoutSvc.RequestPayout(context.Background(), userID, 200_000, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if err := env.payoutSvc.ProcessPayout(context.Background(), payout.ID); err != nil {
		t.Fatalf("first process: %v", err)
	}

	err = env.payoutSvc.ProcessPayout(context.Background(), payout.ID)
	if !errors.Is(err, domain.ErrPayoutAlreadyProcessed) {
		t.Errorf("err = %v, want ErrPayoutAlreadyProcessed", err)
	}
}

func TestPayoutService_ProcessPayout_NotFound(t *testing.T) {
	env := newPayoutTestEnv()

	err := env.payoutSvc.ProcessPayout(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPayoutNotFound) {
		t.Errorf("err = %v, want ErrPayoutNotFound", err)
	}
}

func TestPayoutService_GetPayoutHistory(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 3_000_000) // 30,000 RUB

	// Create multiple payouts
	for i := 0; i < 3; i++ {
		_, err := env.payoutSvc.RequestPayout(context.Background(), userID, 50_000, nil)
		if err != nil {
			t.Fatalf("payout %d: %v", i, err)
		}
	}

	result, err := env.payoutSvc.GetPayoutHistory(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
}

func TestPayoutService_CalculateAvailableBalance(t *testing.T) {
	env := newPayoutTestEnv()
	_, userID := setupOwnerWithBalance(t, env, 1_000_000)

	available, err := env.payoutSvc.CalculateAvailableBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("calc: %v", err)
	}
	if available != 1_000_000 {
		t.Errorf("available = %d, want 1000000", available)
	}

	// Request a payout, available should decrease
	_, err = env.payoutSvc.RequestPayout(context.Background(), userID, 300_000, nil)
	if err != nil {
		t.Fatalf("payout: %v", err)
	}

	available, err = env.payoutSvc.CalculateAvailableBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("calc after payout: %v", err)
	}
	if available != 700_000 {
		t.Errorf("available = %d, want 700000", available)
	}
}

func TestPayoutService_GetAutoPayoutSettings_Default(t *testing.T) {
	env := newPayoutTestEnv()

	settings, err := env.payoutSvc.GetAutoPayoutSettings(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.Threshold != 0 {
		t.Errorf("threshold = %d, want 0 (disabled by default)", settings.Threshold)
	}
}
