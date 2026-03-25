package cron

import (
	"context"
	"fmt"
)

// accountDeletionExecution processes accounts past their 30-day grace period.
func (cs *CronScheduler) accountDeletionExecution(ctx context.Context) error {
	executed, err := cs.accountDeletionSvc.CheckPendingDeletions(ctx)
	if err != nil {
		return fmt.Errorf("check pending deletions: %w", err)
	}
	cs.logger.Info("Account deletion execution done", "executed", executed)
	return nil
}

// accountDeletionReminders sends reminders to users with pending deletions.
func (cs *CronScheduler) accountDeletionReminders(ctx context.Context) error {
	return cs.accountDeletionSvc.SendDeletionReminders(ctx)
}
