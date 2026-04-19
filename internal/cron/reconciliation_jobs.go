package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
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

	if snapshot.Status != "ok" && cs.adminNotificationSvc != nil {
		_ = cs.adminNotificationSvc.Emit(ctx,
			domain.AdminNotifFloatDrift,
			domain.AdminNotifSeverityError,
			"Обнаружено расхождение баланса платформы",
			fmt.Sprintf("Статус: %s. Ожидаемый итог: %d коп.", snapshot.Status, snapshot.ExpectedTotal),
			map[string]interface{}{
				"status":         snapshot.Status,
				"expected_total": snapshot.ExpectedTotal,
			},
		)
	}

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

	if report.PaymentDiscrepancy != 0 && cs.adminNotificationSvc != nil {
		_ = cs.adminNotificationSvc.Emit(ctx,
			domain.AdminNotifReconciliationMismatch,
			domain.AdminNotifSeverityCritical,
			"Расхождение в сверке платежей",
			fmt.Sprintf("Расхождение: %d коп. Внутренние: %d, Провайдер: %d",
				report.PaymentDiscrepancy, report.InternalPaymentsSum, report.ProviderPaymentsSum),
			map[string]interface{}{
				"discrepancy":       report.PaymentDiscrepancy,
				"internal_payments": report.InternalPaymentsSum,
				"provider_payments": report.ProviderPaymentsSum,
			},
		)
	}

	return nil
}
