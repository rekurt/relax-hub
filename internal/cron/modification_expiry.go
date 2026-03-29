package cron

import (
	"context"
	"fmt"
)

func (cs *CronScheduler) modificationRequestExpiry(ctx context.Context) error {
	expired, err := cs.modReqSvc.ExpireTimedOutRequests(ctx)
	if err != nil {
		return fmt.Errorf("expire modification requests: %w", err)
	}
	if expired > 0 {
		cs.logger.Info("expired booking modification requests", "count", expired)
	}
	return nil
}
