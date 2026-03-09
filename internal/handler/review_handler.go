package handler

import (
	"bytes"
	"io"
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
	mediaService  service.MediaService
}

func NewReviewHandler(reviewService service.ReviewService, mediaService service.MediaService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		mediaService:  mediaService,
	}
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

type mediaResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	OriginalName string    `json:"original_name"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type reviewResponse struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	BathhouseID     string           `json:"bathhouse_id"`
	BookingID       string           `json:"booking_id"`
	Rating          int              `json:"rating"`
	Text            string           `json:"text"`
	Status          string           `json:"status"`
	OwnerResponse   string           `json:"owner_response,omitempty"`
	OwnerResponseAt *time.Time       `json:"owner_response_at,omitempty"`
	Images          []string         `json:"images,omitempty"`
	Media           []mediaResponse  `json:"media,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func toMediaResponse(m *domain.Media) mediaResponse {
	return mediaResponse{
		ID:           m.ID.String(),
		Type:         string(m.Type),
		URL:          m.URL,
		ThumbnailURL: m.ThumbnailURL,
		OriginalName: m.OriginalName,
		Size:         m.Size,
		MimeType:     m.MimeType,
		Width:        m.Width,
		Height:       m.Height,
		Status:       string(m.Status),
		CreatedAt:    m.CreatedAt,
	}
}

func toMediaResponses(items []domain.Media) []mediaResponse {
	result := make([]mediaResponse, len(items))
	for i := range items {
		result[i] = toMediaResponse(&items[i])
	}
	return result
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
		resp := toReviewResponse(&rev)
		media, err := h.mediaService.ListByReview(r.Context(), rev.ID)
		if err == nil && len(media) > 0 {
			resp.Media = toMediaResponses(media)
		}
		items[i] = resp
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

const maxMediaUploadSize = 50 << 20 // 50MB (max video size)

func (h *ReviewHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	// Verify the review exists and belongs to the user
	userID := middleware.GetUserID(r.Context())
	review, err := h.reviewService.GetByID(r.Context(), reviewID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if review.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMediaUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "file is required")
		return
	}
	defer file.Close()

	// Detect actual content type from file bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read file")
		return
	}
	if n == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "empty file")
		return
	}
	detectedType := http.DetectContentType(buf[:n])

	// Reconstruct the reader with the already-read bytes prepended
	data := io.MultiReader(bytes.NewReader(buf[:n]), file)

	media, err := h.mediaService.Upload(r.Context(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         data,
		OriginalName: header.Filename,
		Size:         header.Size,
		MimeType:     detectedType,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toMediaResponse(media))
}

func (h *ReviewHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	mediaID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid media id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.mediaService.Delete(r.Context(), mediaID, userID, userRole); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "media deleted"})
}
