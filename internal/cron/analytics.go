package cron

import (
	"context"
	"time"

	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/antifraud"
	"github.com/rekurt/relax-hub/internal/calendar"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

// CronScheduler manages all scheduled tasks
type CronScheduler struct {
	c                  *cron.Cron
	cfg                *config.Config
	logger             *logger.Logger
	jobs               []RegisteredJob
	analyticsService   service.AnalyticsService
	analyticsRepo      repository.AnalyticsRepository
	subscriptionRepo   repository.SubscriptionRepository
	promoRepo          repository.PromoCodeRepository
	notifSvc           service.NotificationService
	calendarSync       *calendar.CalendarSyncService
	walletSvc          service.WalletService
	walletRepo         repository.WalletRepository
	accountDeletionSvc service.AccountDeletionService
	sessionSvc         service.SessionService
	savedSearchSvc     service.SavedSearchService
	bookingSvc         service.BookingService
	escrowSvc          service.EscrowService
	depositSvc         service.SecurityDepositService
	reviewSvc          service.ReviewService
	clientReviewSvc    service.ClientReviewService
	autoScenarioSvc    service.AutoScenarioService
	ticketSvc          service.TicketService
	kycSvc             service.KYCService
	fraudEngine        antifraud.FraudEngine
	bookingRepo        repository.BookingRepository
	redisClient        *redis.Client
	reconciliationSvc      service.ReconciliationService
	adminNotificationSvc   service.AdminNotificationService
	adminNotificationRepo  repository.AdminNotificationRepository
	pmsSvc                 service.PMSService
	promotionSvc           service.PromotionService
	modReqSvc              service.BookingModificationService
	extReqSvc              service.BookingExtensionService
	payoutSvc              service.PayoutService
	payoutRepo             repository.PayoutRepository
	paymentDetailsRepo     repository.PaymentDetailsRepository
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(
	cfg *config.Config,
	l *logger.Logger,
	svc service.AnalyticsService,
	repo repository.AnalyticsRepository,
	subscriptionRepo repository.SubscriptionRepository,
	promoRepo repository.PromoCodeRepository,
	notifSvc service.NotificationService,
	calendarSync *calendar.CalendarSyncService,
	walletSvc service.WalletService,
	walletRepo repository.WalletRepository,
	accountDeletionSvc service.AccountDeletionService,
	sessionSvc service.SessionService,
	savedSearchSvc service.SavedSearchService,
	bookingSvc service.BookingService,
	escrowSvc service.EscrowService,
	depositSvc service.SecurityDepositService,
	reviewSvc service.ReviewService,
	clientReviewSvc service.ClientReviewService,
	autoScenarioSvc service.AutoScenarioService,
	ticketSvc service.TicketService,
	kycSvc service.KYCService,
	fraudEngine antifraud.FraudEngine,
	bookingRepo repository.BookingRepository,
	redisClient *redis.Client,
	reconciliationSvc service.ReconciliationService,
	adminNotificationSvc service.AdminNotificationService,
	adminNotificationRepo repository.AdminNotificationRepository,
	pmsSvc service.PMSService,
	promotionSvc service.PromotionService,
	modReqSvc service.BookingModificationService,
	extReqSvc service.BookingExtensionService,
	payoutSvc service.PayoutService,
	payoutRepo repository.PayoutRepository,
	paymentDetailsRepo repository.PaymentDetailsRepository,
) *CronScheduler {
	timezone := cfg.Cron.Timezone
	if timezone == "" {
		timezone = "Europe/Moscow"
	}

	return &CronScheduler{
		c:                  newCronWithTimezone(timezone, l),
		cfg:                cfg,
		logger:             l,
		analyticsService:   svc,
		analyticsRepo:      repo,
		subscriptionRepo:   subscriptionRepo,
		promoRepo:          promoRepo,
		notifSvc:           notifSvc,
		calendarSync:       calendarSync,
		walletSvc:          walletSvc,
		walletRepo:         walletRepo,
		accountDeletionSvc: accountDeletionSvc,
		sessionSvc:         sessionSvc,
		savedSearchSvc:     savedSearchSvc,
		bookingSvc:         bookingSvc,
		escrowSvc:          escrowSvc,
		depositSvc:         depositSvc,
		reviewSvc:          reviewSvc,
		clientReviewSvc:    clientReviewSvc,
		autoScenarioSvc:    autoScenarioSvc,
		ticketSvc:          ticketSvc,
		kycSvc:             kycSvc,
		fraudEngine:        fraudEngine,
		bookingRepo:        bookingRepo,
		redisClient:        redisClient,
		reconciliationSvc:      reconciliationSvc,
		adminNotificationSvc:   adminNotificationSvc,
		adminNotificationRepo:  adminNotificationRepo,
		pmsSvc:                 pmsSvc,
		promotionSvc:           promotionSvc,
		modReqSvc:              modReqSvc,
		extReqSvc:              extReqSvc,
		payoutSvc:              payoutSvc,
		payoutRepo:             payoutRepo,
		paymentDetailsRepo:     paymentDetailsRepo,
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
	if !cs.cfg.Cron.Enabled {
		cs.logger.Info("Cron scheduler is disabled via config")
		return nil
	}

	// Register all jobs using the new Register method with automatic
	// panic recovery, structured logging, and distributed locking
	jobs := []struct {
		spec string
		name string
		fn   JobFunc
	}{
		{"0 2 * * *", "daily_aggregation", cs.dailyAggregation},
		{"0 3 * * 0", "weekly_cleanup", cs.weeklyCleanup},
		{"0 6 * * *", "subscription_expiry_notify", cs.subscriptionExpiryNotify},
		{"5 6 * * *", "expired_subscription_update", cs.expiredSubscriptionUpdate},
		{"10 6 * * *", "promo_deactivation", cs.promoDeactivation},
		{"*/15 * * * *", "calendar_sync", cs.calendarSyncJob},
		{"0 4 * * *", "bonus_expiration", cs.bonusExpiration},
		{"0 5 * * *", "bonus_expiry_notify", cs.bonusExpiryNotify},
		{"0 * * * *", "expired_hold_cleanup", cs.expiredHoldCleanup},
		{"30 3 * * *", "account_deletion_execution", cs.accountDeletionExecution},
		{"0 7 * * *", "account_deletion_reminders", cs.accountDeletionReminders},
		{"15 3 * * *", "session_cleanup", cs.sessionCleanupJob},
		{"0 8 * * *", "saved_search_check", cs.savedSearchCheck},
		{"*/15 * * * *", "auto_reject_timed_out_requests", cs.autoRejectTimedOutRequests},
		{"*/15 * * * *", "no_show_detection", cs.noShowDetection},
		{"30 * * * *", "escrow_release", cs.escrowReleaseJob},
		{"45 * * * *", "deposit_release", cs.depositReleaseJob},
		{"*/15 * * * *", "booking_reminders", cs.bookingRemindersJob},
		{"0 3 * * *", "response_rate_recalculation", cs.responseRateRecalculation},
		{"30 2 * * *", "platform_average_refresh", cs.platformAverageRefresh},
		{"45 * * * *", "auto_review_requests", cs.autoReviewRequests},
		{"15 * * * *", "auto_scenario_execution", cs.autoScenarioExecution},
		{"*/15 * * * *", "ticket_auto_escalation", cs.ticketAutoEscalation},
		{"30 4 * * *", "ticket_auto_close", cs.ticketAutoCloseJob},
		{"0 * * * *", "antifraud_pattern_detection", cs.antiFraudPatternDetection},
		{"45 3 * * *", "kyc_expiry_check", cs.kycExpiryCheck},
		{"0 6 * * *", "bathhouse_metrics_update", cs.bathhouseMetricsUpdate},
		{"0 * * * *", "review_auto_reveal", cs.reviewAutoReveal},
		{"0 1 * * *", "daily_float_snapshot", cs.dailyFloatSnapshot},
		{"30 1 * * *", "daily_reconciliation", cs.dailyReconciliation},
		{"0 9 * * *", "admin_notification_digest", cs.adminNotificationDigest},
		{"0 5 * * 0", "admin_notification_cleanup", cs.adminNotificationCleanup},
		{"*/15 * * * *", "pms_sync", cs.pmsSyncJob},
		{"0 7 * * *", "promotion_daily_budget", cs.promotionDailyBudget},
		{"*/15 * * * *", "modification_request_expiry", cs.modificationRequestExpiry},
		{"*/5 * * * *", "extension_request_expiry", cs.extensionRequestExpiry},
		{"0 * * * *", "auto_payout", cs.autoPayoutJob},
	}

	for _, j := range jobs {
		if err := cs.Register(j.spec, j.name, j.fn); err != nil {
			return err
		}
	}

	cs.c.Start()
	cs.logger.Info("Cron scheduler started", "timezone", cs.cfg.Cron.Timezone, "jobs", len(cs.jobs))
	return nil
}

// Stop gracefully stops the cron scheduler
func (cs *CronScheduler) Stop(ctx context.Context) error {
	stopCtx := cs.c.Stop()
	select {
	case <-stopCtx.Done():
		cs.logger.Info("Cron scheduler stopped gracefully")
	case <-ctx.Done():
		cs.logger.Warn("Cron scheduler stop timed out")
	}
	return nil
}

// dailyAggregation runs daily analytics aggregation
func (cs *CronScheduler) dailyAggregation(ctx context.Context) error {
	return cs.analyticsService.AggregateDaily(ctx)
}

// weeklyCleanup removes old bathhouse view records (older than 90 days)
func (cs *CronScheduler) weeklyCleanup(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -90)
	deletedCount, err := cs.analyticsRepo.DeleteOldViews(ctx, cutoffDate)
	if err != nil {
		return err
	}
	cs.logger.Info("Weekly cleanup done", "deleted_count", deletedCount, "cutoff_date", cutoffDate)
	return nil
}
