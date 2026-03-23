package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func createTestWallet(t *testing.T, repo *mock.WalletRepo, userID uuid.UUID) *domain.Wallet {
	t.Helper()
	if userID == uuid.Nil {
		userID = uuid.New()
	}
	w := &domain.Wallet{
		UserID:   userID,
		Currency: domain.WalletCurrencyRUB,
	}
	if err := repo.Create(context.Background(), w); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	return w
}

func TestWalletRepo_Create(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	w := &domain.Wallet{
		UserID:   uuid.New(),
		Currency: domain.WalletCurrencyRUB,
	}
	err := repo.Create(context.Background(), w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if w.Status != domain.WalletStatusActive {
		t.Errorf("status = %v, want active", w.Status)
	}
	if w.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestWalletRepo_Create_Duplicate(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	userID := uuid.New()

	createTestWallet(t, repo, userID)

	w2 := &domain.Wallet{UserID: userID, Currency: domain.WalletCurrencyRUB}
	err := repo.Create(context.Background(), w2)
	if err != domain.ErrAlreadyExists {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestWalletRepo_GetByID(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	w := createTestWallet(t, repo, uuid.Nil)

	found, err := repo.GetByID(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.UserID != w.UserID {
		t.Errorf("user_id = %v, want %v", found.UserID, w.UserID)
	}
}

func TestWalletRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrWalletNotFound {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletRepo_GetByUserID(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	userID := uuid.New()
	w := createTestWallet(t, repo, userID)

	found, err := repo.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != w.ID {
		t.Errorf("id = %v, want %v", found.ID, w.ID)
	}
}

func TestWalletRepo_GetByUserID_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	_, err := repo.GetByUserID(context.Background(), uuid.New())
	if err != domain.ErrWalletNotFound {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletRepo_UpdateBalance(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	w := createTestWallet(t, repo, uuid.Nil)

	err := repo.UpdateBalance(context.Background(), w.ID, 0, 100000, 0, 20000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), w.ID)
	if found.Balance != 100000 {
		t.Errorf("balance = %d, want 100000", found.Balance)
	}
	if found.HeldAmount != 20000 {
		t.Errorf("held_amount = %d, want 20000", found.HeldAmount)
	}
}

func TestWalletRepo_UpdateBalance_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	err := repo.UpdateBalance(context.Background(), uuid.New(), 0, 100000, 0, 0)
	if err != domain.ErrWalletNotFound {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletRepo_UpdateStatus(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	w := createTestWallet(t, repo, uuid.Nil)

	err := repo.UpdateStatus(context.Background(), w.ID, domain.WalletStatusFrozen)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), w.ID)
	if found.Status != domain.WalletStatusFrozen {
		t.Errorf("status = %v, want frozen", found.Status)
	}
}

func TestWalletRepo_UpdateStatus_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	err := repo.UpdateStatus(context.Background(), uuid.New(), domain.WalletStatusFrozen)
	if err != domain.ErrWalletNotFound {
		t.Errorf("err = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletRepo_CreateTransaction(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	tx := &domain.WalletTransaction{
		WalletID:    walletID,
		Type:        domain.WalletTxTopUp,
		Amount:      50000,
		BalanceAfter: 50000,
		Description: "Top up",
	}
	err := repo.CreateTransaction(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if tx.Status != domain.WalletTxStatusCompleted {
		t.Errorf("status = %v, want completed", tx.Status)
	}
}

func TestWalletRepo_ListTransactions(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	for i := 0; i < 3; i++ {
		_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
			WalletID:     walletID,
			Type:         domain.WalletTxTopUp,
			Amount:       int64((i + 1) * 10000),
			BalanceAfter: int64((i + 1) * 10000),
		})
	}

	// Unrelated transaction
	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID:     uuid.New(),
		Type:         domain.WalletTxTopUp,
		Amount:       99999,
		BalanceAfter: 99999,
	})

	result, err := repo.ListTransactions(context.Background(), domain.WalletTransactionFilter{
		WalletID: &walletID,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}
}

func TestWalletRepo_ListTransactions_FilterByType(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxTopUp, Amount: 10000, BalanceAfter: 10000,
	})
	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxSpend, Amount: 5000, BalanceAfter: 5000,
	})

	txType := domain.WalletTxTopUp
	result, err := repo.ListTransactions(context.Background(), domain.WalletTransactionFilter{
		WalletID: &walletID,
		Type:     &txType,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("total_count = %d, want 1", result.TotalCount)
	}
}

