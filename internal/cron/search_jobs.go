package cron

import (
	"context"
	"fmt"
)

// bathhouseMetricsUpdate recalculates conversion rate and occupancy rate for all bathhouses.
func (cs *CronScheduler) bathhouseMetricsUpdate(ctx context.Context) error {
	updated, err := cs.analyticsService.UpdateBathhouseMetrics(ctx)
	if err != nil {
		return fmt.Errorf("bathhouse metrics update: %w", err)
	}
	cs.logger.Info("Bathhouse metrics update done", "updated", updated)
	return nil
}
