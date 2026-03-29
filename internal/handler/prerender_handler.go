package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/seo"
	"github.com/nikitaaldaev/bani/internal/service"
)

// PrerenderHandler serves pre-rendered HTML pages for search engine bots.
type PrerenderHandler struct {
	renderer         *seo.Renderer
	bathhouseService service.BathhouseService
	cityService      service.CityService
	reviewService    service.ReviewService
	log              *logger.Logger
	baseURL          string
}

func NewPrerenderHandler(
	renderer *seo.Renderer,
	bathhouseService service.BathhouseService,
	cityService service.CityService,
	reviewService service.ReviewService,
	log *logger.Logger,
	baseURL string,
) *PrerenderHandler {
	return &PrerenderHandler{
		renderer:         renderer,
		bathhouseService: bathhouseService,
		cityService:      cityService,
		reviewService:    reviewService,
		log:              log,
		baseURL:          baseURL,
	}
}

// BathhouseDetail serves pre-rendered HTML for a bathhouse detail page.
//
//	@Summary		Get pre-rendered bathhouse page for SEO
//	@Description	Returns pre-rendered HTML with meta tags and Schema.org JSON-LD for search engine bots
//	@Tags			seo
//	@Produce		html
//	@Param			slug	path	string	true	"Bathhouse slug"
//	@Success		200		{string}	string	"HTML page"
//	@Failure		404		{string}	string	"Not found"
//	@Router			/prerender/bathhouses/{slug} [get]
func (h *PrerenderHandler) BathhouseDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	cacheKey := seo.CacheKeyForDetail(slug)
	if cached, ok := h.renderer.GetCachedPage(r.Context(), cacheKey); ok {
		h.writeHTML(w, cached)
		return
	}

	bh, err := h.bathhouseService.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var cityName, citySlug string
	if bh.CityID > 0 {
		city, err := h.cityService.GetByID(r.Context(), bh.CityID)
		if err == nil {
			cityName = city.Name
			citySlug = city.Slug
		}
	}

	metaInput := seo.MetaInput{
		Name:         bh.Name,
		CityName:     cityName,
		CitySlug:     citySlug,
		Slug:         bh.Slug,
		Description:  bh.Description,
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
	meta := seo.GenerateMetaTags(metaInput)

	schemaInput := seo.SchemaInput{
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
		schemaInput.WorkingHours = append(schemaInput.WorkingHours, seo.WorkingHoursInput{
			DayOfWeek: wh.DayOfWeek,
			OpenTime:  wh.OpenTime,
			CloseTime: wh.CloseTime,
		})
	}

	pageData := h.renderer.BuildBathhouseDetailPage(schemaInput, meta)
	htmlContent := h.renderer.RenderPage(pageData)

	h.renderer.CachePage(r.Context(), cacheKey, htmlContent)
	h.writeHTML(w, htmlContent)
}

