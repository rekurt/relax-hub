package handler

import (
	"context"
	"crypto/md5"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/seo"
	"github.com/nikitaaldaev/bani/internal/service"
)

type BathhouseHandler struct {
	bathhouseService      service.BathhouseService
	bookingService        service.BookingService
	representativeService service.RepresentativeService
	favoriteService       service.FavoriteService
	recommendationService service.RecommendationService
	analyticsService      service.AnalyticsService
	mediaService          service.MediaService
	promotionService      service.PromotionService
	cityService           service.CityService
	savedSearchService    service.SavedSearchService
	suggestionService     service.SearchSuggestionService
	log                   *logger.Logger
	baseURL               string
}

func NewBathhouseHandler(
	bathhouseService service.BathhouseService,
	bookingService service.BookingService,
	representativeService service.RepresentativeService,
	favoriteService service.FavoriteService,
	recommendationService service.RecommendationService,
	analyticsService service.AnalyticsService,
	mediaService service.MediaService,
	promotionService service.PromotionService,
	cityService service.CityService,
	savedSearchService service.SavedSearchService,
	suggestionService service.SearchSuggestionService,
	log *logger.Logger,
	baseURL string,
) *BathhouseHandler {
	return &BathhouseHandler{
		bathhouseService:      bathhouseService,
		bookingService:        bookingService,
		representativeService: representativeService,
		favoriteService:       favoriteService,
		recommendationService: recommendationService,
		analyticsService:      analyticsService,
		mediaService:          mediaService,
		promotionService:      promotionService,
		cityService:           cityService,
		savedSearchService:    savedSearchService,
		suggestionService:     suggestionService,
		log:                   log,
		baseURL:               baseURL,
	}
}

// getIPHash extracts the client IP from the request and returns its hash
func getIPHash(r *http.Request) string {
	// Try X-Forwarded-For first (for proxied requests)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ip = strings.Split(ip, ",")[0]
		ip = strings.TrimSpace(ip)
	}

	// Fallback to RemoteAddr
	if ip == "" {
		ip = r.RemoteAddr
		// Remove port if present
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}

	// Hash the IP
	hash := md5.Sum([]byte(ip))
	return fmt.Sprintf("%x", hash)
}

