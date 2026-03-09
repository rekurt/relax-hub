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

type mockPromoService struct {
	createFn          func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error)
	validateFn        func(ctx context.Context, code string, bathhouseID uuid.UUID, amount int64) (*domain.PromoCode, int64, error)
	applyFn           func(ctx context.Context, userID uuid.UUID, code string, bookingID uuid.UUID, amount int64) (int64, error)
	deactivateFn      func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promoID uuid.UUID) error
	listByBathhouseFn func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error)
}

func (m *mockPromoService) Create(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, userRole, promo)
	}
	return nil, nil
}

func (m *mockPromoService) Validate(ctx context.Context, code string, bathhouseID uuid.UUID, amount int64) (*domain.PromoCode, int64, error) {
	if m.validateFn != nil {
		return m.validateFn(ctx, code, bathhouseID, amount)
	}
	return nil, 0, nil
}

func (m *mockPromoService) Apply(ctx context.Context, userID uuid.UUID, code string, bookingID uuid.UUID, amount int64) (int64, error) {
	if m.applyFn != nil {
		return m.applyFn(ctx, userID, code, bookingID, amount)
	}
	return 0, nil
}

func (m *mockPromoService) Deactivate(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promoID uuid.UUID) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, userID, userRole, promoID)
	}
	return nil
}

func (m *mockPromoService) ListByBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, userID, userRole, bathhouseID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.PromoCode]{}, nil
}

