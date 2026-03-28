package notification_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

type recordingEmailSender struct {
	calls []emailCall
}

type emailCall struct {
	to      string
	subject string
	body    string
}

func (r *recordingEmailSender) Send(_ context.Context, to, subject, body string) error {
	r.calls = append(r.calls, emailCall{to: to, subject: subject, body: body})
	return nil
}

func TestDispatcher_Dispatch_AllChannelsEnabled(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, nil, mock.NewTelegramLinkRepo(), nil, nil, nil, hub, log)

	userID := uuid.New()
	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifBookingConfirmed,
		Title:  "Booking Confirmed",
		Body:   "Your booking has been confirmed",
	}
	prefs := &domain.NotificationPreferences{
		UserID:        userID,
		InApp:         true,
		Email:         true,
		BookingEvents: true,
	}

	d.Dispatch(context.Background(), notif, prefs, "user@example.com")

	// Check in-app was created
	count, err := notifRepo.CountUnread(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("unread count = %d, want 1", count)
	}

	// Check email was sent
	if len(emailSender.calls) != 1 {
		t.Fatalf("email calls = %d, want 1", len(emailSender.calls))
	}
	if emailSender.calls[0].to != "user@example.com" {
		t.Errorf("email to = %s, want user@example.com", emailSender.calls[0].to)
	}
	if emailSender.calls[0].subject != "Booking Confirmed" {
		t.Errorf("email subject = %s, want 'Booking Confirmed'", emailSender.calls[0].subject)
	}
}

func TestDispatcher_Dispatch_InAppOnly(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, nil, mock.NewTelegramLinkRepo(), nil, nil, nil, hub, log)

	userID := uuid.New()
	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifNewReview,
		Title:  "New Review",
		Body:   "You have a new review",
	}
	prefs := &domain.NotificationPreferences{
		UserID:       userID,
		InApp:        true,
		Email:        false,
		ReviewEvents: true,
	}

	d.Dispatch(context.Background(), notif, prefs, "user@example.com")

	count, err := notifRepo.CountUnread(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("unread count = %d, want 1", count)
	}

	if len(emailSender.calls) != 0 {
		t.Errorf("email calls = %d, want 0", len(emailSender.calls))
	}
}

func TestDispatcher_Dispatch_UserOptedOutOfEventType(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, nil, mock.NewTelegramLinkRepo(), nil, nil, nil, hub, log)

	userID := uuid.New()
	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifPromo,
		Title:  "Special Offer",
		Body:   "50% off today",
	}
	prefs := &domain.NotificationPreferences{
		UserID:      userID,
		InApp:       true,
		Email:       true,
		PromoEvents: false, // Opted out of promo
	}

	d.Dispatch(context.Background(), notif, prefs, "user@example.com")

	count, err := notifRepo.CountUnread(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0 (user opted out)", count)
	}

	if len(emailSender.calls) != 0 {
		t.Errorf("email calls = %d, want 0 (user opted out)", len(emailSender.calls))
	}
}

func TestDispatcher_Dispatch_EmailWithoutAddress(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, nil, mock.NewTelegramLinkRepo(), nil, nil, nil, hub, log)

	userID := uuid.New()
	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifSystem,
		Title:  "System Update",
		Body:   "New features available",
	}
	prefs := &domain.NotificationPreferences{
		UserID: userID,
		InApp:  true,
		Email:  true,
	}

	d.Dispatch(context.Background(), notif, prefs, "") // No email

	count, err := notifRepo.CountUnread(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("unread count = %d, want 1", count)
	}

	if len(emailSender.calls) != 0 {
		t.Errorf("email calls = %d, want 0 (no email address)", len(emailSender.calls))
	}
}

type recordingTelegramSender struct {
	calls []telegramCall
}

type telegramCall struct {
	chatID int64
	title  string
	body   string
}

func (r *recordingTelegramSender) Send(_ context.Context, chatID int64, title, body string) error {
	r.calls = append(r.calls, telegramCall{chatID: chatID, title: title, body: body})
	return nil
}

func TestDispatcher_Dispatch_TelegramEnabled(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	tgSender := &recordingTelegramSender{}
	tgRepo := mock.NewTelegramLinkRepo()
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, tgSender, tgRepo, nil, nil, nil, hub, log)

	userID := uuid.New()
	telegramID := int64(123456789)

	// Create telegram link for user
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       telegramID,
		TelegramUsername: "testuser",
	}
	if err := tgRepo.Create(context.Background(), link); err != nil {
		t.Fatalf("create telegram link: %v", err)
	}

	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifBookingConfirmed,
		Title:  "Booking Confirmed",
		Body:   "Your booking has been confirmed",
	}
	prefs := &domain.NotificationPreferences{
		UserID:        userID,
		InApp:         true,
		Email:         false,
		Telegram:      true,
		BookingEvents: true,
	}

	d.Dispatch(context.Background(), notif, prefs, "")

	// Check telegram was sent
	if len(tgSender.calls) != 1 {
		t.Fatalf("telegram calls = %d, want 1", len(tgSender.calls))
	}
	if tgSender.calls[0].chatID != telegramID {
		t.Errorf("telegram chatID = %d, want %d", tgSender.calls[0].chatID, telegramID)
	}
	if tgSender.calls[0].title != "Booking Confirmed" {
		t.Errorf("telegram title = %s, want 'Booking Confirmed'", tgSender.calls[0].title)
	}
}

func TestDispatcher_Dispatch_TelegramDisabled(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	tgSender := &recordingTelegramSender{}
	tgRepo := mock.NewTelegramLinkRepo()
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, tgSender, tgRepo, nil, nil, nil, hub, log)

	userID := uuid.New()

	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifBookingConfirmed,
		Title:  "Booking Confirmed",
		Body:   "Your booking has been confirmed",
	}
	prefs := &domain.NotificationPreferences{
		UserID:        userID,
		InApp:         true,
		Telegram:      false, // Disabled
		BookingEvents: true,
	}

	d.Dispatch(context.Background(), notif, prefs, "")

	// Check telegram was NOT sent
	if len(tgSender.calls) != 0 {
		t.Errorf("telegram calls = %d, want 0 (telegram disabled)", len(tgSender.calls))
	}
}

func TestDispatcher_Dispatch_TelegramNoLink(t *testing.T) {
	notifRepo := mock.NewNotificationRepo()
	emailSender := &recordingEmailSender{}
	tgSender := &recordingTelegramSender{}
	tgRepo := mock.NewTelegramLinkRepo()
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	d := notification.NewDispatcher(notifRepo, emailSender, tgSender, tgRepo, nil, nil, nil, hub, log)

	userID := uuid.New()

	notif := &domain.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.NotifSystem,
		Title:  "System Update",
		Body:   "New features",
	}
	prefs := &domain.NotificationPreferences{
		UserID:   userID,
		InApp:    true,
		Telegram: true,
	}

	d.Dispatch(context.Background(), notif, prefs, "")

	// No telegram link exists, so no message should be sent
	if len(tgSender.calls) != 0 {
		t.Errorf("telegram calls = %d, want 0 (no telegram link)", len(tgSender.calls))
	}
}
