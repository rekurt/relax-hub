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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockPayoutService struct {
	requestPayoutFn          func(ctx context.Context, userID uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error)
	getPayoutHistoryFn       func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error)
	setAutoPayoutThresholdFn func(ctx context.Context, userID uuid.UUID, threshold int64) error
	getAutoPayoutSettingsFn  func(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error)
	calculateAvailableFn     func(ctx context.Context, userID uuid.UUID) (int64, error)
	processPayoutFn          func(ctx context.Context, payoutID uuid.UUID) error
}

func (m *mockPayoutService) RequestPayout(ctx context.Context, userID uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
	if m.requestPayoutFn != nil {
		return m.requestPayoutFn(ctx, userID, amount, bankDetails)
	}
	return nil, nil
}

func (m *mockPayoutService) GetPayoutHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
	if m.getPayoutHistoryFn != nil {
		return m.getPayoutHistoryFn(ctx, userID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Payout]{}, nil
}

func (m *mockPayoutService) SetAutoPayoutThreshold(ctx context.Context, userID uuid.UUID, threshold int64) error {
	if m.setAutoPayoutThresholdFn != nil {
		return m.setAutoPayoutThresholdFn(ctx, userID, threshold)
	}
	return nil
}

func (m *mockPayoutService) GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error) {
	if m.getAutoPayoutSettingsFn != nil {
		return m.getAutoPayoutSettingsFn(ctx, userID)
	}
	return &domain.AutoPayoutSettings{UserID: userID, Threshold: 0, UpdatedAt: time.Now()}, nil
}

func (m *mockPayoutService) CalculateAvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.calculateAvailableFn != nil {
		return m.calculateAvailableFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockPayoutService) ProcessPayout(ctx context.Context, payoutID uuid.UUID) error {
	if m.processPayoutFn != nil {
		return m.processPayoutFn(ctx, payoutID)
	}
	return nil
}

func TestPayoutHandler_RequestPayout_Success(t *testing.T) {
	userID := uuid.New()
	payoutID := uuid.New()
	now := time.Now()

	svc := &mockPayoutService{
		requestPayoutFn: func(ctx context.Context, uid uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
			return &domain.Payout{
				ID:          payoutID,
				UserID:      uid,
				Amount:      amount,
				Status:      domain.PayoutStatusPending,
				BankDetails: bankDetails,
				RequestedAt: now,
				CreatedAt:   now,
			}, nil
		},
	}

	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	body, _ := json.Marshal(requestPayoutRequest{
		Amount:      500_000,
		BankDetails: json.RawMessage(`{"bik":"044525225"}`),
	})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
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
	if data["id"] != payoutID.String() {
		t.Errorf("expected id %s, got %v", payoutID.String(), data["id"])
	}
	if data["status"] != "pending" {
		t.Errorf("expected status pending, got %v", data["status"])
	}
	if data["amount"].(float64) != 500_000 {
		t.Errorf("expected amount 500000, got %v", data["amount"])
	}
}

func TestPayoutHandler_RequestPayout_InvalidBody(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader([]byte("not json")))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPayoutHandler_RequestPayout_ZeroAmount(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	body, _ := json.Marshal(requestPayoutRequest{Amount: 0})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPayoutHandler_RequestPayout_BelowMinimum(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{
		requestPayoutFn: func(ctx context.Context, uid uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
			return nil, domain.ErrPayoutBelowMinimum
		},
	}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	body, _ := json.Marshal(requestPayoutRequest{Amount: 10_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPayoutHandler_SetAutoPayoutThreshold_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := &mockPayoutService{
		setAutoPayoutThresholdFn: func(ctx context.Context, uid uuid.UUID, threshold int64) error {
			return nil
		},
		getAutoPayoutSettingsFn: func(ctx context.Context, uid uuid.UUID) (*domain.AutoPayoutSettings, error) {
			return &domain.AutoPayoutSettings{
				UserID:    uid,
				Threshold: 500_000,
				UpdatedAt: now,
			}, nil
		},
	}

	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Put("/my/wallet/auto-payout", h.SetAutoPayoutThreshold)

	body, _ := json.Marshal(autoPayoutRequest{Threshold: 500_000})
	req := httptest.NewRequest(http.MethodPut, "/my/wallet/auto-payout", bytes.NewReader(body))
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
	if data["threshold"].(float64) != 500_000 {
		t.Errorf("expected threshold 500000, got %v", data["threshold"])
	}
}

func TestPayoutHandler_SetAutoPayoutThreshold_InvalidBody(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Put("/my/wallet/auto-payout", h.SetAutoPayoutThreshold)

	req := httptest.NewRequest(http.MethodPut, "/my/wallet/auto-payout", bytes.NewReader([]byte("bad")))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPayoutHandler_ListPayouts_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	payoutID := uuid.New()

	svc := &mockPayoutService{
		getPayoutHistoryFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
			return &domain.PaginatedResult[domain.Payout]{
				Items: []domain.Payout{
					{
						ID:          payoutID,
						UserID:      uid,
						Amount:      500_000,
						Status:      domain.PayoutStatusCompleted,
						RequestedAt: now,
						CreatedAt:   now,
					},
				},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/payouts", h.ListPayouts)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/payouts?page=1&page_size=20", nil)
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
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta")
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count 1, got %d", resp.Meta.TotalCount)
	}

	items, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be array")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 payout, got %d", len(items))
	}
}

func TestPayoutHandler_ListPayouts_Empty(t *testing.T) {
	userID := uuid.New()

	svc := &mockPayoutService{
		getPayoutHistoryFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
			return &domain.PaginatedResult[domain.Payout]{
				Items:      nil,
				Page:       1,
				PageSize:   20,
				TotalCount: 0,
				TotalPages: 0,
			}, nil
		},
	}

	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/wallet/payouts", h.ListPayouts)

	req := httptest.NewRequest(http.MethodGet, "/my/wallet/payouts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestPayoutHandler_RequestPayout_InsufficientBalance(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{
		requestPayoutFn: func(ctx context.Context, uid uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
			return nil, domain.ErrInsufficientWalletBalance
		},
	}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	body, _ := json.Marshal(requestPayoutRequest{Amount: 500_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPayoutHandler_RequestPayout_DailyLimitExceeded(t *testing.T) {
	userID := uuid.New()
	svc := &mockPayoutService{
		requestPayoutFn: func(ctx context.Context, uid uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
			return nil, domain.ErrPayoutDailyLimitExceeded
		},
	}
	h := NewPayoutHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/wallet/payout", h.RequestPayout)

	body, _ := json.Marshal(requestPayoutRequest{Amount: 500_000})
	req := httptest.NewRequest(http.MethodPost, "/my/wallet/payout", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
