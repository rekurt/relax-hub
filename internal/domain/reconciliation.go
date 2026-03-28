package domain

import (
	"time"

	"github.com/google/uuid"
)

// FloatSnapshotStatus - статус снимка
type FloatSnapshotStatus string

const (
	FloatSnapshotOK          FloatSnapshotStatus = "ok"
	FloatSnapshotDiscrepancy FloatSnapshotStatus = "discrepancy"
)

func (s FloatSnapshotStatus) IsValid() bool {
	switch s {
	case FloatSnapshotOK, FloatSnapshotDiscrepancy:
		return true
	}
	return false
}

// FloatSnapshot - ежедневный снимок баланса платформы
type FloatSnapshot struct {
	ID                  uuid.UUID
	ClientWalletsTotal  int64               // сумма балансов клиентских кошельков (копейки)
	OwnerWalletsTotal   int64               // сумма балансов кошельков владельцев (копейки)
	EscrowHeldTotal     int64               // сумма эскроу в статусе held (копейки)
	WalletHoldsTotal    int64               // сумма активных холдов кошельков (копейки)
	ExpectedTotal       int64               // ожидаемый итог (сумма всех 4 компонент)
	ActualTotal         int64               // фактический итог от провайдера (если доступен)
	Discrepancy         int64               // разница: actual - expected (0 = ok)
	Status              FloatSnapshotStatus // ok / discrepancy
	ClientWalletsCount  int                 // количество клиентских кошельков
	OwnerWalletsCount   int                 // количество кошельков владельцев
	EscrowCount         int                 // количество эскроу записей
	Notes               string              // дополнительные заметки
	SnapshotDate        time.Time           // дата снимка
	CreatedAt           time.Time
}

func (f *FloatSnapshot) Validate() error {
	if !f.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// ReconciliationStatus - статус сверки
type ReconciliationStatus string

const (
	ReconciliationPending  ReconciliationStatus = "pending"
	ReconciliationMatched  ReconciliationStatus = "matched"
	ReconciliationMismatch ReconciliationStatus = "mismatch"
	ReconciliationError    ReconciliationStatus = "error"
)

func (s ReconciliationStatus) IsValid() bool {
	switch s {
	case ReconciliationPending, ReconciliationMatched, ReconciliationMismatch, ReconciliationError:
		return true
	}
	return false
}

// ReconciliationReport - отчёт сверки с платёжным провайдером
type ReconciliationReport struct {
	ID                    uuid.UUID
	PeriodStart           time.Time            // начало периода сверки
	PeriodEnd             time.Time            // конец периода сверки
	InternalPaymentsSum   int64                // сумма платежей по нашим данным (копейки)
	InternalPaymentsCount int                  // количество платежей по нашим данным
	InternalRefundsSum    int64                // сумма возвратов по нашим данным (копейки)
	InternalRefundsCount  int                  // количество возвратов по нашим данным
	ProviderPaymentsSum   int64                // сумма платежей по данным провайдера (копейки)
	ProviderPaymentsCount int                  // количество платежей по данным провайдера
	ProviderRefundsSum    int64                // сумма возвратов по данным провайдера (копейки)
	ProviderRefundsCount  int                  // количество возвратов по данным провайдера
	PaymentDiscrepancy    int64                // расхождение по платежам
	RefundDiscrepancy     int64                // расхождение по возвратам
	Status                ReconciliationStatus // matched / mismatch / error
	MismatchDetails       string               // детали расхождений (JSON)
	ErrorMessage          string               // сообщение об ошибке
	CreatedAt             time.Time
}

func (r *ReconciliationReport) Validate() error {
	if r.PeriodStart.IsZero() || r.PeriodEnd.IsZero() {
		return ErrInvalidInput
	}
	if r.PeriodEnd.Before(r.PeriodStart) {
		return ErrInvalidInput
	}
	if !r.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
