package pages

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
)

var analyticsFuncMap = template.FuncMap{
	"formatRubles": FormatKopecksToRubles,
	"itoa": func(n int64) string {
		return strconv.FormatInt(n, 10)
	},
	"trendClass":  TrendClass,
	"trendArrow":  TrendArrow,
	"formatTrend": FormatTrend,
}

var analyticsTmpl = ParsePageTemplate(analyticsFuncMap, "templates/analytics.tmpl")

// ChartPoint represents a single data point for time-series charts.
type ChartPoint struct {
	Label string
	Value int64
}

// RankedItem represents a top-N item (e.g. bathhouse by bookings).
type RankedItem struct {
	Name  string
	Value int64
}

// DistributionItem represents a slice in a pie chart.
type DistributionItem struct {
	Label string
	Value int64
}

// CityOption represents a city for the filter dropdown.
type CityOption struct {
	ID   int64
	Name string
}

// AnalyticsFilter holds the current filter state.
type AnalyticsFilter struct {
	DateFrom string
	DateTo   string
	CityID   string
}

// DatePreset represents a quick date range selection button.
type DatePreset struct {
	Label    string
	DateFrom string
	DateTo   string
	Active   bool
}

// AnalyticsSummary holds aggregate totals for the selected period.
type AnalyticsSummary struct {
	TotalBookings int64
	TotalRevenue  int64
	TotalNewUsers int64

	BookingsTrend TrendData
	RevenueTrend  TrendData
	UsersTrend    TrendData
}

// AnalyticsData is the full data model for the analytics dashboard page.
type AnalyticsData struct {
	BookingsPerDay     []ChartPoint
	RevenuePerDay      []ChartPoint
	NewUsersPerDay     []ChartPoint
	TopByBookings      []RankedItem
	TopByRevenue       []RankedItem
	BookingStatusDist  []DistributionItem
	ReviewRatingDist   []DistributionItem
	Filter             AnalyticsFilter
	Cities             []CityOption
	DatePresets        []DatePreset
	Summary            AnalyticsSummary
	GeneratedAt        time.Time
	PagesPrefix        string
	AdminPrefix        string
	PageTitle          string
	ActivePage         string
}

// AnalyticsDataProvider fetches analytics data from a data source.
type AnalyticsDataProvider interface {
	GetAnalyticsData(ctx context.Context, filter AnalyticsFilter) (*AnalyticsData, error)
}

// PostgresAnalyticsProvider fetches analytics data from PostgreSQL.
type PostgresAnalyticsProvider struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewPostgresAnalyticsProvider creates a new PostgresAnalyticsProvider.
func NewPostgresAnalyticsProvider(pool *pgxpool.Pool, log *logger.Logger) *PostgresAnalyticsProvider {
	return &PostgresAnalyticsProvider{pool: pool, log: log}
}

func (p *PostgresAnalyticsProvider) GetAnalyticsData(ctx context.Context, filter AnalyticsFilter) (*AnalyticsData, error) {
	data := &AnalyticsData{
		Filter:      filter,
		GeneratedAt: time.Now(),
	}

	if err := p.loadCities(ctx, &data.Cities); err != nil {
		return nil, err
	}

	dateFrom, dateTo := p.parseDateRange(filter)

	if err := p.loadBookingsPerDay(ctx, dateFrom, dateTo, filter.CityID, &data.BookingsPerDay); err != nil {
		return nil, err
	}
	if err := p.loadRevenuePerDay(ctx, dateFrom, dateTo, filter.CityID, &data.RevenuePerDay); err != nil {
		return nil, err
	}
	if err := p.loadNewUsersPerDay(ctx, dateFrom, dateTo, filter.CityID, &data.NewUsersPerDay); err != nil {
		return nil, err
	}
	if err := p.loadTopByBookings(ctx, dateFrom, dateTo, filter.CityID, &data.TopByBookings); err != nil {
		return nil, err
	}
	if err := p.loadTopByRevenue(ctx, dateFrom, dateTo, filter.CityID, &data.TopByRevenue); err != nil {
		return nil, err
	}
	if err := p.loadBookingStatusDist(ctx, dateFrom, dateTo, filter.CityID, &data.BookingStatusDist); err != nil {
		return nil, err
	}
	if err := p.loadReviewRatingDist(ctx, dateFrom, dateTo, filter.CityID, &data.ReviewRatingDist); err != nil {
		return nil, err
	}
	if err := p.loadSummary(ctx, dateFrom, dateTo, filter.CityID, &data.Summary); err != nil {
		return nil, err
	}

	data.DatePresets = BuildDatePresets(time.Now(), filter)

	return data, nil
}

