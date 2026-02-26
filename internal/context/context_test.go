package context

import (
	"context"
	"testing"
	"time"
)

func TestWithQueryTimeout(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	if timeoutCtx == nil {
		t.Error("WithQueryTimeout should return a non-nil context")
	}

	// Verify the context has a deadline
	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Error("WithQueryTimeout should set a deadline")
	}

	// Verify the deadline is in the future
	if deadline.Before(time.Now()) {
		t.Error("WithQueryTimeout deadline should be in the future")
	}
}

func TestWithRedisTimeout(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := WithRedisTimeout(ctx)
	defer cancel()

	if timeoutCtx == nil {
		t.Error("WithRedisTimeout should return a non-nil context")
	}

	// Verify the context has a deadline
	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Error("WithRedisTimeout should set a deadline")
	}

	// Verify the deadline is in the future
	if deadline.Before(time.Now()) {
		t.Error("WithRedisTimeout deadline should be in the future")
	}
}

func TestWithCustomTimeout(t *testing.T) {
	ctx := context.Background()
	customTimeout := 10 * time.Second
	timeoutCtx, cancel := WithCustomTimeout(ctx, customTimeout)
	defer cancel()

	if timeoutCtx == nil {
		t.Error("WithCustomTimeout should return a non-nil context")
	}

	// Verify the context has a deadline
	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Error("WithCustomTimeout should set a deadline")
	}

	// Verify the deadline is in the future
	if deadline.Before(time.Now()) {
		t.Error("WithCustomTimeout deadline should be in the future")
	}
}

func TestTimeoutContextCancellation(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := WithQueryTimeout(ctx)

	// Verify the context is not cancelled yet
	select {
	case <-timeoutCtx.Done():
		t.Error("Context should not be cancelled immediately")
	default:
		// This is expected
	}

	// Cancel the context
	cancel()

	// Verify the context is now cancelled
	select {
	case <-timeoutCtx.Done():
		// This is expected
	default:
		t.Error("Context should be cancelled after cancel() is called")
	}
}

func TestTimeoutContextError(t *testing.T) {
	ctx := context.Background()
	timeoutCtx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	err := timeoutCtx.Err()
	if err != nil {
		t.Errorf("Context should not have an error yet, got: %v", err)
	}
}
