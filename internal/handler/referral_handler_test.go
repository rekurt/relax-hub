package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockReferralService struct {
	generateCodeFn func(ctx context.Context, userID uuid.UUID) (string, error)
	getStatsFn     func(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error)
	getBalanceFn   func(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error)
	registerFn     func(ctx context.Context, referralCode string, newUserID uuid.UUID) error
	completeFn     func(ctx context.Context, refereeID uuid.UUID) error
	useBalanceFn   func(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
}

func (m *mockReferralService) GenerateCode(ctx context.Context, userID uuid.UUID) (string, error) {
	if m.generateCodeFn != nil {
		return m.generateCodeFn(ctx, userID)
	}
	return "abc12345", nil
}

func (m *mockReferralService) RegisterReferral(ctx context.Context, referralCode string, newUserID uuid.UUID) error {
	if m.registerFn != nil {
		return m.registerFn(ctx, referralCode, newUserID)
	}
	return nil
}

func (m *mockReferralService) CompleteReferral(ctx context.Context, refereeID uuid.UUID) error {
	if m.completeFn != nil {
		return m.completeFn(ctx, refereeID)
	}
	return nil
}

func (m *mockReferralService) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error) {
	if m.getBalanceFn != nil {
		return m.getBalanceFn(ctx, userID)
	}
	return &domain.ReferralBalance{UserID: userID, Balance: 0, TotalEarned: 0}, nil
}

func (m *mockReferralService) UseBalance(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if m.useBalanceFn != nil {
		return m.useBalanceFn(ctx, userID, amount, bookingID)
	}
	return nil
}

func (m *mockReferralService) RefundBalance(_ context.Context, _ uuid.UUID, _ int64, _ uuid.UUID) error {
	return nil
}

func (m *mockReferralService) GetStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx, userID)
	}
	return &domain.ReferralStats{}, nil
}

func TestReferralHandler_GetCode(t *testing.T) {
	userID := uuid.New()

	refSvc := &mockReferralService{
		generateCodeFn: func(ctx context.Context, uid uuid.UUID) (string, error) {
			if uid == userID {
				return "ref12345", nil
			}
			return "", domain.ErrNotFound
		},
	}

	h := NewReferralHandler(refSvc, "http://example.com")
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/referral", h.GetCode)

	req := httptest.NewRequest(http.MethodGet, "/my/referral", nil)
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

	if data["referral_code"] != "ref12345" {
		t.Errorf("expected referral_code ref12345, got %v", data["referral_code"])
	}
	if data["referral_link"] != "http://example.com/register?ref=ref12345" {
		t.Errorf("expected referral_link with code, got %v", data["referral_link"])
	}
}

func TestReferralHandler_GetCode_ServiceError(t *testing.T) {
	userID := uuid.New()

	refSvc := &mockReferralService{
		generateCodeFn: func(ctx context.Context, uid uuid.UUID) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	h := NewReferralHandler(refSvc, "")
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/referral", h.GetCode)

	req := httptest.NewRequest(http.MethodGet, "/my/referral", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestReferralHandler_GetStats(t *testing.T) {
	userID := uuid.New()

	refSvc := &mockReferralService{
		getStatsFn: func(ctx context.Context, uid uuid.UUID) (*domain.ReferralStats, error) {
			if uid == userID {
				return &domain.ReferralStats{
					TotalInvited:   5,
					TotalCompleted: 3,
					TotalEarned:    150000,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewReferralHandler(refSvc, "")
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/referral/stats", h.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/my/referral/stats", nil)
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

	if data["total_invited"].(float64) != 5 {
		t.Errorf("expected total_invited 5, got %v", data["total_invited"])
	}
	if data["total_completed"].(float64) != 3 {
		t.Errorf("expected total_completed 3, got %v", data["total_completed"])
	}
	if data["total_earned"].(float64) != 150000 {
		t.Errorf("expected total_earned 150000, got %v", data["total_earned"])
	}
}

func TestReferralHandler_GetBalance(t *testing.T) {
	userID := uuid.New()

	refSvc := &mockReferralService{
		getBalanceFn: func(ctx context.Context, uid uuid.UUID) (*domain.ReferralBalance, error) {
			if uid == userID {
				return &domain.ReferralBalance{
					UserID:      userID,
					Balance:     50000,
					TotalEarned: 100000,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewReferralHandler(refSvc, "")
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/referral/balance", h.GetBalance)

	req := httptest.NewRequest(http.MethodGet, "/my/referral/balance", nil)
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

	if data["balance"].(float64) != 50000 {
		t.Errorf("expected balance 50000, got %v", data["balance"])
	}
	if data["total_earned"].(float64) != 100000 {
		t.Errorf("expected total_earned 100000, got %v", data["total_earned"])
	}
}

func TestReferralHandler_GetCode_Unauthorized(t *testing.T) {
	refSvc := &mockReferralService{}
	h := NewReferralHandler(refSvc, "")
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/referral", h.GetCode)

	req := httptest.NewRequest(http.MethodGet, "/my/referral", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
