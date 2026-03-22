package cron

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAnalyticsService is a mock implementation for testing
type MockAnalyticsService struct {
	AggregateDailyFunc func(ctx context.Context) error
}

func (m *MockAnalyticsService) RecordView(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error {
	return nil
}

func (m *MockAnalyticsService) GetOwnerDashboard(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*service.OwnerDashboard, error) {
	return nil, nil
}

func (m *MockAnalyticsService) GetDailyStats(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error) {
	return nil, nil
}

func (m *MockAnalyticsService) GetAdminDashboard(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error) {
	return nil, nil
}

func (m *MockAnalyticsService) GetTopBathhousesByMetric(ctx context.Context, userRole domain.UserRole, metric domain.TopMetric, limit int64) ([]service.TopBathhouseInfo, error) {
	return nil, nil
}

func (m *MockAnalyticsService) AggregateDaily(ctx context.Context) error {
	if m.AggregateDailyFunc != nil {
		return m.AggregateDailyFunc(ctx)
	}
	return nil
}

func TestNewCronScheduler(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	assert.NotNil(t, cs)
	assert.NotNil(t, cs.c)
	assert.Equal(t, log, cs.logger)
	assert.Equal(t, mockSvc, cs.analyticsService)
	assert.Equal(t, mockRepo, cs.analyticsRepo)
}

func TestCronSchedulerStart(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	err := cs.Start(context.Background())
	assert.NoError(t, err)

	// Verify jobs are registered
	jobs := cs.c.Entries()
	assert.True(t, len(jobs) >= 7, "Expected at least 7 jobs registered")

	// Clean up
	err = cs.Stop(context.Background())
	assert.NoError(t, err)
}

func TestCronSchedulerStop(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	err := cs.Start(context.Background())
	assert.NoError(t, err)

	err = cs.Stop(context.Background())
	assert.NoError(t, err)
}

func TestHandleDailyAggregation(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	aggregationCalled := false

	mockSvc := &MockAnalyticsService{
		AggregateDailyFunc: func(ctx context.Context) error {
			aggregationCalled = true
			return nil
		},
	}
	mockRepo := mock.NewAnalyticsRepo()

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	cs.handleDailyAggregation()

	assert.True(t, aggregationCalled, "Expected AggregateDaily to be called")
}

func TestHandleDailyAggregationError(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	mockSvc := &MockAnalyticsService{
		AggregateDailyFunc: func(ctx context.Context) error {
			return errors.New("aggregation failed")
		},
	}
	mockRepo := mock.NewAnalyticsRepo()

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	// Should not panic even with error
	cs.handleDailyAggregation()
}

func TestHandleWeeklyCleanup(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()

	// Add some views to the mock repo
	oldDate := time.Now().AddDate(0, 0, -100)
	newDate := time.Now().AddDate(0, 0, -10)
	bathhouseID := uuid.New()

	err := mockRepo.RecordView(context.Background(), &domain.BathhouseView{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Source:      domain.ViewSourceSearch,
		IPHash:      "hash1",
		ViewedAt:    oldDate,
	})
	require.NoError(t, err)

	err = mockRepo.RecordView(context.Background(), &domain.BathhouseView{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Source:      domain.ViewSourceSearch,
		IPHash:      "hash2",
		ViewedAt:    newDate,
	})
	require.NoError(t, err)

	cs := NewCronScheduler(log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil)
	cs.handleWeeklyCleanup()

	// Verify old views are deleted by trying to delete again (should return 0)
	cutoffDate := time.Now().AddDate(0, 0, -90)
	deletedCount, err := mockRepo.DeleteOldViews(context.Background(), cutoffDate)
	assert.NoError(t, err)
	// After handleWeeklyCleanup, there should be no old views to delete
	assert.Equal(t, int64(0), deletedCount)
}
