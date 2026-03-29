package payment

import (
	"context"
	"errors"
)

// ErrPayoutsNotConfigured is returned when payout credentials are not set up.
var ErrPayoutsNotConfigured = errors.New("payout provider is not configured")

// NoopPayoutProvider is a PayoutProvider that returns errors.
// Used when payout credentials are not configured.
type NoopPayoutProvider struct{}

func NewNoopPayoutProvider() *NoopPayoutProvider {
	return &NoopPayoutProvider{}
}

func (p *NoopPayoutProvider) CreatePayout(_ context.Context, _ CreatePayoutRequest) (*PayoutResult, error) {
	return nil, ErrPayoutsNotConfigured
}

func (p *NoopPayoutProvider) GetPayoutStatus(_ context.Context, _ string) (string, error) {
	return "", ErrPayoutsNotConfigured
}
