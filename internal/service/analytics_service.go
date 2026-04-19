package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	// cacheKeyOwnerDashboard is the prefix for owner dashboard cache
	cacheKeyOwnerDashboard = "analytics:owner:%s:%s"
	// cacheKeyAdminDashboard is the prefix for admin dashboard cache
	cacheKeyAdminDashboard = "analytics:admin:%s"
	// cacheKeyIPDuplication is the prefix for IP view deduplication
	cacheKeyIPDuplication = "analytics:view:%s:%s:%s" // bathhouse:ip:timestamp_bucket
	// cacheTTL is the cache duration for dashboards (15 minutes)
	cacheTTL = 15 * time.Minute
	// ipDeduplicationWindow is the time window for IP-based deduplication (30 minutes)
	ipDeduplicationWindow = 30 * time.Minute
)

// OwnerDashboard represents analytics data for a bathhouse owner
type OwnerDashboard struct {
	Period         domain.AnalyticsPeriod `json:"period"`
	Views          int64                  `json:"views"`
	UniqueViews    int64                  `json:"unique_views"`
	Bookings       int64                  `json:"bookings"`
	ConversionRate float64                `json:"conversion_rate"` // bookings / views
	Revenue        int64                  `json:"revenue"`         // in kopecks
	AvgCheck       int64                  `json:"avg_check"`       // revenue / bookings
	Rating         float64                `json:"rating"`
	RatingChange   float64                `json:"rating_change"` // % change from previous period
	ViewsChange    float64                `json:"views_change"`
	BookingsChange float64                `json:"bookings_change"`
	RevenueChange  float64                `json:"revenue_change"`
	PreviousPeriod *OwnerDashboard        `json:"previous_period,omitempty"`
}

// AdminDashboard represents analytics data for admin
type AdminDashboard struct {
	Period          domain.AnalyticsPeriod `json:"period"`
	TotalUsers      int64                  `json:"total_users"`
	NewUsers        int64                  `json:"new_users"`
	TotalBathhouses int64                  `json:"total_bathhouses"`
	TotalBookings   int64                  `json:"total_bookings"`
	TotalRevenue    int64                  `json:"total_revenue"`
	TotalViews      int64                  `json:"total_views"`
	AvgRating       float64                `json:"avg_rating"`
	DAU             int64                  `json:"dau"`              // daily active users (from bookings)
	WAU             int64                  `json:"wau"`              // weekly active users
	MAU             int64                  `json:"mau"`              // monthly active users
	ADR             int64                  `json:"adr"`              // average daily rate (avg booking value in kopecks)
	ChurnRate       float64                `json:"churn_rate"`       // % of users with no bookings in last 90 days
	TopBathhouses   []TopBathhouseInfo     `json:"top_bathhouses"`
}

// BusinessMetrics represents detailed business metrics for admin analytics (FR-148, FR-149)
type BusinessMetrics struct {
	Period    domain.AnalyticsPeriod `json:"period"`
	ADR       int64                  `json:"adr"`        // average booking value (kopecks)
	DAU       int64                  `json:"dau"`        // daily active users
	MAU       int64                  `json:"mau"`        // monthly active users
	ChurnRate float64                `json:"churn_rate"` // % users churned (90 days no activity)
	LTV       int64                  `json:"ltv"`        // average lifetime value per user (kopecks)
	ARPU      int64                  `json:"arpu"`       // average revenue per user in period (kopecks)
}

