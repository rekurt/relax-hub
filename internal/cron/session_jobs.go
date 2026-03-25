package cron

import (
	"context"
	"fmt"
)

// sessionCleanupJob removes expired sessions from the database.
func (cs *CronScheduler) sessionCleanupJob(ctx context.Context) error {
	cleaned, err := cs.sessionSvc.CleanExpired(ctx)
	if err != nil {
		return fmt.Errorf("clean expired sessions: %w", err)
	}
	cs.logger.Info("Session cleanup done", "cleaned", cleaned)
	return nil
}
