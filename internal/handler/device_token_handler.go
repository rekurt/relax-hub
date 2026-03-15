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

type DeviceTokenHandler struct {
	svc service.DeviceTokenService
}

func NewDeviceTokenHandler(svc service.DeviceTokenService) *DeviceTokenHandler {
	return &DeviceTokenHandler{svc: svc}
}

type registerDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type deviceTokenResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *DeviceTokenHandler) Register(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req registerDeviceTokenRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "Token is required")
		return
	}
	if req.Platform == "" {
		req.Platform = "web"
	}

	dt := &domain.DeviceToken{
		ID:       uuid.New(),
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
	}

	if err := dt.Validate(); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.svc.Register(r.Context(), dt); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, deviceTokenResponse{
		ID:        dt.ID.String(),
		UserID:    dt.UserID.String(),
		Token:     dt.Token,
		Platform:  dt.Platform,
		CreatedAt: dt.CreatedAt,
	})
}

func (h *DeviceTokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Invalid device token ID")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
