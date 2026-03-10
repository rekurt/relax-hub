package domain

import (
	"time"

	"github.com/google/uuid"
)

// DeviceToken represents a push notification subscription for a user's device.
type DeviceToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string // Web Push subscription JSON or native push token
	Platform  string // "web", "android", "ios"
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d *DeviceToken) Validate() error {
	if d.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if d.Token == "" {
		return ErrInvalidInput
	}
	if d.Platform == "" {
		return ErrInvalidInput
	}
	switch d.Platform {
	case "web", "android", "ios":
		// valid
	default:
		return ErrInvalidInput
	}
	return nil
}
