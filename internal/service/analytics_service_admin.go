package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/redis/go-redis/v9"
)

// GetAdminDashboard returns dashboard data for admin
func (s *analyticsService) GetAdminDashboard(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*AdminDashboard, error) {
	// Only admins can view platform analytics
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	// Check cache
	cacheKey := fmt.Sprintf(cacheKeyAdminDashboard, period)
	cachedResult := s.redis.Get(ctx, cacheKey)
	if cachedErr := cachedResult.Err(); cachedErr != nil && cachedErr != redis.Nil {
		s.logger.Warn("Redis admin dashboard cache check failed", "error", cachedErr)
	}
	cached := cachedResult.Val()
	if cached != "" {
		var dashboard AdminDashboard
		if err := json.Unmarshal([]byte(cached), &dashboard); err == nil {
			return &dashboard, nil
		} else {
			s.logger.Debug("Failed to unmarshal cached admin dashboard", "error", err)
		}
	}

	// Calculate date range
	now := time.Now()
	periodDays := period.Days()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -periodDays)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	// Get platform stats
	stats, err := s.analyticsRepo.GetPlatformStats(ctx, from, to)
	if err != nil {
		return nil, err
	}

	dashboard := &AdminDashboard{
		Period:        period,
		TotalViews:    stats.Views,
		TotalBookings: stats.Bookings,
		TotalRevenue:  stats.Revenue,
		AvgRating:     stats.AvgRating,
	}

	// Get top bathhouses by bookings (default metric)
	topIDs, err := s.analyticsRepo.GetTopBathhouses(ctx, domain.MetricBookings, 10)
	if err != nil {
		s.logger.Warn("Failed to get top bathhouses", "error", err)
	} else {
		dashboard.TopBathhouses = make([]TopBathhouseInfo, 0, len(topIDs))
		for _, id := range topIDs {
			bh, err := s.bhRepo.GetByID(ctx, id)
			if err != nil {
				s.logger.Warn("Failed to get bathhouse", "id", id, "error", err)
				continue
			}

			bhStats, err := s.analyticsRepo.GetBathhouseStats(ctx, id, from, to)
			if err != nil {
				s.logger.Warn("Failed to get bathhouse stats", "id", id, "error", err)
				continue
			}

			dashboard.TopBathhouses = append(dashboard.TopBathhouses, TopBathhouseInfo{
				BathhouseID: id,
				Name:        bh.Name,
				Views:       bhStats.Views,
				Bookings:    bhStats.Bookings,
				Revenue:     bhStats.Revenue,
				Rating:      bhStats.AvgRating,
			})
		}
	}

	// Get user and bathhouse counts
	allUsers, err := s.userRepo.List(ctx, 1, 1)
	if err == nil && allUsers != nil {
		dashboard.TotalUsers = int64(allUsers.TotalCount)
	}

	allBathhouses, err := s.bhRepo.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 1})
	if err == nil && allBathhouses != nil {
		dashboard.TotalBathhouses = int64(allBathhouses.TotalCount)
	}

	// Get user activity metrics
	dashboard.DAU, dashboard.WAU, dashboard.MAU = s.calculateUserActivity(ctx, from, to)

	// Get new users for the period
	dashboard.NewUsers = s.countNewUsers(ctx, from, to)

	// ADR: average booking value
	if dashboard.TotalBookings > 0 {
		dashboard.ADR = dashboard.TotalRevenue / dashboard.TotalBookings
	}

	// Churn rate: % of users active before 90 days ago who had no bookings in last 90 days
	churnRate, err := s.analyticsRepo.GetChurnRate(ctx, 90)
	if err != nil {
		s.logger.Warn("Failed to get churn rate", "error", err)
	} else {
		dashboard.ChurnRate = churnRate
	}

	// Cache the result
	if data, err := json.Marshal(dashboard); err == nil {
		if err := s.redis.Set(ctx, cacheKey, data, cacheTTL).Err(); err != nil {
			s.logger.Warn("Failed to cache admin dashboard", "error", err)
		}
	}

	return dashboard, nil
}

