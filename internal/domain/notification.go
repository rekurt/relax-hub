package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifBookingConfirmed     NotificationType = "booking_confirmed"
	NotifBookingCancelled     NotificationType = "booking_cancelled"
	NotifBookingRejected      NotificationType = "booking_rejected"
	NotifNewReview            NotificationType = "new_review"
	NotifReviewResponse       NotificationType = "review_response"
	NotifReviewApproved       NotificationType = "review_approved"
	NotifReviewRejected       NotificationType = "review_rejected"
	NotifPromo                NotificationType = "promo"
	NotifReminder             NotificationType = "reminder"
	NotifSystem               NotificationType = "system"
	NotifNewMessage           NotificationType = "new_message"
	NotifPhotoVerified        NotificationType = "photo_verified"
	NotifPhotoRejected        NotificationType = "photo_rejected"
	NotifReviewHidden         NotificationType = "review_hidden"
	NotifLoyaltyUpgrade       NotificationType = "loyalty_upgrade"
	NotifReferralBonus        NotificationType = "referral_bonus"
	NotifSubscriptionExpiring NotificationType = "subscription_expiring"
	NotifSubscriptionExpired  NotificationType = "subscription_expired"
	NotifBonusExpiring            NotificationType = "bonus_expiring"
	NotifBonusExpired             NotificationType = "bonus_expired"
	NotifAccountDeletionRequested NotificationType = "account_deletion_requested"
	NotifAccountDeletionReminder  NotificationType = "account_deletion_reminder"
	NotifAccountDeletionFinal     NotificationType = "account_deletion_final"
	NotifSavedSearchMatch         NotificationType = "saved_search_match"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotifBookingConfirmed, NotifBookingCancelled, NotifBookingRejected, NotifNewReview,
		NotifReviewResponse, NotifReviewApproved, NotifReviewRejected, NotifPromo, NotifReminder, NotifSystem,
		NotifNewMessage, NotifPhotoVerified, NotifPhotoRejected, NotifReviewHidden,
		NotifLoyaltyUpgrade, NotifReferralBonus,
		NotifSubscriptionExpiring, NotifSubscriptionExpired,
		NotifBonusExpiring, NotifBonusExpired,
		NotifAccountDeletionRequested, NotifAccountDeletionReminder, NotifAccountDeletionFinal:
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
	Telegram      bool
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
		Telegram:      true,
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
	case NotifBookingConfirmed, NotifBookingCancelled, NotifBookingRejected,
		NotifLoyaltyUpgrade, NotifReferralBonus,
		NotifSubscriptionExpiring, NotifSubscriptionExpired,
		NotifBonusExpiring, NotifBonusExpired:
		return p.BookingEvents
	case NotifNewReview, NotifReviewResponse, NotifReviewApproved, NotifReviewRejected,
		NotifReviewHidden:
		return p.ReviewEvents
	case NotifPromo:
		return p.PromoEvents
	case NotifReminder:
		return p.Reminders
	case NotifSystem, NotifNewMessage, NotifPhotoVerified, NotifPhotoRejected,
		NotifAccountDeletionRequested, NotifAccountDeletionReminder, NotifAccountDeletionFinal:
		return true
	}
	return false
}

type NotificationFilter struct {
	Page     int
	PageSize int
}
