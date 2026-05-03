package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

// promoTestNotifSvc is a no-op notification service for testing.
type promoTestNotifSvc struct{}

func (m *promoTestNotifSvc) Send(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _, _ string, _ map[string]string) error {
	return nil
}
func (m *promoTestNotifSvc) List(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Notification], error) {
	return &domain.PaginatedResult[domain.Notification]{}, nil
}
func (m *promoTestNotifSvc) MarkAsRead(_ context.Context, _, _ uuid.UUID) error   { return nil }
func (m *promoTestNotifSvc) MarkAllAsRead(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *promoTestNotifSvc) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *promoTestNotifSvc) GetPreferences(_ context.Context, _ uuid.UUID) (*domain.NotificationPreferences, error) {
	return nil, nil
}
func (m *promoTestNotifSvc) UpdatePreferences(_ context.Context, _ uuid.UUID, _ *domain.NotificationPreferences) error {
	return nil
}
func (m *promoTestNotifSvc) HasRecentByType(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _ time.Time) (bool, error) {
	return false, nil
}
func (m *promoTestNotifSvc) GetEventPreferences(_ context.Context, _ uuid.UUID) ([]domain.NotificationEventPreference, error) {
	return nil, nil
}
func (m *promoTestNotifSvc) UpdateEventPreferences(_ context.Context, _ uuid.UUID, _ []domain.NotificationEventPreference) error {
	return nil
}

// promoTestWalletSvc is a simple mock for testing promotion budget deduction.
type promoTestWalletSvc struct {
	wallets      map[uuid.UUID]*domain.Wallet
	spendCalled  int
	spendErr     error
}

func newPromoTestWalletSvc() *promoTestWalletSvc {
	return &promoTestWalletSvc{wallets: make(map[uuid.UUID]*domain.Wallet)}
}

func (m *promoTestWalletSvc) GetWallet(_ context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	w, ok := m.wallets[userID]
	if !ok {
		return nil, domain.ErrWalletNotFound
	}
	return w, nil
}

func (m *promoTestWalletSvc) Spend(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	m.spendCalled++
	if m.spendErr != nil {
		return nil, m.spendErr
	}
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}

