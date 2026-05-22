package handler

import (
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/service"
)

type ComparisonHandler struct {
	bathhouseService service.BathhouseService
}

func NewComparisonHandler(bathhouseService service.BathhouseService) *ComparisonHandler {
	return &ComparisonHandler{
		bathhouseService: bathhouseService,
	}
}

type compareRequest struct {
	IDs       []string `json:"ids"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type comparisonItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Address      string   `json:"address"`
	PricePerHour int64    `json:"price_per_hour"`
	Rating       float64  `json:"rating"`
	ReviewCount  int      `json:"review_count"`
	MaxGuests    int      `json:"max_guests"`
	MinDuration  int      `json:"min_duration"`
	HasPool      bool     `json:"has_pool"`
	HasSauna     bool     `json:"has_sauna"`
	HasSteamRoom bool     `json:"has_steam_room"`
	HasHotTub    bool     `json:"has_hot_tub"`
	HasBBQ       bool     `json:"has_bbq"`
	HasKaraoke   bool     `json:"has_karaoke"`
	Images       []string `json:"images"`
	Distance     *float64 `json:"distance,omitempty"`
	Status       string   `json:"status"`
}

type compareResponse struct {
	Items []comparisonItem `json:"items"`
}

// Compare godoc
//
//	@Summary		Compare bathhouses
//	@Description	Compare 2-3 bathhouses side by side
//	@Tags			bathhouses
//	@Accept			json
//	@Produce		json
//	@Param			request	body		compareRequest	true	"Bathhouse IDs to compare"
//	@Success		200		{object}	APIResponse{data=compareResponse}
//	@Failure		400		{object}	APIResponse
//	@Failure		404		{object}	APIResponse
//	@Router			/bathhouses/compare [post]
func (h *ComparisonHandler) Compare(w http.ResponseWriter, r *http.Request) {
	var req compareRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if len(req.IDs) < 2 || len(req.IDs) > 3 {
		writeError(w, http.StatusBadRequest, "invalid_input", "provide 2 or 3 bathhouse IDs for comparison")
		return
	}

	// Parse and deduplicate IDs
	seen := make(map[uuid.UUID]bool, len(req.IDs))
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse ID: "+idStr)
			return
		}
		if seen[id] {
			writeError(w, http.StatusBadRequest, "invalid_input", "duplicate bathhouse ID: "+idStr)
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}

	items := make([]comparisonItem, 0, len(ids))
	for _, id := range ids {
		bh, err := h.bathhouseService.GetByID(r.Context(), id)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		item := comparisonItem{
			ID:           bh.ID.String(),
			Name:         bh.Name,
			Slug:         bh.Slug,
			Address:      bh.Address,
			PricePerHour: bh.PricePerHour,
			Rating:       bh.Rating,
			ReviewCount:  bh.ReviewCount,
			MaxGuests:    bh.MaxGuests,
			MinDuration:  bh.MinDuration,
			HasPool:      bh.HasPool,
			HasSauna:     bh.HasSauna,
			HasSteamRoom: bh.HasSteamRoom,
			HasHotTub:    bh.HasHotTub,
			HasBBQ:       bh.HasBBQ,
			HasKaraoke:   bh.HasKaraoke,
			Images:       bh.Images,
			Status:       string(bh.Status),
		}
		if item.Images == nil {
			item.Images = []string{}
		}

		if req.Latitude != nil && req.Longitude != nil && bh.Latitude != 0 && bh.Longitude != 0 {
			dist := haversineDistance(*req.Latitude, *req.Longitude, bh.Latitude, bh.Longitude)
			item.Distance = &dist
		}

		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, compareResponse{Items: items})
}

// haversineDistance calculates the distance between two points on Earth in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return math.Round(earthRadiusKm*c*100) / 100
}
