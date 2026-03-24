package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingPending      BookingStatus = "pending"
	BookingPendingOwner BookingStatus = "pending_owner"
	BookingConfirmed    BookingStatus = "confirmed"
	BookingCancelled    BookingStatus = "cancelled"
	BookingRejected     BookingStatus = "rejected"
	BookingCompleted    BookingStatus = "completed"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case BookingPending, BookingPendingOwner, BookingConfirmed, BookingCancelled, BookingRejected, BookingCompleted:
		return true
	}
	return false
}

type Booking struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	BathhouseID       uuid.UUID
	StartTime         time.Time
	EndTime           time.Time
	GuestCount        int
	TotalPrice        int64
	AddOnTotal        int64
	PointsSpent       int64
	ReferralBonusUsed int64
	BasePrice             int64
	LongSessionDiscount   int64
	ExtraGuestSurcharge   int64
	LastMinuteDiscount    int64
	ServiceFeeAmount      int64
	HoldID                *uuid.UUID
	RejectionReason       string
	Status                BookingStatus
	Comment           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
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
