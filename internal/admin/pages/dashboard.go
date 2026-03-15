package pages

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
)

var dashboardFuncMap = template.FuncMap{
	"formatRubles": FormatKopecksToRubles,
	"relativeTime": RelativeTime,
	"truncateText": TruncateText,
	"trendClass":   TrendClass,
	"trendArrow":   TrendArrow,
	"formatTrend":  FormatTrend,
}

var dashboardTmpl = ParsePageTemplate(dashboardFuncMap, "templates/dashboard.tmpl")

// TrendData holds comparison data against a previous period.
type TrendData struct {
	Current  int64
	Previous int64
}

// Percent returns the percentage change from Previous to Current.
// Returns 0 if Previous is 0.
func (t TrendData) Percent() float64 {
	if t.Previous == 0 {
		return 0
	}
	return float64(t.Current-t.Previous) / float64(t.Previous) * 100
}

// KPICards holds the main KPI metrics for the admin dashboard.
type KPICards struct {
	TotalUsers      int64
	TotalBathhouses int64

	BookingsToday int64
	BookingsWeek  int64
	BookingsMonth int64

	// Revenue in kopecks.
	RevenueToday int64
	RevenueWeek  int64
	RevenueMonth int64

	// Trends: comparison with previous period.
	UsersTrend         TrendData
	BathhousesTrend    TrendData
	BookingsTodayTrend TrendData
	RevenueTodayTrend  TrendData
	RevenueWeekTrend   TrendData
	RevenueMonthTrend  TrendData
}

// StatusCards holds counts of items needing admin attention.
type StatusCards struct {
	PendingBathhouses   int64
	PendingReviews      int64
	ActiveSubscriptions int64
}

// RecentBooking represents a booking in the activity feed.
type RecentBooking struct {
	ID            string
	UserName      string
	BathhouseName string
	StartTime     time.Time
	TotalPrice    int64
	Status        string
}

// RecentReview represents a review in the activity feed.
type RecentReview struct {
	ID            string
	UserName      string
	BathhouseName string
	Rating        int
	Text          string
	Status        string
	CreatedAt     time.Time
}

// RecentRegistration represents a user registration in the activity feed.
type RecentRegistration struct {
	ID        string
	Name      string
	Email     string
	Role      string
	CreatedAt time.Time
}

// DashboardData is the full data model for the admin dashboard page.
type DashboardData struct {
	KPI           KPICards
	Status        StatusCards
	Bookings      []RecentBooking
	Reviews       []RecentReview
	Registrations []RecentRegistration
	GeneratedAt   time.Time
	AdminPrefix   string
	PagesPrefix   string
	PageTitle     string
	ActivePage    string
}

// DashboardDataProvider fetches dashboard data from a data source.
type DashboardDataProvider interface {
	GetDashboardData(ctx context.Context) (*DashboardData, error)
}

// PostgresDashboardProvider fetches dashboard data from PostgreSQL.
type PostgresDashboardProvider struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewPostgresDashboardProvider creates a new PostgresDashboardProvider.
func NewPostgresDashboardProvider(pool *pgxpool.Pool, log *logger.Logger) *PostgresDashboardProvider {
	return &PostgresDashboardProvider{pool: pool, log: log}
}

