package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type walletTestEnv struct {
	svc service.WalletService
}

func newWalletTestEnv() *walletTestEnv {
	walletRepo := mock.NewWalletRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewWalletService(walletRepo, log)
	return &walletTestEnv{svc: svc}
}

func createTestWallet(t *testing.T, env *walletTestEnv) *domain.Wallet {
	t.Helper()
	userID := uuid.New()
	wallet, err := env.svc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	return wallet
}

func TestWalletService_CreateWallet_Success(t *testing.T) {
	env := newWalletTestEnv()
	userID := uuid.New()

	wallet, err := env.svc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wallet.UserID != userID {
		t.Errorf("user_id = %v, want %v", wallet.UserID, userID)
	}
	if wallet.Balance != 0 {
		t.Errorf("balance = %d, want 0", wallet.Balance)
	}
	if wallet.Currency != domain.WalletCurrencyRUB {
		t.Errorf("currency = %v, want RUB", wallet.Currency)
	}
	if wallet.Status != domain.WalletStatusActive {
		t.Errorf("status = %v, want active", wallet.Status)
	}
}

func TestWalletService_CreateWallet_Duplicate(t *testing.T) {
	env := newWalletTestEnv()
	userID := uuid.New()

	_, err := env.svc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = env.svc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestWalletService_TopUp_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	amount := int64(100_000) // 1000 RUB

	tx, err := env.svc.TopUp(context.Background(), wallet.UserID, amount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Amount != amount {
		t.Errorf("amount = %d, want %d", tx.Amount, amount)
	}
	if tx.Type != domain.WalletTxTopUp {
		t.Errorf("type = %v, want topup", tx.Type)
	}
	if tx.BalanceAfter != amount {
		t.Errorf("balance_after = %d, want %d", tx.BalanceAfter, amount)
	}

	// Verify balance
	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != amount {
		t.Errorf("available = %d, want %d", balance.Available, amount)
	}
}

func TestWalletService_TopUp_BelowMinimum(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 10_000) // 100 RUB < 500 RUB min
	if !errors.Is(err, domain.ErrTopUpBelowMinimum) {
		t.Errorf("err = %v, want ErrTopUpBelowMinimum", err)
	}
}

func TestWalletService_TopUp_AboveMaximum(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 5_000_000) // 50,000 RUB > 30,000 max
	if !errors.Is(err, domain.ErrTopUpAboveMaximum) {
		t.Errorf("err = %v, want ErrTopUpAboveMaximum", err)
	}
}

func TestWalletService_TopUp_ExceedsBalanceLimit(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	// Fill up close to max
	for i := 0; i < 3; i++ {
		_, err := env.svc.TopUp(context.Background(), wallet.UserID, 3_000_000) // 30,000 RUB
		if err != nil {
			t.Fatalf("top up %d: %v", i, err)
		}
	}

	// Now balance is 9,000,000 (90,000 RUB). Another 30,000 would exceed 100,000 limit
	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 3_000_000)
	if !errors.Is(err, domain.ErrWalletLimitExceeded) {
		t.Errorf("err = %v, want ErrWalletLimitExceeded", err)
	}
}

func TestWalletService_TopUp_FrozenWallet(t *testing.T) {
	t.Skip("requires direct repo access to freeze wallet")
}

func TestWalletService_Spend_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	// Top up first
	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 500_000) // 5000 RUB
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	// Spend
	refID := uuid.New()
	tx, err := env.svc.Spend(context.Background(), wallet.ID, 200_000, "booking", &refID, "Оплата бронирования")
	if err != nil {
		t.Fatalf("spend: %v", err)
	}
	if tx.Amount != 200_000 {
		t.Errorf("amount = %d, want 200000", tx.Amount)
	}
	if tx.BalanceAfter != 300_000 {
		t.Errorf("balance_after = %d, want 300000", tx.BalanceAfter)
	}

	// Verify balance
	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != 300_000 {
		t.Errorf("available = %d, want 300000", balance.Available)
	}
}

