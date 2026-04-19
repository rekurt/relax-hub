package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/geo"
	"github.com/rekurt/relax-hub/internal/service"
)

// TransportHandler handles requests for nearby transport information.
type TransportHandler struct {
	transportService *geo.TransportService
	bathhouseService service.BathhouseService
}

// NewTransportHandler creates a new TransportHandler.
func NewTransportHandler(transportService *geo.TransportService, bathhouseService service.BathhouseService) *TransportHandler {
	return &TransportHandler{
		transportService: transportService,
		bathhouseService: bathhouseService,
	}
}

type transportInfoResponse struct {
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Distance int     `json:"distance_meters"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type transportResponse struct {
	Items []transportInfoResponse `json:"items"`
}

// GetTransport godoc
//
//	@Summary		Get nearby transport for a bathhouse
//	@Description	Returns nearest metro stations, bus stops, and parking near the bathhouse
//	@Tags			bathhouses
//	@Produce		json
//	@Param			id	path		string	true	"Bathhouse ID"
//	@Success		200	{object}	APIResponse{data=transportResponse}
//	@Failure		400	{object}	APIResponse
//	@Failure		404	{object}	APIResponse
//	@Router			/bathhouses/{id}/transport [get]
func (h *TransportHandler) GetTransport(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.transportService.GetNearbyTransport(r.Context(), bh.ID.String(), bh.Latitude, bh.Longitude)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "transport_error", "failed to fetch transport data")
		return
	}

	items := make([]transportInfoResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, transportInfoResponse{
			Type:     string(item.Type),
			Name:     item.Name,
			Distance: item.Distance,
			Lat:      item.Lat,
			Lng:      item.Lng,
		})
	}

	writeJSON(w, http.StatusOK, transportResponse{Items: items})
}