func (p *PostgresDashboardProvider) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	data := &DashboardData{GeneratedAt: time.Now()}

	if err := p.loadKPIs(ctx, &data.KPI); err != nil {
		return nil, err
	}
	if err := p.loadTrends(ctx, &data.KPI); err != nil {
		return nil, err
	}
	if err := p.loadStatusCards(ctx, &data.Status); err != nil {
		return nil, err
	}
	if err := p.loadRecentBookings(ctx, &data.Bookings); err != nil {
		return nil, err
	}
	if err := p.loadRecentReviews(ctx, &data.Reviews); err != nil {
		return nil, err
	}
	if err := p.loadRecentRegistrations(ctx, &data.Registrations); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *PostgresDashboardProvider) loadKPIs(ctx context.Context, kpi *KPICards) error {
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&kpi.TotalUsers)
	if err != nil {
		p.log.Error("dashboard: count users", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bathhouses").Scan(&kpi.TotalBathhouses)
	if err != nil {
		p.log.Error("dashboard: count bathhouses", "error", err)
		return err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	if now.Weekday() == time.Sunday {
		weekStart = todayStart.AddDate(0, 0, -6)
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	bookingQuery := "SELECT COUNT(*), COALESCE(SUM(total_price), 0) FROM bookings WHERE created_at >= $1"

	err = p.pool.QueryRow(ctx, bookingQuery, todayStart).Scan(&kpi.BookingsToday, &kpi.RevenueToday)
	if err != nil {
		p.log.Error("dashboard: bookings today", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, bookingQuery, weekStart).Scan(&kpi.BookingsWeek, &kpi.RevenueWeek)
	if err != nil {
		p.log.Error("dashboard: bookings week", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, bookingQuery, monthStart).Scan(&kpi.BookingsMonth, &kpi.RevenueMonth)
	if err != nil {
		p.log.Error("dashboard: bookings month", "error", err)
		return err
	}

	return nil
}

func (p *PostgresDashboardProvider) loadTrends(ctx context.Context, kpi *KPICards) error {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	if now.Weekday() == time.Sunday {
		weekStart = todayStart.AddDate(0, 0, -6)
	}
	prevWeekStart := weekStart.AddDate(0, 0, -7)

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonthStart := monthStart.AddDate(0, -1, 0)

	// Users trend: registered this week vs previous week.
	var usersThisWeek, usersPrevWeek int64
	err := p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE created_at >= $1", weekStart,
	).Scan(&usersThisWeek)
	if err != nil {
		p.log.Error("dashboard: users trend this week", "error", err)
		return err
	}
	err = p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2",
		prevWeekStart, weekStart,
	).Scan(&usersPrevWeek)
	if err != nil {
		p.log.Error("dashboard: users trend prev week", "error", err)
		return err
	}
	kpi.UsersTrend = TrendData{Current: usersThisWeek, Previous: usersPrevWeek}

	// Bathhouses trend: new this week vs previous week.
	var bhThisWeek, bhPrevWeek int64
	err = p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM bathhouses WHERE created_at >= $1", weekStart,
	).Scan(&bhThisWeek)
	if err != nil {
		p.log.Error("dashboard: bathhouses trend this week", "error", err)
		return err
	}
	err = p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM bathhouses WHERE created_at >= $1 AND created_at < $2",
		prevWeekStart, weekStart,
	).Scan(&bhPrevWeek)
	if err != nil {
		p.log.Error("dashboard: bathhouses trend prev week", "error", err)
		return err
	}
	kpi.BathhousesTrend = TrendData{Current: bhThisWeek, Previous: bhPrevWeek}

	// Bookings today trend: today vs yesterday.
	bookingCountQuery := "SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at < $2"
	revenueQuery := "SELECT COALESCE(SUM(total_price), 0) FROM bookings WHERE created_at >= $1 AND created_at < $2"

	var bookingsYesterday int64
	err = p.pool.QueryRow(ctx, bookingCountQuery, yesterdayStart, todayStart).Scan(&bookingsYesterday)
	if err != nil {
		p.log.Error("dashboard: bookings yesterday", "error", err)
		return err
	}
	kpi.BookingsTodayTrend = TrendData{Current: kpi.BookingsToday, Previous: bookingsYesterday}

	// Revenue today trend: today vs yesterday.
	var revenueYesterday int64
	err = p.pool.QueryRow(ctx, revenueQuery, yesterdayStart, todayStart).Scan(&revenueYesterday)
	if err != nil {
		p.log.Error("dashboard: revenue yesterday", "error", err)
		return err
	}
	kpi.RevenueTodayTrend = TrendData{Current: kpi.RevenueToday, Previous: revenueYesterday}

	// Revenue week trend: this week vs previous week.
	var revenuePrevWeek int64
	err = p.pool.QueryRow(ctx, revenueQuery, prevWeekStart, weekStart).Scan(&revenuePrevWeek)
	if err != nil {
		p.log.Error("dashboard: revenue prev week", "error", err)
		return err
	}
	kpi.RevenueWeekTrend = TrendData{Current: kpi.RevenueWeek, Previous: revenuePrevWeek}

	// Revenue month trend: this month vs previous month.
	var revenuePrevMonth int64
	err = p.pool.QueryRow(ctx, revenueQuery, prevMonthStart, monthStart).Scan(&revenuePrevMonth)
	if err != nil {
		p.log.Error("dashboard: revenue prev month", "error", err)
		return err
	}
	kpi.RevenueMonthTrend = TrendData{Current: kpi.RevenueMonth, Previous: revenuePrevMonth}

	return nil
}

func (p *PostgresDashboardProvider) loadStatusCards(ctx context.Context, status *StatusCards) error {
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bathhouses WHERE status = 'pending'").Scan(&status.PendingBathhouses)
	if err != nil {
		p.log.Error("dashboard: pending bathhouses", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending'").Scan(&status.PendingReviews)
	if err != nil {
		p.log.Error("dashboard: pending reviews", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM subscriptions WHERE status = 'active'").Scan(&status.ActiveSubscriptions)
	if err != nil {
		p.log.Error("dashboard: active subscriptions", "error", err)
		return err
	}

	return nil
}

func (p *PostgresDashboardProvider) loadRecentBookings(ctx context.Context, bookings *[]RecentBooking) error {
	rows, err := p.pool.Query(ctx, `
		SELECT b.id, u.name, bh.name, b.start_time, b.total_price, b.status
		FROM bookings b
		JOIN users u ON u.id = b.user_id
		JOIN bathhouses bh ON bh.id = b.bathhouse_id
		ORDER BY b.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		p.log.Error("dashboard: recent bookings", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rb RecentBooking
		if err := rows.Scan(&rb.ID, &rb.UserName, &rb.BathhouseName, &rb.StartTime, &rb.TotalPrice, &rb.Status); err != nil {
			return err
		}
		*bookings = append(*bookings, rb)
	}
	return rows.Err()
}

func (p *PostgresDashboardProvider) loadRecentReviews(ctx context.Context, reviews *[]RecentReview) error {
	rows, err := p.pool.Query(ctx, `
		SELECT r.id, u.name, bh.name, r.rating, r.text, r.status, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		JOIN bathhouses bh ON bh.id = r.bathhouse_id
		ORDER BY r.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		p.log.Error("dashboard: recent reviews", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rr RecentReview
		if err := rows.Scan(&rr.ID, &rr.UserName, &rr.BathhouseName, &rr.Rating, &rr.Text, &rr.Status, &rr.CreatedAt); err != nil {
			return err
		}
		*reviews = append(*reviews, rr)
	}
	return rows.Err()
}

func (p *PostgresDashboardProvider) loadRecentRegistrations(ctx context.Context, registrations *[]RecentRegistration) error {
	rows, err := p.pool.Query(ctx, `
		SELECT id, name, email, role, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		p.log.Error("dashboard: recent registrations", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rr RecentRegistration
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.Email, &rr.Role, &rr.CreatedAt); err != nil {
			return err
		}
		*registrations = append(*registrations, rr)
	}
	return rows.Err()
}

// DashboardHandler serves the admin dashboard page.
type DashboardHandler struct {
	provider    DashboardDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(provider DashboardDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *DashboardHandler {
	return &DashboardHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the admin dashboard page.
func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.provider.GetDashboardData(r.Context())
	if err != nil {
		h.log.Error("dashboard: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.AdminPrefix = h.adminPrefix
	data.PagesPrefix = h.pagesPrefix
	data.PageTitle = "Панель управления"
	data.ActivePage = "dashboard"

	var buf bytes.Buffer
	if err := dashboardTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("dashboard: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}

// FormatKopecksToRubles converts kopecks to rubles string (e.g. 150050 -> "1 500.50").
func FormatKopecksToRubles(kopecks int64) string {
	negative := kopecks < 0
	if negative {
		kopecks = -kopecks
	}

	rubles := kopecks / 100
	kop := kopecks % 100

	// Format with thousands separator.
	s := fmt.Sprintf("%d", rubles)
	// Insert space separators from right to left.
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, " ")
	}

	result := fmt.Sprintf("%s.%02d", s, kop)
	if negative {
		return "-" + result
	}
	return result
}

// RelativeTime returns a human-readable Russian string for elapsed time.
func RelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "только что"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%d мин. назад", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%d ч. назад", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "вчера"
		}
		return fmt.Sprintf("%d дн. назад", days)
	}
}

// TruncateText truncates text to maxLen characters and appends "..." if truncated.
func TruncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// TrendClass returns CSS class name based on trend direction.
func TrendClass(t TrendData) string {
	p := t.Percent()
	switch {
	case p > 0:
		return "trend-up"
	case p < 0:
		return "trend-down"
	default:
		return "trend-neutral"
	}
}

// TrendArrow returns an arrow indicator for the trend.
func TrendArrow(t TrendData) string {
	p := t.Percent()
	switch {
	case p > 0:
		return "↑"
	case p < 0:
		return "↓"
	default:
		return "="
	}
}

// FormatTrend returns the absolute percentage value as a formatted string.
func FormatTrend(t TrendData) string {
	p := math.Abs(t.Percent())
	if p == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.0f%%", p)
}
