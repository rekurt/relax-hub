package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*clientLimiter
}

type clientLimiter struct {
	tokens    float64
	lastCheck time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*clientLimiter),
	}
	// Cleanup stale entries every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow(key string, ratePerSecond float64) bool {
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
		return true
	}

	// Add tokens based on time elapsed
	elapsed := now.Sub(limiter.lastCheck).Seconds()
	limiter.tokens += elapsed * ratePerSecond
	if limiter.tokens > ratePerSecond {
		limiter.tokens = ratePerSecond
	}
	limiter.lastCheck = now

	if limiter.tokens >= 1 {
		limiter.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limiter := range rl.limiters {
		// Remove entries that haven't been used in 5 minutes
		if now.Sub(limiter.lastCheck) > 5*time.Minute {
			delete(rl.limiters, key)
		}
	}
}

func WidgetRateLimit(rl *RateLimiter, ratePerSecond float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from URL path /widget/{api_key}/...
			apiKey := ""
			pathParts := strings.Split(r.URL.Path, "/")
			// pathParts[0] = "", pathParts[1] = "widget", pathParts[2] = api_key
			if len(pathParts) > 2 && pathParts[1] == "widget" && pathParts[2] != "" {
				apiKey = pathParts[2]
			}

			if apiKey == "" {
				// Fallback to IP address if API key not found
				apiKey = r.RemoteAddr
			}

			if !rl.Allow(apiKey, ratePerSecond) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
