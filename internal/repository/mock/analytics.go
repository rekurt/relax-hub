package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

// AnalyticsRepo is an in-memory mock implementation of repository.AnalyticsRepository.
type AnalyticsRepo struct {
	mu        sync.RWMutex
	views     []domain.BathhouseView
	snapshots map[string]*domain.AnalyticsSnapshot // key: bathhouse_id:date
}

func NewAnalyticsRepo() *AnalyticsRepo {
	return &AnalyticsRepo{
		views:     []domain.BathhouseView{},
		snapshots: make(map[string]*domain.AnalyticsSnapshot),
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
		id       uuid.UUID
		views    int64
		bookings int64
		revenue  int64
		rating   float64
	}

	bathMap := make(map[uuid.UUID]*bathStats)
	for _, snapshot := range r.snapshots {
		if bathMap[snapshot.BathhouseID] == nil {
			bathMap[snapshot.BathhouseID] = &bathStats{id: snapshot.BathhouseID}
		}
		bathMap[snapshot.BathhouseID].views += snapshot.Views
		bathMap[snapshot.BathhouseID].bookings += snapshot.Bookings
		bathMap[snapshot.BathhouseID].revenue += snapshot.Revenue
		bathMap[snapshot.BathhouseID].rating = snapshot.AvgRating
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
