package handler

import (
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
	"github.com/nikitaaldaev/bani/internal/service"
)

type mockLoyaltyService struct {
	getAccountFn       func(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error)
	earnPointsFn       func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, totalPrice int64) (int64, error)
	spendPointsFn      func(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	refundPointsFn     func(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	getDiscountFn      func(ctx context.Context, userID uuid.UUID) (int, error)
	recalculateLevelFn func(ctx context.Context, userID uuid.UUID) (*service.LevelChangeResult, error)
	listTransactionsFn func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error)
}

func (m *mockLoyaltyService) GetAccount(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(ctx, userID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLoyaltyService) EarnPoints(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, totalPrice int64) (int64, error) {
	if m.earnPointsFn != nil {
		return m.earnPointsFn(ctx, userID, bookingID, totalPrice)
	}
	return 0, nil
}

func (m *mockLoyaltyService) SpendPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if m.spendPointsFn != nil {
		return m.spendPointsFn(ctx, userID, amount, bookingID)
	}
	return nil
}

func (m *mockLoyaltyService) RefundPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if m.refundPointsFn != nil {
		return m.refundPointsFn(ctx, userID, amount, bookingID)
	}
	return nil
}

func (m *mockLoyaltyService) GetDiscount(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.getDiscountFn != nil {
		return m.getDiscountFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockLoyaltyService) RecalculateLevel(ctx context.Context, userID uuid.UUID) (*service.LevelChangeResult, error) {
	if m.recalculateLevelFn != nil {
		return m.recalculateLevelFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockLoyaltyService) ListTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
	if m.listTransactionsFn != nil {
		return m.listTransactionsFn(ctx, userID, page, pageSize)
	}
	return nil, nil
}

func TestLoyaltyHandler_GetAccount(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	loyaltySvc := &mockLoyaltyService{
		getAccountFn: func(ctx context.Context, uid uuid.UUID) (*domain.LoyaltyAccount, error) {
			if uid == userID {
				return &domain.LoyaltyAccount{
					UserID:      userID,
					Level:       domain.LoyaltySilver,
					Points:      1500,
					TotalEarned: 3000,
					TotalSpent:  1500,
					VisitCount:  7,
					UpdatedAt:   now,
					CreatedAt:   now,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewLoyaltyHandler(loyaltySvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty", h.GetAccount)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty", nil)
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

	if data["level"] != "silver" {
		t.Errorf("expected level silver, got %v", data["level"])
	}
	if data["points"].(float64) != 1500 {
		t.Errorf("expected points 1500, got %v", data["points"])
	}

	privileges, ok := data["privileges"].(map[string]interface{})
	if !ok {
		t.Fatal("expected privileges to be a map")
	}
	if privileges["discount_percent"].(float64) != 3 {
		t.Errorf("expected discount 3, got %v", privileges["discount_percent"])
	}
	if privileges["next_level"] != "gold" {
		t.Errorf("expected next_level gold, got %v", privileges["next_level"])
	}
}

func TestLoyaltyHandler_GetAccount_AutoCreate(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	loyaltySvc := &mockLoyaltyService{
		getAccountFn: func(ctx context.Context, uid uuid.UUID) (*domain.LoyaltyAccount, error) {
			return &domain.LoyaltyAccount{
				UserID:      uid,
				Level:       domain.LoyaltyBronze,
				Points:      0,
				TotalEarned: 0,
				TotalSpent:  0,
				VisitCount:  0,
				UpdatedAt:   now,
				CreatedAt:   now,
			}, nil
		},
	}

	h := NewLoyaltyHandler(loyaltySvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty", h.GetAccount)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["level"] != "bronze" {
		t.Errorf("expected level bronze, got %v", data["level"])
	}
}

func TestLoyaltyHandler_ListTransactions(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	loyaltySvc := &mockLoyaltyService{
		listTransactionsFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
			if uid == userID {
				return &domain.PaginatedResult[domain.LoyaltyTransaction]{
					Items: []domain.LoyaltyTransaction{
						{
							ID:          uuid.New(),
							UserID:      userID,
							Type:        domain.LoyaltyTransactionEarn,
							Amount:      50,
							BookingID:   &bookingID,
							Description: "Начисление за бронирование",
							CreatedAt:   now,
						},
					},
					Page:       1,
					PageSize:   20,
					TotalCount: 1,
					TotalPages: 1,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewLoyaltyHandler(loyaltySvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty/transactions", h.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty/transactions?page=1&page_size=20", nil)
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

func TestLoyaltyHandler_ListTransactions_ServiceError(t *testing.T) {
	userID := uuid.New()

	loyaltySvc := &mockLoyaltyService{
		listTransactionsFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
			return nil, domain.ErrNotFound
		},
	}

	h := NewLoyaltyHandler(loyaltySvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty/transactions", h.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty/transactions", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestLoyaltyHandler_GetLevels(t *testing.T) {
	loyaltySvc := &mockLoyaltyService{}
	h := NewLoyaltyHandler(loyaltySvc)

	// GetLevels is a public-like endpoint, but routed under auth.
	// We still need auth middleware in the router, but since it just reads domain constants,
	// let's test it with auth.
	userID := uuid.New()
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty/levels", h.GetLevels)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty/levels", nil)
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

	levels, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}

	if len(levels) != 4 {
		t.Errorf("expected 4 levels, got %d", len(levels))
		return
	}

	// Verify order: bronze first, platinum last
	first := levels[0].(map[string]interface{})
	if first["level"] != "bronze" {
		t.Errorf("expected first level to be bronze, got %v", first["level"])
	}

	last := levels[3].(map[string]interface{})
	if last["level"] != "platinum" {
		t.Errorf("expected last level to be platinum, got %v", last["level"])
	}
}

func TestLoyaltyHandler_GetAccount_Unauthorized(t *testing.T) {
	loyaltySvc := &mockLoyaltyService{}
	h := NewLoyaltyHandler(loyaltySvc)
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/loyalty", h.GetAccount)

	req := httptest.NewRequest(http.MethodGet, "/my/loyalty", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
