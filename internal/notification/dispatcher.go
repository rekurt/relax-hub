package notification

import (
	"context"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// Dispatcher routes notifications to appropriate delivery channels based on user preferences.
type Dispatcher struct {
	notifRepo   repository.NotificationRepository
	emailSender EmailSender
	logger      *logger.Logger
}

// NewDispatcher creates a new notification Dispatcher.
func NewDispatcher(
	notifRepo repository.NotificationRepository,
	emailSender EmailSender,
	log *logger.Logger,
) *Dispatcher {
	return &Dispatcher{
		notifRepo:   notifRepo,
		emailSender: emailSender,
		logger:      log,
	}
}

// Dispatch sends a notification through all enabled channels for the user.
// It checks user preferences to determine which channels to use and whether the
// user wants this event type at all.
func (d *Dispatcher) Dispatch(ctx context.Context, notif *domain.Notification, prefs *domain.NotificationPreferences, email string) {
	if !prefs.WantsEventType(notif.Type) {
		d.logger.Debug("user opted out of event type", "user_id", notif.UserID, "type", notif.Type)
		return
	}

	if prefs.InApp {
		if err := d.notifRepo.Create(ctx, notif); err != nil {
			d.logger.Error("failed to create in-app notification", "user_id", notif.UserID, "error", err)
		}
	}

	if prefs.Email && email != "" {
		if err := d.emailSender.Send(ctx, email, notif.Title, notif.Body); err != nil {
			d.logger.Warn("failed to send email notification", "user_id", notif.UserID, "error", err)
		}
	}
}