func TestWalletRepo_GetExpiringBonuses(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 1000, BalanceAfter: 1000,
		IsBonus: true, ExpiresAt: &past,
	})
	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 2000, BalanceAfter: 3000,
		IsBonus: true, ExpiresAt: &future,
	})

	result, err := repo.GetExpiringBonuses(context.Background(), walletID, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Amount != 1000 {
		t.Errorf("amount = %d, want 1000", result[0].Amount)
	}
}

func TestWalletRepo_GetBonusTransactionsForSpending(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	soon := time.Now().Add(1 * time.Hour)
	later := time.Now().Add(48 * time.Hour)

	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 2000, BalanceAfter: 2000,
		IsBonus: true, ExpiresAt: &later,
	})
	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 1000, BalanceAfter: 3000,
		IsBonus: true, ExpiresAt: &soon,
	})

	result, err := repo.GetBonusTransactionsForSpending(context.Background(), walletID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	// Expiring sooner should come first
	if result[0].Amount != 1000 {
		t.Errorf("first amount = %d, want 1000 (sooner expiry)", result[0].Amount)
	}
}

func TestWalletRepo_ExpireBonuses(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	past := time.Now().Add(-1 * time.Hour)
	tx := &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 1000, BalanceAfter: 1000,
		IsBonus: true, ExpiresAt: &past,
	}
	_ = repo.CreateTransaction(context.Background(), tx)

	err := repo.ExpireBonuses(context.Background(), []uuid.UUID{tx.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should no longer appear in bonus transactions for spending
	result, _ := repo.GetBonusTransactionsForSpending(context.Background(), walletID)
	if len(result) != 0 {
		t.Errorf("len = %d, want 0 (expired)", len(result))
	}
}

func TestWalletRepo_GetExpiringBonusesSoon(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	in3Days := time.Now().Add(3 * 24 * time.Hour)
	in30Days := time.Now().Add(30 * 24 * time.Hour)

	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 1000, BalanceAfter: 1000,
		IsBonus: true, ExpiresAt: &in3Days,
	})
	_ = repo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		WalletID: walletID, Type: domain.WalletTxBonus, Amount: 2000, BalanceAfter: 3000,
		IsBonus: true, ExpiresAt: &in30Days,
	})

	from := time.Now()
	to := time.Now().Add(7 * 24 * time.Hour)
	result, err := repo.GetExpiringBonusesSoon(context.Background(), walletID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Amount != 1000 {
		t.Errorf("amount = %d, want 1000", result[0].Amount)
	}
}

