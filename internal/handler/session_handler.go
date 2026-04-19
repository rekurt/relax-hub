package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type SessionHandler struct {
	sessionService service.SessionService
}

func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

type sessionResponse struct {
	ID           string    `json:"id"`
	DeviceInfo   string    `json:"device_info"`
	Browser      string    `json:"browser"`
	IP           string    `json:"ip"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsCurrent    bool      `json:"is_current"`
}

// ListSessions godoc
//
//	@Summary		List active sessions
//	@Description	Returns all active sessions for the authenticated user
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]sessionResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/sessions [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	currentSessionID := middleware.GetSessionID(r.Context())

	sessions, err := h.sessionService.ListSessions(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var resp []sessionResponse
	for _, s := range sessions {
		resp = append(resp, sessionResponse{
			ID:           s.ID.String(),
			DeviceInfo:   s.DeviceInfo,
			Browser:      s.Browser,
			IP:           s.IP,
			LastActiveAt: s.LastActiveAt,
			CreatedAt:    s.CreatedAt,
			IsCurrent:    s.ID == currentSessionID,
		})
	}

	if resp == nil {
		resp = []sessionResponse{}
	}

	writeJSON(w, http.StatusOK, resp)
}

// TerminateSession godoc
//
//	@Summary		Terminate a session
//	@Description	Terminates a specific session by ID
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Session ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/sessions/{id} [delete]
func (h *SessionHandler) TerminateSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid session id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.sessionService.TerminateSession(r.Context(), userID, sessionID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "session terminated"})
}

// TerminateAllOtherSessions godoc
//
//	@Summary		Terminate all other sessions
//	@Description	Terminates all sessions except the current one
//	@Tags			sessions
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/sessions [delete]
func (h *SessionHandler) TerminateAllOtherSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	currentSessionID := middleware.GetSessionID(r.Context())

	if err := h.sessionService.TerminateAllExceptCurrent(r.Context(), userID, currentSessionID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "all other sessions terminated"})
}
