package domain

import (
	"time"

	"github.com/google/uuid"
)

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
	ID           uuid.UUID
	BookingID    uuid.UUID
	UserID       uuid.UUID
	Amount       int64             // в копейках
	Currency     string            // "RUB"
	Status       PaymentStatus
	Provider     string            // "yookassa"
	ExternalID   string            // ID транзакции в платежной системе
	RefundAmount int64             // сумма возврата в копейках
	RefundedAt   *time.Time
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	if p.RefundAmount < 0 {
		return ErrInvalidInput
	}
	if p.RefundAmount > p.Amount {
		return ErrRefundExceedsAmount
	}
	return nil
}
