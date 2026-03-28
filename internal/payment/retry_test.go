package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingProvider fails N times then succeeds.
type failingProvider struct {
	inner       PaymentProvider
	failsLeft   int
	retryable   bool
	createCalls int
	statusCalls int
	refundCalls int
}

func (f *failingProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error) {
	f.createCalls++
	if f.failsLeft > 0 {
		f.failsLeft--
		return nil, f.makeError()
	}
	return f.inner.CreatePayment(ctx, req)
}

func (f *failingProvider) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	f.statusCalls++
	if f.failsLeft > 0 {
		f.failsLeft--
		return "", f.makeError()
	}
	return f.inner.GetPaymentStatus(ctx, externalID)
}

func (f *failingProvider) CreateRefund(ctx context.Context, externalID string, amount int64) error {
	f.refundCalls++
	if f.failsLeft > 0 {
		f.failsLeft--
		return f.makeError()
	}
	return f.inner.CreateRefund(ctx, externalID, amount)
}

func (f *failingProvider) CapturePayment(ctx context.Context, externalID string, amount int64) error {
	if f.failsLeft > 0 {
		f.failsLeft--
		return f.makeError()
	}
	return f.inner.CapturePayment(ctx, externalID, amount)
}

func (f *failingProvider) CancelPayment(ctx context.Context, externalID string) error {
	if f.failsLeft > 0 {
		f.failsLeft--
		return f.makeError()
	}
	return f.inner.CancelPayment(ctx, externalID)
}

func (f *failingProvider) makeError() error {
	if f.retryable {
		return errors.New("connection refused")
	}
	return errors.New("insufficient_funds")
}

func TestRetryingProvider_CreatePayment_RetriesOnRetryableError(t *testing.T) {
	mock := NewMockProvider()
	fp := &failingProvider{inner: mock, failsLeft: 2, retryable: true}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	result, err := rp.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.ExternalID)
	assert.Equal(t, 3, fp.createCalls) // 2 failures + 1 success
}

func TestRetryingProvider_CreatePayment_NoRetryOnPermanentError(t *testing.T) {
	mock := NewMockProvider()
	fp := &failingProvider{inner: mock, failsLeft: 5, retryable: false}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	_, err := rp.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})

	require.Error(t, err)
	assert.Equal(t, 1, fp.createCalls) // stops after first permanent error
}

func TestRetryingProvider_CreatePayment_ExhaustsRetries(t *testing.T) {
	mock := NewMockProvider()
	fp := &failingProvider{inner: mock, failsLeft: 5, retryable: true}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	_, err := rp.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})

	require.Error(t, err)
	assert.Equal(t, 3, fp.createCalls) // all 3 attempts used
}

func TestRetryingProvider_GetPaymentStatus_Retries(t *testing.T) {
	mock := NewMockProvider()
	result, _ := mock.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})

	fp := &failingProvider{inner: mock, failsLeft: 1, retryable: true}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	status, err := rp.GetPaymentStatus(context.Background(), result.ExternalID)
	require.NoError(t, err)
	assert.Equal(t, "pending", status)
	assert.Equal(t, 2, fp.statusCalls) // 1 failure + 1 success
}

func TestRetryingProvider_CreateRefund_Retries(t *testing.T) {
	mock := NewMockProvider()
	result, _ := mock.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})
	mock.SetPaymentStatus(result.ExternalID, "succeeded")

	fp := &failingProvider{inner: mock, failsLeft: 1, retryable: true}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	err := rp.CreateRefund(context.Background(), result.ExternalID, 5000)
	require.NoError(t, err)
	assert.Equal(t, 2, fp.refundCalls)
}

func TestRetryingProvider_ContextCancellation(t *testing.T) {
	mock := NewMockProvider()
	fp := &failingProvider{inner: mock, failsLeft: 5, retryable: true}
	rp := NewRetryingProvider(fp, RetryConfig{MaxAttempts: 3, BaseDelay: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := rp.CreatePayment(ctx, CreatePaymentRequest{
		Amount: 10000, Currency: "RUB", Method: "card", Capture: true,
	})

	require.Error(t, err)
}

func TestRetryingProvider_ExponentialBackoff(t *testing.T) {
	rp := &RetryingProvider{config: RetryConfig{BaseDelay: 2 * time.Second}}
	assert.Equal(t, 2*time.Second, rp.delay(0))
	assert.Equal(t, 4*time.Second, rp.delay(1))
	assert.Equal(t, 8*time.Second, rp.delay(2))
}

// Verify RetryingProvider implements PaymentProvider
var _ PaymentProvider = (*RetryingProvider)(nil)