// PnLMetrics represents P&L and unit-economics for admin analytics (FR-150)
type PnLMetrics struct {
	Period             domain.AnalyticsPeriod `json:"period"`
	GMV                int64                  `json:"gmv"`                  // gross merchandise value (kopecks) — total booking value
	ServiceFeesTotal   int64                  `json:"service_fees_total"`   // platform service fees (kopecks)
	SubscriptionsTotal int64                  `json:"subscriptions_total"`  // subscription revenue (kopecks)
	PromotionsTotal    int64                  `json:"promotions_total"`     // promotion revenue (kopecks)
	PlatformRevenue    int64                  `json:"platform_revenue"`     // total platform revenue (kopecks)
	TakeRate           float64                `json:"take_rate"`            // platform_revenue / GMV * 100
	TotalBookings      int64                  `json:"total_bookings"`       // booking count in period
	RevenuePerBooking  int64                  `json:"revenue_per_booking"`  // platform revenue per booking (kopecks)
	GMVPerBooking      int64                  `json:"gmv_per_booking"`      // average booking value (kopecks)
}

// TopBathhouseInfo contains info about a top-ranked bathhouse
type TopBathhouseInfo struct {
	BathhouseID uuid.UUID `json:"bathhouse_id"`
	Name        string    `json:"name"`
	Views       int64     `json:"views"`
	Bookings    int64     `json:"bookings"`
	Revenue     int64     `json:"revenue"`
	Rating      float64   `json:"rating"`
}

// AnalyticsService defines the analytics business logic
type AnalyticsService interface {
	RecordView(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error
	GetOwnerDashboard(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*OwnerDashboard, error)
	GetDailyStats(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error)
	GetAdminDashboard(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*AdminDashboard, error)
	GetTopBathhousesByMetric(ctx context.Context, userRole domain.UserRole, metric domain.TopMetric, limit int64) ([]TopBathhouseInfo, error)
	AggregateDaily(ctx context.Context) error
	UpdateBathhouseMetrics(ctx context.Context) (int, error)

	// Advanced analytics (FR-147-154)
	GetConversionFunnel(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.ConversionFunnel, error)
	GetCohortAnalysis(ctx context.Context, userRole domain.UserRole, months int) (*domain.CohortAnalysis, error)
	GetGeoDemandSupply(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.GeoDemandSupplyMap, error)
	GetWalletMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.WalletMetrics, error)
	GetOwnerPerformance(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*domain.OwnerPerformance, error)

	// Business metrics (FR-148, FR-149)
	GetBusinessMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*BusinessMetrics, error)

	// P&L and unit economics (FR-150)
	GetPnL(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*PnLMetrics, error)

	// Heatmap (FR-153)
	GetHeatmapData(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	bhRepo        repository.BathhouseRepository
	bookingRepo   repository.BookingRepository
	userRepo      repository.UserRepository
	access        *AccessChecker
	redis         *redis.Client
	logger        *logger.Logger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	analyticsRepo repository.AnalyticsRepository,
	bhRepo repository.BathhouseRepository,
	bookingRepo repository.BookingRepository,
	userRepo repository.UserRepository,
	access *AccessChecker,
	redis *redis.Client,
	log *logger.Logger,
) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		bhRepo:        bhRepo,
		bookingRepo:   bookingRepo,
		userRepo:      userRepo,
		access:        access,
		redis:         redis,
		logger:        log,
	}
}

