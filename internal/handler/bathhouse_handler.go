package handler

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/seo"
	"github.com/rekurt/relax-hub/internal/service"
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
	reviewService         service.ReviewService
	userService           service.UserService
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
	reviewService service.ReviewService,
	userService service.UserService,
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
		reviewService:         reviewService,
		userService:           userService,
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
	ID                         string             `json:"id"`
	OwnerID                    string             `json:"owner_id"`
	Name                       string             `json:"name"`
	Slug                       string             `json:"slug"`
	Description                string             `json:"description"`
	Address                    string             `json:"address"`
	CityID                     int64              `json:"city_id"`
	Latitude                   float64            `json:"latitude"`
	Longitude                  float64            `json:"longitude"`
	PricePerHour               int64              `json:"price_per_hour"`
	MinDuration                int                `json:"min_duration"`
	MaxGuests                  int                `json:"max_guests"`
	HasPool                    bool               `json:"has_pool"`
	HasSauna                   bool               `json:"has_sauna"`
	HasSteamRoom               bool               `json:"has_steam_room"`
	HasHotTub                  bool               `json:"has_hot_tub"`
	HasBBQ                     bool               `json:"has_bbq"`
	HasKaraoke                 bool               `json:"has_karaoke"`
	LongSessionThresholdHours  int                `json:"long_session_threshold_hours"`
	LongSessionDiscountPercent int                `json:"long_session_discount_percent"`
	BaseCapacity               int                `json:"base_capacity"`
	ExtraGuestSurcharge        int64              `json:"extra_guest_surcharge"`
	LastMinuteEnabled          bool               `json:"last_minute_enabled"`
	LastMinuteDiscountPercent  int                `json:"last_minute_discount_percent,omitempty"`
	LastMinuteHoursThreshold   int                `json:"last_minute_hours_threshold,omitempty"`
	BufferMinutes              int                `json:"buffer_minutes"`
	LeadTimeHours              int                `json:"lead_time_hours"`
	MaxAdvanceDays             int                `json:"max_advance_days"`
	BookingMode                string             `json:"booking_mode"`
	RequestTimeout             int                `json:"request_timeout"`
	CancellationPolicy         string             `json:"cancellation_policy"`
	SecurityDepositPercent     int                `json:"security_deposit_percent"`
	ResponseRate               float64            `json:"response_rate"`
	AvgResponseTimeMinutes     int                `json:"avg_response_time_minutes"`
	Rating                     float64            `json:"rating"`
	BayesianRating             float64            `json:"bayesian_rating"`
	AvgCleanliness             float64            `json:"avg_cleanliness,omitempty"`
	AvgAccuracy                float64            `json:"avg_accuracy,omitempty"`
	AvgCommunication           float64            `json:"avg_communication,omitempty"`
	AvgValueForMoney           float64            `json:"avg_value_for_money,omitempty"`
	ReviewCount                int                `json:"review_count"`
	Images                     []string           `json:"images"`
	WorkingHours               []workingHoursResp `json:"working_hours"`
	Status                     string             `json:"status"`
	IsFavorite                 bool               `json:"is_favorite"`
	IsPromoted                 bool               `json:"is_promoted"`
	IsPhotoVerified            bool               `json:"is_photo_verified"`
	LastMinuteActive           bool               `json:"last_minute_active"`
	Badges                     []string           `json:"badges"`
	AreaAvgPricePerHour        int64                 `json:"area_avg_price_per_hour,omitempty"`
	GalleryPreview             []mediaResponse       `json:"gallery_preview,omitempty"`
	OwnerProfile               *ownerProfileResponse `json:"owner_profile,omitempty"`
	Meta                       *seo.MetaTags         `json:"meta,omitempty"`
	CreatedAt                  time.Time             `json:"created_at"`
	UpdatedAt                  time.Time             `json:"updated_at"`
}

type ownerProfileResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	Rating         float64   `json:"rating"`
	ObjectCount    int       `json:"object_count"`
	MemberSince    time.Time `json:"member_since"`
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
		ID:                         b.ID.String(),
		OwnerID:                    b.OwnerID.String(),
		Name:                       b.Name,
		Slug:                       b.Slug,
		Description:                b.Description,
		Address:                    b.Address,
		CityID:                     b.CityID,
		Latitude:                   b.Latitude,
		Longitude:                  b.Longitude,
		PricePerHour:               b.PricePerHour,
		MinDuration:                b.MinDuration,
		MaxGuests:                  b.MaxGuests,
		HasPool:                    b.HasPool,
		HasSauna:                   b.HasSauna,
		HasSteamRoom:               b.HasSteamRoom,
		HasHotTub:                  b.HasHotTub,
		HasBBQ:                     b.HasBBQ,
		HasKaraoke:                 b.HasKaraoke,
		LongSessionThresholdHours:  b.LongSessionThresholdHours,
		LongSessionDiscountPercent: b.LongSessionDiscountPercent,
		BaseCapacity:               b.BaseCapacity,
		ExtraGuestSurcharge:        b.ExtraGuestSurcharge,
		LastMinuteEnabled:          b.LastMinuteEnabled,
		LastMinuteDiscountPercent:  b.LastMinuteDiscountPercent,
		LastMinuteHoursThreshold:   b.LastMinuteHoursThreshold,
		BufferMinutes:              b.BufferMinutes,
		LeadTimeHours:              b.LeadTimeHours,
		MaxAdvanceDays:             b.MaxAdvanceDays,
		BookingMode:                b.BookingMode,
		RequestTimeout:             b.RequestTimeout,
		CancellationPolicy:         string(b.CancellationPolicy),
		SecurityDepositPercent:     b.SecurityDepositPercent,
		ResponseRate:               b.ResponseRate,
		AvgResponseTimeMinutes:     b.AvgResponseTimeMinutes,
		Rating:                     b.Rating,
		BayesianRating:             b.BayesianRating,
		ReviewCount:                b.ReviewCount,
		Images:                     images,
		WorkingHours:               wh,
		Status:                     string(b.Status),
		IsPromoted:                 b.IsPromoted,
		IsPhotoVerified:            b.IsPhotoVerified,
		CreatedAt:                  b.CreatedAt,
		UpdatedAt:                  b.UpdatedAt,
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
	Name                       *string               `json:"name"`
	Description                *string               `json:"description"`
	Address                    *string               `json:"address"`
	CityID                     *int64                `json:"city_id"`
	Latitude                   *float64              `json:"latitude"`
	Longitude                  *float64              `json:"longitude"`
	PricePerHour               *int64                `json:"price_per_hour"`
	MinDuration                *int                  `json:"min_duration"`
	MaxGuests                  *int                  `json:"max_guests"`
	HasPool                    *bool                 `json:"has_pool"`
	HasSauna                   *bool                 `json:"has_sauna"`
	HasSteamRoom               *bool                 `json:"has_steam_room"`
	HasHotTub                  *bool                 `json:"has_hot_tub"`
	HasBBQ                     *bool                 `json:"has_bbq"`
	HasKaraoke                 *bool                 `json:"has_karaoke"`
	LongSessionThresholdHours  *int                  `json:"long_session_threshold_hours"`
	LongSessionDiscountPercent *int                  `json:"long_session_discount_percent"`
	BaseCapacity               *int                  `json:"base_capacity"`
	ExtraGuestSurcharge        *int64                `json:"extra_guest_surcharge"`
	LastMinuteEnabled          *bool                 `json:"last_minute_enabled"`
	LastMinuteDiscountPercent  *int                  `json:"last_minute_discount_percent"`
	LastMinuteHoursThreshold   *int                  `json:"last_minute_hours_threshold"`
	BufferMinutes              *int                  `json:"buffer_minutes"`
	LeadTimeHours              *int                  `json:"lead_time_hours"`
	MaxAdvanceDays             *int                  `json:"max_advance_days"`
	BookingMode                *string               `json:"booking_mode"`
	RequestTimeout             *int                  `json:"request_timeout"`
	CancellationPolicy         *string               `json:"cancellation_policy"`
	SecurityDepositPercent     *int                  `json:"security_deposit_percent"`
	Images                     []string              `json:"images"`
	WorkingHours               []workingHoursRequest `json:"working_hours"`
}

