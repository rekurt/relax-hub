package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestAnalyticsRepoRecordView(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	bathhouseID := uuid.New()
	viewerID := uuid.New()

	view := &domain.BathhouseView{
		BathhouseID: bathhouseID,
		ViewerID:    &viewerID,
		Source:      domain.ViewSourceSearch,
		IPHash:      "abc123",
		ViewedAt:    time.Now(),
	}

	err := repo.RecordView(ctx, view)
	if err != nil {
		t.Fatalf("RecordView failed: %v", err)
	}

	if view.ID == uuid.Nil {
		t.Error("RecordView should set view ID")
	}
}

func TestAnalyticsRepoCreateSnapshot(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	bathhouseID := uuid.New()
	snapshot := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        time.Now().Truncate(24 * time.Hour),
		Views:       100,
		UniqueViews: 50,
		Bookings:    5,
		Revenue:     50000,
		ReviewCount: 2,
		AvgRating:   4.5,
	}

	err := repo.CreateSnapshot(ctx, snapshot)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
}

func TestAnalyticsRepoGetBathhouseStats(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	bathhouseID := uuid.New()
	date := time.Now().Truncate(24 * time.Hour)

	snapshot := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        date,
		Views:       100,
		UniqueViews: 50,
		Bookings:    5,
		Revenue:     50000,
		ReviewCount: 2,
		AvgRating:   4.5,
	}

	_ = repo.CreateSnapshot(ctx, snapshot)

	stats, err := repo.GetBathhouseStats(ctx, bathhouseID, date, date)
	if err != nil {
		t.Fatalf("GetBathhouseStats failed: %v", err)
	}

	if stats.Views != 100 {
		t.Errorf("Expected views=100, got %d", stats.Views)
	}
	if stats.Bookings != 5 {
		t.Errorf("Expected bookings=5, got %d", stats.Bookings)
	}
}

func TestAnalyticsRepoGetDailyStats(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	bathhouseID := uuid.New()
	date1 := time.Now().Truncate(24 * time.Hour)
	date2 := date1.AddDate(0, 0, 1)

	snapshot1 := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        date1,
		Views:       100,
		UniqueViews: 50,
		Bookings:    5,
		Revenue:     50000,
		ReviewCount: 2,
		AvgRating:   4.5,
	}

	snapshot2 := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        date2,
		Views:       80,
		UniqueViews: 40,
		Bookings:    3,
		Revenue:     30000,
		ReviewCount: 1,
		AvgRating:   4.0,
	}

	_ = repo.CreateSnapshot(ctx, snapshot1)
	_ = repo.CreateSnapshot(ctx, snapshot2)

	stats, err := repo.GetDailyStats(ctx, bathhouseID, date1, date2)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(stats))
	}
}

func TestAnalyticsRepoGetPlatformStats(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	date := time.Now().Truncate(24 * time.Hour)

	snapshot1 := &domain.AnalyticsSnapshot{
		BathhouseID: uuid.New(),
		Date:        date,
		Views:       100,
		UniqueViews: 50,
		Bookings:    5,
		Revenue:     50000,
		ReviewCount: 2,
		AvgRating:   4.5,
	}

	snapshot2 := &domain.AnalyticsSnapshot{
		BathhouseID: uuid.New(),
		Date:        date,
		Views:       80,
		UniqueViews: 40,
		Bookings:    3,
		Revenue:     30000,
		ReviewCount: 1,
		AvgRating:   4.0,
	}

	_ = repo.CreateSnapshot(ctx, snapshot1)
	_ = repo.CreateSnapshot(ctx, snapshot2)

	stats, err := repo.GetPlatformStats(ctx, date, date)
	if err != nil {
		t.Fatalf("GetPlatformStats failed: %v", err)
	}

	if stats.Views != 180 {
		t.Errorf("Expected total views=180, got %d", stats.Views)
	}
}

func TestAnalyticsRepoGetTopBathhouses(t *testing.T) {
	repo := NewAnalyticsRepo()
	ctx := context.Background()

	date := time.Now().Truncate(24 * time.Hour)

	snapshots := []domain.AnalyticsSnapshot{
		{
			BathhouseID: uuid.New(),
			Date:        date,
			Views:       100,
			Bookings:    5,
			Revenue:     50000,
			AvgRating:   4.5,
		},
		{
			BathhouseID: uuid.New(),
			Date:        date,
			Views:       80,
			Bookings:    3,
			Revenue:     30000,
			AvgRating:   4.0,
		},
	}

	for i := range snapshots {
		_ = repo.CreateSnapshot(ctx, &snapshots[i])
	}

	ids, err := repo.GetTopBathhouses(ctx, domain.MetricViews, 10)
	if err != nil {
		t.Fatalf("GetTopBathhouses failed: %v", err)
	}

	if len(ids) == 0 {
		t.Error("Expected at least one bathhouse")
	}
}
