package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
	BookingRejected  BookingStatus = "rejected"
	BookingCompleted BookingStatus = "completed"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case BookingPending, BookingConfirmed, BookingCancelled, BookingRejected, BookingCompleted:
		return true
	}
	return false
}

type Booking struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	GuestCount  int
	TotalPrice  int64
	PointsSpent int64
	Status      BookingStatus
	Comment     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (b *Booking) Validate() error {
	if b.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if b.GuestCount <= 0 {
		return ErrInvalidInput
	}
	if b.TotalPrice <= 0 {
		return ErrInvalidInput
	}
	if !b.EndTime.After(b.StartTime) {
		return ErrInvalidInput
	}
	return nil
}
