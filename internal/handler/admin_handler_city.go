package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/service"
)

type createCityRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Region    string  `json:"region"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type updateCityRequest struct {
	Name      *string  `json:"name"`
	Slug      *string  `json:"slug"`
	Region    *string  `json:"region"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// CreateCity godoc
//
//	@Summary		Create city
//	@Description	Create a new city. Admin only.
//	@Tags			admin-cities
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createCityRequest	true	"City data"
//	@Success		201		{object}	APIResponse{data=cityResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/admin/cities [post]
func (h *AdminHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var req createCityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Create(r.Context(), service.CreateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
		Region:    req.Region,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, cityResponse{
		ID:        city.ID,
		Name:      city.Name,
		Slug:      city.Slug,
		Region:    city.Region,
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

// UpdateCity godoc
//
//	@Summary		Update city
//	@Description	Update an existing city. Admin only.
//	@Tags			admin-cities
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"City ID"
//	@Param			body	body		updateCityRequest	true	"Fields to update"
//	@Success		200		{object}	APIResponse{data=cityResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/cities/{id} [put]
func (h *AdminHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city id")
		return
	}

	var req updateCityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Update(r.Context(), id, service.UpdateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
		Region:    req.Region,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cityResponse{
		ID:        city.ID,
		Name:      city.Name,
		Slug:      city.Slug,
		Region:    city.Region,
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

// DeleteCity godoc
//
//	@Summary		Delete city
//	@Description	Delete a city. Admin only.
//	@Tags			admin-cities
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"City ID"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/cities/{id} [delete]
func (h *AdminHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city id")
		return
	}

	if err := h.cityService.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "city deleted"})
}

// Review Moderation Handlers
