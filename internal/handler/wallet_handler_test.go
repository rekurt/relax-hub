package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/service"
)

type mockWalletService struct {
	createWalletFn     func(ctx context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error)
	getWalletFn        func(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	topUpFn            func(ctx context.Context, userID uuid.UUID, amount int64) (*domain.WalletTransaction, error)
	spendFn            func(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error)
	holdFn             func(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string, expiresAt time.Time) (*domain.WalletHold, error)
	captureHoldFn      func(ctx context.Context, holdID uuid.UUID) (*domain.WalletTransaction, error)
	releaseHoldFn      func(ctx context.Context, holdID uuid.UUID) error
	refundFn           func(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error)
	addBonusFn         func(ctx context.Context, walletID uuid.UUID, amount int64, bonusType domain.WalletTransactionType, expiresAt *time.Time, description string) (*domain.WalletTransaction, error)
	getBalanceFn       func(ctx context.Context, userID uuid.UUID) (*service.WalletBalanceSummary, error)
	listTransactionsFn func(ctx context.Context, userID uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error)
	getActiveHoldsFn   func(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error)
	expireBonusesFn    func(ctx context.Context) (int, error)
	adminCreditFn      func(ctx context.Context, walletID uuid.UUID, amount int64, reason string, adminID uuid.UUID) (*domain.WalletTransaction, error)
	adminDebitFn       func(ctx context.Context, walletID uuid.UUID, amount int64, reason string, adminID uuid.UUID) (*domain.WalletTransaction, error)
	adminFreezeFn      func(ctx context.Context, walletID uuid.UUID, reason string, adminID uuid.UUID) error
	adminUnfreezeFn    func(ctx context.Context, walletID uuid.UUID, reason string, adminID uuid.UUID) error
	getWalletByIDFn    func(ctx context.Context, walletID uuid.UUID) (*domain.Wallet, error)
}

func (m *mockWalletService) CreateWallet(ctx context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
	if m.createWalletFn != nil {
		return m.createWalletFn(ctx, userID, currency)
	}
	return nil, nil
}

func (m *mockWalletService) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	if m.getWalletFn != nil {
		return m.getWalletFn(ctx, userID)
	}
	return nil, domain.ErrWalletNotFound
}

func (m *mockWalletService) TopUp(ctx context.Context, userID uuid.UUID, amount int64) (*domain.WalletTransaction, error) {
	if m.topUpFn != nil {
		return m.topUpFn(ctx, userID, amount)
	}
	return nil, nil
}

func (m *mockWalletService) Spend(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	if m.spendFn != nil {
		return m.spendFn(ctx, walletID, amount, refType, refID, description)
	}
	return nil, nil
}

func (m *mockWalletService) Hold(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string, expiresAt time.Time) (*domain.WalletHold, error) {
	if m.holdFn != nil {
		return m.holdFn(ctx, walletID, amount, refType, refID, description, expiresAt)
	}
	return nil, nil
}

func (m *mockWalletService) CaptureHold(ctx context.Context, holdID uuid.UUID) (*domain.WalletTransaction, error) {
	if m.captureHoldFn != nil {
		return m.captureHoldFn(ctx, holdID)
	}
	return nil, nil
}

func (m *mockWalletService) ReleaseHold(ctx context.Context, holdID uuid.UUID) error {
	if m.releaseHoldFn != nil {
		return m.releaseHoldFn(ctx, holdID)
	}
	return nil
}

func (m *mockWalletService) Refund(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	if m.refundFn != nil {
		return m.refundFn(ctx, walletID, amount, refType, refID, description)
	}
	return nil, nil
}

func (m *mockWalletService) AddBonus(ctx context.Context, walletID uuid.UUID, amount int64, bonusType domain.WalletTransactionType, expiresAt *time.Time, description string) (*domain.WalletTransaction, error) {
	if m.addBonusFn != nil {
		return m.addBonusFn(ctx, walletID, amount, bonusType, expiresAt, description)
	}
	return nil, nil
}

