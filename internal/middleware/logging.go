package middleware

import (
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rekurt/relax-hub/internal/logger"
)

func Logging(l logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				reqID := chiMiddleware.GetReqID(r.Context())
				l.Info(
					"Request",
					"request_id", reqID,
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration", time.Since(start),
					"remote_addr", r.RemoteAddr,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
