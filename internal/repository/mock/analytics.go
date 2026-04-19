package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// AnalyticsRepo is an in-memory mock implementation of repository.AnalyticsRepository.
type AnalyticsRepo struct {
	mu        sync.RWMutex
	views     []domain.BathhouseView
	snapshots map[string]*domain.AnalyticsSnapshot // key: bathhouse_id:date

	// Advanced analytics mock data
	FunnelSteps    []domain.FunnelStep
	CohortRows     []domain.CohortRow
	GeoDemand      []domain.GeoSupplyDemand
	WalletMetrics  *domain.WalletMetrics
	OwnerPerfData  map[uuid.UUID]*domain.OwnerPerformance
	HeatmapCells   []domain.HeatmapCell
}

func NewAnalyticsRepo() *AnalyticsRepo {
	return &AnalyticsRepo{
		views:         []domain.BathhouseView{},
		snapshots:     make(map[string]*domain.AnalyticsSnapshot),
		OwnerPerfData: make(map[uuid.UUID]*domain.OwnerPerformance),
	}
}

func (r *AnalyticsRepo) RecordView(ctx context.Context, view *domain.BathhouseView) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if view.ID == uuid.Nil {
		view.ID = uuid.New()
	}
	if view.ViewedAt.IsZero() {
		view.ViewedAt = time.Now()
	}

	cp := *view
	r.views = append(r.views, cp)
	return nil
}

func (r *AnalyticsRepo) GetBathhouseStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) (*domain.AnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        from,
	}

	// Aggregate snapshots within the date range
	for _, snapshot := range r.snapshots {
		if snapshot.BathhouseID == bathhouseID && !snapshot.Date.Before(from) && !snapshot.Date.After(to) {
			result.Views += snapshot.Views
			result.UniqueViews += snapshot.UniqueViews
			result.Bookings += snapshot.Bookings
			result.Revenue += snapshot.Revenue
			result.ReviewCount += snapshot.ReviewCount
			// Aggregate weighted average rating
			if snapshot.ReviewCount > 0 {
				if result.ReviewCount-snapshot.ReviewCount > 0 {
					oldCount := result.ReviewCount - snapshot.ReviewCount
					newCount := result.ReviewCount
					result.AvgRating = (result.AvgRating*float64(oldCount) + snapshot.AvgRating*float64(snapshot.ReviewCount)) / float64(newCount)
				} else {
					result.AvgRating = snapshot.AvgRating
				}
			}
		}
	}

	return result, nil
}

func (r *AnalyticsRepo) GetDailyStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.AnalyticsSnapshot

	// Collect snapshots within the date range
	for _, snapshot := range r.snapshots {
		if snapshot.BathhouseID == bathhouseID && !snapshot.Date.Before(from) && !snapshot.Date.After(to) {
			cp := *snapshot
			results = append(results, cp)
		}
	}

	return results, nil
}

func (r *AnalyticsRepo) GetPlatformStats(ctx context.Context, from, to time.Time) (*domain.AnalyticsSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := &domain.AnalyticsSnapshot{
		Date: from,
	}

	// Aggregate all snapshots within the date range
	for _, snapshot := range r.snapshots {
		if !snapshot.Date.Before(from) && !snapshot.Date.After(to) {
			result.Views += snapshot.Views
			result.UniqueViews += snapshot.UniqueViews
			result.Bookings += snapshot.Bookings
			result.Revenue += snapshot.Revenue
			result.ReviewCount += snapshot.ReviewCount
			// Aggregate weighted average rating
			if snapshot.ReviewCount > 0 {
				if result.ReviewCount-snapshot.ReviewCount > 0 {
					oldCount := result.ReviewCount - snapshot.ReviewCount
					newCount := result.ReviewCount
					result.AvgRating = (result.AvgRating*float64(oldCount) + snapshot.AvgRating*float64(snapshot.ReviewCount)) / float64(newCount)
				} else {
					result.AvgRating = snapshot.AvgRating
				}
			}
		}
	}

	return result, nil
}

