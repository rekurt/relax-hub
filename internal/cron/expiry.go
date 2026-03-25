package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

// subscriptionExpiryNotify notifies owners about subscriptions expiring within 3 days.
func (cs *CronScheduler) subscriptionExpiryNotify(ctx context.Context) error {
	now := time.Now()
	threshold := now.Add(3 * 24 * time.Hour)

	subs, err := cs.subscriptionRepo.GetExpiring(ctx, threshold)
	if err != nil {
		return fmt.Errorf("get expiring subscriptions: %w", err)
	}

	notified := 0
	for _, sub := range subs {
		if sub.EndDate != nil && sub.EndDate.Before(now) {
			continue
		}
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

	cs.logger.Info("Subscription expiry notify done", "notified", notified)
	return nil
}

// expiredSubscriptionUpdate marks expired subscriptions as expired and notifies owners.
func (cs *CronScheduler) expiredSubscriptionUpdate(ctx context.Context) error {
	now := time.Now()

	subs, err := cs.subscriptionRepo.GetExpiring(ctx, now)
	if err != nil {
		return fmt.Errorf("get expired subscriptions: %w", err)
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

	cs.logger.Info("Expired subscription update done", "updated", updated)
	return nil
}

// promoDeactivation deactivates promo codes that have passed their valid_until date.
func (cs *CronScheduler) promoDeactivation(ctx context.Context) error {
	count, err := cs.promoRepo.DeactivateExpired(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("deactivate expired promos: %w", err)
	}

	cs.logger.Info("Promo deactivation done", "deactivated", count)
	return nil
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
