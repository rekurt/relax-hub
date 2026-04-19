package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekurt/relax-hub/config"
)

func TestNewCORSMiddleware_DefaultOrigins(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	}

	cors := NewCORSMiddleware(cfg)
	if cors == nil {
		t.Fatal("expected non-nil CORSMiddleware")
	}

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin = *, got %s", origin)
	}
}

func TestNewCORSMiddleware_SpecificOrigins(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"https://example.com", "https://app.example.com"}},
	}

	cors := NewCORSMiddleware(cfg)

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Allowed origin
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin = https://example.com, got %s", origin)
	}

	// Disallowed origin
	req2 := httptest.NewRequest(http.MethodOptions, "/", nil)
	req2.Header.Set("Origin", "https://evil.com")
	req2.Header.Set("Access-Control-Request-Method", "GET")
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	origin2 := rec2.Header().Get("Access-Control-Allow-Origin")
	if origin2 != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin for disallowed origin, got %s", origin2)
	}
}

func TestNewCORSMiddleware_EmptyOriginsFallsBackToWildcard(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: nil},
	}

	cors := NewCORSMiddleware(cfg)

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://anything.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin = * for empty config, got %s", origin)
	}
}

func TestNewCORSMiddleware_AllowedMethods(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	}

	cors := NewCORSMiddleware(cfg)

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected non-empty Access-Control-Allow-Methods")
	}
}
