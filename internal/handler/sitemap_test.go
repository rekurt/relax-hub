package handler

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/seo"
	"github.com/nikitaaldaev/bani/internal/service"
)

// sitemapMockBHService implements service.BathhouseService for sitemap tests.
type sitemapMockBHService struct {
	searchFn  func(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
}

func (m *sitemapMockBHService) Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.Bathhouse]{}, nil
}
func (m *sitemapMockBHService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *sitemapMockBHService) GetBySlug(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, nil
}
func (m *sitemapMockBHService) Create(_ context.Context, _ uuid.UUID, _ service.CreateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}
func (m *sitemapMockBHService) Update(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ service.UpdateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}
func (m *sitemapMockBHService) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *sitemapMockBHService) ListByOwner(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}
func (m *sitemapMockBHService) GetWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}
func (m *sitemapMockBHService) RegenerateWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}
func (m *sitemapMockBHService) Approve(_ context.Context, _ uuid.UUID) error { return nil }
func (m *sitemapMockBHService) Reject(_ context.Context, _ uuid.UUID) error  { return nil }

// mockCityRepo implements repository.CityRepository for sitemap tests.
type mockCityRepo struct {
	getAllFn   func(ctx context.Context) ([]domain.City, error)
	getByIDFn func(ctx context.Context, id int64) (*domain.City, error)
}

func (m *mockCityRepo) Create(_ context.Context, _ *domain.City) error { return nil }
func (m *mockCityRepo) GetAll(ctx context.Context) ([]domain.City, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}
func (m *mockCityRepo) GetBySlug(_ context.Context, _ string) (*domain.City, error) {
	return nil, nil
}
func (m *mockCityRepo) GetByID(ctx context.Context, id int64) (*domain.City, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *mockCityRepo) Update(_ context.Context, _ *domain.City) error { return nil }
func (m *mockCityRepo) Delete(_ context.Context, _ int64) error        { return nil }

func TestSitemapHandler_GetSchema(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	bhID := uuid.New()
	bh := &domain.Bathhouse{
		ID:           bhID,
		Name:         "Тест баня",
		Slug:         "test-banya",
		Description:  "Описание",
		Address:      "ул. Тестовая, 1",
		CityID:       1,
		Latitude:     55.75,
		Longitude:    37.62,
		PricePerHour: 200000,
		Rating:       4.5,
		ReviewCount:  10,
		Images:       []string{"https://example.com/img.jpg"},
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		Status:       domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "23:00"},
		},
	}

	svc := &sitemapMockBHService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Bathhouse, error) {
			return bh, nil
		},
	}
	cityRepo := &mockCityRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.City, error) {
			return &domain.City{ID: 1, Name: "Москва", Slug: "moscow"}, nil
		},
	}

	h := NewSitemapHandler(svc, cityRepo, nil, log, "https://bani.ru")

	r := chi.NewRouter()
	r.Get("/api/v1/bathhouses/{id}/schema", h.GetSchema)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses/"+bhID.String()+"/schema", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/ld+json" {
		t.Errorf("expected Content-Type application/ld+json, got %s", contentType)
	}

	var schema seo.SchemaLocalBusiness
	if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
		t.Fatalf("failed to decode schema: %v", err)
	}

	if schema.Context != "https://schema.org" {
		t.Errorf("expected context https://schema.org, got %s", schema.Context)
	}
	if schema.Type != "LocalBusiness" {
		t.Errorf("expected type LocalBusiness, got %s", schema.Type)
	}
	if schema.Name != "Тест баня" {
		t.Errorf("expected name 'Тест баня', got %s", schema.Name)
	}
	if schema.URL != "https://bani.ru/moscow/test-banya" {
		t.Errorf("expected URL with city, got %s", schema.URL)
	}
	if schema.AggregateRating == nil {
		t.Fatal("expected non-nil aggregateRating")
	}
	if schema.AggregateRating.RatingValue != "4.5" {
		t.Errorf("expected ratingValue 4.5, got %s", schema.AggregateRating.RatingValue)
	}
	if len(schema.OpeningHours) != 1 {
		t.Errorf("expected 1 opening hours, got %d", len(schema.OpeningHours))
	}
	if len(schema.AmenityFeature) != 3 {
		t.Errorf("expected 3 amenities, got %d", len(schema.AmenityFeature))
	}
}

