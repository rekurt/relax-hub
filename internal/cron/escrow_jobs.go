package cron

import (
	"context"
	"time"
)

func (cs *CronScheduler) handleEscrowRelease() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting escrow release processing")

	released, err := cs.escrowSvc.ProcessMaturedEscrows(ctx)
	if err != nil {
		cs.logger.Error("Escrow release processing failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Escrow release processing completed", "released", released, "duration", time.Since(start))
}
