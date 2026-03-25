package cron

import "context"

func (cs *CronScheduler) handleTicketAutoEscalation() {
	ctx := context.Background()
	if err := cs.ticketSvc.AutoEscalateStaleTickets(ctx); err != nil {
		cs.logger.Error("ticket auto-escalation failed", "error", err)
	}
}

func (cs *CronScheduler) handleTicketAutoClose() {
	ctx := context.Background()
	if err := cs.ticketSvc.AutoCloseResolvedTickets(ctx); err != nil {
		cs.logger.Error("ticket auto-close failed", "error", err)
	}
}
