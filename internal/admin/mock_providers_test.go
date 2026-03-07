package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/admin/pages"
)

// mockDashProvider returns empty dashboard data for testing.
type mockDashProvider struct{}

func (m *mockDashProvider) GetDashboardData(_ context.Context) (*pages.DashboardData, error) {
	return &pages.DashboardData{GeneratedAt: time.Now()}, nil
}

// mockModProvider returns empty moderation data for testing.
type mockModProvider struct{}

func (m *mockModProvider) GetModerationData(_ context.Context, _ pages.ModerationFilter, _, _ int) (*pages.ModerationData, error) {
	return &pages.ModerationData{Page: 1, PageSize: 20, TotalPages: 1}, nil
}

func (m *mockModProvider) ApproveReview(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *mockModProvider) RejectReview(_ context.Context, _ uuid.UUID, _ []string) error {
	return nil
}
func (m *mockModProvider) BatchApproveReviews(_ context.Context, _ []uuid.UUID) (int, int) {
	return 0, 0
}
func (m *mockModProvider) BatchRejectReviews(_ context.Context, _ []uuid.UUID, _ []string) (int, int) {
	return 0, 0
}

// mockAnalyticsProvider returns empty analytics data for testing.
type mockAnalyticsProvider struct{}

func (m *mockAnalyticsProvider) GetAnalyticsData(_ context.Context, _ pages.AnalyticsFilter) (*pages.AnalyticsData, error) {
	return &pages.AnalyticsData{GeneratedAt: time.Now()}, nil
}

// mockHealthProvider returns empty health data for testing.
type mockHealthProvider struct{}

func (m *mockHealthProvider) GetHealthData(_ context.Context) (*pages.HealthData, error) {
	return &pages.HealthData{GeneratedAt: time.Now()}, nil
}
