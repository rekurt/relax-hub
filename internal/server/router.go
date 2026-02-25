package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

func NewRouter(cors *middleware.CORSMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler)

	r.Get("/health", healthCheck)

	r.Route("/api/v1", func(r chi.Router) {
		// Routes will be added in subsequent tasks
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
