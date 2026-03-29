package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPromoTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		typ      PromoType
		expected bool
	}{
		{
			name:     "percentage type is valid",
			typ:      PromoTypePercentage,
			expected: true,
		},
		{
			name:     "fixed amount type is valid",
			typ:      PromoTypeFixedAmount,
			expected: true,
		},
		{
			name:     "free hour type is valid",
			typ:      PromoTypeFreeHour,
			expected: true,
		},
		{
			name:     "free addon type is valid",
			typ:      PromoTypeFreeAddon,
			expected: true,
		},
		{
			name:     "invalid type",
			typ:      PromoType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.typ.IsValid())
		})
	}
}

func TestPromoCodeValidate(t *testing.T) {
	now := time.Now()
	creatorID := uuid.New()
	bathhouseID := uuid.New()

	tests := []struct {
		name    string
		promo   *PromoCode
		wantErr bool
	}{
		{
			name: "valid promo code - global",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				MaxUses:    100,
				CurrentUses: 0,
				MinAmount:  0,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
				IsActive:   true,
			},
			wantErr: false,
		},
		{
			name: "valid promo code - bathhouse specific",
			promo: &PromoCode{
				Code:        "BATH20",
				Type:        PromoTypeFixedAmount,
				Value:       5000,
				BathhouseID: &bathhouseID,
				CreatorID:   creatorID,
				MaxUses:     50,
				CurrentUses: 0,
				MinAmount:   10000,
				ValidFrom:   now,
				ValidUntil:  now.AddDate(0, 0, 7),
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "missing code",
			promo: &PromoCode{
				Code:       "",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoType("invalid"),
				Value:      10,
				CreatorID:  creatorID,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "zero value",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      0,
				CreatorID:  creatorID,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "negative value",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      -10,
				CreatorID:  creatorID,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "missing creator id",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  uuid.Nil,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "negative max uses",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				MaxUses:    -1,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "negative current uses",
			promo: &PromoCode{
				Code:        "PROMO10",
				Type:        PromoTypePercentage,
				Value:       10,
				CreatorID:   creatorID,
				CurrentUses: -1,
				ValidFrom:   now,
				ValidUntil:  now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "negative min amount",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				MinAmount:  -100,
				ValidFrom:  now,
				ValidUntil: now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "valid until before valid from",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				ValidFrom:  now.AddDate(0, 0, 30),
				ValidUntil: now,
			},
			wantErr: true,
		},
		{
			name: "valid until equals valid from",
			promo: &PromoCode{
				Code:       "PROMO10",
				Type:       PromoTypePercentage,
				Value:      10,
				CreatorID:  creatorID,
				ValidFrom:  now,
				ValidUntil: now,
			},
			wantErr: true,
		},
		{
			name: "valid free_addon promo",
			promo: &PromoCode{
				Code:          "FREEADD",
				Type:          PromoTypeFreeAddon,
				Value:         0,
				BathhouseID:   &bathhouseID,
				TargetAddOnID: &bathhouseID, // reuse as UUID placeholder
				CreatorID:     creatorID,
				ValidFrom:     now,
				ValidUntil:    now.AddDate(0, 0, 30),
			},
			wantErr: false,
		},
		{
			name: "free_addon without target addon id",
			promo: &PromoCode{
				Code:        "FREEADD2",
				Type:        PromoTypeFreeAddon,
				Value:       0,
				BathhouseID: &bathhouseID,
				CreatorID:   creatorID,
				ValidFrom:   now,
				ValidUntil:  now.AddDate(0, 0, 30),
			},
			wantErr: true,
		},
		{
			name: "free_addon without bathhouse id",
			promo: &PromoCode{
				Code:          "FREEADD3",
				Type:          PromoTypeFreeAddon,
				Value:         0,
				TargetAddOnID: &bathhouseID,
				CreatorID:     creatorID,
				ValidFrom:     now,
				ValidUntil:    now.AddDate(0, 0, 30),
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