func TestWalletService_Spend_InsufficientBalance(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 100_000) // 1000 RUB
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	_, err = env.svc.Spend(context.Background(), wallet.ID, 200_000, "booking", nil, "test")
	if !errors.Is(err, domain.ErrInsufficientWalletBalance) {
		t.Errorf("err = %v, want ErrInsufficientWalletBalance", err)
	}
}

func TestWalletService_Hold_And_Capture(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 500_000)
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	// Create hold
	bookingID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	hold, err := env.svc.Hold(context.Background(), wallet.ID, 200_000, "booking", &bookingID, "Бронирование", expiresAt)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if hold.Amount != 200_000 {
		t.Errorf("hold amount = %d, want 200000", hold.Amount)
	}

	// Check available balance is reduced
	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != 300_000 {
		t.Errorf("available = %d, want 300000", balance.Available)
	}
	if balance.HeldAmount != 200_000 {
		t.Errorf("held = %d, want 200000", balance.HeldAmount)
	}

	// Capture hold
	tx, err := env.svc.CaptureHold(context.Background(), hold.ID)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if tx.Amount != 200_000 {
		t.Errorf("capture amount = %d, want 200000", tx.Amount)
	}
	if tx.Type != domain.WalletTxHoldCapture {
		t.Errorf("type = %v, want hold_capture", tx.Type)
	}

	// Balance should be 300,000 with 0 held
	balance, err = env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance after capture: %v", err)
	}
	if balance.Balance != 300_000 {
		t.Errorf("balance = %d, want 300000", balance.Balance)
	}
	if balance.HeldAmount != 0 {
		t.Errorf("held = %d, want 0", balance.HeldAmount)
	}
}

func TestWalletService_Hold_And_Release(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 500_000)
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	hold, err := env.svc.Hold(context.Background(), wallet.ID, 200_000, "booking", nil, "test", expiresAt)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}

	// Release hold
	if err := env.svc.ReleaseHold(context.Background(), hold.ID); err != nil {
		t.Fatalf("release: %v", err)
	}

	// Full balance should be available again
	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != 500_000 {
		t.Errorf("available = %d, want 500000", balance.Available)
	}
	if balance.HeldAmount != 0 {
		t.Errorf("held = %d, want 0", balance.HeldAmount)
	}
}

func TestWalletService_Hold_InsufficientBalance(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 100_000)
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = env.svc.Hold(context.Background(), wallet.ID, 200_000, "booking", nil, "test", expiresAt)
	if !errors.Is(err, domain.ErrInsufficientWalletBalance) {
		t.Errorf("err = %v, want ErrInsufficientWalletBalance", err)
	}
}

func TestWalletService_Refund(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 500_000)
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	// Spend some
	_, err = env.svc.Spend(context.Background(), wallet.ID, 200_000, "booking", nil, "test")
	if err != nil {
		t.Fatalf("spend: %v", err)
	}

	// Refund
	refID := uuid.New()
	tx, err := env.svc.Refund(context.Background(), wallet.ID, 100_000, "booking", &refID, "Возврат за отмену")
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if tx.Type != domain.WalletTxRefund {
		t.Errorf("type = %v, want refund", tx.Type)
	}
	if tx.BalanceAfter != 400_000 {
		t.Errorf("balance_after = %d, want 400000", tx.BalanceAfter)
	}
}

func TestWalletService_AddBonus(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	tx, err := env.svc.AddBonus(context.Background(), wallet.ID, 50_000, domain.WalletTxWelcomeBonus, &expiresAt, "Приветственный бонус")
	if err != nil {
		t.Fatalf("add bonus: %v", err)
	}
	if tx.Amount != 50_000 {
		t.Errorf("amount = %d, want 50000", tx.Amount)
	}
	if !tx.IsBonus {
		t.Error("expected is_bonus = true")
	}
	if tx.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}

	// Check balance
	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != 50_000 {
		t.Errorf("available = %d, want 50000", balance.Available)
	}
}

