package cron

import (
	"context"
	"time"
)

// handleAccountDeletionExecution processes accounts past their 30-day grace period.
func (cs *CronScheduler) handleAccountDeletionExecution() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting account deletion execution")

	executed, err := cs.accountDeletionSvc.CheckPendingDeletions(ctx)
	if err != nil {
		cs.logger.Error("Account deletion execution failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Account deletion execution completed",
		"executed", executed,
		"duration", time.Since(start),
	)
}

// handleAccountDeletionReminders sends reminders to users with pending deletions.
func (cs *CronScheduler) handleAccountDeletionReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting account deletion reminders")

	if err := cs.accountDeletionSvc.SendDeletionReminders(ctx); err != nil {
		cs.logger.Error("Account deletion reminders failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Account deletion reminders completed", "duration", time.Since(start))
}
