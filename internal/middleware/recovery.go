package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// RecoveryMiddleware recovers from panics and logs the error with stack trace
// Stack traces are only shown in dev mode (via environment detection)
func RecoveryMiddleware(isDev bool, log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					reqID := chiMiddleware.GetReqID(r.Context())

					if isDev {
						// In dev mode, log full stack trace
						stackTrace := string(debug.Stack())
						log.Error("PANIC recovered",
							"request_id", reqID,
							"path", r.URL.Path,
							"method", r.Method,
							"panic", fmt.Sprintf("%v", rvr),
							"stack", stackTrace,
						)
					} else {
						// In production, log without detailed stack trace
						log.Error("PANIC recovered",
							"request_id", reqID,
							"path", r.URL.Path,
							"method", r.Method,
							"panic", fmt.Sprintf("%v", rvr),
						)
					}

					// Send generic error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"error": map[string]string{
							"code":    "internal_error",
							"message": "internal server error",
						},
					}); err != nil {
						log.Error("Failed to encode panic response", "error", err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// IsDevEnvironment checks if the given environment string represents dev mode
func IsDevEnvironment(env string) bool {
	return env == "" || env == "dev"
}
