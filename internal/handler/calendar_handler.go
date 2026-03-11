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

// ExportICal handles GET /api/v1/my/bathhouses/{id}/calendar.ics
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

// ExportICalByToken handles GET /calendar/{token}.ics
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

// GetCalendarToken handles GET /api/v1/my/bathhouses/{id}/calendar-token
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

// AddExternalCalendar handles POST /api/v1/my/bathhouses/{id}/external-calendars
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

// ListExternalCalendars handles GET /api/v1/my/bathhouses/{id}/external-calendars
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

// RemoveExternalCalendar handles DELETE /api/v1/my/external-calendars/{id}
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

// SyncExternalCalendars handles POST /api/v1/my/bathhouses/{id}/external-calendars/sync
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

// CreateSlotBlock handles POST /api/v1/my/bathhouses/{id}/slot-blocks
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

// DeleteSlotBlock handles DELETE /api/v1/my/slot-blocks/{id}
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
