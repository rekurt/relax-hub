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

func (cs *CronScheduler) handleAutoReviewRequests() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting auto review requests")

	delayHours := cs.cfg.Review.RequestDelayHours
	if delayHours <= 0 {
		delayHours = 2
	}

	sent, err := cs.reviewSvc.SendReviewRequests(ctx, delayHours)
	if err != nil {
		cs.logger.Error("Auto review requests failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Auto review requests completed", "sent", sent, "duration", time.Since(start))
}
