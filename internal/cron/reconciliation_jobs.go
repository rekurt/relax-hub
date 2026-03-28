package cron

import (
	"context"
	"fmt"
	"time"
)

func (cs *CronScheduler) dailyFloatSnapshot(ctx context.Context) error {
	if cs.reconciliationSvc == nil {
		return nil
	}
	snapshot, err := cs.reconciliationSvc.TakeFloatSnapshot(ctx)
	if err != nil {
		return fmt.Errorf("take float snapshot: %w", err)
	}
	cs.logger.Info("Daily float snapshot taken",
		"status", snapshot.Status,
		"client_wallets", snapshot.ClientWalletsTotal,
		"owner_wallets", snapshot.OwnerWalletsTotal,
		"escrow_held", snapshot.EscrowHeldTotal,
		"expected_total", snapshot.ExpectedTotal,
	)
	return nil
}

func (cs *CronScheduler) dailyReconciliation(ctx context.Context) error {
	if cs.reconciliationSvc == nil {
		return nil
	}
	now := time.Now()
	from := now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
	to := now.Truncate(24 * time.Hour)

	report, err := cs.reconciliationSvc.ReconcileWithProvider(ctx, from, to)
	if err != nil {
		return fmt.Errorf("reconcile with provider: %w", err)
	}
	cs.logger.Info("Daily reconciliation done",
		"status", report.Status,
		"internal_payments", report.InternalPaymentsSum,
		"provider_payments", report.ProviderPaymentsSum,
		"payment_discrepancy", report.PaymentDiscrepancy,
	)
	return nil
}
