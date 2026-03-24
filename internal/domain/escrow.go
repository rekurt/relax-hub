package domain

import (
	"time"

	"github.com/google/uuid"
)

type EscrowStatus string

const (
	EscrowHeld     EscrowStatus = "held"
	EscrowReleased EscrowStatus = "released"
	EscrowDisputed EscrowStatus = "disputed"
	EscrowRefunded EscrowStatus = "refunded"
)

func (s EscrowStatus) IsValid() bool {
	switch s {
	case EscrowHeld, EscrowReleased, EscrowDisputed, EscrowRefunded:
		return true
	}
	return false
}

type Escrow struct {
	ID                uuid.UUID
	BookingID         uuid.UUID
	Amount            int64        // total payment amount in kopecks
	ServiceFee        int64        // platform service fee in kopecks
	Status            EscrowStatus
	ClaimPeriodEndsAt time.Time    // when funds can be released to owner
	ReleasedAt        *time.Time
	CreatedAt         time.Time
}

func (e *Escrow) Validate() error {
	if e.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if e.Amount <= 0 {
		return ErrInvalidInput
	}
	if e.ServiceFee < 0 {
		return ErrInvalidInput
	}
	if e.ServiceFee > e.Amount {
		return ErrInvalidInput
	}
	if !e.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
