package cron

import (
	"context"
	"fmt"
)

func (cs *CronScheduler) autoScenarioExecution(ctx context.Context) error {
	sent, err := cs.autoScenarioSvc.ExecuteScenarios(ctx)
	if err != nil {
		return fmt.Errorf("execute scenarios: %w", err)
	}
	cs.logger.Info("Auto scenario execution done", "sent", sent)
	return nil
}