// CityListing serves pre-rendered HTML for a city listing page.
//
//	@Summary		Get pre-rendered city listing page for SEO
//	@Description	Returns pre-rendered HTML for city bathhouse catalog
//	@Tags			seo
//	@Produce		html
//	@Param			citySlug	path	string	true	"City slug"
//	@Success		200		{string}	string	"HTML page"
//	@Failure		404		{string}	string	"Not found"
//	@Router			/prerender/cities/{citySlug} [get]
func (h *PrerenderHandler) CityListing(w http.ResponseWriter, r *http.Request) {
	citySlug := chi.URLParam(r, "citySlug")
	if citySlug == "" {
		http.NotFound(w, r)
		return
	}

	cacheKey := seo.CacheKeyForCity(citySlug)
	if cached, ok := h.renderer.GetCachedPage(r.Context(), cacheKey); ok {
		h.writeHTML(w, cached)
		return
	}

	city, err := h.cityService.GetBySlug(r.Context(), citySlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	activeStatus := domain.BathhouseStatusActive
	result, err := h.bathhouseService.Search(r.Context(), domain.BathhouseFilter{
		CityID:   &city.ID,
		Status:   &activeStatus,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		h.log.Error("prerender city listing search failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pageData := h.renderer.BuildCityListingPage(city.Name, city.Slug, result.TotalCount)
	htmlContent := h.renderer.RenderPage(pageData)

	h.renderer.CachePage(r.Context(), cacheKey, htmlContent)
	h.writeHTML(w, htmlContent)
}

// MainListing serves pre-rendered HTML for the main catalog page.
//
//	@Summary		Get pre-rendered main listing page for SEO
//	@Description	Returns pre-rendered HTML for main bathhouse catalog
//	@Tags			seo
//	@Produce		html
//	@Success		200		{string}	string	"HTML page"
//	@Router			/prerender/catalog [get]
func (h *PrerenderHandler) MainListing(w http.ResponseWriter, r *http.Request) {
	cacheKey := seo.CacheKeyForListing()
	if cached, ok := h.renderer.GetCachedPage(r.Context(), cacheKey); ok {
		h.writeHTML(w, cached)
		return
	}

	activeStatus := domain.BathhouseStatusActive
	result, err := h.bathhouseService.Search(r.Context(), domain.BathhouseFilter{
		Status:   &activeStatus,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		h.log.Error("prerender main listing search failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pageData := h.renderer.BuildMainListingPage(result.TotalCount)
	htmlContent := h.renderer.RenderPage(pageData)

	h.renderer.CachePage(r.Context(), cacheKey, htmlContent)
	h.writeHTML(w, htmlContent)
}

// BathhouseReviews serves pre-rendered HTML for a bathhouse reviews page.
//
//	@Summary		Get pre-rendered reviews page for SEO
//	@Description	Returns pre-rendered HTML with reviews content for search engine bots
//	@Tags			seo
//	@Produce		html
//	@Param			slug	path	string	true	"Bathhouse slug"
//	@Success		200		{string}	string	"HTML page"
//	@Failure		404		{string}	string	"Not found"
//	@Router			/prerender/bathhouses/{slug}/reviews [get]
func (h *PrerenderHandler) BathhouseReviews(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	cacheKey := seo.CacheKeyForReviews(slug)
	if cached, ok := h.renderer.GetCachedPage(r.Context(), cacheKey); ok {
		h.writeHTML(w, cached)
		return
	}

	bh, err := h.bathhouseService.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var cityName, citySlug string
	if bh.CityID > 0 {
		city, err := h.cityService.GetByID(r.Context(), bh.CityID)
		if err == nil {
			cityName = city.Name
			citySlug = city.Slug
		}
	}

	// Get first page of reviews
	var reviewTexts []string
	reviews, err := h.reviewService.ListByBathhouse(r.Context(), bh.ID, 1, 10)
	if err == nil {
		for _, rev := range reviews.Items {
			if rev.Text != "" && rev.IsRevealed {
				reviewTexts = append(reviewTexts, rev.Text)
			}
		}
	}

	pageData := h.renderer.BuildReviewsPage(bh.Name, cityName, citySlug, bh.Slug, bh.Rating, bh.ReviewCount, reviewTexts)
	htmlContent := h.renderer.RenderPage(pageData)

	h.renderer.CachePage(r.Context(), cacheKey, htmlContent)
	h.writeHTML(w, htmlContent)
}

// InvalidateBathhouseCache invalidates pre-rendered cache for a bathhouse.
//
//	@Summary		Invalidate pre-rendered cache for a bathhouse
//	@Description	Removes cached pre-rendered HTML pages for a specific bathhouse (admin use)
//	@Tags			seo
//	@Param			slug	path	string	true	"Bathhouse slug"
//	@Success		204
//	@Router			/admin/prerender/invalidate/{slug} [post]
func (h *PrerenderHandler) InvalidateBathhouseCache(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	h.renderer.InvalidateCache(r.Context(), slug)
	w.WriteHeader(http.StatusNoContent)
}

func (h *PrerenderHandler) writeHTML(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noarchive")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}
