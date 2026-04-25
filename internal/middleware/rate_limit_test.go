package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.168.1.1:12345"

	got := ClientIP(r)
	if got != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", got)
	}
}

func TestClientIP_RemoteAddrNoPort(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.168.1.1"

	got := ClientIP(r)
	if got != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", got)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	got := ClientIP(r)
	if got != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %s", got)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Real-IP", "203.0.113.99")

	got := ClientIP(r)
	if got != "203.0.113.99" {
		t.Errorf("expected 203.0.113.99, got %s", got)
	}
}

func TestClientIP_IPv6(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "[::1]:12345"

	got := ClientIP(r)
	if got != "::1" {
		t.Errorf("expected ::1, got %s", got)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}

	// First request should always be allowed
	if !rl.Allow("test", 1.0) {
		t.Error("first request should be allowed")
	}
}

func TestRateLimiter_Check_FirstRequest(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}

	// Capacity = max(rate*60, 5). For rate=2.0 → 120.
	info := rl.Check("client1", 2.0)
	if !info.Allowed {
		t.Error("first request should be allowed")
	}
	if info.Remaining != 119 {
		t.Errorf("expected remaining=119 (started with 120 tokens, used 1), got %d", info.Remaining)
	}
}

func TestRateLimiter_Check_ExhaustedTokens(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}

	// Bucket capacity floor is 5 — drain it, then the next request denies.
	const rate = 0.001 // capacity = max(0.06, 5) = 5
	for i := 0; i < 5; i++ {
		if !rl.Check("client1", rate).Allowed {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}

	// The 6th immediate call should be denied — refill rate is tiny, no
	// token has been earned in test runtime.
	info := rl.Check("client1", rate)
	if info.Allowed {
		t.Error("6th request should be denied when tokens exhausted")
	}
	if info.Remaining != 0 {
		t.Errorf("expected remaining=0, got %d", info.Remaining)
	}
	if info.RetryAfter <= 0 {
		t.Error("expected positive RetryAfter")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}

	// Add a stale entry
	rl.limiters["stale"] = &clientLimiter{
		tokens:    1.0,
		lastCheck: time.Now().Add(-10 * time.Minute),
	}
	// Add a fresh entry
	rl.limiters["fresh"] = &clientLimiter{
		tokens:    1.0,
		lastCheck: time.Now(),
	}

	rl.cleanup()

	if _, exists := rl.limiters["stale"]; exists {
		t.Error("stale entry should have been cleaned up")
	}
	if _, exists := rl.limiters["fresh"]; !exists {
		t.Error("fresh entry should not have been cleaned up")
	}
}

func TestRateLimit_Middleware_AllowsRequest(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}
	handler := RateLimit(rl, 10.0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("expected X-RateLimit-Remaining header")
	}
}

func TestRateLimit_Middleware_Returns429(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}
	// rate=0.001/s gives capacity=max(0.06, 5)=5 — small bucket, easy to drain.
	const rate = 0.001
	handler := RateLimit(rl, rate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"

	// Drain the bucket (5 tokens).
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("call %d should be 200, got %d", i+1, rec.Code)
		}
	}

	// 6th request should be rate limited.
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec2.Code)
	}

	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header on 429 response")
	}

	remainingHeader := rec2.Header().Get("X-RateLimit-Remaining")
	if remainingHeader != "0" {
		t.Errorf("expected X-RateLimit-Remaining=0, got %s", remainingHeader)
	}

	var resp APIResponse
	if err := json.NewDecoder(rec2.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "rate_limit_exceeded" {
		t.Error("expected rate_limit_exceeded error code")
	}
}

func TestRateLimit_Middleware_StripsPort(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}
	// Use the floor-capacity rate (5 tokens) so we can drain quickly.
	const rate = 0.001
	handler := RateLimit(rl, rate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Same IP, different ports — should share the bucket.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	// Now drained. Hit again from different port on same IP — must 429.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:2222"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for same IP with different port, got %d", rec2.Code)
	}
}

func TestWidgetRateLimit_UsesAPIKey(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}
	const rate = 0.001 // capacity floor 5
	handler := WidgetRateLimit(rl, rate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Drain bucket for a particular API key.
	req := httptest.NewRequest(http.MethodGet, "/widget/test-key-123/bathhouse", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("call %d should be 200, got %d", i+1, rec.Code)
		}
	}

	// Next request for the same API key should be rate-limited.
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for repeated widget request, got %d", rec2.Code)
	}

	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header on widget 429 response")
	}
}

func TestWidgetRateLimit_FallsBackToIP(t *testing.T) {
	rl := &RateLimiter{limiters: make(map[string]*clientLimiter)}
	handler := WidgetRateLimit(rl, 1.0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without API key in expected path position
	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
