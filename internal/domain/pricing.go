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
