package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

// Mock SubscriptionService for testing
type mockSubscriptionService struct {
	subscribeFn      func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, plan domain.SubscriptionPlan) (*domain.Subscription, error)
	cancelFn         func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, subscriptionID uuid.UUID) error
	getActiveFn      func(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error)
	listByOwnerFn    func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error)
}

func (m *mockSubscriptionService) Subscribe(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, plan domain.SubscriptionPlan) (*domain.Subscription, error) {
	if m.subscribeFn != nil {
		return m.subscribeFn(ctx, userID, userRole, bathhouseID, plan)
	}
	return nil, nil
}

func (m *mockSubscriptionService) Cancel(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, subscriptionID uuid.UUID) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, userID, userRole, subscriptionID)
	}
	return nil
}

func (m *mockSubscriptionService) GetActive(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx, bathhouseID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockSubscriptionService) ListByOwner(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
	if m.listByOwnerFn != nil {
		return m.listByOwnerFn(ctx, userID, page, pageSize)
	}
	return nil, nil
}

// Mock PromotionService for testing
type mockPromotionService struct {
	createFn                func(ctx context.Context, promo *domain.Promotion) error
	getActiveBybathhouseFn  func(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error)
	recordImpressionFn      func(ctx context.Context, bathhouseID uuid.UUID) error
	recordClickFn           func(ctx context.Context, bathhouseID uuid.UUID) error
}

func (m *mockPromotionService) Create(ctx context.Context, promo *domain.Promotion) error {
	if m.createFn != nil {
		return m.createFn(ctx, promo)
	}
	return nil
}

func (m *mockPromotionService) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	if m.getActiveBybathhouseFn != nil {
		return m.getActiveBybathhouseFn(ctx, bathhouseID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockPromotionService) RecordImpression(ctx context.Context, bathhouseID uuid.UUID) error {
	if m.recordImpressionFn != nil {
		return m.recordImpressionFn(ctx, bathhouseID)
	}
	return nil
}

func (m *mockPromotionService) RecordClick(ctx context.Context, bathhouseID uuid.UUID) error {
	if m.recordClickFn != nil {
		return m.recordClickFn(ctx, bathhouseID)
	}
	return nil
}

func (m *mockPromotionService) Update(_ context.Context, _ *domain.Promotion) error { return nil }
func (m *mockPromotionService) Pause(_ context.Context, _ uuid.UUID) error         { return nil }
func (m *mockPromotionService) Resume(_ context.Context, _ uuid.UUID) error        { return nil }
func (m *mockPromotionService) GetByID(_ context.Context, _ uuid.UUID) (*domain.Promotion, error) {
	return nil, domain.ErrNotFound
}
func (m *mockPromotionService) ListByBathhouse(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Promotion], error) {
	return &domain.PaginatedResult[domain.Promotion]{}, nil
}
func (m *mockPromotionService) ListByOwner(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Promotion], error) {
	return &domain.PaginatedResult[domain.Promotion]{}, nil
}
func (m *mockPromotionService) DeductDailyBudgets(_ context.Context) error { return nil }