func (p *PostgresAnalyticsProvider) parseDateRange(filter AnalyticsFilter) (time.Time, time.Time) {
	now := time.Now()
	dateTo := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	dateFrom := dateTo.AddDate(0, 0, -29)
	dateFrom = time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, now.Location())

	if filter.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", filter.DateFrom); err == nil {
			dateFrom = t
		}
	}
	if filter.DateTo != "" {
		if t, err := time.Parse("2006-01-02", filter.DateTo); err == nil {
			dateTo = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		}
	}

	return dateFrom, dateTo
}

func (p *PostgresAnalyticsProvider) loadCities(ctx context.Context, cities *[]CityOption) error {
	rows, err := p.pool.Query(ctx, "SELECT id, name FROM cities ORDER BY name")
	if err != nil {
		p.log.Error("analytics: load cities", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var c CityOption
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return err
		}
		*cities = append(*cities, c)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadBookingsPerDay(ctx context.Context, from, to time.Time, cityID string, points *[]ChartPoint) error {
	cityFilter := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityFilter = " AND b.bathhouse_id IN (SELECT id FROM bathhouses WHERE city_id = $3)"
		args = append(args, cityID)
	}

	query := `
		SELECT d::date AS day, COUNT(b.id)
		FROM generate_series($1::date, $2::date, '1 day') d
		LEFT JOIN bookings b ON b.created_at::date = d::date` + cityFilter + `
		GROUP BY day ORDER BY day`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: bookings per day", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cp ChartPoint
		var day time.Time
		if err := rows.Scan(&day, &cp.Value); err != nil {
			return err
		}
		cp.Label = day.Format("02.01")
		*points = append(*points, cp)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadRevenuePerDay(ctx context.Context, from, to time.Time, cityID string, points *[]ChartPoint) error {
	cityFilter := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityFilter = " AND b.bathhouse_id IN (SELECT id FROM bathhouses WHERE city_id = $3)"
		args = append(args, cityID)
	}

	query := `
		SELECT d::date AS day, COALESCE(SUM(b.total_price), 0)
		FROM generate_series($1::date, $2::date, '1 day') d
		LEFT JOIN bookings b ON b.created_at::date = d::date` + cityFilter + `
		GROUP BY day ORDER BY day`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: revenue per day", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cp ChartPoint
		var day time.Time
		if err := rows.Scan(&day, &cp.Value); err != nil {
			return err
		}
		cp.Label = day.Format("02.01")
		*points = append(*points, cp)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadNewUsersPerDay(ctx context.Context, from, to time.Time, cityID string, points *[]ChartPoint) error {
	cityFilter := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityFilter = " AND u.id IN (SELECT DISTINCT b2.user_id FROM bookings b2 JOIN bathhouses bh2 ON bh2.id = b2.bathhouse_id WHERE bh2.city_id = $3)"
		args = append(args, cityID)
	}

	query := `
		SELECT d::date AS day, COUNT(u.id)
		FROM generate_series($1::date, $2::date, '1 day') d
		LEFT JOIN users u ON u.created_at::date = d::date` + cityFilter + `
		GROUP BY day ORDER BY day`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: new users per day", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cp ChartPoint
		var day time.Time
		if err := rows.Scan(&day, &cp.Value); err != nil {
			return err
		}
		cp.Label = day.Format("02.01")
		*points = append(*points, cp)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadTopByBookings(ctx context.Context, from, to time.Time, cityID string, items *[]RankedItem) error {
	cityWhere := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityWhere = " AND bh.city_id = $3"
		args = append(args, cityID)
	}

	query := `
		SELECT bh.name, COUNT(b.id)
		FROM bookings b
		JOIN bathhouses bh ON bh.id = b.bathhouse_id
		WHERE b.created_at >= $1 AND b.created_at <= $2` + cityWhere + `
		GROUP BY bh.name
		ORDER BY COUNT(b.id) DESC
		LIMIT 10`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: top by bookings", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ri RankedItem
		if err := rows.Scan(&ri.Name, &ri.Value); err != nil {
			return err
		}
		*items = append(*items, ri)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadTopByRevenue(ctx context.Context, from, to time.Time, cityID string, items *[]RankedItem) error {
	cityWhere := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityWhere = " AND bh.city_id = $3"
		args = append(args, cityID)
	}

	query := `
		SELECT bh.name, COALESCE(SUM(b.total_price), 0)
		FROM bookings b
		JOIN bathhouses bh ON bh.id = b.bathhouse_id
		WHERE b.created_at >= $1 AND b.created_at <= $2` + cityWhere + `
		GROUP BY bh.name
		ORDER BY SUM(b.total_price) DESC
		LIMIT 10`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: top by revenue", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ri RankedItem
		if err := rows.Scan(&ri.Name, &ri.Value); err != nil {
			return err
		}
		*items = append(*items, ri)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadBookingStatusDist(ctx context.Context, from, to time.Time, cityID string, items *[]DistributionItem) error {
	cityJoin := ""
	cityWhere := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityJoin = " JOIN bathhouses bh ON bh.id = b.bathhouse_id"
		cityWhere = " AND bh.city_id = $3"
		args = append(args, cityID)
	}

	query := `
		SELECT b.status, COUNT(*)
		FROM bookings b` + cityJoin + `
		WHERE b.created_at >= $1 AND b.created_at <= $2` + cityWhere + `
		GROUP BY b.status
		ORDER BY COUNT(*) DESC`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: booking status distribution", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var di DistributionItem
		if err := rows.Scan(&di.Label, &di.Value); err != nil {
			return err
		}
		*items = append(*items, di)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadReviewRatingDist(ctx context.Context, from, to time.Time, cityID string, items *[]DistributionItem) error {
	cityJoin := ""
	cityWhere := ""
	args := []interface{}{from, to}
	if cityID != "" {
		cityJoin = " JOIN bathhouses bh ON bh.id = r.bathhouse_id"
		cityWhere = " AND bh.city_id = $3"
		args = append(args, cityID)
	}

	query := `
		SELECT r.rating::text, COUNT(*)
		FROM reviews r` + cityJoin + `
		WHERE r.created_at >= $1 AND r.created_at <= $2` + cityWhere + `
		GROUP BY r.rating
		ORDER BY r.rating`

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("analytics: review rating distribution", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var di DistributionItem
		if err := rows.Scan(&di.Label, &di.Value); err != nil {
			return err
		}
		*items = append(*items, di)
	}
	return rows.Err()
}

func (p *PostgresAnalyticsProvider) loadSummary(ctx context.Context, dateFrom, dateTo time.Time, cityID string, summary *AnalyticsSummary) error {
	// Total bookings and revenue for the period.
	cityFilter := ""
	args := []interface{}{dateFrom, dateTo}
	if cityID != "" {
		cityFilter = " AND b.bathhouse_id IN (SELECT id FROM bathhouses WHERE city_id = $3)"
		args = append(args, cityID)
	}

	query := `SELECT COUNT(*), COALESCE(SUM(b.total_price), 0) FROM bookings b WHERE b.created_at >= $1 AND b.created_at <= $2` + cityFilter
	if err := p.pool.QueryRow(ctx, query, args...).Scan(&summary.TotalBookings, &summary.TotalRevenue); err != nil {
		p.log.Error("analytics: summary bookings/revenue", "error", err)
		return err
	}

	// Total new users for the period.
	userFilter := ""
	userArgs := []interface{}{dateFrom, dateTo}
	if cityID != "" {
		userFilter = " AND u.id IN (SELECT DISTINCT b2.user_id FROM bookings b2 JOIN bathhouses bh2 ON bh2.id = b2.bathhouse_id WHERE bh2.city_id = $3)"
		userArgs = append(userArgs, cityID)
	}
	userQuery := `SELECT COUNT(*) FROM users u WHERE u.created_at >= $1 AND u.created_at <= $2` + userFilter
	if err := p.pool.QueryRow(ctx, userQuery, userArgs...).Scan(&summary.TotalNewUsers); err != nil {
		p.log.Error("analytics: summary new users", "error", err)
		return err
	}

	// Previous period comparison.
	duration := dateTo.Sub(dateFrom)
	prevTo := dateFrom.Add(-time.Second)
	prevFrom := prevTo.Add(-duration)

	prevArgs := []interface{}{prevFrom, prevTo}
	prevCityFilter := ""
	if cityID != "" {
		prevCityFilter = " AND b.bathhouse_id IN (SELECT id FROM bathhouses WHERE city_id = $3)"
		prevArgs = append(prevArgs, cityID)
	}

	var prevBookings, prevRevenue int64
	prevQuery := `SELECT COUNT(*), COALESCE(SUM(b.total_price), 0) FROM bookings b WHERE b.created_at >= $1 AND b.created_at <= $2` + prevCityFilter
	if err := p.pool.QueryRow(ctx, prevQuery, prevArgs...).Scan(&prevBookings, &prevRevenue); err != nil {
		p.log.Error("analytics: prev period bookings/revenue", "error", err)
		return err
	}

	var prevUsers int64
	prevUserArgs := []interface{}{prevFrom, prevTo}
	prevUserFilter := ""
	if cityID != "" {
		prevUserFilter = " AND u.id IN (SELECT DISTINCT b2.user_id FROM bookings b2 JOIN bathhouses bh2 ON bh2.id = b2.bathhouse_id WHERE bh2.city_id = $3)"
		prevUserArgs = append(prevUserArgs, cityID)
	}
	prevUserQuery := `SELECT COUNT(*) FROM users u WHERE u.created_at >= $1 AND u.created_at <= $2` + prevUserFilter
	if err := p.pool.QueryRow(ctx, prevUserQuery, prevUserArgs...).Scan(&prevUsers); err != nil {
		p.log.Error("analytics: prev period users", "error", err)
		return err
	}

	summary.BookingsTrend = TrendData{Current: summary.TotalBookings, Previous: prevBookings}
	summary.RevenueTrend = TrendData{Current: summary.TotalRevenue, Previous: prevRevenue}
	summary.UsersTrend = TrendData{Current: summary.TotalNewUsers, Previous: prevUsers}

	return nil
}

// BuildDatePresets returns date preset buttons with the active one highlighted.
func BuildDatePresets(now time.Time, filter AnalyticsFilter) []DatePreset {
	today := now.Format("2006-01-02")
	day7 := now.AddDate(0, 0, -6).Format("2006-01-02")
	day30 := now.AddDate(0, 0, -29).Format("2006-01-02")

	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	prevMonthEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
	prevMonthStart := time.Date(prevMonthEnd.Year(), prevMonthEnd.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	prevMonthEndStr := prevMonthEnd.Format("2006-01-02")

	thisYearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	presets := []DatePreset{
		{Label: "Сегодня", DateFrom: today, DateTo: today},
		{Label: "7 дней", DateFrom: day7, DateTo: today},
		{Label: "30 дней", DateFrom: day30, DateTo: today},
		{Label: "Этот месяц", DateFrom: thisMonthStart, DateTo: today},
		{Label: "Прошлый месяц", DateFrom: prevMonthStart, DateTo: prevMonthEndStr},
		{Label: "Этот год", DateFrom: thisYearStart, DateTo: today},
	}

	for i := range presets {
		if filter.DateFrom == presets[i].DateFrom && filter.DateTo == presets[i].DateTo {
			presets[i].Active = true
		}
	}

	return presets
}

// AnalyticsHandler serves the analytics dashboard page.
type AnalyticsHandler struct {
	provider    AnalyticsDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(provider AnalyticsDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *AnalyticsHandler {
	return &AnalyticsHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the analytics dashboard page.
func (h *AnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	cityID := q.Get("city_id")
	if cityID != "" {
		if _, err := strconv.ParseInt(cityID, 10, 64); err != nil {
			cityID = ""
		}
	}

	filter := AnalyticsFilter{
		DateFrom: q.Get("date_from"),
		DateTo:   q.Get("date_to"),
		CityID:   cityID,
	}

	data, err := h.provider.GetAnalyticsData(r.Context(), filter)
	if err != nil {
		h.log.Error("analytics: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.PagesPrefix = h.pagesPrefix
	data.AdminPrefix = h.adminPrefix
	data.PageTitle = "Аналитика платформы"
	data.ActivePage = "analytics"

	var buf bytes.Buffer
	if err := analyticsTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("analytics: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}

// HandleCSVExport serves CSV data for a specific chart type.
func (h *AnalyticsHandler) HandleCSVExport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	exportType := q.Get("type")
	if exportType == "" {
		http.Error(w, "missing type parameter", http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{
		"bookings": true, "revenue": true, "users": true,
		"top_bookings": true, "top_revenue": true,
	}
	if !validTypes[exportType] {
		http.Error(w, "invalid type parameter", http.StatusBadRequest)
		return
	}

	cityID := q.Get("city_id")
	if cityID != "" {
		if _, err := strconv.ParseInt(cityID, 10, 64); err != nil {
			cityID = ""
		}
	}

	filter := AnalyticsFilter{
		DateFrom: q.Get("from"),
		DateTo:   q.Get("to"),
		CityID:   cityID,
	}

	data, err := h.provider.GetAnalyticsData(r.Context(), filter)
	if err != nil {
		h.log.Error("analytics: csv export get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fromStr := filter.DateFrom
	if fromStr == "" {
		fromStr = "start"
	}
	toStr := filter.DateTo
	if toStr == "" {
		toStr = "end"
	}
	filename := fmt.Sprintf("analytics_%s_%s_%s.csv", exportType, fromStr, toStr)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	// BOM for Excel to correctly detect UTF-8.
	w.Write([]byte{0xEF, 0xBB, 0xBF}) //nolint:errcheck

	csvW := csv.NewWriter(w)
	defer csvW.Flush()

	switch exportType {
	case "bookings":
		csvW.Write([]string{"Дата", "Бронирования"}) //nolint:errcheck
		for _, p := range data.BookingsPerDay {
			csvW.Write([]string{p.Label, strconv.FormatInt(p.Value, 10)}) //nolint:errcheck
		}
	case "revenue":
		csvW.Write([]string{"Дата", "Выручка (руб.)"}) //nolint:errcheck
		for _, p := range data.RevenuePerDay {
			csvW.Write([]string{p.Label, FormatKopecksToRubles(p.Value)}) //nolint:errcheck
		}
	case "users":
		csvW.Write([]string{"Дата", "Новые пользователи"}) //nolint:errcheck
		for _, p := range data.NewUsersPerDay {
			csvW.Write([]string{p.Label, strconv.FormatInt(p.Value, 10)}) //nolint:errcheck
		}
	case "top_bookings":
		csvW.Write([]string{"Название", "Бронирования"}) //nolint:errcheck
		for _, item := range data.TopByBookings {
			csvW.Write([]string{item.Name, strconv.FormatInt(item.Value, 10)}) //nolint:errcheck
		}
	case "top_revenue":
		csvW.Write([]string{"Название", "Выручка (руб.)"}) //nolint:errcheck
		for _, item := range data.TopByRevenue {
			csvW.Write([]string{item.Name, FormatKopecksToRubles(item.Value)}) //nolint:errcheck
		}
	}
}

// SummaryTrendPercent returns the formatted absolute percentage for a summary trend.
func SummaryTrendPercent(t TrendData) string {
	p := math.Abs(t.Percent())
	if p == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.0f%%", p)
}
