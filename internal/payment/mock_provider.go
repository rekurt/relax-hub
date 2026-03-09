package payment

import (
	"context"
	"fmt"
	"sync"
)

// MockProvider implements PaymentProvider for testing.
type MockProvider struct {
	mu       sync.Mutex
	payments map[string]*mockPayment
	refunds  map[string]int64
	counter  int
}

type mockPayment struct {
	ExternalID string
	Amount     int64
	Currency   string
	Status     string
}

// NewMockProvider creates a new mock payment provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		payments: make(map[string]*mockPayment),
		refunds:  make(map[string]int64),
	}
}

func (m *MockProvider) CreatePayment(_ context.Context, amount int64, currency string, _ string, returnURL string, _ map[string]string) (*PaymentResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	externalID := fmt.Sprintf("mock-pay-%d", m.counter)

	m.payments[externalID] = &mockPayment{
		ExternalID: externalID,
		Amount:     amount,
		Currency:   currency,
		Status:     "pending",
	}

	return &PaymentResult{
		ExternalID:      externalID,
		ConfirmationURL: returnURL + "?payment_id=" + externalID,
	}, nil
}

func (m *MockProvider) GetPaymentStatus(_ context.Context, externalID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.payments[externalID]
	if !ok {
		return "", fmt.Errorf("payment %s not found", externalID)
	}
	return p.Status, nil
}

func (m *MockProvider) CreateRefund(_ context.Context, externalID string, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.payments[externalID]
	if !ok {
		return fmt.Errorf("payment %s not found", externalID)
	}

	if p.Status != "succeeded" {
		return fmt.Errorf("payment %s is not succeeded, current status: %s", externalID, p.Status)
	}

	m.refunds[externalID] += amount
	return nil
}

// SetPaymentStatus allows tests to change the status of a mock payment.
func (m *MockProvider) SetPaymentStatus(externalID, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.payments[externalID]; ok {
		p.Status = status
	}
}
