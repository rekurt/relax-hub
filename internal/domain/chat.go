package domain

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID            uuid.UUID
	BathhouseID   uuid.UUID
	ClientID      uuid.UUID
	BookingID     *uuid.UUID
	LastMessageAt *time.Time
	CreatedAt     time.Time
}

func (c *Conversation) Validate() error {
	if c.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if c.ClientID == uuid.Nil {
		return ErrInvalidInput
	}
	return nil
}

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Text           string
	IsRead         bool
	ReadAt         *time.Time
	CreatedAt      time.Time
}

const MaxMessageTextLength = 4000

func (m *Message) Validate() error {
	if m.ConversationID == uuid.Nil {
		return ErrInvalidInput
	}
	if m.SenderID == uuid.Nil {
		return ErrInvalidInput
	}
	if m.Text == "" {
		return ErrInvalidInput
	}
	if len(m.Text) > MaxMessageTextLength {
		return ErrInvalidInput
	}
	return nil
}