func TestSitemapHandler_GetSchema_InvalidID(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	h := NewSitemapHandler(nil, nil, nil, log, "")

	r := chi.NewRouter()
	r.Get("/api/v1/bathhouses/{id}/schema", h.GetSchema)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses/invalid-id/schema", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSitemapHandler_GetSchema_NotFound(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := &sitemapMockBHService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Bathhouse, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := NewSitemapHandler(svc, nil, nil, log, "")

	r := chi.NewRouter()
	r.Get("/api/v1/bathhouses/{id}/schema", h.GetSchema)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses/"+uuid.New().String()+"/schema", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSitemapHandler_GenerateSitemap(t *testing.T) {
	log := logger.New(logger.LevelWarn)

	cities := []domain.City{
		{ID: 1, Name: "Москва", Slug: "moscow"},
		{ID: 2, Name: "Санкт-Петербург", Slug: "saint-petersburg"},
	}

	bathhouses := []domain.Bathhouse{
		{
			ID:        uuid.New(),
			Name:      "Баня 1",
			Slug:      "banya-1",
			CityID:    1,
			Status:    domain.BathhouseStatusActive,
			UpdatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.New(),
			Name:      "Баня 2",
			Slug:      "banya-2",
			CityID:    2,
			Status:    domain.BathhouseStatusActive,
			UpdatedAt: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	svc := &sitemapMockBHService{
		searchFn: func(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items:      bathhouses,
				TotalCount: 2,
				Page:       1,
				PageSize:   1000,
				TotalPages: 1,
			}, nil
		},
	}
	cityRepo := &mockCityRepo{
		getAllFn: func(_ context.Context) ([]domain.City, error) {
			return cities, nil
		},
	}

	h := NewSitemapHandler(svc, cityRepo, nil, log, "https://bani.ru")

	xmlData, err := h.generateSitemap(context.Background())
	if err != nil {
		t.Fatalf("generateSitemap failed: %v", err)
	}

	xmlStr := string(xmlData)

	// Verify XML structure
	if !strings.Contains(xmlStr, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Error("missing sitemap xmlns")
	}

	// Main page
	if !strings.Contains(xmlStr, "<loc>https://bani.ru</loc>") {
		t.Error("missing main page URL")
	}

	// Cities
	if !strings.Contains(xmlStr, "<loc>https://bani.ru/moscow</loc>") {
		t.Error("missing Moscow city URL")
	}
	if !strings.Contains(xmlStr, "<loc>https://bani.ru/saint-petersburg</loc>") {
		t.Error("missing Saint Petersburg city URL")
	}

	// Bathhouses with city slugs
	if !strings.Contains(xmlStr, "<loc>https://bani.ru/moscow/banya-1</loc>") {
		t.Error("missing bathhouse 1 URL with city slug")
	}
	if !strings.Contains(xmlStr, "<loc>https://bani.ru/saint-petersburg/banya-2</loc>") {
		t.Error("missing bathhouse 2 URL with city slug")
	}

	// Lastmod
	if !strings.Contains(xmlStr, "<lastmod>2026-03-01</lastmod>") {
		t.Error("missing lastmod for bathhouse 1")
	}

	// Priority
	if !strings.Contains(xmlStr, "<priority>0.8</priority>") {
		t.Error("missing priority 0.8 for bathhouses")
	}
	if !strings.Contains(xmlStr, "<priority>0.7</priority>") {
		t.Error("missing priority 0.7 for cities")
	}

	// Verify valid XML
	var urlset sitemapURLSet
	xmlContent := xmlStr
	if idx := strings.Index(xmlContent, "<urlset"); idx > 0 {
		xmlContent = xmlContent[idx:]
	}
	if err := xml.Unmarshal([]byte(xmlContent), &urlset); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}

	// 1 main + 2 cities + 2 bathhouses = 5 URLs
	if len(urlset.URLs) != 5 {
		t.Errorf("expected 5 URLs, got %d", len(urlset.URLs))
	}
}

func TestSitemapHandler_GenerateSitemap_EmptySlugs(t *testing.T) {
	log := logger.New(logger.LevelWarn)

	svc := &sitemapMockBHService{
		searchFn: func(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{
					{ID: uuid.New(), Name: "No Slug Bath", Slug: "", CityID: 1, Status: domain.BathhouseStatusActive},
				},
				TotalCount: 1,
				Page:       1,
				PageSize:   1000,
				TotalPages: 1,
			}, nil
		},
	}
	cityRepo := &mockCityRepo{
		getAllFn: func(_ context.Context) ([]domain.City, error) {
			return []domain.City{{ID: 1, Name: "NoSlug", Slug: ""}}, nil
		},
	}

	h := NewSitemapHandler(svc, cityRepo, nil, log, "https://bani.ru")

	xmlData, err := h.generateSitemap(context.Background())
	if err != nil {
		t.Fatalf("generateSitemap failed: %v", err)
	}

	xmlStr := string(xmlData)

	var urlset sitemapURLSet
	xmlContent := xmlStr
	if idx := strings.Index(xmlContent, "<urlset"); idx > 0 {
		xmlContent = xmlContent[idx:]
	}
	_ = xml.Unmarshal([]byte(xmlContent), &urlset)

	// Only 1 URL (main page) since city and bathhouse have empty slugs
	if len(urlset.URLs) != 1 {
		t.Errorf("expected 1 URL (main page only), got %d", len(urlset.URLs))
	}
}