func (m *promoTestWalletSvc) CreateWallet(_ context.Context, _ uuid.UUID, _ domain.WalletCurrency) (*domain.Wallet, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) EnsureWallet(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) TopUp(_ context.Context, _ uuid.UUID, _ int64) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) Hold(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) CaptureHold(_ context.Context, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) ReleaseHold(_ context.Context, _ uuid.UUID) error { return nil }
func (m *promoTestWalletSvc) Refund(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) AddBonus(_ context.Context, _ uuid.UUID, _ int64, _ domain.WalletTransactionType, _ *time.Time, _ string) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) GetBalance(_ context.Context, _ uuid.UUID) (*service.WalletBalanceSummary, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) ListTransactions(_ context.Context, _ uuid.UUID, _ domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	return nil, nil
}
func (m *promoTestWalletSvc) GetActiveHolds(_ context.Context, _ uuid.UUID) ([]domain.WalletHold, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) ExpireBonuses(_ context.Context) (int, error)                        { return 0, nil }
func (m *promoTestWalletSvc) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error)  { return 0, nil }
func (m *promoTestWalletSvc) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error           { return nil }
func (m *promoTestWalletSvc) AdminCredit(_ context.Context, _ uuid.UUID, _ int64, _ string, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) AdminDebit(_ context.Context, _ uuid.UUID, _ int64, _ string, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *promoTestWalletSvc) AdminFreeze(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error   { return nil }
func (m *promoTestWalletSvc) AdminUnfreeze(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error { return nil }
func (m *promoTestWalletSvc) GetWalletByID(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	return nil, domain.ErrWalletNotFound
}

func newPromotionTestEnv() (*mock.PromotionRepo, *mock.BathhouseRepo, *promoTestWalletSvc, service.PromotionService) {
	promoRepo := mock.NewPromotionRepo()
	bhRepo := mock.NewBathhouseRepo()
	walletSvc := newPromoTestWalletSvc()
	notifSvc := &promoTestNotifSvc{}
	log := logger.New(logger.LevelWarn)
	svc := service.NewPromotionService(promoRepo, bhRepo, walletSvc, notifSvc, log)
	return promoRepo, bhRepo, walletSvc, svc
}

func TestPromotionService_Create(t *testing.T) {
	promoRepo, _, _, svc := newPromotionTestEnv()

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     uuid.New(),
		DailyBidKopecks: 5000,
		BudgetKopecks:   100000,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionActive,
	}

	if err := svc.Create(context.Background(), promo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := promoRepo.GetByID(context.Background(), promo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.DailyBidKopecks != 5000 {
		t.Errorf("daily_bid = %d, want 5000", got.DailyBidKopecks)
	}
}

func TestPromotionService_PauseResume(t *testing.T) {
	promoRepo, _, _, svc := newPromotionTestEnv()

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     uuid.New(),
		DailyBidKopecks: 5000,
		BudgetKopecks:   100000,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionActive,
	}
	_ = promoRepo.Create(context.Background(), promo)

	// Pause
	if err := svc.Pause(context.Background(), promo.ID); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	got, _ := promoRepo.GetByID(context.Background(), promo.ID)
	if got.Status != domain.PromotionPaused {
		t.Errorf("status after pause = %q, want %q", got.Status, domain.PromotionPaused)
	}

	// Resume
	if err := svc.Resume(context.Background(), promo.ID); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	got, _ = promoRepo.GetByID(context.Background(), promo.ID)
	if got.Status != domain.PromotionActive {
		t.Errorf("status after resume = %q, want %q", got.Status, domain.PromotionActive)
	}
}

func TestPromotionService_Resume_InsufficientBudget(t *testing.T) {
	promoRepo, _, _, svc := newPromotionTestEnv()

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     uuid.New(),
		DailyBidKopecks: 5000,
		BudgetKopecks:   6000,
		SpentKopecks:    5500,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionPaused,
	}
	_ = promoRepo.Create(context.Background(), promo)

	err := svc.Resume(context.Background(), promo.ID)
	if err != domain.ErrPromotionBudgetExhausted {
		t.Errorf("Resume should fail with ErrPromotionBudgetExhausted, got %v", err)
	}
}

func TestPromotionService_DeductDailyBudgets_Success(t *testing.T) {
	promoRepo, bhRepo, walletSvc, svc := newPromotionTestEnv()

	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	walletSvc.wallets[ownerID] = &domain.Wallet{
		ID:       uuid.New(),
		UserID:   ownerID,
		Balance:  100000,
		Currency: domain.WalletCurrencyRUB,
	}

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     bh.ID,
		DailyBidKopecks: 5000,
		BudgetKopecks:   50000,
		SpentKopecks:    0,
		StartDate:       time.Now().Add(-24 * time.Hour),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionActive,
	}
	_ = promoRepo.Create(context.Background(), promo)

	err := svc.DeductDailyBudgets(context.Background())
	if err != nil {
		t.Fatalf("DeductDailyBudgets: %v", err)
	}

	// Verify spend was called
	if walletSvc.spendCalled != 1 {
		t.Errorf("spendCalled = %d, want 1", walletSvc.spendCalled)
	}

	// Verify spent was updated
	got, _ := promoRepo.GetByID(context.Background(), promo.ID)
	if got.SpentKopecks != 5000 {
		t.Errorf("spent = %d, want 5000", got.SpentKopecks)
	}
}

func TestPromotionService_DeductDailyBudgets_Exhausted(t *testing.T) {
	promoRepo, bhRepo, walletSvc, svc := newPromotionTestEnv()

	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	walletSvc.wallets[ownerID] = &domain.Wallet{
		ID:       uuid.New(),
		UserID:   ownerID,
		Balance:  100000,
		Currency: domain.WalletCurrencyRUB,
	}

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     bh.ID,
		DailyBidKopecks: 5000,
		BudgetKopecks:   7000,
		SpentKopecks:    5000,
		StartDate:       time.Now().Add(-24 * time.Hour),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionActive,
	}
	_ = promoRepo.Create(context.Background(), promo)

	err := svc.DeductDailyBudgets(context.Background())
	if err != nil {
		t.Fatalf("DeductDailyBudgets: %v", err)
	}

	// Remaining = 2000 < daily bid 5000 -> should be exhausted
	got, _ := promoRepo.GetByID(context.Background(), promo.ID)
	if got.Status != domain.PromotionExhausted {
		t.Errorf("status = %q, want %q", got.Status, domain.PromotionExhausted)
	}
}

