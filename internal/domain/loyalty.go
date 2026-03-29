package domain

import (
	"time"

	"github.com/google/uuid"
)

type LoyaltyLevel string

const (
	LoyaltyBronze   LoyaltyLevel = "bronze"
	LoyaltySilver   LoyaltyLevel = "silver"
	LoyaltyGold     LoyaltyLevel = "gold"
	LoyaltyPlatinum LoyaltyLevel = "platinum"
)

func (l LoyaltyLevel) IsValid() bool {
	switch l {
	case LoyaltyBronze, LoyaltySilver, LoyaltyGold, LoyaltyPlatinum:
		return true
	}
	return false
}

type LoyaltyLevelInfo struct {
	Level           LoyaltyLevel
	MinVisits       int
	PointMultiplier float64
	DiscountPercent int
	CashbackPercent int
}

// loyaltyLevels is unexported to prevent accidental mutation. Use GetAllLoyaltyLevels() for a copy.
var loyaltyLevels = []LoyaltyLevelInfo{
	{Level: LoyaltyPlatinum, MinVisits: 30, PointMultiplier: 2.0, DiscountPercent: 10, CashbackPercent: 10},
	{Level: LoyaltyGold, MinVisits: 15, PointMultiplier: 1.5, DiscountPercent: 5, CashbackPercent: 5},
	{Level: LoyaltySilver, MinVisits: 5, PointMultiplier: 1.2, DiscountPercent: 3, CashbackPercent: 3},
	{Level: LoyaltyBronze, MinVisits: 0, PointMultiplier: 1.0, DiscountPercent: 0, CashbackPercent: 0},
}

// GetAllLoyaltyLevels returns a copy of the loyalty levels (sorted from highest to lowest).
func GetAllLoyaltyLevels() []LoyaltyLevelInfo {
	result := make([]LoyaltyLevelInfo, len(loyaltyLevels))
	copy(result, loyaltyLevels)
	return result
}

func LevelForVisitCount(visitCount int) LoyaltyLevel {
	for _, info := range loyaltyLevels {
		if visitCount >= info.MinVisits {
			return info.Level
		}
	}
	return LoyaltyBronze
}

func GetLoyaltyLevelInfo(level LoyaltyLevel) LoyaltyLevelInfo {
	for _, info := range loyaltyLevels {
		if info.Level == level {
			return info
		}
	}
	return loyaltyLevels[len(loyaltyLevels)-1]
}

type LoyaltyAccount struct {
	UserID      uuid.UUID
	Level       LoyaltyLevel
	Points      int64
	TotalEarned int64
	TotalSpent  int64
	VisitCount  int
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

func (a *LoyaltyAccount) Validate() error {
	if a.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !a.Level.IsValid() {
		return ErrInvalidInput
	}
	if a.Points < 0 {
		return ErrInvalidInput
	}
	if a.TotalEarned < 0 {
		return ErrInvalidInput
	}
	if a.TotalSpent < 0 {
		return ErrInvalidInput
	}
	if a.VisitCount < 0 {
		return ErrInvalidInput
	}
	return nil
}

type LoyaltyTransactionType string

const (
	LoyaltyTransactionEarn   LoyaltyTransactionType = "earn"
	LoyaltyTransactionSpend  LoyaltyTransactionType = "spend"
	LoyaltyTransactionRefund LoyaltyTransactionType = "refund"
)

func (t LoyaltyTransactionType) IsValid() bool {
	switch t {
	case LoyaltyTransactionEarn, LoyaltyTransactionSpend, LoyaltyTransactionRefund:
		return true
	}
	return false
}

type LoyaltyTransaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Type        LoyaltyTransactionType
	Amount      int64
	BookingID   *uuid.UUID
	Description string
	CreatedAt   time.Time
}

func (t *LoyaltyTransaction) Validate() error {
	if t.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !t.Type.IsValid() {
		return ErrInvalidInput
	}
	if t.Amount <= 0 {
		return ErrInvalidInput
	}
	return nil
}
