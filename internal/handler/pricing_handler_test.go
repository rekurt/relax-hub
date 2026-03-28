package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

// Mock PricingService for testing
type mockPricingService struct {
	calculatePriceFn func(ctx context.Context, bathhouseID uuid.UUID, basePrice int64, startTime, endTime time.Time) (int64, error)
	createRuleFn     func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error)
	updateRuleFn     func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error
	deleteRuleFn     func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, ruleID uuid.UUID) error
	listRulesFn      func(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	getActiveRulesFn func(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
}

func (m *mockPricingService) CalculatePrice(ctx context.Context, bathhouseID uuid.UUID, basePrice int64, startTime, endTime time.Time) (int64, error) {
	if m.calculatePriceFn != nil {
		return m.calculatePriceFn(ctx, bathhouseID, basePrice, startTime, endTime)
	}
	return basePrice * int64(endTime.Sub(startTime).Hours()), nil
}

func (m *mockPricingService) CalculateFullPrice(ctx context.Context, input service.PriceCalculationInput) (int64, *service.PriceBreakdown, error) {
	price, err := m.CalculatePrice(ctx, input.BathhouseID, input.BasePrice, input.StartTime, input.EndTime)
	return price, &service.PriceBreakdown{BasePrice: price}, err
}

func (m *mockPricingService) CreateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error) {
	if m.createRuleFn != nil {
		return m.createRuleFn(ctx, userID, userRole, rule)
	}
	return rule, nil
}

func (m *mockPricingService) UpdateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error {
	if m.updateRuleFn != nil {
		return m.updateRuleFn(ctx, userID, userRole, rule)
	}
	return nil
}

func (m *mockPricingService) DeleteRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, ruleID uuid.UUID) error {
	if m.deleteRuleFn != nil {
		return m.deleteRuleFn(ctx, userID, userRole, ruleID)
	}
	return nil
}

func (m *mockPricingService) ListRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	if m.listRulesFn != nil {
		return m.listRulesFn(ctx, bathhouseID)
	}
	return nil, nil
}

func (m *mockPricingService) GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	if m.getActiveRulesFn != nil {
		return m.getActiveRulesFn(ctx, bathhouseID)
	}
	return nil, nil
}

func (m *mockPricingService) CreateSeasonalTariff(_ context.Context, _ uuid.UUID, _ domain.UserRole, tariff *domain.SeasonalTariff) (*domain.SeasonalTariff, error) {
	return tariff, nil
}

func (m *mockPricingService) UpdateSeasonalTariff(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ *domain.SeasonalTariff) error {
	return nil
}

func (m *mockPricingService) DeleteSeasonalTariff(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}

func (m *mockPricingService) ListSeasonalTariffs(_ context.Context, _ uuid.UUID) ([]domain.SeasonalTariff, error) {
	return nil, nil
}

type mockSmartPricingService struct {
	getRecommendationFn func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*service.PriceRecommendation, error)
}

func (m *mockSmartPricingService) GetRecommendation(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*service.PriceRecommendation, error) {
	if m.getRecommendationFn != nil {
		return m.getRecommendationFn(ctx, userID, userRole, bathhouseID)
	}
	return &service.PriceRecommendation{
		CurrentPrice:     200000,
		RecommendedPrice: 220000,
		Coefficient:      1.1,
		AvgAreaPrice:     180000,
		OccupancyRate:    0.65,
		DemandTrend:      "stable",
	}, nil
}

