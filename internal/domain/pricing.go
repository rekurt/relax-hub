package domain

import (
	"time"

	"github.com/google/uuid"
)

type PricingRuleType string

const (
	RuleTypeWeekday   PricingRuleType = "weekday"
	RuleTypeWeekend   PricingRuleType = "weekend"
	RuleTypeHoliday   PricingRuleType = "holiday"
	RuleTypeTimeRange PricingRuleType = "time_range"
	RuleTypeSeason    PricingRuleType = "season"
)

func (t PricingRuleType) IsValid() bool {
	switch t {
	case RuleTypeWeekday, RuleTypeWeekend, RuleTypeHoliday, RuleTypeTimeRange, RuleTypeSeason:
		return true
	}
	return false
}

type PricingRule struct {
	ID          uuid.UUID
	BathhouseID uuid.UUID
	Name        string
	Type        PricingRuleType
	Multiplier  float64
	DaysOfWeek  []int
	TimeFrom    *string
	TimeTo      *string
	DateFrom    *time.Time
	DateTo      *time.Time
	Priority    int
	IsActive    bool
	CreatedAt   time.Time
}

func (pr *PricingRule) Validate() error {
	if pr.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if pr.Name == "" {
		return ErrInvalidInput
	}
	if !pr.Type.IsValid() {
		return ErrInvalidInput
	}
	if pr.Multiplier <= 0 {
		return ErrInvalidInput
	}
	if pr.Priority < 0 {
		return ErrInvalidInput
	}

	// Validate type-specific fields
	switch pr.Type {
	case RuleTypeWeekday, RuleTypeWeekend:
		if len(pr.DaysOfWeek) == 0 {
			return ErrInvalidInput
		}
		for _, day := range pr.DaysOfWeek {
			if day < 0 || day > 6 {
				return ErrInvalidInput
			}
		}
	case RuleTypeHoliday:
		if pr.DateFrom == nil || pr.DateTo == nil {
			return ErrInvalidInput
		}
		if pr.DateTo.Before(*pr.DateFrom) {
			return ErrInvalidInput
		}
	case RuleTypeTimeRange:
		if pr.TimeFrom == nil || pr.TimeTo == nil {
			return ErrInvalidInput
		}
		// Validate time format (HH:MM)
		if !isValidTimeFormat(*pr.TimeFrom) || !isValidTimeFormat(*pr.TimeTo) {
			return ErrInvalidInput
		}
		// For non-wraparound times, ensure from < to
		// Note: wraparound times (22:00-06:00) are allowed by not validating the order
		// since we need to support overnight shifts
		if *pr.TimeFrom > *pr.TimeTo && !isWrapAroundTime(*pr.TimeFrom, *pr.TimeTo) {
			// Only reject if it's clearly wrong (e.g., 12:00 to 11:00 non-wraparound)
			// This is handled implicitly - we allow times where from >= to as wraparound
		}
	case RuleTypeSeason:
		if pr.DateFrom == nil || pr.DateTo == nil {
			return ErrInvalidInput
		}
		if pr.DateTo.Before(*pr.DateFrom) {
			return ErrInvalidInput
		}
	}

	return nil
}

// isValidTimeFormat checks if the time string is in HH:MM format
func isValidTimeFormat(timeStr string) bool {
	if len(timeStr) != 5 {
		return false
	}
	if timeStr[2] != ':' {
		return false
	}
	hours := timeStr[0:2]
	minutes := timeStr[3:5]

	// Check if hours and minutes are digits
	for _, c := range hours {
		if c < '0' || c > '9' {
			return false
		}
	}
	for _, c := range minutes {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Parse and validate ranges
	hh := int(hours[0]-'0')*10 + int(hours[1]-'0')
	mm := int(minutes[0]-'0')*10 + int(minutes[1]-'0')

	return hh >= 0 && hh <= 23 && mm >= 0 && mm <= 59
}

// isWrapAroundTime checks if a time range wraps around midnight (e.g., 22:00-06:00)
func isWrapAroundTime(from, to string) bool {
	// Wraparound is when from > to (e.g., "22:00" > "06:00")
	return from > to
}