func (m *mockWalletService) GetBalance(ctx context.Context, userID uuid.UUID) (*service.WalletBalanceSummary, error) {
	if m.getBalanceFn != nil {
		return m.getBalanceFn(ctx, userID)
	}
	return nil, domain.ErrWalletNotFound
}

func (m *mockWalletService) ListTransactions(ctx context.Context, userID uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	if m.listTransactionsFn != nil {
		return m.listTransactionsFn(ctx, userID, filter)
	}
	return nil, nil
}

func (m *mockWalletService) GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error) {
	if m.getActiveHoldsFn != nil {
		return m.getActiveHoldsFn(ctx, walletID)
	}
	return nil, nil
}

func (m *mockWalletService) ExpireBonuses(ctx context.Context) (int, error) {
	if m.expireBonusesFn != nil {
		return m.expireBonusesFn(ctx)
	}
	return 0, nil
}

func (m *mockWalletService) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockWalletService) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockWalletService) AdminCredit(ctx context.Context, walletID uuid.UUID, amount int64, reason string, adminID uuid.UUID) (*domain.WalletTransaction, error) {
	if m.adminCreditFn != nil {
		return m.adminCreditFn(ctx, walletID, amount, reason, adminID)
	}
	return nil, nil
}

func (m *mockWalletService) AdminDebit(ctx context.Context, walletID uuid.UUID, amount int64, reason string, adminID uuid.UUID) (*domain.WalletTransaction, error) {
	if m.adminDebitFn != nil {
		return m.adminDebitFn(ctx, walletID, amount, reason, adminID)
	}
	return nil, nil
}

func (m *mockWalletService) AdminFreeze(ctx context.Context, walletID uuid.UUID, reason string, adminID uuid.UUID) error {
	if m.adminFreezeFn != nil {
		return m.adminFreezeFn(ctx, walletID, reason, adminID)
	}
	return nil
}

func (m *mockWalletService) AdminUnfreeze(ctx context.Context, walletID uuid.UUID, reason string, adminID uuid.UUID) error {
	if m.adminUnfreezeFn != nil {
		return m.adminUnfreezeFn(ctx, walletID, reason, adminID)
	}
	return nil
}

func (m *mockWalletService) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*domain.Wallet, error) {
	if m.getWalletByIDFn != nil {
		return m.getWalletByIDFn(ctx, walletID)
	}
	return nil, nil
}

type mockPaymentProvider struct {
	createPaymentFn func(ctx context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error)
}

func (m *mockPaymentProvider) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
	if m.createPaymentFn != nil {
		return m.createPaymentFn(ctx, req)
	}
	return nil, nil
}

func (m *mockPaymentProvider) GetPaymentStatus(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockPaymentProvider) CreateRefund(_ context.Context, _ string, _ int64) error { return nil }
func (m *mockPaymentProvider) CapturePayment(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *mockPaymentProvider) CancelPayment(_ context.Context, _ string) error { return nil }

