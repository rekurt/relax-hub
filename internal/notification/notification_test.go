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
	d := notification.NewDispatcher(notifRepo, emailSender, log)

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
	d := notification.NewDispatcher(notifRepo, emailSender, log)

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
	d := notification.NewDispatcher(notifRepo, emailSender, log)

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
	d := notification.NewDispatcher(notifRepo, emailSender, log)

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

func TestNoopEmailSender_Send(t *testing.T) {
	sender := notification.NewNoopEmailSender()
	err := sender.Send(context.Background(), "test@example.com", "Subject", "Body")
	if err != nil {
		t.Errorf("NoopEmailSender should not return error, got: %v", err)
	}
}
