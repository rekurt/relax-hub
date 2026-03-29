package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type PricingRuleType string

const (
	RuleTypeWeekday   PricingRuleType = "weekday"
	RuleTypeWeekend   PricingRuleType = "weekend"
	RuleTypePerDay    PricingRuleType = "per_day"
	RuleTypeHoliday   PricingRuleType = "holiday"
	RuleTypeTimeRange PricingRuleType = "time_range"
	RuleTypeSeason    PricingRuleType = "season"
)

func (t PricingRuleType) IsValid() bool {
	switch t {
	case RuleTypeWeekday, RuleTypeWeekend, RuleTypePerDay, RuleTypeHoliday, RuleTypeTimeRange, RuleTypeSeason:
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
	if pr.Name == "" || len(pr.Name) > 255 || strings.TrimSpace(pr.Name) == "" {
		return ErrInvalidInput
	}
	if !pr.Type.IsValid() {
		return ErrInvalidInput
	}
	if pr.Multiplier <= 0 || pr.Multiplier > 999.99 {
		return ErrInvalidInput
	}
	if pr.Priority < 0 {
		return ErrInvalidInput
	}

	// Validate type-specific fields
	switch pr.Type {
	case RuleTypeWeekday, RuleTypeWeekend, RuleTypePerDay:
		if len(pr.DaysOfWeek) == 0 {
			return ErrInvalidInput
		}
		for _, day := range pr.DaysOfWeek {
			if day < 0 || day > 6 {
				return ErrInvalidInput
			}
		}
		// For weekday rules, ensure only weekdays are included (0-4 = Monday-Friday)
		if pr.Type == RuleTypeWeekday {
			for _, day := range pr.DaysOfWeek {
				if day > 4 { // 5=Saturday, 6=Sunday
					return ErrInvalidInput
				}
			}
		}
		// For weekend rules, ensure only weekend days are included (5-6 = Saturday-Sunday)
		if pr.Type == RuleTypeWeekend {
			for _, day := range pr.DaysOfWeek {
				if day < 5 {
					return ErrInvalidInput
				}
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
		// Disallow same time for both boundaries - rule would never apply
		if *pr.TimeFrom == *pr.TimeTo {
			return ErrInvalidInput
		}
		// Allow both from < to (normal) and from > to (wraparound like 22:00-06:00)
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
