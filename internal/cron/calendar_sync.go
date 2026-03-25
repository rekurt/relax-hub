package cron

import (
	"context"
)

// calendarSyncJob syncs all registered external calendar feeds.
func (cs *CronScheduler) calendarSyncJob(ctx context.Context) error {
	cs.calendarSync.SyncAllCalendars(ctx)
	return nil
}
