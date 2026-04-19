package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func (s *analyticsService) GetConversionFunnel(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.ConversionFunnel, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	steps, err := s.analyticsRepo.GetConversionFunnel(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return &domain.ConversionFunnel{Period: period, Steps: steps}, nil
}

func (s *analyticsService) GetCohortAnalysis(ctx context.Context, userRole domain.UserRole, months int) (*domain.CohortAnalysis, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if months < 1 || months > 24 {
		months = 6
	}

	cohorts, err := s.analyticsRepo.GetCohortAnalysis(ctx, months)
	if err != nil {
		return nil, err
	}
	return &domain.CohortAnalysis{Cohorts: cohorts}, nil
}

func (s *analyticsService) GetGeoDemandSupply(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.GeoDemandSupplyMap, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	cities, err := s.analyticsRepo.GetGeoSupplyDemand(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return &domain.GeoDemandSupplyMap{Period: period, Cities: cities}, nil
}

func (s *analyticsService) GetWalletMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.WalletMetrics, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	return s.analyticsRepo.GetWalletMetrics(ctx, from, to)
}

func (s *analyticsService) GetOwnerPerformance(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*domain.OwnerPerformance, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	return s.analyticsRepo.GetOwnerPerformance(ctx, bathhouseID, from, to)
}

// GetBusinessMetrics returns detailed business metrics for admin analytics (FR-148, FR-149)
func (s *analyticsService) GetBusinessMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*BusinessMetrics, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	metrics := &BusinessMetrics{Period: period}

	stats, err := s.analyticsRepo.GetPlatformStats(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}

	if stats.Bookings > 0 {
		metrics.ADR = stats.Revenue / stats.Bookings
	}

	metrics.DAU, _, metrics.MAU = s.calculateUserActivity(ctx, from, to)

	churnRate, err := s.analyticsRepo.GetChurnRate(ctx, 90)
	if err != nil {
		s.logger.Warn("Failed to get churn rate", "error", err)
	} else {
		metrics.ChurnRate = churnRate
	}

	ltv, err := s.analyticsRepo.GetLTV(ctx)
	if err != nil {
		s.logger.Warn("Failed to get LTV", "error", err)
	} else {
		metrics.LTV = ltv
	}

	arpu, err := s.analyticsRepo.GetARPU(ctx, from, to)
	if err != nil {
		s.logger.Warn("Failed to get ARPU", "error", err)
	} else {
		metrics.ARPU = arpu
	}

	return metrics, nil
}

// GetPnL returns P&L and unit-economics metrics for admin (FR-150)
func (s *analyticsService) GetPnL(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*PnLMetrics, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	from, to := s.periodToRange(period)
	pnl := &PnLMetrics{Period: period}

	gmv, bookingCount, err := s.analyticsRepo.GetGMV(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get gmv: %w", err)
	}
	pnl.GMV = gmv
	pnl.TotalBookings = bookingCount

	serviceFees, subscriptions, promotions, err := s.analyticsRepo.GetPlatformRevenue(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get platform revenue: %w", err)
	}
	pnl.ServiceFeesTotal = serviceFees
	pnl.SubscriptionsTotal = subscriptions
	pnl.PromotionsTotal = promotions
	pnl.PlatformRevenue = serviceFees + subscriptions + promotions

	if pnl.GMV > 0 {
		pnl.TakeRate = float64(pnl.PlatformRevenue) / float64(pnl.GMV) * 100
	}

	if bookingCount > 0 {
		pnl.RevenuePerBooking = pnl.PlatformRevenue / bookingCount
		pnl.GMVPerBooking = gmv / bookingCount
	}

	return pnl, nil
}

func (s *analyticsService) UpdateBathhouseMetrics(ctx context.Context) (int, error) {
	allBhs, err := s.bhRepo.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 10000})
	if err != nil {
		return 0, fmt.Errorf("list bathhouses: %w", err)
	}

	now := time.Now()
	from := now.AddDate(0, 0, -30)
	updated := 0

	for _, bh := range allBhs.Items {
		stats, err := s.analyticsRepo.GetBathhouseStats(ctx, bh.ID, from, now)
		if err != nil {
			s.logger.Warn("Failed to get bathhouse stats for metrics", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		var conversionRate float64
		if stats != nil && stats.Views > 0 {
			conversionRate = float64(stats.Bookings) / float64(stats.Views)
		}

		bookingsList, err := s.bookingRepo.ListByBathhouse(ctx, bh.ID, 1, 1)
		if err != nil {
			s.logger.Warn("Failed to get bookings for occupancy", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		var occupancyRate float64
		if bookingsList != nil && bookingsList.TotalCount > 0 {
			totalAvailableHours := float64(30 * 12) // 30 days * 12 hours average
			occupancyRate = float64(bookingsList.TotalCount) / totalAvailableHours
			if occupancyRate > 1.0 {
				occupancyRate = 1.0
			}
		}

		if err := s.bhRepo.UpdateRankingFields(ctx, bh.ID, conversionRate, occupancyRate); err != nil {
			s.logger.Error("Failed to update ranking fields", "bathhouse_id", bh.ID, "error", err)
			continue
		}
		updated++
	}

	return updated, nil
}

// GetHeatmapData returns geographic heatmap data with supply/demand per grid cell (FR-153)
func (s *analyticsService) GetHeatmapData(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}
	if cellSize <= 0 || cellSize > 1.0 {
		cellSize = 0.01 // default ~1km grid
	}

	from, to := s.periodToRange(period)

	cacheKey := fmt.Sprintf("analytics:heatmap:%s:%.4f", period, cellSize)
	cachedResult := s.redis.Get(ctx, cacheKey)
	if cachedResult.Err() == nil && cachedResult.Val() != "" {
		var data domain.HeatmapData
		if err := json.Unmarshal([]byte(cachedResult.Val()), &data); err == nil {
			return &data, nil
		}
	}

	cells, err := s.analyticsRepo.GetHeatmapData(ctx, from, to, cellSize)
	if err != nil {
		return nil, err
	}

	result := &domain.HeatmapData{
		Period:   period,
		CellSize: cellSize,
		Cells:    cells,
	}

	if encoded, err := json.Marshal(result); err == nil {
		if err := s.redis.Set(ctx, cacheKey, encoded, cacheTTL).Err(); err != nil {
			s.logger.Warn("Failed to cache heatmap data", "error", err)
		}
	}

	return result, nil
}
