package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type RecommendationHandler struct {
	recommendationService service.RecommendationService
	bathhouseService      service.BathhouseService
}

func NewRecommendationHandler(
	recommendationService service.RecommendationService,
	bathhouseService service.BathhouseService,
) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: recommendationService,
		bathhouseService:      bathhouseService,
	}
}

type recommendationResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Address      string  `json:"address"`
	CityID       int64   `json:"city_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	PricePerHour int64   `json:"price_per_hour"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
}

type userPreferencesResponse struct {
	PreferredCityID *int64 `json:"preferred_city_id,omitempty"`
	PriceRangeMin   *int64 `json:"price_range_min,omitempty"`
	PriceRangeMax   *int64 `json:"price_range_max,omitempty"`
	PreferPool      bool   `json:"prefer_pool"`
	PreferSauna     bool   `json:"prefer_sauna"`
	PreferSteamRoom bool   `json:"prefer_steam_room"`
	PreferHotTub    bool   `json:"prefer_hot_tub"`
	PreferBBQ       bool   `json:"prefer_bbq"`
	PreferKaraoke   bool   `json:"prefer_karaoke"`
}

type updateRecommendationPreferencesRequest struct {
	PreferredCityID *int64 `json:"preferred_city_id,omitempty"`
	PriceRangeMin   *int64 `json:"price_range_min,omitempty"`
	PriceRangeMax   *int64 `json:"price_range_max,omitempty"`
	PreferPool      *bool  `json:"prefer_pool,omitempty"`
	PreferSauna     *bool  `json:"prefer_sauna,omitempty"`
	PreferSteamRoom *bool  `json:"prefer_steam_room,omitempty"`
	PreferHotTub    *bool  `json:"prefer_hot_tub,omitempty"`
	PreferBBQ       *bool  `json:"prefer_bbq,omitempty"`
	PreferKaraoke   *bool  `json:"prefer_karaoke,omitempty"`
}

