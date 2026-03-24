package cron

import (
	"context"
	"time"
)

func (cs *CronScheduler) handleAutoRejectTimedOutRequests() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting auto-reject timed out booking requests")

	rejected, err := cs.bookingSvc.AutoRejectTimedOutRequests(ctx)
	if err != nil {
		cs.logger.Error("Auto-reject timed out requests failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Auto-reject timed out requests completed", "rejected", rejected, "duration", time.Since(start))
}

func (cs *CronScheduler) handleNoShowDetection() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting no-show detection")

	marked, err := cs.bookingSvc.MarkNoShows(ctx)
	if err != nil {
		cs.logger.Error("No-show detection failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("No-show detection completed", "marked", marked, "duration", time.Since(start))
}