func TestPricingHandler_CreateRule(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	ruleID := uuid.New()

	pricingSvc := &mockPricingService{
		createRuleFn: func(ctx context.Context, uid uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error) {
			if uid == userID && rule.BathhouseID == bathhouseID {
				rule.ID = ruleID
				rule.CreatedAt = time.Now()
				return rule, nil
			}
			return nil, domain.ErrInvalidInput
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == bathhouseID {
				return &domain.Bathhouse{ID: id, OwnerID: userID}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/bathhouses/{id}/pricing-rules", h.CreateRule)

	requestBody := `{"name":"Happy Hour","type":"time_range","multiplier":0.8,"time_from":"08:00","time_to":"12:00","priority":1,"is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/pricing-rules", strings.NewReader(requestBody))
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestPricingHandler_ListRules(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()

	now := time.Now()
	rules := []domain.PricingRule{
		{
			ID:          uuid.New(),
			BathhouseID: bathhouseID,
			Name:        "Weekends",
			Type:        domain.RuleTypeWeekend,
			Multiplier:  1.5,
			DaysOfWeek:  []int{5, 6},
			Priority:    1,
			IsActive:    true,
			CreatedAt:   now,
		},
	}

	pricingSvc := &mockPricingService{
		listRulesFn: func(ctx context.Context, bhid uuid.UUID) ([]domain.PricingRule, error) {
			if bhid == bathhouseID {
				return rules, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == bathhouseID {
				return &domain.Bathhouse{ID: id, OwnerID: userID}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/bathhouses/{id}/pricing-rules", h.ListRules)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/pricing-rules", nil)
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

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestPricingHandler_UpdateRule(t *testing.T) {
	userID := uuid.New()
	ruleID := uuid.New()

	pricingSvc := &mockPricingService{
		updateRuleFn: func(ctx context.Context, uid uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error {
			if uid == userID && rule.ID == ruleID {
				return nil
			}
			return domain.ErrInvalidInput
		},
	}

	bhSvc := &mockBHService{}
	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Put("/pricing-rules/{id}", h.UpdateRule)

	requestBody := `{"name":"Happy Hour Updated","type":"time_range","multiplier":0.9,"time_from":"08:00","time_to":"12:00","priority":2,"is_active":true}`
	req := httptest.NewRequest(http.MethodPut, "/pricing-rules/"+ruleID.String(), strings.NewReader(requestBody))
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

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestPricingHandler_DeleteRule(t *testing.T) {
	userID := uuid.New()
	ruleID := uuid.New()

	pricingSvc := &mockPricingService{
		deleteRuleFn: func(ctx context.Context, uid uuid.UUID, userRole domain.UserRole, rid uuid.UUID) error {
			if uid == userID && rid == ruleID {
				return nil
			}
			return domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{}
	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Delete("/pricing-rules/{id}", h.DeleteRule)

	req := httptest.NewRequest(http.MethodDelete, "/pricing-rules/"+ruleID.String(), nil)
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

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestPricingHandler_CalculatePrice(t *testing.T) {
	bathhouseID := uuid.New()
	startTime := time.Now().Round(time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	pricingSvc := &mockPricingService{
		calculatePriceFn: func(ctx context.Context, bhid uuid.UUID, basePrice int64, st, et time.Time) (int64, error) {
			if bhid == bathhouseID {
				hours := int64(et.Sub(st).Hours())
				return basePrice * hours, nil
			}
			return 0, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == bathhouseID {
				return &domain.Bathhouse{ID: id, PricePerHour: 1000}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})

	r := chi.NewRouter()
	r.Get("/bathhouses/{id}/price-calculator", h.CalculatePrice)

	q := url.Values{}
	q.Set("start", startTime.Format(time.RFC3339))
	q.Set("end", endTime.Format(time.RFC3339))
	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bathhouseID.String()+"/price-calculator?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestPricingHandler_CalculatePrice_MissingParams(t *testing.T) {
	bathhouseID := uuid.New()

	pricingSvc := &mockPricingService{}
	bhSvc := &mockBHService{}
	h := NewPricingHandler(pricingSvc, bhSvc, &mockSmartPricingService{})

	r := chi.NewRouter()
	r.Get("/bathhouses/{id}/price-calculator", h.CalculatePrice)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bathhouseID.String()+"/price-calculator", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success=false for missing params")
	}
}
