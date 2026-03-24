package handler

import (
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ServiceFeeHandler struct {
	svc service.ServiceFeeService
}

func NewServiceFeeHandler(svc service.ServiceFeeService) *ServiceFeeHandler {
	return &ServiceFeeHandler{svc: svc}
}

type serviceFeeConfigResponse struct {
	ID         string  `json:"id"`
	Region     string  `json:"region"`
	Category   *string `json:"category,omitempty"`
	FeePercent float64 `json:"fee_percent"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type upsertServiceFeeRequest struct {
	Region     string  `json:"region"`
	Category   *string `json:"category,omitempty"`
	FeePercent float64 `json:"fee_percent"`
}

// List godoc
// @Summary      List service fee configs
// @Description  Returns all service fee configurations
// @Tags         admin,service-fee
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  APIResponse{data=[]serviceFeeConfigResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /admin/service-fee [get]
func (h *ServiceFeeHandler) List(w http.ResponseWriter, r *http.Request) {
	configs, err := h.svc.ListConfigs(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]serviceFeeConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = serviceFeeConfigResponse{
			ID:         c.ID.String(),
			Region:     c.Region,
			Category:   c.Category,
			FeePercent: c.FeePercent,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// Upsert godoc
// @Summary      Create or update service fee config
// @Description  Creates or updates a service fee configuration by region and category
// @Tags         admin,service-fee
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      upsertServiceFeeRequest  true  "Service fee config"
// @Success      200   {object}  APIResponse{data=serviceFeeConfigResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /admin/service-fee [put]
func (h *ServiceFeeHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req upsertServiceFeeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	config := &domain.ServiceFeeConfig{
		Region:     req.Region,
		Category:   req.Category,
		FeePercent: req.FeePercent,
	}

	if err := h.svc.UpsertConfig(r.Context(), config); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serviceFeeConfigResponse{
		ID:         config.ID.String(),
		Region:     config.Region,
		Category:   config.Category,
		FeePercent: config.FeePercent,
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  config.UpdatedAt,
	})
}
