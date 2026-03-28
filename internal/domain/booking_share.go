package domain

import (
	"time"

	"github.com/google/uuid"
)

// BookingShare represents a shareable booking link with pre-filled parameters.
type BookingShare struct {
	ID          uuid.UUID `json:"id"`
	Token       string    `json:"token"`
	CreatedBy   uuid.UUID `json:"created_by"`
	BathhouseID uuid.UUID `json:"bathhouse_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	GuestCount  int       `json:"guest_count"`
	BookingID   *uuid.UUID `json:"booking_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
