package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type extendBookingRequest struct {
	ExtraHours    int    `json:"extra_hours"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

type extendBookingResponse struct {
	Booking        bookingResponse `json:"booking"`
	ExtensionPrice int64           `json:"extension_price"`
}

// Extend godoc
//
//	@Summary		Extend booking session
//	@Description	Extends an active booking session by 1-2 hours. Only the booking owner (client) can extend. Booking must be confirmed or checked-in. Extension slots must be available.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		extendBookingRequest	true	"Extension details"
//	@Success		200		{object}	APIResponse{data=extendBookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/extend [post]
func (h *BookingHandler) Extend(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req extendBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.ExtraHours < 1 || req.ExtraHours > 2 {
		writeError(w, http.StatusBadRequest, "invalid_input", "extra_hours must be 1 or 2")
		return
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.bookingService.Extend(r.Context(), userID, bookingID, req.ExtraHours)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := extendBookingResponse{
		Booking:        toBookingResponse(result.Booking),
		ExtensionPrice: result.ExtensionPrice,
	}
	writeJSON(w, http.StatusOK, resp)
}

type rebookAddOnResponse struct {
	AddOnID  string `json:"addon_id"`
	Quantity int    `json:"quantity"`
}

type rebookDataResponse struct {
	BathhouseID   string                `json:"bathhouse_id"`
	DurationHours int                   `json:"duration_hours"`
	TimeFrom      string                `json:"time_from"`
	TimeTo        string                `json:"time_to"`
	GuestCount    int                   `json:"guest_count"`
	AddOns        []rebookAddOnResponse `json:"addons,omitempty"`
}

// GetRebookData godoc
//
//	@Summary		Get re-booking data
//	@Description	Returns pre-filled booking parameters from a past booking (completed or cancelled) for quick re-booking. Prices are not included as they are recalculated at booking time.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=rebookDataResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/rebook-data [get]
func (h *BookingHandler) GetRebookData(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	data, err := h.bookingService.GetRebookData(r.Context(), userID, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := rebookDataResponse{
		BathhouseID:   data.BathhouseID.String(),
		DurationHours: data.DurationHours,
		TimeFrom:      data.TimeFrom,
		TimeTo:        data.TimeTo,
		GuestCount:    data.GuestCount,
	}
	for _, a := range data.AddOns {
		resp.AddOns = append(resp.AddOns, rebookAddOnResponse{
			AddOnID:  a.AddOnID.String(),
			Quantity: a.Quantity,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

type modifyBookingRequest struct {
	StartTime  string                  `json:"start_time"`
	EndTime    string                  `json:"end_time"`
	GuestCount int                     `json:"guest_count"`
	AddOns     []addOnSelectionRequest `json:"addons,omitempty"`
}

type modifyBookingResponse struct {
	Booking   bookingResponse `json:"booking"`
	OldPrice  int64           `json:"old_price"`
	NewPrice  int64           `json:"new_price"`
	PriceDiff int64           `json:"price_diff"`
}

// Modify godoc
//
//	@Summary		Modify booking
//	@Description	Modifies a pending or confirmed booking. Allows changing date/time, duration, guest count, and add-ons. Maximum 3 modifications per booking. If the new price differs, a refund or additional charge is handled automatically.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		modifyBookingRequest	true	"Modification data"
//	@Success		200		{object}	APIResponse{data=modifyBookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/modify [put]
func (h *BookingHandler) Modify(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req modifyBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid start_time format, use RFC3339")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid end_time format, use RFC3339")
		return
	}

	if req.GuestCount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "guest_count must be positive")
		return
	}

	var addOnSelections []service.AddOnSelection
	for _, a := range req.AddOns {
		addonID, err := uuid.Parse(a.AddOnID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid addon_id")
			return
		}
		addOnSelections = append(addOnSelections, service.AddOnSelection{
			AddOnID:  addonID,
			Quantity: a.Quantity,
		})
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.bookingService.Modify(r.Context(), userID, bookingID, service.ModifyBookingInput{
		StartTime:  startTime,
		EndTime:    endTime,
		GuestCount: req.GuestCount,
		AddOns:     addOnSelections,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := modifyBookingResponse{
		Booking:   toBookingResponse(result.Booking),
		OldPrice:  result.OldPrice,
		NewPrice:  result.NewPrice,
		PriceDiff: result.PriceDiff,
	}
	writeJSON(w, http.StatusOK, resp)
}

type disputeNoShowRequest struct {
	GPSLat  float64 `json:"gps_lat"`
	GPSLon  float64 `json:"gps_lon"`
	Comment string  `json:"comment"`
}

// DisputeNoShow godoc
//
//	@Summary		Dispute no-show
//	@Description	Client disputes a no-show status within 2 hours. Creates a support ticket with GPS coordinates.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		disputeNoShowRequest	true	"Dispute details"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/dispute-noshow [post]
func (h *BookingHandler) DisputeNoShow(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req disputeNoShowRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.bookingService.DisputeNoShow(r.Context(), userID, bookingID, req.GPSLat, req.GPSLon, req.Comment); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "dispute_created"})
}
