package payment

import (
	"context"
	"time"
)

// RetryConfig defines retry parameters for payment operations.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

// DefaultRetryConfig returns the default retry configuration: 3 attempts with 2s base delay.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   2 * time.Second,
}

// RetryingProvider wraps a PaymentProvider with automatic retry for retryable errors.
type RetryingProvider struct {
	inner  PaymentProvider
	config RetryConfig
}

// NewRetryingProvider creates a RetryingProvider wrapping the given provider.
func NewRetryingProvider(inner PaymentProvider, cfg RetryConfig) *RetryingProvider {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultRetryConfig.MaxAttempts
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = DefaultRetryConfig.BaseDelay
	}
	return &RetryingProvider{inner: inner, config: cfg}
}

func (p *RetryingProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error) {
	var lastErr error
	for attempt := 0; attempt < p.config.MaxAttempts; attempt++ {
		result, err := p.inner.CreatePayment(ctx, req)
		if err == nil {
			return result, nil
		}
		info := ClassifyError(err)
		if !info.Retryable {
			return nil, err
		}
		lastErr = err
		if attempt < p.config.MaxAttempts-1 {
			if err := sleepWithContext(ctx, p.delay(attempt)); err != nil {
				return nil, lastErr
			}
		}
	}
	return nil, lastErr
}

func (p *RetryingProvider) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < p.config.MaxAttempts; attempt++ {
		status, err := p.inner.GetPaymentStatus(ctx, externalID)
		if err == nil {
			return status, nil
		}
		info := ClassifyError(err)
		if !info.Retryable {
			return "", err
		}
		lastErr = err
		if attempt < p.config.MaxAttempts-1 {
			if err := sleepWithContext(ctx, p.delay(attempt)); err != nil {
				return "", lastErr
			}
		}
	}
	return "", lastErr
}

func (p *RetryingProvider) CreateRefund(ctx context.Context, externalID string, amount int64) error {
	return p.retryOp(ctx, func() error {
		return p.inner.CreateRefund(ctx, externalID, amount)
	})
}

func (p *RetryingProvider) CapturePayment(ctx context.Context, externalID string, amount int64) error {
	return p.retryOp(ctx, func() error {
		return p.inner.CapturePayment(ctx, externalID, amount)
	})
}

func (p *RetryingProvider) CancelPayment(ctx context.Context, externalID string) error {
	return p.retryOp(ctx, func() error {
		return p.inner.CancelPayment(ctx, externalID)
	})
}

func (p *RetryingProvider) retryOp(ctx context.Context, op func() error) error {
	var lastErr error
	for attempt := 0; attempt < p.config.MaxAttempts; attempt++ {
		err := op()
		if err == nil {
			return nil
		}
		info := ClassifyError(err)
		if !info.Retryable {
			return err
		}
		lastErr = err
		if attempt < p.config.MaxAttempts-1 {
			if err := sleepWithContext(ctx, p.delay(attempt)); err != nil {
				return lastErr
			}
		}
	}
	return lastErr
}

// delay returns the delay for the given attempt: 2s, 4s, 8s (exponential backoff).
func (p *RetryingProvider) delay(attempt int) time.Duration {
	d := p.config.BaseDelay
	for i := 0; i < attempt; i++ {
		d *= 2
	}
	return d
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
