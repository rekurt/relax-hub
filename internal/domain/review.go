package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	BookingID   uuid.UUID
	Rating      int
	Text        string
	CreatedAt   time.Time
}

func (r *Review) Validate() error {
	if r.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.Rating < 1 || r.Rating > 5 {
		return ErrInvalidInput
	}
	return nil
}
