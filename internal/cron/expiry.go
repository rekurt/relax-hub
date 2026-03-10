package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

// handleSubscriptionExpiryNotify notifies owners about subscriptions expiring within 3 days.
func (cs *CronScheduler) handleSubscriptionExpiryNotify() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting subscription expiry notification check")

	now := time.Now()
	threshold := now.Add(3 * 24 * time.Hour) // 3 days from now

	subs, err := cs.subscriptionRepo.GetExpiring(ctx, threshold)
	if err != nil {
		cs.logger.Error("Failed to get expiring subscriptions", "error", err, "duration", time.Since(start))
		return
	}

	notified := 0
	for _, sub := range subs {
		// Skip already expired subscriptions — they'll be handled by handleExpiredSubscriptionUpdate
		if sub.EndDate != nil && sub.EndDate.Before(now) {
			continue
		}
		// Skip non-active subscriptions
		if sub.Status != domain.SubscriptionActive {
			continue
		}

		daysLeft := 0
		if sub.EndDate != nil {
			daysLeft = int(sub.EndDate.Sub(now).Hours() / 24)
		}

		err := cs.notifSvc.Send(ctx, sub.OwnerID, domain.NotifSubscriptionExpiring,
			"Подписка скоро истекает",
			"Ваша подписка истекает через "+formatDays(daysLeft)+". Продлите подписку, чтобы сохранить преимущества.",
			map[string]string{
				"subscription_id": sub.ID.String(),
				"bathhouse_id":    sub.BathhouseID.String(),
			},
		)
		if err != nil {
			cs.logger.Warn("Failed to send subscription expiry notification", "owner_id", sub.OwnerID, "error", err)
			continue
		}
		notified++
	}

	cs.logger.Info("Subscription expiry notification check completed", "notified", notified, "duration", time.Since(start))
}

// handleExpiredSubscriptionUpdate marks expired subscriptions as expired and notifies owners.
func (cs *CronScheduler) handleExpiredSubscriptionUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting expired subscription update")

	now := time.Now()

	subs, err := cs.subscriptionRepo.GetExpiring(ctx, now)
	if err != nil {
		cs.logger.Error("Failed to get expired subscriptions", "error", err, "duration", time.Since(start))
		return
	}

	updated := 0
	for _, sub := range subs {
		if sub.Status != domain.SubscriptionActive {
			continue
		}
		if sub.EndDate == nil || sub.EndDate.After(now) {
			continue
		}

		sub.Status = domain.SubscriptionExpired
		if err := cs.subscriptionRepo.Update(ctx, &sub); err != nil {
			cs.logger.Error("Failed to update expired subscription", "subscription_id", sub.ID, "error", err)
			continue
		}

		_ = cs.notifSvc.Send(ctx, sub.OwnerID, domain.NotifSubscriptionExpired,
			"Подписка истекла",
			"Ваша подписка истекла. Продлите подписку для возобновления преимуществ.",
			map[string]string{
				"subscription_id": sub.ID.String(),
				"bathhouse_id":    sub.BathhouseID.String(),
			},
		)
		updated++
	}

	cs.logger.Info("Expired subscription update completed", "updated", updated, "duration", time.Since(start))
}

// handlePromoDeactivation deactivates promo codes that have passed their valid_until date.
func (cs *CronScheduler) handlePromoDeactivation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting promo code deactivation")

	count, err := cs.promoRepo.DeactivateExpired(ctx, time.Now())
	if err != nil {
		cs.logger.Error("Promo code deactivation failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Promo code deactivation completed", "deactivated", count, "duration", time.Since(start))
}

func formatDays(days int) string {
	if days <= 0 {
		return "менее суток"
	}
	if days == 1 {
		return "1 день"
	}
	if days >= 2 && days <= 4 {
		return fmt.Sprintf("%d дня", days)
	}
	return fmt.Sprintf("%d дней", days)
}
