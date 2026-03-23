package cron

import (
	"context"
	"time"
)

// handleSessionCleanup removes expired sessions from the database.
func (cs *CronScheduler) handleSessionCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting session cleanup")

	cleaned, err := cs.sessionSvc.CleanExpired(ctx)
	if err != nil {
		cs.logger.Error("Session cleanup failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Session cleanup completed",
		"cleaned", cleaned,
		"duration", time.Since(start),
	)
}
