package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/database"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
	log   *logger.Logger
}

type HealthParams struct {
	fx.In

	DB    *pgxpool.Pool
	Redis *redis.Client
	Log   *logger.Logger
}

func NewHealthHandler(p HealthParams) *HealthHandler {
	return &HealthHandler{
		db:    p.DB,
		redis: p.Redis,
		log:   p.Log,
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ReadyResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Health returns a simple liveness probe response
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthResponse{Status: "ok"}); err != nil {
		h.log.Error("Failed to encode health response", "error", err)
	}
}

// Ready returns readiness status after checking external dependencies
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allReady := true

	// Check PostgreSQL
	pgCtx, pgCancel := context.WithTimeout(ctx, 3*time.Second)
	if err := h.db.Ping(pgCtx); err != nil {
		h.log.Error("Readiness check failed: PostgreSQL unavailable", "error", err)
		services["postgres"] = "down"
		allReady = false
	} else {
		services["postgres"] = "up"
	}
	pgCancel()

	// Check Redis
	redisCtx, redisCancel := context.WithTimeout(ctx, database.RedisOperationTimeout)
	if err := h.redis.Ping(redisCtx).Err(); err != nil {
		h.log.Error("Readiness check failed: Redis unavailable", "error", err)
		services["redis"] = "down"
		allReady = false
	} else {
		services["redis"] = "up"
	}
	redisCancel()

	w.Header().Set("Content-Type", "application/json")

	if !allReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(ReadyResponse{
			Status:   "not_ready",
			Services: services,
		}); err != nil {
			h.log.Error("Failed to encode ready response", "error", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ReadyResponse{
		Status:   "ready",
		Services: services,
	}); err != nil {
		h.log.Error("Failed to encode ready response", "error", err)
	}
}
