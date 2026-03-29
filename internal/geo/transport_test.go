package geo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaaldaev/bani/internal/logger"
)

func TestNewTransportService(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.apiKey != "test-key" {
		t.Errorf("expected api key 'test-key', got %s", svc.apiKey)
	}
}

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		expected int
		delta    int // acceptable error in meters
	}{
		{
			name:     "Moscow center to Kremlin (very short)",
			lat1:     55.7558,
			lng1:     37.6173,
			lat2:     55.7520,
			lng2:     37.6175,
			expected: 423,
			delta:    50,
		},
		{
			name:     "same point",
			lat1:     55.7558,
			lng1:     37.6173,
			lat2:     55.7558,
			lng2:     37.6173,
			expected: 0,
			delta:    0,
		},
		{
			name:     "Moscow to SPB (approx 635km)",
			lat1:     55.7558,
			lng1:     37.6173,
			lat2:     59.9343,
			lng2:     30.3351,
			expected: 635000,
			delta:    10000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HaversineDistance(tc.lat1, tc.lng1, tc.lat2, tc.lng2)
			diff := got - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.delta {
				t.Errorf("HaversineDistance() = %d, want %d (±%d)", got, tc.expected, tc.delta)
			}
		})
	}
}

func TestMetersToDegrees(t *testing.T) {
	deg := metersToDegrees(1000)
	if deg < 0.008 || deg > 0.01 {
		t.Errorf("metersToDegrees(1000) = %f, expected ~0.009", deg)
	}
}

func TestMetersToDegreesLng(t *testing.T) {
	// At equator, 1 degree lng ~ 111km, same as lat
	degEquator := metersToDegreesLng(1000, 0)
	degLat := metersToDegrees(1000)
	if degEquator < degLat*0.9 || degEquator > degLat*1.1 {
		t.Errorf("metersToDegreesLng at equator should be ~= metersToDegrees, got %f vs %f", degEquator, degLat)
	}

	// At 60 degrees, lng degrees should be about 2x lat degrees
	deg60 := metersToDegreesLng(1000, 60)
	if deg60 < degLat*1.8 || deg60 > degLat*2.2 {
		t.Errorf("metersToDegreesLng at 60° should be ~2x metersToDegrees, got %f vs %f", deg60, degLat)
	}
}

func TestTransportTypeConstants(t *testing.T) {
	if TransportTypeMetro != "metro" {
		t.Errorf("expected 'metro', got %s", TransportTypeMetro)
	}
	if TransportTypeBus != "bus_stop" {
		t.Errorf("expected 'bus_stop', got %s", TransportTypeBus)
	}
	if TransportTypeParking != "parking" {
		t.Errorf("expected 'parking', got %s", TransportTypeParking)
	}
}

