package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type ClientReviewHandler struct {
	clientReviewService service.ClientReviewService
	log                 *logger.Logger
}

func NewClientReviewHandler(clientReviewService service.ClientReviewService, log *logger.Logger) *ClientReviewHandler {
	return &ClientReviewHandler{
		clientReviewService: clientReviewService,
		log:                 log,
	}
}

type createClientReviewRequest struct {
	BookingID      string   `json:"booking_id"`
	Punctuality    *float64 `json:"punctuality,omitempty"`
	Cleanliness    *float64 `json:"cleanliness,omitempty"`
	RuleCompliance *float64 `json:"rule_compliance,omitempty"`
	Rating         int      `json:"rating"`
	Text           string   `json:"text"`
}

type clientReviewResponse struct {
	ID             string    `json:"id"`
	OwnerID        string    `json:"owner_id"`
	ClientID       string    `json:"client_id"`
	BookingID      string    `json:"booking_id"`
	BathhouseID    string    `json:"bathhouse_id"`
	Punctuality    *float64  `json:"punctuality,omitempty"`
	Cleanliness    *float64  `json:"cleanliness,omitempty"`
	RuleCompliance *float64  `json:"rule_compliance,omitempty"`
	Rating         int       `json:"rating"`
	Text           string    `json:"text"`
	IsRevealed     bool      `json:"is_revealed"`
	RevealAt       time.Time `json:"reveal_at"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toClientReviewResponse(r *domain.ClientReview) clientReviewResponse {
	return clientReviewResponse{
		ID:             r.ID.String(),
		OwnerID:        r.OwnerID.String(),
		ClientID:       r.ClientID.String(),
		BookingID:      r.BookingID.String(),
		BathhouseID:    r.BathhouseID.String(),
		Punctuality:    r.Punctuality,
		Cleanliness:    r.Cleanliness,
		RuleCompliance: r.RuleCompliance,
		Rating:         r.Rating,
		Text:           r.Text,
		IsRevealed:     r.IsRevealed,
		RevealAt:       r.RevealAt,
		Status:         string(r.Status),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// Create godoc
//
//	@Summary		Create client review
//	@Description	Creates a review of a client by the bathhouse owner/representative. Only available after booking completion.
//	@Tags			client-reviews
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createClientReviewRequest	true	"Client review data"
//	@Success		201		{object}	APIResponse{data=clientReviewResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/client-reviews [post]
func (h *ClientReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createClientReviewRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking_id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	review, err := h.clientReviewService.Create(r.Context(), userID, userRole, service.CreateClientReviewInput{
		BookingID:      bookingID,
		Punctuality:    req.Punctuality,
		Cleanliness:    req.Cleanliness,
		RuleCompliance: req.RuleCompliance,
		Rating:         req.Rating,
		Text:           req.Text,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toClientReviewResponse(review))
}

// GetByBooking godoc
//
//	@Summary		Get client review by booking
//	@Description	Returns the owner's review of a client for a specific booking.
//	@Tags			client-reviews
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=clientReviewResponse}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/client-review [get]
func (h *ClientReviewHandler) GetByBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	review, err := h.clientReviewService.GetByBookingID(r.Context(), bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Only return the full review if it's revealed, or if caller is the owner
	userID := middleware.GetUserID(r.Context())
	if !review.IsRevealed && review.OwnerID != userID {
		writeError(w, http.StatusForbidden, "review_blind_period", "review is in blind period")
		return
	}

	writeJSON(w, http.StatusOK, toClientReviewResponse(review))
}

// ListMyClientReviews godoc
//
//	@Summary		List my client reviews
//	@Description	Returns revealed reviews that owners have written about the current user as a client.
//	@Tags			client-reviews
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]clientReviewResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/client-reviews [get]
func (h *ClientReviewHandler) ListMyClientReviews(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.clientReviewService.ListByClient(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]clientReviewResponse, len(result.Items))
	for i, rev := range result.Items {
		items[i] = toClientReviewResponse(&rev)
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
