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

type updateReviewRequest struct {
	Rating *int     `json:"rating,omitempty"`
	Text   *string  `json:"text,omitempty"`
	Images []string `json:"images,omitempty"`
}

type ownerResponseRequest struct {
	Response string `json:"response"`
}

type reviewResponse struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	BathhouseID     string     `json:"bathhouse_id"`
	BookingID       string     `json:"booking_id"`
	Rating          int        `json:"rating"`
	Text            string     `json:"text"`
	Status          string     `json:"status"`
	OwnerResponse   string     `json:"owner_response,omitempty"`
	OwnerResponseAt *time.Time `json:"owner_response_at,omitempty"`
	Images          []string   `json:"images,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func toReviewResponse(rev *domain.Review) reviewResponse {
	return reviewResponse{
		ID:              rev.ID.String(),
		UserID:          rev.UserID.String(),
		BathhouseID:     rev.BathhouseID.String(),
		BookingID:       rev.BookingID.String(),
		Rating:          rev.Rating,
		Text:            rev.Text,
		Status:          string(rev.Status),
		OwnerResponse:   rev.OwnerResponse,
		OwnerResponseAt: rev.OwnerResponseAt,
		Images:          rev.Images,
		CreatedAt:       rev.CreatedAt,
		UpdatedAt:       rev.UpdatedAt,
	}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createReviewRequest
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

	writeJSON(w, http.StatusCreated, toReviewResponse(review))
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	var req updateReviewRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	review, err := h.reviewService.Update(r.Context(), userID, reviewID, service.UpdateReviewInput{
		Rating: req.Rating,
		Text:   req.Text,
		Images: req.Images,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReviewResponse(review))
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.reviewService.Delete(r.Context(), userID, userRole, reviewID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "review deleted"})
}

func (h *ReviewHandler) AddOwnerResponse(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	var req ownerResponseRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	review, err := h.reviewService.AddOwnerResponse(r.Context(), userID, userRole, reviewID, req.Response)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReviewResponse(review))
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
		items[i] = toReviewResponse(&rev)
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
