package cron

import (
	"context"
	"fmt"
)

// ownerResponseRateMonitoring recalculates response rates for request-mode bathhouses.
func (cs *CronScheduler) ownerResponseRateMonitoring(ctx context.Context) error {
	updated, err := cs.bookingSvc.RecalculateResponseRates(ctx)
	if err != nil {
		return fmt.Errorf("owner response rate monitoring: %w", err)
	}
	cs.logger.Info("Owner response rate monitoring done", "updated", updated)
	return nil
}
