package cron

import (
	"context"
	"time"
)

func (cs *CronScheduler) handleAutoScenarioExecution() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting auto scenario execution")

	sent, err := cs.autoScenarioSvc.ExecuteScenarios(ctx)
	if err != nil {
		cs.logger.Error("Auto scenario execution failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Auto scenario execution completed", "sent", sent, "duration", time.Since(start))
}
