package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"github.com/rekurt/relax-hub/config"
	"go.uber.org/fx"
)

var CORSModule = fx.Module("cors",
	fx.Provide(NewCORSMiddleware),
)

type CORSMiddleware struct {
	handler func(http.Handler) http.Handler
}

func NewCORSMiddleware(cfg *config.Config) *CORSMiddleware {
	origins := cfg.CORS.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	return &CORSMiddleware{handler: c.Handler}
}

func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return m.handler(next)
}
