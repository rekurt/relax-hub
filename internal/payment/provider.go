package payment

import "context"

// CreatePaymentRequest contains parameters for creating a payment in the external provider.
type CreatePaymentRequest struct {
	Amount       int64
	Currency     string
	Description  string
	ReturnURL    string
	Metadata     map[string]string
	Method       string // "card", "sbp", "apple_pay", "google_pay"
	Capture      bool   // true for instant charge, false for authorization hold
	PaymentToken string // client-side token for Apple Pay / Google Pay
}

// PaymentResult contains the result of creating a payment in the external provider.
type PaymentResult struct {
	ExternalID      string
	ConfirmationURL string
}

// PaymentProvider abstracts the payment gateway (e.g. YooKassa).
type PaymentProvider interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error)
	GetPaymentStatus(ctx context.Context, externalID string) (status string, err error)
	CreateRefund(ctx context.Context, externalID string, amount int64) error
	CapturePayment(ctx context.Context, externalID string, amount int64) error
	CancelPayment(ctx context.Context, externalID string) error
}
