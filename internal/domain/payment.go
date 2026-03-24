package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodSBP    PaymentMethod = "sbp"
	PaymentMethodWallet PaymentMethod = "wallet"
	PaymentMethodCombo  PaymentMethod = "combo"
)

func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCard, PaymentMethodSBP, PaymentMethodWallet, PaymentMethodCombo:
		return true
	}
	return false
}

type PaymentStatus string

const (
	PaymentPending            PaymentStatus = "pending"
	PaymentProcessing         PaymentStatus = "processing"
	PaymentSucceeded          PaymentStatus = "succeeded"
	PaymentFailed             PaymentStatus = "failed"
	PaymentRefunded           PaymentStatus = "refunded"
	PaymentPartiallyRefunded  PaymentStatus = "partially_refunded"
)

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentPending, PaymentProcessing, PaymentSucceeded, PaymentFailed, PaymentRefunded, PaymentPartiallyRefunded:
		return true
	}
	return false
}

type Payment struct {
	ID            uuid.UUID
	BookingID     uuid.UUID
	UserID        uuid.UUID
	Amount        int64             // в копейках
	Currency      string            // "RUB"
	Status        PaymentStatus
	Provider      string            // "yookassa"
	ExternalID    string            // ID транзакции в платежной системе
	PaymentMethod PaymentMethod     // "card", "sbp", "wallet", "combo"
	WalletAmount  int64             // копейки, оплачено из кошелька
	CardAmount    int64             // копейки, оплачено картой/СБП
	IsHold        bool              // true = authorization hold, not yet captured
	CapturedAt    *time.Time        // when the hold was captured
	RefundAmount  int64             // сумма возврата в копейках
	RefundedAt    *time.Time
	Metadata      map[string]string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (p *Payment) Validate() error {
	if p.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.Amount <= 0 {
		return ErrInvalidInput
	}
	if p.Currency == "" {
		return ErrInvalidInput
	}
	if !p.Status.IsValid() {
		return ErrInvalidInput
	}
	if p.Provider == "" {
		return ErrInvalidInput
	}
	if p.PaymentMethod != "" && !p.PaymentMethod.IsValid() {
		return ErrInvalidInput
	}
	if p.RefundAmount < 0 {
		return ErrInvalidInput
	}
	if p.RefundAmount > p.Amount {
		return ErrRefundExceedsAmount
	}
	return nil
}
