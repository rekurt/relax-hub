package handler

import (
	"net/http"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type RegionHandler struct {
	regionService service.RegionService
}

func NewRegionHandler(regionService service.RegionService) *RegionHandler {
	return &RegionHandler{regionService: regionService}
}

type switchRegionRequest struct {
	Region string `json:"region"`
}

type regionResponse struct {
	Region string `json:"region"`
}

// GetRegion godoc
//
//	@Summary		Get current region
//	@Description	Returns the current region of the authenticated user
//	@Tags			region
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=regionResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/region [get]
func (h *RegionHandler) GetRegion(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	region, err := h.regionService.GetRegion(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, regionResponse{Region: string(region)})
}

// SwitchRegion godoc
//
//	@Summary		Switch region
//	@Description	Switches user region (RU/BY). Blocks if wallet has balance, active bookings, open disputes.
//	@Tags			region
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		switchRegionRequest	true	"New region"
//	@Success		200		{object}	APIResponse{data=regionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/region [put]
func (h *RegionHandler) SwitchRegion(w http.ResponseWriter, r *http.Request) {
	var req switchRegionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	newRegion := domain.UserRegion(req.Region)

	if err := h.regionService.SwitchRegion(r.Context(), userID, newRegion); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, regionResponse{Region: string(newRegion)})
}
