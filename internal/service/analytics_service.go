package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
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
	DAU             int64                  `json:"dau"`        // daily active users (from bookings)
	WAU             int64                  `json:"wau"`        // weekly active users
	MAU             int64                  `json:"mau"`        // monthly active users
	ADR             int64                  `json:"adr"`        // average daily rate (avg booking value in kopecks)
	ChurnRate       float64                `json:"churn_rate"` // % of users with no bookings in last 90 days
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
	GMV                int64                  `json:"gmv"`                 // gross merchandise value (kopecks) — total booking value
	ServiceFeesTotal   int64                  `json:"service_fees_total"`  // platform service fees (kopecks)
	SubscriptionsTotal int64                  `json:"subscriptions_total"` // subscription revenue (kopecks)
	PromotionsTotal    int64                  `json:"promotions_total"`    // promotion revenue (kopecks)
	PlatformRevenue    int64                  `json:"platform_revenue"`    // total platform revenue (kopecks)
	TakeRate           float64                `json:"take_rate"`           // platform_revenue / GMV * 100
	TotalBookings      int64                  `json:"total_bookings"`      // booking count in period
	RevenuePerBooking  int64                  `json:"revenue_per_booking"` // platform revenue per booking (kopecks)
	GMVPerBooking      int64                  `json:"gmv_per_booking"`     // average booking value (kopecks)
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
