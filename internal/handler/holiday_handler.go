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

type HolidayHandler struct {
	svc service.HolidayService
}

func NewHolidayHandler(svc service.HolidayService) *HolidayHandler {
	return &HolidayHandler{svc: svc}
}

type holidayResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Date        string    `json:"date"`
	Region      string    `json:"region"`
	IsRecurring bool      `json:"is_recurring"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type createHolidayRequest struct {
	Name        string `json:"name"`
	Date        string `json:"date"`
	Region      string `json:"region"`
	IsRecurring bool   `json:"is_recurring"`
}

type updateHolidayRequest struct {
	Name        string `json:"name"`
	Date        string `json:"date"`
	Region      string `json:"region"`
	IsRecurring bool   `json:"is_recurring"`
}

type setMultiplierRequest struct {
	Multiplier float64 `json:"multiplier"`
}

func toHolidayResponse(h *domain.Holiday) holidayResponse {
	return holidayResponse{
		ID:          h.ID.String(),
		Name:        h.Name,
		Date:        h.Date.Format("2006-01-02"),
		Region:      h.Region,
		IsRecurring: h.IsRecurring,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}
}

// ListHolidays godoc
// @Summary      List holidays
// @Description  Returns all holidays, optionally filtered by region
// @Tags         admin,holidays
// @Produce      json
// @Security     BearerAuth
// @Param        region  query     string  false  "Filter by region (RU, BY)"
// @Success      200     {object}  APIResponse{data=[]holidayResponse}
// @Failure      401     {object}  APIResponse{error=APIError}
// @Failure      403     {object}  APIResponse{error=APIError}
// @Router       /admin/holidays [get]
func (h *HolidayHandler) ListHolidays(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")

	var holidays []domain.Holiday
	var err error
	if region != "" {
		holidays, err = h.svc.ListByRegion(r.Context(), region)
	} else {
		holidays, err = h.svc.ListAll(r.Context())
	}
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]holidayResponse, len(holidays))
	for i := range holidays {
		result[i] = toHolidayResponse(&holidays[i])
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateHoliday godoc
// @Summary      Create holiday
// @Description  Creates a new holiday entry
// @Tags         admin,holidays
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createHolidayRequest  true  "Holiday data"
// @Success      201   {object}  APIResponse{data=holidayResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /admin/holidays [post]
func (h *HolidayHandler) CreateHoliday(w http.ResponseWriter, r *http.Request) {
	var req createHolidayRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date format, use YYYY-MM-DD")
		return
	}

	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        req.Name,
		Date:        date,
		Region:      req.Region,
		IsRecurring: req.IsRecurring,
	}

	if err := h.svc.Create(r.Context(), holiday); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toHolidayResponse(holiday))
}

// UpdateHoliday godoc
// @Summary      Update holiday
// @Description  Updates an existing holiday entry
// @Tags         admin,holidays
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "Holiday ID (UUID)"
// @Param        body  body      updateHolidayRequest  true  "Updated holiday data"
// @Success      200   {object}  APIResponse{data=holidayResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /admin/holidays/{id} [put]
func (h *HolidayHandler) UpdateHoliday(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid holiday id")
		return
	}

	var req updateHolidayRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date format, use YYYY-MM-DD")
		return
	}

	holiday := &domain.Holiday{
		ID:          id,
		Name:        req.Name,
		Date:        date,
		Region:      req.Region,
		IsRecurring: req.IsRecurring,
	}

	if err := h.svc.Update(r.Context(), holiday); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toHolidayResponse(holiday))
}

// DeleteHoliday godoc
// @Summary      Delete holiday
// @Description  Deletes a holiday entry
// @Tags         admin,holidays
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Holiday ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /admin/holidays/{id} [delete]
func (h *HolidayHandler) DeleteHoliday(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid holiday id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "holiday deleted"})
}

// SetBathhouseMultiplier godoc
// @Summary      Set bathhouse holiday multiplier
// @Description  Sets a custom holiday price multiplier for a bathhouse. Owner or representative only.
// @Tags         pricing,holidays
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "Bathhouse ID (UUID)"
// @Param        body  body      setMultiplierRequest  true  "Multiplier data"
// @Success      200   {object}  APIResponse
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/holiday-multiplier [put]
func (h *HolidayHandler) SetBathhouseMultiplier(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req setMultiplierRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.svc.SetBathhouseMultiplier(r.Context(), userID, userRole, bathhouseID, req.Multiplier); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "holiday multiplier updated",
		"multiplier": req.Multiplier,
	})
}
