package cron

import (
	"context"
	"encoding/json"
	"fmt"
)

// autoPayoutJob checks owners with auto-payout threshold configured and initiates
// payouts when their available balance exceeds the threshold.
func (cs *CronScheduler) autoPayoutJob(ctx context.Context) error {
	if cs.payoutSvc == nil || cs.payoutRepo == nil {
		return nil
	}

	settings, err := cs.payoutRepo.ListActiveAutoPayoutSettings(ctx)
	if err != nil {
		return fmt.Errorf("list active auto-payout settings: %w", err)
	}

	if len(settings) == 0 {
		return nil
	}

	triggered := 0
	failed := 0

	for _, s := range settings {
		available, err := cs.payoutSvc.CalculateAvailableBalance(ctx, s.UserID)
		if err != nil {
			cs.logger.Error("Auto-payout: failed to calculate available balance",
				"user_id", s.UserID, "error", err)
			failed++
			continue
		}

		if available < s.Threshold {
			continue
		}

		// Get payment details to include in the payout request
		var bankDetails json.RawMessage
		if cs.paymentDetailsRepo != nil {
			pd, err := cs.paymentDetailsRepo.GetByUserID(ctx, s.UserID)
			if err != nil {
				cs.logger.Error("Auto-payout: failed to get payment details",
					"user_id", s.UserID, "error", err)
				failed++
				continue
			}
			bankDetails, _ = json.Marshal(pd)
		}

		payout, err := cs.payoutSvc.RequestPayout(ctx, s.UserID, available, bankDetails)
		if err != nil {
			cs.logger.Error("Auto-payout: failed to request payout",
				"user_id", s.UserID, "amount", available, "error", err)
			failed++
			continue
		}

		if err := cs.payoutSvc.ProcessPayout(ctx, payout.ID); err != nil {
			cs.logger.Error("Auto-payout: failed to process payout",
				"user_id", s.UserID, "payout_id", payout.ID, "error", err)
			failed++
			continue
		}

		triggered++
		cs.logger.Info("Auto-payout: payout processed",
			"user_id", s.UserID, "amount", available, "payout_id", payout.ID)
	}

	cs.logger.Info("Auto-payout job done",
		"eligible_users", len(settings),
		"triggered", triggered,
		"failed", failed,
	)
	return nil
}
