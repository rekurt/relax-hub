package domain

import (
	"time"

	"github.com/google/uuid"
)

type Holiday struct {
	ID          uuid.UUID
	Name        string    // e.g. "Новый год", "23 февраля"
	Date        time.Time // specific date (only year-month-day matters)
	Region      string    // "RU", "BY"
	IsRecurring bool      // if true, same month-day every year
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (h *Holiday) Validate() error {
	if h.Name == "" {
		return ErrInvalidInput
	}
	if h.Region != "RU" && h.Region != "BY" {
		return ErrInvalidInput
	}
	if h.Date.IsZero() {
		return ErrInvalidInput
	}
	return nil
}

type BathhouseHolidayPrice struct {
	BathhouseID uuid.UUID
	Multiplier  float64 // range 1.0-2.0, platform default 1.5
}

func (p *BathhouseHolidayPrice) Validate() error {
	if p.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.Multiplier < 1.0 || p.Multiplier > 2.0 {
		return ErrInvalidInput
	}
	return nil
}

// DefaultHolidayMultiplier is the platform default multiplier for holiday pricing.
const DefaultHolidayMultiplier = 1.5
