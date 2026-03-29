package payment

import (
	"context"
)

// NoopPayoutProvider is a PayoutProvider that does nothing.
// Used when payout credentials are not configured.
type NoopPayoutProvider struct{}

func NewNoopPayoutProvider() *NoopPayoutProvider {
	return &NoopPayoutProvider{}
}

func (p *NoopPayoutProvider) CreatePayout(_ context.Context, _ CreatePayoutRequest) (*PayoutResult, error) {
	return &PayoutResult{
		ExternalID: "",
		Status:     "pending",
	}, nil
}

func (p *NoopPayoutProvider) GetPayoutStatus(_ context.Context, _ string) (string, error) {
	return "pending", nil
}

// IsNoop returns true — used by services to check if real payouts are available.
func (p *NoopPayoutProvider) IsNoop() bool {
	return true
}