func TestSubscriptionHandler_Subscribe(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	subID := uuid.New()

	subSvc := &mockSubscriptionService{
		subscribeFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, bhid uuid.UUID, plan domain.SubscriptionPlan) (*domain.Subscription, error) {
			if uid == userID && bhid == bathhouseID && plan == domain.PlanPremium {
				return &domain.Subscription{
					ID:           subID,
					BathhouseID:  bhid,
					OwnerID:      userID,
					Plan:         plan,
					Status:       domain.SubscriptionActive,
					StartDate:    time.Now(),
					EndDate:      nil,
					AutoRenew:    true,
					PriceKopecks: 5000,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil
			}
			return nil, domain.ErrInvalidInput
		},
	}

	promoSvc := &mockPromotionService{}
	accessCheck := newTestAccessChecker(userID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoSvc, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/bathhouses/{id}/subscription", h.Subscribe)

	requestBody := `{"plan":"premium"}`
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/subscription", strings.NewReader(requestBody))
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

func TestSubscriptionHandler_GetSubscription(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	subID := uuid.New()
	now := time.Now()

	subSvc := &mockSubscriptionService{
		getActiveFn: func(ctx context.Context, bhid uuid.UUID) (*domain.Subscription, error) {
			if bhid == bathhouseID {
				return &domain.Subscription{
					ID:           subID,
					BathhouseID:  bathhouseID,
					OwnerID:      userID,
					Plan:         domain.PlanPremium,
					Status:       domain.SubscriptionActive,
					StartDate:    now,
					EndDate:      nil,
					AutoRenew:    true,
					PriceKopecks: 5000,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	promoSvc := &mockPromotionService{}
	accessCheck := newTestAccessChecker(userID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoSvc, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/bathhouses/{id}/subscription", h.GetSubscription)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/subscription", nil)
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

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
	userID := uuid.New()
	ownerID := uuid.New()
	subID := uuid.New()
	bathhouseID := uuid.New()
	now := time.Now()

	subSvc := &mockSubscriptionService{
		listByOwnerFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
			if uid == userID {
				return &domain.PaginatedResult[domain.Subscription]{
					Items: []domain.Subscription{
						{
							ID:           subID,
							BathhouseID:  bathhouseID,
							OwnerID:      ownerID,
							Plan:         domain.PlanPremium,
							Status:       domain.SubscriptionActive,
							StartDate:    now,
							EndDate:      nil,
							AutoRenew:    true,
							PriceKopecks: 5000,
							CreatedAt:    now,
							UpdatedAt:    now,
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

	promoSvc := &mockPromotionService{}
	accessCheck := newTestAccessChecker(ownerID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoSvc, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/subscriptions", h.ListSubscriptions)

	req := httptest.NewRequest(http.MethodGet, "/my/subscriptions?page=1&page_size=20", nil)
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

func TestSubscriptionHandler_CancelSubscription(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	subID := uuid.New()
	now := time.Now()

	subSvc := &mockSubscriptionService{
		getActiveFn: func(ctx context.Context, bhid uuid.UUID) (*domain.Subscription, error) {
			if bhid == bathhouseID {
				return &domain.Subscription{
					ID:           subID,
					BathhouseID:  bathhouseID,
					OwnerID:      userID,
					Plan:         domain.PlanPremium,
					Status:       domain.SubscriptionActive,
					StartDate:    now,
					EndDate:      nil,
					AutoRenew:    true,
					PriceKopecks: 5000,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
		cancelFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, sid uuid.UUID) error {
			if uid == userID && sid == subID {
				return nil
			}
			return domain.ErrForbidden
		},
	}

	promoSvc := &mockPromotionService{}
	accessCheck := newTestAccessChecker(userID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoSvc, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Delete("/my/bathhouses/{id}/subscription", h.CancelSubscription)

	req := httptest.NewRequest(http.MethodDelete, "/my/bathhouses/"+bathhouseID.String()+"/subscription", nil)
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

func TestSubscriptionHandler_CreatePromotion(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	promoID := uuid.New()

	promoRepo := &mockPromotionService{
		getActiveBybathhouseFn: func(ctx context.Context, bhid uuid.UUID) (*domain.Promotion, error) {
			return nil, domain.ErrNotFound
		},
		createFn: func(ctx context.Context, promo *domain.Promotion) error {
			if promo.BathhouseID == bathhouseID {
				promo.ID = promoID
				return nil
			}
			return domain.ErrInvalidInput
		},
	}

	subSvc := &mockSubscriptionService{
		getActiveFn: func(ctx context.Context, bhid uuid.UUID) (*domain.Subscription, error) {
			if bhid == bathhouseID {
				return &domain.Subscription{
					ID:           uuid.New(),
					BathhouseID:  bathhouseID,
					OwnerID:      userID,
					Plan:         domain.PlanPromoted,
					Status:       domain.SubscriptionActive,
					StartDate:    time.Now(),
					EndDate:      nil,
					AutoRenew:    true,
					PriceKopecks: 10000,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	accessCheck := newTestAccessChecker(userID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoRepo, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/my/bathhouses/{id}/promotion", h.CreatePromotion)

	requestBody := `{"budget_kopecks":100000,"duration_days":30,"daily_bid_kopecks":5000}`
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/promotion", strings.NewReader(requestBody))
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

func TestSubscriptionHandler_GetPromotion(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	promoID := uuid.New()
	now := time.Now()

	promoRepo := &mockPromotionService{
		getActiveBybathhouseFn: func(ctx context.Context, bhid uuid.UUID) (*domain.Promotion, error) {
			if bhid == bathhouseID {
				return &domain.Promotion{
					ID:              promoID,
					BathhouseID:     bathhouseID,
					BudgetKopecks:   100000,
					SpentKopecks:    0,
					StartDate:       now,
					EndDate:         now.AddDate(0, 0, 30),
					TargetCityID:    nil,
					Status:          domain.PromotionActive,
					ImpressionCount: 0,
					ClickCount:      0,
					CreatedAt:       now,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	subSvc := &mockSubscriptionService{}
	accessCheck := newTestAccessChecker(userID, bathhouseID)

	h := NewSubscriptionHandler(subSvc, promoRepo, accessCheck)
	authService := &mockAuthService{userID: userID, role: domain.RoleOwner}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/bathhouses/{id}/promotion", h.GetPromotion)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/promotion", nil)
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

// Test AccessChecker for mocking
func newTestAccessChecker(ownerID uuid.UUID, bathhouseID uuid.UUID) *service.AccessChecker {
	repRepo := &mockRepresentativeRepository{
		getByUserAndBathhouseFn: func(ctx context.Context, userID uuid.UUID, bhid uuid.UUID) (*domain.Representative, error) {
			return nil, domain.ErrNotFound
		},
	}
	bhRepo := &mockBathhouseRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == bathhouseID {
				return &domain.Bathhouse{
					ID:      id,
					OwnerID: ownerID,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	return service.NewAccessChecker(repRepo, bhRepo)
}

type mockRepresentativeRepository struct {
	createFn                    func(ctx context.Context, rep *domain.Representative) error
	getByIDFn                   func(ctx context.Context, id uuid.UUID) (*domain.Representative, error)
	deleteFn                    func(ctx context.Context, id uuid.UUID) error
	getByUserAndBathhouseFn     func(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
	listByBathhouseFn           func(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
	listByUserFn                func(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
	listBathhouseIDsByUserFn    func(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockRepresentativeRepository) Create(ctx context.Context, rep *domain.Representative) error {
	if m.createFn != nil {
		return m.createFn(ctx, rep)
	}
	return nil
}

func (m *mockRepresentativeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Representative, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepresentativeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRepresentativeRepository) GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error) {
	if m.getByUserAndBathhouseFn != nil {
		return m.getByUserAndBathhouseFn(ctx, userID, bathhouseID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepresentativeRepository) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, bathhouseID)
	}
	return nil, nil
}

func (m *mockRepresentativeRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepresentativeRepository) ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if m.listBathhouseIDsByUserFn != nil {
		return m.listBathhouseIDsByUserFn(ctx, userID)
	}
	return nil, nil
}
