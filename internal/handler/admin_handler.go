package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AdminHandler struct {
	userService      service.UserService
	bathhouseService service.BathhouseService
	cityService      service.CityService
}

func NewAdminHandler(
	userService service.UserService,
	bathhouseService service.BathhouseService,
	cityService service.CityService,
) *AdminHandler {
	return &AdminHandler{
		userService:      userService,
		bathhouseService: bathhouseService,
		cityService:      cityService,
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]userResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toUserResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	if err := h.userService.Block(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user blocked"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	if err := h.userService.Unblock(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user unblocked"})
}

func (h *AdminHandler) ApproveBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Approve(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse approved"})
}

func (h *AdminHandler) RejectBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Reject(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse rejected"})
}

func (h *AdminHandler) ListBathhouses(w http.ResponseWriter, r *http.Request) {
	filter := domain.BathhouseFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.BathhouseStatus(v)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status value")
			return
		}
		filter.Status = &status
	} else {
		filter.ShowAllStatuses = true
	}

	result, err := h.bathhouseService.Search(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBathhouseResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

type createCityRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type updateCityRequest struct {
	Name      *string  `json:"name"`
	Slug      *string  `json:"slug"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (h *AdminHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var req createCityRequest
	if err := readJSON(r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Create(r.Context(), service.CreateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
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
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

func (h *AdminHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city id")
		return
	}

	var req updateCityRequest
	if err := readJSON(r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Update(r.Context(), id, service.UpdateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
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
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

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
