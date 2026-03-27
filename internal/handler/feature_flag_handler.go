package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type FeatureFlagHandler struct {
	svc service.FeatureFlagService
}

func NewFeatureFlagHandler(svc service.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{svc: svc}
}

type featureFlagResponse struct {
	Key         string    `json:"key"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	Region      *string   `json:"region,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
}

type updateFeatureFlagRequest struct {
	Enabled bool    `json:"enabled"`
	Region  *string `json:"region,omitempty"`
}

// List godoc
//
//	@Summary		List all feature flags
//	@Description	Returns all feature flags (admin only)
//	@Tags			admin,feature-flags
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]featureFlagResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/feature-flags [get]
func (h *FeatureFlagHandler) List(w http.ResponseWriter, r *http.Request) {
	flags, err := h.svc.GetAll(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]featureFlagResponse, len(flags))
	for i, f := range flags {
		result[i] = featureFlagResponse{
			Key:         f.Key,
			Enabled:     f.Enabled,
			Description: f.Description,
			Region:      f.Region,
			UpdatedAt:   f.UpdatedAt,
			UpdatedBy:   f.UpdatedBy,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// Update godoc
//
//	@Summary		Update a feature flag
//	@Description	Updates the enabled state and optional region of a feature flag (admin only)
//	@Tags			admin,feature-flags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			key		path		string						true	"Feature flag key"
//	@Param			body	body		updateFeatureFlagRequest	true	"New state"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/feature-flags/{key} [put]
func (h *FeatureFlagHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "feature flag key is required")
		return
	}

	var req updateFeatureFlagRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.svc.SetFlag(r.Context(), key, req.Enabled, req.Region, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "feature flag updated"})
}
