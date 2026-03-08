package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReferralStatus string

const (
	ReferralStatusPending   ReferralStatus = "pending"
	ReferralStatusCompleted ReferralStatus = "completed"
	ReferralStatusExpired   ReferralStatus = "expired"
)

func (s ReferralStatus) IsValid() bool {
	switch s {
	case ReferralStatusPending, ReferralStatusCompleted, ReferralStatusExpired:
		return true
	}
	return false
}

type Referral struct {
	ID           uuid.UUID
	ReferrerID   uuid.UUID
	RefereeID    uuid.UUID
	ReferralCode string
	Status       ReferralStatus
	BonusAmount  int64
	CompletedAt  *time.Time
	CreatedAt    time.Time
}

func (r *Referral) Validate() error {
	if r.ReferrerID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.RefereeID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.ReferralCode == "" {
		return ErrInvalidInput
	}
	if r.ReferrerID == r.RefereeID {
		return ErrSelfReferral
	}
	if r.Status != "" && !r.Status.IsValid() {
		return ErrInvalidInput
	}
	if r.BonusAmount < 0 {
		return ErrInvalidInput
	}
	return nil
}

type ReferralBalance struct {
	UserID      uuid.UUID
	Balance     int64
	TotalEarned int64
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

func (b *ReferralBalance) Validate() error {
	if b.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if b.Balance < 0 {
		return ErrInvalidInput
	}
	if b.TotalEarned < 0 {
		return ErrInvalidInput
	}
	return nil
}

type ReferralStats struct {
	TotalInvited   int
	TotalCompleted int
	TotalEarned    int64
}
