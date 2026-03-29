package domain

import (
	"time"

	"github.com/google/uuid"
)

// WalletCurrency - валюта кошелька
type WalletCurrency string

const (
	WalletCurrencyRUB WalletCurrency = "RUB"
	WalletCurrencyBYN WalletCurrency = "BYN"
)

func (c WalletCurrency) IsValid() bool {
	switch c {
	case WalletCurrencyRUB, WalletCurrencyBYN:
		return true
	}
	return false
}

// WalletStatus - статус кошелька
type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "active"
	WalletStatusFrozen   WalletStatus = "frozen"
	WalletStatusArchived WalletStatus = "archived"
)

func (s WalletStatus) IsValid() bool {
	switch s {
	case WalletStatusActive, WalletStatusFrozen, WalletStatusArchived:
		return true
	}
	return false
}

// WalletTransactionType - тип транзакции
type WalletTransactionType string

const (
	WalletTxTopUp         WalletTransactionType = "topup"
	WalletTxSpend         WalletTransactionType = "spend"
	WalletTxRefund        WalletTransactionType = "refund"
	WalletTxBonus         WalletTransactionType = "bonus"
	WalletTxBonusExpiry   WalletTransactionType = "bonus_expiry"
	WalletTxHoldCapture   WalletTransactionType = "hold_capture"
	WalletTxHoldRelease   WalletTransactionType = "hold_release"
	WalletTxPayout        WalletTransactionType = "payout"
	WalletTxWelcomeBonus  WalletTransactionType = "welcome_bonus"
	WalletTxReferralBonus WalletTransactionType = "referral_bonus"
	WalletTxAdminCredit   WalletTransactionType = "admin_credit"
	WalletTxAdminDebit    WalletTransactionType = "admin_debit"
	WalletTxCashback      WalletTransactionType = "cashback"
)

func (t WalletTransactionType) IsValid() bool {
	switch t {
	case WalletTxTopUp, WalletTxSpend, WalletTxRefund, WalletTxBonus,
		WalletTxBonusExpiry, WalletTxHoldCapture, WalletTxHoldRelease,
		WalletTxPayout, WalletTxWelcomeBonus, WalletTxReferralBonus,
		WalletTxAdminCredit, WalletTxAdminDebit, WalletTxCashback:
		return true
	}
	return false
}

// WalletTransactionStatus - статус транзакции
type WalletTransactionStatus string

const (
	WalletTxStatusPending   WalletTransactionStatus = "pending"
	WalletTxStatusCompleted WalletTransactionStatus = "completed"
	WalletTxStatusFailed    WalletTransactionStatus = "failed"
	WalletTxStatusCancelled WalletTransactionStatus = "cancelled"
)

func (s WalletTransactionStatus) IsValid() bool {
	switch s {
	case WalletTxStatusPending, WalletTxStatusCompleted, WalletTxStatusFailed, WalletTxStatusCancelled:
		return true
	}
	return false
}

// WalletHoldStatus - статус холда
type WalletHoldStatus string

const (
	WalletHoldStatusActive   WalletHoldStatus = "active"
	WalletHoldStatusCaptured WalletHoldStatus = "captured"
	WalletHoldStatusReleased WalletHoldStatus = "released"
	WalletHoldStatusExpired  WalletHoldStatus = "expired"
)

func (s WalletHoldStatus) IsValid() bool {
	switch s {
	case WalletHoldStatusActive, WalletHoldStatusCaptured, WalletHoldStatusReleased, WalletHoldStatusExpired:
		return true
	}
	return false
}

