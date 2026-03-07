package domain

import (
	"time"

	"github.com/google/uuid"
)

// ViewSource represents the source of a bathhouse view.
type ViewSource string

const (
	ViewSourceSearch   ViewSource = "search"
	ViewSourceDirect   ViewSource = "direct"
	ViewSourceWidget   ViewSource = "widget"
	ViewSourceTelegram ViewSource = "telegram"
)

func (s ViewSource) IsValid() bool {
	switch s {
	case ViewSourceSearch, ViewSourceDirect, ViewSourceWidget, ViewSourceTelegram:
		return true
	}
	return false
}

// BathhouseView records a single view event for a bathhouse.
type BathhouseView struct {
	ID          uuid.UUID  `json:"id"`
	BathhouseID uuid.UUID  `json:"bathhouse_id"`
	ViewerID    *uuid.UUID `json:"viewer_id,omitempty"`
	Source      ViewSource `json:"source"`
	IPHash      string     `json:"ip_hash"`
	ViewedAt    time.Time  `json:"viewed_at"`
}

// AnalyticsSnapshot stores aggregated analytics data for a bathhouse per day.
type AnalyticsSnapshot struct {
	BathhouseID uuid.UUID `json:"bathhouse_id"`
	Date        time.Time `json:"date"`
	Views       int64     `json:"views"`
	UniqueViews int64     `json:"unique_views"`
	Bookings    int64     `json:"bookings"`
	Revenue     int64     `json:"revenue"`
	ReviewCount int       `json:"review_count"`
	AvgRating   float64   `json:"avg_rating"`
}

// AnalyticsPeriod represents a time period for analytics queries.
type AnalyticsPeriod string

const (
	PeriodDay   AnalyticsPeriod = "1d"
	PeriodWeek  AnalyticsPeriod = "7d"
	PeriodMonth AnalyticsPeriod = "30d"
	Period90d   AnalyticsPeriod = "90d"
)

func (p AnalyticsPeriod) IsValid() bool {
	switch p {
	case PeriodDay, PeriodWeek, PeriodMonth, Period90d:
		return true
	}
	return false
}

func (p AnalyticsPeriod) Days() int {
	switch p {
	case PeriodDay:
		return 1
	case PeriodWeek:
		return 7
	case PeriodMonth:
		return 30
	case Period90d:
		return 90
	default:
		return 30
	}
}

// TopMetric represents a metric for ranking bathhouses.
type TopMetric string

const (
	MetricViews    TopMetric = "views"
	MetricBookings TopMetric = "bookings"
	MetricRevenue  TopMetric = "revenue"
	MetricRating   TopMetric = "rating"
)

func (m TopMetric) IsValid() bool {
	switch m {
	case MetricViews, MetricBookings, MetricRevenue, MetricRating:
		return true
	}
	return false
}
