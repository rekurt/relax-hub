package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/sms"
)

// Dispatcher routes notifications to appropriate delivery channels based on user preferences.
type Dispatcher struct {
	notifRepo       repository.NotificationRepository
	userRepo        repository.UserRepository
	emailSender     EmailSender
	telegramSender  TelegramSender
	telegramRepo    repository.TelegramLinkRepository
	pushSender      PushSender
	deviceTokenRepo repository.DeviceTokenRepository
	smsProvider     sms.Provider
	hub             *Hub
	logger          *logger.Logger
}

// NewDispatcher creates a new notification Dispatcher.
func NewDispatcher(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailSender EmailSender,
	telegramSender TelegramSender,
	telegramRepo repository.TelegramLinkRepository,
	pushSender PushSender,
	deviceTokenRepo repository.DeviceTokenRepository,
	smsProvider sms.Provider,
	hub *Hub,
	log *logger.Logger,
) *Dispatcher {
	return &Dispatcher{
		notifRepo:       notifRepo,
		userRepo:        userRepo,
		emailSender:     emailSender,
		telegramSender:  telegramSender,
		telegramRepo:    telegramRepo,
		pushSender:      pushSender,
		deviceTokenRepo: deviceTokenRepo,
		smsProvider:     smsProvider,
		hub:             hub,
		logger:          log,
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

	eventType := domain.MapNotificationTypeToEvent(notif.Type)
	eventPref, err := d.notifRepo.GetEventPreference(ctx, notif.UserID, eventType)
	if err != nil {
		d.logger.Warn("failed to get event preference, using defaults", "user_id", notif.UserID, "event", eventType, "error", err)
		def := domain.DefaultEventPreference(notif.UserID, eventType)
		eventPref = &def
	}

	if prefs.InApp {
		if err := d.notifRepo.Create(ctx, notif); err != nil {
			d.logger.Error("failed to create in-app notification", "user_id", notif.UserID, "error", err)
		} else if d.hub != nil {
			d.hub.SendToUser(notif.UserID, notif)
		}
	}

	emailEnabled := prefs.Email && eventPref.EmailEnabled
	if emailEnabled && email != "" && d.emailSender != nil {
		if err := d.emailSender.Send(ctx, email, notif.Title, notif.Body); err != nil {
			d.logger.Warn("failed to send email notification", "user_id", notif.UserID, "error", err)
		}
	}

	if prefs.Telegram && d.telegramSender != nil && d.telegramRepo != nil {
		link, err := d.telegramRepo.GetByUserID(ctx, notif.UserID)
		if err == nil && link != nil {
			if err := d.telegramSender.Send(ctx, link.TelegramID, notif.Title, notif.Body); err != nil {
				d.logger.Warn("failed to send telegram notification", "user_id", notif.UserID, "error", err)
			}
		}
	}

	pushEnabled := prefs.Push && eventPref.PushEnabled
	pushSent := false
	if pushEnabled && d.pushSender != nil && d.deviceTokenRepo != nil {
		tokens, err := d.deviceTokenRepo.ListByUser(ctx, notif.UserID)
		if err == nil && len(tokens) > 0 {
			for _, dt := range tokens {
				if err := d.pushSender.Send(ctx, dt.Token, notif.Title, notif.Body, notif.Data); err != nil {
					d.logger.Warn("failed to send push notification", "user_id", notif.UserID, "error", err)
				} else {
					pushSent = true
				}
			}
		}
	}

	// Log push delivery for fallback chain tracking
	if pushEnabled && pushSent {
		deliveryLog := &domain.PushDeliveryLog{
			ID:             uuid.New(),
			NotificationID: notif.ID,
			UserID:         notif.UserID,
			Status:         domain.PushStatusSent,
			SentAt:         time.Now(),
			CreatedAt:      time.Now(),
		}
		if err := d.notifRepo.CreatePushDeliveryLog(ctx, deliveryLog); err != nil {
			d.logger.Warn("failed to create push delivery log", "user_id", notif.UserID, "error", err)
		}
	}

	// For critical events, send SMS immediately if push was not sent and SMS is enabled
	smsEnabled := prefs.SMS && eventPref.SMSEnabled
	if smsEnabled && !pushSent && eventType.IsCritical() && d.smsProvider != nil {
		d.sendSMSFallback(ctx, notif)
	}
}

// DispatchFallback checks for undelivered push notifications older than the threshold
// and sends fallback notifications (email/SMS) for them.
func (d *Dispatcher) DispatchFallback(ctx context.Context, fallbackThreshold time.Duration) {
	olderThan := time.Now().Add(-fallbackThreshold)
	pending, err := d.notifRepo.GetPendingPushDeliveries(ctx, olderThan)
	if err != nil {
		d.logger.Error("failed to get pending push deliveries", "error", err)
		return
	}

	for _, p := range pending {
		notif, err := d.notifRepo.GetByID(ctx, p.NotificationID)
		if err != nil {
			d.logger.Warn("failed to get notification for fallback", "notification_id", p.NotificationID, "error", err)
			continue
		}

		eventType := domain.MapNotificationTypeToEvent(notif.Type)

		prefs, err := d.notifRepo.GetPreferences(ctx, p.UserID)
		if err != nil {
			d.logger.Warn("failed to get preferences for fallback", "user_id", p.UserID, "error", err)
			continue
		}

		eventPref, err := d.notifRepo.GetEventPreference(ctx, p.UserID, eventType)
		if err != nil {
			d.logger.Warn("failed to get event preference for fallback", "user_id", p.UserID, "error", err)
			continue
		}

		// Fallback to email if not already sent via email
		if prefs.Email && eventPref.EmailEnabled && d.emailSender != nil && d.userRepo != nil {
			user, err := d.userRepo.GetByID(ctx, p.UserID)
			if err != nil {
				d.logger.Warn("failed to get user for email fallback", "user_id", p.UserID, "error", err)
			} else if user.Email != "" {
				d.logger.Info("sending email fallback for undelivered push", "user_id", p.UserID, "notification_id", p.NotificationID)
				if err := d.emailSender.Send(ctx, user.Email, notif.Title, notif.Body); err != nil {
					d.logger.Warn("failed to send email fallback", "user_id", p.UserID, "error", err)
				}
			}
		}

		// Fallback to SMS for critical events
		if prefs.SMS && eventPref.SMSEnabled && eventType.IsCritical() && d.smsProvider != nil {
			d.sendSMSFallback(ctx, notif)
		}

		if err := d.notifRepo.MarkPushFallbackSent(ctx, p.ID); err != nil {
			d.logger.Warn("failed to mark push fallback sent", "id", p.ID, "error", err)
		}
	}
}

func (d *Dispatcher) sendSMSFallback(ctx context.Context, notif *domain.Notification) {
	if d.smsProvider == nil {
		return
	}

	// Look up user phone from repository
	var phone string
	if d.userRepo != nil {
		user, err := d.userRepo.GetByID(ctx, notif.UserID)
		if err != nil {
			d.logger.Warn("failed to get user for SMS fallback", "user_id", notif.UserID, "error", err)
			return
		}
		phone = user.Phone
	}
	if phone == "" {
		d.logger.Debug("no phone number for SMS fallback", "user_id", notif.UserID)
		return
	}

	message := notif.Title + ": " + notif.Body
	if len(message) > 160 {
		message = message[:157] + "..."
	}
	if err := d.smsProvider.SendSMS(ctx, phone, message); err != nil {
		d.logger.Warn("failed to send SMS fallback", "user_id", notif.UserID, "error", err)
	}
}
