package payment

import "context"

// CreatePayoutRequest contains parameters for creating a payout.
type CreatePayoutRequest struct {
	Amount      int64             // in kopecks
	Currency    string            // "RUB", "BYN"
	Method      string            // "sbp", "bank_transfer"
	Phone       string            // for SBP payouts — recipient phone
	BankDetails map[string]string // for bank transfer — BIK, account, etc.
	Description string
	Metadata    map[string]string
}

// PayoutResult contains the result of creating a payout.
type PayoutResult struct {
	ExternalID string
	Status     string // "pending", "succeeded", "canceled"
}

// PayoutProvider abstracts the payout gateway (e.g. YooKassa Payouts API).
type PayoutProvider interface {
	CreatePayout(ctx context.Context, req CreatePayoutRequest) (*PayoutResult, error)
	GetPayoutStatus(ctx context.Context, externalID string) (string, error)
}