type bathhouseResponse struct {
	ID           string             `json:"id"`
	OwnerID      string             `json:"owner_id"`
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	Address      string             `json:"address"`
	CityID       int64              `json:"city_id"`
	Latitude     float64            `json:"latitude"`
	Longitude    float64            `json:"longitude"`
	PricePerHour int64              `json:"price_per_hour"`
	MinDuration  int                `json:"min_duration"`
	MaxGuests    int                `json:"max_guests"`
	HasPool      bool               `json:"has_pool"`
	HasSauna     bool               `json:"has_sauna"`
	HasSteamRoom bool               `json:"has_steam_room"`
	HasHotTub    bool               `json:"has_hot_tub"`
	HasBBQ       bool               `json:"has_bbq"`
	HasKaraoke   bool               `json:"has_karaoke"`
	Rating       float64            `json:"rating"`
	ReviewCount  int                `json:"review_count"`
	Images       []string           `json:"images"`
	WorkingHours []workingHoursResp `json:"working_hours"`
	Status       string             `json:"status"`
	IsFavorite   bool               `json:"is_favorite"`
	IsPromoted       bool               `json:"is_promoted"`
	IsPhotoVerified  bool               `json:"is_photo_verified"`
	GalleryPreview   []mediaResponse    `json:"gallery_preview,omitempty"`
	Meta         *seo.MetaTags      `json:"meta,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type workingHoursResp struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
}

func toBathhouseResponse(b *domain.Bathhouse) bathhouseResponse {
	wh := make([]workingHoursResp, len(b.WorkingHours))
	for i, h := range b.WorkingHours {
		wh[i] = workingHoursResp{
			DayOfWeek: h.DayOfWeek,
			OpenTime:  h.OpenTime,
			CloseTime: h.CloseTime,
		}
	}
	images := b.Images
	if images == nil {
		images = []string{}
	}
	return bathhouseResponse{
		ID:           b.ID.String(),
		OwnerID:      b.OwnerID.String(),
		Name:         b.Name,
		Slug:         b.Slug,
		Description:  b.Description,
		Address:      b.Address,
		CityID:       b.CityID,
		Latitude:     b.Latitude,
		Longitude:    b.Longitude,
		PricePerHour: b.PricePerHour,
		MinDuration:  b.MinDuration,
		MaxGuests:    b.MaxGuests,
		HasPool:      b.HasPool,
		HasSauna:     b.HasSauna,
		HasSteamRoom: b.HasSteamRoom,
		HasHotTub:    b.HasHotTub,
		HasBBQ:       b.HasBBQ,
		HasKaraoke:   b.HasKaraoke,
		Rating:       b.Rating,
		ReviewCount:  b.ReviewCount,
		Images:       images,
		WorkingHours: wh,
		Status:       string(b.Status),
		IsPromoted:      b.IsPromoted,
		IsPhotoVerified: b.IsPhotoVerified,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
}

type createBathhouseRequest struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Address      string                `json:"address"`
	CityID       int64                 `json:"city_id"`
	Latitude     float64               `json:"latitude"`
	Longitude    float64               `json:"longitude"`
	PricePerHour int64                 `json:"price_per_hour"`
	MinDuration  int                   `json:"min_duration"`
	MaxGuests    int                   `json:"max_guests"`
	HasPool      bool                  `json:"has_pool"`
	HasSauna     bool                  `json:"has_sauna"`
	HasSteamRoom bool                  `json:"has_steam_room"`
	HasHotTub    bool                  `json:"has_hot_tub"`
	HasBBQ       bool                  `json:"has_bbq"`
	HasKaraoke   bool                  `json:"has_karaoke"`
	Images       []string              `json:"images"`
	WorkingHours []workingHoursRequest `json:"working_hours"`
}

type workingHoursRequest struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
}

type updateBathhouseRequest struct {
	Name         *string               `json:"name"`
	Description  *string               `json:"description"`
	Address      *string               `json:"address"`
	CityID       *int64                `json:"city_id"`
	Latitude     *float64              `json:"latitude"`
	Longitude    *float64              `json:"longitude"`
	PricePerHour *int64                `json:"price_per_hour"`
	MinDuration  *int                  `json:"min_duration"`
	MaxGuests    *int                  `json:"max_guests"`
	HasPool      *bool                 `json:"has_pool"`
	HasSauna     *bool                 `json:"has_sauna"`
	HasSteamRoom *bool                 `json:"has_steam_room"`
	HasHotTub    *bool                 `json:"has_hot_tub"`
	HasBBQ       *bool                 `json:"has_bbq"`
	HasKaraoke   *bool                 `json:"has_karaoke"`
	Images       []string              `json:"images"`
	WorkingHours []workingHoursRequest `json:"working_hours"`
}

// @Summary      Search bathhouses
// @Description  Search and filter bathhouses with pagination. Supports geo-search, amenity filters, availability checks.
// @Tags         bathhouses
// @Produce      json
// @Param        page               query   int     false  "Page number"              default(1)
// @Param        page_size          query   int     false  "Items per page"           default(20)
// @Param        sort_by            query   string  false  "Sort field (relevance, price_asc, price_desc, rating, distance, newest)"
// @Param        sort_order         query   string  false  "Sort order (asc, desc)"
// @Param        city_id            query   int     false  "Filter by city ID"
// @Param        city_slug          query   string  false  "Filter by city slug"
// @Param        price_min          query   int     false  "Minimum price in kopecks"
// @Param        price_max          query   int     false  "Maximum price in kopecks"
// @Param        min_guests         query   int     false  "Minimum guest capacity"
// @Param        has_pool           query   bool    false  "Has pool"
// @Param        has_sauna          query   bool    false  "Has sauna"
// @Param        has_steam_room     query   bool    false  "Has steam room"
// @Param        has_hot_tub        query   bool    false  "Has hot tub"
// @Param        has_bbq            query   bool    false  "Has BBQ"
// @Param        has_karaoke        query   bool    false  "Has karaoke"
// @Param        min_rating         query   number  false  "Minimum rating"
// @Param        lat                query   number  false  "Latitude for geo-search"
// @Param        lng                query   number  false  "Longitude for geo-search"
// @Param        radius_km          query   number  false  "Search radius in km"
// @Param        guest_count        query   int     false  "Number of guests"
// @Param        available_date     query   string  false  "Check availability date (YYYY-MM-DD)"
// @Param        available_time_from query  string  false  "Available from time (HH:MM)"
// @Param        available_time_to  query   string  false  "Available to time (HH:MM)"
// @Param        open_now           query   bool    false  "Only open now"
// @Param        q                  query   string  false  "Search query"
// @Success      200  {object}  APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure      500  {object}  APIResponse{error=APIError}
// @Router       /bathhouses [get]
func (h *BathhouseHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.BathhouseFilter{
		Page:      getPage(q.Get("page")),
		PageSize:  getPageSize(q.Get("page_size"), 20),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
	}

	// Only show active bathhouses for public search
	activeStatus := domain.BathhouseStatusActive
	filter.Status = &activeStatus

	if v := q.Get("city_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.CityID = &id
		}
	}
	if v := q.Get("city_slug"); v != "" {
		filter.CitySlug = &v
	}
	if v := q.Get("price_min"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.PriceMin = &val
		}
	}
	if v := q.Get("price_max"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.PriceMax = &val
		}
	}
	if v := q.Get("min_guests"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			filter.MinGuests = &val
		}
	}
	if v := q.Get("has_pool"); v != "" {
		val := v == "true"
		filter.HasPool = &val
	}
	if v := q.Get("has_sauna"); v != "" {
		val := v == "true"
		filter.HasSauna = &val
	}
	if v := q.Get("has_steam_room"); v != "" {
		val := v == "true"
		filter.HasSteamRoom = &val
	}
	if v := q.Get("has_hot_tub"); v != "" {
		val := v == "true"
		filter.HasHotTub = &val
	}
	if v := q.Get("has_bbq"); v != "" {
		val := v == "true"
		filter.HasBBQ = &val
	}
	if v := q.Get("has_karaoke"); v != "" {
		val := v == "true"
		filter.HasKaraoke = &val
	}
	if v := q.Get("min_rating"); v != "" {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinRating = &val
		}
	}
	if v := q.Get("lat"); v != "" {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			filter.Latitude = &val
		}
	}
	if v := q.Get("lng"); v != "" {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			filter.Longitude = &val
		}
	}
	if v := q.Get("radius_km"); v != "" {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			filter.RadiusKm = &val
		}
	}
	if v := q.Get("guest_count"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			filter.GuestCount = &val
		}
	}
	if v := q.Get("available_date"); v != "" {
		if d, err := time.Parse("2006-01-02", v); err == nil {
			filter.AvailableDate = &d
		}
	}
	if v := q.Get("available_time_from"); v != "" {
		if domain.IsValidTimeFormat(v) {
			filter.AvailableTimeFrom = &v
		}
	}
	if v := q.Get("available_time_to"); v != "" {
		if domain.IsValidTimeFormat(v) {
			filter.AvailableTimeTo = &v
		}
	}
	if v := q.Get("open_now"); v != "" {
		val := v == "true"
		filter.OpenNow = &val
	}
	if v := q.Get("q"); v != "" {
		filter.SearchQuery = &v
	}

	result, err := h.bathhouseService.Search(r.Context(), filter)
	if err == nil && filter.SearchQuery != nil && h.suggestionService != nil {
		_ = h.suggestionService.RecordQuery(r.Context(), *filter.SearchQuery)
	}
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBathhouseResponse(&result.Items[i])
	}

	userID := middleware.GetUserID(r.Context())
	if userID != uuid.Nil && h.favoriteService != nil {
		for i := range result.Items {
			fav, err := h.favoriteService.IsFavorite(r.Context(), userID, result.Items[i].ID)
			if err == nil {
				items[i].IsFavorite = fav
			}
		}
	}

	// Record impressions for promoted bathhouses
	if h.promotionService != nil {
		for _, bh := range result.Items {
			if bh.IsPromoted {
				_ = h.promotionService.RecordImpression(r.Context(), bh.ID)
			}
		}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// @Summary      Search bathhouses by city slug
// @Description  Get bathhouses in a city identified by its slug. SEO-friendly endpoint.
// @Tags         bathhouses
// @Produce      json
// @Param        slug       path    string  true   "City slug"
// @Param        page       query   int     false  "Page number"     default(1)
// @Param        page_size  query   int     false  "Items per page"  default(20)
// @Success      200  {object}  APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      500  {object}  APIResponse{error=APIError}
// @Router       /cities/{slug}/bathhouses [get]
func (h *BathhouseHandler) SearchByCitySlug(w http.ResponseWriter, r *http.Request) {
	citySlug := chi.URLParam(r, "slug")
	if citySlug == "" || !seo.IsValidSlug(citySlug) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city slug")
		return
	}

	q := r.URL.Query()
	filter := domain.BathhouseFilter{
		Page:     getPage(q.Get("page")),
		PageSize: getPageSize(q.Get("page_size"), 20),
		CitySlug: &citySlug,
	}

	activeStatus := domain.BathhouseStatusActive
	filter.Status = &activeStatus

	result, err := h.bathhouseService.Search(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBathhouseResponse(&result.Items[i])
	}

	userID := middleware.GetUserID(r.Context())
	if userID != uuid.Nil && h.favoriteService != nil {
		for i := range result.Items {
			fav, err := h.favoriteService.IsFavorite(r.Context(), userID, result.Items[i].ID)
			if err == nil {
				items[i].IsFavorite = fav
			}
		}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// @Summary      Get bathhouse by ID
// @Description  Get detailed bathhouse information including meta tags and gallery preview.
// @Tags         bathhouses
// @Produce      json
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=bathhouseResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id} [get]
func (h *BathhouseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	bh, err := h.bathhouseService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := toBathhouseResponse(bh)
	resp.Meta = h.buildMeta(r.Context(), bh)

	userID := middleware.GetUserID(r.Context())
	if userID != uuid.Nil && h.favoriteService != nil {
		fav, err := h.favoriteService.IsFavorite(r.Context(), userID, bh.ID)
		if err == nil {
			resp.IsFavorite = fav
		}
	}

	h.recordBathhouseView(r, bh.ID, userID)

	// Load gallery preview (first 4 photos from reviews)
	if h.mediaService != nil {
		gallery, err := h.mediaService.ListByBathhouse(r.Context(), bh.ID, 1, 4)
		if err == nil && len(gallery.Items) > 0 {
			resp.GalleryPreview = toMediaResponses(gallery.Items)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Get bathhouse by slug
// @Description  Get detailed bathhouse information by its URL-friendly slug.
// @Tags         bathhouses
// @Produce      json
// @Param        slug  path      string  true  "Bathhouse slug"
// @Success      200   {object}  APIResponse{data=bathhouseResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /bathhouses/by-slug/{slug} [get]
func (h *BathhouseHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" || !seo.IsValidSlug(slug) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid slug format")
		return
	}

	bh, err := h.bathhouseService.GetBySlug(r.Context(), slug)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := toBathhouseResponse(bh)
	resp.Meta = h.buildMeta(r.Context(), bh)

	userID := middleware.GetUserID(r.Context())
	if userID != uuid.Nil && h.favoriteService != nil {
		fav, err := h.favoriteService.IsFavorite(r.Context(), userID, bh.ID)
		if err == nil {
			resp.IsFavorite = fav
		}
	}

	h.recordBathhouseView(r, bh.ID, userID)

	// Load gallery preview (first 4 photos from reviews)
	if h.mediaService != nil {
		gallery, err := h.mediaService.ListByBathhouse(r.Context(), bh.ID, 1, 4)
		if err == nil && len(gallery.Items) > 0 {
			resp.GalleryPreview = toMediaResponses(gallery.Items)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Create bathhouse
// @Description  Create a new bathhouse. Only owners can create bathhouses.
// @Tags         bathhouses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createBathhouseRequest  true  "Bathhouse data"
// @Success      201   {object}  APIResponse{data=bathhouseResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /bathhouses [post]
func (h *BathhouseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBathhouseRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	ownerID := middleware.GetUserID(r.Context())

	wh := make([]domain.WorkingHours, len(req.WorkingHours))
	for i, h := range req.WorkingHours {
		wh[i] = domain.WorkingHours{
			DayOfWeek: h.DayOfWeek,
			OpenTime:  h.OpenTime,
			CloseTime: h.CloseTime,
		}
	}

	bh, err := h.bathhouseService.Create(r.Context(), ownerID, service.CreateBathhouseInput{
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		CityID:       req.CityID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		PricePerHour: req.PricePerHour,
		MinDuration:  req.MinDuration,
		MaxGuests:    req.MaxGuests,
		HasPool:      req.HasPool,
		HasSauna:     req.HasSauna,
		HasSteamRoom: req.HasSteamRoom,
		HasHotTub:    req.HasHotTub,
		HasBBQ:       req.HasBBQ,
		HasKaraoke:   req.HasKaraoke,
		Images:       req.Images,
		WorkingHours: wh,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBathhouseResponse(bh))
}

// @Summary      Update bathhouse
// @Description  Update bathhouse details. Owner or representative only.
// @Tags         bathhouses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Bathhouse ID (UUID)"
// @Param        body  body      updateBathhouseRequest  true  "Fields to update"
// @Success      200   {object}  APIResponse{data=bathhouseResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id} [put]
func (h *BathhouseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req updateBathhouseRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var wh []domain.WorkingHours
	if req.WorkingHours != nil {
		wh = make([]domain.WorkingHours, len(req.WorkingHours))
		for i, h := range req.WorkingHours {
			wh[i] = domain.WorkingHours{
				DayOfWeek: h.DayOfWeek,
				OpenTime:  h.OpenTime,
				CloseTime: h.CloseTime,
			}
		}
	}

	bh, err := h.bathhouseService.Update(r.Context(), userID, role, id, service.UpdateBathhouseInput{
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		CityID:       req.CityID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		PricePerHour: req.PricePerHour,
		MinDuration:  req.MinDuration,
		MaxGuests:    req.MaxGuests,
		HasPool:      req.HasPool,
		HasSauna:     req.HasSauna,
		HasSteamRoom: req.HasSteamRoom,
		HasHotTub:    req.HasHotTub,
		HasBBQ:       req.HasBBQ,
		HasKaraoke:   req.HasKaraoke,
		Images:       req.Images,
		WorkingHours: wh,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBathhouseResponse(bh))
}

// @Summary      Delete bathhouse
// @Description  Delete a bathhouse. Only the owner can delete.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Failure      409  {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id} [delete]
func (h *BathhouseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	ownerID := middleware.GetUserID(r.Context())
	if err := h.bathhouseService.Delete(r.Context(), ownerID, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// @Summary      Get available time slots
// @Description  Get available booking time slots for a bathhouse on a specific date.
// @Tags         bathhouses
// @Produce      json
// @Param        id    path      string  true  "Bathhouse ID (UUID)"
// @Param        date  query     string  true  "Date in YYYY-MM-DD format"
// @Success      200   {object}  APIResponse{data=[]service.TimeSlot}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/available-slots [get]
func (h *BathhouseHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date parameter is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date format, use YYYY-MM-DD")
		return
	}

	slots, err := h.bookingService.GetAvailableSlots(r.Context(), id, date)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, slots)
}

// @Summary      Get my bathhouses
// @Description  List bathhouses owned by or assigned to the current user.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        page       query   int  false  "Page number"     default(1)
// @Param        page_size  query   int  false  "Items per page"  default(20)
// @Success      200  {object}  APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses [get]
func (h *BathhouseHandler) MyBathhouses(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	if role == domain.RoleAdmin {
		writeError(w, http.StatusBadRequest, "invalid_input", "admin does not have personal bathhouses")
		return
	}

	if role == domain.RoleOwner {
		result, err := h.bathhouseService.ListByOwner(r.Context(), userID, page, pageSize)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		items := make([]bathhouseResponse, len(result.Items))
		for i := range result.Items {
			items[i] = toBathhouseResponse(&result.Items[i])
		}

		writeJSONWithMeta(w, http.StatusOK, items, &Meta{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalCount: result.TotalCount,
			TotalPages: result.TotalPages,
		})
		return
	}

	// Representative
	bathhouses, err := h.representativeService.GetMyBathhouses(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(bathhouses))
	for i := range bathhouses {
		items[i] = toBathhouseResponse(&bathhouses[i])
	}

	total := int64(len(items))
	totalPages := 1
	if int(total) > pageSize {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}
	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       1,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: totalPages,
	})
}

type widgetKeyResponse struct {
	ApiKey string `json:"api_key"`
}

type widgetCodeRequest struct {
	Color      string `json:"color"`
	FontFamily string `json:"font_family"`
	ShowPrice  bool   `json:"show_price"`
	ShowRating bool   `json:"show_rating"`
	Language   string `json:"language"`
}

type widgetCodeResponse struct {
	Code     string `json:"code"`
	ApiKey   string `json:"api_key"`
	ScriptURL string `json:"script_url"`
	StyleURL string `json:"style_url"`
}

// @Summary      Get widget API key
// @Description  Get the widget API key for a bathhouse. Owner or representative only.
// @Tags         widgets
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=widgetKeyResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/widget-key [get]
func (h *BathhouseHandler) GetWidgetKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	apiKey, err := h.bathhouseService.GetWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widgetKeyResponse{ApiKey: apiKey})
}

// @Summary      Regenerate widget API key
// @Description  Generate a new widget API key, invalidating the old one.
// @Tags         widgets
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=widgetKeyResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/widget-key/regenerate [post]
func (h *BathhouseHandler) RegenerateWidgetKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	newKey, err := h.bathhouseService.RegenerateWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widgetKeyResponse{ApiKey: newKey})
}

// @Summary      Get widget embed code
// @Description  Generate HTML embed code for the booking widget with customization options.
// @Tags         widgets
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string  true   "Bathhouse ID (UUID)"
// @Param        color        query     string  false  "Widget accent color (#RRGGBB)"       default(#4CAF50)
// @Param        font_family  query     string  false  "Font family for the widget"
// @Param        show_price   query     bool    false  "Show price in widget"
// @Param        show_rating  query     bool    false  "Show rating in widget"
// @Param        language     query     string  false  "Widget language"                      default(en)
// @Success      200  {object}  APIResponse{data=widgetCodeResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/widget-code [get]
func (h *BathhouseHandler) GetWidgetCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	// Get the API key (this verifies authorization via the service)
	apiKey, err := h.bathhouseService.GetWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Parse optional request parameters
	var req widgetCodeRequest
	req.Color = r.URL.Query().Get("color")
	req.FontFamily = r.URL.Query().Get("font_family")
	req.ShowPrice = r.URL.Query().Get("show_price") == "true"
	req.ShowRating = r.URL.Query().Get("show_rating") == "true"
	req.Language = r.URL.Query().Get("language")
	if req.Language == "" {
		req.Language = "en"
	}

	// Validate parameters
	if req.Color != "" && !isValidColor(req.Color) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid color format")
		return
	}
	if req.FontFamily != "" && !isValidFontFamily(req.FontFamily) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid font family")
		return
	}

	// Set defaults
	if req.Color == "" {
		req.Color = "#4CAF50"
	}
	if req.FontFamily == "" {
		req.FontFamily = "Helvetica, Arial, sans-serif"
	}

	// Generate HTML embed code with absolute URLs
	code := generateWidgetCode(apiKey, req, r)

	// Get API base URL for response
	scheme := "https"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	apiBaseURL := scheme + "://" + host

	writeJSON(w, http.StatusOK, widgetCodeResponse{
		Code:      code,
		ApiKey:    apiKey,
		ScriptURL: apiBaseURL + "/widget.js",
		StyleURL:  apiBaseURL + "/widget.css",
	})
}

func generateWidgetCode(apiKey string, req widgetCodeRequest, r *http.Request) string {
	// Calculate API base URL
	scheme := "https"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	apiBaseURL := scheme + "://" + host

	// Properly escape all user-supplied parameters to prevent injection
	escapedApiKey := html.EscapeString(apiKey)
	escapedColor := html.EscapeString(req.Color)
	escapedFontFamily := html.EscapeString(req.FontFamily)
	escapedLanguage := html.EscapeString(req.Language)
	escapedApiBaseURL := html.EscapeString(apiBaseURL)

	return `<div id="bani-widget" data-api-key="` + escapedApiKey + `" data-color="` + escapedColor + `" data-font-family="` + escapedFontFamily + `" data-language="` + escapedLanguage + `" data-show-price="` + boolToString(req.ShowPrice) + `" data-show-rating="` + boolToString(req.ShowRating) + `"></div>
<link rel="stylesheet" href="` + escapedApiBaseURL + `/widget.css">
<script src="` + escapedApiBaseURL + `/widget.js"></script>
<script>
  window.BANI_WIDGET_API_URL = '` + strings.ReplaceAll(strings.ReplaceAll(escapedApiBaseURL, `\`, `\\`), `'`, `\'`) + `';
</script>`
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// isValidColor validates that a color is in valid hex format (#RRGGBB)
func isValidColor(color string) bool {
	colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	return colorRegex.MatchString(color)
}

// isValidFontFamily validates font family to prevent CSS injection
func isValidFontFamily(fontFamily string) bool {
	// Allow comma-separated font families with basic validation
	// Reject if it contains special CSS characters that could be injection vectors
	forbidden := []string{";", "}", "{", "(", ")", "!", "@"}
	for _, char := range forbidden {
		if strings.Contains(fontFamily, char) {
			return false
		}
	}
	// Ensure it's not too long (reasonable max is 256 chars)
	if len(fontFamily) > 256 {
		return false
	}
	return true
}

// @Summary      Get bathhouse SEO meta tags
// @Description  Get SEO meta tags (title, description, og:image, canonical) for a bathhouse.
// @Tags         bathhouses
// @Produce      json
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=seo.MetaTags}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/meta [get]
func (h *BathhouseHandler) GetMeta(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	bh, err := h.bathhouseService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	meta := h.buildMeta(r.Context(), bh)
	if meta == nil {
		meta = &seo.MetaTags{}
	}

	writeJSON(w, http.StatusOK, meta)
}

func (h *BathhouseHandler) buildMeta(ctx context.Context, bh *domain.Bathhouse) *seo.MetaTags {
	var cityName, citySlug string
	if h.cityService != nil && bh.CityID > 0 {
		city, err := h.cityService.GetByID(ctx, bh.CityID)
		if err == nil && city != nil {
			cityName = city.Name
			citySlug = city.Slug
		}
	}

	meta := seo.GenerateMetaTags(seo.MetaInput{
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
	})
	return &meta
}

// @Summary      Check listing completeness
// @Description  Check how complete a bathhouse listing is before submitting for moderation.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=service.CompletenessResult}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/completeness [get]
func (h *BathhouseHandler) CheckCompleteness(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.bathhouseService.CheckCompleteness(r.Context(), userID, role, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// @Summary      Submit bathhouse for moderation
// @Description  Submit a bathhouse listing for admin moderation. All required fields must be complete.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/submit [post]
func (h *BathhouseHandler) SubmitForModeration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.SubmitForModeration(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submitted for moderation"})
}

// @Summary      Duplicate bathhouse
// @Description  Create a copy of an existing bathhouse listing with draft status.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      201  {object}  APIResponse{data=bathhouseResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/duplicate [post]
func (h *BathhouseHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	bh, err := h.bathhouseService.DuplicateBathhouse(r.Context(), userID, role, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBathhouseResponse(bh))
}

// @Summary      Deactivate bathhouse
// @Description  Temporarily deactivate a bathhouse. It will be hidden from search but existing bookings are kept.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/deactivate [post]
func (h *BathhouseHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.DeactivateBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse deactivated"})
}

// @Summary      Activate bathhouse
// @Description  Restore a temporarily deactivated bathhouse back to active status.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/activate [post]
func (h *BathhouseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.ActivateBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse activated"})
}

// @Summary      Archive bathhouse
// @Description  Permanently archive a bathhouse. Only possible if there are no active bookings. Not reversible via API.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Failure      409  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/archive [delete]
func (h *BathhouseHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.ArchiveBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse archived"})
}

func (h *BathhouseHandler) recordBathhouseView(r *http.Request, bathhouseID uuid.UUID, userID uuid.UUID) {
	ctx := r.Context()

	if userID != uuid.Nil && h.recommendationService != nil {
		if err := h.recommendationService.RecordView(ctx, userID, bathhouseID); err != nil {
			h.log.Warn("failed to record recommendation view", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if h.analyticsService != nil {
		ipHash := getIPHash(r)
		source := domain.ViewSourceDirect
		var viewerID *uuid.UUID
		if userID != uuid.Nil {
			viewerID = &userID
		}
		if err := h.analyticsService.RecordView(ctx, bathhouseID, viewerID, source, ipHash); err != nil {
			h.log.Warn("failed to record analytics view", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if h.promotionService != nil {
		if err := h.promotionService.RecordClick(ctx, bathhouseID); err != nil {
			h.log.Warn("failed to record promotion click", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if userID != uuid.Nil && h.savedSearchService != nil {
		if err := h.savedSearchService.RecordView(ctx, userID, bathhouseID); err != nil {
			h.log.Warn("failed to record recently viewed", "bathhouse_id", bathhouseID, "error", err)
		}
	}
}

const maxPageSize = 100

func getPage(s string) int {
	if s == "" {
		return 1
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return 1
	}
	return v
}

func getPageSize(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	if v > maxPageSize {
		return maxPageSize
	}
	return v
}