func (r *AnalyticsRepo) GetTopBathhouses(ctx context.Context, metric domain.TopMetric, limit int) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !metric.IsValid() {
		return nil, domain.ErrInvalidInput
	}

	if limit < 1 {
		limit = 10
	}

	// Build a map of bathhouse aggregates
	type bathStats struct {
		id          uuid.UUID
		views       int64
		bookings    int64
		revenue     int64
		rating      float64
		reviewCount int
	}

	bathMap := make(map[uuid.UUID]*bathStats)
	for _, snapshot := range r.snapshots {
		if bathMap[snapshot.BathhouseID] == nil {
			bathMap[snapshot.BathhouseID] = &bathStats{id: snapshot.BathhouseID}
		}
		bath := bathMap[snapshot.BathhouseID]
		bath.views += snapshot.Views
		bath.bookings += snapshot.Bookings
		bath.revenue += snapshot.Revenue
		// Aggregate weighted average rating
		if snapshot.ReviewCount > 0 {
			if bath.reviewCount > 0 {
				oldCount := bath.reviewCount
				newCount := oldCount + snapshot.ReviewCount
				bath.rating = (bath.rating*float64(oldCount) + snapshot.AvgRating*float64(snapshot.ReviewCount)) / float64(newCount)
				bath.reviewCount = newCount
			} else {
				bath.rating = snapshot.AvgRating
				bath.reviewCount = snapshot.ReviewCount
			}
		}
	}

	// Sort by metric
	var stats []*bathStats
	for _, s := range bathMap {
		stats = append(stats, s)
	}

	// Simple sort - in real usage, this would be a proper sort.
	// For now, we'll return just the IDs in order
	var result []uuid.UUID
	for _, s := range stats {
		if len(result) >= limit {
			break
		}
		result = append(result, s.id)
	}

	return result, nil
}

// AggregateRawData aggregates analytics data from mock views/bookings/reviews for a specific date
func (r *AnalyticsRepo) AggregateRawData(ctx context.Context, bathhouseID uuid.UUID, date time.Time) (*domain.AnalyticsSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Normalize date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	snapshot := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        startOfDay,
	}

	// Count views and unique views for the day
	uniqueIPs := make(map[string]bool)
	for _, view := range r.views {
		if view.BathhouseID == bathhouseID && view.ViewedAt.After(startOfDay) && view.ViewedAt.Before(endOfDay) {
			snapshot.Views++
			uniqueIPs[view.IPHash] = true
		}
	}
	snapshot.UniqueViews = int64(len(uniqueIPs))

	// For mock: assume bookings and reviews are tracked separately
	// In real implementation, these would come from booking and review repos
	// For now, return what we have in snapshots
	for _, s := range r.snapshots {
		if s.BathhouseID == bathhouseID && s.Date.Equal(startOfDay) {
			snapshot.Bookings = s.Bookings
			snapshot.Revenue = s.Revenue
			snapshot.ReviewCount = s.ReviewCount
			snapshot.AvgRating = s.AvgRating
			break
		}
	}

	return snapshot, nil
}

func (r *AnalyticsRepo) CreateSnapshot(ctx context.Context, snapshot *domain.AnalyticsSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := snapshot.BathhouseID.String() + ":" + snapshot.Date.Format("2006-01-02")
	cp := *snapshot
	r.snapshots[key] = &cp
	return nil
}

// DeleteOldViews removes bathhouse view records older than the specified date
func (r *AnalyticsRepo) DeleteOldViews(ctx context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var count int64
	var newViews []domain.BathhouseView

	for _, view := range r.views {
		if view.ViewedAt.Before(before) {
			count++
		} else {
			newViews = append(newViews, view)
		}
	}

	r.views = newViews
	return count, nil
}

// --- Advanced analytics mock methods ---

func (r *AnalyticsRepo) GetConversionFunnel(_ context.Context, _, _ time.Time) ([]domain.FunnelStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.FunnelSteps != nil {
		return r.FunnelSteps, nil
	}
	return []domain.FunnelStep{
		{Name: "visit", Count: 1000, Percentage: 100},
		{Name: "search", Count: 600, Percentage: 60},
		{Name: "view_card", Count: 300, Percentage: 30},
		{Name: "start_booking", Count: 100, Percentage: 10},
		{Name: "pay", Count: 80, Percentage: 8},
		{Name: "complete_visit", Count: 70, Percentage: 7},
	}, nil
}