// @Summary		Search bathhouses
// @Description	Search and filter bathhouses with pagination. Supports geo-search, amenity filters, availability checks.
// @Tags			bathhouses
// @Produce		json
// @Param			page				query		int		false	"Page number"		default(1)
// @Param			page_size			query		int		false	"Items per page"	default(20)
// @Param			sort_by				query		string	false	"Sort field (relevance, price_asc, price_desc, rating, distance, newest)"
// @Param			sort_order			query		string	false	"Sort order (asc, desc)"
// @Param			city_id				query		int		false	"Filter by city ID"
// @Param			city_slug			query		string	false	"Filter by city slug"
// @Param			price_min			query		int		false	"Minimum price in kopecks"
// @Param			price_max			query		int		false	"Maximum price in kopecks"
// @Param			min_guests			query		int		false	"Minimum guest capacity"
// @Param			has_pool			query		bool	false	"Has pool"
// @Param			has_sauna			query		bool	false	"Has sauna"
// @Param			has_steam_room		query		bool	false	"Has steam room"
// @Param			has_hot_tub			query		bool	false	"Has hot tub"
// @Param			has_bbq				query		bool	false	"Has BBQ"
// @Param			has_karaoke			query		bool	false	"Has karaoke"
// @Param			min_rating			query		number	false	"Minimum rating"
// @Param			lat					query		number	false	"Latitude for geo-search"
// @Param			lng					query		number	false	"Longitude for geo-search"
// @Param			radius_km			query		number	false	"Search radius in km"
// @Param			guest_count			query		int		false	"Number of guests"
// @Param			available_date		query		string	false	"Check availability date (YYYY-MM-DD)"
// @Param			available_time_from	query		string	false	"Available from time (HH:MM)"
// @Param			available_time_to	query		string	false	"Available to time (HH:MM)"
// @Param			open_now			query		bool	false	"Only open now"
// @Param			q					query		string	false	"Search query"
// @Success		200					{object}	APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure		500					{object}	APIResponse{error=APIError}
// @Router			/bathhouses [get]
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
	if v := q.Get("last_minute"); v != "" {
		val := v == "true"
		filter.LastMinute = &val
	}
	if v := q.Get("q"); v != "" {
		filter.SearchQuery = &v
	}
	if v := q.Get("isochrone_wkt"); v != "" {
		filter.IsochroneWKT = &v
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

	// Compute badges and last-minute status for each bathhouse
	for i := range result.Items {
		badges := h.bathhouseService.ComputeBadges(r.Context(), &result.Items[i])
		if badges == nil {
			badges = []string{}
		}
		items[i].Badges = badges
		items[i].LastMinuteActive = h.bathhouseService.IsLastMinuteActive(&result.Items[i])
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

// @Summary		Search bathhouses by city slug
// @Description	Get bathhouses in a city identified by its slug. SEO-friendly endpoint.
// @Tags			bathhouses
// @Produce		json
// @Param			slug		path		string	true	"City slug"
// @Param			page		query		int		false	"Page number"		default(1)
// @Param			page_size	query		int		false	"Items per page"	default(20)
// @Success		200			{object}	APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure		400			{object}	APIResponse{error=APIError}
// @Failure		500			{object}	APIResponse{error=APIError}
// @Router			/cities/{slug}/bathhouses [get]
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

	// Compute badges for each bathhouse
	for i := range result.Items {
		badges := h.bathhouseService.ComputeBadges(r.Context(), &result.Items[i])
		if badges == nil {
			badges = []string{}
		}
		items[i].Badges = badges
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// @Summary		Get bathhouse by ID
// @Description	Get detailed bathhouse information including meta tags and gallery preview.
// @Tags			bathhouses
// @Produce		json
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=bathhouseResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/bathhouses/{id} [get]
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

	// Load review criteria averages
	if h.reviewService != nil {
		avgs, err := h.reviewService.GetCriteriaAverages(r.Context(), bh.ID)
		if err == nil {
			resp.AvgCleanliness = avgs.AvgCleanliness
			resp.AvgAccuracy = avgs.AvgAccuracy
			resp.AvgCommunication = avgs.AvgCommunication
			resp.AvgValueForMoney = avgs.AvgValueForMoney
		}
	}

	// Compute badges and last-minute status
	badges := h.bathhouseService.ComputeBadges(r.Context(), bh)
	if badges == nil {
		badges = []string{}
	}
	resp.Badges = badges
	resp.LastMinuteActive = h.bathhouseService.IsLastMinuteActive(bh)

	// Area average price for price context
	if avgPrice, err := h.bathhouseService.GetAreaAvgPrice(r.Context(), bh.CityID, bh.Latitude, bh.Longitude); err == nil && avgPrice > 0 {
		resp.AreaAvgPricePerHour = avgPrice
	}

	// Owner profile section
	resp.OwnerProfile = h.buildOwnerProfile(r.Context(), bh.OwnerID)

	writeJSON(w, http.StatusOK, resp)
}

// @Summary		Get bathhouse by slug
// @Description	Get detailed bathhouse information by its URL-friendly slug.
// @Tags			bathhouses
// @Produce		json
// @Param			slug	path		string	true	"Bathhouse slug"
// @Success		200		{object}	APIResponse{data=bathhouseResponse}
// @Failure		400		{object}	APIResponse{error=APIError}
// @Failure		404		{object}	APIResponse{error=APIError}
// @Router			/bathhouses/by-slug/{slug} [get]
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

	// Load review criteria averages
	if h.reviewService != nil {
		avgs, err := h.reviewService.GetCriteriaAverages(r.Context(), bh.ID)
		if err == nil {
			resp.AvgCleanliness = avgs.AvgCleanliness
			resp.AvgAccuracy = avgs.AvgAccuracy
			resp.AvgCommunication = avgs.AvgCommunication
			resp.AvgValueForMoney = avgs.AvgValueForMoney
		}
	}

	// Compute badges and last-minute status
	badges := h.bathhouseService.ComputeBadges(r.Context(), bh)
	if badges == nil {
		badges = []string{}
	}
	resp.Badges = badges
	resp.LastMinuteActive = h.bathhouseService.IsLastMinuteActive(bh)

	// Area average price for price context
	if avgPrice, err := h.bathhouseService.GetAreaAvgPrice(r.Context(), bh.CityID, bh.Latitude, bh.Longitude); err == nil && avgPrice > 0 {
		resp.AreaAvgPricePerHour = avgPrice
	}

	// Owner profile section
	resp.OwnerProfile = h.buildOwnerProfile(r.Context(), bh.OwnerID)

	writeJSON(w, http.StatusOK, resp)
}

// @Summary		Create bathhouse
// @Description	Create a new bathhouse. Only owners can create bathhouses.
// @Tags			bathhouses
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			body	body		createBathhouseRequest	true	"Bathhouse data"
// @Success		201		{object}	APIResponse{data=bathhouseResponse}
// @Failure		400		{object}	APIResponse{error=APIError}
// @Failure		401		{object}	APIResponse{error=APIError}
// @Failure		403		{object}	APIResponse{error=APIError}
// @Router			/bathhouses [post]
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

// @Summary		Update bathhouse
// @Description	Update bathhouse details. Owner or representative only.
// @Tags			bathhouses
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string					true	"Bathhouse ID (UUID)"
// @Param			body	body		updateBathhouseRequest	true	"Fields to update"
// @Success		200		{object}	APIResponse{data=bathhouseResponse}
// @Failure		400		{object}	APIResponse{error=APIError}
// @Failure		401		{object}	APIResponse{error=APIError}
// @Failure		403		{object}	APIResponse{error=APIError}
// @Failure		404		{object}	APIResponse{error=APIError}
// @Router			/bathhouses/{id} [put]
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
		Name:                       req.Name,
		Description:                req.Description,
		Address:                    req.Address,
		CityID:                     req.CityID,
		Latitude:                   req.Latitude,
		Longitude:                  req.Longitude,
		PricePerHour:               req.PricePerHour,
		MinDuration:                req.MinDuration,
		MaxGuests:                  req.MaxGuests,
		HasPool:                    req.HasPool,
		HasSauna:                   req.HasSauna,
		HasSteamRoom:               req.HasSteamRoom,
		HasHotTub:                  req.HasHotTub,
		HasBBQ:                     req.HasBBQ,
		HasKaraoke:                 req.HasKaraoke,
		LongSessionThresholdHours:  req.LongSessionThresholdHours,
		LongSessionDiscountPercent: req.LongSessionDiscountPercent,
		BaseCapacity:               req.BaseCapacity,
		ExtraGuestSurcharge:        req.ExtraGuestSurcharge,
		LastMinuteEnabled:          req.LastMinuteEnabled,
		LastMinuteDiscountPercent:  req.LastMinuteDiscountPercent,
		LastMinuteHoursThreshold:   req.LastMinuteHoursThreshold,
		BufferMinutes:              req.BufferMinutes,
		LeadTimeHours:              req.LeadTimeHours,
		MaxAdvanceDays:             req.MaxAdvanceDays,
		BookingMode:                req.BookingMode,
		RequestTimeout:             req.RequestTimeout,
		CancellationPolicy:         req.CancellationPolicy,
		SecurityDepositPercent:     req.SecurityDepositPercent,
		Images:                     req.Images,
		WorkingHours:               wh,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBathhouseResponse(bh))
}

// @Summary		Delete bathhouse
// @Description	Delete a bathhouse. Only the owner can delete.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Failure		409	{object}	APIResponse{error=APIError}
// @Router			/bathhouses/{id} [delete]
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

// @Summary		Get available time slots
// @Description	Get available booking time slots for a bathhouse on a specific date.
// @Tags			bathhouses
// @Produce		json
// @Param			id		path		string	true	"Bathhouse ID (UUID)"
// @Param			date	query		string	true	"Date in YYYY-MM-DD format"
// @Success		200		{object}	APIResponse{data=[]service.TimeSlot}
// @Failure		400		{object}	APIResponse{error=APIError}
// @Failure		404		{object}	APIResponse{error=APIError}
// @Router			/bathhouses/{id}/available-slots [get]
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

// @Summary		Get my bathhouses
// @Description	List bathhouses owned by or assigned to the current user.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			page		query		int	false	"Page number"		default(1)
// @Param			page_size	query		int	false	"Items per page"	default(20)
// @Success		200			{object}	APIResponse{data=[]bathhouseResponse,meta=Meta}
// @Failure		400			{object}	APIResponse{error=APIError}
// @Failure		401			{object}	APIResponse{error=APIError}
// @Failure		403			{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses [get]
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

