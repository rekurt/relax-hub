package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
)

type cancelBookingRequest struct {
	RefundTo string `json:"refund_to"` // "wallet" or "card", default "card"
}

// Cancel godoc
//
//	@Summary		Cancel booking
//	@Description	Cancels a booking. Clients can cancel their own bookings, owners/representatives can cancel bookings for their bathhouses. Optional refund_to parameter to choose refund destination.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		cancelBookingRequest	false	"Cancel options"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/cancel [patch]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req cancelBookingRequest
	// Body is optional for cancel
	_ = readJSON(w, r, &req)

	if req.RefundTo != "" && req.RefundTo != "wallet" && req.RefundTo != "card" {
		writeError(w, http.StatusBadRequest, "invalid_input", "refund_to must be 'wallet' or 'card'")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Cancel(r.Context(), userID, role, bookingID, req.RefundTo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cancelled"})
}

// Confirm godoc
//
//	@Summary		Confirm booking
//	@Description	Confirms a pending booking. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/confirm [patch]
func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Confirm(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "confirmed"})
}

type rejectBookingRequest struct {
	Reason string `json:"reason"`
}

// Reject godoc
//
//	@Summary		Reject booking
//	@Description	Rejects a pending or pending_owner booking. Only available to bathhouse owners and representatives. Optional rejection reason.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		rejectBookingRequest	false	"Rejection reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/reject [patch]
func (h *BookingHandler) Reject(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req rejectBookingRequest
	// Body is optional for reject
	_ = readJSON(w, r, &req)

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Reject(r.Context(), userID, role, bookingID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rejected"})
}

// Approve godoc
//
//	@Summary		Approve booking request
//	@Description	Approves a pending_owner booking request. Captures wallet hold. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/approve [patch]
func (h *BookingHandler) Approve(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Approve(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "approved"})
}

// Complete godoc
//
//	@Summary		Complete booking
//	@Description	Marks a confirmed booking as completed. Awards loyalty points. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=bookingResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/complete [patch]
func (h *BookingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.bookingService.Complete(r.Context(), userID, role, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBookingResultResponse(result))
}

// ListByBathhouse godoc
//
//	@Summary		List bathhouse bookings
//	@Description	Returns a paginated list of bookings for a specific bathhouse. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Bathhouse ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]bookingResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/bookings [get]
func (h *BookingHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.bookingService.ListByBathhouse(r.Context(), userID, role, bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bookingResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBookingResponse(&result.Items[i])
		h.enrichWithPaymentStatus(r.Context(), &items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// CheckIn godoc
//
//	@Summary		Check-in guest
//	@Description	Marks a guest as arrived for a confirmed booking. Available 15 min before to 30 min after booking start time. Only owners/representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/check-in [patch]
func (h *BookingHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.CheckIn(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "checked_in"})
}

// CheckOut godoc
//
//	@Summary		Check-out guest
//	@Description	Marks a guest as departed, completing the booking. Requires prior check-in. Awards loyalty points. Only owners/representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/check-out [patch]
func (h *BookingHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.CheckOut(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "checked_out"})
}

