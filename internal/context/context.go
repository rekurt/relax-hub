package context

import (
	"context"
	"time"

	"github.com/nikitaaldaev/bani/internal/database"
)

// WithQueryTimeout wraps the given context with a timeout for database queries
// Uses the default query timeout constant from the database package
func WithQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, database.DefaultQueryTimeout)
}

// WithRedisTimeout wraps the given context with a timeout for Redis operations
// Uses the Redis operation timeout constant from the database package
func WithRedisTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, database.RedisOperationTimeout)
}

// WithCustomTimeout wraps the given context with a custom timeout duration
func WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
