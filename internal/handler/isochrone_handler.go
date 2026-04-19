package handler

import (
	"net/http"
	"strconv"

	"github.com/rekurt/relax-hub/internal/geo"
)

// IsochroneHandler handles isochrone-related endpoints.
type IsochroneHandler struct {
	isochroneService *geo.IsochroneService
}

// NewIsochroneHandler creates a new IsochroneHandler.
func NewIsochroneHandler(isochroneService *geo.IsochroneService) *IsochroneHandler {
	return &IsochroneHandler{isochroneService: isochroneService}
}

type isochroneResponse struct {
	Coordinates [][]float64 `json:"coordinates"` // [lng, lat] pairs
	Mode        string      `json:"mode"`
	Minutes     int         `json:"minutes"`
	CenterLat   float64     `json:"center_lat"`
	CenterLng   float64     `json:"center_lng"`
	WKT         string      `json:"wkt"` // WKT polygon for use with search filter
}

// GetIsochrone godoc
//
//	@Summary		Get isochrone polygon
//	@Description	Returns a polygon representing the area reachable within the specified travel time from a point
//	@Tags			geo
//	@Produce		json
//	@Param			lat		query		number	true	"Latitude of the origin point"
//	@Param			lon		query		number	true	"Longitude of the origin point"
//	@Param			mode	query		string	true	"Travel mode: car or transit"
//	@Param			minutes	query		int		true	"Travel time in minutes (1-60)"
//	@Success		200		{object}	APIResponse{data=isochroneResponse}
//	@Failure		400		{object}	APIResponse
//	@Router			/isochrone [get]
func (h *IsochroneHandler) GetIsochrone(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	lat, err := strconv.ParseFloat(q.Get("lat"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "lat is required and must be a number")
		return
	}
	lon, err := strconv.ParseFloat(q.Get("lon"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "lon is required and must be a number")
		return
	}

	modeStr := q.Get("mode")
	if modeStr == "" {
		modeStr = "car"
	}
	mode := geo.TravelMode(modeStr)
	if mode != geo.TravelModeCar && mode != geo.TravelModeTransit {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "mode must be 'car' or 'transit'")
		return
	}

	minutes := 15
	if v := q.Get("minutes"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			minutes = parsed
		}
	}
	if minutes < 1 || minutes > 60 {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "minutes must be between 1 and 60")
		return
	}

	result, err := h.isochroneService.GetIsochrone(r.Context(), lat, lon, mode, minutes)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ISOCHRONE_ERROR", err.Error())
		return
	}

	resp := isochroneResponse{
		Coordinates: result.Coordinates,
		Mode:        string(result.Mode),
		Minutes:     result.Minutes,
		CenterLat:   result.CenterLat,
		CenterLng:   result.CenterLng,
		WKT:         result.ToWKTPolygon(),
	}

	writeJSON(w, http.StatusOK, resp)
}
