package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

// bonusExpiration expires bonus transactions that have passed their expiry date.
func (cs *CronScheduler) bonusExpiration(ctx context.Context) error {
	walletIDs, err := cs.walletRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list wallet IDs: %w", err)
	}

	totalExpired := 0
	walletsProcessed := 0
	for _, wid := range walletIDs {
		expired, err := cs.walletSvc.ExpireBonusesForWallet(ctx, wid)
		if err != nil {
			cs.logger.Error("Failed to expire bonuses for wallet", "wallet_id", wid, "error", err)
			continue
		}

		if expired > 0 {
			wallet, err := cs.walletRepo.GetByID(ctx, wid)
			if err == nil {
				_ = cs.notifSvc.Send(ctx, wallet.UserID, domain.NotifBonusExpired,
					"Бонусы истекли",
					fmt.Sprintf("Истёк срок действия %d бонусных начислений.", expired),
					map[string]string{
						"wallet_id":     wid.String(),
						"expired_count": fmt.Sprintf("%d", expired),
					},
				)
			}
		}

		totalExpired += expired
		walletsProcessed++
	}

	cs.logger.Info("Bonus expiration done",
		"wallets_processed", walletsProcessed,
		"total_expired", totalExpired,
	)
	return nil
}

// bonusExpiryNotify warns users about bonuses expiring soon (at 14 and 3 days before expiry).
func (cs *CronScheduler) bonusExpiryNotify(ctx context.Context) error {
	walletIDs, err := cs.walletRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list wallet IDs: %w", err)
	}

	now := time.Now()
	notified := 0

	for _, wid := range walletIDs {
		for _, daysAhead := range []int{14, 3} {
			windowStart := now.Add(time.Duration(daysAhead-1) * 24 * time.Hour)
			windowEnd := now.Add(time.Duration(daysAhead) * 24 * time.Hour)

			expiring, err := cs.walletRepo.GetExpiringBonusesSoon(ctx, wid, windowStart, windowEnd)
			if err != nil {
				cs.logger.Error("Failed to get expiring bonuses soon", "wallet_id", wid, "error", err)
				continue
			}

			if len(expiring) == 0 {
				continue
			}

			var totalAmount int64
			for _, tx := range expiring {
				totalAmount += tx.Amount
			}

			wallet, err := cs.walletRepo.GetByID(ctx, wid)
			if err != nil {
				cs.logger.Error("Failed to get wallet for bonus expiry notify", "wallet_id", wid, "error", err)
				continue
			}

			expiryDays := cs.getBonusExpiryDays()
			_ = cs.notifSvc.Send(ctx, wallet.UserID, domain.NotifBonusExpiring,
				"Бонусы скоро истекут",
				fmt.Sprintf("Через %s истекут бонусы на сумму %d ₽. Срок действия бонусов — %d дней.",
					formatDays(daysAhead), totalAmount/100, expiryDays),
				map[string]string{
					"wallet_id":    wid.String(),
					"days_left":    fmt.Sprintf("%d", daysAhead),
					"total_amount": fmt.Sprintf("%d", totalAmount),
				},
			)
			notified++
		}
	}

	cs.logger.Info("Bonus expiry notifications done", "notified", notified)
	return nil
}

// expiredHoldCleanup releases wallet holds that have passed their expiry time.
func (cs *CronScheduler) expiredHoldCleanup(ctx context.Context) error {
	holds, err := cs.walletRepo.GetExpiredHolds(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("get expired holds: %w", err)
	}

	released := 0
	for _, hold := range holds {
		if err := cs.walletSvc.ReleaseHold(ctx, hold.ID); err != nil {
			cs.logger.Error("Failed to release expired hold", "hold_id", hold.ID, "error", err)
			continue
		}
		released++
	}

	cs.logger.Info("Expired hold cleanup done", "total_expired", len(holds), "released", released)
	return nil
}

// getBonusExpiryDays returns the configured bonus expiry period in days.
func (cs *CronScheduler) getBonusExpiryDays() int {
	days := cs.cfg.Wallet.BonusExpiryDays
	if days <= 0 {
		return 180
	}
	return days
}
