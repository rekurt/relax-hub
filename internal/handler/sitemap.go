package handler

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/seo"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	sitemapCacheKey = "seo:sitemap:xml"
	sitemapCacheTTL = 24 * time.Hour
)

// SitemapHandler handles sitemap.xml and Schema.org endpoints.
type SitemapHandler struct {
	bathhouseService service.BathhouseService
	cityRepo         repository.CityRepository
	redis            *redis.Client
	log              *logger.Logger
	baseURL          string
}

func NewSitemapHandler(
	bathhouseService service.BathhouseService,
	cityRepo repository.CityRepository,
	redisClient *redis.Client,
	log *logger.Logger,
	baseURL string,
) *SitemapHandler {
	return &SitemapHandler{
		bathhouseService: bathhouseService,
		cityRepo:         cityRepo,
		redis:            redisClient,
		log:              log,
		baseURL:          baseURL,
	}
}

// XML sitemap types

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap serves /sitemap.xml with Redis caching.
func (h *SitemapHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Try cache first
	if h.redis != nil {
		cached, err := h.redis.Get(ctx, sitemapCacheKey).Result()
		if err == nil {
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(cached))
			return
		}
		if err != redis.Nil {
			h.log.Warn("redis get sitemap cache failed", "error", err)
		}
	}

	// Generate sitemap
	xmlData, err := h.generateSitemap(ctx)
	if err != nil {
		h.log.Error("failed to generate sitemap", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Cache in Redis
	if h.redis != nil {
		if cacheErr := h.redis.Set(ctx, sitemapCacheKey, string(xmlData), sitemapCacheTTL).Err(); cacheErr != nil {
			h.log.Error("failed to cache sitemap", "error", cacheErr)
		}
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xmlData)
}

func (h *SitemapHandler) generateSitemap(ctx context.Context) ([]byte, error) {
	baseURL := h.baseURL
	if baseURL == "" {
		baseURL = "https://bani.ru"
	}

	urlset := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	// Main page
	urlset.URLs = append(urlset.URLs, sitemapURL{
		Loc:        baseURL,
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	// Cities
	cities, err := h.cityRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cities: %w", err)
	}
	for _, city := range cities {
		if city.Slug == "" {
			continue
		}
		urlset.URLs = append(urlset.URLs, sitemapURL{
			Loc:        fmt.Sprintf("%s/%s", baseURL, city.Slug),
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	// Active bathhouses - fetch all with large page size
	activeStatus := domain.BathhouseStatusActive
	page := 1
	pageSize := 1000

	// Build city slug lookup
	citySlugMap := make(map[int64]string)
	for _, city := range cities {
		citySlugMap[city.ID] = city.Slug
	}

	for {
		result, err := h.bathhouseService.Search(ctx, domain.BathhouseFilter{
			Status:   &activeStatus,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("search bathhouses: %w", err)
		}

		for _, bh := range result.Items {
			if bh.Slug == "" {
				continue
			}
			loc := fmt.Sprintf("%s/bathhouses/%s", baseURL, bh.Slug)
			citySlug := citySlugMap[bh.CityID]
			if citySlug != "" {
				loc = fmt.Sprintf("%s/%s/%s", baseURL, citySlug, bh.Slug)
			}
			urlset.URLs = append(urlset.URLs, sitemapURL{
				Loc:        loc,
				LastMod:    bh.UpdatedAt.Format("2006-01-02"),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			})
		}

		if page >= result.TotalPages {
			break
		}
		page++
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal xml: %w", err)
	}

	return append([]byte(xml.Header), output...), nil
}

// GetSchema serves Schema.org JSON-LD for a bathhouse.
func (h *SitemapHandler) GetSchema(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid bathhouse ID")
		return
	}

	bh, err := h.bathhouseService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Get city info
	var cityName, citySlug string
	if bh.CityID > 0 {
		city, err := h.cityRepo.GetByID(r.Context(), bh.CityID)
		if err == nil {
			cityName = city.Name
			citySlug = city.Slug
		} else {
			h.log.Warn("failed to get city for schema", "city_id", bh.CityID, "error", err)
		}
	}

	input := seo.SchemaInput{
		Name:         bh.Name,
		Description:  bh.Description,
		Slug:         bh.Slug,
		CityName:     cityName,
		CitySlug:     citySlug,
		Address:      bh.Address,
		Latitude:     bh.Latitude,
		Longitude:    bh.Longitude,
		PricePerHour: bh.PricePerHour,
		Rating:       bh.Rating,
		ReviewCount:  bh.ReviewCount,
		Images:       bh.Images,
		HasPool:      bh.HasPool,
		HasSauna:     bh.HasSauna,
		HasSteamRoom: bh.HasSteamRoom,
		HasHotTub:    bh.HasHotTub,
		HasBBQ:       bh.HasBBQ,
		HasKaraoke:   bh.HasKaraoke,
		BaseURL:      h.baseURL,
	}

	for _, wh := range bh.WorkingHours {
		input.WorkingHours = append(input.WorkingHours, seo.WorkingHoursInput{
			DayOfWeek: wh.DayOfWeek,
			OpenTime:  wh.OpenTime,
			CloseTime: wh.CloseTime,
		})
	}

	schema := seo.GenerateSchema(input)

	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(schema)
}
