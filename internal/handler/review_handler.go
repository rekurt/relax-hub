package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ReviewHandler struct {
	reviewService service.ReviewService
	mediaService  service.MediaService
	log           *logger.Logger
}

func NewReviewHandler(reviewService service.ReviewService, mediaService service.MediaService, log *logger.Logger) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		mediaService:  mediaService,
		log:           log,
	}
}

type createReviewRequest struct {
	BookingID     string   `json:"booking_id"`
	Rating        int      `json:"rating"`
	Cleanliness   *float64 `json:"cleanliness,omitempty"`
	Accuracy      *float64 `json:"accuracy,omitempty"`
	Communication *float64 `json:"communication,omitempty"`
	ValueForMoney *float64 `json:"value_for_money,omitempty"`
	Text          string   `json:"text"`
}

type updateReviewRequest struct {
	Rating        *int     `json:"rating,omitempty"`
	Cleanliness   *float64 `json:"cleanliness,omitempty"`
	Accuracy      *float64 `json:"accuracy,omitempty"`
	Communication *float64 `json:"communication,omitempty"`
	ValueForMoney *float64 `json:"value_for_money,omitempty"`
	Text          *string  `json:"text,omitempty"`
	Images        []string `json:"images,omitempty"`
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
	Cleanliness     *float64         `json:"cleanliness,omitempty"`
	Accuracy        *float64         `json:"accuracy,omitempty"`
	Communication   *float64         `json:"communication,omitempty"`
	ValueForMoney   *float64         `json:"value_for_money,omitempty"`
	Text            string           `json:"text"`
	Status          string           `json:"status"`
	ModerationScore *float64         `json:"moderation_score,omitempty"`
	ModerationFlags []string         `json:"moderation_flags,omitempty"`
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
		Cleanliness:     rev.Cleanliness,
		Accuracy:        rev.Accuracy,
		Communication:   rev.Communication,
		ValueForMoney:   rev.ValueForMoney,
		Text:            rev.Text,
		Status:          string(rev.Status),
		ModerationScore: rev.ModerationScore,
		ModerationFlags: rev.ModerationFlags,
		OwnerResponse:   rev.OwnerResponse,
		OwnerResponseAt: rev.OwnerResponseAt,
		Images:          rev.Images,
		CreatedAt:       rev.CreatedAt,
		UpdatedAt:       rev.UpdatedAt,
	}
}

// Create godoc
// @Summary      Create review
// @Description  Creates a review for a bathhouse. Requires a completed booking. Only clients can write reviews.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string              true  "Bathhouse ID (UUID)"
// @Param        body  body      createReviewRequest  true  "Review data"
// @Success      201   {object}  APIResponse{data=reviewResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/reviews [post]
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
		BookingID:     bookingID,
		BathhouseID:   bathhouseID,
		Rating:        req.Rating,
		Cleanliness:   req.Cleanliness,
		Accuracy:      req.Accuracy,
		Communication: req.Communication,
		ValueForMoney: req.ValueForMoney,
		Text:          req.Text,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toReviewResponse(review))
}

// Update godoc
// @Summary      Update review
// @Description  Updates a review. Only the review author can update it.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string               true  "Review ID (UUID)"
// @Param        body  body      updateReviewRequest   true  "Fields to update"
// @Success      200   {object}  APIResponse{data=reviewResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /reviews/{id} [put]
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
		Rating:        req.Rating,
		Cleanliness:   req.Cleanliness,
		Accuracy:      req.Accuracy,
		Communication: req.Communication,
		ValueForMoney: req.ValueForMoney,
		Text:          req.Text,
		Images:        req.Images,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReviewResponse(review))
}

// Delete godoc
// @Summary      Delete review
// @Description  Deletes a review. The author or an admin can delete it.
// @Tags         reviews
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Review ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /reviews/{id} [delete]
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

// AddOwnerResponse godoc
// @Summary      Respond to review
// @Description  Adds an owner/representative response to a review. Only available to bathhouse owners and representatives.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "Review ID (UUID)"
// @Param        body  body      ownerResponseRequest  true  "Response text"
// @Success      200   {object}  APIResponse{data=reviewResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /reviews/{id}/response [post]
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

// ListByBathhouse godoc
// @Summary      List bathhouse reviews
// @Description  Returns a paginated list of reviews for a bathhouse with attached media
// @Tags         reviews
// @Produce      json
// @Param        id         path      string  true   "Bathhouse ID (UUID)"
// @Param        page       query     int     false  "Page number"  default(1)
// @Param        page_size  query     int     false  "Page size"    default(20)
// @Success      200        {object}  APIResponse{data=[]reviewResponse,meta=Meta}
// @Failure      400        {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/reviews [get]
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
	reviewIDs := make([]uuid.UUID, len(result.Items))
	for i, rev := range result.Items {
		reviewIDs[i] = rev.ID
	}

	mediaByReview, err := h.mediaService.ListByReviewIDs(r.Context(), reviewIDs)
	if err != nil {
		h.log.Warn("failed to load media for reviews", "error", err)
		mediaByReview = make(map[uuid.UUID][]domain.Media)
	}

	for i, rev := range result.Items {
		resp := toReviewResponse(&rev)
		if media, ok := mediaByReview[rev.ID]; ok && len(media) > 0 {
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

// UploadMedia godoc
// @Summary      Upload review media
// @Description  Uploads a photo or video attachment to a review. Max 10 photos and 1 video per review. Max file size 50MB.
// @Tags         review-media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Review ID (UUID)"
// @Param        file  formData  file    true  "Media file (image or video)"
// @Success      201   {object}  APIResponse{data=mediaResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /reviews/{id}/media [post]
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
	if err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read file")
		return
	}
	if n == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "empty file")
		return
	}
	detectedType := http.DetectContentType(buf[:n])

	// http.DetectContentType cannot reliably detect video MIME types.
	// Fall back to file extension for video detection.
	if detectedType == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".mp4":
			detectedType = "video/mp4"
		case ".webm":
			detectedType = "video/webm"
		}
	}

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

// DeleteMedia godoc
// @Summary      Delete review media
// @Description  Deletes a media attachment from a review. Only the review author or an admin can delete.
// @Tags         review-media
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Media ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /media/{id} [delete]
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
