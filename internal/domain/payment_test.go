package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPaymentStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   PaymentStatus
		expected bool
	}{
		{"pending is valid", PaymentPending, true},
		{"processing is valid", PaymentProcessing, true},
		{"succeeded is valid", PaymentSucceeded, true},
		{"failed is valid", PaymentFailed, true},
		{"refunded is valid", PaymentRefunded, true},
		{"partially_refunded is valid", PaymentPartiallyRefunded, true},
		{"invalid status", PaymentStatus("invalid"), false},
		{"empty status", PaymentStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestPaymentMethodIsValid(t *testing.T) {
	tests := []struct {
		name     string
		method   PaymentMethod
		expected bool
	}{
		{"card", PaymentMethodCard, true},
		{"sbp", PaymentMethodSBP, true},
		{"wallet", PaymentMethodWallet, true},
		{"combo", PaymentMethodCombo, true},
		{"mir", PaymentMethodMIR, true},
		{"belkart", PaymentMethodBelkart, true},
		{"erip", PaymentMethodERIP, true},
		{"apple_pay", PaymentMethodApplePay, true},
		{"google_pay", PaymentMethodGooglePay, true},
		{"invalid method", PaymentMethod("bitcoin"), false},
		{"empty method", PaymentMethod(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.IsValid())
		})
	}
}

func TestPaymentMethodIsTokenBased(t *testing.T) {
	tests := []struct {
		name     string
		method   PaymentMethod
		expected bool
	}{
		{"apple_pay is token-based", PaymentMethodApplePay, true},
		{"google_pay is token-based", PaymentMethodGooglePay, true},
		{"card is not token-based", PaymentMethodCard, false},
		{"sbp is not token-based", PaymentMethodSBP, false},
		{"wallet is not token-based", PaymentMethodWallet, false},
		{"combo is not token-based", PaymentMethodCombo, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.IsTokenBased())
		})
	}
}

func TestPaymentValidate(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	userID := uuid.New()

	validPayment := func() *Payment {
		return &Payment{
			ID:        uuid.New(),
			BookingID: bookingID,
			UserID:    userID,
			Amount:    500000,
			Currency:  "RUB",
			Status:    PaymentPending,
			Provider:  "yookassa",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	tests := []struct {
		name    string
		modify  func(p *Payment)
		wantErr error
	}{
		{
			name:    "valid payment",
			modify:  func(p *Payment) {},
			wantErr: nil,
		},
		{
			name:    "valid payment with refund",
			modify:  func(p *Payment) { p.RefundAmount = 100000; p.Status = PaymentPartiallyRefunded },
			wantErr: nil,
		},
		{
			name:    "missing booking id",
			modify:  func(p *Payment) { p.BookingID = uuid.Nil },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "missing user id",
			modify:  func(p *Payment) { p.UserID = uuid.Nil },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "zero amount",
			modify:  func(p *Payment) { p.Amount = 0 },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "negative amount",
			modify:  func(p *Payment) { p.Amount = -100 },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty currency",
			modify:  func(p *Payment) { p.Currency = "" },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "invalid status",
			modify:  func(p *Payment) { p.Status = PaymentStatus("invalid") },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty provider",
			modify:  func(p *Payment) { p.Provider = "" },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "negative refund amount",
			modify:  func(p *Payment) { p.RefundAmount = -100 },
			wantErr: ErrInvalidInput,
		},
		{
			name:    "refund exceeds amount",
			modify:  func(p *Payment) { p.RefundAmount = 600000 },
			wantErr: ErrRefundExceedsAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validPayment()
			tt.modify(p)
			err := p.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
