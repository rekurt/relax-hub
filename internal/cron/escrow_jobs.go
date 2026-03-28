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

func (cs *CronScheduler) depositReleaseJob(ctx context.Context) error {
	if cs.depositSvc == nil {
		return nil
	}
	released, err := cs.depositSvc.ProcessMaturedDeposits(ctx)
	if err != nil {
		return fmt.Errorf("process matured deposits: %w", err)
	}
	if released > 0 {
		cs.logger.Info("Deposit release done", "released", released)
	}
	return nil
}
