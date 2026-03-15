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

type CalendarHandler struct {
	calendarSvc service.CalendarService
}

func NewCalendarHandler(calendarSvc service.CalendarService) *CalendarHandler {
	return &CalendarHandler{calendarSvc: calendarSvc}
}

// ExportICal godoc
// @Summary      Export calendar as ICS
// @Description  Exports bathhouse bookings as an iCalendar (.ics) file. Owner or representative only.
// @Tags         calendar
// @Produce      text/calendar
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {string}  string  "ICS file content"
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/calendar.ics [get]
func (h *CalendarHandler) ExportICal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	ical, err := h.calendarSvc.ExportICal(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ical))
}

// ExportICalByToken exports bathhouse bookings as an iCalendar (.ics) file using a shareable token.
// No authentication required. URL: GET /calendar/{token}.ics (outside /api/v1 prefix, not in OpenAPI spec).
func (h *CalendarHandler) ExportICalByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "missing token")
		return
	}

	ical, err := h.calendarSvc.ExportICalByToken(r.Context(), token)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(ical))
}

// GetCalendarToken godoc
// @Summary      Get calendar token
// @Description  Returns or creates a shareable calendar token for ICS export. Owner or representative only.
// @Tags         calendar
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=object}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/calendar-token [get]
func (h *CalendarHandler) GetCalendarToken(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	token, err := h.calendarSvc.GetOrCreateCalendarToken(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
		"url":   "/calendar/" + token + ".ics",
	})
}

// AddExternalCalendar godoc
// @Summary      Add external calendar
// @Description  Adds an external iCal calendar URL for syncing blocked slots. Owner or representative only.
// @Tags         calendar
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Bathhouse ID (UUID)"
// @Param        body  body      object  true  "External calendar data (url, source)"
// @Success      201   {object}  APIResponse{data=object}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/external-calendars [post]
func (h *CalendarHandler) AddExternalCalendar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var input struct {
		URL    string                  `json:"url"`
		Source domain.SlotBlockSource  `json:"source"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	cal, err := h.calendarSvc.AddExternalCalendar(r.Context(), userID, userRole, bathhouseID, input.URL, input.Source)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, cal)
}

// ListExternalCalendars godoc
// @Summary      List external calendars
// @Description  Returns all external calendars linked to a bathhouse. Owner or representative only.
// @Tags         calendar
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=[]object}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/external-calendars [get]
func (h *CalendarHandler) ListExternalCalendars(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	calendars, err := h.calendarSvc.ListExternalCalendars(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, calendars)
}

// RemoveExternalCalendar godoc
// @Summary      Remove external calendar
// @Description  Removes an external calendar link. Owner or representative only.
// @Tags         calendar
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "External calendar ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/external-calendars/{id} [delete]
func (h *CalendarHandler) RemoveExternalCalendar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	calendarID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid calendar id")
		return
	}

	if err := h.calendarSvc.RemoveExternalCalendar(r.Context(), userID, userRole, calendarID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// SyncExternalCalendars godoc
// @Summary      Sync external calendars
// @Description  Manually triggers synchronization of all external calendars for a bathhouse. Owner or representative only.
// @Tags         calendar
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/external-calendars/sync [post]
func (h *CalendarHandler) SyncExternalCalendars(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.calendarSvc.SyncExternalCalendars(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}

// CreateSlotBlock godoc
// @Summary      Create slot block
// @Description  Creates a manual time slot block for a bathhouse, preventing bookings during that period. Owner or representative only.
// @Tags         calendar
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Bathhouse ID (UUID)"
// @Param        body  body      object  true  "Slot block data (start_time, end_time, description)"
// @Success      201   {object}  APIResponse{data=object}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/slot-blocks [post]
func (h *CalendarHandler) CreateSlotBlock(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var input struct {
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		Description string    `json:"description"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	block := &domain.SlotBlock{
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Description: input.Description,
	}

	result, err := h.calendarSvc.CreateSlotBlock(r.Context(), userID, userRole, bathhouseID, block)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// DeleteSlotBlock godoc
// @Summary      Delete slot block
// @Description  Deletes a slot block, freeing the time period for bookings. Owner or representative only.
// @Tags         calendar
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Slot block ID (UUID)"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/slot-blocks/{id} [delete]
func (h *CalendarHandler) DeleteSlotBlock(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	blockID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid block id")
		return
	}

	if err := h.calendarSvc.DeleteSlotBlock(r.Context(), userID, userRole, blockID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
