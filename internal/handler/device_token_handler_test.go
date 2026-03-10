package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestDeviceTokenHandler_Register(t *testing.T) {
	repo := mock.NewDeviceTokenRepo()
	h := handler.NewDeviceTokenHandler(repo)

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
	tokens, err := repo.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser error: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("tokens count = %d, want 1", len(tokens))
	}
}

func TestDeviceTokenHandler_Register_EmptyToken(t *testing.T) {
	repo := mock.NewDeviceTokenRepo()
	h := handler.NewDeviceTokenHandler(repo)

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
	repo := mock.NewDeviceTokenRepo()
	h := handler.NewDeviceTokenHandler(repo)

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
	json.NewDecoder(w.Body).Decode(&resp)
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
	tokens, _ := repo.ListByUser(context.Background(), userID)
	if len(tokens) != 0 {
		t.Errorf("tokens count = %d, want 0", len(tokens))
	}
}

func TestDeviceTokenHandler_Delete_InvalidID(t *testing.T) {
	repo := mock.NewDeviceTokenRepo()
	h := handler.NewDeviceTokenHandler(repo)

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
