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

type NotificationHandler struct {
	notifService service.NotificationService
}

func NewNotificationHandler(notifService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

type notificationResponse struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data,omitempty"`
	IsRead    bool              `json:"is_read"`
	ReadAt    *time.Time        `json:"read_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

func toNotificationResponse(n *domain.Notification) notificationResponse {
	return notificationResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		Body:      n.Body,
		Data:      n.Data,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

type preferencesResponse struct {
	InApp         bool `json:"in_app"`
	Email         bool `json:"email"`
	Push          bool `json:"push"`
	BookingEvents bool `json:"booking_events"`
	ReviewEvents  bool `json:"review_events"`
	PromoEvents   bool `json:"promo_events"`
	Reminders     bool `json:"reminders"`
}

type updatePreferencesRequest struct {
	InApp         *bool `json:"in_app"`
	Email         *bool `json:"email"`
	Push          *bool `json:"push"`
	BookingEvents *bool `json:"booking_events"`
	ReviewEvents  *bool `json:"review_events"`
	PromoEvents   *bool `json:"promo_events"`
	Reminders     *bool `json:"reminders"`
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.notifService.List(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]notificationResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toNotificationResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	count, err := h.notifService.GetUnreadCount(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid notification id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.notifService.MarkAsRead(r.Context(), userID, notifID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.notifService.MarkAllAsRead(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "all marked as read"})
}

func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	prefs, err := h.notifService.GetPreferences(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, preferencesResponse{
		InApp:         prefs.InApp,
		Email:         prefs.Email,
		Push:          prefs.Push,
		BookingEvents: prefs.BookingEvents,
		ReviewEvents:  prefs.ReviewEvents,
		PromoEvents:   prefs.PromoEvents,
		Reminders:     prefs.Reminders,
	})
}

func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var req updatePreferencesRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	current, err := h.notifService.GetPreferences(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if req.InApp != nil {
		current.InApp = *req.InApp
	}
	if req.Email != nil {
		current.Email = *req.Email
	}
	if req.Push != nil {
		current.Push = *req.Push
	}
	if req.BookingEvents != nil {
		current.BookingEvents = *req.BookingEvents
	}
	if req.ReviewEvents != nil {
		current.ReviewEvents = *req.ReviewEvents
	}
	if req.PromoEvents != nil {
		current.PromoEvents = *req.PromoEvents
	}
	if req.Reminders != nil {
		current.Reminders = *req.Reminders
	}

	if err := h.notifService.UpdatePreferences(r.Context(), userID, current); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, preferencesResponse{
		InApp:         current.InApp,
		Email:         current.Email,
		Push:          current.Push,
		BookingEvents: current.BookingEvents,
		ReviewEvents:  current.ReviewEvents,
		PromoEvents:   current.PromoEvents,
		Reminders:     current.Reminders,
	})
}
