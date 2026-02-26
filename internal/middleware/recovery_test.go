package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaaldaev/bani/internal/logger"
)

func TestRecoveryMiddleware_DevMode(t *testing.T) {
	log := logger.New(logger.LevelError)
	handler := RecoveryMiddleware(true, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	success, ok := resp["success"].(bool)
	if !ok || success {
		t.Error("Expected success=false")
	}
}

func TestRecoveryMiddleware_ProdMode(t *testing.T) {
	log := logger.New(logger.LevelError)
	handler := RecoveryMiddleware(false, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	success, ok := resp["success"].(bool)
	if !ok || success {
		t.Error("Expected success=false")
	}
}

func TestRecoveryMiddleware_RequestID(t *testing.T) {
	log := logger.New(logger.LevelError)
	handler := RecoveryMiddleware(true, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// The handler should recover and return a 500 error regardless of request ID
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	log := logger.New(logger.LevelError)
	handler := RecoveryMiddleware(true, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestRecoveryMiddleware_ErrorFormatting(t *testing.T) {
	log := logger.New(logger.LevelError)
	handler := RecoveryMiddleware(true, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	handler.ServeHTTP(w, r)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that error object has proper structure
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error to be an object")
	}

	code, ok := errObj["code"].(string)
	if !ok || code != "internal_error" {
		t.Errorf("Expected error code 'internal_error', got %v", code)
	}

	message, ok := errObj["message"].(string)
	if !ok || message != "internal server error" {
		t.Errorf("Expected error message 'internal server error', got %v", message)
	}
}

func TestIsDevEnvironment_Dev(t *testing.T) {
	t.Setenv("BANI_ENVIRONMENT", "dev")
	if !IsDevEnvironment() {
		t.Error("Expected IsDevEnvironment to return true for dev environment")
	}
}

func TestIsDevEnvironment_Empty(t *testing.T) {
	t.Setenv("BANI_ENVIRONMENT", "")
	if !IsDevEnvironment() {
		t.Error("Expected IsDevEnvironment to return true for empty environment")
	}
}

func TestIsDevEnvironment_Production(t *testing.T) {
	t.Setenv("BANI_ENVIRONMENT", "production")
	if IsDevEnvironment() {
		t.Error("Expected IsDevEnvironment to return false for production environment")
	}
}

func TestIsDevEnvironment_Staging(t *testing.T) {
	t.Setenv("BANI_ENVIRONMENT", "staging")
	if IsDevEnvironment() {
		t.Error("Expected IsDevEnvironment to return false for staging environment")
	}
}
