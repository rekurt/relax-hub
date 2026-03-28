package domain

import (
	"strconv"
	"strings"
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
	BroadcastChannelSMS      BroadcastChannel = "sms"
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
	Clicked     int64
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
		case BroadcastChannelPush, BroadcastChannelEmail, BroadcastChannelTelegram, BroadcastChannelSMS:
		default:
			return ErrInvalidInput
		}
	}
	return nil
}

// BroadcastPersonalizationData holds per-guest data for template token replacement.
type BroadcastPersonalizationData struct {
	GuestName     string
	LastVisitDate string
	VisitCount    int
	PromoCode     string
}

// AvailablePersonalizationTokens returns all supported tokens for broadcast templates.
func AvailablePersonalizationTokens() []string {
	return []string{
		"{{guest_name}}",
		"{{last_visit_date}}",
		"{{visit_count}}",
		"{{promo_code}}",
	}
}

// PersonalizeMessage replaces personalization tokens in a message with guest-specific data.
func PersonalizeMessage(template string, data BroadcastPersonalizationData) string {
	r := strings.NewReplacer(
		"{{guest_name}}", data.GuestName,
		"{{last_visit_date}}", data.LastVisitDate,
		"{{visit_count}}", strconv.Itoa(data.VisitCount),
		"{{promo_code}}", data.PromoCode,
	)
	return r.Replace(template)
}

// HasChannel checks if the broadcast uses a specific channel.
func (b *Broadcast) HasChannel(ch BroadcastChannel) bool {
	for _, c := range b.Channels {
		if c == ch {
			return true
		}
	}
	return false
}

type BroadcastFilter struct {
	OwnerID  uuid.UUID
	Status   *BroadcastStatus
	Page     int
	PageSize int
}
