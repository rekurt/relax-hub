package payment

import (
	"context"
	"fmt"
	"sync"
)

// MockPayoutProvider implements PayoutProvider for testing.
type MockPayoutProvider struct {
	mu      sync.Mutex
	payouts map[string]*mockPayout
	counter int
	Fail    bool // when true, CreatePayout returns an error
}

type mockPayout struct {
	ExternalID string
	Amount     int64
	Method     string
	Phone      string
	Status     string
}

func NewMockPayoutProvider() *MockPayoutProvider {
	return &MockPayoutProvider{
		payouts: make(map[string]*mockPayout),
	}
}

func (m *MockPayoutProvider) CreatePayout(_ context.Context, req CreatePayoutRequest) (*PayoutResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Fail {
		return nil, fmt.Errorf("payout failed (mock)")
	}

	m.counter++
	externalID := fmt.Sprintf("mock-payout-%d", m.counter)

	m.payouts[externalID] = &mockPayout{
		ExternalID: externalID,
		Amount:     req.Amount,
		Method:     req.Method,
		Phone:      req.Phone,
		Status:     "succeeded",
	}

	return &PayoutResult{
		ExternalID: externalID,
		Status:     "succeeded",
	}, nil
}

func (m *MockPayoutProvider) GetPayoutStatus(_ context.Context, externalID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.payouts[externalID]
	if !ok {
		return "", fmt.Errorf("payout %s not found", externalID)
	}
	return p.Status, nil
}

// GetLastPayout returns the last created payout details for testing.
func (m *MockPayoutProvider) GetLastPayout() *mockPayout {
	m.mu.Lock()
	defer m.mu.Unlock()
	lastID := fmt.Sprintf("mock-payout-%d", m.counter)
	return m.payouts[lastID]
}
