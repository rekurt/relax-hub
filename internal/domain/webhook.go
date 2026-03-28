package domain

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

type WebhookEventType string

const (
	WebhookEventBookingCreated   WebhookEventType = "booking.created"
	WebhookEventBookingConfirmed WebhookEventType = "booking.confirmed"
	WebhookEventBookingCancelled WebhookEventType = "booking.cancelled"
	WebhookEventBookingCompleted WebhookEventType = "booking.completed"
	WebhookEventPaymentReceived  WebhookEventType = "payment.received"
)

func AllWebhookEventTypes() []WebhookEventType {
	return []WebhookEventType{
		WebhookEventBookingCreated,
		WebhookEventBookingConfirmed,
		WebhookEventBookingCancelled,
		WebhookEventBookingCompleted,
		WebhookEventPaymentReceived,
	}
}

func (t WebhookEventType) IsValid() bool {
	for _, v := range AllWebhookEventTypes() {
		if v == t {
			return true
		}
	}
	return false
}

type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending WebhookDeliveryStatus = "pending"
	WebhookDeliverySuccess WebhookDeliveryStatus = "success"
	WebhookDeliveryFailed  WebhookDeliveryStatus = "failed"
)

const (
	MaxWebhooksPerOwner    = 20
	MaxWebhookRetries      = 3
	WebhookRetryBaseDelay  = 2 // seconds
)

type Webhook struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	URL       string
	Secret    string
	Events    []WebhookEventType
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *Webhook) Validate() error {
	if w.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if w.URL == "" {
		return ErrInvalidInput
	}
	u, err := url.ParseRequestURI(w.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrInvalidInput
	}
	if len(w.Events) == 0 {
		return ErrInvalidInput
	}
	for _, e := range w.Events {
		if !e.IsValid() {
			return ErrInvalidInput
		}
	}
	if w.Secret == "" {
		return ErrInvalidInput
	}
	return nil
}

func (w *Webhook) SubscribedTo(event WebhookEventType) bool {
	for _, e := range w.Events {
		if e == event {
			return true
		}
	}
	return false
}

type WebhookDelivery struct {
	ID           uuid.UUID
	WebhookID    uuid.UUID
	EventType    WebhookEventType
	Payload      []byte
	Status       WebhookDeliveryStatus
	HTTPStatus   int
	ErrorMessage string
	AttemptCount int
	NextRetryAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