func TestWalletService_AddBonus_ExceedsLimit(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	// Fill up to near max
	for i := 0; i < 3; i++ {
		_, err := env.svc.TopUp(context.Background(), wallet.UserID, 3_000_000)
		if err != nil {
			t.Fatalf("top up %d: %v", i, err)
		}
	}

	// Try to add bonus that would exceed limit
	_, err := env.svc.AddBonus(context.Background(), wallet.ID, 2_000_000, domain.WalletTxBonus, nil, "test")
	if !errors.Is(err, domain.ErrWalletLimitExceeded) {
		t.Errorf("err = %v, want ErrWalletLimitExceeded", err)
	}
}

func TestWalletService_ListTransactions(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	// Create some transactions
	_, _ = env.svc.TopUp(context.Background(), wallet.UserID, 500_000)
	_, _ = env.svc.Spend(context.Background(), wallet.ID, 100_000, "booking", nil, "test1")
	_, _ = env.svc.Spend(context.Background(), wallet.ID, 50_000, "booking", nil, "test2")

	result, err := env.svc.ListTransactions(context.Background(), wallet.UserID, domain.WalletTransactionFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
}

func TestWalletService_GetBalance_WithExpiringSoon(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	// Add bonus expiring in 7 days (within 14-day window)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err := env.svc.AddBonus(context.Background(), wallet.ID, 30_000, domain.WalletTxBonus, &expiresAt, "Expiring bonus")
	if err != nil {
		t.Fatalf("add bonus: %v", err)
	}

	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.ExpiringSoon != 30_000 {
		t.Errorf("expiring_soon = %d, want 30000", balance.ExpiringSoon)
	}
	if balance.EarliestExpiry == nil {
		t.Error("expected earliest_expiry to be set")
	}
}

func TestWalletService_CaptureHold_Expired(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	_, err := env.svc.TopUp(context.Background(), wallet.UserID, 500_000)
	if err != nil {
		t.Fatalf("top up: %v", err)
	}

	// Create hold that's already expired
	expiresAt := time.Now().Add(-1 * time.Hour)
	hold, err := env.svc.Hold(context.Background(), wallet.ID, 100_000, "booking", nil, "test", expiresAt)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}

	_, err = env.svc.CaptureHold(context.Background(), hold.ID)
	if !errors.Is(err, domain.ErrHoldExpired) {
		t.Errorf("err = %v, want ErrHoldExpired", err)
	}
}

