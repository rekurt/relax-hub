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

// --- Advanced Analytics (FR-147-154) ---

// FunnelStep represents one step in the conversion funnel.
type FunnelStep struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"` // % of the first step
}

// ConversionFunnel represents the full conversion funnel.
type ConversionFunnel struct {
	Period AnalyticsPeriod `json:"period"`
	Steps  []FunnelStep    `json:"steps"`
}

// CohortRow represents a single cohort (users registered in a given month).
type CohortRow struct {
	CohortMonth    string    `json:"cohort_month"`    // "2026-01"
	UsersCount     int64     `json:"users_count"`     // users in cohort
	RetentionWeeks []float64 `json:"retention_weeks"` // retention % per week (0..N)
	TotalSpending  int64     `json:"total_spending"`  // kopecks
}

// CohortAnalysis holds all cohort rows.
type CohortAnalysis struct {
	Cohorts []CohortRow `json:"cohorts"`
}

// GeoSupplyDemand represents demand vs supply for a city.
type GeoSupplyDemand struct {
	CityID       int64  `json:"city_id"`
	CityName     string `json:"city_name"`
	SearchCount  int64  `json:"search_count"`  // demand (searches)
	ListingCount int64  `json:"listing_count"` // supply (active listings)
	BookingCount int64  `json:"booking_count"` // realized demand
}

// GeoDemandSupplyMap holds all city-level demand/supply data.
type GeoDemandSupplyMap struct {
	Period AnalyticsPeriod   `json:"period"`
	Cities []GeoSupplyDemand `json:"cities"`
}

// HeatmapCell represents one grid cell on the geographic heatmap.
type HeatmapCell struct {
	Latitude     float64 `json:"latitude"`      // center of the cell
	Longitude    float64 `json:"longitude"`     // center of the cell
	ListingCount int64   `json:"listing_count"` // supply (active bathhouses)
	BookingCount int64   `json:"booking_count"` // realized demand
	SearchCount  int64   `json:"search_count"`  // demand (search views)
}

// HeatmapData holds the full heatmap response.
type HeatmapData struct {
	Period   AnalyticsPeriod `json:"period"`
	CellSize float64         `json:"cell_size"` // grid cell size in degrees
	Cells    []HeatmapCell   `json:"cells"`
}

// WalletMetrics holds aggregate wallet statistics.
type WalletMetrics struct {
	TotalClientBalance int64   `json:"total_client_balance"` // kopecks
	TotalOwnerBalance  int64   `json:"total_owner_balance"`  // kopecks
	TotalEscrow        int64   `json:"total_escrow"`         // kopecks
	WalletPaymentShare float64 `json:"wallet_payment_share"` // % of bookings paid via wallet
	ExpiredBonusVolume int64   `json:"expired_bonus_volume"` // kopecks expired in period
	ActiveWallets      int64   `json:"active_wallets"`
}

// OwnerPerformance holds per-owner analytics (with optional anonymized competitor benchmarking).
type OwnerPerformance struct {
	BathhouseID           uuid.UUID `json:"bathhouse_id"`
	BathhouseName         string    `json:"bathhouse_name"`
	ConversionRate        float64   `json:"conversion_rate"` // views -> bookings
	OccupancyRate         float64   `json:"occupancy_rate"`  // booked hours / available hours
	AvgRating             float64   `json:"avg_rating"`
	Revenue               int64     `json:"revenue"`                  // kopecks
	AvgCityConversionRate float64   `json:"avg_city_conversion_rate"` // anonymous benchmark
	AvgCityOccupancyRate  float64   `json:"avg_city_occupancy_rate"`
	AvgCityRating         float64   `json:"avg_city_rating"`
}
