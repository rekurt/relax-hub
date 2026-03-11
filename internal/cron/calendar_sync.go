package cron

import (
	"context"
	"time"
)

// handleCalendarSync syncs all registered external calendar feeds.
func (cs *CronScheduler) handleCalendarSync() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting external calendar sync")

	cs.calendarSync.SyncAllCalendars(ctx)

	cs.logger.Info("External calendar sync completed", "duration", time.Since(start))
}
