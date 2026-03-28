package cron

import (
	"context"
	"time"
)

// adminNotificationDigest sends daily email digest of unread critical alerts per admin role.
func (cs *CronScheduler) adminNotificationDigest(ctx context.Context) error {
	if cs.adminNotificationSvc == nil {
		return nil
	}
	return cs.adminNotificationSvc.SendDailyDigest(ctx)
}

// adminNotificationCleanup deletes read admin notifications older than 90 days.
func (cs *CronScheduler) adminNotificationCleanup(ctx context.Context) error {
	if cs.adminNotificationRepo == nil {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -90)
	deleted, err := cs.adminNotificationRepo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return err
	}
	cs.logger.Info("Admin notification cleanup done", "deleted_count", deleted, "cutoff", cutoff)
	return nil
}