func TestSearchCategory_Success(t *testing.T) {
	mockResp := yandexSearchResponse{
		Features: []yandexFeature{
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name: "Станция Партизанская",
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{37.7490, 55.7880},
				},
			},
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name: "Станция Измайловская",
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{37.7810, 55.7870},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("apikey") != "test-key" {
			t.Errorf("expected apikey=test-key, got %s", q.Get("apikey"))
		}
		if q.Get("lang") != "ru_RU" {
			t.Errorf("expected lang=ru_RU, got %s", q.Get("lang"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	// Override the API URL to use our test server
	items, err := svc.searchCategoryWithURL(context.Background(), 55.7558, 37.6173, "метро", TransportTypeMetro, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "Станция Партизанская" && items[1].Name != "Станция Партизанская" {
		t.Error("expected Станция Партизанская in results")
	}
	if items[0].Type != TransportTypeMetro {
		t.Errorf("expected type metro, got %s", items[0].Type)
	}
	// Items should be sorted by distance
	if len(items) >= 2 && items[0].Distance > items[1].Distance {
		t.Errorf("items not sorted by distance: %d > %d", items[0].Distance, items[1].Distance)
	}
}

func TestSearchCategory_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	_, err := svc.searchCategoryWithURL(context.Background(), 55.7558, 37.6173, "метро", TransportTypeMetro, ts.URL)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestSearchCategory_EmptyResults(t *testing.T) {
	mockResp := yandexSearchResponse{Features: []yandexFeature{}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	items, err := svc.searchCategoryWithURL(context.Background(), 55.7558, 37.6173, "метро", TransportTypeMetro, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestSearchCategory_InvalidCoordinates(t *testing.T) {
	// Feature with missing coordinates should be skipped
	mockResp := yandexSearchResponse{
		Features: []yandexFeature{
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name: "Test",
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{}, // empty coordinates
				},
			},
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name: "Valid Stop",
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{37.62, 55.75},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	items, err := svc.searchCategoryWithURL(context.Background(), 55.7558, 37.6173, "метро", TransportTypeMetro, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 valid item (invalid coords skipped), got %d", len(items))
	}
	if len(items) > 0 && items[0].Name != "Valid Stop" {
		t.Errorf("expected 'Valid Stop', got %s", items[0].Name)
	}
}

func TestSearchCategory_FallbackName(t *testing.T) {
	// Test name fallback: CompanyMetaData.Name then Description
	mockResp := yandexSearchResponse{
		Features: []yandexFeature{
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name:        "",
					Description: "",
					CompanyMetaData: &struct {
						Name string `json:"name"`
					}{Name: "Company Name"},
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{37.62, 55.75},
				},
			},
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{
					Name:        "",
					Description: "Description Fallback",
				},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{
					Coordinates: []float64{37.63, 55.76},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	items, err := svc.searchCategoryWithURL(context.Background(), 55.7558, 37.6173, "test", TransportTypeBus, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Find each by checking names
	foundCompany := false
	foundDesc := false
	for _, item := range items {
		if item.Name == "Company Name" {
			foundCompany = true
		}
		if item.Name == "Description Fallback" {
			foundDesc = true
		}
	}
	if !foundCompany {
		t.Error("expected CompanyMetaData.Name fallback")
	}
	if !foundDesc {
		t.Error("expected Description fallback")
	}
}

func TestGetNearbyTransport_NilRedis(t *testing.T) {
	// Full integration test with mock HTTP server for all 3 categories
	metroResp := yandexSearchResponse{
		Features: []yandexFeature{
			{
				Properties: struct {
					Name            string `json:"name"`
					Description     string `json:"description"`
					CompanyMetaData *struct {
						Name string `json:"name"`
					} `json:"CompanyMetaData,omitempty"`
				}{Name: "Метро Тест"},
				Geometry: struct {
					Coordinates []float64 `json:"coordinates"`
				}{Coordinates: []float64{37.62, 55.75}},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metroResp)
	}))
	defer ts.Close()

	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")
	svc.testBaseURL = ts.URL

	result, err := svc.GetNearbyTransport(context.Background(), "test-bathhouse-id", 55.7558, 37.6173)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.BathhouseID != "test-bathhouse-id" {
		t.Errorf("expected bathhouse ID 'test-bathhouse-id', got %s", result.BathhouseID)
	}
	// Should have results from all 3 categories (mock returns same for each)
	if len(result.Items) == 0 {
		t.Error("expected non-empty transport items")
	}
	// Items should be sorted by distance
	for i := 1; i < len(result.Items); i++ {
		if result.Items[i].Distance < result.Items[i-1].Distance {
			t.Errorf("items not sorted by distance at index %d", i)
		}
	}
}

func TestInvalidateCache_NilRedis(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	svc := NewTransportService(nil, log, "test-key")

	// Should not panic with nil redis
	err := svc.InvalidateCache(context.Background(), "test-id")
	if err != nil {
		t.Errorf("expected nil error for nil redis, got %v", err)
	}
}
