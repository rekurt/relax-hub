package fiscal

import (
	"context"

	"github.com/rekurt/relax-hub/internal/domain"
)

// ReceiptType identifies the type of fiscal receipt.
type ReceiptType string

const (
	ReceiptAdvance       ReceiptType = "advance"
	ReceiptAdvanceCredit ReceiptType = "advance_credit"
	ReceiptFullPayment   ReceiptType = "full_payment"
	ReceiptRefund        ReceiptType = "refund"
)

// TaxSystem represents the taxation system used in fiscal receipts (ATOL sno parameter).
type TaxSystem string

const (
	TaxSystemOSN     TaxSystem = "osn"        // Общая система налогообложения
	TaxSystemUSN     TaxSystem = "usn_income" // Упрощённая (доходы)
	TaxSystemPatent  TaxSystem = "patent"     // Патент
	TaxSystemNPD     TaxSystem = "npd"        // Налог на профессиональный доход (самозанятые)
	TaxSystemDefault TaxSystem = "osn"        // По умолчанию
)

// ReceiptRequest contains data needed to create a fiscal receipt.
type ReceiptRequest struct {
	Type      ReceiptType
	Amount    int64 // kopecks
	Email     string
	Phone     string
	Items     []ReceiptItem
	TaxSystem TaxSystem // система налогообложения продавца
}

// ReceiptItem represents a single line item in a fiscal receipt.
type ReceiptItem struct {
	Name     string
	Quantity int
	Price    int64  // kopecks
	VAT      string // "none", "vat0", "vat10", "vat20"
}

// TaxInfoForEntityType returns the appropriate VAT rate and tax system
// for a given KYC entity type. This adapts fiscal receipts to the
// seller's tax status as required by Russian tax law.
func TaxInfoForEntityType(entityType domain.KYCEntityType) (vat string, taxSystem TaxSystem) {
	switch entityType {
	case domain.KYCEntityLegalEntity:
		// Юрлица: ОСНО, НДС 20%
		return "vat20", TaxSystemOSN
	case domain.KYCEntitySoleProprietor:
		// ИП: ОСНО, НДС 20%
		return "vat20", TaxSystemOSN
	case domain.KYCEntitySelfEmployed:
		// Самозанятые: НПД, без НДС
		return "none", TaxSystemNPD
	case domain.KYCEntityIndividual:
		// Физлица: без НДС, ОСНО по умолчанию
		return "none", TaxSystemOSN
	default:
		return "none", TaxSystemDefault
	}
}

// Receipt represents a created fiscal receipt.
type Receipt struct {
	ID         string
	ExternalID string
	Status     string
}

// FiscalProvider abstracts the fiscal receipt creation (e.g. ATOL Online).
type FiscalProvider interface {
	CreateReceipt(ctx context.Context, req ReceiptRequest) (*Receipt, error)
}
