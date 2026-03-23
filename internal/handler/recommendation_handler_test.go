package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/service"
)

// Mock RecommendationService for testing
type mockRecommendationService struct {
	getPersonalizedFn    func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error)
	getSimilarFn         func(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error)
	getPopularFn         func(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error)
	getUserPreferencesFn func(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	updatePreferencesFn  func(ctx context.Context, userID uuid.UUID, prefs *domain.UserPreferences) error
	recordViewFn         func(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID) error
}

func (m *mockRecommendationService) GetPersonalized(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error) {
	if m.getPersonalizedFn != nil {
		return m.getPersonalizedFn(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRecommendationService) GetSimilar(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if m.getSimilarFn != nil {
		return m.getSimilarFn(ctx, bathhouseID, limit)
	}
	return nil, nil
}

func (m *mockRecommendationService) GetPopular(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
	if m.getPopularFn != nil {
		return m.getPopularFn(ctx, cityID, limit)
	}
	return nil, nil
}

func (m *mockRecommendationService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	if m.getUserPreferencesFn != nil {
		return m.getUserPreferencesFn(ctx, userID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockRecommendationService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.UserPreferences) error {
	if m.updatePreferencesFn != nil {
		return m.updatePreferencesFn(ctx, userID, prefs)
	}
	return nil
}

func (m *mockRecommendationService) RecordView(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID) error {
	if m.recordViewFn != nil {
		return m.recordViewFn(ctx, userID, bathhouseID)
	}
	return nil
}

// mockBHService implements service.BathhouseService for handler tests in package handler.
type mockBHService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
}

func (m *mockBHService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockBHService) GetByAPIKey(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockBHService) GetBySlug(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockBHService) Create(_ context.Context, _ uuid.UUID, _ service.CreateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}

func (m *mockBHService) Update(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ service.UpdateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}

func (m *mockBHService) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBHService) Search(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockBHService) ListByOwner(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockBHService) GetWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockBHService) RegenerateWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockBHService) Approve(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockBHService) Reject(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *mockBHService) CheckCompleteness(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.CompletenessResult, error) {
	return nil, nil
}
func (m *mockBHService) SubmitForModeration(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBHService) DuplicateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Bathhouse, error) {
	return nil, nil
}
func (m *mockBHService) DeactivateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBHService) ActivateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBHService) ArchiveBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}

// mockBathhouseRepository implements repository.BathhouseRepository for AccessChecker tests.
type mockBathhouseRepository struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
}

func (m *mockBathhouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockBathhouseRepository) GetByAPIKey(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockBathhouseRepository) Create(_ context.Context, _ *domain.Bathhouse) error {
	return nil
}

func (m *mockBathhouseRepository) Update(_ context.Context, _ *domain.Bathhouse) error {
	return nil
}

func (m *mockBathhouseRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockBathhouseRepository) List(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockBathhouseRepository) ListByOwner(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockBathhouseRepository) ListIDsByOwner(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockBathhouseRepository) UpdateRating(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockBathhouseRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.BathhouseStatus) error {
	return nil
}

func (m *mockBathhouseRepository) UpdatePhotoVerified(_ context.Context, _ uuid.UUID, _ bool) error {
	return nil
}

func (m *mockBathhouseRepository) GetBySlug(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockBathhouseRepository) SlugExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockBathhouseRepository) GetCalendarToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockBathhouseRepository) SetCalendarToken(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockBathhouseRepository) GetByCalendarToken(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockBathhouseRepository) SuggestNames(_ context.Context, _ repository.SuggestionFilter) ([]string, error) {
	return nil, nil
}

func (m *mockBathhouseRepository) IncrementViewCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockBathhouseRepository) UpdateRankingFields(_ context.Context, _ uuid.UUID, _, _ float64) error {
	return nil
}

// Mock AuthService for testing
type mockAuthService struct {
	userID uuid.UUID
	role   domain.UserRole
	err    error
}

func (m *mockAuthService) ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error) {
	return m.userID, m.role, m.err
}

func (m *mockAuthService) ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	userID, role, err := m.ParseToken(ctx, token)
	return userID, role, uuid.Nil, err
}

