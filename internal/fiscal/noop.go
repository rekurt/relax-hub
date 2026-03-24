package fiscal

import "context"

// NoOpProvider is a FiscalProvider that does nothing.
// Used when fiscalization is not configured.
type NoOpProvider struct{}

// NewNoOpProvider creates a new no-op fiscal provider.
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{}
}

// CreateReceipt does nothing and returns nil.
func (p *NoOpProvider) CreateReceipt(_ context.Context, _ ReceiptRequest) (*Receipt, error) {
	return nil, nil
}
