package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

// mockDeviceTokenService implements service.DeviceTokenService for testing.
type mockDeviceTokenService struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*domain.DeviceToken
}

func newMockDeviceTokenService() *mockDeviceTokenService {
	return &mockDeviceTokenService{
		tokens: make(map[uuid.UUID]*domain.DeviceToken),
	}
}

func (m *mockDeviceTokenService) Register(_ context.Context, token *domain.DeviceToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token.Token == "" {
		return domain.ErrInvalidInput
	}
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}
	cp := *token
	m.tokens[token.ID] = &cp
	return nil
}

func (m *mockDeviceTokenService) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tokens[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.tokens, id)
	return nil
}

// ListByUser is a test helper, not part of the service interface.
func (m *mockDeviceTokenService) ListByUser(userID uuid.UUID) []domain.DeviceToken {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []domain.DeviceToken
	for _, dt := range m.tokens {
		if dt.UserID == userID {
			cp := *dt
			result = append(result, cp)
		}
	}
	return result
}

func TestDeviceTokenHandler_Register(t *testing.T) {
	svc := newMockDeviceTokenService()
	h := handler.NewDeviceTokenHandler(svc)

	userID := uuid.New()
	body := map[string]string{
		"token":    `{"endpoint":"https://push.example.com","keys":{"p256dh":"test","auth":"test"}}`,
		"platform": "web",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/device-tokens", bytes.NewReader(bodyBytes))
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// Verify token was stored
	tokens := svc.ListByUser(userID)
	if len(tokens) != 1 {
		t.Errorf("tokens count = %d, want 1", len(tokens))
	}
}

func TestDeviceTokenHandler_Register_EmptyToken(t *testing.T) {
	svc := newMockDeviceTokenService()
	h := handler.NewDeviceTokenHandler(svc)

	userID := uuid.New()
	body := map[string]string{
		"token":    "",
		"platform": "web",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/device-tokens", bytes.NewReader(bodyBytes))
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDeviceTokenHandler_Delete(t *testing.T) {
	svc := newMockDeviceTokenService()
	h := handler.NewDeviceTokenHandler(svc)

	userID := uuid.New()
	body := map[string]string{
		"token":    `{"endpoint":"https://push.example.com","keys":{"p256dh":"test","auth":"test"}}`,
		"platform": "web",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/device-tokens", bytes.NewReader(bodyBytes))
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	w := httptest.NewRecorder()
	h.Register(w, req)

	// Get the token ID from the response
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	tokenID := resp.Data.ID

	// Delete the token
	r := chi.NewRouter()
	r.Delete("/api/v1/device-tokens/{id}", h.Delete)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/device-tokens/"+tokenID, nil)
	req2 = req2.WithContext(createTestContext(userID, domain.RoleClient))
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w2.Code, http.StatusOK, w2.Body.String())
	}

	// Verify token was deleted
	tokens := svc.ListByUser(userID)
	if len(tokens) != 0 {
		t.Errorf("tokens count = %d, want 0", len(tokens))
	}
}

func TestDeviceTokenHandler_Delete_InvalidID(t *testing.T) {
	svc := newMockDeviceTokenService()
	h := handler.NewDeviceTokenHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/device-tokens/{id}", h.Delete)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/device-tokens/not-a-uuid", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