func TestWalletHandler_GetWallet(t *testing.T) {
	userID := uuid.New()
	expiry := time.Now().Add(7 * 24 * time.Hour)

	walletSvc := &mockWalletService{
		getBalanceFn: func(ctx context.Context, uid uuid.UUID) (*service.WalletBalanceSummary, error) {
			if uid == userID {
				return &service.WalletBalanceSummary{
					Balance:        500_000,
					HeldAmount:     50_000,
					Available:      450_000,
					Currency:       domain.WalletCurrencyRUB,
					ExpiringSoon:   100_000,
					EarliestExpiry: &expiry,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet", h.GetWallet)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	if data["balance"].(float64) != 500_000 {
		t.Errorf("expected balance 500000, got %v", data["balance"])
	}
	if data["available"].(float64) != 450_000 {
		t.Errorf("expected available 450000, got %v", data["available"])
	}
	if data["held_amount"].(float64) != 50_000 {
		t.Errorf("expected held_amount 50000, got %v", data["held_amount"])
	}
	if data["currency"] != "RUB" {
		t.Errorf("expected currency RUB, got %v", data["currency"])
	}
	if data["expiring_soon"].(float64) != 100_000 {
		t.Errorf("expected expiring_soon 100000, got %v", data["expiring_soon"])
	}
	if data["earliest_expiry"] == nil {
		t.Error("expected earliest_expiry to be set")
	}
}

func TestWalletHandler_GetWallet_CreatesMissingWallet(t *testing.T) {
	userID := uuid.New()
	calls := 0
	created := false

	walletSvc := &mockWalletService{
		getBalanceFn: func(ctx context.Context, uid uuid.UUID) (*service.WalletBalanceSummary, error) {
			calls++
			if calls == 1 {
				return nil, domain.ErrWalletNotFound
			}
			if uid != userID {
				return nil, domain.ErrWalletNotFound
			}
			return &service.WalletBalanceSummary{
				Balance:    0,
				HeldAmount: 0,
				Available:  0,
				Currency:   domain.WalletCurrencyRUB,
			}, nil
		},
		createWalletFn: func(ctx context.Context, uid uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
			if uid != userID {
				t.Fatalf("expected userID %s, got %s", userID, uid)
			}
			if currency != domain.WalletCurrencyRUB {
				t.Fatalf("expected RUB wallet, got %s", currency)
			}
			created = true
			return &domain.Wallet{ID: uuid.New(), UserID: uid, Currency: currency}, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet", h.GetWallet)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !created {
		t.Fatal("expected missing wallet to be created")
	}
	if calls != 2 {
		t.Fatalf("expected balance to be loaded twice, got %d", calls)
	}
}

func TestWalletHandler_TopUp(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			return &domain.Wallet{ID: walletID, UserID: uid, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive}, nil
		},
	}

	mockPay := &mockPaymentProvider{
		createPaymentFn: func(ctx context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
			return &payment.PaymentResult{ExternalID: "ext-123", ConfirmationURL: "https://pay.example.com/confirm"}, nil
		},
	}

	h := NewWalletHandler(walletSvc, mockPay, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	body, _ := json.Marshal(topUpRequest{Amount: 100_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data := resp.Data.(map[string]interface{})
	if data["confirmation_url"] != "https://pay.example.com/confirm" {
		t.Errorf("expected confirmation_url, got %v", data["confirmation_url"])
	}
	if data["amount"].(float64) != 100_000 {
		t.Errorf("expected amount 100000, got %v", data["amount"])
	}
}

func TestWalletHandler_TopUp_NoPaymentProvider(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			return &domain.Wallet{ID: walletID, UserID: uid, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive}, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	body, _ := json.Marshal(topUpRequest{Amount: 100_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWalletHandler_TopUp_BelowMinimum(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			return &domain.Wallet{ID: walletID, UserID: uid, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive}, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	body, _ := json.Marshal(topUpRequest{Amount: 1_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWalletHandler_TopUp_InvalidBody(t *testing.T) {
	userID := uuid.New()

	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader([]byte("not json")))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWalletHandler_TopUp_ZeroAmount(t *testing.T) {
	userID := uuid.New()

	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	body, _ := json.Marshal(topUpRequest{Amount: 0})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWalletHandler_ListTransactions(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	txID := uuid.New()

	walletSvc := &mockWalletService{
		listTransactionsFn: func(ctx context.Context, uid uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
			if uid == userID {
				return &domain.PaginatedResult[domain.WalletTransaction]{
					Items: []domain.WalletTransaction{
						{
							ID:           txID,
							WalletID:     uuid.New(),
							Type:         domain.WalletTxTopUp,
							Amount:       100_000,
							BalanceAfter: 100_000,
							Status:       domain.WalletTxStatusCompleted,
							Description:  "Пополнение кошелька",
							CreatedAt:    now,
						},
					},
					Page:       1,
					PageSize:   20,
					TotalCount: 1,
					TotalPages: 1,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/transactions", h.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/transactions?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	if resp.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count 1, got %d", resp.Meta.TotalCount)
	}

	items, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(items))
	}
}

func TestWalletHandler_ListTransactions_WithTypeFilter(t *testing.T) {
	userID := uuid.New()

	var capturedFilter domain.WalletTransactionFilter
	walletSvc := &mockWalletService{
		listTransactionsFn: func(ctx context.Context, uid uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
			capturedFilter = filter
			return &domain.PaginatedResult[domain.WalletTransaction]{
				Items:      []domain.WalletTransaction{},
				Page:       1,
				PageSize:   20,
				TotalCount: 0,
				TotalPages: 0,
			}, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/transactions", h.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/transactions?type=topup&is_bonus=true", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if capturedFilter.Type == nil || *capturedFilter.Type != domain.WalletTxTopUp {
		t.Error("expected type filter to be topup")
	}
	if capturedFilter.IsBonus == nil || *capturedFilter.IsBonus != true {
		t.Error("expected is_bonus filter to be true")
	}
}

func TestWalletHandler_ListHolds(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	holdID := uuid.New()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			if uid == userID {
				return &domain.Wallet{
					ID:       walletID,
					UserID:   userID,
					Balance:  500_000,
					Currency: domain.WalletCurrencyRUB,
					Status:   domain.WalletStatusActive,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
		getActiveHoldsFn: func(ctx context.Context, wid uuid.UUID) ([]domain.WalletHold, error) {
			if wid == walletID {
				return []domain.WalletHold{
					{
						ID:            holdID,
						WalletID:      walletID,
						Amount:        50_000,
						Status:        domain.WalletHoldStatusActive,
						Description:   "Бронирование",
						ReferenceType: "booking",
						ExpiresAt:     expiresAt,
						CreatedAt:     now,
					},
				}, nil
			}
			return nil, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/holds", h.ListHolds)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/holds", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	items, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 hold, got %d", len(items))
	}

	hold := items[0].(map[string]interface{})
	if hold["id"] != holdID.String() {
		t.Errorf("expected hold id %s, got %v", holdID.String(), hold["id"])
	}
	if hold["amount"].(float64) != 50_000 {
		t.Errorf("expected amount 50000, got %v", hold["amount"])
	}
}

func TestWalletHandler_ListHolds_WalletNotFound(t *testing.T) {
	userID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/holds", h.ListHolds)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/holds", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestWalletHandler_GetWallet_Unauthorized(t *testing.T) {
	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet", h.GetWallet)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestWalletHandler_TopUp_WalletNotFound(t *testing.T) {
	userID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletFn: func(ctx context.Context, uid uuid.UUID) (*domain.Wallet, error) {
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/topup", h.TopUp)

	body, _ := json.Marshal(topUpRequest{Amount: 100_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/topup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestWalletHandler_AdminCreditWallet(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()
	txID := uuid.New()

	walletSvc := &mockWalletService{
		adminCreditFn: func(_ context.Context, wid uuid.UUID, amount int64, reason string, aid uuid.UUID) (*domain.WalletTransaction, error) {
			if wid == walletID && amount == 100_000 && aid == adminID {
				return &domain.WalletTransaction{
					ID:           txID,
					WalletID:     walletID,
					Type:         domain.WalletTxAdminCredit,
					Amount:       100_000,
					BalanceAfter: 100_000,
					Status:       domain.WalletTxStatusCompleted,
					Description:  "Начисление администратором: " + reason,
					CreatedAt:    time.Now(),
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/credit", h.AdminCreditWallet)

	body, _ := json.Marshal(map[string]interface{}{"amount": 100_000, "reason": "компенсация"})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/"+walletID.String()+"/credit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	data := resp.Data.(map[string]interface{})
	if data["type"] != "admin_credit" {
		t.Errorf("expected type admin_credit, got %v", data["type"])
	}
}

func TestWalletHandler_AdminCreditWallet_MissingReason(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/credit", h.AdminCreditWallet)

	body, _ := json.Marshal(map[string]interface{}{"amount": 100_000})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/"+walletID.String()+"/credit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWalletHandler_AdminDebitWallet(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()
	txID := uuid.New()

	walletSvc := &mockWalletService{
		adminDebitFn: func(_ context.Context, wid uuid.UUID, amount int64, _ string, aid uuid.UUID) (*domain.WalletTransaction, error) {
			if wid == walletID && amount == 50_000 && aid == adminID {
				return &domain.WalletTransaction{
					ID:           txID,
					WalletID:     walletID,
					Type:         domain.WalletTxAdminDebit,
					Amount:       50_000,
					BalanceAfter: 450_000,
					Status:       domain.WalletTxStatusCompleted,
					CreatedAt:    time.Now(),
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/debit", h.AdminDebitWallet)

	body, _ := json.Marshal(map[string]interface{}{"amount": 50_000, "reason": "штраф"})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/"+walletID.String()+"/debit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWalletHandler_AdminFreezeWallet(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{
		adminFreezeFn: func(_ context.Context, wid uuid.UUID, _ string, _ uuid.UUID) error {
			if wid == walletID {
				return nil
			}
			return domain.ErrWalletNotFound
		},
		getWalletByIDFn: func(_ context.Context, wid uuid.UUID) (*domain.Wallet, error) {
			if wid == walletID {
				return &domain.Wallet{
					ID:       walletID,
					UserID:   uuid.New(),
					Balance:  500_000,
					Currency: domain.WalletCurrencyRUB,
					Status:   domain.WalletStatusFrozen,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/freeze", h.AdminFreezeWallet)

	body, _ := json.Marshal(map[string]interface{}{"reason": "подозрение на фрод"})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/"+walletID.String()+"/freeze", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data := resp.Data.(map[string]interface{})
	if data["status"] != "frozen" {
		t.Errorf("expected status frozen, got %v", data["status"])
	}
}

func TestWalletHandler_AdminUnfreezeWallet(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()

	walletSvc := &mockWalletService{
		adminUnfreezeFn: func(_ context.Context, wid uuid.UUID, _ string, _ uuid.UUID) error {
			if wid == walletID {
				return nil
			}
			return domain.ErrWalletNotFound
		},
		getWalletByIDFn: func(_ context.Context, wid uuid.UUID) (*domain.Wallet, error) {
			if wid == walletID {
				return &domain.Wallet{
					ID:       walletID,
					UserID:   uuid.New(),
					Balance:  500_000,
					Currency: domain.WalletCurrencyRUB,
					Status:   domain.WalletStatusActive,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/unfreeze", h.AdminUnfreezeWallet)

	body, _ := json.Marshal(map[string]interface{}{"reason": "проверка завершена"})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/"+walletID.String()+"/unfreeze", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data := resp.Data.(map[string]interface{})
	if data["status"] != "active" {
		t.Errorf("expected status active, got %v", data["status"])
	}
}

func TestWalletHandler_AdminGetWallet(t *testing.T) {
	adminID := uuid.New()
	walletID := uuid.New()
	userID := uuid.New()

	walletSvc := &mockWalletService{
		getWalletByIDFn: func(_ context.Context, wid uuid.UUID) (*domain.Wallet, error) {
			if wid == walletID {
				return &domain.Wallet{
					ID:       walletID,
					UserID:   userID,
					Balance:  300_000,
					Currency: domain.WalletCurrencyRUB,
					Status:   domain.WalletStatusActive,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/admin/wallets/{id}", h.AdminGetWallet)

	req := httptest.NewRequest(http.MethodGet, "/admin/wallets/"+walletID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data := resp.Data.(map[string]interface{})
	if data["id"] != walletID.String() {
		t.Errorf("expected id %s, got %v", walletID.String(), data["id"])
	}
	if data["balance"].(float64) != 300_000 {
		t.Errorf("expected balance 300000, got %v", data["balance"])
	}
}

func TestWalletHandler_AdminCreditWallet_InvalidID(t *testing.T) {
	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/{id}/credit", h.AdminCreditWallet)

	body, _ := json.Marshal(map[string]interface{}{"amount": 100_000, "reason": "test"})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/not-a-uuid/credit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWalletHandler_AdminBatchCreditWallets(t *testing.T) {
	adminID := uuid.New()
	w1 := uuid.New()
	w2 := uuid.New()

	walletSvc := &mockWalletService{
		adminCreditFn: func(_ context.Context, wid uuid.UUID, amount int64, reason string, aid uuid.UUID) (*domain.WalletTransaction, error) {
			return &domain.WalletTransaction{
				ID:           uuid.New(),
				WalletID:     wid,
				Type:         domain.WalletTxAdminCredit,
				Amount:       amount,
				BalanceAfter: amount,
				Status:       domain.WalletTxStatusCompleted,
			}, nil
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/batch-credit", h.AdminBatchCreditWallets)

	body, _ := json.Marshal(batchCreditRequest{
		IDs:    []string{w1.String(), w2.String()},
		Amount: 50_000,
		Reason: "компенсация",
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/batch-credit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})
	failed := data["failed"].([]interface{})
	if len(succeeded) != 2 {
		t.Errorf("expected 2 succeeded, got %d", len(succeeded))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
}

func TestWalletHandler_AdminBatchCreditWallets_Validation(t *testing.T) {
	adminID := uuid.New()
	walletSvc := &mockWalletService{}
	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	tests := []struct {
		name string
		body batchCreditRequest
	}{
		{"empty ids", batchCreditRequest{IDs: []string{}, Amount: 1000, Reason: "test"}},
		{"zero amount", batchCreditRequest{IDs: []string{uuid.New().String()}, Amount: 0, Reason: "test"}},
		{"negative amount", batchCreditRequest{IDs: []string{uuid.New().String()}, Amount: -100, Reason: "test"}},
		{"empty reason", batchCreditRequest{IDs: []string{uuid.New().String()}, Amount: 1000, Reason: ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middleware.RequireAuth(authService))
			r.Post("/admin/wallets/batch-credit", h.AdminBatchCreditWallets)

			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/admin/wallets/batch-credit", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestWalletHandler_AdminBatchCreditWallets_PartialFailure(t *testing.T) {
	adminID := uuid.New()
	goodWallet := uuid.New()
	badWallet := uuid.New()

	walletSvc := &mockWalletService{
		adminCreditFn: func(_ context.Context, wid uuid.UUID, amount int64, _ string, _ uuid.UUID) (*domain.WalletTransaction, error) {
			if wid == goodWallet {
				return &domain.WalletTransaction{
					ID:       uuid.New(),
					WalletID: wid,
					Amount:   amount,
					Status:   domain.WalletTxStatusCompleted,
				}, nil
			}
			return nil, domain.ErrWalletNotFound
		},
	}

	h := NewWalletHandler(walletSvc, nil, &config.Config{})
	authService := &mockAuthService{userID: adminID, role: domain.RoleAdmin}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/admin/wallets/batch-credit", h.AdminBatchCreditWallets)

	body, _ := json.Marshal(batchCreditRequest{
		IDs:    []string{goodWallet.String(), badWallet.String(), "not-a-uuid"},
		Amount: 10_000,
		Reason: "тест",
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/wallets/batch-credit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})
	failed := data["failed"].([]interface{})

	if len(succeeded) != 1 {
		t.Errorf("expected 1 succeeded, got %d", len(succeeded))
	}
	if len(failed) != 2 {
		t.Errorf("expected 2 failed, got %d", len(failed))
	}
}
