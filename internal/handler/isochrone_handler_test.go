package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekurt/relax-hub/internal/geo"
	"github.com/rekurt/relax-hub/internal/logger"
)

func newTestIsochroneHandler(apiURL string) *IsochroneHandler {
	log := logger.New(logger.LevelWarn)
	svc := geo.NewIsochroneService(nil, log, apiURL, "test-key")
	return NewIsochroneHandler(svc)
}

func TestGetIsochrone_MissingLat(t *testing.T) {
	h := newTestIsochroneHandler("")

	req := httptest.NewRequest(http.MethodGet, "/isochrone?lon=37.62&mode=car&minutes=15", nil)
	rec := httptest.NewRecorder()

	h.GetIsochrone(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetIsochrone_MissingLon(t *testing.T) {
	h := newTestIsochroneHandler("")

	req := httptest.NewRequest(http.MethodGet, "/isochrone?lat=55.75&mode=car&minutes=15", nil)
	rec := httptest.NewRecorder()

	h.GetIsochrone(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetIsochrone_InvalidMode(t *testing.T) {
	h := newTestIsochroneHandler("")

	req := httptest.NewRequest(http.MethodGet, "/isochrone?lat=55.75&lon=37.62&mode=bicycle&minutes=15", nil)
	rec := httptest.NewRecorder()

	h.GetIsochrone(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetIsochrone_InvalidMinutes(t *testing.T) {
	h := newTestIsochroneHandler("")

	tests := []string{"0", "61", "-1"}
	for _, m := range tests {
		req := httptest.NewRequest(http.MethodGet, "/isochrone?lat=55.75&lon=37.62&mode=car&minutes="+m, nil)
		rec := httptest.NewRecorder()
		h.GetIsochrone(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("minutes=%s: expected 400, got %d", m, rec.Code)
		}
	}
}

func TestGetIsochrone_DefaultMode(t *testing.T) {
	// Mock ORS API that returns a valid polygon
	mockResp := map[string]interface{}{
		"features": []map[string]interface{}{
			{
				"geometry": map[string]interface{}{
					"coordinates": [][][]float64{
						{
							{37.5, 55.7},
							{37.6, 55.8},
							{37.7, 55.7},
							{37.5, 55.7},
						},
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	h := newTestIsochroneHandler(ts.URL)

	// No mode param - should default to "car"
	req := httptest.NewRequest(http.MethodGet, "/isochrone?lat=55.75&lon=37.62&minutes=15", nil)
	rec := httptest.NewRecorder()

	h.GetIsochrone(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetIsochrone_SuccessWithPolygon(t *testing.T) {
	mockResp := map[string]interface{}{
		"features": []map[string]interface{}{
			{
				"geometry": map[string]interface{}{
					"coordinates": [][][]float64{
						{
							{37.5, 55.7},
							{37.6, 55.8},
							{37.7, 55.7},
							{37.5, 55.7},
						},
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	h := newTestIsochroneHandler(ts.URL)

	req := httptest.NewRequest(http.MethodGet, "/isochrone?lat=55.75&lon=37.62&mode=car&minutes=15", nil)
	rec := httptest.NewRecorder()

	h.GetIsochrone(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Coordinates [][]float64 `json:"coordinates"`
			Mode        string      `json:"mode"`
			Minutes     int         `json:"minutes"`
			WKT         string      `json:"wkt"`
		} `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected success=true")
	}
	if len(resp.Data.Coordinates) != 4 {
		t.Errorf("expected 4 coordinates, got %d", len(resp.Data.Coordinates))
	}
	if resp.Data.Mode != "car" {
		t.Errorf("expected mode 'car', got %s", resp.Data.Mode)
	}
	if resp.Data.Minutes != 15 {
		t.Errorf("expected 15 minutes, got %d", resp.Data.Minutes)
	}
	if resp.Data.WKT == "" {
		t.Error("expected non-empty WKT")
	}
}
