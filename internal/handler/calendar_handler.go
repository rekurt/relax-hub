package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/calendar"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type CalendarHandler struct {
	calendarSvc service.CalendarService
}

// swagger types for calendar endpoints

type calendarTokenResponse struct {
	Token string `json:"token" example:"abc123def456"`
	URL   string `json:"url" example:"/calendar/abc123def456.ics"`
}

type addExternalCalendarRequest struct {
	URL    string `json:"url" example:"https://calendar.google.com/calendar/ical/xxx/basic.ics"`
	Source string `json:"source" example:"google_calendar"`
}

type externalCalendarResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BathhouseID string  `json:"bathhouse_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	URL         string  `json:"url" example:"https://calendar.google.com/calendar/ical/xxx/basic.ics"`
	Source      string  `json:"source" example:"google_calendar"`
	LastSyncAt  *string `json:"last_sync_at,omitempty" example:"2026-01-15T10:00:00Z"`
	LastError   string  `json:"last_error,omitempty" example:""`
	CreatedAt   string  `json:"created_at" example:"2026-01-15T10:00:00Z"`
}

type createSlotBlockRequest struct {
	StartTime   string `json:"start_time" example:"2026-03-20T10:00:00Z"`
	EndTime     string `json:"end_time" example:"2026-03-20T14:00:00Z"`
	Description string `json:"description" example:"Maintenance"`
}

type slotBlockResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BathhouseID string `json:"bathhouse_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	StartTime   string `json:"start_time" example:"2026-03-20T10:00:00Z"`
	EndTime     string `json:"end_time" example:"2026-03-20T14:00:00Z"`
	Source      string `json:"source" example:"manual"`
	Description string `json:"description" example:"Maintenance"`
	CreatedAt   string `json:"created_at" example:"2026-01-15T10:00:00Z"`
}

type calendarConflictResponse struct {
	BlockID       string `json:"block_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BlockStart    string `json:"block_start" example:"2026-03-20T10:00:00Z"`
	BlockEnd      string `json:"block_end" example:"2026-03-20T14:00:00Z"`
	BlockSource   string `json:"block_source" example:"google_calendar"`
	BlockSummary  string `json:"block_summary" example:"Personal event"`
	BookingID     string `json:"booking_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	BookingStart  string `json:"booking_start" example:"2026-03-20T11:00:00Z"`
	BookingEnd    string `json:"booking_end" example:"2026-03-20T13:00:00Z"`
	BookingStatus string `json:"booking_status" example:"confirmed"`
}

type syncResultResponse struct {
	Status        string                     `json:"status" example:"synced"`
	ConflictCount int                        `json:"conflict_count" example:"0"`
	Conflicts     []calendarConflictResponse `json:"conflicts,omitempty"`
}

func NewCalendarHandler(calendarSvc service.CalendarService) *CalendarHandler {
	return &CalendarHandler{calendarSvc: calendarSvc}
}

// ExportICal godoc
//
//	@Summary		Export calendar as ICS
//	@Description	Exports bathhouse bookings as an iCalendar (.ics) file. Owner or representative only.
//	@Tags			calendar
//	@Produce		text/calendar
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{string}	string	"ICS file content"
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/calendar.ics [get]
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
//
//	@Summary		Get calendar token
//	@Description	Returns or creates a shareable calendar token for ICS export. The token can be used at /calendar/{token}.ics to access the calendar without authentication. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=calendarTokenResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/calendar-token [get]
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
//
//	@Summary		Add external calendar
//	@Description	Adds an external iCal calendar URL for syncing blocked slots. Owner or representative only.
//	@Tags			calendar
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Bathhouse ID (UUID)"
//	@Param			body	body		addExternalCalendarRequest	true	"External calendar data"
//	@Success		201		{object}	APIResponse{data=externalCalendarResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/external-calendars [post]
func (h *CalendarHandler) AddExternalCalendar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var input struct {
		URL    string                 `json:"url"`
		Source domain.SlotBlockSource `json:"source"`
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
//
//	@Summary		List external calendars
//	@Description	Returns all external calendars linked to a bathhouse. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]externalCalendarResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/external-calendars [get]
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
//
//	@Summary		Remove external calendar
//	@Description	Removes an external calendar link. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"External calendar ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/external-calendars/{id} [delete]
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
//
//	@Summary		Sync external calendars
//	@Description	Manually triggers synchronization of all external calendars for a bathhouse. Returns any conflicts with existing bookings. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=syncResultResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/external-calendars/sync [post]
func (h *CalendarHandler) SyncExternalCalendars(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	conflicts, err := h.calendarSvc.SyncExternalCalendars(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapSyncResult(conflicts))
}

// GetCalendarConflicts godoc
//
//	@Summary		Get calendar conflicts
//	@Description	Returns conflicts between external calendar slot blocks and existing confirmed bookings. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=syncResultResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/calendar-conflicts [get]
func (h *CalendarHandler) GetCalendarConflicts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	conflicts, err := h.calendarSvc.GetCalendarConflicts(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapSyncResult(conflicts))
}

// CreateSlotBlock godoc
//
//	@Summary		Create slot block
//	@Description	Creates a manual time slot block for a bathhouse, preventing bookings during that period. Owner or representative only.
//	@Tags			calendar
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Bathhouse ID (UUID)"
//	@Param			body	body		createSlotBlockRequest	true	"Slot block data"
//	@Success		201		{object}	APIResponse{data=slotBlockResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/slot-blocks [post]
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
//
//	@Summary		Delete slot block
//	@Description	Deletes a slot block, freeing the time period for bookings. Owner or representative only.
//	@Tags			calendar
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Slot block ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/slot-blocks/{id} [delete]
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

func mapSyncResult(conflicts []calendar.CalendarConflict) syncResultResponse {
	resp := syncResultResponse{
		Status:        "synced",
		ConflictCount: len(conflicts),
	}
	for _, c := range conflicts {
		resp.Conflicts = append(resp.Conflicts, calendarConflictResponse{
			BlockID:       c.SlotBlock.ID.String(),
			BlockStart:    c.SlotBlock.StartTime.Format(time.RFC3339),
			BlockEnd:      c.SlotBlock.EndTime.Format(time.RFC3339),
			BlockSource:   string(c.SlotBlock.Source),
			BlockSummary:  c.SlotBlock.Description,
			BookingID:     c.Booking.ID.String(),
			BookingStart:  c.Booking.StartTime.Format(time.RFC3339),
			BookingEnd:    c.Booking.EndTime.Format(time.RFC3339),
			BookingStatus: string(c.Booking.Status),
		})
	}
	return resp
}