func TestWalletRepo_CreateHold(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	hold := &domain.WalletHold{
		WalletID:    walletID,
		Amount:      50000,
		Description: "booking hold",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	err := repo.CreateHold(context.Background(), hold)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hold.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if hold.Status != domain.WalletHoldStatusActive {
		t.Errorf("status = %v, want active", hold.Status)
	}
}

func TestWalletRepo_GetHoldByID(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	hold := &domain.WalletHold{
		WalletID:  uuid.New(),
		Amount:    50000,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateHold(context.Background(), hold)

	found, err := repo.GetHoldByID(context.Background(), hold.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Amount != hold.Amount {
		t.Errorf("amount = %d, want %d", found.Amount, hold.Amount)
	}
}

func TestWalletRepo_GetHoldByID_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	_, err := repo.GetHoldByID(context.Background(), uuid.New())
	if err != domain.ErrHoldNotFound {
		t.Errorf("err = %v, want ErrHoldNotFound", err)
	}
}

func TestWalletRepo_UpdateHoldStatus_Capture(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	hold := &domain.WalletHold{
		WalletID:  uuid.New(),
		Amount:    50000,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateHold(context.Background(), hold)

	now := time.Now()
	err := repo.UpdateHoldStatus(context.Background(), hold.ID, domain.WalletHoldStatusCaptured, &now, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetHoldByID(context.Background(), hold.ID)
	if found.Status != domain.WalletHoldStatusCaptured {
		t.Errorf("status = %v, want captured", found.Status)
	}
	if found.CapturedAt == nil {
		t.Error("expected CapturedAt to be set")
	}
}

func TestWalletRepo_UpdateHoldStatus_Release(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	hold := &domain.WalletHold{
		WalletID:  uuid.New(),
		Amount:    50000,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateHold(context.Background(), hold)

	now := time.Now()
	err := repo.UpdateHoldStatus(context.Background(), hold.ID, domain.WalletHoldStatusReleased, nil, &now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetHoldByID(context.Background(), hold.ID)
	if found.Status != domain.WalletHoldStatusReleased {
		t.Errorf("status = %v, want released", found.Status)
	}
	if found.ReleasedAt == nil {
		t.Error("expected ReleasedAt to be set")
	}
}

func TestWalletRepo_UpdateHoldStatus_NotFound(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	err := repo.UpdateHoldStatus(context.Background(), uuid.New(), domain.WalletHoldStatusCaptured, nil, nil)
	if err != domain.ErrHoldNotFound {
		t.Errorf("err = %v, want ErrHoldNotFound", err)
	}
}

func TestWalletRepo_GetActiveHolds(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	walletID := uuid.New()

	_ = repo.CreateHold(context.Background(), &domain.WalletHold{
		WalletID: walletID, Amount: 10000, ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	_ = repo.CreateHold(context.Background(), &domain.WalletHold{
		WalletID: walletID, Amount: 20000, ExpiresAt: time.Now().Add(48 * time.Hour),
	})

	// Captured hold - should not appear
	capturedHold := &domain.WalletHold{
		WalletID: walletID, Amount: 30000, ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateHold(context.Background(), capturedHold)
	now := time.Now()
	_ = repo.UpdateHoldStatus(context.Background(), capturedHold.ID, domain.WalletHoldStatusCaptured, &now, nil)

	// Other wallet's hold
	_ = repo.CreateHold(context.Background(), &domain.WalletHold{
		WalletID: uuid.New(), Amount: 40000, ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	result, err := repo.GetActiveHolds(context.Background(), walletID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len = %d, want 2", len(result))
	}
}

func TestWalletRepo_GetExpiredHolds(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)

	_ = repo.CreateHold(context.Background(), &domain.WalletHold{
		WalletID: uuid.New(), Amount: 10000, ExpiresAt: time.Now().Add(-1 * time.Hour),
	})
	_ = repo.CreateHold(context.Background(), &domain.WalletHold{
		WalletID: uuid.New(), Amount: 20000, ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	result, err := repo.GetExpiredHolds(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("len = %d, want 1", len(result))
	}
	if result[0].Amount != 10000 {
		t.Errorf("amount = %d, want 10000", result[0].Amount)
	}
}

func TestWalletRepo_IsolatesData(t *testing.T) {
	repo := mock.NewWalletRepo().(*mock.WalletRepo)
	w := createTestWallet(t, repo, uuid.Nil)

	found, _ := repo.GetByID(context.Background(), w.ID)
	found.Balance = 999999

	found2, _ := repo.GetByID(context.Background(), w.ID)
	if found2.Balance != 0 {
		t.Errorf("balance was mutated: got %d, want 0", found2.Balance)
	}
}
