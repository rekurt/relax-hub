package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PhotoHandler struct {
	photoService service.PhotoVerificationService
}

func NewPhotoHandler(photoService service.PhotoVerificationService) *PhotoHandler {
	return &PhotoHandler{photoService: photoService}
}

type uploadPhotoRequest struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type reorderPhotosRequest struct {
	PhotoIDs []string `json:"photo_ids"`
}

type rejectPhotoRequest struct {
	Reason string `json:"reason"`
}

type photoResponse struct {
	ID              string     `json:"id"`
	BathhouseID     string     `json:"bathhouse_id"`
	URL             string     `json:"url"`
	ThumbnailURL    string     `json:"thumbnail_url,omitempty"`
	Position        int        `json:"position"`
	Status          string     `json:"status"`
	VerifiedByID    *string    `json:"verified_by_id,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	UploadedAt      time.Time  `json:"uploaded_at"`
}

func toPhotoResponse(p *domain.BathhousePhoto) photoResponse {
	resp := photoResponse{
		ID:              p.ID.String(),
		BathhouseID:     p.BathhouseID.String(),
		URL:             p.URL,
		ThumbnailURL:    p.ThumbnailURL,
		Position:        p.Position,
		Status:          string(p.Status),
		RejectionReason: p.RejectionReason,
		UploadedAt:      p.UploadedAt,
		VerifiedAt:      p.VerifiedAt,
	}
	if p.VerifiedByID != nil {
		s := p.VerifiedByID.String()
		resp.VerifiedByID = &s
	}
	return resp
}

func toPhotoResponses(photos []domain.BathhousePhoto) []photoResponse {
	result := make([]photoResponse, len(photos))
	for i := range photos {
		result[i] = toPhotoResponse(&photos[i])
	}
	return result
}

func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req uploadPhotoRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	photo, err := h.photoService.UploadPhoto(r.Context(), userID, userRole, service.UploadPhotoInput{
		BathhouseID:  bathhouseID,
		URL:          req.URL,
		ThumbnailURL: req.ThumbnailURL,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPhotoResponse(photo))
}

func (h *PhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	photoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid photo id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.photoService.DeletePhoto(r.Context(), photoID, userID, userRole); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *PhotoHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req reorderPhotosRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	photoIDs := make([]uuid.UUID, len(req.PhotoIDs))
	for i, idStr := range req.PhotoIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid photo id in list")
			return
		}
		photoIDs[i] = id
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.photoService.ReorderPhotos(r.Context(), bathhouseID, userID, userRole, photoIDs); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *PhotoHandler) GetPending(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.photoService.GetPendingPhotos(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := toPhotoResponses(result.Items)
	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *PhotoHandler) Verify(w http.ResponseWriter, r *http.Request) {
	photoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid photo id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	photo, err := h.photoService.VerifyPhoto(r.Context(), photoID, adminID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPhotoResponse(photo))
}

func (h *PhotoHandler) Reject(w http.ResponseWriter, r *http.Request) {
	photoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid photo id")
		return
	}

	var req rejectPhotoRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	adminID := middleware.GetUserID(r.Context())

	photo, err := h.photoService.RejectPhoto(r.Context(), photoID, adminID, req.Reason)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPhotoResponse(photo))
}

func (h *PhotoHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	photos, err := h.photoService.ListByBathhouse(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPhotoResponses(photos))
}
