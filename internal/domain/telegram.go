package domain

import (
	"time"

	"github.com/google/uuid"
)

// TelegramLink represents a connection between a Telegram account and a platform user
type TelegramLink struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	TelegramID       int64
	TelegramUsername string
	LinkedAt         time.Time
}

// Validate checks if the TelegramLink is valid
func (t *TelegramLink) Validate() error {
	if t.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if t.TelegramID <= 0 {
		return ErrInvalidInput
	}
	return nil
}
