package cron

import (
	"context"
	"fmt"
)

func (cs *CronScheduler) platformAverageRefresh(ctx context.Context) error {
	return cs.reviewSvc.RefreshPlatformAverage(ctx)
}

func (cs *CronScheduler) autoReviewRequests(ctx context.Context) error {
	delayHours := cs.cfg.Review.RequestDelayHours
	if delayHours <= 0 {
		delayHours = 2
	}

	sent, err := cs.reviewSvc.SendReviewRequests(ctx, delayHours)
	if err != nil {
		return fmt.Errorf("send review requests: %w", err)
	}
	cs.logger.Info("Auto review requests done", "sent", sent)
	return nil
}

func (cs *CronScheduler) reviewAutoReveal(ctx context.Context) error {
	if cs.clientReviewSvc == nil {
		return nil
	}
	revealed, err := cs.clientReviewSvc.RevealExpired(ctx)
	if err != nil {
		return fmt.Errorf("reveal expired reviews: %w", err)
	}
	if revealed > 0 {
		cs.logger.Info("Auto-revealed reviews", "count", revealed)
	}
	return nil
}
