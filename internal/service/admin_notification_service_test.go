package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEmailSender implements notification.EmailSender for testing.
type mockEmailSender struct {
	sent []emailCall
}

type emailCall struct {
	To      string
	Subject string
	Body    string
}

func (m *mockEmailSender) Send(_ context.Context, to, subject, body string) error {
	m.sent = append(m.sent, emailCall{To: to, Subject: subject, Body: body})
	return nil
}

func setupAdminNotifTest(t *testing.T) (service.AdminNotificationService, *mock.AdminNotificationRepo, *mock.UserRepo, *mockEmailSender) {
	t.Helper()
	repo := mock.NewAdminNotificationRepo()
	userRepo := mock.NewUserRepo()
	emailSender := &mockEmailSender{}
	log := logger.New(logger.LevelWarn)
	svc := service.NewAdminNotificationService(repo, userRepo, emailSender, log)
	return svc, repo, userRepo, emailSender
}

func TestAdminNotification_Emit_Success(t *testing.T) {
	svc, repo, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	err := svc.Emit(ctx, domain.AdminNotifAntifraudFlag, domain.AdminNotifSeverityCritical,
		"Fraud detected", "Suspicious activity on wallet", map[string]interface{}{"wallet_id": "abc"})
	require.NoError(t, err)

	// AdminNotifAntifraudFlag targets super_admin and finance = 2 notifications
	assert.Equal(t, 2, repo.CreatedCount())
}

func TestAdminNotification_Emit_InvalidType(t *testing.T) {
	svc, repo, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	err := svc.Emit(ctx, domain.AdminNotificationType("bogus"), domain.AdminNotifSeverityInfo,
		"title", "body", nil)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Equal(t, 0, repo.CreatedCount())
}

func TestAdminNotification_Emit_InvalidSeverity(t *testing.T) {
	svc, repo, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	err := svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverity("bogus"),
		"title", "body", nil)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Equal(t, 0, repo.CreatedCount())
}

func TestAdminNotification_Emit_NilData(t *testing.T) {
	svc, repo, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// SystemAlert targets only super_admin = 1 notification
	err := svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityWarning,
		"System check", "Memory high", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.CreatedCount())
}

func TestAdminNotification_Emit_AllTypes_TargetCorrectRoles(t *testing.T) {
	tests := []struct {
		name          string
		notifType     domain.AdminNotificationType
		expectedCount int
	}{
		{"antifraud_flag", domain.AdminNotifAntifraudFlag, 2},                   // super_admin + finance
		{"sla_violation", domain.AdminNotifSLAViolation, 2},                     // super_admin + moderator
		{"reconciliation_mismatch", domain.AdminNotifReconciliationMismatch, 2}, // super_admin + finance
		{"float_drift", domain.AdminNotifFloatDrift, 2},                         // super_admin + finance
		{"ticket_escalation", domain.AdminNotifTicketEscalation, 3},             // super_admin + support_l2 + support_l3
		{"dispute_opened", domain.AdminNotifDisputeOpened, 3},                   // super_admin + support_l2 + support_l3
		{"kyc_pending", domain.AdminNotifKYCPending, 2},                         // super_admin + moderator
		{"system_alert", domain.AdminNotifSystemAlert, 1},                       // super_admin only
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _, _ := setupAdminNotifTest(t)
			ctx := context.Background()

			err := svc.Emit(ctx, tc.notifType, domain.AdminNotifSeverityInfo, "title", "body", nil)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, repo.CreatedCount())
		})
	}
}

func TestAdminNotification_MarkAsRead(t *testing.T) {
	svc, repo, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// Emit a notification
	err := svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityInfo, "Test", "Body", nil)
	require.NoError(t, err)

	// Get the created notification
	result, err := svc.List(ctx, domain.AdminNotificationFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	adminID := uuid.New()
	notifID := result.Items[0].ID

	// Mark as read
	err = svc.MarkAsRead(ctx, notifID, adminID)
	require.NoError(t, err)

	// Verify it's read
	updated, err := repo.GetByID(ctx, notifID)
	require.NoError(t, err)
	assert.True(t, updated.IsRead)
	assert.NotNil(t, updated.ReadAt)
	assert.Equal(t, adminID, *updated.ReadBy)
}

