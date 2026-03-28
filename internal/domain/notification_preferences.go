package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationEventType represents a specific notification event for per-event preferences.
type NotificationEventType string

const (
	EventBookingConfirmed    NotificationEventType = "booking_confirmed"
	EventBookingCancelled    NotificationEventType = "booking_cancelled"
	EventBookingRejected     NotificationEventType = "booking_rejected"
	EventBookingRequest      NotificationEventType = "booking_request"
	EventBookingReminder24h  NotificationEventType = "booking_reminder_24h"
	EventBookingReminder2h   NotificationEventType = "booking_reminder_2h"
	EventBookingCheckedIn    NotificationEventType = "booking_checked_in"
	EventBookingNoShow       NotificationEventType = "booking_no_show"
	EventBookingExtended     NotificationEventType = "booking_extended"
	EventNewReview           NotificationEventType = "new_review"
	EventReviewResponse      NotificationEventType = "review_response"
	EventReviewApproved      NotificationEventType = "review_approved"
	EventReviewRejected      NotificationEventType = "review_rejected"
	EventReviewRequest       NotificationEventType = "review_request"
	EventPromo               NotificationEventType = "promo"
	EventBroadcast           NotificationEventType = "broadcast"
	EventAutoScenario        NotificationEventType = "auto_scenario"
	EventNewMessage          NotificationEventType = "new_message"
	EventLoyaltyUpgrade      NotificationEventType = "loyalty_upgrade"
	EventReferralBonus       NotificationEventType = "referral_bonus"
	EventBonusExpiring       NotificationEventType = "bonus_expiring"
	EventSubscriptionExpiring NotificationEventType = "subscription_expiring"
	EventPhotoVerified       NotificationEventType = "photo_verified"
	EventPhotoRejected       NotificationEventType = "photo_rejected"
	EventSavedSearchMatch    NotificationEventType = "saved_search_match"
	EventDisputeResolution   NotificationEventType = "dispute_resolution"
	EventPaymentReceived     NotificationEventType = "payment_received"
	EventSystem              NotificationEventType = "system"
)

// AllNotificationEventTypes returns all supported event types for the preferences matrix.
var AllNotificationEventTypes = []NotificationEventType{
	EventBookingConfirmed, EventBookingCancelled, EventBookingRejected,
	EventBookingRequest, EventBookingReminder24h, EventBookingReminder2h,
	EventBookingCheckedIn, EventBookingNoShow, EventBookingExtended,
	EventNewReview, EventReviewResponse, EventReviewApproved,
	EventReviewRejected, EventReviewRequest,
	EventPromo, EventBroadcast, EventAutoScenario,
	EventNewMessage,
	EventLoyaltyUpgrade, EventReferralBonus, EventBonusExpiring,
	EventSubscriptionExpiring,
	EventPhotoVerified, EventPhotoRejected,
	EventSavedSearchMatch,
	EventDisputeResolution, EventPaymentReceived,
	EventSystem,
}

// MandatoryEvents are events that cannot be disabled by the user.
// These include critical transactional notifications.
var MandatoryEvents = map[NotificationEventType]bool{
	EventBookingConfirmed:  true,
	EventBookingCancelled:  true,
	EventBookingRejected:   true,
	EventDisputeResolution: true,
	EventPaymentReceived:   true,
	EventSystem:            true,
}

// CriticalEvents are events where SMS fallback should be used if push/email fail.
var CriticalEvents = map[NotificationEventType]bool{
	EventBookingConfirmed:  true,
	EventBookingCancelled:  true,
	EventDisputeResolution: true,
	EventPaymentReceived:   true,
}

// IsMandatory returns true if this event type cannot be disabled.
func (e NotificationEventType) IsMandatory() bool {
	return MandatoryEvents[e]
}

// IsCritical returns true if this event type should use SMS fallback.
func (e NotificationEventType) IsCritical() bool {
	return CriticalEvents[e]
}

