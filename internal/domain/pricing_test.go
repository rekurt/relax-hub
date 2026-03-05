package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPricingRuleTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		ruleType PricingRuleType
		expected bool
	}{
		{
			name:     "weekday rule type is valid",
			ruleType: RuleTypeWeekday,
			expected: true,
		},
		{
			name:     "weekend rule type is valid",
			ruleType: RuleTypeWeekend,
			expected: true,
		},
		{
			name:     "holiday rule type is valid",
			ruleType: RuleTypeHoliday,
			expected: true,
		},
		{
			name:     "time_range rule type is valid",
			ruleType: RuleTypeTimeRange,
			expected: true,
		},
		{
			name:     "season rule type is valid",
			ruleType: RuleTypeSeason,
			expected: true,
		},
		{
			name:     "invalid rule type",
			ruleType: PricingRuleType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ruleType.IsValid())
		})
	}
}

func TestPricingRuleValidate(t *testing.T) {
	now := time.Now()
	bathhouseID := uuid.New()
	tomorrow := now.AddDate(0, 0, 1)
	timeFrom := "08:00"
	timeTo := "12:00"

	tests := []struct {
		name    string
		rule    *PricingRule
		wantErr bool
	}{
		{
			name: "valid weekday rule",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Будни скидка",
				Type:        RuleTypeWeekday,
				Multiplier:  0.8,
				DaysOfWeek:  []int{0, 1, 2, 3, 4},
				Priority:    1,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "valid weekend rule",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Выходные надбавка",
				Type:        RuleTypeWeekend,
				Multiplier:  1.5,
				DaysOfWeek:  []int{5, 6},
				Priority:    2,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "valid holiday rule",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Новый Год",
				Type:        RuleTypeHoliday,
				Multiplier:  2.0,
				DateFrom:    &now,
				DateTo:      &tomorrow,
				Priority:    10,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "valid time_range rule",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Happy Hour",
				Type:        RuleTypeTimeRange,
				Multiplier:  0.5,
				TimeFrom:    &timeFrom,
				TimeTo:      &timeTo,
				Priority:    3,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "valid season rule",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Летний сезон",
				Type:        RuleTypeSeason,
				Multiplier:  1.3,
				DateFrom:    &now,
				DateTo:      &tomorrow,
				Priority:    5,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "missing bathhouse id",
			rule: &PricingRule{
				BathhouseID: uuid.Nil,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{0},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{0},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "invalid rule type",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        PricingRuleType("invalid"),
				Multiplier:  1.0,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "zero multiplier",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  0,
				DaysOfWeek:  []int{0},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "negative multiplier",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  -1.0,
				DaysOfWeek:  []int{0},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "negative priority",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{0},
				Priority:    -1,
			},
			wantErr: true,
		},
		{
			name: "weekday rule without days_of_week",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "weekday rule with invalid day",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{7},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "weekday rule with negative day",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Rule",
				Type:        RuleTypeWeekday,
				Multiplier:  1.0,
				DaysOfWeek:  []int{-1},
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "holiday rule without date_from",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Holiday",
				Type:        RuleTypeHoliday,
				Multiplier:  1.0,
				DateTo:      &tomorrow,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "holiday rule without date_to",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Holiday",
				Type:        RuleTypeHoliday,
				Multiplier:  1.0,
				DateFrom:    &now,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "holiday rule with date_to before date_from",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Holiday",
				Type:        RuleTypeHoliday,
				Multiplier:  1.0,
				DateFrom:    &tomorrow,
				DateTo:      &now,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "time_range rule without time_from",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Happy Hour",
				Type:        RuleTypeTimeRange,
				Multiplier:  1.0,
				TimeTo:      &timeTo,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "time_range rule without time_to",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Happy Hour",
				Type:        RuleTypeTimeRange,
				Multiplier:  1.0,
				TimeFrom:    &timeFrom,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "season rule without date_from",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Season",
				Type:        RuleTypeSeason,
				Multiplier:  1.0,
				DateTo:      &tomorrow,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "season rule without date_to",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Season",
				Type:        RuleTypeSeason,
				Multiplier:  1.0,
				DateFrom:    &now,
				Priority:    1,
			},
			wantErr: true,
		},
		{
			name: "season rule with date_to before date_from",
			rule: &PricingRule{
				BathhouseID: bathhouseID,
				Name:        "Season",
				Type:        RuleTypeSeason,
				Multiplier:  1.0,
				DateFrom:    &tomorrow,
				DateTo:      &now,
				Priority:    1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
