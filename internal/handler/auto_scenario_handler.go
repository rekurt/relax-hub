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

type AutoScenarioHandler struct {
	scenarioService service.AutoScenarioService
}

func NewAutoScenarioHandler(scenarioService service.AutoScenarioService) *AutoScenarioHandler {
	return &AutoScenarioHandler{scenarioService: scenarioService}
}

type autoScenarioResponse struct {
	ID          string  `json:"id"`
	OwnerID     string  `json:"owner_id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Enabled     bool    `json:"enabled"`
	CustomText  string  `json:"custom_text"`
	Channel     string  `json:"channel"`
	DelayHours  int     `json:"delay_hours"`
	PromoCodeID *string `json:"promo_code_id,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

func toAutoScenarioResponse(s *domain.AutoScenario) autoScenarioResponse {
	name, description, _ := domain.ScenarioMeta(s.Type)
	resp := autoScenarioResponse{
		ID:          s.ID.String(),
		OwnerID:     s.OwnerID.String(),
		Type:        string(s.Type),
		Name:        name,
		Description: description,
		Enabled:     s.Enabled,
		CustomText:  s.CustomText,
		Channel:     string(s.Channel),
		DelayHours:  s.DelayHours,
	}
	if s.PromoCodeID != nil {
		pid := s.PromoCodeID.String()
		resp.PromoCodeID = &pid
	}
	if !s.CreatedAt.IsZero() {
		resp.CreatedAt = s.CreatedAt.Format(time.RFC3339)
	}
	if !s.UpdatedAt.IsZero() {
		resp.UpdatedAt = s.UpdatedAt.Format(time.RFC3339)
	}
	return resp
}

type updateAutoScenarioRequest struct {
	Enabled     *bool   `json:"enabled"`
	CustomText  *string `json:"custom_text"`
	Channel     *string `json:"channel"`
	DelayHours  *int    `json:"delay_hours"`
	PromoCodeID *string `json:"promo_code_id"`
}

// ListAutoScenarios godoc
//
//	@Summary		List auto-scenarios
//	@Description	Get all predefined auto-scenarios with their current configuration
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]autoScenarioResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/auto-scenarios [get]
func (h *AutoScenarioHandler) ListAutoScenarios(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	scenarios, err := h.scenarioService.ListScenarios(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []autoScenarioResponse
	for _, s := range scenarios {
		items = append(items, toAutoScenarioResponse(&s))
	}
	if items == nil {
		items = []autoScenarioResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// UpdateAutoScenario godoc
//
//	@Summary		Update auto-scenario
//	@Description	Update configuration for a specific auto-scenario type
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type	path		string						true	"Scenario type"
//	@Param			body	body		updateAutoScenarioRequest	true	"Scenario configuration"
//	@Success		200		{object}	APIResponse{data=autoScenarioResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/auto-scenarios/{type} [put]
func (h *AutoScenarioHandler) UpdateAutoScenario(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	scenarioType := domain.AutoScenarioType(chi.URLParam(r, "type"))
	if !scenarioType.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_type", "invalid scenario type")
		return
	}

	var req updateAutoScenarioRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	// Get defaults
	_, _, defaultDelay := domain.ScenarioMeta(scenarioType)

	scenario := &domain.AutoScenario{
		ID:         uuid.New(),
		Type:       scenarioType,
		Enabled:    false,
		Channel:    domain.BroadcastChannelPush,
		DelayHours: defaultDelay,
	}

	if req.Enabled != nil {
		scenario.Enabled = *req.Enabled
	}
	if req.CustomText != nil {
		scenario.CustomText = *req.CustomText
	}
	if req.Channel != nil {
		scenario.Channel = domain.BroadcastChannel(*req.Channel)
	}
	if req.DelayHours != nil {
		scenario.DelayHours = *req.DelayHours
	}
	if req.PromoCodeID != nil && *req.PromoCodeID != "" {
		id, err := uuid.Parse(*req.PromoCodeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_promo_code_id", "invalid promo code ID")
			return
		}
		scenario.PromoCodeID = &id
	}

	if err := h.scenarioService.UpdateScenario(r.Context(), userID, role, scenario); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAutoScenarioResponse(scenario))
}
