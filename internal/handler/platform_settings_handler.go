package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PlatformSettingsHandler struct {
	svc service.PlatformSettingsService
}

func NewPlatformSettingsHandler(svc service.PlatformSettingsService) *PlatformSettingsHandler {
	return &PlatformSettingsHandler{svc: svc}
}

type platformSettingResponse struct {
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdatedBy   *string    `json:"updated_by,omitempty"`
}

type updateSettingRequest struct {
	Value string `json:"value"`
}

// List godoc
//
//	@Summary		List all platform settings
//	@Description	Returns all platform settings (admin only)
//	@Tags			admin,platform-settings
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]platformSettingResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/settings [get]
func (h *PlatformSettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.GetAll(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]platformSettingResponse, len(settings))
	for i, s := range settings {
		result[i] = platformSettingResponse{
			Key:         s.Key,
			Value:       s.Value,
			Description: s.Description,
			Type:        string(s.Type),
			UpdatedAt:   s.UpdatedAt,
			UpdatedBy:   s.UpdatedBy,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// Update godoc
//
//	@Summary		Update a platform setting
//	@Description	Updates the value of a platform setting by key (admin only)
//	@Tags			admin,platform-settings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			key		path		string					true	"Setting key"
//	@Param			body	body		updateSettingRequest	true	"New value"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/settings/{key} [put]
func (h *PlatformSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "setting key is required")
		return
	}

	var req updateSettingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.svc.Set(r.Context(), key, req.Value, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "setting updated"})
}
