package payment

import "context"

// PaymentResult contains the result of creating a payment in the external provider.
type PaymentResult struct {
	ExternalID      string
	ConfirmationURL string
}

// PaymentProvider abstracts the payment gateway (e.g. YooKassa).
type PaymentProvider interface {
	CreatePayment(ctx context.Context, amount int64, currency string, description string, returnURL string, metadata map[string]string) (*PaymentResult, error)
	GetPaymentStatus(ctx context.Context, externalID string) (status string, err error)
	CreateRefund(ctx context.Context, externalID string, amount int64) error
}
