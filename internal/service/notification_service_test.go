package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type notifTestEnv struct {
	svc       service.NotificationService
	notifRepo *mock.NotificationRepo
	userRepo  *mock.UserRepo
}

func newNotifTestEnv() *notifTestEnv {
	notifRepo := mock.NewNotificationRepo()
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelError)
	emailSender := notification.NewNoopEmailSender()
	dispatcher := notification.NewDispatcher(notifRepo, emailSender, notification.NewNoopTelegramSender(), mock.NewTelegramLinkRepo(), nil, nil, notification.NewHub(log), log)
	svc := service.NewNotificationService(notifRepo, userRepo, dispatcher, log)
	return &notifTestEnv{
		svc:       svc,
		notifRepo: notifRepo,
		userRepo:  userRepo,
	}
}

func createUser(t *testing.T, repo *mock.UserRepo) *domain.User {
	t.Helper()
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestNotificationService_Send_Success(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	err := env.svc.Send(context.Background(), user.ID, domain.NotifBookingConfirmed,
		"Booking Confirmed", "Your booking has been confirmed", map[string]string{"booking_id": uuid.New().String()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, err := env.svc.GetUnreadCount(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("unread count = %d, want 1", count)
	}
}

func TestNotificationService_Send_InvalidType(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	err := env.svc.Send(context.Background(), user.ID, "invalid_type",
		"Title", "Body", nil)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for invalid type, got: %v", err)
	}
}

func TestNotificationService_Send_EmptyTitle(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	err := env.svc.Send(context.Background(), user.ID, domain.NotifSystem,
		"", "Body", nil)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for empty title, got: %v", err)
	}
}

func TestNotificationService_Send_RespectsPreferences(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	// Disable booking events
	prefs := &domain.NotificationPreferences{
		UserID:        user.ID,
		InApp:         true,
		Email:         false,
		BookingEvents: false,
	}
	if err := env.notifRepo.UpdatePreferences(context.Background(), prefs); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	err := env.svc.Send(context.Background(), user.ID, domain.NotifBookingConfirmed,
		"Booking Confirmed", "Your booking has been confirmed", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Notification should not be created because user opted out
	count, err := env.svc.GetUnreadCount(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0 (user opted out)", count)
	}
}

func TestNotificationService_List(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 1", "Body 1", nil)
	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 2", "Body 2", nil)

	result, err := env.svc.List(context.Background(), user.ID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("totalCount = %d, want 2", result.TotalCount)
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title", "Body", nil)

	result, _ := env.svc.List(context.Background(), user.ID, 1, 10)
	notifID := result.Items[0].ID

	err := env.svc.MarkAsRead(context.Background(), user.ID, notifID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, _ := env.svc.GetUnreadCount(context.Background(), user.ID)
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestNotificationService_MarkAsRead_OtherUserForbidden(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title", "Body", nil)

	result, _ := env.svc.List(context.Background(), user.ID, 1, 10)
	notifID := result.Items[0].ID

	otherUserID := uuid.New()
	err := env.svc.MarkAsRead(context.Background(), otherUserID, notifID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden for other user, got: %v", err)
	}
}

func TestNotificationService_MarkAsRead_NotFound(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	err := env.svc.MarkAsRead(context.Background(), user.ID, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("should be not found, got: %v", err)
	}
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 1", "Body 1", nil)
	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 2", "Body 2", nil)

	err := env.svc.MarkAllAsRead(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, _ := env.svc.GetUnreadCount(context.Background(), user.ID)
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	env := newNotifTestEnv()
	user := createUser(t, env.userRepo)

	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 1", "Body 1", nil)
	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 2", "Body 2", nil)
	_ = env.svc.Send(context.Background(), user.ID, domain.NotifSystem, "Title 3", "Body 3", nil)

	count, err := env.svc.GetUnreadCount(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("unread count = %d, want 3", count)
	}
}

func TestNotificationService_GetPreferences_Defaults(t *testing.T) {
	env := newNotifTestEnv()
	userID := uuid.New()

	prefs, err := env.svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !prefs.InApp {
		t.Error("InApp should default to true")
	}
	if !prefs.Email {
		t.Error("Email should default to true")
	}
	if prefs.Push {
		t.Error("Push should default to false")
	}
	if !prefs.BookingEvents {
		t.Error("BookingEvents should default to true")
	}
	if !prefs.Telegram {
		t.Error("Telegram should default to true")
	}
}

func TestNotificationService_UpdatePreferences(t *testing.T) {
	env := newNotifTestEnv()
	userID := uuid.New()

	prefs := &domain.NotificationPreferences{
		InApp:         true,
		Email:         false,
		Push:          true,
		BookingEvents: true,
		ReviewEvents:  false,
		PromoEvents:   false,
		Reminders:     true,
	}

	err := env.svc.UpdatePreferences(context.Background(), userID, prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := env.svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email {
		t.Error("Email should be false after update")
	}
	if !got.Push {
		t.Error("Push should be true after update")
	}
	if got.ReviewEvents {
		t.Error("ReviewEvents should be false after update")
	}
}

func TestNotificationService_UpdatePreferences_SetsUserID(t *testing.T) {
	env := newNotifTestEnv()
	userID := uuid.New()

	prefs := &domain.NotificationPreferences{
		InApp: true,
	}

	err := env.svc.UpdatePreferences(context.Background(), userID, prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := env.svc.GetPreferences(context.Background(), userID)
	if got.UserID != userID {
		t.Errorf("UserID = %s, want %s", got.UserID, userID)
	}
}
