package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/service"
)

type AmenityHandler struct {
	svc service.AmenityService
}

func NewAmenityHandler(svc service.AmenityService) *AmenityHandler {
	return &AmenityHandler{svc: svc}
}

type amenityResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createAmenityRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

type updateAmenityRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

func toAmenityResponse(a *domain.Amenity) amenityResponse {
	return amenityResponse{
		ID:        a.ID.String(),
		Name:      a.Name,
		Icon:      a.Icon,
		SortOrder: a.SortOrder,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ListAmenities godoc
//
//	@Summary		List amenities
//	@Description	Returns all amenities
//	@Tags			admin,amenities
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]amenityResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/amenities [get]
func (h *AmenityHandler) ListAmenities(w http.ResponseWriter, r *http.Request) {
	amenities, err := h.svc.ListAll(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]amenityResponse, len(amenities))
	for i := range amenities {
		result[i] = toAmenityResponse(&amenities[i])
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateAmenity godoc
//
//	@Summary		Create amenity
//	@Description	Creates a new amenity
//	@Tags			admin,amenities
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createAmenityRequest	true	"Amenity data"
//	@Success		201		{object}	APIResponse{data=amenityResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/amenities [post]
func (h *AmenityHandler) CreateAmenity(w http.ResponseWriter, r *http.Request) {
	var req createAmenityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	amenity := &domain.Amenity{
		ID:        uuid.New(),
		Name:      req.Name,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
	}

	if err := h.svc.Create(r.Context(), amenity); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAmenityResponse(amenity))
}

// UpdateAmenity godoc
//
//	@Summary		Update amenity
//	@Description	Updates an existing amenity
//	@Tags			admin,amenities
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Amenity ID (UUID)"
//	@Param			body	body		updateAmenityRequest	true	"Updated amenity data"
//	@Success		200		{object}	APIResponse{data=amenityResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/amenities/{id} [put]
func (h *AmenityHandler) UpdateAmenity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid amenity id")
		return
	}

	var req updateAmenityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	amenity := &domain.Amenity{
		ID:        id,
		Name:      req.Name,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
	}

	if err := h.svc.Update(r.Context(), amenity); err != nil {
		handleServiceError(w, err)
		return
	}

	updated, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAmenityResponse(updated))
}

// DeleteAmenity godoc
//
//	@Summary		Delete amenity
//	@Description	Deletes an amenity
//	@Tags			admin,amenities
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Amenity ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/amenities/{id} [delete]
func (h *AmenityHandler) DeleteAmenity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid amenity id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "amenity deleted"})
}
