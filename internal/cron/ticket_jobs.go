package cron

import (
	"context"
	"time"
)

func (cs *CronScheduler) handleTicketAutoEscalation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := cs.ticketSvc.AutoEscalateStaleTickets(ctx); err != nil {
		cs.logger.Error("ticket auto-escalation failed", "error", err)
	}
}

func (cs *CronScheduler) handleTicketAutoClose() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := cs.ticketSvc.AutoCloseResolvedTickets(ctx); err != nil {
		cs.logger.Error("ticket auto-close failed", "error", err)
	}
}
