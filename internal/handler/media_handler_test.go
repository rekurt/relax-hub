package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

func TestMediaHandler_BathhouseGallery(t *testing.T) {
	bathhouseID := uuid.New()
	mediaID1 := uuid.New()
	mediaID2 := uuid.New()

	mediaSvc := &mockMediaService{
		listByBathhouseFn: func(_ context.Context, bhID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
			if bhID != bathhouseID {
				return nil, domain.ErrNotFound
			}
			return &domain.PaginatedResult[domain.Media]{
				Items: []domain.Media{
					{
						ID:           mediaID1,
						OwnerType:    domain.MediaOwnerReview,
						OwnerID:      uuid.New(),
						UserID:       uuid.New(),
						Type:         domain.MediaTypeImage,
						URL:          "https://s3.example.com/media/1/main.jpg",
						ThumbnailURL: "https://s3.example.com/media/1/thumb.jpg",
						OriginalName: "photo1.jpg",
						Size:         102400,
						MimeType:     "image/jpeg",
						Width:        1920,
						Height:       1080,
						Status:       domain.MediaStatusPending,
						CreatedAt:    time.Now(),
					},
					{
						ID:           mediaID2,
						OwnerType:    domain.MediaOwnerReview,
						OwnerID:      uuid.New(),
						UserID:       uuid.New(),
						Type:         domain.MediaTypeImage,
						URL:          "https://s3.example.com/media/2/main.jpg",
						ThumbnailURL: "https://s3.example.com/media/2/thumb.jpg",
						OriginalName: "photo2.jpg",
						Size:         204800,
						MimeType:     "image/jpeg",
						Width:        1280,
						Height:       720,
						Status:       domain.MediaStatusPending,
						CreatedAt:    time.Now(),
					},
				},
				TotalCount: 2,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewMediaHandler(mediaSvc)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/gallery", h.BathhouseGallery)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/gallery", bathhouseID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, mediaID1.String()) {
			t.Errorf("expected body to contain media ID %s", mediaID1)
		}
		if !strings.Contains(body, mediaID2.String()) {
			t.Errorf("expected body to contain media ID %s", mediaID2)
		}
		if !strings.Contains(body, `"total_count":2`) {
			t.Errorf("expected total_count 2 in response")
		}
	})

	t.Run("invalid bathhouse id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/bathhouses/invalid/gallery", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/gallery?page=1&page_size=10", bathhouseID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestMediaHandler_BathhouseGallery_Empty(t *testing.T) {
	bathhouseID := uuid.New()

	mediaSvc := &mockMediaService{
		listByBathhouseFn: func(_ context.Context, _ uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
			return &domain.PaginatedResult[domain.Media]{
				Items:      []domain.Media{},
				TotalCount: 0,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewMediaHandler(mediaSvc)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/gallery", h.BathhouseGallery)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/gallery", bathhouseID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"total_count":0`) {
		t.Errorf("expected total_count 0 in response")
	}
}

func TestBathhouseHandler_GetByID_WithGalleryPreview(t *testing.T) {
	bathhouseID := uuid.New()

	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID:     bathhouseID,
				Name:   "Test Banya",
				Status: domain.BathhouseStatusActive,
				Images: []string{},
			}, nil
		},
	}

	mediaID := uuid.New()
	mediaSvc := &mockMediaService{
		listByBathhouseFn: func(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Media], error) {
			return &domain.PaginatedResult[domain.Media]{
				Items: []domain.Media{
					{
						ID:           mediaID,
						Type:         domain.MediaTypeImage,
						URL:          "https://s3.example.com/media/1/main.jpg",
						ThumbnailURL: "https://s3.example.com/media/1/thumb.jpg",
						Status:       domain.MediaStatusPending,
						CreatedAt:    time.Now(),
					},
				},
				TotalCount: 1,
				Page:       1,
				PageSize:   4,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil, nil, nil, mediaSvc, nil, nil, nil, nil, nil, "")

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s", bathhouseID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"gallery_preview"`) {
		t.Errorf("expected gallery_preview in response")
	}
	if !strings.Contains(body, mediaID.String()) {
		t.Errorf("expected media ID %s in gallery_preview", mediaID)
	}
}

func TestBathhouseHandler_GetByID_NoGalleryPreview(t *testing.T) {
	bathhouseID := uuid.New()

	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID:     bathhouseID,
				Name:   "Test Banya",
				Status: domain.BathhouseStatusActive,
				Images: []string{},
			}, nil
		},
	}

	mediaSvc := &mockMediaService{
		listByBathhouseFn: func(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Media], error) {
			return &domain.PaginatedResult[domain.Media]{
				Items:      []domain.Media{},
				TotalCount: 0,
				Page:       1,
				PageSize:   4,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil, nil, nil, mediaSvc, nil, nil, nil, nil, nil, "")

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s", bathhouseID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// When no gallery items, gallery_preview should be omitted (omitempty)
	if strings.Contains(rec.Body.String(), `"gallery_preview"`) {
		t.Errorf("expected no gallery_preview when empty")
	}
}
