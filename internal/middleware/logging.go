package middleware

import (
	"log"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			reqID := chiMiddleware.GetReqID(r.Context())
			log.Printf(
				"[%s] %s %s %d %s %s",
				reqID,
				r.Method,
				r.URL.Path,
				ww.Status(),
				time.Since(start),
				r.RemoteAddr,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
