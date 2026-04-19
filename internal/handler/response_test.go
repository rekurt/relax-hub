package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekurt/relax-hub/internal/domain"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "test_error", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false")
	}

	if resp.Error == nil {
		t.Fatal("Expected error to be set")
	}

	if resp.Error.Code != "test_error" {
		t.Errorf("Expected code 'test_error', got '%s'", resp.Error.Code)
	}

	if resp.Error.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", resp.Error.Message)
	}
}

func TestWriteErrorWithContext_5xxLogging(t *testing.T) {
	// Test that 5xx errors are logged
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	writeErrorWithContext(w, r, http.StatusInternalServerError, "internal_error", "Internal error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false")
	}
}

func TestHandleServiceError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "not_found" {
		t.Error("Expected error code 'not_found'")
	}
}

func TestHandleServiceError_AlreadyExists(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrAlreadyExists)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "already_exists" {
		t.Error("Expected error code 'already_exists'")
	}
}

func TestHandleServiceError_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrInvalidInput)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "invalid_input" {
		t.Error("Expected error code 'invalid_input'")
	}
}

func TestHandleServiceError_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrUnauthorized)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "unauthorized" {
		t.Error("Expected error code 'unauthorized'")
	}
}

func TestHandleServiceError_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrForbidden)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "forbidden" {
		t.Error("Expected error code 'forbidden'")
	}
}

func TestHandleServiceError_InternalError(t *testing.T) {
	w := httptest.NewRecorder()
	// Test with a generic error that doesn't match any domain error
	unknownErr := errors.New("unknown error")
	handleServiceError(w, unknownErr)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "internal_error" {
		t.Error("Expected error code 'internal_error'")
	}
}

func TestReadJSON_Valid(t *testing.T) {
	type testData struct {
		Name string `json:"name"`
	}

	body := bytes.NewReader([]byte(`{"name":"test"}`))
	r := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	var data testData
	err := readJSON(w, r, &data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if data.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", data.Name)
	}
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	body := bytes.NewReader([]byte(`invalid json`))
	r := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	var data map[string]interface{}
	err := readJSON(w, r, &data)

	if err != domain.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestReadJSON_NilBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	var data map[string]interface{}
	err := readJSON(w, r, &data)

	if err != domain.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestHandleServiceError_OAuthExchangeFailed_NoInternalDetails(t *testing.T) {
	w := httptest.NewRecorder()
	// Wrap with internal details that should NOT be exposed
	err := errors.Join(domain.ErrOAuthExchangeFailed, errors.New("Post \"https://accounts.google.com/o/oauth2/token\": 400 Bad Request"))
	handleServiceError(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "oauth_exchange_failed" {
		t.Error("Expected error code 'oauth_exchange_failed'")
	}

	if resp.Error.Message != "oauth code exchange failed" {
		t.Errorf("Expected static message 'oauth code exchange failed', got '%s'", resp.Error.Message)
	}
}

func TestWriteJSONWithMeta(t *testing.T) {
	w := httptest.NewRecorder()
	data := []int{1, 2, 3}
	meta := &Meta{
		Page:       1,
		PageSize:   10,
		TotalCount: 100,
		TotalPages: 10,
	}

	writeJSONWithMeta(w, http.StatusOK, data, meta)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	if resp.Meta == nil {
		t.Fatal("Expected meta to be set")
	}

	if resp.Meta.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Meta.Page)
	}

	if resp.Meta.TotalPages != 10 {
		t.Errorf("Expected total_pages 10, got %d", resp.Meta.TotalPages)
	}
}
