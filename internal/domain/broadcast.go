package domain

import (
	"time"

	"github.com/google/uuid"
)

type BroadcastStatus string

const (
	BroadcastStatusDraft   BroadcastStatus = "draft"
	BroadcastStatusSending BroadcastStatus = "sending"
	BroadcastStatusSent    BroadcastStatus = "sent"
	BroadcastStatusFailed  BroadcastStatus = "failed"
)

type BroadcastChannel string

const (
	BroadcastChannelPush     BroadcastChannel = "push"
	BroadcastChannelEmail    BroadcastChannel = "email"
	BroadcastChannelTelegram BroadcastChannel = "telegram"
)

type Broadcast struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Segment     GuestSegmentSlug
	Title       string
	Body        string
	ImageURL    string
	PromoCodeID *uuid.UUID
	Channels    []BroadcastChannel
	Status      BroadcastStatus
	Delivered   int64
	Read        int64
	SentAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (b *Broadcast) Validate() error {
	if b.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if b.Title == "" || b.Body == "" {
		return ErrInvalidInput
	}
	if len(b.Channels) == 0 {
		return ErrInvalidInput
	}
	// Validate segment
	valid := false
	for _, s := range AllSegments() {
		if s == b.Segment {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidInput
	}
	// Validate channels
	for _, ch := range b.Channels {
		switch ch {
		case BroadcastChannelPush, BroadcastChannelEmail, BroadcastChannelTelegram:
		default:
			return ErrInvalidInput
		}
	}
	return nil
}

type BroadcastFilter struct {
	OwnerID  uuid.UUID
	Status   *BroadcastStatus
	Page     int
	PageSize int
}
