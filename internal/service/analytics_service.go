package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
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
	Period          domain.AnalyticsPeriod `json:"period"`
	Views           int64                  `json:"views"`
	UniqueViews     int64                  `json:"unique_views"`
	Bookings        int64                  `json:"bookings"`
	ConversionRate  float64                `json:"conversion_rate"` // bookings / views
	Revenue         int64                  `json:"revenue"`         // in kopecks
	AvgCheck        int64                  `json:"avg_check"`       // revenue / bookings
	Rating          float64                `json:"rating"`
	RatingChange    float64                `json:"rating_change"` // % change from previous period
	ViewsChange     float64                `json:"views_change"`
	BookingsChange  float64                `json:"bookings_change"`
	RevenueChange   float64                `json:"revenue_change"`
	PreviousPeriod  *OwnerDashboard        `json:"previous_period,omitempty"`
}

// AdminDashboard represents analytics data for admin
type AdminDashboard struct {
	Period              domain.AnalyticsPeriod `json:"period"`
	TotalUsers          int64                  `json:"total_users"`
	NewUsers            int64                  `json:"new_users"`
	TotalBathhouses     int64                  `json:"total_bathhouses"`
	TotalBookings       int64                  `json:"total_bookings"`
	TotalRevenue        int64                  `json:"total_revenue"`
	TotalViews          int64                  `json:"total_views"`
	AvgRating           float64                `json:"avg_rating"`
	DAU                 int64                  `json:"dau"`  // daily active users (from bookings)
	WAU                 int64                  `json:"wau"`  // weekly active users
	MAU                 int64                  `json:"mau"`  // monthly active users
	TopBathhouses       []TopBathhouseInfo     `json:"top_bathhouses"`
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
	AggregateDaily(ctx context.Context) error
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
	if !source.IsValid() {
		return fmt.Errorf("invalid view source: %s", source)
	}

	// Check IP-based deduplication: if same IP viewed same bathhouse in last 30 min, skip
	now := time.Now()
	// Create a time bucket for the 30-minute window (align to bucket boundaries)
	timeBucket := (now.Unix() / int64(ipDeduplicationWindow.Seconds())) * int64(ipDeduplicationWindow.Seconds())
	dedupeKey := fmt.Sprintf(cacheKeyIPDuplication, bathhouseID.String(), ipHash, fmt.Sprintf("%d", timeBucket))

	// Try to get from Redis (if this IP+bathhouse was seen recently)
	exists := s.redis.Exists(ctx, dedupeKey).Val()
	if exists > 0 {
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
	cached := s.redis.Get(ctx, cacheKey).Val()
	if cached != "" {
		var dashboard OwnerDashboard
		if err := json.Unmarshal([]byte(cached), &dashboard); err == nil {
			return &dashboard, nil
		}
	}

	// Calculate date range
	now := time.Now()
	periodDays := period.Days()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -periodDays)

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
		_ = s.redis.Set(ctx, cacheKey, data, cacheTTL).Err()
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
	cached := s.redis.Get(ctx, cacheKey).Val()
	if cached != "" {
		var dashboard AdminDashboard
		if err := json.Unmarshal([]byte(cached), &dashboard); err == nil {
			return &dashboard, nil
		}
	}

	// Calculate date range
	now := time.Now()
	periodDays := period.Days()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -periodDays)

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

	// Cache the result
	if data, err := json.Marshal(dashboard); err == nil {
		_ = s.redis.Set(ctx, cacheKey, data, cacheTTL).Err()
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