func TestPromotionService_DeductDailyBudgets_Expired(t *testing.T) {
	promoRepo, _, _, svc := newPromotionTestEnv()

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     uuid.New(),
		DailyBidKopecks: 5000,
		BudgetKopecks:   50000,
		SpentKopecks:    0,
		StartDate:       time.Now().Add(-48 * time.Hour),
		EndDate:         time.Now().Add(-1 * time.Hour), // expired
		Status:          domain.PromotionActive,
	}
	_ = promoRepo.Create(context.Background(), promo)

	err := svc.DeductDailyBudgets(context.Background())
	if err != nil {
		t.Fatalf("DeductDailyBudgets: %v", err)
	}

	got, _ := promoRepo.GetByID(context.Background(), promo.ID)
	if got.Status != domain.PromotionExpired {
		t.Errorf("status = %q, want %q", got.Status, domain.PromotionExpired)
	}
}

func TestPromotionService_DeductDailyBudgets_InsufficientWallet(t *testing.T) {
	promoRepo, bhRepo, walletSvc, svc := newPromotionTestEnv()

	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	walletSvc.wallets[ownerID] = &domain.Wallet{
		ID:       uuid.New(),
		UserID:   ownerID,
		Balance:  1000, // less than daily bid
		Currency: domain.WalletCurrencyRUB,
	}
	walletSvc.spendErr = domain.ErrInsufficientWalletBalance

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     bh.ID,
		DailyBidKopecks: 5000,
		BudgetKopecks:   50000,
		SpentKopecks:    0,
		StartDate:       time.Now().Add(-24 * time.Hour),
		EndDate:         time.Now().Add(30 * 24 * time.Hour),
		Status:          domain.PromotionActive,
	}
	_ = promoRepo.Create(context.Background(), promo)

	err := svc.DeductDailyBudgets(context.Background())
	if err != nil {
		t.Fatalf("DeductDailyBudgets: %v", err)
	}

	// Should be paused due to insufficient wallet balance
	got, _ := promoRepo.GetByID(context.Background(), promo.ID)
	if got.Status != domain.PromotionPaused {
		t.Errorf("status = %q, want %q (insufficient wallet should pause)", got.Status, domain.PromotionPaused)
	}
}

func TestPromotionDomain_Validate_MinBid(t *testing.T) {
	promo := &domain.Promotion{
		BathhouseID:     uuid.New(),
		DailyBidKopecks: 1000, // below MinDailyBidKopecks (5000)
		BudgetKopecks:   50000,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(24 * time.Hour),
		Status:          domain.PromotionActive,
	}

	if err := promo.Validate(); err == nil {
		t.Error("Validate should fail when DailyBidKopecks < MinDailyBidKopecks")
	}

	promo.DailyBidKopecks = domain.MinDailyBidKopecks
	if err := promo.Validate(); err != nil {
		t.Errorf("Validate should pass at MinDailyBidKopecks, got %v", err)
	}
}

func TestPromotionDomain_Statistics(t *testing.T) {
	promo := &domain.Promotion{
		BudgetKopecks:   100000,
		SpentKopecks:    30000,
		ImpressionCount: 1000,
		ClickCount:      50,
	}

	remaining := promo.RemainingBudget()
	if remaining != 70000 {
		t.Errorf("RemainingBudget = %d, want 70000", remaining)
	}

	ctr := promo.CTR()
	if ctr != 5.0 {
		t.Errorf("CTR = %f, want 5.0", ctr)
	}

	cpc := promo.CostPerClick()
	expectedCPC := float64(30000) / float64(50)
	if cpc != expectedCPC {
		t.Errorf("CostPerClick = %f, want %f", cpc, expectedCPC)
	}
}

func TestPromotionDomain_Statistics_ZeroDivision(t *testing.T) {
	promo := &domain.Promotion{
		BudgetKopecks:   100000,
		SpentKopecks:    0,
		ImpressionCount: 0,
		ClickCount:      0,
	}

	if promo.CTR() != 0 {
		t.Errorf("CTR with 0 impressions should be 0")
	}
	if promo.CostPerClick() != 0 {
		t.Errorf("CPC with 0 clicks should be 0")
	}
}
