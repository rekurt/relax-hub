package domain

import (
	"time"

	"github.com/google/uuid"
)

// SeasonalTariff represents a named seasonal pricing period for a bathhouse.
// Seasonal tariffs apply as a multiplicative layer on the base price before dynamic pricing rules.
type SeasonalTariff struct {
	ID          uuid.UUID
	BathhouseID uuid.UUID
	Name        string    // e.g. "Летний сезон", "Новогодние каникулы"
	DateFrom    time.Time // start of seasonal period (date only)
	DateTo      time.Time // end of seasonal period (date only, inclusive)
	Multiplier  float64   // price multiplier, e.g. 1.3 = +30%
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (st *SeasonalTariff) Validate() error {
	if st.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if st.Name == "" || len(st.Name) > 255 {
		return ErrInvalidInput
	}
	if st.DateFrom.IsZero() || st.DateTo.IsZero() {
		return ErrInvalidInput
	}
	if st.DateTo.Before(st.DateFrom) {
		return ErrInvalidInput
	}
	if st.Multiplier <= 0 || st.Multiplier > 10.0 {
		return ErrInvalidInput
	}
	return nil
}

// AppliesToDate checks whether the tariff applies to a given date (date-only comparison).
func (st *SeasonalTariff) AppliesToDate(t time.Time) bool {
	if !st.IsActive {
		return false
	}
	tDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	fromDate := time.Date(st.DateFrom.Year(), st.DateFrom.Month(), st.DateFrom.Day(), 0, 0, 0, 0, st.DateFrom.Location())
	toDate := time.Date(st.DateTo.Year(), st.DateTo.Month(), st.DateTo.Day(), 0, 0, 0, 0, st.DateTo.Location())
	return !tDate.Before(fromDate) && !tDate.After(toDate)
}