// RecordView records a bathhouse view with IP-based deduplication over 30 minutes
func (s *analyticsService) RecordView(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error {
	if bathhouseID == uuid.Nil {
		return fmt.Errorf("invalid bathhouse_id")
	}
	if !source.IsValid() {
		return fmt.Errorf("invalid view source: %s", source)
	}

	// Check IP-based deduplication: if same IP viewed same bathhouse in last 30 min, skip
	now := time.Now()
	// Create a time bucket for the 30-minute window (align to bucket boundaries)
	timeBucket := (now.Unix() / int64(ipDeduplicationWindow.Seconds())) * int64(ipDeduplicationWindow.Seconds())
	dedupeKey := fmt.Sprintf(cacheKeyIPDuplication, bathhouseID.String(), ipHash, fmt.Sprintf("%d", timeBucket))

	// Try to get from Redis (if this IP+bathhouse was seen recently)
	exists := s.redis.Exists(ctx, dedupeKey)
	if existsErr := exists.Err(); existsErr != nil && existsErr != redis.Nil {
		s.logger.Warn("Redis deduplication check failed", "error", existsErr)
	}
	if exists.Val() > 0 {
		// Already viewed recently, skip recording
		return nil
	}

	// Record the view in database
	view := &domain.BathhouseView{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		ViewerID:    viewerID,
		Source:      source,
		IPHash:      ipHash,
		ViewedAt:    now,
	}

	if err := s.analyticsRepo.RecordView(ctx, view); err != nil {
		s.logger.Error("Failed to record view", "bathhouse_id", bathhouseID, "error", err)
		return err
	}

	// Set the deduplication flag in Redis with 30-minute expiration
	if err := s.redis.Set(ctx, dedupeKey, "1", ipDeduplicationWindow).Err(); err != nil {
		// Log but don't fail - view was recorded in DB
		s.logger.Warn("Failed to set deduplication flag in Redis", "error", err)
	}

	return nil
}

// GetOwnerDashboard returns dashboard data for a bathhouse owner
func (s *analyticsService) GetOwnerDashboard(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*OwnerDashboard, error) {
	// RBAC check
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	if !period.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	// Check cache first
	cacheKey := fmt.Sprintf(cacheKeyOwnerDashboard, bathhouseID.String(), period)
	cachedResult := s.redis.Get(ctx, cacheKey)
	if cachedErr := cachedResult.Err(); cachedErr != nil && cachedErr != redis.Nil {
		s.logger.Warn("Redis cache check failed", "error", cachedErr)
	}
	cached := cachedResult.Val()
	if cached != "" {
		var dashboard OwnerDashboard
		if err := json.Unmarshal([]byte(cached), &dashboard); err == nil {
			return &dashboard, nil
		} else {
			s.logger.Debug("Failed to unmarshal cached dashboard", "error", err)
		}
	}

	// Calculate date range
	now := time.Now()
	periodDays := period.Days()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -periodDays)
	// Fix off-by-one: from should be at 00:00:00 of the start day
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	// Get current period stats
	stats, err := s.analyticsRepo.GetBathhouseStats(ctx, bathhouseID, from, to)
	if err != nil {
		return nil, err
	}

	// Get previous period stats for comparison
	prevTo := from.AddDate(0, 0, -1)
	prevFrom := prevTo.AddDate(0, 0, -periodDays)
	prevStats, err := s.analyticsRepo.GetBathhouseStats(ctx, bathhouseID, prevFrom, prevTo)
	if err != nil {
		s.logger.Warn("Failed to get previous period stats", "error", err)
		prevStats = &domain.AnalyticsSnapshot{}
	}

	dashboard := &OwnerDashboard{
		Period:      period,
		Views:       stats.Views,
		UniqueViews: stats.UniqueViews,
		Bookings:    stats.Bookings,
		Revenue:     stats.Revenue,
		Rating:      stats.AvgRating,
	}

	// Calculate conversion rate
	if stats.Views > 0 {
		dashboard.ConversionRate = float64(stats.Bookings) / float64(stats.Views)
	}

	// Calculate average check
	if stats.Bookings > 0 {
		dashboard.AvgCheck = stats.Revenue / stats.Bookings
	}

	// Calculate period-over-period changes
	if prevStats.Views > 0 {
		dashboard.ViewsChange = float64(stats.Views-prevStats.Views) / float64(prevStats.Views) * 100
	}
	if prevStats.Bookings > 0 {
		dashboard.BookingsChange = float64(stats.Bookings-prevStats.Bookings) / float64(prevStats.Bookings) * 100
	}
	if prevStats.Revenue > 0 {
		dashboard.RevenueChange = float64(stats.Revenue-prevStats.Revenue) / float64(prevStats.Revenue) * 100
	}
	if prevStats.AvgRating > 0 {
		dashboard.RatingChange = ((stats.AvgRating - prevStats.AvgRating) / prevStats.AvgRating) * 100
	}

	// Cache the result
	if data, err := json.Marshal(dashboard); err == nil {
		if err := s.redis.Set(ctx, cacheKey, data, cacheTTL).Err(); err != nil {
			s.logger.Warn("Failed to cache owner dashboard", "error", err)
		}
	}

	return dashboard, nil
}