func TestAdminNotification_MarkAsRead_NotFound(t *testing.T) {
	svc, _, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	err := svc.MarkAsRead(ctx, uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAdminNotification_MarkAllAsRead(t *testing.T) {
	svc, _, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// Emit ticket escalation -> super_admin + support_l2 + support_l3
	err := svc.Emit(ctx, domain.AdminNotifTicketEscalation, domain.AdminNotifSeverityWarning, "Escalated", "Ticket escalated", nil)
	require.NoError(t, err)

	adminID := uuid.New()
	// Mark all as read for super_admin role
	err = svc.MarkAllAsRead(ctx, domain.AdminSubRoleSuperAdmin, adminID)
	require.NoError(t, err)

	// super_admin should have 0 unread
	count, err := svc.CountUnread(ctx, domain.AdminSubRoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// support_l2 should still have 1 unread
	count, err = svc.CountUnread(ctx, domain.AdminSubRoleSupportL2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestAdminNotification_CountUnread(t *testing.T) {
	svc, _, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// Initially 0
	count, err := svc.CountUnread(ctx, domain.AdminSubRoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Emit 2 system alerts (each creates 1 for super_admin)
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityInfo, "A", "a", nil))
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityWarning, "B", "b", nil))

	count, err = svc.CountUnread(ctx, domain.AdminSubRoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Finance should have 0 (system_alert only targets super_admin)
	count, err = svc.CountUnread(ctx, domain.AdminSubRoleFinance)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestAdminNotification_ListUnreadCritical(t *testing.T) {
	svc, _, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// Emit info-level -> should not appear in critical list
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityInfo, "Info", "info body", nil))
	// Emit critical-level
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityCritical, "Critical", "critical body", nil))
	// Emit error-level
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityError, "Error", "error body", nil))

	alerts, err := svc.ListUnreadCritical(ctx, domain.AdminSubRoleSuperAdmin)
	require.NoError(t, err)
	assert.Len(t, alerts, 2, "should return only error and critical severity")
}

func TestAdminNotification_SendDailyDigest_NoAlerts(t *testing.T) {
	svc, _, _, emailSender := setupAdminNotifTest(t)
	ctx := context.Background()

	// No alerts -> no emails
	err := svc.SendDailyDigest(ctx)
	require.NoError(t, err)
	assert.Empty(t, emailSender.sent)
}

func TestAdminNotification_SendDailyDigest_WithAlerts(t *testing.T) {
	svc, _, userRepo, emailSender := setupAdminNotifTest(t)
	ctx := context.Background()

	// Create an admin user with super_admin role
	admin := &domain.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Role:         domain.RoleAdmin,
		AdminSubRole: domain.AdminSubRoleSuperAdmin,
		IsActive:     true,
	}
	require.NoError(t, userRepo.Create(ctx, admin))

	// Emit critical alert
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityCritical,
		"Server down", "Primary DB unreachable", nil))

	err := svc.SendDailyDigest(ctx)
	require.NoError(t, err)

	// Should have sent 1 email to the admin
	require.Len(t, emailSender.sent, 1)
	assert.Equal(t, "admin@example.com", emailSender.sent[0].To)
	assert.Contains(t, emailSender.sent[0].Subject, "super_admin")
	assert.Contains(t, emailSender.sent[0].Body, "Server down")
}

func TestAdminNotification_List_WithFilters(t *testing.T) {
	svc, _, _, _ := setupAdminNotifTest(t)
	ctx := context.Background()

	// Emit different types
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifAntifraudFlag, domain.AdminNotifSeverityCritical, "Fraud", "body", nil))
	require.NoError(t, svc.Emit(ctx, domain.AdminNotifSystemAlert, domain.AdminNotifSeverityInfo, "Info", "body", nil))

	// Filter by role=finance -> should get antifraud_flag only (1 notif for finance)
	role := domain.AdminSubRoleFinance
	result, err := svc.List(ctx, domain.AdminNotificationFilter{Role: &role, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, domain.AdminNotifAntifraudFlag, result.Items[0].Type)
}