// AggregateDaily aggregates analytics data for the previous day into snapshots
func (s *analyticsService) AggregateDaily(ctx context.Context) error {
	allBhs, err := s.bhRepo.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 10000})
	if err != nil {
		return fmt.Errorf("failed to list bathhouses: %w", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	dayDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	for _, bh := range allBhs.Items {
		stats, err := s.analyticsRepo.AggregateRawData(ctx, bh.ID, dayDate)
		if err != nil {
			s.logger.Warn("Failed to aggregate raw data for bathhouse", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		snapshot := &domain.AnalyticsSnapshot{
			BathhouseID: bh.ID,
			Date:        dayDate,
			Views:       stats.Views,
			UniqueViews: stats.UniqueViews,
			Bookings:    stats.Bookings,
			Revenue:     stats.Revenue,
			ReviewCount: stats.ReviewCount,
			AvgRating:   stats.AvgRating,
		}

		if err := s.analyticsRepo.CreateSnapshot(ctx, snapshot); err != nil {
			s.logger.Error("Failed to create analytics snapshot", "bathhouse_id", bh.ID, "date", dayDate, "error", err)
			continue
		}
	}

	s.logger.Info("Daily analytics aggregation completed", "date", dayDate)
	return nil
}

// GetTopBathhousesByMetric returns top bathhouses ranked by the specified metric
func (s *analyticsService) GetTopBathhousesByMetric(ctx context.Context, userRole domain.UserRole, metric domain.TopMetric, limit int64) ([]TopBathhouseInfo, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	topIDs, err := s.analyticsRepo.GetTopBathhouses(ctx, metric, int(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get top bathhouses: %w", err)
	}

	topBathhouses := make([]TopBathhouseInfo, 0, len(topIDs))
	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -30)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	for _, id := range topIDs {
		bh, err := s.bhRepo.GetByID(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to get bathhouse", "id", id, "error", err)
			continue
		}

		bhStats, err := s.analyticsRepo.GetBathhouseStats(ctx, id, from, to)
		if err != nil {
			s.logger.Warn("Failed to get bathhouse stats", "id", id, "error", err)
			continue
		}

		topBathhouses = append(topBathhouses, TopBathhouseInfo{
			BathhouseID: id,
			Name:        bh.Name,
			Views:       bhStats.Views,
			Bookings:    bhStats.Bookings,
			Revenue:     bhStats.Revenue,
			Rating:      bhStats.AvgRating,
		})
	}

	return topBathhouses, nil
}

// calculateUserActivity returns daily, weekly, and monthly active users
func (s *analyticsService) calculateUserActivity(ctx context.Context, from, to time.Time) (dau, wau, mau int64) {
	dayStart := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	dayEnd := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	if count, err := s.analyticsRepo.CountDistinctActiveUsers(ctx, dayStart, dayEnd); err == nil {
		dau = count
	}

	weekStart := to.AddDate(0, 0, -7)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	if count, err := s.analyticsRepo.CountDistinctActiveUsers(ctx, weekStart, dayEnd); err == nil {
		wau = count
	}

	fromStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toEnd := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	if count, err := s.analyticsRepo.CountDistinctActiveUsers(ctx, fromStart, toEnd); err == nil {
		mau = count
	}

	return dau, wau, mau
}

// countNewUsers returns the count of users created in the given period
func (s *analyticsService) countNewUsers(ctx context.Context, from, to time.Time) int64 {
	count, err := s.userRepo.CountByCreatedAtRange(ctx, from, to)
	if err != nil {
		s.logger.Warn("Failed to count new users", "error", err)
		return 0
	}
	return count
}

// periodToRange converts an AnalyticsPeriod to a (from, to) time range ending today
func (s *analyticsService) periodToRange(period domain.AnalyticsPeriod) (time.Time, time.Time) {
	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -period.Days())
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	return from, to
}
