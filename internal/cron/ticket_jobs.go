package cron

import (
	"context"

	"github.com/nikitaaldaev/bani/internal/domain"
)

func (cs *CronScheduler) ticketAutoEscalation(ctx context.Context) error {
	err := cs.ticketSvc.AutoEscalateStaleTickets(ctx)
	if err != nil {
		return err
	}

	// Emit SLA violation notification if there are stale tickets
	if cs.adminNotificationSvc != nil {
		// Check for tickets that breached SLA (older than 24h without review)
		counts, countErr := cs.ticketSvc.GetStats(ctx)
		if countErr == nil && counts != nil && counts.Open > 0 {
			_ = cs.adminNotificationSvc.Emit(ctx,
				domain.AdminNotifSLAViolation,
				domain.AdminNotifSeverityWarning,
				"Тикеты ожидают обработки",
				"Есть открытые тикеты, требующие внимания. Проверьте очередь поддержки.",
				map[string]interface{}{"open_count": counts.Open},
			)
		}
	}
	return nil
}

func (cs *CronScheduler) ticketAutoCloseJob(ctx context.Context) error {
	return cs.ticketSvc.AutoCloseResolvedTickets(ctx)
}
