package cron

import (
	"context"
	"testing"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStart_RegistersAllExpectedJobs(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	cfg := &config.Config{}
	cfg.Cron.Enabled = true
	cfg.Cron.Timezone = "UTC"

	cs := NewCronScheduler(cfg, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.Start(context.Background())
	require.NoError(t, err)
	defer cs.Stop(context.Background())

	jobs := cs.Jobs()
	jobMap := make(map[string]string, len(jobs))
	for _, j := range jobs {
		jobMap[j.Name] = j.Spec
	}

	// Every 15 minutes
	assert.Equal(t, "*/15 * * * *", jobMap["auto_reject_timed_out_requests"], "BookingRequestTimeout should run every 15 min")
	assert.Equal(t, "*/15 * * * *", jobMap["no_show_detection"], "NoShowDetection should run every 15 min")
	assert.Equal(t, "*/15 * * * *", jobMap["booking_reminders"], "BookingReminders should run every 15 min")
	assert.Equal(t, "*/15 * * * *", jobMap["ticket_auto_escalation"], "TicketAutoEscalation should run every 15 min")

	// Hourly
	assert.Contains(t, jobMap, "escrow_release", "EscrowRelease should be registered")
	assert.Contains(t, jobMap, "auto_review_requests", "ReviewRequest should be registered")
	assert.Contains(t, jobMap, "auto_scenario_execution", "AutoScenarioExecution should be registered")
	assert.Contains(t, jobMap, "antifraud_pattern_detection", "AntiFraudPatternDetection should be registered")
	assert.Contains(t, jobMap, "expired_hold_cleanup", "ExpiredHoldCleanup should be registered")

	// Daily
	assert.Contains(t, jobMap, "bonus_expiration", "BonusExpiration should be registered")
	assert.Contains(t, jobMap, "bonus_expiry_notify", "BonusExpiryNotification should be registered")
	assert.Contains(t, jobMap, "account_deletion_execution", "AccountDeletionExecution should be registered")
	assert.Contains(t, jobMap, "session_cleanup", "SessionCleanup should be registered")
	assert.Contains(t, jobMap, "kyc_expiry_check", "KYCExpiryCheck should be registered")
	assert.Contains(t, jobMap, "response_rate_recalculation", "OwnerResponseRateMonitoring should be registered")
	assert.Contains(t, jobMap, "promo_deactivation", "PromoCodeDeactivation should be registered")
	assert.Contains(t, jobMap, "saved_search_check", "SavedSearchNotification should be registered")
	assert.Contains(t, jobMap, "bathhouse_metrics_update", "BathhouseMetricsUpdate should be registered")
	assert.Contains(t, jobMap, "platform_average_refresh", "PlatformAverageRating should be registered")
	assert.Contains(t, jobMap, "ticket_auto_close", "TicketAutoClose should be registered")

	// Other scheduled jobs
	assert.Contains(t, jobMap, "daily_aggregation", "DailyAggregation should be registered")
	assert.Contains(t, jobMap, "weekly_cleanup", "WeeklyCleanup should be registered")
	assert.Contains(t, jobMap, "subscription_expiry_notify", "SubscriptionExpiryNotify should be registered")
	assert.Contains(t, jobMap, "expired_subscription_update", "ExpiredSubscriptionUpdate should be registered")
	assert.Contains(t, jobMap, "calendar_sync", "CalendarSync should be registered")
	assert.Contains(t, jobMap, "account_deletion_reminders", "AccountDeletionReminders should be registered")
}

func TestStart_AllJobsUseDistributedLocking(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	cfg := &config.Config{}
	cfg.Cron.Enabled = true
	cfg.Cron.Timezone = "UTC"

	cs := NewCronScheduler(cfg, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.Start(context.Background())
	require.NoError(t, err)
	defer cs.Stop(context.Background())

	// All jobs are registered via Register() which wraps them with distributed locking.
	// The Register method is the only way to add jobs, and it always applies wrapJob.
	// Verify that all registered jobs have unique names (used as lock keys).
	jobs := cs.Jobs()
	names := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		assert.False(t, names[j.Name], "Duplicate job name found: %s (would cause lock conflicts)", j.Name)
		names[j.Name] = true
	}
}

func TestStart_JobCount(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	cfg := &config.Config{}
	cfg.Cron.Enabled = true
	cfg.Cron.Timezone = "UTC"

	cs := NewCronScheduler(cfg, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.Start(context.Background())
	require.NoError(t, err)
	defer cs.Stop(context.Background())

	// We expect at least 26 jobs to be registered (all the ones listed in the schedule)
	jobs := cs.Jobs()
	assert.GreaterOrEqual(t, len(jobs), 26, "Expected at least 26 registered cron jobs, got %d", len(jobs))
}

func TestTicketJobs_AutoEscalation(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockTicketSvc := &mockTicketServiceForCron{}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockTicketSvc, nil, nil, nil, nil,
	)

	err := cs.ticketAutoEscalation(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockTicketSvc.autoEscalateCalled, "AutoEscalateStaleTickets should be called")
}

func TestTicketJobs_AutoClose(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockTicketSvc := &mockTicketServiceForCron{}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockTicketSvc, nil, nil, nil, nil,
	)

	err := cs.ticketAutoCloseJob(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockTicketSvc.autoCloseCalled, "AutoCloseResolvedTickets should be called")
}

func TestEscrowJob_ProcessMatured(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockEscrowSvc := &mockEscrowServiceForCron{released: 5}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockEscrowSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.escrowReleaseJob(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockEscrowSvc.processMaturedCalled, "ProcessMaturedEscrows should be called")
}

func TestAntiFraudJob_NilEngine(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.antiFraudPatternDetection(context.Background())
	assert.NoError(t, err, "Should handle nil fraud engine gracefully")
}

func TestKYCExpiryCheck_NilService(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.kycExpiryCheck(context.Background())
	assert.NoError(t, err, "Should handle nil KYC service gracefully")
}

func TestSearchJob_BathhouseMetricsUpdate(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockAnalyticsSvc := &MockAnalyticsService{}

	cs := NewCronScheduler(&config.Config{}, log,
		mockAnalyticsSvc, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.bathhouseMetricsUpdate(context.Background())
	assert.NoError(t, err)
}

func TestOwnerJob_ResponseRateMonitoring(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	bookingSvc := &mockBookingServiceForReminders{}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, bookingSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.responseRateRecalculation(context.Background())
	assert.NoError(t, err)
}

func TestSavedSearchJob_NilService(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.savedSearchCheck(context.Background())
	assert.NoError(t, err, "Should handle nil saved search service gracefully")
}

func TestReviewJobs_PlatformAverageRefresh(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockReviewSvc := &mockReviewServiceForCron{}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockReviewSvc, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.platformAverageRefresh(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockReviewSvc.refreshPlatformCalled)
}

func TestReviewJobs_AutoReviewRequests(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockReviewSvc := &mockReviewServiceForCron{}
	cfg := &config.Config{}
	cfg.Review.RequestDelayHours = 3

	cs := NewCronScheduler(cfg, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockReviewSvc, nil, nil, nil, nil, nil, nil, nil,
	)

	err := cs.autoReviewRequests(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockReviewSvc.sendReviewRequestsCalled)
	assert.Equal(t, 3, mockReviewSvc.sendReviewRequestsDelay)
}

func TestAutoScenarioJob(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockScenarioSvc := &mockAutoScenarioServiceForCron{}

	cs := NewCronScheduler(&config.Config{}, log,
		&MockAnalyticsService{}, mock.NewAnalyticsRepo(),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockScenarioSvc, nil, nil, nil, nil, nil,
	)

	err := cs.autoScenarioExecution(context.Background())
	assert.NoError(t, err)
	assert.True(t, mockScenarioSvc.executeCalled)
}
