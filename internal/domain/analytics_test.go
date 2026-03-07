package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestViewSourceIsValid(t *testing.T) {
	tests := []struct {
		name     string
		source   ViewSource
		expected bool
	}{
		{name: "search is valid", source: ViewSourceSearch, expected: true},
		{name: "direct is valid", source: ViewSourceDirect, expected: true},
		{name: "widget is valid", source: ViewSourceWidget, expected: true},
		{name: "telegram is valid", source: ViewSourceTelegram, expected: true},
		{name: "invalid source", source: ViewSource("unknown"), expected: false},
		{name: "empty source", source: ViewSource(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.source.IsValid())
		})
	}
}

func TestAnalyticsPeriodIsValid(t *testing.T) {
	tests := []struct {
		name     string
		period   AnalyticsPeriod
		expected bool
	}{
		{name: "1d is valid", period: PeriodDay, expected: true},
		{name: "7d is valid", period: PeriodWeek, expected: true},
		{name: "30d is valid", period: PeriodMonth, expected: true},
		{name: "90d is valid", period: Period90d, expected: true},
		{name: "invalid period", period: AnalyticsPeriod("365d"), expected: false},
		{name: "empty period", period: AnalyticsPeriod(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.period.IsValid())
		})
	}
}

func TestAnalyticsPeriodDays(t *testing.T) {
	tests := []struct {
		name     string
		period   AnalyticsPeriod
		expected int
	}{
		{name: "1d returns 1", period: PeriodDay, expected: 1},
		{name: "7d returns 7", period: PeriodWeek, expected: 7},
		{name: "30d returns 30", period: PeriodMonth, expected: 30},
		{name: "90d returns 90", period: Period90d, expected: 90},
		{name: "unknown defaults to 30", period: AnalyticsPeriod("unknown"), expected: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.period.Days())
		})
	}
}

func TestTopMetricIsValid(t *testing.T) {
	tests := []struct {
		name     string
		metric   TopMetric
		expected bool
	}{
		{name: "views is valid", metric: MetricViews, expected: true},
		{name: "bookings is valid", metric: MetricBookings, expected: true},
		{name: "revenue is valid", metric: MetricRevenue, expected: true},
		{name: "rating is valid", metric: MetricRating, expected: true},
		{name: "invalid metric", metric: TopMetric("unknown"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.metric.IsValid())
		})
	}
}

func TestBathhouseViewStruct(t *testing.T) {
	id := uuid.New()
	bathhouseID := uuid.New()
	viewerID := uuid.New()
	now := time.Now()

	view := BathhouseView{
		ID:          id,
		BathhouseID: bathhouseID,
		ViewerID:    &viewerID,
		Source:      ViewSourceSearch,
		IPHash:      "abc123hash",
		ViewedAt:    now,
	}

	assert.Equal(t, id, view.ID)
	assert.Equal(t, bathhouseID, view.BathhouseID)
	assert.Equal(t, &viewerID, view.ViewerID)
	assert.Equal(t, ViewSourceSearch, view.Source)
	assert.Equal(t, "abc123hash", view.IPHash)
	assert.Equal(t, now, view.ViewedAt)
}

func TestBathhouseViewAnonymous(t *testing.T) {
	view := BathhouseView{
		ID:          uuid.New(),
		BathhouseID: uuid.New(),
		ViewerID:    nil,
		Source:      ViewSourceDirect,
		IPHash:      "anonymous_hash",
		ViewedAt:    time.Now(),
	}

	assert.Nil(t, view.ViewerID)
}

func TestAnalyticsSnapshotStruct(t *testing.T) {
	bathhouseID := uuid.New()
	date := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)

	snapshot := AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        date,
		Views:       150,
		UniqueViews: 80,
		Bookings:    10,
		Revenue:     500000,
		ReviewCount: 3,
		AvgRating:   4.5,
	}

	assert.Equal(t, bathhouseID, snapshot.BathhouseID)
	assert.Equal(t, date, snapshot.Date)
	assert.Equal(t, int64(150), snapshot.Views)
	assert.Equal(t, int64(80), snapshot.UniqueViews)
	assert.Equal(t, int64(10), snapshot.Bookings)
	assert.Equal(t, int64(500000), snapshot.Revenue)
	assert.Equal(t, 3, snapshot.ReviewCount)
	assert.Equal(t, 4.5, snapshot.AvgRating)
}