// Wallet - кошелёк пользователя
type Wallet struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Balance    int64 // в копейках
	HeldAmount int64 // заморожено в холдах, в копейках
	Currency   WalletCurrency
	Status     WalletStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (w *Wallet) Validate() error {
	if w.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !w.Currency.IsValid() {
		return ErrInvalidInput
	}
	if w.Balance < 0 {
		return ErrInvalidInput
	}
	if w.HeldAmount < 0 {
		return ErrInvalidInput
	}
	if w.Status != "" && !w.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// AvailableBalance возвращает доступный баланс (без учёта холдов)
func (w *Wallet) AvailableBalance() int64 {
	return w.Balance - w.HeldAmount
}

// IsFrozen проверяет, заморожен ли кошелёк
func (w *Wallet) IsFrozen() bool {
	return w.Status == WalletStatusFrozen
}

// WalletTransaction - транзакция кошелька
type WalletTransaction struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	Type          WalletTransactionType
	Amount        int64 // в копейках, всегда положительное
	BalanceAfter  int64 // баланс после транзакции
	Status        WalletTransactionStatus
	Description   string
	ReferenceType string     // "booking", "payment", "payout" и т.д.
	ReferenceID   *uuid.UUID // ID связанной сущности
	IsBonus       bool       // бонусная транзакция (для приоритетного списания)
	ExpiresAt     *time.Time // срок действия бонуса
	CreatedAt     time.Time
}

func (t *WalletTransaction) Validate() error {
	if t.WalletID == uuid.Nil {
		return ErrInvalidInput
	}
	if !t.Type.IsValid() {
		return ErrInvalidInput
	}
	if t.Amount <= 0 {
		return ErrInvalidInput
	}
	if t.Status != "" && !t.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// WalletHold - холд (заморозка) средств
type WalletHold struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	Amount        int64 // в копейках
	Status        WalletHoldStatus
	Description   string
	ReferenceType string // "booking" и т.д.
	ReferenceID   *uuid.UUID
	ExpiresAt     time.Time // когда холд автоматически снимается
	CapturedAt    *time.Time
	ReleasedAt    *time.Time
	CreatedAt     time.Time
}

func (h *WalletHold) Validate() error {
	if h.WalletID == uuid.Nil {
		return ErrInvalidInput
	}
	if h.Amount <= 0 {
		return ErrInvalidInput
	}
	if h.ExpiresAt.IsZero() {
		return ErrInvalidInput
	}
	if h.Status != "" && !h.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// IsExpired проверяет, истёк ли холд
func (h *WalletHold) IsExpired() bool {
	return time.Now().After(h.ExpiresAt)
}

// WalletTransactionFilter - фильтр для пагинации транзакций
type WalletTransactionFilter struct {
	WalletID *uuid.UUID
	Type     *WalletTransactionType
	IsBonus  *bool
	DateFrom *time.Time
	DateTo   *time.Time
	Page     int
	PageSize int
}

// Лимиты кошелька
const (
	WalletMaxBalanceRUB int64 = 10_000_000 // 100 000 RUB в копейках
	WalletMaxBalanceBYN int64 = 300_000    // 3 000 BYN в копейках
	WalletTopUpMinRUB   int64 = 50_000     // 500 RUB в копейках
	WalletTopUpMaxRUB   int64 = 3_000_000  // 30 000 RUB в копейках
	WalletTopUpMinBYN   int64 = 1_000      // 10 BYN в копейках
	WalletTopUpMaxBYN   int64 = 100_000    // 1 000 BYN в копейках
)

// MaxBalanceForCurrency возвращает максимальный баланс для валюты
func MaxBalanceForCurrency(currency WalletCurrency) int64 {
	switch currency {
	case WalletCurrencyBYN:
		return WalletMaxBalanceBYN
	default:
		return WalletMaxBalanceRUB
	}
}

// TopUpMinForCurrency возвращает минимальную сумму пополнения для валюты
func TopUpMinForCurrency(currency WalletCurrency) int64 {
	switch currency {
	case WalletCurrencyBYN:
		return WalletTopUpMinBYN
	default:
		return WalletTopUpMinRUB
	}
}

// TopUpMaxForCurrency возвращает максимальную сумму пополнения для валюты
func TopUpMaxForCurrency(currency WalletCurrency) int64 {
	switch currency {
	case WalletCurrencyBYN:
		return WalletTopUpMaxBYN
	default:
		return WalletTopUpMaxRUB
	}
}
