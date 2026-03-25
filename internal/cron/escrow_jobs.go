package cron

import (
	"context"
	"fmt"
)

func (cs *CronScheduler) escrowReleaseJob(ctx context.Context) error {
	released, err := cs.escrowSvc.ProcessMaturedEscrows(ctx)
	if err != nil {
		return fmt.Errorf("process matured escrows: %w", err)
	}
	cs.logger.Info("Escrow release done", "released", released)
	return nil
}
