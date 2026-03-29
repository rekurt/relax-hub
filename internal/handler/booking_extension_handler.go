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

type BookingExtensionHandler struct {
	extSvc service.BookingExtensionService
}

func NewBookingExtensionHandler(extSvc service.BookingExtensionService) *BookingExtensionHandler {
	return &BookingExtensionHandler{extSvc: extSvc}
}

type extensionRequestResponse struct {
	ID              string     `json:"id"`
	BookingID       string     `json:"booking_id"`
	UserID          string     `json:"user_id"`
	BathhouseID     string     `json:"bathhouse_id"`
	Status          string     `json:"status"`
	ExtraHours      int        `json:"extra_hours"`
	ExtensionPrice  int64      `json:"extension_price"`
	NewEndTime      time.Time  `json:"new_end_time"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

func toExtensionRequestResponse(r *domain.BookingExtensionRequest) extensionRequestResponse {
	return extensionRequestResponse{
		ID:              r.ID.String(),
		BookingID:       r.BookingID.String(),
		UserID:          r.UserID.String(),
		BathhouseID:     r.BathhouseID.String(),
		Status:          string(r.Status),
		ExtraHours:      r.ExtraHours,
		ExtensionPrice:  r.ExtensionPrice,
		NewEndTime:      r.NewEndTime,
		RejectionReason: r.RejectionReason,
		CreatedAt:       r.CreatedAt,
		ExpiresAt:       r.ExpiresAt,
		ResolvedAt:      r.ResolvedAt,
	}
}

type requestExtensionRequest struct {
	ExtraHours int `json:"extra_hours"`
}

// RequestExtension godoc
//
//	@Summary		Request booking extension
//	@Description	Creates a pending extension request that the bathhouse owner must approve. Funds are held in the client's wallet.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Booking ID (UUID)"
//	@Param			body	body		requestExtensionRequest		true	"Extension details"
//	@Success		201		{object}	APIResponse{data=extensionRequestResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/extension-request [post]
func (h *BookingExtensionHandler) RequestExtension(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req requestExtensionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.extSvc.RequestExtension(r.Context(), userID, bookingID, req.ExtraHours)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := toExtensionRequestResponse(result)
	writeJSON(w, http.StatusCreated, resp)
}

// ListExtensionRequests godoc
//
//	@Summary		List extension requests for a booking
//	@Description	Returns all extension requests for the specified booking.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]extensionRequestResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/extension-requests [get]
func (h *BookingExtensionHandler) ListExtensionRequests(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	// Authorization: verify the caller is related to this booking
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	if err := h.extSvc.AuthorizeListAccess(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	requests, err := h.extSvc.ListByBooking(r.Context(), bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]extensionRequestResponse, len(requests))
	for i, req := range requests {
		resp[i] = toExtensionRequestResponse(&req)
	}
	writeJSON(w, http.StatusOK, resp)
}

type rejectExtensionRequest struct {
	Reason string `json:"reason"`
}

// ApproveExtension godoc
//
//	@Summary		Approve extension request
//	@Description	Approves a pending extension request and applies the extension to the booking.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Extension Request ID (UUID)"
//	@Success		200	{object}	APIResponse{data=extendBookingResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/extension-requests/{id}/approve [patch]
func (h *BookingExtensionHandler) ApproveExtension(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.extSvc.ApproveExtension(r.Context(), userID, role, requestID)
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

// RejectExtension godoc
//
//	@Summary		Reject extension request
//	@Description	Rejects a pending extension request and releases the wallet hold.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Extension Request ID (UUID)"
//	@Param			body	body		rejectExtensionRequest		true	"Rejection details"
//	@Success		200		{object}	APIResponse
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/extension-requests/{id}/reject [patch]
func (h *BookingExtensionHandler) RejectExtension(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request id")
		return
	}

	var req rejectExtensionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.extSvc.RejectExtension(r.Context(), userID, role, requestID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "extension request rejected"})
}
