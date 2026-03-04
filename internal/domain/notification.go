package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifBookingConfirmed NotificationType = "booking_confirmed"
	NotifBookingCancelled NotificationType = "booking_cancelled"
	NotifNewReview        NotificationType = "new_review"
	NotifReviewResponse   NotificationType = "review_response"
	NotifPromo            NotificationType = "promo"
	NotifReminder         NotificationType = "reminder"
	NotifSystem           NotificationType = "system"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotifBookingConfirmed, NotifBookingCancelled, NotifNewReview,
		NotifReviewResponse, NotifPromo, NotifReminder, NotifSystem:
		return true
	}
	return false
}

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      NotificationType
	Title     string
	Body      string
	Data      map[string]string
	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

func (n *Notification) Validate() error {
	if n.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !n.Type.IsValid() {
		return ErrInvalidInput
	}
	if n.Title == "" {
		return ErrInvalidInput
	}
	if n.Body == "" {
		return ErrInvalidInput
	}
	return nil
}

type NotificationPreferences struct {
	UserID        uuid.UUID
	InApp         bool
	Email         bool
	Push          bool
	BookingEvents bool
	ReviewEvents  bool
	PromoEvents   bool
	Reminders     bool
}

func DefaultNotificationPreferences(userID uuid.UUID) NotificationPreferences {
	return NotificationPreferences{
		UserID:        userID,
		InApp:         true,
		Email:         true,
		Push:          false,
		BookingEvents: true,
		ReviewEvents:  true,
		PromoEvents:   true,
		Reminders:     true,
	}
}

func (p *NotificationPreferences) Validate() error {
	if p.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	return nil
}

// WantsEventType checks if the user wants notifications for the given event type.
func (p *NotificationPreferences) WantsEventType(t NotificationType) bool {
	switch t {
	case NotifBookingConfirmed, NotifBookingCancelled:
		return p.BookingEvents
	case NotifNewReview, NotifReviewResponse:
		return p.ReviewEvents
	case NotifPromo:
		return p.PromoEvents
	case NotifReminder:
		return p.Reminders
	case NotifSystem:
		return true
	}
	return false
}

type NotificationFilter struct {
	Page     int
	PageSize int
}
