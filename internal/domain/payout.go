package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PayoutStatus - статус выплаты
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "pending"
	PayoutStatusProcessing PayoutStatus = "processing"
	PayoutStatusCompleted  PayoutStatus = "completed"
	PayoutStatusFailed     PayoutStatus = "failed"
)

func (s PayoutStatus) IsValid() bool {
	switch s {
	case PayoutStatusPending, PayoutStatusProcessing, PayoutStatusCompleted, PayoutStatusFailed:
		return true
	}
	return false
}

// Payout - запрос на вывод средств
type Payout struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Amount        int64 // в копейках
	Status        PayoutStatus
	BankDetails   json.RawMessage // JSONB с реквизитами
	RequestedAt   time.Time
	ProcessedAt   *time.Time
	FailureReason string
	CreatedAt     time.Time
}

func (p *Payout) Validate() error {
	if p.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.Amount <= 0 {
		return ErrInvalidInput
	}
	if p.Status != "" && !p.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// AutoPayoutSettings - настройки автовыплаты
type AutoPayoutSettings struct {
	UserID    uuid.UUID
	Threshold int64 // порог в копейках, 0 = отключено
	UpdatedAt time.Time
}

// PayoutFilter - фильтр для списка выплат
type PayoutFilter struct {
	UserID   *uuid.UUID
	Status   *PayoutStatus
	DateFrom *time.Time
	DateTo   *time.Time
	Page     int
	PageSize int
}

// Лимиты выплат
const (
	PayoutMinAmountRUB      int64 = 50_000      // 500 RUB в копейках
	PayoutDailyLimitRUB     int64 = 50_000_000   // 500,000 RUB в копейках
	PayoutMonthlyLimitRUB   int64 = 300_000_000  // 3,000,000 RUB в копейках
)
