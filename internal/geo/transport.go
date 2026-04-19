package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/redis/go-redis/v9"
)

// TransportType represents a type of nearby transport infrastructure.
type TransportType string

const (
	TransportTypeMetro   TransportType = "metro"
	TransportTypeBus     TransportType = "bus_stop"
	TransportTypeParking TransportType = "parking"
)

// TransportInfo describes a single nearby transport point.
type TransportInfo struct {
	Type     TransportType `json:"type"`
	Name     string        `json:"name"`
	Distance int           `json:"distance_meters"` // distance from bathhouse in meters
	Lat      float64       `json:"lat"`
	Lng      float64       `json:"lng"`
}

// TransportResult contains all nearby transport for a bathhouse.
type TransportResult struct {
	BathhouseID string          `json:"bathhouse_id"`
	Items       []TransportInfo `json:"items"`
	FetchedAt   time.Time       `json:"fetched_at"`
}

// TransportService fetches and caches nearby transport data for bathhouses.
type TransportService struct {
	redis       *redis.Client
	httpClient  *http.Client
	apiKey      string
	log         *logger.Logger
	testBaseURL string // override API URL for testing; empty in production
}

const (
	transportCachePrefix = "transport:"
	transportCacheTTL    = 7 * 24 * time.Hour // 7 days
	transportSearchRadius = 1500              // meters
	maxTransportResults   = 5                 // max results per transport type
)

// NewTransportService creates a new TransportService.
func NewTransportService(redisClient *redis.Client, log *logger.Logger, apiKey string) *TransportService {
	return &TransportService{
		redis: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey: apiKey,
		log:    log,
	}
}

// GetNearbyTransport returns nearby transport infrastructure for a location.
// Results are cached per bathhouse ID in Redis.
func (s *TransportService) GetNearbyTransport(ctx context.Context, bathhouseID string, lat, lng float64) (*TransportResult, error) {
	key := fmt.Sprintf("%s%s", transportCachePrefix, bathhouseID)

	// Check cache
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var result TransportResult
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return &result, nil
			}
		}
	}

	// Fetch from API
	items, err := s.fetchNearbyTransport(ctx, lat, lng)
	if err != nil {
		return nil, err
	}

	result := &TransportResult{
		BathhouseID: bathhouseID,
		Items:       items,
		FetchedAt:   time.Now().UTC(),
	}

	// Cache result
	if s.redis != nil {
		data, err := json.Marshal(result)
		if err == nil {
			if err := s.redis.Set(ctx, key, data, transportCacheTTL).Err(); err != nil {
				s.log.Error("failed to cache transport data", "error", err)
			}
		}
	}

	return result, nil
}

// InvalidateCache removes cached transport data for a bathhouse.
func (s *TransportService) InvalidateCache(ctx context.Context, bathhouseID string) error {
	if s.redis == nil {
		return nil
	}
	key := fmt.Sprintf("%s%s", transportCachePrefix, bathhouseID)
	return s.redis.Del(ctx, key).Err()
}

// fetchNearbyTransport queries the Yandex Maps Geocoder/Search API for nearby POIs.
func (s *TransportService) fetchNearbyTransport(ctx context.Context, lat, lng float64) ([]TransportInfo, error) {
	var allItems []TransportInfo

	categories := []struct {
		query         string
		transportType TransportType
	}{
		{"метро", TransportTypeMetro},
		{"остановка общественного транспорта", TransportTypeBus},
		{"парковка", TransportTypeParking},
	}

	for _, cat := range categories {
		items, err := s.searchCategory(ctx, lat, lng, cat.query, cat.transportType)
		if err != nil {
			s.log.Error("failed to fetch transport category",
				"category", cat.transportType,
				"error", err,
			)
			continue // partial results are acceptable
		}
		allItems = append(allItems, items...)
	}

	// Sort by distance
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Distance < allItems[j].Distance
	})

	return allItems, nil
}

// yandexSearchResponse represents the Yandex Maps Search API response.
type yandexSearchResponse struct {
	Features []yandexFeature `json:"features"`
}

type yandexFeature struct {
	Properties struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		CompanyMetaData  *struct {
			Name string `json:"name"`
		} `json:"CompanyMetaData,omitempty"`
	} `json:"properties"`
	Geometry struct {
		Coordinates []float64 `json:"coordinates"` // [lng, lat]
	} `json:"geometry"`
}

func (s *TransportService) searchCategory(ctx context.Context, lat, lng float64, query string, transportType TransportType) ([]TransportInfo, error) {
	baseURL := "https://search-maps.yandex.ru/v1/"
	if s.testBaseURL != "" {
		baseURL = s.testBaseURL + "/"
	}
	return s.searchCategoryWithURL(ctx, lat, lng, query, transportType, baseURL)
}

func (s *TransportService) searchCategoryWithURL(ctx context.Context, lat, lng float64, query string, transportType TransportType, baseURL string) ([]TransportInfo, error) {
	params := url.Values{}
	params.Set("apikey", s.apiKey)
	params.Set("text", query)
	params.Set("lang", "ru_RU")
	params.Set("ll", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("spn", fmt.Sprintf("%.4f,%.4f", metersToDegreesLng(transportSearchRadius, lat), metersToDegrees(transportSearchRadius)))
	params.Set("type", "biz")
	params.Set("results", fmt.Sprintf("%d", maxTransportResults))

	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yandex search API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex search API returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp yandexSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	var items []TransportInfo
	for _, feature := range searchResp.Features {
		if len(feature.Geometry.Coordinates) < 2 {
			continue
		}

		poiLng := feature.Geometry.Coordinates[0]
		poiLat := feature.Geometry.Coordinates[1]
		distance := HaversineDistance(lat, lng, poiLat, poiLng)

		name := feature.Properties.Name
		if name == "" && feature.Properties.CompanyMetaData != nil {
			name = feature.Properties.CompanyMetaData.Name
		}
		if name == "" {
			name = feature.Properties.Description
		}

		items = append(items, TransportInfo{
			Type:     transportType,
			Name:     name,
			Distance: distance,
			Lat:      poiLat,
			Lng:      poiLng,
		})
	}

	// Sort by distance and limit
	sort.Slice(items, func(i, j int) bool {
		return items[i].Distance < items[j].Distance
	})
	if len(items) > maxTransportResults {
		items = items[:maxTransportResults]
	}

	return items, nil
}

// HaversineDistance calculates the distance in meters between two lat/lng points.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) int {
	const earthRadius = 6371000 // meters

	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(math.Round(earthRadius * c))
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// metersToDegrees converts meters to approximate degrees of latitude.
func metersToDegrees(meters int) float64 {
	return float64(meters) / 111320.0
}

// metersToDegreesLng converts meters to approximate degrees of longitude at given latitude.
func metersToDegreesLng(meters int, lat float64) float64 {
	return float64(meters) / (111320.0 * math.Cos(degreesToRadians(lat)))
}