func TestWalletService_GetWallet_NotFound(t *testing.T) {
	env := newWalletTestEnv()

	_, err := env.svc.GetWallet(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletService_Spend_WalletNotFound(t *testing.T) {
	env := newWalletTestEnv()

	_, err := env.svc.Spend(context.Background(), uuid.New(), 100, "booking", nil, "test")
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletService_AdminCredit_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	tx, err := env.svc.AdminCredit(context.Background(), wallet.ID, 200_000, "компенсация", adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != domain.WalletTxAdminCredit {
		t.Errorf("type = %v, want admin_credit", tx.Type)
	}
	if tx.Amount != 200_000 {
		t.Errorf("amount = %d, want 200000", tx.Amount)
	}
	if tx.BalanceAfter != 200_000 {
		t.Errorf("balance_after = %d, want 200000", tx.BalanceAfter)
	}

	balance, err := env.svc.GetBalance(context.Background(), wallet.UserID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != 200_000 {
		t.Errorf("available = %d, want 200000", balance.Available)
	}
}

func TestWalletService_AdminCredit_InvalidAmount(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	_, err := env.svc.AdminCredit(context.Background(), wallet.ID, 0, "test", adminID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}

	_, err = env.svc.AdminCredit(context.Background(), wallet.ID, -100, "test", adminID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestWalletService_AdminCredit_ExceedsLimit(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	// Credit close to max
	_, err := env.svc.AdminCredit(context.Background(), wallet.ID, 9_500_000, "test", adminID)
	if err != nil {
		t.Fatalf("first credit: %v", err)
	}

	// Try to exceed 10,000,000 (100,000 RUB)
	_, err = env.svc.AdminCredit(context.Background(), wallet.ID, 1_000_000, "test", adminID)
	if !errors.Is(err, domain.ErrWalletLimitExceeded) {
		t.Errorf("err = %v, want ErrWalletLimitExceeded", err)
	}
}

func TestWalletService_AdminCredit_WalletNotFound(t *testing.T) {
	env := newWalletTestEnv()
	adminID := uuid.New()

	_, err := env.svc.AdminCredit(context.Background(), uuid.New(), 100_000, "test", adminID)
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletService_AdminDebit_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	// Credit first to have balance
	_, err := env.svc.AdminCredit(context.Background(), wallet.ID, 500_000, "seed", adminID)
	if err != nil {
		t.Fatalf("credit: %v", err)
	}

	tx, err := env.svc.AdminDebit(context.Background(), wallet.ID, 200_000, "штраф", adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != domain.WalletTxAdminDebit {
		t.Errorf("type = %v, want admin_debit", tx.Type)
	}
	if tx.Amount != 200_000 {
		t.Errorf("amount = %d, want 200000", tx.Amount)
	}
	if tx.BalanceAfter != 300_000 {
		t.Errorf("balance_after = %d, want 300000", tx.BalanceAfter)
	}
}

func TestWalletService_AdminDebit_InsufficientBalance(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	_, err := env.svc.AdminDebit(context.Background(), wallet.ID, 100, "test", adminID)
	if !errors.Is(err, domain.ErrInsufficientWalletBalance) {
		t.Errorf("err = %v, want ErrInsufficientWalletBalance", err)
	}
}

func TestWalletService_AdminFreeze_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	err := env.svc.AdminFreeze(context.Background(), wallet.ID, "подозрительная активность", adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, err := env.svc.GetWalletByID(context.Background(), wallet.ID)
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if w.Status != domain.WalletStatusFrozen {
		t.Errorf("status = %v, want frozen", w.Status)
	}
}

func TestWalletService_AdminFreeze_AlreadyFrozen(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	if err := env.svc.AdminFreeze(context.Background(), wallet.ID, "test", adminID); err != nil {
		t.Fatalf("first freeze: %v", err)
	}

	err := env.svc.AdminFreeze(context.Background(), wallet.ID, "test", adminID)
	if !errors.Is(err, domain.ErrWalletFrozen) {
		t.Errorf("err = %v, want ErrWalletFrozen", err)
	}
}

func TestWalletService_AdminUnfreeze_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	if err := env.svc.AdminFreeze(context.Background(), wallet.ID, "test", adminID); err != nil {
		t.Fatalf("freeze: %v", err)
	}

	if err := env.svc.AdminUnfreeze(context.Background(), wallet.ID, "проверка завершена", adminID); err != nil {
		t.Fatalf("unfreeze: %v", err)
	}

	w, err := env.svc.GetWalletByID(context.Background(), wallet.ID)
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if w.Status != domain.WalletStatusActive {
		t.Errorf("status = %v, want active", w.Status)
	}
}

func TestWalletService_AdminUnfreeze_NotFrozen(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)
	adminID := uuid.New()

	err := env.svc.AdminUnfreeze(context.Background(), wallet.ID, "test", adminID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestWalletService_GetWalletByID_Success(t *testing.T) {
	env := newWalletTestEnv()
	wallet := createTestWallet(t, env)

	w, err := env.svc.GetWalletByID(context.Background(), wallet.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID != wallet.ID {
		t.Errorf("id = %v, want %v", w.ID, wallet.ID)
	}
}

func TestWalletService_GetWalletByID_NotFound(t *testing.T) {
	env := newWalletTestEnv()

	_, err := env.svc.GetWalletByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}
