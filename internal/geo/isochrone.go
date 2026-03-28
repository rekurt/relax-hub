package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
)

// TravelMode represents the mode of transportation for isochrone calculation.
type TravelMode string

const (
	TravelModeCar     TravelMode = "car"
	TravelModeTransit TravelMode = "transit"
)

// IsochroneResult contains the polygon and metadata for an isochrone query.
type IsochroneResult struct {
	Coordinates [][]float64 `json:"coordinates"` // [lng, lat] pairs forming polygon
	Mode        TravelMode  `json:"mode"`
	Minutes     int         `json:"minutes"`
	CenterLat   float64     `json:"center_lat"`
	CenterLng   float64     `json:"center_lng"`
}

// IsochroneService fetches and caches isochrone polygons.
type IsochroneService struct {
	redis      *redis.Client
	httpClient *http.Client
	apiURL     string
	apiKey     string
	log        *logger.Logger
}

const (
	isochroneCachePrefix = "isochrone:"
	isochroneCacheTTL    = 1 * time.Hour
)

// NewIsochroneService creates a new IsochroneService.
func NewIsochroneService(redisClient *redis.Client, log *logger.Logger, apiURL, apiKey string) *IsochroneService {
	if apiURL == "" {
		apiURL = "https://api.openrouteservice.org"
	}
	return &IsochroneService{
		redis: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: apiURL,
		apiKey: apiKey,
		log:    log,
	}
}

func cacheKey(lat, lng float64, mode TravelMode, minutes int) string {
	return fmt.Sprintf("%s%.4f_%.4f_%s_%d", isochroneCachePrefix, lat, lng, mode, minutes)
}

// GetIsochrone returns the reachable polygon for a given origin, mode, and time.
func (s *IsochroneService) GetIsochrone(ctx context.Context, lat, lng float64, mode TravelMode, minutes int) (*IsochroneResult, error) {
	if minutes < 1 || minutes > 120 {
		return nil, fmt.Errorf("minutes must be between 1 and 120")
	}
	if mode != TravelModeCar && mode != TravelModeTransit {
		return nil, fmt.Errorf("unsupported travel mode: %s", mode)
	}

	key := cacheKey(lat, lng, mode, minutes)

	// Check cache
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var result IsochroneResult
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				return &result, nil
			}
		}
	}

	// Fetch from API
	result, err := s.fetchIsochrone(ctx, lat, lng, mode, minutes)
	if err != nil {
		return nil, err
	}

	// Cache result
	if s.redis != nil {
		data, err := json.Marshal(result)
		if err == nil {
			if err := s.redis.Set(ctx, key, data, isochroneCacheTTL).Err(); err != nil {
				s.log.Error("failed to cache isochrone", "error", err)
			}
		}
	}

	return result, nil
}

// orsProfile maps our travel mode to OpenRouteService profile names.
func orsProfile(mode TravelMode) string {
	switch mode {
	case TravelModeTransit:
		return "foot-walking" // ORS has no transit; foot-walking approximates short transit trips
	default:
		return "driving-car"
	}
}

// orsIsochroneResponse is the OpenRouteService isochrones response.
type orsIsochroneResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func (s *IsochroneService) fetchIsochrone(ctx context.Context, lat, lng float64, mode TravelMode, minutes int) (*IsochroneResult, error) {
	profile := orsProfile(mode)
	url := fmt.Sprintf("%s/v2/isochrones/%s", s.apiURL, profile)

	body := fmt.Sprintf(`{
		"locations": [[%f, %f]],
		"range": [%d],
		"range_type": "time"
	}`, lng, lat, minutes*60)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("isochrone API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("isochrone API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var orsResp orsIsochroneResponse
	if err := json.Unmarshal(respBody, &orsResp); err != nil {
		return nil, fmt.Errorf("parse isochrone response: %w", err)
	}

	if len(orsResp.Features) == 0 || len(orsResp.Features[0].Geometry.Coordinates) == 0 {
		return nil, fmt.Errorf("no isochrone polygon returned")
	}

	ring := orsResp.Features[0].Geometry.Coordinates[0]
	coords := make([][]float64, len(ring))
	for i, pt := range ring {
		coords[i] = []float64{pt[0], pt[1]} // [lng, lat]
	}

	return &IsochroneResult{
		Coordinates: coords,
		Mode:        mode,
		Minutes:     minutes,
		CenterLat:   lat,
		CenterLng:   lng,
	}, nil
}

// ToWKTPolygon converts the isochrone result to a WKT POLYGON string for PostGIS.
func (r *IsochroneResult) ToWKTPolygon() string {
	if len(r.Coordinates) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("POLYGON((")
	for i, pt := range r.Coordinates {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%f %f", pt[0], pt[1])) // lng lat (WKT order)
	}
	// Close the ring if not already closed
	first := r.Coordinates[0]
	last := r.Coordinates[len(r.Coordinates)-1]
	if first[0] != last[0] || first[1] != last[1] {
		b.WriteString(fmt.Sprintf(", %f %f", first[0], first[1]))
	}
	b.WriteString("))")
	return b.String()
}