// GetPersonalized returns personalized recommendations for the authenticated user
func (h *RecommendationHandler) GetPersonalized(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	q := r.URL.Query()
	page := getPage(q.Get("page"))
	pageSize := getPageSize(q.Get("page_size"), 20)

	recommendations, totalCount, err := h.recommendationService.GetPersonalized(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]recommendationResponse, 0, len(recommendations))
	for _, recID := range recommendations {
		bh, err := h.bathhouseService.GetByID(r.Context(), recID)
		if err != nil {
			continue
		}
		items = append(items, recommendationResponse{
			ID:           bh.ID.String(),
			Name:         bh.Name,
			Description:  bh.Description,
			Address:      bh.Address,
			CityID:       bh.CityID,
			Latitude:     bh.Latitude,
			Longitude:    bh.Longitude,
			PricePerHour: bh.PricePerHour,
			Rating:       bh.Rating,
			ReviewCount:  bh.ReviewCount,
		})
	}

	totalPages := (int(totalCount) + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

// GetSimilar returns bathhouses similar to the specified bathhouse
func (h *RecommendationHandler) GetSimilar(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	q := r.URL.Query()
	limit := 10
	if v := q.Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	recommendations, err := h.recommendationService.GetSimilar(r.Context(), bathhouseID, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]recommendationResponse, 0, len(recommendations))
	for _, recID := range recommendations {
		bh, err := h.bathhouseService.GetByID(r.Context(), recID)
		if err != nil {
			continue
		}
		items = append(items, recommendationResponse{
			ID:           bh.ID.String(),
			Name:         bh.Name,
			Description:  bh.Description,
			Address:      bh.Address,
			CityID:       bh.CityID,
			Latitude:     bh.Latitude,
			Longitude:    bh.Longitude,
			PricePerHour: bh.PricePerHour,
			Rating:       bh.Rating,
			ReviewCount:  bh.ReviewCount,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

// GetPopular returns popular bathhouses in a city
func (h *RecommendationHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cityIDStr := q.Get("city_id")
	if cityIDStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "city_id query parameter required")
		return
	}

	cityID, err := strconv.ParseInt(cityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city_id")
		return
	}

	limit := 10
	if v := q.Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	recommendations, err := h.recommendationService.GetPopular(r.Context(), cityID, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]recommendationResponse, 0, len(recommendations))
	for _, recID := range recommendations {
		bh, err := h.bathhouseService.GetByID(r.Context(), recID)
		if err != nil {
			continue
		}
		items = append(items, recommendationResponse{
			ID:           bh.ID.String(),
			Name:         bh.Name,
			Description:  bh.Description,
			Address:      bh.Address,
			CityID:       bh.CityID,
			Latitude:     bh.Latitude,
			Longitude:    bh.Longitude,
			PricePerHour: bh.PricePerHour,
			Rating:       bh.Rating,
			ReviewCount:  bh.ReviewCount,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

// GetPreferences returns the user's current preferences
func (h *RecommendationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	prefs, err := h.recommendationService.GetUserPreferences(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := userPreferencesResponse{
		PreferredCityID: prefs.PreferredCityID,
		PriceRangeMin:   prefs.PriceRangeMin,
		PriceRangeMax:   prefs.PriceRangeMax,
		PreferPool:      prefs.PreferPool,
		PreferSauna:     prefs.PreferSauna,
		PreferSteamRoom: prefs.PreferSteamRoom,
		PreferHotTub:    prefs.PreferHotTub,
		PreferBBQ:       prefs.PreferBBQ,
		PreferKaraoke:   prefs.PreferKaraoke,
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdatePreferences updates the user's preferences
func (h *RecommendationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	defer r.Body.Close()
	var req updateRecommendationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	// Build preferences object
	prefs := &domain.UserPreferences{
		UserID: userID,
	}

	// Try to get existing preferences to preserve unchanged fields
	existing, err := h.recommendationService.GetUserPreferences(r.Context(), userID)
	if err == nil && existing != nil {
		// Copy existing values
		prefs.PreferredCityID = existing.PreferredCityID
		prefs.PriceRangeMin = existing.PriceRangeMin
		prefs.PriceRangeMax = existing.PriceRangeMax
		prefs.PreferPool = existing.PreferPool
		prefs.PreferSauna = existing.PreferSauna
		prefs.PreferSteamRoom = existing.PreferSteamRoom
		prefs.PreferHotTub = existing.PreferHotTub
		prefs.PreferBBQ = existing.PreferBBQ
		prefs.PreferKaraoke = existing.PreferKaraoke
	}

	// Update with provided values
	if req.PreferredCityID != nil {
		prefs.PreferredCityID = req.PreferredCityID
	}
	if req.PriceRangeMin != nil {
		prefs.PriceRangeMin = req.PriceRangeMin
	}
	if req.PriceRangeMax != nil {
		prefs.PriceRangeMax = req.PriceRangeMax
	}
	if req.PreferPool != nil {
		prefs.PreferPool = *req.PreferPool
	}
	if req.PreferSauna != nil {
		prefs.PreferSauna = *req.PreferSauna
	}
	if req.PreferSteamRoom != nil {
		prefs.PreferSteamRoom = *req.PreferSteamRoom
	}
	if req.PreferHotTub != nil {
		prefs.PreferHotTub = *req.PreferHotTub
	}
	if req.PreferBBQ != nil {
		prefs.PreferBBQ = *req.PreferBBQ
	}
	if req.PreferKaraoke != nil {
		prefs.PreferKaraoke = *req.PreferKaraoke
	}

	err = h.recommendationService.UpdatePreferences(r.Context(), userID, prefs)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	updated, err := h.recommendationService.GetUserPreferences(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := userPreferencesResponse{
		PreferredCityID: updated.PreferredCityID,
		PriceRangeMin:   updated.PriceRangeMin,
		PriceRangeMax:   updated.PriceRangeMax,
		PreferPool:      updated.PreferPool,
		PreferSauna:     updated.PreferSauna,
		PreferSteamRoom: updated.PreferSteamRoom,
		PreferHotTub:    updated.PreferHotTub,
		PreferBBQ:       updated.PreferBBQ,
		PreferKaraoke:   updated.PreferKaraoke,
	}

	writeJSON(w, http.StatusOK, resp)
}
