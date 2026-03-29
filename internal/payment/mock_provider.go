package payment

import (
	"context"
	"fmt"
	"sync"
)

// MockProvider implements PaymentProvider for testing.
type MockProvider struct {
	mu            sync.Mutex
	payments      map[string]*mockPayment
	refunds       map[string]int64
	captures      map[string]int64
	cancellations map[string]bool
	counter       int
	FailCapture   bool // when true, CapturePayment returns an error
	FailCancel    bool // when true, CancelPayment returns an error
}

type mockPayment struct {
	ExternalID string
	Amount     int64
	Currency   string
	Status     string
	Method     string
	Capture    bool
}

// NewMockProvider creates a new mock payment provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		payments:      make(map[string]*mockPayment),
		refunds:       make(map[string]int64),
		captures:      make(map[string]int64),
		cancellations: make(map[string]bool),
	}
}

func (m *MockProvider) CreatePayment(_ context.Context, req CreatePaymentRequest) (*PaymentResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	externalID := fmt.Sprintf("mock-pay-%d", m.counter)

	m.payments[externalID] = &mockPayment{
		ExternalID: externalID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     "pending",
		Method:     req.Method,
		Capture:    req.Capture,
	}

	return &PaymentResult{
		ExternalID:      externalID,
		ConfirmationURL: req.ReturnURL + "?payment_id=" + externalID,
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

func (m *MockProvider) CapturePayment(_ context.Context, externalID string, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.FailCapture {
		return fmt.Errorf("capture failed (mock)")
	}

	p, ok := m.payments[externalID]
	if !ok {
		return fmt.Errorf("payment %s not found", externalID)
	}

	if p.Status != "waiting_for_capture" && p.Status != "pending" {
		return fmt.Errorf("payment %s cannot be captured, current status: %s", externalID, p.Status)
	}

	p.Status = "succeeded"
	m.captures[externalID] = amount
	return nil
}

func (m *MockProvider) CancelPayment(_ context.Context, externalID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.FailCancel {
		return fmt.Errorf("cancel failed (mock)")
	}

	p, ok := m.payments[externalID]
	if !ok {
		return fmt.Errorf("payment %s not found", externalID)
	}

	p.Status = "canceled"
	m.cancellations[externalID] = true
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

// GetLastPaymentMethod returns the method used for the last created payment (for testing).
func (m *MockProvider) GetLastPaymentMethod() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	lastID := fmt.Sprintf("mock-pay-%d", m.counter)
	if p, ok := m.payments[lastID]; ok {
		return p.Method
	}
	return ""
}

// GetPaymentCapture returns the capture flag for a given payment (for testing).
func (m *MockProvider) GetPaymentCapture(externalID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.payments[externalID]; ok {
		return p.Capture
	}
	return false
}

// WasCaptured returns true if CapturePayment was called for the given external ID.
func (m *MockProvider) WasCaptured(externalID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.captures[externalID]
	return ok
}

// WasCancelled returns true if CancelPayment was called for the given external ID.
func (m *MockProvider) WasCancelled(externalID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cancellations[externalID]
}
