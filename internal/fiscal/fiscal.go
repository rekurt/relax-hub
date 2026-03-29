package fiscal

import "context"

// ReceiptType identifies the type of fiscal receipt.
type ReceiptType string

const (
	ReceiptAdvance       ReceiptType = "advance"
	ReceiptAdvanceCredit ReceiptType = "advance_credit"
	ReceiptFullPayment   ReceiptType = "full_payment"
	ReceiptRefund        ReceiptType = "refund"
)

// ReceiptRequest contains data needed to create a fiscal receipt.
type ReceiptRequest struct {
	Type   ReceiptType
	Amount int64 // kopecks
	Email  string
	Phone  string
	Items  []ReceiptItem
}

// ReceiptItem represents a single line item in a fiscal receipt.
type ReceiptItem struct {
	Name     string
	Quantity int
	Price    int64  // kopecks
	VAT      string // "none", "vat0", "vat10", "vat20"
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
