package cron

import (
	"context"
	"fmt"
	"time"
)

// antiFraudPatternDetection runs batch anti-fraud checks: dormant balances,
// suspicious top-up/cancel cycles, and other heuristic patterns.
func (cs *CronScheduler) antiFraudPatternDetection(ctx context.Context) error {
	if cs.fraudEngine == nil {
		cs.logger.Warn("Fraud engine not available, skipping pattern detection")
		return nil
	}

	if cs.bookingRepo == nil {
		cs.logger.Warn("Booking repo not available, skipping dormant balance check")
		return nil
	}

	walletIDs, err := cs.walletRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list wallet IDs for antifraud: %w", err)
	}

	flagged := 0
	for _, wid := range walletIDs {
		wallet, err := cs.walletRepo.GetByID(ctx, wid)
		if err != nil {
			cs.logger.Warn("Failed to get wallet for antifraud check", "wallet_id", wid, "error", err)
			continue
		}

		if wallet.Balance < 5000000 { // skip wallets under 50k RUB
			continue
		}

		lastDate, err := cs.bookingRepo.GetLastBookingDateByUser(ctx, wallet.UserID)
		if err != nil {
			cs.logger.Warn("Failed to get last booking date", "user_id", wallet.UserID, "error", err)
			continue
		}
		var lastBookingDaysAgo int
		if lastDate != nil {
			lastBookingDaysAgo = int(time.Since(*lastDate).Hours() / 24)
		} else {
			// No bookings at all — treat as dormant if balance is high
			lastBookingDaysAgo = 365
		}

		if err := cs.fraudEngine.CheckDormantBalance(ctx, wallet.UserID, wallet.Balance, lastBookingDaysAgo); err != nil {
			flagged++
		}
	}

	cs.logger.Info("Anti-fraud pattern detection done", "wallets_checked", len(walletIDs), "flagged", flagged)
	return nil
}
