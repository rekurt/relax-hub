package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/google/uuid"
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
	w.Write([]byte(ical))
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
	w.Write([]byte(ical))
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
