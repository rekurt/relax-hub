package payment

import (
	"context"
	"fmt"
)

// ProviderFactory selects the appropriate PaymentProvider based on region.
// It implements PaymentProvider itself, defaulting to the RU provider for backward compatibility.
type ProviderFactory struct {
	providers map[string]PaymentProvider
	fallback  PaymentProvider
}

// NewProviderFactory creates a factory with region-keyed providers.
// The "RU" provider is used as the default fallback.
func NewProviderFactory(providers map[string]PaymentProvider) *ProviderFactory {
	f := &ProviderFactory{
		providers: providers,
	}
	if p, ok := providers["RU"]; ok {
		f.fallback = p
	}
	return f
}

// ProviderForRegion returns the PaymentProvider for the given region.
// Returns an error if no provider is configured for the region.
func (f *ProviderFactory) ProviderForRegion(region string) (PaymentProvider, error) {
	if p, ok := f.providers[region]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("no payment provider configured for region %q", region)
}

// ProviderName returns the provider name for a given region.
func (f *ProviderFactory) ProviderName(region string) string {
	switch region {
	case "BY":
		return "bepaid"
	default:
		return "yookassa"
	}
}

// Below: PaymentProvider interface implementation using the fallback (RU) provider.
// This ensures backward compatibility with existing code that injects a single PaymentProvider.

func (f *ProviderFactory) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error) {
	if f.fallback == nil {
		return nil, fmt.Errorf("no default payment provider configured")
	}
	return f.fallback.CreatePayment(ctx, req)
}

func (f *ProviderFactory) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	if f.fallback == nil {
		return "", fmt.Errorf("no default payment provider configured")
	}
	return f.fallback.GetPaymentStatus(ctx, externalID)
}

func (f *ProviderFactory) CreateRefund(ctx context.Context, externalID string, amount int64) error {
	if f.fallback == nil {
		return fmt.Errorf("no default payment provider configured")
	}
	return f.fallback.CreateRefund(ctx, externalID, amount)
}

func (f *ProviderFactory) CapturePayment(ctx context.Context, externalID string, amount int64) error {
	if f.fallback == nil {
		return fmt.Errorf("no default payment provider configured")
	}
	return f.fallback.CapturePayment(ctx, externalID, amount)
}

func (f *ProviderFactory) CancelPayment(ctx context.Context, externalID string) error {
	if f.fallback == nil {
		return fmt.Errorf("no default payment provider configured")
	}
	return f.fallback.CancelPayment(ctx, externalID)
}
