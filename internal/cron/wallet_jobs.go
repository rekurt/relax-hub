package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/spf13/viper"
)

// handleBonusExpiration expires bonus transactions that have passed their expiry date.
// Runs daily, iterates all active wallets and delegates to WalletService.ExpireBonusesForWallet.
func (cs *CronScheduler) handleBonusExpiration() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting wallet bonus expiration")

	walletIDs, err := cs.walletRepo.ListAllIDs(ctx)
	if err != nil {
		cs.logger.Error("Failed to list wallet IDs for bonus expiration", "error", err, "duration", time.Since(start))
		return
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

	cs.logger.Info("Wallet bonus expiration completed",
		"wallets_processed", walletsProcessed,
		"total_expired", totalExpired,
		"duration", time.Since(start),
	)
}

// handleBonusExpiryNotify warns users about bonuses expiring soon (at 14 and 3 days before expiry).
func (cs *CronScheduler) handleBonusExpiryNotify() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting wallet bonus expiry notifications")

	walletIDs, err := cs.walletRepo.ListAllIDs(ctx)
	if err != nil {
		cs.logger.Error("Failed to list wallet IDs for bonus expiry notify", "error", err, "duration", time.Since(start))
		return
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

			expiryDays := getBonusExpiryDays()
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

	cs.logger.Info("Wallet bonus expiry notifications completed",
		"notified", notified,
		"duration", time.Since(start),
	)
}

// handleExpiredHoldCleanup releases wallet holds that have passed their expiry time.
func (cs *CronScheduler) handleExpiredHoldCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting expired hold cleanup")

	holds, err := cs.walletRepo.GetExpiredHolds(ctx, time.Now())
	if err != nil {
		cs.logger.Error("Failed to get expired holds", "error", err, "duration", time.Since(start))
		return
	}

	released := 0
	for _, hold := range holds {
		if err := cs.walletSvc.ReleaseHold(ctx, hold.ID); err != nil {
			cs.logger.Error("Failed to release expired hold", "hold_id", hold.ID, "error", err)
			continue
		}
		released++
	}

	cs.logger.Info("Expired hold cleanup completed",
		"total_expired", len(holds),
		"released", released,
		"duration", time.Since(start),
	)
}

// getBonusExpiryDays returns the configured bonus expiry period in days.
func getBonusExpiryDays() int {
	days := viper.GetInt("WALLET_BONUS_EXPIRY_DAYS")
	if days <= 0 {
		return 180
	}
	return days
}
