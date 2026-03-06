package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoyaltyLevelIsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    LoyaltyLevel
		expected bool
	}{
		{name: "bronze is valid", level: LoyaltyBronze, expected: true},
		{name: "silver is valid", level: LoyaltySilver, expected: true},
		{name: "gold is valid", level: LoyaltyGold, expected: true},
		{name: "platinum is valid", level: LoyaltyPlatinum, expected: true},
		{name: "invalid level", level: LoyaltyLevel("diamond"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.IsValid())
		})
	}
}

func TestLevelForVisitCount(t *testing.T) {
	tests := []struct {
		name       string
		visitCount int
		expected   LoyaltyLevel
	}{
		{name: "0 visits is bronze", visitCount: 0, expected: LoyaltyBronze},
		{name: "4 visits is bronze", visitCount: 4, expected: LoyaltyBronze},
		{name: "5 visits is silver", visitCount: 5, expected: LoyaltySilver},
		{name: "14 visits is silver", visitCount: 14, expected: LoyaltySilver},
		{name: "15 visits is gold", visitCount: 15, expected: LoyaltyGold},
		{name: "29 visits is gold", visitCount: 29, expected: LoyaltyGold},
		{name: "30 visits is platinum", visitCount: 30, expected: LoyaltyPlatinum},
		{name: "100 visits is platinum", visitCount: 100, expected: LoyaltyPlatinum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LevelForVisitCount(tt.visitCount))
		})
	}
}

func TestGetLoyaltyLevelInfo(t *testing.T) {
	info := GetLoyaltyLevelInfo(LoyaltyGold)
	assert.Equal(t, LoyaltyGold, info.Level)
	assert.Equal(t, 15, info.MinVisits)
	assert.Equal(t, 1.5, info.PointMultiplier)
	assert.Equal(t, 5, info.DiscountPercent)

	info = GetLoyaltyLevelInfo(LoyaltyBronze)
	assert.Equal(t, 0, info.MinVisits)
	assert.Equal(t, 1.0, info.PointMultiplier)
	assert.Equal(t, 0, info.DiscountPercent)

	// invalid level returns bronze (last in sorted list)
	info = GetLoyaltyLevelInfo(LoyaltyLevel("unknown"))
	assert.Equal(t, LoyaltyBronze, info.Level)
}

func TestLoyaltyAccountValidate(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		account *LoyaltyAccount
		wantErr bool
	}{
		{
			name: "valid account",
			account: &LoyaltyAccount{
				UserID:      userID,
				Level:       LoyaltyBronze,
				Points:      100,
				TotalEarned: 200,
				TotalSpent:  100,
				VisitCount:  3,
			},
			wantErr: false,
		},
		{
			name: "missing user id",
			account: &LoyaltyAccount{
				UserID: uuid.Nil,
				Level:  LoyaltyBronze,
			},
			wantErr: true,
		},
		{
			name: "invalid level",
			account: &LoyaltyAccount{
				UserID: userID,
				Level:  LoyaltyLevel("diamond"),
			},
			wantErr: true,
		},
		{
			name: "negative points",
			account: &LoyaltyAccount{
				UserID: userID,
				Level:  LoyaltyBronze,
				Points: -10,
			},
			wantErr: true,
		},
		{
			name: "negative total earned",
			account: &LoyaltyAccount{
				UserID:      userID,
				Level:       LoyaltyBronze,
				TotalEarned: -1,
			},
			wantErr: true,
		},
		{
			name: "negative total spent",
			account: &LoyaltyAccount{
				UserID:     userID,
				Level:      LoyaltyBronze,
				TotalSpent: -1,
			},
			wantErr: true,
		},
		{
			name: "negative visit count",
			account: &LoyaltyAccount{
				UserID:     userID,
				Level:      LoyaltyBronze,
				VisitCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoyaltyTransactionTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		txType   LoyaltyTransactionType
		expected bool
	}{
		{name: "earn is valid", txType: LoyaltyTransactionEarn, expected: true},
		{name: "spend is valid", txType: LoyaltyTransactionSpend, expected: true},
		{name: "invalid type", txType: LoyaltyTransactionType("refund"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.txType.IsValid())
		})
	}
}

func TestLoyaltyTransactionValidate(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name    string
		tx      *LoyaltyTransaction
		wantErr bool
	}{
		{
			name: "valid earn transaction",
			tx: &LoyaltyTransaction{
				UserID:      userID,
				Type:        LoyaltyTransactionEarn,
				Amount:      100,
				BookingID:   &bookingID,
				Description: "Points for booking",
			},
			wantErr: false,
		},
		{
			name: "valid spend transaction without booking",
			tx: &LoyaltyTransaction{
				UserID:      userID,
				Type:        LoyaltyTransactionSpend,
				Amount:      50,
				Description: "Spent on discount",
			},
			wantErr: false,
		},
		{
			name: "missing user id",
			tx: &LoyaltyTransaction{
				UserID: uuid.Nil,
				Type:   LoyaltyTransactionEarn,
				Amount: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			tx: &LoyaltyTransaction{
				UserID: userID,
				Type:   LoyaltyTransactionType("refund"),
				Amount: 100,
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			tx: &LoyaltyTransaction{
				UserID: userID,
				Type:   LoyaltyTransactionEarn,
				Amount: 0,
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			tx: &LoyaltyTransaction{
				UserID: userID,
				Type:   LoyaltyTransactionEarn,
				Amount: -10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
