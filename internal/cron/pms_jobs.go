package cron

import (
	"context"
)

// pmsSyncJob synchronizes bookings with external PMS systems for all active connections.
func (cs *CronScheduler) pmsSyncJob(ctx context.Context) error {
	return cs.pmsSvc.SyncAllActive(ctx)
}