func TestPromoHandler_CreateForBathhouse(t *testing.T) {
	bathhouseID := uuid.New()
	userID := uuid.New()
	promoID := uuid.New()
	now := time.Now()

	svc := &mockPromoService{
		createFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error) {
			return &domain.PromoCode{
				ID:          promoID,
				Code:        "SUMMER20",
				Type:        domain.PromoTypePercentage,
				Value:       20,
				BathhouseID: &bathhouseID,
				CreatorID:   uid,
				MaxUses:     100,
				MinAmount:   500000,
				ValidFrom:   now,
				ValidUntil:  now.Add(30 * 24 * time.Hour),
				IsActive:    true,
				CreatedAt:   now,
			}, nil
		},
	}

	h := NewPromoHandler(svc)

	body, _ := json.Marshal(createPromoRequest{
		Code:       "SUMMER20",
		Type:       "percentage",
		Value:      20,
		MaxUses:    100,
		MinAmount:  500000,
		ValidFrom:  now,
		ValidUntil: now.Add(30 * 24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/promo-codes", bytes.NewReader(body))
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bathhouseID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.CreateForBathhouse(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success true")
	}
}

func TestPromoHandler_CreateForBathhouse_InvalidID(t *testing.T) {
	h := NewPromoHandler(&mockPromoService{})

	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/invalid/promo-codes", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.CreateForBathhouse(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPromoHandler_CreateForBathhouse_ServiceError(t *testing.T) {
	bathhouseID := uuid.New()
	userID := uuid.New()

	svc := &mockPromoService{
		createFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := NewPromoHandler(svc)

	body, _ := json.Marshal(createPromoRequest{
		Code:       "TEST",
		Type:       "percentage",
		Value:      10,
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/promo-codes", bytes.NewReader(body))
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bathhouseID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.CreateForBathhouse(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestPromoHandler_ListByBathhouse(t *testing.T) {
	bathhouseID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	svc := &mockPromoService{
		listByBathhouseFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, bhID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
			return &domain.PaginatedResult[domain.PromoCode]{
				Items: []domain.PromoCode{
					{
						ID:          uuid.New(),
						Code:        "PROMO1",
						Type:        domain.PromoTypePercentage,
						Value:       10,
						BathhouseID: &bathhouseID,
						CreatorID:   uid,
						IsActive:    true,
						ValidFrom:   now,
						ValidUntil:  now.Add(24 * time.Hour),
						CreatedAt:   now,
					},
				},
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewPromoHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/promo-codes", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bathhouseID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.ListByBathhouse(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success true")
	}
	if resp.Meta == nil {
		t.Error("expected meta to be present")
	}
}

func TestPromoHandler_Deactivate(t *testing.T) {
	promoID := uuid.New()
	userID := uuid.New()

	svc := &mockPromoService{
		deactivateFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, pid uuid.UUID) error {
			if pid != promoID {
				t.Errorf("expected promo ID %s, got %s", promoID, pid)
			}
			return nil
		},
	}

	h := NewPromoHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/promo-codes/"+promoID.String(), nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", promoID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.Deactivate(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPromoHandler_Deactivate_NotFound(t *testing.T) {
	svc := &mockPromoService{
		deactivateFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, pid uuid.UUID) error {
			return domain.ErrPromoNotFound
		},
	}

	h := NewPromoHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/promo-codes/"+uuid.New().String(), nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", uuid.New().String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.Deactivate(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestPromoHandler_Validate(t *testing.T) {
	bathhouseID := uuid.New()
	now := time.Now()

	svc := &mockPromoService{
		validateFn: func(ctx context.Context, code string, bhID uuid.UUID, amount int64) (*domain.PromoCode, int64, error) {
			return &domain.PromoCode{
				Code:       "SUMMER20",
				Type:       domain.PromoTypePercentage,
				Value:      20,
				IsActive:   true,
				ValidFrom:  now,
				ValidUntil: now.Add(24 * time.Hour),
			}, 200000, nil
		},
	}

	h := NewPromoHandler(svc)

	body, _ := json.Marshal(validatePromoRequest{
		Code:        "SUMMER20",
		BathhouseID: bathhouseID.String(),
		Amount:      1000000,
	})

	req := httptest.NewRequest(http.MethodPost, "/promo-codes/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Validate(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success true")
	}
}

func TestPromoHandler_Validate_Expired(t *testing.T) {
	svc := &mockPromoService{
		validateFn: func(ctx context.Context, code string, bhID uuid.UUID, amount int64) (*domain.PromoCode, int64, error) {
			return nil, 0, domain.ErrPromoExpired
		},
	}

	h := NewPromoHandler(svc)

	body, _ := json.Marshal(validatePromoRequest{
		Code:        "EXPIRED",
		BathhouseID: uuid.New().String(),
		Amount:      1000000,
	})

	req := httptest.NewRequest(http.MethodPost, "/promo-codes/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Validate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPromoHandler_Validate_InvalidBathhouseID(t *testing.T) {
	h := NewPromoHandler(&mockPromoService{})

	body, _ := json.Marshal(validatePromoRequest{
		Code:        "TEST",
		BathhouseID: "invalid",
		Amount:      1000000,
	})

	req := httptest.NewRequest(http.MethodPost, "/promo-codes/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Validate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPromoHandler_CreateGlobal(t *testing.T) {
	userID := uuid.New()
	promoID := uuid.New()
	now := time.Now()

	svc := &mockPromoService{
		createFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error) {
			if promo.BathhouseID != nil {
				t.Error("expected nil bathhouse ID for global promo")
			}
			return &domain.PromoCode{
				ID:         promoID,
				Code:       "GLOBAL10",
				Type:       domain.PromoTypeFixedAmount,
				Value:      100000,
				CreatorID:  uid,
				MaxUses:    1000,
				IsActive:   true,
				ValidFrom:  now,
				ValidUntil: now.Add(30 * 24 * time.Hour),
				CreatedAt:  now,
			}, nil
		},
	}

	h := NewPromoHandler(svc)

	body, _ := json.Marshal(createPromoRequest{
		Code:       "GLOBAL10",
		Type:       "fixed_amount",
		Value:      100000,
		MaxUses:    1000,
		ValidFrom:  now,
		ValidUntil: now.Add(30 * 24 * time.Hour),
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/promo-codes", bytes.NewReader(body))
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.CreateGlobal(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}
