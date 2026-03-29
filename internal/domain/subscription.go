package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan string

const (
	PlanFree     SubscriptionPlan = "free"
	PlanPremium  SubscriptionPlan = "premium"
	PlanPromoted SubscriptionPlan = "promoted"
)

func (p SubscriptionPlan) IsValid() bool {
	switch p {
	case PlanFree, PlanPremium, PlanPromoted:
		return true
	}
	return false
}

type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionExpired   SubscriptionStatus = "expired"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
)

func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionActive, SubscriptionExpired, SubscriptionCancelled:
		return true
	}
	return false
}

type Subscription struct {
	ID           uuid.UUID
	BathhouseID  uuid.UUID
	OwnerID      uuid.UUID
	Plan         SubscriptionPlan
	Status       SubscriptionStatus
	StartDate    time.Time
	EndDate      *time.Time
	AutoRenew    bool
	PriceKopecks int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Subscription) Validate() error {
	if s.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if s.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if !s.Plan.IsValid() {
		return ErrInvalidInput
	}
	if !s.Status.IsValid() {
		return ErrInvalidInput
	}
	if s.PriceKopecks < 0 {
		return ErrInvalidInput
	}
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return ErrInvalidInput
	}
	return nil
}

type PromotionStatus string

const (
	PromotionActive    PromotionStatus = "active"
	PromotionPaused    PromotionStatus = "paused"
	PromotionExhausted PromotionStatus = "exhausted"
	PromotionExpired   PromotionStatus = "expired"
)

func (p PromotionStatus) IsValid() bool {
	switch p {
	case PromotionActive, PromotionPaused, PromotionExhausted, PromotionExpired:
		return true
	}
	return false
}

type Promotion struct {
	ID              uuid.UUID
	BathhouseID     uuid.UUID
	DailyBidKopecks int64
	BudgetKopecks   int64
	SpentKopecks    int64
	StartDate       time.Time
	EndDate         time.Time
	TargetCityID    *int64
	Status          PromotionStatus
	ImpressionCount int64
	ClickCount      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// MinDailyBidKopecks is the minimum daily bid for a promotion campaign (50 rubles).
const MinDailyBidKopecks int64 = 5000

func (p *Promotion) Validate() error {
	if p.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.DailyBidKopecks < MinDailyBidKopecks {
		return ErrInvalidInput
	}
	if p.BudgetKopecks <= 0 {
		return ErrInvalidInput
	}
	if p.SpentKopecks < 0 {
		return ErrInvalidInput
	}
	if p.SpentKopecks > p.BudgetKopecks {
		return ErrInvalidInput
	}
	if !p.EndDate.After(p.StartDate) {
		return ErrInvalidInput
	}
	if !p.Status.IsValid() {
		return ErrInvalidInput
	}
	if p.ImpressionCount < 0 {
		return ErrInvalidInput
	}
	if p.ClickCount < 0 {
		return ErrInvalidInput
	}
	if p.ClickCount > p.ImpressionCount {
		return ErrInvalidInput
	}
	return nil
}

// RemainingBudget returns the unspent budget in kopecks.
func (p *Promotion) RemainingBudget() int64 {
	return p.BudgetKopecks - p.SpentKopecks
}

// CTR returns the click-through rate as a percentage.
func (p *Promotion) CTR() float64 {
	if p.ImpressionCount == 0 {
		return 0
	}
	return float64(p.ClickCount) / float64(p.ImpressionCount) * 100
}

// CostPerClick returns the average cost per click in kopecks.
func (p *Promotion) CostPerClick() float64 {
	if p.ClickCount == 0 {
		return 0
	}
	return float64(p.SpentKopecks) / float64(p.ClickCount)
}