func (r *AnalyticsRepo) GetCohortAnalysis(_ context.Context, months int) ([]domain.CohortRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.CohortRows != nil {
		return r.CohortRows, nil
	}
	// Return sample cohort data
	return []domain.CohortRow{
		{CohortMonth: "2026-01", UsersCount: 100, RetentionWeeks: []float64{100, 60, 40, 30}, TotalSpending: 5000000},
		{CohortMonth: "2026-02", UsersCount: 120, RetentionWeeks: []float64{100, 55, 35}, TotalSpending: 4500000},
	}, nil
}

func (r *AnalyticsRepo) GetGeoSupplyDemand(_ context.Context, _, _ time.Time) ([]domain.GeoSupplyDemand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.GeoDemand != nil {
		return r.GeoDemand, nil
	}
	return []domain.GeoSupplyDemand{
		{CityID: 1, CityName: "Москва", SearchCount: 5000, ListingCount: 100, BookingCount: 800},
		{CityID: 2, CityName: "Санкт-Петербург", SearchCount: 3000, ListingCount: 60, BookingCount: 500},
	}, nil
}

func (r *AnalyticsRepo) GetWalletMetrics(_ context.Context, _, _ time.Time) (*domain.WalletMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.WalletMetrics != nil {
		return r.WalletMetrics, nil
	}
	return &domain.WalletMetrics{
		TotalClientBalance: 15000000,
		TotalOwnerBalance:  25000000,
		TotalEscrow:        5000000,
		WalletPaymentShare: 25.5,
		ExpiredBonusVolume: 1200000,
		ActiveWallets:      850,
	}, nil
}

func (r *AnalyticsRepo) GetOwnerPerformance(_ context.Context, bathhouseID uuid.UUID, _, _ time.Time) (*domain.OwnerPerformance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if perf, ok := r.OwnerPerfData[bathhouseID]; ok {
		return perf, nil
	}
	return &domain.OwnerPerformance{
		BathhouseID:           bathhouseID,
		BathhouseName:         "Test",
		ConversionRate:        0.08,
		OccupancyRate:         0.65,
		AvgRating:             4.3,
		Revenue:               5000000,
		AvgCityConversionRate: 0.06,
		AvgCityOccupancyRate:  0.55,
		AvgCityRating:         4.1,
	}, nil
}

// --- Heatmap mock method (FR-153) ---

func (r *AnalyticsRepo) GetHeatmapData(_ context.Context, _, _ time.Time, cellSize float64) ([]domain.HeatmapCell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.HeatmapCells != nil {
		return r.HeatmapCells, nil
	}
	return []domain.HeatmapCell{
		{Latitude: 55.75, Longitude: 37.62, ListingCount: 30, BookingCount: 120, SearchCount: 500},
		{Latitude: 55.76, Longitude: 37.63, ListingCount: 15, BookingCount: 60, SearchCount: 250},
		{Latitude: 59.93, Longitude: 30.32, ListingCount: 20, BookingCount: 80, SearchCount: 350},
	}, nil
}

// --- Business metrics mock methods (FR-148, FR-149) ---

// ChurnRateOverride allows tests to set a custom churn rate value
var ChurnRateDefault float64 = 12.5

func (r *AnalyticsRepo) GetChurnRate(_ context.Context, inactiveDays int) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return ChurnRateDefault, nil
}

func (r *AnalyticsRepo) GetLTV(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Default mock LTV: 15000 rub = 1500000 kopecks
	return 1500000, nil
}

func (r *AnalyticsRepo) GetARPU(_ context.Context, _, _ time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Default mock ARPU: 3000 rub = 300000 kopecks
	return 300000, nil
}

// --- P&L metrics mock methods (FR-150) ---

func (r *AnalyticsRepo) GetGMV(_ context.Context, _, _ time.Time) (int64, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Default mock GMV: 500,000 rub = 50_000_000 kopecks, 200 bookings
	return 50000000, 200, nil
}

func (r *AnalyticsRepo) GetPlatformRevenue(_ context.Context, _, _ time.Time) (int64, int64, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Default mock: service_fees=5M kop, subscriptions=1M kop, promotions=500K kop
	return 5000000, 1000000, 500000, nil
}

func (r *AnalyticsRepo) CountDistinctActiveUsers(_ context.Context, _, _ time.Time) (int64, error) {
	return 42, nil
}
