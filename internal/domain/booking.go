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
	BookingNoShow       BookingStatus = "no_show"
	BookingForceMajeure BookingStatus = "force_majeure_cancelled"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case BookingPending, BookingPendingOwner, BookingConfirmed, BookingCancelled, BookingRejected, BookingCompleted, BookingNoShow, BookingForceMajeure:
		return true
	}
	return false
}

// MaxBookingModifications is the maximum number of times a booking can be modified.
const MaxBookingModifications = 3

type DepositStatus string

const (
	DepositNone     DepositStatus = "none"
	DepositHeld     DepositStatus = "held"
	DepositReleased DepositStatus = "released"
	DepositClaimed  DepositStatus = "claimed"
	DepositDisputed DepositStatus = "disputed"
)

func (s DepositStatus) IsValid() bool {
	switch s {
	case DepositNone, DepositHeld, DepositReleased, DepositClaimed, DepositDisputed:
		return true
	}
	return false
}

type Booking struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	BathhouseID         uuid.UUID
	StartTime           time.Time
	EndTime             time.Time
	GuestCount          int
	TotalPrice          int64
	AddOnTotal          int64
	PointsSpent         int64
	ReferralBonusUsed   int64
	BasePrice           int64
	LongSessionDiscount int64
	ExtraGuestSurcharge int64
	LastMinuteDiscount  int64
	ServiceFeeAmount    int64
	ModificationCount   int
	DepositAmount       int64
	DepositStatus       DepositStatus
	DepositExternalID   string
	DepositReleasedAt   *time.Time
	CheckedInAt         *time.Time
	CheckedOutAt        *time.Time
	HoldID              *uuid.UUID
	RejectionReason     string
	CancelledByOwner    bool
	Status              BookingStatus
	Comment             string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// AdminBookingFilter contains filter parameters for admin booking search.
type AdminBookingFilter struct {
	UserID      *uuid.UUID
	BathhouseID *uuid.UUID
	Status      *BookingStatus
	FromDate    *time.Time
	ToDate      *time.Time
	Page        int
	PageSize    int
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
