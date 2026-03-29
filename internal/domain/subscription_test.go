package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionPlanIsValid(t *testing.T) {
	tests := []struct {
		name     string
		plan     SubscriptionPlan
		expected bool
	}{
		{
			name:     "free plan is valid",
			plan:     PlanFree,
			expected: true,
		},
		{
			name:     "premium plan is valid",
			plan:     PlanPremium,
			expected: true,
		},
		{
			name:     "promoted plan is valid",
			plan:     PlanPromoted,
			expected: true,
		},
		{
			name:     "invalid plan",
			plan:     SubscriptionPlan("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.plan.IsValid())
		})
	}
}

func TestSubscriptionStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   SubscriptionStatus
		expected bool
	}{
		{
			name:     "active status is valid",
			status:   SubscriptionActive,
			expected: true,
		},
		{
			name:     "expired status is valid",
			status:   SubscriptionExpired,
			expected: true,
		},
		{
			name:     "cancelled status is valid",
			status:   SubscriptionCancelled,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   SubscriptionStatus("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestSubscriptionValidate(t *testing.T) {
	now := time.Now()
	bathhouseID := uuid.New()
	ownerID := uuid.New()

	tests := []struct {
		name    string
		sub     *Subscription
		wantErr bool
	}{
		{
			name: "valid subscription",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      ownerID,
				Plan:         PlanPremium,
				Status:       SubscriptionActive,
				StartDate:    now,
				PriceKopecks: 100000,
			},
			wantErr: false,
		},
		{
			name: "missing bathhouse id",
			sub: &Subscription{
				BathhouseID:  uuid.Nil,
				OwnerID:      ownerID,
				Plan:         PlanPremium,
				Status:       SubscriptionActive,
				StartDate:    now,
				PriceKopecks: 100000,
			},
			wantErr: true,
		},
		{
			name: "missing owner id",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      uuid.Nil,
				Plan:         PlanPremium,
				Status:       SubscriptionActive,
				StartDate:    now,
				PriceKopecks: 100000,
			},
			wantErr: true,
		},
		{
			name: "invalid plan",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      ownerID,
				Plan:         SubscriptionPlan("invalid"),
				Status:       SubscriptionActive,
				StartDate:    now,
				PriceKopecks: 100000,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      ownerID,
				Plan:         PlanPremium,
				Status:       SubscriptionStatus("invalid"),
				StartDate:    now,
				PriceKopecks: 100000,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      ownerID,
				Plan:         PlanPremium,
				Status:       SubscriptionActive,
				StartDate:    now,
				PriceKopecks: -100,
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			sub: &Subscription{
				BathhouseID:  bathhouseID,
				OwnerID:      ownerID,
				Plan:         PlanPremium,
				Status:       SubscriptionActive,
				StartDate:    now.AddDate(0, 0, 1),
				EndDate:      &now,
				PriceKopecks: 100000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromotionStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   PromotionStatus
		expected bool
	}{
		{
			name:     "active status is valid",
			status:   PromotionActive,
			expected: true,
		},
		{
			name:     "paused status is valid",
			status:   PromotionPaused,
			expected: true,
		},
		{
			name:     "exhausted status is valid",
			status:   PromotionExhausted,
			expected: true,
		},
		{
			name:     "expired status is valid",
			status:   PromotionExpired,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   PromotionStatus("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestPromotionValidate(t *testing.T) {
	now := time.Now()
	bathhouseID := uuid.New()
	cityID := int64(1)

	tests := []struct {
		name    string
		promo   *Promotion
		wantErr bool
	}{
		{
			name: "valid promotion",
			promo: &Promotion{
				BathhouseID:     bathhouseID,
				DailyBidKopecks: MinDailyBidKopecks,
				BudgetKopecks:   100000,
				SpentKopecks:    50000,
				StartDate:       now,
				EndDate:         now.AddDate(0, 0, 30),
				TargetCityID:    &cityID,
				Status:          PromotionActive,
			},
			wantErr: false,
		},
		{
			name: "missing bathhouse id",
			promo: &Promotion{
				BathhouseID:   uuid.Nil,
				BudgetKopecks: 100000,
				SpentKopecks:  50000,
				StartDate:     now,
				EndDate:       now.AddDate(0, 0, 30),
				Status:        PromotionActive,
			},
			wantErr: true,
		},
		{
			name: "zero budget",
			promo: &Promotion{
				BathhouseID:   bathhouseID,
				BudgetKopecks: 0,
				SpentKopecks:  0,
				StartDate:     now,
				EndDate:       now.AddDate(0, 0, 30),
				Status:        PromotionActive,
			},
			wantErr: true,
		},
		{
			name: "negative budget",
			promo: &Promotion{
				BathhouseID:   bathhouseID,
				BudgetKopecks: -1000,
				SpentKopecks:  0,
				StartDate:     now,
				EndDate:       now.AddDate(0, 0, 30),
				Status:        PromotionActive,
			},
			wantErr: true,
		},
		{
			name: "spent more than budget",
			promo: &Promotion{
				BathhouseID:   bathhouseID,
				BudgetKopecks: 100000,
				SpentKopecks:  150000,
				StartDate:     now,
				EndDate:       now.AddDate(0, 0, 30),
				Status:        PromotionActive,
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			promo: &Promotion{
				BathhouseID:   bathhouseID,
				BudgetKopecks: 100000,
				SpentKopecks:  50000,
				StartDate:     now.AddDate(0, 0, 30),
				EndDate:       now,
				Status:        PromotionActive,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			promo: &Promotion{
				BathhouseID:   bathhouseID,
				BudgetKopecks: 100000,
				SpentKopecks:  50000,
				StartDate:     now,
				EndDate:       now.AddDate(0, 0, 30),
				Status:        PromotionStatus("invalid"),
			},
			wantErr: true,
		},
		{
			name: "negative impression count",
			promo: &Promotion{
				BathhouseID:     bathhouseID,
				BudgetKopecks:   100000,
				SpentKopecks:    50000,
				StartDate:       now,
				EndDate:         now.AddDate(0, 0, 30),
				Status:          PromotionActive,
				ImpressionCount: -1,
			},
			wantErr: true,
		},
		{
			name: "more clicks than impressions",
			promo: &Promotion{
				BathhouseID:     bathhouseID,
				BudgetKopecks:   100000,
				SpentKopecks:    50000,
				StartDate:       now,
				EndDate:         now.AddDate(0, 0, 30),
				Status:          PromotionActive,
				ImpressionCount: 100,
				ClickCount:      150,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.promo.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
