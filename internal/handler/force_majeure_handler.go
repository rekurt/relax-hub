package handler

import (
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ForceMajeureHandler struct {
	svc service.ForceMajeureService
}

func NewForceMajeureHandler(svc service.ForceMajeureService) *ForceMajeureHandler {
	return &ForceMajeureHandler{svc: svc}
}

type activateForceMajeureRequest struct {
	Region   string `json:"region"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Reason   string `json:"reason"`
}

type forceMajeureEventResponse struct {
	ID            string    `json:"id"`
	AdminID       string    `json:"admin_id"`
	Region        string    `json:"region"`
	DateFrom      time.Time `json:"date_from"`
	DateTo        time.Time `json:"date_to"`
	Reason        string    `json:"reason"`
	AffectedCount int       `json:"affected_count"`
	TotalRefund   int64     `json:"total_refund"`
	CreatedAt     time.Time `json:"created_at"`
}

// Activate godoc
//
//	@Summary		Activate force majeure
//	@Description	Mass-cancel bookings in a region for a date range due to force majeure
//	@Tags			admin,force-majeure
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		activateForceMajeureRequest	true	"Force majeure parameters"
//	@Success		200		{object}	APIResponse{data=forceMajeureEventResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/force-majeure [post]
func (h *ForceMajeureHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req activateForceMajeureRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	dateFrom, err := time.Parse(time.RFC3339, req.DateFrom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from format, use RFC3339")
		return
	}
	dateTo, err := time.Parse(time.RFC3339, req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to format, use RFC3339")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	event, err := h.svc.Activate(r.Context(), adminID, req.Region, dateFrom, dateTo, req.Reason)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, forceMajeureEventResponse{
		ID:            event.ID.String(),
		AdminID:       event.AdminID.String(),
		Region:        event.Region,
		DateFrom:      event.DateFrom,
		DateTo:        event.DateTo,
		Reason:        event.Reason,
		AffectedCount: event.AffectedCount,
		TotalRefund:   event.TotalRefund,
		CreatedAt:     event.CreatedAt,
	})
}

// List godoc
//
//	@Summary		List force majeure events
//	@Description	Returns all force majeure events ordered by creation date desc
//	@Tags			admin,force-majeure
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]forceMajeureEventResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/force-majeure [get]
func (h *ForceMajeureHandler) List(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]forceMajeureEventResponse, len(events))
	for i, e := range events {
		result[i] = forceMajeureEventResponse{
			ID:            e.ID.String(),
			AdminID:       e.AdminID.String(),
			Region:        e.Region,
			DateFrom:      e.DateFrom,
			DateTo:        e.DateTo,
			Reason:        e.Reason,
			AffectedCount: e.AffectedCount,
			TotalRefund:   e.TotalRefund,
			CreatedAt:     e.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, result)
}
