package cron

import "context"

// savedSearchCheck checks for new bathhouses matching saved searches with notifications enabled.
func (cs *CronScheduler) savedSearchCheck(ctx context.Context) error {
	if cs.savedSearchSvc == nil {
		cs.logger.Warn("Saved search service not available, skipping")
		return nil
	}
	return cs.savedSearchSvc.CheckNewMatches(ctx)
}
