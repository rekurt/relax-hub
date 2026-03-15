package middleware

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// APIError represents an error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIResponse represents the standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*clientLimiter
}

type clientLimiter struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimitInfo holds the result of a rate limit check.
type RateLimitInfo struct {
	Allowed   bool
	Remaining int
	RetryAfter time.Duration
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*clientLimiter),
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

// Allow checks if the request is allowed under the given rate.
func (rl *RateLimiter) Allow(key string, ratePerSecond float64) bool {
	info := rl.Check(key, ratePerSecond)
	return info.Allowed
}

// Check performs a rate limit check and returns detailed info including remaining tokens.
func (rl *RateLimiter) Check(key string, ratePerSecond float64) RateLimitInfo {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limiter, exists := rl.limiters[key]

	if !exists {
		limiter = &clientLimiter{
			tokens:    ratePerSecond,
			lastCheck: now,
		}
		rl.limiters[key] = limiter
		remaining := int(limiter.tokens) - 1
		limiter.tokens--
		return RateLimitInfo{Allowed: true, Remaining: remaining}
	}

	elapsed := now.Sub(limiter.lastCheck).Seconds()
	limiter.tokens += elapsed * ratePerSecond
	if limiter.tokens > ratePerSecond {
		limiter.tokens = ratePerSecond
	}
	limiter.lastCheck = now

	if limiter.tokens >= 1 {
		limiter.tokens--
		return RateLimitInfo{
			Allowed:   true,
			Remaining: int(limiter.tokens),
		}
	}

	// Calculate how long until next token is available
	deficit := 1.0 - limiter.tokens
	retryAfter := time.Duration(math.Ceil(deficit/ratePerSecond)) * time.Second
	return RateLimitInfo{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: retryAfter,
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limiter := range rl.limiters {
		if now.Sub(limiter.lastCheck) > 5*time.Minute {
			delete(rl.limiters, key)
		}
	}
}

// ClientIP extracts the client IP address from the request, stripping the port.
// It checks X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs; take the first (original client)
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Strip port from RemoteAddr (e.g., "192.168.1.1:12345" -> "192.168.1.1")
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit creates a middleware that rate-limits requests by client IP.
func RateLimit(rl *RateLimiter, ratePerSecond float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r)
			info := rl.Check(ip, ratePerSecond)

			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))

			if !info.Allowed {
				retryAfterSec := int(math.Ceil(info.RetryAfter.Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Error: &APIError{
						Code:    "rate_limit_exceeded",
						Message: "rate limit exceeded",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WidgetRateLimit creates a middleware that rate-limits widget requests by API key.
func WidgetRateLimit(rl *RateLimiter, ratePerSecond float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ""
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) > 2 && pathParts[1] == "widget" && pathParts[2] != "" {
				key = pathParts[2]
			}

			if key == "" {
				key = ClientIP(r)
			}

			info := rl.Check(key, ratePerSecond)

			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))

			if !info.Allowed {
				retryAfterSec := int(math.Ceil(info.RetryAfter.Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Error: &APIError{
						Code:    "rate_limit_exceeded",
						Message: "rate limit exceeded",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
