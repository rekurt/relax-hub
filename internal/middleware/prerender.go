package middleware

import (
	"net/http"

	"github.com/rekurt/relax-hub/internal/seo"
)

// Prerender is a middleware that detects bot user-agents and redirects them
// to pre-rendered HTML pages for SEO. Non-bot requests pass through to the
// next handler (typically the SPA).
func Prerender(renderer *seo.Renderer, botHandler http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if seo.IsBotUserAgent(ua) && botHandler != nil {
				botHandler.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
