package cron

import "context"

func (cs *CronScheduler) ticketAutoEscalation(ctx context.Context) error {
	return cs.ticketSvc.AutoEscalateStaleTickets(ctx)
}

func (cs *CronScheduler) ticketAutoCloseJob(ctx context.Context) error {
	return cs.ticketSvc.AutoCloseResolvedTickets(ctx)
}
