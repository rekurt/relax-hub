package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type AdminNotificationHandler struct {
	svc service.AdminNotificationService
}

func NewAdminNotificationHandler(svc service.AdminNotificationService) *AdminNotificationHandler {
	return &AdminNotificationHandler{svc: svc}
}

type adminNotificationResponse struct {
	ID        string      `json:"id"`
	Role      string      `json:"role"`
	Severity  string      `json:"severity"`
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Body      string      `json:"body"`
	Data      interface{} `json:"data,omitempty"`
	IsRead    bool        `json:"is_read"`
	ReadAt    *string     `json:"read_at,omitempty"`
	ReadBy    *string     `json:"read_by,omitempty"`
	CreatedAt string      `json:"created_at"`
}

func toAdminNotificationResponse(n *domain.AdminNotification) adminNotificationResponse {
	resp := adminNotificationResponse{
		ID:        n.ID.String(),
		Role:      string(n.Role),
		Severity:  string(n.Severity),
		Type:      string(n.Type),
		Title:     n.Title,
		Body:      n.Body,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}
	if n.Data != nil {
		resp.Data = n.Data
	}
	if n.ReadAt != nil {
		t := n.ReadAt.Format(time.RFC3339)
		resp.ReadAt = &t
	}
	if n.ReadBy != nil {
		s := n.ReadBy.String()
		resp.ReadBy = &s
	}
	return resp
}

// ListAdminNotifications godoc
//
//	@Summary		List admin notifications
//	@Description	Returns paginated list of admin notifications filtered by role, severity, type
//	@Tags			admin-notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			severity	query		string	false	"Filter by severity (info, warning, error, critical)"
//	@Param			type		query		string	false	"Filter by notification type"
//	@Param			is_read		query		string	false	"Filter by read status (true/false)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]adminNotificationResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/notifications [get]
func (h *AdminNotificationHandler) ListAdminNotifications(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	adminRole := middleware.GetAdminSubRole(r.Context())

	filter := domain.AdminNotificationFilter{
		Role:     &adminRole,
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("severity"); s != "" {
		severity := domain.AdminNotifSeverity(s)
		if !severity.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_severity", "invalid severity value")
			return
		}
		filter.Severity = &severity
	}
	if s := r.URL.Query().Get("type"); s != "" {
		notifType := domain.AdminNotificationType(s)
		if !notifType.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_type", "invalid notification type")
			return
		}
		filter.Type = &notifType
	}
	if s := r.URL.Query().Get("is_read"); s != "" {
		var isRead bool
		switch s {
		case "true":
			isRead = true
		case "false":
			isRead = false
		default:
			writeError(w, http.StatusBadRequest, "invalid_is_read", "is_read must be true or false")
			return
		}
		filter.IsRead = &isRead
	}

	result, err := h.svc.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]adminNotificationResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toAdminNotificationResponse(&result.Items[i]))
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// MarkNotificationRead godoc
//
//	@Summary		Mark admin notification as read
//	@Description	Marks a single admin notification as read
//	@Tags			admin-notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Notification ID (UUID)"
//	@Success		200	{object}	APIResponse{data=statusResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/notifications/{id}/read [put]
func (h *AdminNotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid notification ID")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.svc.MarkAsRead(r.Context(), notifID, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

// MarkAllNotificationsRead godoc
//
//	@Summary		Mark all admin notifications as read
//	@Description	Marks all notifications for the admin's role as read
//	@Tags			admin-notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=statusResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/admin/notifications/read-all [put]
func (h *AdminNotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())
	adminRole := middleware.GetAdminSubRole(r.Context())

	if err := h.svc.MarkAllAsRead(r.Context(), adminRole, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "all_read"})
}

// GetUnreadCount godoc
//
//	@Summary		Get unread notification count
//	@Description	Returns the count of unread admin notifications for the admin's role
//	@Tags			admin-notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=unreadCountResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/admin/notifications/unread-count [get]
func (h *AdminNotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	adminRole := middleware.GetAdminSubRole(r.Context())

	count, err := h.svc.CountUnread(r.Context(), adminRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}
