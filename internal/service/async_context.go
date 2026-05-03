package service

import (
	"context"
	"time"
)

func runDetached(ctx context.Context, timeout time.Duration, fn func(context.Context)) {
	detached := context.WithoutCancel(ctx)
	go func() {
		runCtx, cancel := context.WithTimeout(detached, timeout)
		defer cancel()
		fn(runCtx)
	}()
}