func TestRecommendationHandler_GetPersonalized(t *testing.T) {
	userID := uuid.New()
	bathID1 := uuid.New()
	bathID2 := uuid.New()

	recSvc := &mockRecommendationService{
		getPersonalizedFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error) {
			if uid == userID {
				return []uuid.UUID{bathID1, bathID2}, 2, nil
			}
			return nil, 0, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Test Bath", Address: "123 Main St", CityID: 1,
				Latitude: 10.0, Longitude: 20.0, PricePerHour: 5000, Rating: 4.5, ReviewCount: 10,
			}, nil
		},
	}

	h := NewRecommendationHandler(recSvc, bhSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/recommendations", h.GetPersonalized)

	req := httptest.NewRequest(http.MethodGet, "/recommendations?page=1&page_size=20", nil)
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

func TestRecommendationHandler_GetSimilar(t *testing.T) {
	bathhouseID := uuid.New()
	similarID1 := uuid.New()
	similarID2 := uuid.New()

	recSvc := &mockRecommendationService{
		getSimilarFn: func(ctx context.Context, bid uuid.UUID, limit int) ([]uuid.UUID, error) {
			if bid == bathhouseID {
				return []uuid.UUID{similarID1, similarID2}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Similar Bath", Address: "456 Oak St", CityID: 1,
				Latitude: 11.0, Longitude: 21.0, PricePerHour: 4500, Rating: 4.3, ReviewCount: 8,
			}, nil
		},
	}

	h := NewRecommendationHandler(recSvc, bhSvc)

	r := chi.NewRouter()
	r.Get("/bathhouses/{id}/similar", h.GetSimilar)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bathhouseID.String()+"/similar", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestRecommendationHandler_GetPopular(t *testing.T) {
	bathID1 := uuid.New()
	bathID2 := uuid.New()

	recSvc := &mockRecommendationService{
		getPopularFn: func(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
			if cityID == 1 {
				return []uuid.UUID{bathID1, bathID2}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Popular Bath", Address: "789 Elm St", CityID: 1,
				Latitude: 12.0, Longitude: 22.0, PricePerHour: 6000, Rating: 4.7, ReviewCount: 50,
			}, nil
		},
	}

	h := NewRecommendationHandler(recSvc, bhSvc)

	r := chi.NewRouter()
	r.Get("/popular", h.GetPopular)

	req := httptest.NewRequest(http.MethodGet, "/popular?city_id=1&limit=10", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRecommendationHandler_GetPopularMissingCityID(t *testing.T) {
	recSvc := &mockRecommendationService{}
	bhSvc := &mockBHService{}
	h := NewRecommendationHandler(recSvc, bhSvc)

	r := chi.NewRouter()
	r.Get("/popular", h.GetPopular)

	req := httptest.NewRequest(http.MethodGet, "/popular", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestRecommendationHandler_UpdatePreferences(t *testing.T) {
	userID := uuid.New()

	recSvc := &mockRecommendationService{
		getUserPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*domain.UserPreferences, error) {
			if uid == userID {
				return &domain.UserPreferences{
					UserID:          uid,
					PreferredCityID: nil,
					PriceRangeMin:   nil,
					PriceRangeMax:   nil,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
		updatePreferencesFn: func(ctx context.Context, uid uuid.UUID, prefs *domain.UserPreferences) error {
			return nil
		},
	}

	bhSvc := &mockBHService{}
	h := NewRecommendationHandler(recSvc, bhSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Put("/preferences", h.UpdatePreferences)

	requestBody := `{"prefer_pool": true, "prefer_sauna": false}`
	req := httptest.NewRequest(http.MethodPut, "/preferences", strings.NewReader(requestBody))
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

func TestRecommendationHandler_GetPreferences(t *testing.T) {
	userID := uuid.New()
	cityID := int64(1)
	priceMin := int64(1000)
	priceMax := int64(5000)

	recSvc := &mockRecommendationService{
		getUserPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*domain.UserPreferences, error) {
			if uid == userID {
				return &domain.UserPreferences{
					UserID:          uid,
					PreferredCityID: &cityID,
					PriceRangeMin:   &priceMin,
					PriceRangeMax:   &priceMax,
					PreferPool:      true,
					PreferSauna:     false,
					PreferSteamRoom: true,
					PreferHotTub:    false,
					PreferBBQ:       true,
					PreferKaraoke:   false,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{}
	h := NewRecommendationHandler(recSvc, bhSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/preferences", h.GetPreferences)

	req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
		return
	}

	var apiResp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Extract the userPreferencesResponse from Data
	var resp userPreferencesResponse
	respBytes, _ := json.Marshal(apiResp.Data)
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal preferences: %v", err)
	}

	if resp.PreferredCityID == nil || *resp.PreferredCityID != cityID {
		t.Errorf("expected city id %d, got %v", cityID, resp.PreferredCityID)
	}
	if !resp.PreferPool {
		t.Error("expected prefer_pool to be true")
	}
}

func TestRecommendationHandler_GetPreferencesNotFound(t *testing.T) {
	userID := uuid.New()

	recSvc := &mockRecommendationService{
		getUserPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*domain.UserPreferences, error) {
			return nil, domain.ErrNotFound
		},
	}

	bhSvc := &mockBHService{}
	h := NewRecommendationHandler(recSvc, bhSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/preferences", h.GetPreferences)

	req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
