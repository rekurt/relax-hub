package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

type adminCancelBookingRequest struct {
	Reason string `json:"reason"`
}

// AdminCancel godoc
//
//	@Summary		Admin cancel booking
//	@Description	Admin cancels any booking regardless of ownership. Issues full wallet refund. Logged to audit.
//	@Tags			admin-bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Booking ID (UUID)"
//	@Param			body	body		adminCancelBookingRequest	true	"Cancel reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings/{id}/cancel [post]
func (h *BookingHandler) AdminCancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req adminCancelBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.bookingService.AdminCancel(r.Context(), adminID, bookingID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cancelled"})
}

type adminChangeStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// AdminChangeStatus godoc
//
//	@Summary		Admin change booking status
//	@Description	Admin changes the status of any booking. Use for edge cases that normal flows don't cover.
//	@Tags			admin-bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Booking ID (UUID)"
//	@Param			body	body		adminChangeStatusRequest	true	"New status and reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings/{id}/change-status [post]
func (h *BookingHandler) AdminChangeStatus(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req adminChangeStatusRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	status := domain.BookingStatus(req.Status)
	if !status.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid status")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.bookingService.AdminChangeStatus(r.Context(), adminID, bookingID, status, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status_changed"})
}

// AdminListBookings godoc
//
//	@Summary		Admin list bookings
//	@Description	Returns a paginated list of all bookings with optional filters for admin.
//	@Tags			admin-bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id			query		string	false	"Filter by user ID"
//	@Param			bathhouse_id	query		string	false	"Filter by bathhouse ID"
//	@Param			status			query		string	false	"Filter by status"
//	@Param			from_date		query		string	false	"Filter from date (RFC3339)"
//	@Param			to_date			query		string	false	"Filter to date (RFC3339)"
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			page_size		query		int		false	"Page size"		default(20)
//	@Success		200				{object}	APIResponse{data=[]bookingResponse,meta=Meta}
//	@Failure		400				{object}	APIResponse{error=APIError}
//	@Failure		401				{object}	APIResponse{error=APIError}
//	@Failure		403				{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings [get]
func (h *BookingHandler) AdminListBookings(w http.ResponseWriter, r *http.Request) {
	filter := domain.AdminBookingFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if uid := r.URL.Query().Get("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid user_id")
			return
		}
		filter.UserID = &id
	}

	if bid := r.URL.Query().Get("bathhouse_id"); bid != "" {
		id, err := uuid.Parse(bid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
			return
		}
		filter.BathhouseID = &id
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.BookingStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status")
			return
		}
		filter.Status = &status
	}

	if fd := r.URL.Query().Get("from_date"); fd != "" {
		t, err := time.Parse(time.RFC3339, fd)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid from_date format, use RFC3339")
			return
		}
		filter.FromDate = &t
	}

	if td := r.URL.Query().Get("to_date"); td != "" {
		t, err := time.Parse(time.RFC3339, td)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid to_date format, use RFC3339")
			return
		}
		filter.ToDate = &t
	}

	result, err := h.bookingService.AdminListBookings(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bookingResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBookingResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
