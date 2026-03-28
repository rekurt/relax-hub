package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type mockNotificationService struct {
	sendFn              func(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error
	listFn              func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	markAsReadFn        func(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	markAllAsReadFn     func(ctx context.Context, userID uuid.UUID) error
	getUnreadCountFn    func(ctx context.Context, userID uuid.UUID) (int64, error)
	getPreferencesFn    func(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	updatePreferencesFn func(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error
}

func (m *mockNotificationService) Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error {
	if m.sendFn != nil {
		return m.sendFn(ctx, userID, notifType, title, body, data)
	}
	return nil
}

func (m *mockNotificationService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Notification]{}, nil
}

func (m *mockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(ctx, userID, notificationID)
	}
	return nil
}

func (m *mockNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.markAllAsReadFn != nil {
		return m.markAllAsReadFn(ctx, userID)
	}
	return nil
}

func (m *mockNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.getUnreadCountFn != nil {
		return m.getUnreadCountFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockNotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	if m.getPreferencesFn != nil {
		return m.getPreferencesFn(ctx, userID)
	}
	prefs := domain.DefaultNotificationPreferences(userID)
	return &prefs, nil
}

func (m *mockNotificationService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error {
	if m.updatePreferencesFn != nil {
		return m.updatePreferencesFn(ctx, userID, prefs)
	}
	return nil
}

func (m *mockNotificationService) HasRecentByType(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _ time.Time) (bool, error) {
	return false, nil
}

func (m *mockNotificationService) GetEventPreferences(_ context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error) {
	return nil, nil
}

func (m *mockNotificationService) UpdateEventPreferences(_ context.Context, _ uuid.UUID, _ []domain.NotificationEventPreference) error {
	return nil
}

// --- Notification Handler Tests ---

func TestNotificationHandler_List(t *testing.T) {
	userID := uuid.New()
	notifID := uuid.New()
	now := time.Now()

	notifSvc := &mockNotificationService{
		listFn: func(_ context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
			return &domain.PaginatedResult[domain.Notification]{
				Items: []domain.Notification{
					{
						ID:        notifID,
						UserID:    uid,
						Type:      domain.NotifBookingConfirmed,
						Title:     "Бронирование подтверждено",
						Body:      "Ваше бронирование подтверждено",
						CreatedAt: now,
					},
				},
				TotalCount: 1,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 1,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/notifications", h.List)

	req := httptest.NewRequest(http.MethodGet, "/my/notifications?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta in response")
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count=1, got %d", resp.Meta.TotalCount)
	}
}

func TestNotificationHandler_UnreadCount(t *testing.T) {
	userID := uuid.New()

	notifSvc := &mockNotificationService{
		getUnreadCountFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 5, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/notifications/unread-count", h.UnreadCount)

	req := httptest.NewRequest(http.MethodGet, "/my/notifications/unread-count", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	var data map[string]int64
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if data["unread_count"] != 5 {
		t.Errorf("expected unread_count=5, got %d", data["unread_count"])
	}
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	userID := uuid.New()
	notifID := uuid.New()

	notifSvc := &mockNotificationService{
		markAsReadFn: func(_ context.Context, uid uuid.UUID, nid uuid.UUID) error {
			if uid != userID || nid != notifID {
				t.Errorf("unexpected args: uid=%s, nid=%s", uid, nid)
			}
			return nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/my/notifications/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/my/notifications/"+notifID.String()+"/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHandler_MarkAsRead_InvalidID(t *testing.T) {
	userID := uuid.New()
	notifSvc := &mockNotificationService{}
	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/my/notifications/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/my/notifications/invalid-id/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestNotificationHandler_MarkAsRead_NotFound(t *testing.T) {
	userID := uuid.New()
	notifSvc := &mockNotificationService{
		markAsReadFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}
	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/my/notifications/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/my/notifications/"+uuid.New().String()+"/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestNotificationHandler_MarkAsRead_Forbidden(t *testing.T) {
	userID := uuid.New()
	notifSvc := &mockNotificationService{
		markAsReadFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}
	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/my/notifications/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/my/notifications/"+uuid.New().String()+"/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	userID := uuid.New()
	called := false

	notifSvc := &mockNotificationService{
		markAllAsReadFn: func(_ context.Context, uid uuid.UUID) error {
			called = true
			if uid != userID {
				t.Errorf("unexpected userID: %s", uid)
			}
			return nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/my/notifications/read-all", h.MarkAllAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/my/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if !called {
		t.Error("expected markAllAsRead to be called")
	}
}

func TestNotificationHandler_GetPreferences(t *testing.T) {
	userID := uuid.New()

	notifSvc := &mockNotificationService{
		getPreferencesFn: func(_ context.Context, uid uuid.UUID) (*domain.NotificationPreferences, error) {
			return &domain.NotificationPreferences{
				UserID:        uid,
				InApp:         true,
				Email:         true,
				Push:          false,
				BookingEvents: true,
				ReviewEvents:  true,
				PromoEvents:   false,
				Reminders:     true,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/notification-preferences", h.GetPreferences)

	req := httptest.NewRequest(http.MethodGet, "/my/notification-preferences", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	var prefs struct {
		InApp         bool `json:"in_app"`
		Email         bool `json:"email"`
		Push          bool `json:"push"`
		PromoEvents   bool `json:"promo_events"`
		BookingEvents bool `json:"booking_events"`
	}
	if err := json.Unmarshal(resp.Data, &prefs); err != nil {
		t.Fatalf("failed to unmarshal prefs: %v", err)
	}
	if !prefs.InApp {
		t.Error("expected in_app=true")
	}
	if prefs.Push {
		t.Error("expected push=false")
	}
	if prefs.PromoEvents {
		t.Error("expected promo_events=false")
	}
}

func TestNotificationHandler_UpdatePreferences(t *testing.T) {
	userID := uuid.New()
	var updatedPrefs *domain.NotificationPreferences

	notifSvc := &mockNotificationService{
		getPreferencesFn: func(_ context.Context, uid uuid.UUID) (*domain.NotificationPreferences, error) {
			return &domain.NotificationPreferences{
				UserID:        uid,
				InApp:         true,
				Email:         true,
				Push:          false,
				BookingEvents: true,
				ReviewEvents:  true,
				PromoEvents:   true,
				Reminders:     true,
			}, nil
		},
		updatePreferencesFn: func(_ context.Context, _ uuid.UUID, prefs *domain.NotificationPreferences) error {
			updatedPrefs = prefs
			return nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Put("/my/notification-preferences", h.UpdatePreferences)

	body := jsonBody(map[string]interface{}{
		"push":         true,
		"promo_events": false,
	})
	req := httptest.NewRequest(http.MethodPut, "/my/notification-preferences", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	if updatedPrefs == nil {
		t.Fatal("updatePreferences was not called")
	}
	if !updatedPrefs.Push {
		t.Error("expected push=true after update")
	}
	if updatedPrefs.PromoEvents {
		t.Error("expected promo_events=false after update")
	}
	if !updatedPrefs.InApp {
		t.Error("expected in_app to remain true")
	}
}

func TestNotificationHandler_UpdatePreferences_InvalidBody(t *testing.T) {
	userID := uuid.New()
	notifSvc := &mockNotificationService{}
	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Put("/my/notification-preferences", h.UpdatePreferences)

	req := httptest.NewRequest(http.MethodPut, "/my/notification-preferences", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestNotificationHandler_Unauthenticated(t *testing.T) {
	notifSvc := &mockNotificationService{}
	authSvc := &mockAuthService{}
	h := handler.NewNotificationHandler(notifSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/notifications", h.List)

	req := httptest.NewRequest(http.MethodGet, "/my/notifications", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// Verify unused imports are from this file's test package
var _ service.NotificationService = (*mockNotificationService)(nil)
