package cron

import (
	"context"
	"time"
)

func (cs *CronScheduler) handlePlatformAverageRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting platform average rating refresh")

	if err := cs.reviewSvc.RefreshPlatformAverage(ctx); err != nil {
		cs.logger.Error("Platform average rating refresh failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Platform average rating refresh completed", "duration", time.Since(start))
}
