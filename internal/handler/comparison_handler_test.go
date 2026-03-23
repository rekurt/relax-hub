package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

func TestComparisonHandler_Compare(t *testing.T) {
	bh1ID := uuid.New()
	bh2ID := uuid.New()
	bh3ID := uuid.New()

	activeBathhouse := func(id uuid.UUID, name string, lat, lon float64) *domain.Bathhouse {
		return &domain.Bathhouse{
			ID:           id,
			OwnerID:      uuid.New(),
			Name:         name,
			Slug:         name,
			Address:      "Test Address",
			CityID:       1,
			Latitude:     lat,
			Longitude:    lon,
			PricePerHour: 500000,
			MinDuration:  1,
			MaxGuests:    10,
			Rating:       4.5,
			ReviewCount:  20,
			HasPool:      true,
			HasSauna:     true,
			Images:       []string{"img1.jpg"},
			Status:       domain.BathhouseStatusActive,
		}
	}

	bhMap := map[uuid.UUID]*domain.Bathhouse{
		bh1ID: activeBathhouse(bh1ID, "Баня 1", 55.75, 37.62),
		bh2ID: activeBathhouse(bh2ID, "Баня 2", 55.76, 37.63),
		bh3ID: activeBathhouse(bh3ID, "Баня 3", 55.77, 37.64),
	}

	mockSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if bh, ok := bhMap[id]; ok {
				return bh, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewComparisonHandler(mockSvc)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantItems  int
	}{
		{
			name:       "compare 2 bathhouses",
			body:       map[string]interface{}{"ids": []string{bh1ID.String(), bh2ID.String()}},
			wantStatus: http.StatusOK,
			wantItems:  2,
		},
		{
			name:       "compare 3 bathhouses",
			body:       map[string]interface{}{"ids": []string{bh1ID.String(), bh2ID.String(), bh3ID.String()}},
			wantStatus: http.StatusOK,
			wantItems:  3,
		},
		{
			name: "with user location calculates distance",
			body: map[string]interface{}{
				"ids":       []string{bh1ID.String(), bh2ID.String()},
				"latitude":  55.74,
				"longitude": 37.61,
			},
			wantStatus: http.StatusOK,
			wantItems:  2,
		},
		{
			name:       "too few IDs",
			body:       map[string]interface{}{"ids": []string{bh1ID.String()}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "too many IDs",
			body:       map[string]interface{}{"ids": []string{bh1ID.String(), bh2ID.String(), bh3ID.String(), uuid.New().String()}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid UUID",
			body:       map[string]interface{}{"ids": []string{"not-a-uuid", bh2ID.String()}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duplicate IDs",
			body:       map[string]interface{}{"ids": []string{bh1ID.String(), bh1ID.String()}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found bathhouse",
			body:       map[string]interface{}{"ids": []string{uuid.New().String(), bh2ID.String()}},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bathhouses/compare", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Compare(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Success bool `json:"success"`
					Data    struct {
						Items []json.RawMessage `json:"items"`
					} `json:"data"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if !resp.Success {
					t.Error("expected success=true")
				}
				if len(resp.Data.Items) != tt.wantItems {
					t.Errorf("got %d items, want %d", len(resp.Data.Items), tt.wantItems)
				}
			}
		})
	}
}

func TestComparisonHandler_Compare_InactiveBathhouse(t *testing.T) {
	activeID := uuid.New()
	inactiveID := uuid.New()

	mockSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == activeID {
				return &domain.Bathhouse{
					ID:           activeID,
					Name:         "Active",
					PricePerHour: 100000,
					MaxGuests:    5,
					MinDuration:  1,
					Status:       domain.BathhouseStatusActive,
				}, nil
			}
			// Service returns ErrNotFound for non-active bathhouses
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewComparisonHandler(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"ids": []string{activeID.String(), inactiveID.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bathhouses/compare", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Compare(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d for inactive bathhouse", rr.Code, http.StatusNotFound)
	}
}

func TestComparisonHandler_Compare_DistanceCalculation(t *testing.T) {
	bh1ID := uuid.New()
	bh2ID := uuid.New()

	mockSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			if id == bh1ID {
				return &domain.Bathhouse{
					ID: bh1ID, Name: "Near", Latitude: 55.75, Longitude: 37.62,
					PricePerHour: 100000, MaxGuests: 5, MinDuration: 1,
					Status: domain.BathhouseStatusActive,
				}, nil
			}
			return &domain.Bathhouse{
				ID: bh2ID, Name: "Far", Latitude: 55.85, Longitude: 37.72,
				PricePerHour: 200000, MaxGuests: 10, MinDuration: 2,
				Status: domain.BathhouseStatusActive,
			}, nil
		},
	}

	h := handler.NewComparisonHandler(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{
		"ids":       []string{bh1ID.String(), bh2ID.String()},
		"latitude":  55.75,
		"longitude": 37.62,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bathhouses/compare", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Compare(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}

	var resp struct {
		Data struct {
			Items []struct {
				Name     string   `json:"name"`
				Distance *float64 `json:"distance"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// First bathhouse is at the same location, distance should be ~0
	if resp.Data.Items[0].Distance == nil {
		t.Fatal("expected distance for first item")
	}
	if *resp.Data.Items[0].Distance > 0.1 {
		t.Errorf("expected near-zero distance for first item, got %f", *resp.Data.Items[0].Distance)
	}

	// Second bathhouse should have non-zero distance
	if resp.Data.Items[1].Distance == nil {
		t.Fatal("expected distance for second item")
	}
	if *resp.Data.Items[1].Distance < 1.0 {
		t.Errorf("expected >1km distance for second item, got %f", *resp.Data.Items[1].Distance)
	}
}
