package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type BookingModificationHandler struct {
	modSvc service.BookingModificationService
}

func NewBookingModificationHandler(modSvc service.BookingModificationService) *BookingModificationHandler {
	return &BookingModificationHandler{modSvc: modSvc}
}

type modificationRequestResponse struct {
	ID                 string     `json:"id"`
	BookingID          string     `json:"booking_id"`
	UserID             string     `json:"user_id"`
	BathhouseID        string     `json:"bathhouse_id"`
	Status             string     `json:"status"`
	OldStartTime       time.Time  `json:"old_start_time"`
	OldEndTime         time.Time  `json:"old_end_time"`
	OldGuestCount      int        `json:"old_guest_count"`
	OldTotalPrice      int64      `json:"old_total_price"`
	ProposedStartTime  time.Time  `json:"proposed_start_time"`
	ProposedEndTime    time.Time  `json:"proposed_end_time"`
	ProposedGuestCount int        `json:"proposed_guest_count"`
	ProposedTotalPrice int64      `json:"proposed_total_price"`
	RejectionReason    string     `json:"rejection_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ExpiresAt          time.Time  `json:"expires_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
}

func toModificationRequestResponse(r *domain.BookingModificationRequest) modificationRequestResponse {
	return modificationRequestResponse{
		ID:                 r.ID.String(),
		BookingID:          r.BookingID.String(),
		UserID:             r.UserID.String(),
		BathhouseID:        r.BathhouseID.String(),
		Status:             string(r.Status),
		OldStartTime:       r.OldStartTime,
		OldEndTime:         r.OldEndTime,
		OldGuestCount:      r.OldGuestCount,
		OldTotalPrice:      r.OldTotalPrice,
		ProposedStartTime:  r.ProposedStartTime,
		ProposedEndTime:    r.ProposedEndTime,
		ProposedGuestCount: r.ProposedGuestCount,
		ProposedTotalPrice: r.ProposedTotalPrice,
		RejectionReason:    r.RejectionReason,
		CreatedAt:          r.CreatedAt,
		ExpiresAt:          r.ExpiresAt,
		ResolvedAt:         r.ResolvedAt,
	}
}

type requestModificationRequest struct {
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	GuestCount int    `json:"guest_count"`
}

// RequestModification godoc
//
//	@Summary		Request booking modification
//	@Description	Creates a modification request that the owner must approve. Maximum 3 modifications per booking.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Booking ID (UUID)"
//	@Param			body	body		requestModificationRequest		true	"Proposed changes"
//	@Success		201		{object}	APIResponse{data=modificationRequestResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/modification-request [post]
func (h *BookingModificationHandler) RequestModification(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req requestModificationRequest
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

	userID := middleware.GetUserID(r.Context())

	result, err := h.modSvc.RequestModification(r.Context(), userID, bookingID, service.ModifyBookingInput{
		StartTime:  startTime,
		EndTime:    endTime,
		GuestCount: req.GuestCount,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toModificationRequestResponse(result))
}

type rejectModificationRequest struct {
	Reason string `json:"reason"`
}

// ApproveModification godoc
//
//	@Summary		Approve modification request
//	@Description	Owner approves a pending booking modification request. Changes are applied immediately.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Modification Request ID (UUID)"
//	@Success		200	{object}	APIResponse{data=modifyBookingResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/modification-requests/{id}/approve [patch]
func (h *BookingModificationHandler) ApproveModification(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.modSvc.ApproveModification(r.Context(), userID, role, requestID)
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

// RejectModification godoc
//
//	@Summary		Reject modification request
//	@Description	Owner rejects a pending booking modification request.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Modification Request ID (UUID)"
//	@Param			body	body		rejectModificationRequest	true	"Rejection reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/modification-requests/{id}/reject [patch]
func (h *BookingModificationHandler) RejectModification(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request id")
		return
	}

	var req rejectModificationRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.modSvc.RejectModification(r.Context(), userID, role, requestID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "modification request rejected"})
}

// ListModificationRequests godoc
//
//	@Summary		List modification requests
//	@Description	Lists all modification requests for a booking.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]modificationRequestResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/modification-requests [get]
func (h *BookingModificationHandler) ListModificationRequests(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	// Authorization: verify the caller is related to this booking
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if err := h.modSvc.AuthorizeListAccess(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	requests, err := h.modSvc.ListByBooking(r.Context(), bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]modificationRequestResponse, 0, len(requests))
	for _, req := range requests {
		resp = append(resp, toModificationRequestResponse(&req))
	}

	writeJSON(w, http.StatusOK, resp)
}
