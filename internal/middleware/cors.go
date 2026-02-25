package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"go.uber.org/fx"
)

var CORSModule = fx.Module("cors",
	fx.Provide(NewCORSMiddleware),
)

type CORSMiddleware struct {
	handler func(http.Handler) http.Handler
}

func NewCORSMiddleware() *CORSMiddleware {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return &CORSMiddleware{handler: c.Handler}
}

func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return m.handler(next)
}