// IsValid checks if the event type is a recognized value.
func (e NotificationEventType) IsValid() bool {
	for _, et := range AllNotificationEventTypes {
		if e == et {
			return true
		}
	}
	return false
}

// MapNotificationTypeToEvent maps a NotificationType to a NotificationEventType.
func MapNotificationTypeToEvent(t NotificationType) NotificationEventType {
	switch t {
	case NotifBookingConfirmed:
		return EventBookingConfirmed
	case NotifBookingCancelled:
		return EventBookingCancelled
	case NotifBookingRejected:
		return EventBookingRejected
	case NotifBookingRequest:
		return EventBookingRequest
	case NotifBookingReminder24h:
		return EventBookingReminder24h
	case NotifBookingReminder2h:
		return EventBookingReminder2h
	case NotifBookingCheckedIn:
		return EventBookingCheckedIn
	case NotifBookingNoShow, NotifBookingNoShowOwner:
		return EventBookingNoShow
	case NotifBookingExtended, NotifBookingExtendedOwner:
		return EventBookingExtended
	case NotifNewReview, NotifClientReview, NotifReviewRevealed:
		return EventNewReview
	case NotifReviewResponse:
		return EventReviewResponse
	case NotifReviewApproved:
		return EventReviewApproved
	case NotifReviewRejected:
		return EventReviewRejected
	case NotifReviewRequest:
		return EventReviewRequest
	case NotifPromo:
		return EventPromo
	case NotifBroadcast:
		return EventBroadcast
	case NotifAutoScenario:
		return EventAutoScenario
	case NotifNewMessage:
		return EventNewMessage
	case NotifLoyaltyUpgrade:
		return EventLoyaltyUpgrade
	case NotifReferralBonus:
		return EventReferralBonus
	case NotifBonusExpiring, NotifBonusExpired:
		return EventBonusExpiring
	case NotifSubscriptionExpiring, NotifSubscriptionExpired:
		return EventSubscriptionExpiring
	case NotifPhotoVerified:
		return EventPhotoVerified
	case NotifPhotoRejected:
		return EventPhotoRejected
	case NotifSavedSearchMatch:
		return EventSavedSearchMatch
	case NotifLowRatingWarning, NotifBathhouseDepublished,
		NotifOwnerCancellationWarning, NotifOwnerCancellationPenalty,
		NotifOwnerResponseRateWarning, NotifOwnerCancellationCompensation,
		NotifAccountDeletionRequested, NotifAccountDeletionReminder, NotifAccountDeletionFinal,
		NotifSystem, NotifReminder, NotifReviewHidden, NotifBookingReminderOwner5min:
		return EventSystem
	default:
		return EventSystem
	}
}

// NotificationEventPreference represents per-event channel preferences.
type NotificationEventPreference struct {
	UserID       uuid.UUID
	EventType    NotificationEventType
	PushEnabled  bool
	EmailEnabled bool
	SMSEnabled   bool
}

// DefaultEventPreference returns default preference for a given event type.
func DefaultEventPreference(userID uuid.UUID, eventType NotificationEventType) NotificationEventPreference {
	return NotificationEventPreference{
		UserID:       userID,
		EventType:    eventType,
		PushEnabled:  true,
		EmailEnabled: true,
		SMSEnabled:   eventType.IsCritical(),
	}
}

// PushDeliveryStatus tracks push notification delivery for fallback chain.
type PushDeliveryStatus string

const (
	PushStatusSent      PushDeliveryStatus = "sent"
	PushStatusDelivered PushDeliveryStatus = "delivered"
	PushStatusFailed    PushDeliveryStatus = "failed"
)

// PushDeliveryLog records push delivery attempts for fallback chain logic.
type PushDeliveryLog struct {
	ID             uuid.UUID
	NotificationID uuid.UUID
	UserID         uuid.UUID
	Status         PushDeliveryStatus
	SentAt         time.Time
	DeliveredAt    *time.Time
	FallbackSent   bool
	CreatedAt      time.Time
}
