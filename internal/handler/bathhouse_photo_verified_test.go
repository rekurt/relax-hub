package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

func TestBathhouseHandler_GetByID_IsPhotoVerified(t *testing.T) {
	bathhouseID := uuid.New()
	ownerID := uuid.New()

	tests := []struct {
		name            string
		isPhotoVerified bool
	}{
		{
			name:            "photo verified true",
			isPhotoVerified: true,
		},
		{
			name:            "photo verified false",
			isPhotoVerified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBathhouse := &mockBathhouseService{
				getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Bathhouse, error) {
					return &domain.Bathhouse{
						ID:              bathhouseID,
						OwnerID:         ownerID,
						Name:            "Test Bathhouse",
						Status:          domain.BathhouseStatusActive,
						IsPhotoVerified: tt.isPhotoVerified,
					}, nil
				},
			}

			h := handler.NewBathhouseHandler(
				mockBathhouse,
				&mockBookingService{},
				nil, // representativeService
				nil, // favoriteService
				nil, // recommendationService
				nil, // analyticsService
				nil, // mediaService
				nil, // promotionService
				nil, // cityService
				nil, // savedSearchService
				nil, // suggestionService
				nil, // reviewService
				nil, // logger
				"",  // baseURL
			)

			r := chi.NewRouter()
			r.Get("/api/v1/bathhouses/{id}", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses/"+bathhouseID.String(), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
			}

			var resp struct {
				Success bool `json:"success"`
				Data    struct {
					IsPhotoVerified bool `json:"is_photo_verified"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Data.IsPhotoVerified != tt.isPhotoVerified {
				t.Errorf("expected is_photo_verified=%v, got %v", tt.isPhotoVerified, resp.Data.IsPhotoVerified)
			}
		})
	}
}

func TestBathhouseHandler_Search_IsPhotoVerified(t *testing.T) {
	bh1 := domain.Bathhouse{
		ID:              uuid.New(),
		OwnerID:         uuid.New(),
		Name:            "Verified Bathhouse",
		Status:          domain.BathhouseStatusActive,
		IsPhotoVerified: true,
	}
	bh2 := domain.Bathhouse{
		ID:              uuid.New(),
		OwnerID:         uuid.New(),
		Name:            "Unverified Bathhouse",
		Status:          domain.BathhouseStatusActive,
		IsPhotoVerified: false,
	}

	mockBathhouse := &mockBathhouseService{
		searchFn: func(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items:      []domain.Bathhouse{bh1, bh2},
				Page:       1,
				PageSize:   20,
				TotalCount: 2,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(
		mockBathhouse,
		&mockBookingService{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "",
	)

	r := chi.NewRouter()
	r.Get("/api/v1/bathhouses", h.Search)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			Name            string `json:"name"`
			IsPhotoVerified bool   `json:"is_photo_verified"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Data))
	}
	if !resp.Data[0].IsPhotoVerified {
		t.Error("expected first bathhouse to have is_photo_verified=true")
	}
	if resp.Data[1].IsPhotoVerified {
		t.Error("expected second bathhouse to have is_photo_verified=false")
	}
}
