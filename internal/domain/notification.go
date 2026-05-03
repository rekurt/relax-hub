package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifBookingConfirmed              NotificationType = "booking_confirmed"
	NotifBookingCancelled              NotificationType = "booking_cancelled"
	NotifBookingRejected               NotificationType = "booking_rejected"
	NotifNewReview                     NotificationType = "new_review"
	NotifReviewResponse                NotificationType = "review_response"
	NotifReviewApproved                NotificationType = "review_approved"
	NotifReviewRejected                NotificationType = "review_rejected"
	NotifPromo                         NotificationType = "promo"
	NotifReminder                      NotificationType = "reminder"
	NotifSystem                        NotificationType = "system"
	NotifNewMessage                    NotificationType = "new_message"
	NotifPhotoVerified                 NotificationType = "photo_verified"
	NotifPhotoRejected                 NotificationType = "photo_rejected"
	NotifReviewHidden                  NotificationType = "review_hidden"
	NotifLoyaltyUpgrade                NotificationType = "loyalty_upgrade"
	NotifReferralBonus                 NotificationType = "referral_bonus"
	NotifSubscriptionExpiring          NotificationType = "subscription_expiring"
	NotifSubscriptionExpired           NotificationType = "subscription_expired"
	NotifBonusExpiring                 NotificationType = "bonus_expiring"
	NotifBonusExpired                  NotificationType = "bonus_expired"
	NotifAccountDeletionRequested      NotificationType = "account_deletion_requested"
	NotifAccountDeletionReminder       NotificationType = "account_deletion_reminder"
	NotifAccountDeletionFinal          NotificationType = "account_deletion_final"
	NotifSavedSearchMatch              NotificationType = "saved_search_match"
	NotifBookingRequest                NotificationType = "booking_request"
	NotifBookingCheckedIn              NotificationType = "booking_checked_in"
	NotifBookingNoShow                 NotificationType = "booking_no_show"
	NotifBookingNoShowOwner            NotificationType = "booking_no_show_owner"
	NotifBookingReminder24h            NotificationType = "booking_reminder_24h"
	NotifBookingReminder2h             NotificationType = "booking_reminder_2h"
	NotifBookingReminderOwner5min      NotificationType = "booking_reminder_owner_5min"
	NotifBookingExtended               NotificationType = "booking_extended"
	NotifBookingExtendedOwner          NotificationType = "booking_extended_owner"
	NotifOwnerCancellationWarning      NotificationType = "owner_cancellation_warning"
	NotifOwnerCancellationPenalty      NotificationType = "owner_cancellation_penalty"
	NotifOwnerResponseRateWarning      NotificationType = "owner_response_rate_warning"
	NotifOwnerCancellationCompensation NotificationType = "owner_cancellation_compensation"
	NotifReviewRequest                 NotificationType = "review_request"
	NotifLowRatingWarning              NotificationType = "low_rating_warning"
	NotifBathhouseDepublished          NotificationType = "bathhouse_depublished"
	NotifBroadcast                     NotificationType = "broadcast"
	NotifAutoScenario                  NotificationType = "auto_scenario"
	NotifClientReview                  NotificationType = "client_review"
	NotifReviewRevealed                NotificationType = "review_revealed"
	NotifContactInfoFiltered           NotificationType = "contact_info_filtered"
	NotifBookingModificationRequested  NotificationType = "booking_modification_requested"
	NotifBookingModificationApproved   NotificationType = "booking_modification_approved"
	NotifBookingModificationRejected   NotificationType = "booking_modification_rejected"
	NotifBookingModificationExpired    NotificationType = "booking_modification_expired"
	NotifBookingExtensionRequested     NotificationType = "booking_extension_requested"
	NotifBookingExtensionApproved      NotificationType = "booking_extension_approved"
	NotifBookingExtensionRejected      NotificationType = "booking_extension_rejected"
	NotifBookingExtensionExpired       NotificationType = "booking_extension_expired"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotifBookingConfirmed, NotifBookingCancelled, NotifBookingRejected, NotifBookingRequest,
		NotifBookingCheckedIn, NotifBookingNoShow, NotifBookingNoShowOwner,
		NotifBookingReminder24h, NotifBookingReminder2h, NotifBookingReminderOwner5min,
		NotifBookingExtended, NotifBookingExtendedOwner,
		NotifOwnerCancellationWarning, NotifOwnerCancellationPenalty,
		NotifOwnerResponseRateWarning, NotifOwnerCancellationCompensation,
		NotifReviewRequest,
		NotifNewReview,
		NotifReviewResponse, NotifReviewApproved, NotifReviewRejected, NotifPromo, NotifReminder, NotifSystem,
		NotifNewMessage, NotifPhotoVerified, NotifPhotoRejected, NotifReviewHidden,
		NotifLoyaltyUpgrade, NotifReferralBonus,
		NotifSubscriptionExpiring, NotifSubscriptionExpired,
		NotifBonusExpiring, NotifBonusExpired,
		NotifAccountDeletionRequested, NotifAccountDeletionReminder, NotifAccountDeletionFinal,
		NotifSavedSearchMatch,
		NotifLowRatingWarning, NotifBathhouseDepublished,
		NotifBroadcast, NotifAutoScenario,
		NotifClientReview, NotifReviewRevealed,
		NotifContactInfoFiltered,
		NotifBookingModificationRequested, NotifBookingModificationApproved,
		NotifBookingModificationRejected, NotifBookingModificationExpired,
		NotifBookingExtensionRequested, NotifBookingExtensionApproved,
		NotifBookingExtensionRejected, NotifBookingExtensionExpired:
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
	SMS           bool
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
		SMS:           false,
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
	case NotifBookingConfirmed, NotifBookingCancelled, NotifBookingRejected, NotifBookingRequest,
		NotifBookingCheckedIn, NotifBookingNoShow, NotifBookingNoShowOwner,
		NotifBookingExtended, NotifBookingExtendedOwner,
		NotifOwnerCancellationWarning, NotifOwnerCancellationPenalty,
		NotifOwnerResponseRateWarning, NotifOwnerCancellationCompensation,
		NotifLoyaltyUpgrade, NotifReferralBonus,
		NotifSubscriptionExpiring, NotifSubscriptionExpired,
		NotifBonusExpiring, NotifBonusExpired,
		NotifBookingModificationRequested, NotifBookingModificationApproved,
		NotifBookingModificationRejected, NotifBookingModificationExpired,
		NotifBookingExtensionRequested, NotifBookingExtensionApproved,
		NotifBookingExtensionRejected, NotifBookingExtensionExpired:
		return p.BookingEvents
	case NotifNewReview, NotifReviewResponse, NotifReviewApproved, NotifReviewRejected,
		NotifReviewHidden:
		return p.ReviewEvents
	case NotifPromo, NotifBroadcast, NotifAutoScenario:
		return p.PromoEvents
	case NotifReminder, NotifBookingReminder24h, NotifBookingReminder2h, NotifBookingReminderOwner5min,
		NotifReviewRequest:
		return p.Reminders
	case NotifSystem, NotifNewMessage, NotifPhotoVerified, NotifPhotoRejected,
		NotifAccountDeletionRequested, NotifAccountDeletionReminder, NotifAccountDeletionFinal,
		NotifSavedSearchMatch,
		NotifLowRatingWarning, NotifBathhouseDepublished,
		NotifClientReview, NotifReviewRevealed,
		NotifContactInfoFiltered:
		return true
	}
	return false
}

type NotificationFilter struct {
	Page     int
	PageSize int
}
