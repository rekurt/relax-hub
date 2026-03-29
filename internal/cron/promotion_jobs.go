package cron

import "context"

// promotionDailyBudget deducts the daily bid from owner wallets for all active promotions.
// Campaigns are paused when wallet balance is insufficient or budget is exhausted.
// Campaigns are expired when past their end date.
func (cs *CronScheduler) promotionDailyBudget(ctx context.Context) error {
	if cs.promotionSvc == nil {
		return nil
	}
	return cs.promotionSvc.DeductDailyBudgets(ctx)
}
