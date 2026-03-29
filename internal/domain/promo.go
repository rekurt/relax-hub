package domain

import (
	"time"

	"github.com/google/uuid"
)

type PromoType string

const (
	PromoTypePercentage  PromoType = "percentage"
	PromoTypeFixedAmount PromoType = "fixed_amount"
	PromoTypeFreeHour    PromoType = "free_hour"
)

func (t PromoType) IsValid() bool {
	switch t {
	case PromoTypePercentage, PromoTypeFixedAmount, PromoTypeFreeHour:
		return true
	}
	return false
}

type PromoCode struct {
	ID          uuid.UUID
	Code        string     // uppercase, unique
	Type        PromoType  // percentage, fixed_amount, free_hour
	Value       int64      // percent (10 = 10%) or amount in kopecks
	BathhouseID *uuid.UUID // nil = global
	CreatorID   uuid.UUID  // owner or admin
	MaxUses     int        // 0 = unlimited
	CurrentUses int
	MinAmount   int64 // minimum booking amount in kopecks
	ValidFrom   time.Time
	ValidUntil  time.Time
	IsActive    bool
	CreatedAt   time.Time
}

func (p *PromoCode) Validate() error {
	if p.Code == "" {
		return ErrInvalidInput
	}
	if !p.Type.IsValid() {
		return ErrInvalidInput
	}
	if p.Value <= 0 {
		return ErrInvalidInput
	}
	if p.Type == PromoTypePercentage && p.Value > 100 {
		return ErrInvalidInput
	}
	if p.CreatorID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.MaxUses < 0 {
		return ErrInvalidInput
	}
	if p.CurrentUses < 0 {
		return ErrInvalidInput
	}
	if p.MinAmount < 0 {
		return ErrInvalidInput
	}
	if !p.ValidUntil.After(p.ValidFrom) {
		return ErrInvalidInput
	}
	return nil
}

type PromoUsage struct {
	ID             uuid.UUID
	PromoCodeID    uuid.UUID
	UserID         uuid.UUID
	BookingID      uuid.UUID
	DiscountAmount int64
	UsedAt         time.Time
}
