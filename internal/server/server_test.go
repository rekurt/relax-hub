package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/server"
)

func TestNewRouter(t *testing.T) {
	cors := middleware.NewCORSMiddleware()
	router := server.NewRouter(cors)

	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestHealthCheck(t *testing.T) {
	cors := middleware.NewCORSMiddleware()
	router := server.NewRouter(cors)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestHealthCheck_WrongMethod(t *testing.T) {
	cors := middleware.NewCORSMiddleware()
	router := server.NewRouter(cors)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}
