package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/nikitaaldaev/bani/internal/calendar"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

// CronScheduler manages all scheduled tasks
type CronScheduler struct {
	c                *cron.Cron
	logger           *logger.Logger
	analyticsService service.AnalyticsService
	analyticsRepo    repository.AnalyticsRepository
	subscriptionRepo repository.SubscriptionRepository
	promoRepo        repository.PromoCodeRepository
	notifSvc         service.NotificationService
	calendarSync     *calendar.CalendarSyncService
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(
	l *logger.Logger,
	svc service.AnalyticsService,
	repo repository.AnalyticsRepository,
	subscriptionRepo repository.SubscriptionRepository,
	promoRepo repository.PromoCodeRepository,
	notifSvc service.NotificationService,
	calendarSync *calendar.CalendarSyncService,
) *CronScheduler {
	return &CronScheduler{
		c:                cron.New(),
		logger:           l,
		analyticsService: svc,
		analyticsRepo:    repo,
		subscriptionRepo: subscriptionRepo,
		promoRepo:        promoRepo,
		notifSvc:         notifSvc,
		calendarSync:     calendarSync,
	}
}

// RegisterCron registers all scheduled tasks with fx lifecycle
func RegisterCron(lc fx.Lifecycle, cs *CronScheduler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return cs.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return cs.Stop(ctx)
		},
	})
}

// Start begins the cron scheduler and registers jobs
func (cs *CronScheduler) Start(ctx context.Context) error {
	// Daily aggregation at 02:00 UTC
	if _, err := cs.c.AddFunc("0 2 * * *", cs.handleDailyAggregation); err != nil {
		cs.logger.Error("Failed to register daily aggregation job", "error", err)
		return fmt.Errorf("failed to register daily aggregation: %w", err)
	}
	cs.logger.Info("Registered daily aggregation job at 02:00 UTC")

	// Weekly cleanup at 03:00 UTC on Sunday
	if _, err := cs.c.AddFunc("0 3 * * 0", cs.handleWeeklyCleanup); err != nil {
		cs.logger.Error("Failed to register weekly cleanup job", "error", err)
		return fmt.Errorf("failed to register weekly cleanup: %w", err)
	}
	cs.logger.Info("Registered weekly cleanup job at 03:00 UTC on Sunday")

	// Subscription expiry notification at 06:00 UTC daily
	if _, err := cs.c.AddFunc("0 6 * * *", cs.handleSubscriptionExpiryNotify); err != nil {
		cs.logger.Error("Failed to register subscription expiry notify job", "error", err)
		return fmt.Errorf("failed to register subscription expiry notify: %w", err)
	}
	cs.logger.Info("Registered subscription expiry notification job at 06:00 UTC")

	// Expired subscription update at 06:05 UTC daily
	if _, err := cs.c.AddFunc("5 6 * * *", cs.handleExpiredSubscriptionUpdate); err != nil {
		cs.logger.Error("Failed to register expired subscription update job", "error", err)
		return fmt.Errorf("failed to register expired subscription update: %w", err)
	}
	cs.logger.Info("Registered expired subscription update job at 06:05 UTC")

	// Promo code deactivation at 06:10 UTC daily
	if _, err := cs.c.AddFunc("10 6 * * *", cs.handlePromoDeactivation); err != nil {
		cs.logger.Error("Failed to register promo deactivation job", "error", err)
		return fmt.Errorf("failed to register promo deactivation: %w", err)
	}
	cs.logger.Info("Registered promo deactivation job at 06:10 UTC")

	// Calendar sync every 15 minutes
	if _, err := cs.c.AddFunc("*/15 * * * *", cs.handleCalendarSync); err != nil {
		cs.logger.Error("Failed to register calendar sync job", "error", err)
		return fmt.Errorf("failed to register calendar sync: %w", err)
	}
	cs.logger.Info("Registered calendar sync job every 15 minutes")

	cs.c.Start()
	cs.logger.Info("Cron scheduler started")
	return nil
}

// Stop gracefully stops the cron scheduler
func (cs *CronScheduler) Stop(ctx context.Context) error {
	cs.c.Stop()
	cs.logger.Info("Cron scheduler stopped")
	return nil
}

// handleDailyAggregation runs daily analytics aggregation
func (cs *CronScheduler) handleDailyAggregation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting daily analytics aggregation")

	if err := cs.analyticsService.AggregateDaily(ctx); err != nil {
		cs.logger.Error("Daily analytics aggregation failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Daily analytics aggregation completed", "duration", time.Since(start))
}

// handleWeeklyCleanup removes old bathhouse view records (older than 90 days)
func (cs *CronScheduler) handleWeeklyCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting weekly analytics cleanup")

	// Calculate the cutoff date (90 days ago)
	cutoffDate := time.Now().AddDate(0, 0, -90)

	deletedCount, err := cs.analyticsRepo.DeleteOldViews(ctx, cutoffDate)
	if err != nil {
		cs.logger.Error("Weekly analytics cleanup failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Weekly analytics cleanup completed", "deleted_count", deletedCount, "cutoff_date", cutoffDate, "duration", time.Since(start))
}
