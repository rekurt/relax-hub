package cron

import (
	"context"
	"fmt"
)

// kycExpiryCheck marks approved KYC applications as expired when past their expiry date.
func (cs *CronScheduler) kycExpiryCheck(ctx context.Context) error {
	if cs.kycSvc == nil {
		cs.logger.Warn("KYC service not available, skipping expiry check")
		return nil
	}

	expired, err := cs.kycSvc.CheckExpiredApplications(ctx)
	if err != nil {
		return fmt.Errorf("kyc expiry check: %w", err)
	}
	cs.logger.Info("KYC expiry check done", "expired", expired)
	return nil
}
