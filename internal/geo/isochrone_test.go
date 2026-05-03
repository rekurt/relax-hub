package geo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekurt/relax-hub/internal/logger"
)

func TestNewIsochroneService(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, "", "test-key")

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.apiURL != "https://api.openrouteservice.org" {
		t.Errorf("expected default API URL, got %s", svc.apiURL)
	}
	if svc.apiKey != "test-key" {
		t.Errorf("expected api key 'test-key', got %s", svc.apiKey)
	}
}

func TestNewIsochroneService_CustomURL(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, "https://custom.ors.example.com", "key")

	if svc.apiURL != "https://custom.ors.example.com" {
		t.Errorf("expected custom URL, got %s", svc.apiURL)
	}
}

func TestCacheKey(t *testing.T) {
	key := cacheKey(55.7512, 37.6184, TravelModeCar, 15)
	expected := "isochrone:55.7512_37.6184_car_15"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestOrsProfile(t *testing.T) {
	tests := []struct {
		mode     TravelMode
		expected string
	}{
		{TravelModeCar, "driving-car"},
		{TravelModeTransit, "foot-walking"},
	}

	for _, tc := range tests {
		got := orsProfile(tc.mode)
		if got != tc.expected {
			t.Errorf("orsProfile(%s) = %s, want %s", tc.mode, got, tc.expected)
		}
	}
}

func TestIsochroneResult_ToWKTPolygon(t *testing.T) {
	result := &IsochroneResult{
		Coordinates: [][]float64{
			{37.5, 55.7},
			{37.6, 55.8},
			{37.7, 55.7},
			{37.5, 55.7}, // closed ring
		},
	}

	wkt := result.ToWKTPolygon()
	expected := "POLYGON((37.500000 55.700000, 37.600000 55.800000, 37.700000 55.700000, 37.500000 55.700000))"
	if wkt != expected {
		t.Errorf("expected WKT:\n%s\ngot:\n%s", expected, wkt)
	}
}

func TestIsochroneResult_ToWKTPolygon_AutoClose(t *testing.T) {
	result := &IsochroneResult{
		Coordinates: [][]float64{
			{37.5, 55.7},
			{37.6, 55.8},
			{37.7, 55.7},
			// not closed
		},
	}

	wkt := result.ToWKTPolygon()
	// Should auto-close by appending the first point
	expected := "POLYGON((37.500000 55.700000, 37.600000 55.800000, 37.700000 55.700000, 37.500000 55.700000))"
	if wkt != expected {
		t.Errorf("expected WKT:\n%s\ngot:\n%s", expected, wkt)
	}
}

func TestIsochroneResult_ToWKTPolygon_Empty(t *testing.T) {
	result := &IsochroneResult{}
	wkt := result.ToWKTPolygon()
	if wkt != "" {
		t.Errorf("expected empty WKT, got %q", wkt)
	}
}

func TestFetchIsochrone_Success(t *testing.T) {
	// Mock ORS API
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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "test-key" {
			t.Errorf("expected Authorization: test-key, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, ts.URL, "test-key")

	result, err := svc.fetchIsochrone(context.Background(), 55.75, 37.62, TravelModeCar, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Coordinates) != 4 {
		t.Errorf("expected 4 coordinates, got %d", len(result.Coordinates))
	}
	if result.Mode != TravelModeCar {
		t.Errorf("expected mode car, got %s", result.Mode)
	}
	if result.Minutes != 15 {
		t.Errorf("expected 15 minutes, got %d", result.Minutes)
	}
	if result.CenterLat != 55.75 {
		t.Errorf("expected center lat 55.75, got %f", result.CenterLat)
	}
}

func TestFetchIsochrone_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, ts.URL, "test-key")

	_, err := svc.fetchIsochrone(context.Background(), 55.75, 37.62, TravelModeCar, 15)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestFetchIsochrone_EmptyFeatures(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"features": []interface{}{}})
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, ts.URL, "test-key")

	_, err := svc.fetchIsochrone(context.Background(), 55.75, 37.62, TravelModeCar, 15)
	if err == nil {
		t.Fatal("expected error for empty features")
	}
}

func TestGetIsochrone_InvalidMinutes(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, "", "")

	_, err := svc.GetIsochrone(context.Background(), 55.75, 37.62, TravelModeCar, 0)
	if err == nil {
		t.Fatal("expected error for invalid minutes")
	}

	_, err = svc.GetIsochrone(context.Background(), 55.75, 37.62, TravelModeCar, 121)
	if err == nil {
		t.Fatal("expected error for minutes > 120")
	}
}

func TestGetIsochrone_InvalidMode(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewIsochroneService(nil, log, "", "")

	_, err := svc.GetIsochrone(context.Background(), 55.75, 37.62, "bicycle", 15)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestTravelModeConstants(t *testing.T) {
	if TravelModeCar != "car" {
		t.Errorf("expected 'car', got %s", TravelModeCar)
	}
	if TravelModeTransit != "transit" {
		t.Errorf("expected 'transit', got %s", TravelModeTransit)
	}
}
