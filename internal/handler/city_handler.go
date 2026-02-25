package handler

import (
	"net/http"

	"github.com/nikitaaldaev/bani/internal/service"
)

type CityHandler struct {
	cityService service.CityService
}

func NewCityHandler(cityService service.CityService) *CityHandler {
	return &CityHandler{cityService: cityService}
}

type cityResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *CityHandler) List(w http.ResponseWriter, r *http.Request) {
	cities, err := h.cityService.GetAll(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]cityResponse, len(cities))
	for i, c := range cities {
		items[i] = cityResponse{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			Latitude:  c.Latitude,
			Longitude: c.Longitude,
		}
	}

	writeJSON(w, http.StatusOK, items)
}
