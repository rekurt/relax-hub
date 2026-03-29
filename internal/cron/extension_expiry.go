package cron

import (
	"context"
	"fmt"
)

func (cs *CronScheduler) extensionRequestExpiry(ctx context.Context) error {
	expired, err := cs.extReqSvc.ExpireTimedOutRequests(ctx)
	if err != nil {
		return fmt.Errorf("expire extension requests: %w", err)
	}
	if expired > 0 {
		cs.logger.Info("expired booking extension requests", "count", expired)
	}
	return nil
}
