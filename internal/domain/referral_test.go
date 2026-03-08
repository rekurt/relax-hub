package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestReferralStatus_IsValid(t *testing.T) {
	tests := []struct {
		status ReferralStatus
		valid  bool
	}{
		{ReferralStatusPending, true},
		{ReferralStatusCompleted, true},
		{ReferralStatusExpired, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("ReferralStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestReferral_Validate(t *testing.T) {
	referrerID := uuid.New()
	refereeID := uuid.New()

	valid := &Referral{
		ReferrerID:   referrerID,
		RefereeID:    refereeID,
		ReferralCode: "ABC123",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid referral returned error: %v", err)
	}

	validWithStatus := &Referral{
		ReferrerID:   referrerID,
		RefereeID:    refereeID,
		ReferralCode: "ABC123",
		Status:       ReferralStatusPending,
		BonusAmount:  50000,
	}
	if err := validWithStatus.Validate(); err != nil {
		t.Errorf("valid referral with status returned error: %v", err)
	}

	tests := []struct {
		name     string
		referral Referral
		wantErr  error
	}{
		{
			"nil referrer",
			Referral{ReferrerID: uuid.Nil, RefereeID: refereeID, ReferralCode: "ABC"},
			ErrInvalidInput,
		},
		{
			"nil referee",
			Referral{ReferrerID: referrerID, RefereeID: uuid.Nil, ReferralCode: "ABC"},
			ErrInvalidInput,
		},
		{
			"empty code",
			Referral{ReferrerID: referrerID, RefereeID: refereeID, ReferralCode: ""},
			ErrInvalidInput,
		},
		{
			"self referral",
			Referral{ReferrerID: referrerID, RefereeID: referrerID, ReferralCode: "ABC"},
			ErrSelfReferral,
		},
		{
			"invalid status",
			Referral{ReferrerID: referrerID, RefereeID: refereeID, ReferralCode: "ABC", Status: "bad"},
			ErrInvalidInput,
		},
		{
			"negative bonus",
			Referral{ReferrerID: referrerID, RefereeID: refereeID, ReferralCode: "ABC", BonusAmount: -1},
			ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.referral.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestReferralBalance_Validate(t *testing.T) {
	valid := &ReferralBalance{
		UserID:      uuid.New(),
		Balance:     5000,
		TotalEarned: 10000,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid balance returned error: %v", err)
	}

	tests := []struct {
		name    string
		balance ReferralBalance
	}{
		{"nil user", ReferralBalance{UserID: uuid.Nil, Balance: 0, TotalEarned: 0}},
		{"negative balance", ReferralBalance{UserID: uuid.New(), Balance: -1, TotalEarned: 0}},
		{"negative total earned", ReferralBalance{UserID: uuid.New(), Balance: 0, TotalEarned: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.balance.Validate(); err == nil {
				t.Error("expected error for invalid balance")
			}
		})
	}
}
