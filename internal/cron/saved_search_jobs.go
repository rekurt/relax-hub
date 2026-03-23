package cron

import (
	"context"
	"time"
)

// handleSavedSearchCheck checks for new bathhouses matching saved searches with notifications enabled
func (cs *CronScheduler) handleSavedSearchCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting saved search new matches check")

	if cs.savedSearchSvc == nil {
		cs.logger.Warn("Saved search service not available, skipping")
		return
	}

	if err := cs.savedSearchSvc.CheckNewMatches(ctx); err != nil {
		cs.logger.Error("Saved search check failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Saved search check completed", "duration", time.Since(start))
}
