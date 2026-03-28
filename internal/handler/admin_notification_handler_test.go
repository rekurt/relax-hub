package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock admin notification service ---

type mockAdminNotifCenterService struct {
	emitFn              func(ctx context.Context, notifType domain.AdminNotificationType, severity domain.AdminNotifSeverity, title, body string, data map[string]interface{}) error
	listFn              func(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error)
	markAsReadFn        func(ctx context.Context, notifID, adminID uuid.UUID) error
	markAllAsReadFn     func(ctx context.Context, role domain.AdminSubRole, adminID uuid.UUID) error
	countUnreadFn       func(ctx context.Context, role domain.AdminSubRole) (int64, error)
	listUnreadCritFn    func(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error)
	sendDailyDigestFn   func(ctx context.Context) error
}

func (m *mockAdminNotifCenterService) Emit(ctx context.Context, notifType domain.AdminNotificationType, severity domain.AdminNotifSeverity, title, body string, data map[string]interface{}) error {
	if m.emitFn != nil {
		return m.emitFn(ctx, notifType, severity, title, body, data)
	}
	return nil
}

func (m *mockAdminNotifCenterService) List(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.AdminNotification]{}, nil
}

func (m *mockAdminNotifCenterService) MarkAsRead(ctx context.Context, notifID, adminID uuid.UUID) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(ctx, notifID, adminID)
	}
	return nil
}

func (m *mockAdminNotifCenterService) MarkAllAsRead(ctx context.Context, role domain.AdminSubRole, adminID uuid.UUID) error {
	if m.markAllAsReadFn != nil {
		return m.markAllAsReadFn(ctx, role, adminID)
	}
	return nil
}

func (m *mockAdminNotifCenterService) CountUnread(ctx context.Context, role domain.AdminSubRole) (int64, error) {
	if m.countUnreadFn != nil {
		return m.countUnreadFn(ctx, role)
	}
	return 0, nil
}

func (m *mockAdminNotifCenterService) ListUnreadCritical(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error) {
	if m.listUnreadCritFn != nil {
		return m.listUnreadCritFn(ctx, role)
	}
	return nil, nil
}

func (m *mockAdminNotifCenterService) SendDailyDigest(ctx context.Context) error {
	if m.sendDailyDigestFn != nil {
		return m.sendDailyDigestFn(ctx)
	}
	return nil
}

var _ service.AdminNotificationService = (*mockAdminNotifCenterService)(nil)

func setupAdminNotifRequest(method, path string, userID uuid.UUID, role domain.AdminSubRole) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := middleware.SetUserID(req.Context(), userID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	ctx = middleware.SetAdminSubRole(ctx, role)
	return req.WithContext(ctx)
}

func TestListAdminNotifications_Success(t *testing.T) {
	notifID := uuid.New()
	svc := &mockAdminNotifCenterService{
		listFn: func(_ context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error) {
			assert.NotNil(t, filter.Role)
			assert.Equal(t, domain.AdminSubRoleSuperAdmin, *filter.Role)
			return &domain.PaginatedResult[domain.AdminNotification]{
				Items: []domain.AdminNotification{
					{ID: notifID, Role: domain.AdminSubRoleSuperAdmin, Severity: domain.AdminNotifSeverityWarning, Type: domain.AdminNotifSLAViolation, Title: "Test"},
				},
				Page: 1, PageSize: 20, TotalCount: 1, TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewAdminNotificationHandler(svc)
	req := setupAdminNotifRequest("GET", "/admin/notifications?page=1&page_size=20", uuid.New(), domain.AdminSubRoleSuperAdmin)
	rr := httptest.NewRecorder()
	h.ListAdminNotifications(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestMarkNotificationRead_Success(t *testing.T) {
	notifID := uuid.New()
	adminID := uuid.New()
	var calledID uuid.UUID
	svc := &mockAdminNotifCenterService{
		markAsReadFn: func(_ context.Context, nID, aID uuid.UUID) error {
			calledID = nID
			assert.Equal(t, adminID, aID)
			return nil
		},
	}

	h := handler.NewAdminNotificationHandler(svc)

	r := chi.NewRouter()
	r.Put("/admin/notifications/{id}/read", h.MarkNotificationRead)

	req := setupAdminNotifRequest("PUT", "/admin/notifications/"+notifID.String()+"/read", adminID, domain.AdminSubRoleSuperAdmin)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, notifID, calledID)
}

func TestMarkAllNotificationsRead_Success(t *testing.T) {
	adminID := uuid.New()
	svc := &mockAdminNotifCenterService{
		markAllAsReadFn: func(_ context.Context, role domain.AdminSubRole, aID uuid.UUID) error {
			assert.Equal(t, domain.AdminSubRoleFinance, role)
			assert.Equal(t, adminID, aID)
			return nil
		},
	}

	h := handler.NewAdminNotificationHandler(svc)
	req := setupAdminNotifRequest("PUT", "/admin/notifications/read-all", adminID, domain.AdminSubRoleFinance)
	rr := httptest.NewRecorder()
	h.MarkAllNotificationsRead(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetUnreadCount_Success(t *testing.T) {
	svc := &mockAdminNotifCenterService{
		countUnreadFn: func(_ context.Context, role domain.AdminSubRole) (int64, error) {
			assert.Equal(t, domain.AdminSubRoleSuperAdmin, role)
			return 7, nil
		},
	}

	h := handler.NewAdminNotificationHandler(svc)
	req := setupAdminNotifRequest("GET", "/admin/notifications/unread-count", uuid.New(), domain.AdminSubRoleSuperAdmin)
	rr := httptest.NewRecorder()
	h.GetUnreadCount(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(7), data["unread_count"])
}

func TestListAdminNotifications_InvalidSeverity(t *testing.T) {
	svc := &mockAdminNotifCenterService{}
	h := handler.NewAdminNotificationHandler(svc)
	req := setupAdminNotifRequest("GET", "/admin/notifications?severity=invalid", uuid.New(), domain.AdminSubRoleSuperAdmin)
	rr := httptest.NewRecorder()
	h.ListAdminNotifications(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMarkNotificationRead_InvalidID(t *testing.T) {
	svc := &mockAdminNotifCenterService{}
	h := handler.NewAdminNotificationHandler(svc)

	r := chi.NewRouter()
	r.Put("/admin/notifications/{id}/read", h.MarkNotificationRead)

	req := setupAdminNotifRequest("PUT", "/admin/notifications/not-a-uuid/read", uuid.New(), domain.AdminSubRoleSuperAdmin)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
