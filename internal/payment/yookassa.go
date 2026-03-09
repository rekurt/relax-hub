package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
	yoorefund "github.com/rvinnie/yookassa-sdk-go/yookassa/refund"
)

// YooKassaProvider implements PaymentProvider using the YooKassa payment gateway.
type YooKassaProvider struct {
	paymentHandler *yookassa.PaymentHandler
	refundHandler  *yookassa.RefundHandler
}

// NewYooKassaProvider creates a new YooKassa provider with the given credentials.
func NewYooKassaProvider(shopID, secretKey string) *YooKassaProvider {
	client := yookassa.NewClient(shopID, secretKey)
	return &YooKassaProvider{
		paymentHandler: yookassa.NewPaymentHandler(client),
		refundHandler:  yookassa.NewRefundHandler(client),
	}
}

// CreatePayment creates a payment in YooKassa and returns the external ID and confirmation URL.
func (p *YooKassaProvider) CreatePayment(ctx context.Context, amount int64, currency string, description string, returnURL string, metadata map[string]string) (*PaymentResult, error) {
	payment := &yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    kopecksToString(amount),
			Currency: currency,
		},
		Capture:     true,
		Description: description,
		Confirmation: &yoopayment.Redirect{
			Type:      yoopayment.TypeRedirect,
			ReturnURL: returnURL,
		},
		Metadata: metadata,
	}

	handler := p.paymentHandler.WithIdempotencyKey(uuid.New().String())
	result, err := handler.CreatePayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("yookassa create payment: %w", err)
	}

	var confirmationURL string
	if redirect, ok := result.Confirmation.(*yoopayment.Redirect); ok {
		confirmationURL = redirect.ConfirmationURL
	}

	return &PaymentResult{
		ExternalID:      result.ID,
		ConfirmationURL: confirmationURL,
	}, nil
}

// GetPaymentStatus returns the current status of a payment by its external ID.
func (p *YooKassaProvider) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	result, err := p.paymentHandler.FindPayment(ctx, externalID)
	if err != nil {
		return "", fmt.Errorf("yookassa find payment: %w", err)
	}
	return string(result.Status), nil
}

// CreateRefund creates a refund for a payment in YooKassa.
func (p *YooKassaProvider) CreateRefund(ctx context.Context, externalID string, amount int64) error {
	refund := &yoorefund.Refund{
		PaymentId: externalID,
		Amount: &yoocommon.Amount{
			Value:    kopecksToString(amount),
			Currency: "RUB",
		},
	}

	handler := p.refundHandler.WithIdempotencyKey(uuid.New().String())
	_, err := handler.CreateRefund(ctx, refund)
	if err != nil {
		return fmt.Errorf("yookassa create refund: %w", err)
	}
	return nil
}

// kopecksToString converts an amount in kopecks to a string with 2 decimal places.
// Example: 10050 -> "100.50", 500 -> "5.00"
func kopecksToString(kopecks int64) string {
	return fmt.Sprintf("%d.%02d", kopecks/100, kopecks%100)
}