// GetDailyStats returns daily analytics breakdown for a bathhouse owner
func (s *analyticsService) GetDailyStats(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error) {
	// RBAC check
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	// Get daily stats from repository
	dailyStats, err := s.analyticsRepo.GetDailyStats(ctx, bathhouseID, from, to)
	if err != nil {
		return nil, err
	}

	return dailyStats, nil
}

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
	// Fix off-by-one: from should be at 00:00:00 of the start day
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

			// Get stats for this bathhouse
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
	// Get all active bathhouses
	allBhs, err := s.bhRepo.List(ctx, domain.BathhouseFilter{Page: 1, PageSize: 10000})
	if err != nil {
		return fmt.Errorf("failed to list bathhouses: %w", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	dayDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())

	for _, bh := range allBhs.Items {
		// Get raw data aggregated from source tables for yesterday
		stats, err := s.analyticsRepo.AggregateRawData(ctx, bh.ID, dayDate)
		if err != nil {
			s.logger.Warn("Failed to aggregate raw data for bathhouse", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		// Create snapshot
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
	// RBAC check
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
	// DAU: last day of the period
	dayStart := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	dayEnd := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	if count, err := s.analyticsRepo.CountDistinctActiveUsers(ctx, dayStart, dayEnd); err == nil {
		dau = count
	}

	// WAU: last 7 days of the period
	weekStart := to.AddDate(0, 0, -7)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	if count, err := s.analyticsRepo.CountDistinctActiveUsers(ctx, weekStart, dayEnd); err == nil {
		wau = count
	}

	// MAU: entire period
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

// --- Advanced analytics (FR-147-154) ---

func (s *analyticsService) periodToRange(period domain.AnalyticsPeriod) (time.Time, time.Time) {
	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -period.Days())
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	return from, to
}

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

	// Get platform stats for ADR calculation
	stats, err := s.analyticsRepo.GetPlatformStats(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}

	// ADR: average booking value
	if stats.Bookings > 0 {
		metrics.ADR = stats.Revenue / stats.Bookings
	}

	// DAU and MAU
	metrics.DAU, _, metrics.MAU = s.calculateUserActivity(ctx, from, to)

	// Churn rate (90 days inactivity window)
	churnRate, err := s.analyticsRepo.GetChurnRate(ctx, 90)
	if err != nil {
		s.logger.Warn("Failed to get churn rate", "error", err)
	} else {
		metrics.ChurnRate = churnRate
	}

	// LTV
	ltv, err := s.analyticsRepo.GetLTV(ctx)
	if err != nil {
		s.logger.Warn("Failed to get LTV", "error", err)
	} else {
		metrics.LTV = ltv
	}

	// ARPU
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

	// GMV
	gmv, bookingCount, err := s.analyticsRepo.GetGMV(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get gmv: %w", err)
	}
	pnl.GMV = gmv
	pnl.TotalBookings = bookingCount

	// Platform revenue breakdown
	serviceFees, subscriptions, promotions, err := s.analyticsRepo.GetPlatformRevenue(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get platform revenue: %w", err)
	}
	pnl.ServiceFeesTotal = serviceFees
	pnl.SubscriptionsTotal = subscriptions
	pnl.PromotionsTotal = promotions
	pnl.PlatformRevenue = serviceFees + subscriptions + promotions

	// Take Rate
	if pnl.GMV > 0 {
		pnl.TakeRate = float64(pnl.PlatformRevenue) / float64(pnl.GMV) * 100
	}

	// Unit economics
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

// GetHeatmapData returns geographic heatmap data with supply/demand per grid cell (FR-153).
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
