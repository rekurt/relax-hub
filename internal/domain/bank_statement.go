package domain

import (
	"time"

	"github.com/google/uuid"
)

// BankStatementEntryStatus - статус строки банковской выписки
type BankStatementEntryStatus string

const (
	BankEntryPending  BankStatementEntryStatus = "pending"  // ожидает сопоставления
	BankEntryMatched  BankStatementEntryStatus = "matched"  // автоматически сопоставлена
	BankEntryManual   BankStatementEntryStatus = "manual"   // сопоставлена вручную
	BankEntryIgnored  BankStatementEntryStatus = "ignored"  // пропущена (не относится к платформе)
)

func (s BankStatementEntryStatus) IsValid() bool {
	switch s {
	case BankEntryPending, BankEntryMatched, BankEntryManual, BankEntryIgnored:
		return true
	}
	return false
}

// BankStatementEntry - строка из банковской выписки
type BankStatementEntry struct {
	ID            uuid.UUID                `json:"id"`
	Date          time.Time                `json:"date"`            // дата операции
	Amount        int64                    `json:"amount"`          // сумма в копейках (положительная для поступлений, отрицательная для списаний)
	Description   string                   `json:"description"`     // назначение платежа
	Counterparty  string                   `json:"counterparty"`    // контрагент
	ReferenceNum  string                   `json:"reference_num"`   // номер платёжного поручения / референс
	MatchedTxID   *uuid.UUID               `json:"matched_tx_id"`   // ID сопоставленной транзакции (payment или wallet_transaction)
	MatchedTxType string                   `json:"matched_tx_type"` // тип сопоставленной транзакции: "payment" или "wallet_transaction"
	Status        BankStatementEntryStatus `json:"status"`
	UploadBatchID uuid.UUID                `json:"upload_batch_id"` // ID пакета загрузки
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

func (e *BankStatementEntry) Validate() error {
	if e.Date.IsZero() {
		return ErrInvalidInput
	}
	if e.Amount == 0 {
		return ErrInvalidInput
	}
	if e.Description == "" && e.Counterparty == "" {
		return ErrInvalidInput
	}
	if !e.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// BankStatementUpload - метаданные загрузки выписки
type BankStatementUpload struct {
	ID           uuid.UUID `json:"id"`
	FileName     string    `json:"file_name"`
	Format       string    `json:"format"` // "csv" или "1c"
	TotalRows    int       `json:"total_rows"`
	MatchedCount int       `json:"matched_count"`
	PendingCount int       `json:"pending_count"`
	IgnoredCount int       `json:"ignored_count"`
	UploadedBy   uuid.UUID `json:"uploaded_by"` // admin user ID
	CreatedAt    time.Time `json:"created_at"`
}

// BankStatementFilter - фильтры для списка строк выписки
type BankStatementFilter struct {
	Status        *BankStatementEntryStatus
	UploadBatchID *uuid.UUID
	DateFrom      *time.Time
	DateTo        *time.Time
	Page          int
	PageSize      int
}
