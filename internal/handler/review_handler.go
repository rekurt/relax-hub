package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

type createReviewRequest struct {
	BookingID string `json:"booking_id"`
	Rating    int    `json:"rating"`
	Text      string `json:"text"`
}

type reviewResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BathhouseID string    `json:"bathhouse_id"`
	BookingID   string    `json:"booking_id"`
	Rating      int       `json:"rating"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createReviewRequest
	if err := readJSON(r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking_id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	review, err := h.reviewService.Create(r.Context(), userID, service.CreateReviewInput{
		BookingID:   bookingID,
		BathhouseID: bathhouseID,
		Rating:      req.Rating,
		Text:        req.Text,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, reviewResponse{
		ID:          review.ID.String(),
		UserID:      review.UserID.String(),
		BathhouseID: review.BathhouseID.String(),
		BookingID:   review.BookingID.String(),
		Rating:      review.Rating,
		Text:        review.Text,
		CreatedAt:   review.CreatedAt,
	})
}

func (h *ReviewHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.reviewService.ListByBathhouse(r.Context(), bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]reviewResponse, len(result.Items))
	for i, rev := range result.Items {
		items[i] = reviewResponse{
			ID:          rev.ID.String(),
			UserID:      rev.UserID.String(),
			BathhouseID: rev.BathhouseID.String(),
			BookingID:   rev.BookingID.String(),
			Rating:      rev.Rating,
			Text:        rev.Text,
			CreatedAt:   rev.CreatedAt,
		}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
