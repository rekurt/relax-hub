package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type analyticsRepo struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) repository.AnalyticsRepository {
	return &analyticsRepo{pool: pool}
}

func (r *analyticsRepo) RecordView(ctx context.Context, view *domain.BathhouseView) error {
	if view.ID == uuid.Nil {
		view.ID = uuid.New()
	}
	if view.ViewedAt.IsZero() {
		view.ViewedAt = time.Now()
	}

	query := `
		INSERT INTO bathhouse_views (id, bathhouse_id, viewer_id, source, ip_hash, viewed_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		view.ID, view.BathhouseID, view.ViewerID, view.Source, view.IPHash, view.ViewedAt)
	if err != nil {
		return fmt.Errorf("record view: %w", err)
	}
	return nil
}

func (r *analyticsRepo) GetBathhouseStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) (*domain.AnalyticsSnapshot, error) {
	// Normalize to date only (midnight)
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT bathhouse_id, date, views, unique_views, bookings, revenue, review_count, avg_rating
		FROM analytics_snapshots
		WHERE bathhouse_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC`

	rows, err := r.pool.Query(ctx, query, bathhouseID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse stats: %w", err)
	}
	defer rows.Close()

	// Aggregate all snapshots into a single result
	result := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        fromDate,
	}

	for rows.Next() {
		var snapshot domain.AnalyticsSnapshot
		if err := rows.Scan(&snapshot.BathhouseID, &snapshot.Date, &snapshot.Views,
			&snapshot.UniqueViews, &snapshot.Bookings, &snapshot.Revenue,
			&snapshot.ReviewCount, &snapshot.AvgRating); err != nil {
			return nil, fmt.Errorf("scan bathhouse stats: %w", err)
		}
		result.Views += snapshot.Views
		result.UniqueViews += snapshot.UniqueViews
		result.Bookings += snapshot.Bookings
		result.Revenue += snapshot.Revenue
		// Weighted average for ratings - calculate BEFORE incrementing count
		if snapshot.ReviewCount > 0 && result.AvgRating >= 0 {
			oldCount := result.ReviewCount
			newCount := oldCount + snapshot.ReviewCount
			if newCount > 0 {
				result.AvgRating = (result.AvgRating*float64(oldCount) + snapshot.AvgRating*float64(snapshot.ReviewCount)) / float64(newCount)
			} else {
				result.AvgRating = snapshot.AvgRating
			}
		}
		result.ReviewCount += snapshot.ReviewCount
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("bathhouse stats error: %w", err)
	}

	return result, nil
}

func (r *analyticsRepo) GetDailyStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error) {
	// Normalize to date only (midnight)
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT bathhouse_id, date, views, unique_views, bookings, revenue, review_count, avg_rating
		FROM analytics_snapshots
		WHERE bathhouse_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC`

	rows, err := r.pool.Query(ctx, query, bathhouseID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get daily stats: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.AnalyticsSnapshot
	for rows.Next() {
		var snapshot domain.AnalyticsSnapshot
		if err := rows.Scan(&snapshot.BathhouseID, &snapshot.Date, &snapshot.Views,
			&snapshot.UniqueViews, &snapshot.Bookings, &snapshot.Revenue,
			&snapshot.ReviewCount, &snapshot.AvgRating); err != nil {
			return nil, fmt.Errorf("scan daily stats: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("daily stats error: %w", err)
	}

	return snapshots, nil
}

func (r *analyticsRepo) GetPlatformStats(ctx context.Context, from, to time.Time) (*domain.AnalyticsSnapshot, error) {
	// Normalize to date only (midnight)
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT
			SUM(views), SUM(unique_views), SUM(bookings), SUM(revenue), SUM(review_count), AVG(avg_rating)
		FROM analytics_snapshots
		WHERE date >= $1 AND date <= $2`

	result := &domain.AnalyticsSnapshot{
		Date: fromDate,
	}

	var views, uniqueViews, bookings, revenue, reviewCount sql.NullInt64
	var avgRating sql.NullFloat64

	err := r.pool.QueryRow(ctx, query, fromDate, toDate).Scan(
		&views, &uniqueViews, &bookings, &revenue, &reviewCount, &avgRating)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get platform stats: %w", err)
	}

	if views.Valid {
		result.Views = views.Int64
	}
	if uniqueViews.Valid {
		result.UniqueViews = uniqueViews.Int64
	}
	if bookings.Valid {
		result.Bookings = bookings.Int64
	}
	if revenue.Valid {
		result.Revenue = revenue.Int64
	}
	if reviewCount.Valid {
		result.ReviewCount = int(reviewCount.Int64)
	}
	if avgRating.Valid {
		result.AvgRating = avgRating.Float64
	}

	return result, nil
}

func (r *analyticsRepo) GetTopBathhouses(ctx context.Context, metric domain.TopMetric, limit int) ([]uuid.UUID, error) {
	if !metric.IsValid() {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	if limit < 1 {
		limit = 10
	}

	// Use CASE statement for safe metric-based ordering (no fmt.Sprintf string injection)
	query := `
		SELECT bathhouse_id
		FROM analytics_snapshots
		GROUP BY bathhouse_id
		ORDER BY CASE $1
			WHEN 'views' THEN SUM(views)
			WHEN 'bookings' THEN SUM(bookings)
			WHEN 'revenue' THEN SUM(revenue)
			WHEN 'rating' THEN AVG(avg_rating)
			ELSE SUM(bookings)
		END DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, string(metric), limit)
	if err != nil {
		return nil, fmt.Errorf("get top bathhouses: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan top bathhouses: %w", err)
		}
		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("top bathhouses error: %w", err)
	}

	return ids, nil
}

// AggregateRawData aggregates analytics data from source tables for a specific date
func (r *analyticsRepo) AggregateRawData(ctx context.Context, bathhouseID uuid.UUID, date time.Time) (*domain.AnalyticsSnapshot, error) {
	// Normalize to date only
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	// Aggregate views from bathhouse_views
	var views, uniqueViews int64
	viewQuery := `
		SELECT COUNT(*) as views, COUNT(DISTINCT ip_hash) as unique_views
		FROM bathhouse_views
		WHERE bathhouse_id = $1 AND viewed_at >= $2 AND viewed_at <= $3`
	err := r.pool.QueryRow(ctx, viewQuery, bathhouseID, startOfDay, endOfDay).Scan(&views, &uniqueViews)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("aggregate views: %w", err)
	}

	// Aggregate bookings and revenue from bookings table (only completed bookings)
	var bookings, revenue sql.NullInt64
	bookingQuery := `
		SELECT COUNT(*) as bookings, COALESCE(SUM(total_price), 0) as revenue
		FROM bookings
		WHERE bathhouse_id = $1 AND status = 'completed' AND created_at >= $2 AND created_at <= $3`
	err = r.pool.QueryRow(ctx, bookingQuery, bathhouseID, startOfDay, endOfDay).Scan(&bookings, &revenue)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("aggregate bookings: %w", err)
	}

	// Aggregate reviews from reviews table
	var reviewCount sql.NullInt64
	var avgRating sql.NullFloat64
	reviewQuery := `
		SELECT COUNT(*) as review_count, COALESCE(AVG(rating), 0.0) as avg_rating
		FROM reviews
		WHERE bathhouse_id = $1 AND created_at >= $2 AND created_at <= $3`
	err = r.pool.QueryRow(ctx, reviewQuery, bathhouseID, startOfDay, endOfDay).Scan(&reviewCount, &avgRating)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("aggregate reviews: %w", err)
	}

	snapshot := &domain.AnalyticsSnapshot{
		BathhouseID: bathhouseID,
		Date:        startOfDay,
		Views:       views,
		UniqueViews: uniqueViews,
		Bookings:    bookings.Int64,
		Revenue:     revenue.Int64,
		ReviewCount: int(reviewCount.Int64),
		AvgRating:   avgRating.Float64,
	}

	return snapshot, nil
}

func (r *analyticsRepo) CreateSnapshot(ctx context.Context, snapshot *domain.AnalyticsSnapshot) error {
	query := `
		INSERT INTO analytics_snapshots (bathhouse_id, date, views, unique_views, bookings, revenue, review_count, avg_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (bathhouse_id, date) DO UPDATE SET
			views = EXCLUDED.views,
			unique_views = EXCLUDED.unique_views,
			bookings = EXCLUDED.bookings,
			revenue = EXCLUDED.revenue,
			review_count = EXCLUDED.review_count,
			avg_rating = EXCLUDED.avg_rating`

	_, err := r.pool.Exec(ctx, query,
		snapshot.BathhouseID, snapshot.Date, snapshot.Views, snapshot.UniqueViews,
		snapshot.Bookings, snapshot.Revenue, snapshot.ReviewCount, snapshot.AvgRating)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}
	return nil
}

// DeleteOldViews removes bathhouse view records older than the specified date
func (r *analyticsRepo) DeleteOldViews(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM bathhouse_views WHERE viewed_at < $1`

	result, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("delete old views: %w", err)
	}

	return result.RowsAffected(), nil
}

// --- Advanced analytics (FR-147-154) ---

func (r *analyticsRepo) GetConversionFunnel(ctx context.Context, from, to time.Time) ([]domain.FunnelStep, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	// Step 1: unique visitors (distinct IPs in views)
	var visits int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT ip_hash) FROM bathhouse_views WHERE viewed_at >= $1 AND viewed_at <= $2`,
		fromDate, toDate).Scan(&visits)
	if err != nil {
		return nil, fmt.Errorf("funnel visits: %w", err)
	}

	// Step 2: searches (views with source='search')
	var searches int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT ip_hash) FROM bathhouse_views WHERE source = 'search' AND viewed_at >= $1 AND viewed_at <= $2`,
		fromDate, toDate).Scan(&searches)
	if err != nil {
		return nil, fmt.Errorf("funnel searches: %w", err)
	}

	// Step 3: card views (views with source='direct' or 'search' deduplicated by viewer_id)
	var cardViews int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT COALESCE(viewer_id::text, ip_hash)) FROM bathhouse_views WHERE viewed_at >= $1 AND viewed_at <= $2`,
		fromDate, toDate).Scan(&cardViews)
	if err != nil {
		return nil, fmt.Errorf("funnel card views: %w", err)
	}

	// Step 4: bookings started (all bookings created in period)
	var bookingsStarted int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&bookingsStarted)
	if err != nil {
		return nil, fmt.Errorf("funnel bookings started: %w", err)
	}

	// Step 5: paid bookings
	var paid int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at <= $2 AND status IN ('confirmed', 'completed')`,
		fromDate, toDate).Scan(&paid)
	if err != nil {
		return nil, fmt.Errorf("funnel paid: %w", err)
	}

	// Step 6: completed visits
	var completed int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at <= $2 AND status = 'completed'`,
		fromDate, toDate).Scan(&completed)
	if err != nil {
		return nil, fmt.Errorf("funnel completed: %w", err)
	}

	steps := []domain.FunnelStep{
		{Name: "visit", Count: visits},
		{Name: "search", Count: searches},
		{Name: "view_card", Count: cardViews},
		{Name: "start_booking", Count: bookingsStarted},
		{Name: "pay", Count: paid},
		{Name: "complete_visit", Count: completed},
	}

	// Calculate percentages relative to first step
	if visits > 0 {
		for i := range steps {
			steps[i].Percentage = float64(steps[i].Count) / float64(visits) * 100
		}
	}

	return steps, nil
}

func (r *analyticsRepo) GetCohortAnalysis(ctx context.Context, months int) ([]domain.CohortRow, error) {
	if months < 1 {
		months = 6
	}

	query := `
		WITH cohorts AS (
			SELECT
				to_char(u.created_at, 'YYYY-MM') AS cohort_month,
				u.id AS user_id
			FROM users u
			WHERE u.created_at >= NOW() - ($1 || ' months')::interval
		),
		activity AS (
			SELECT
				c.cohort_month,
				c.user_id,
				EXTRACT(WEEK FROM (b.created_at - DATE_TRUNC('month', c.cohort_month::date)))::int AS week_num,
				b.total_price
			FROM cohorts c
			JOIN bookings b ON b.user_id = c.user_id AND b.status = 'completed'
		)
		SELECT
			c.cohort_month,
			COUNT(DISTINCT c.user_id) AS users_count,
			COALESCE(SUM(a.total_price), 0) AS total_spending
		FROM cohorts c
		LEFT JOIN activity a ON a.cohort_month = c.cohort_month AND a.user_id = c.user_id
		GROUP BY c.cohort_month
		ORDER BY c.cohort_month`

	rows, err := r.pool.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("cohort analysis: %w", err)
	}
	defer rows.Close()

	var cohorts []domain.CohortRow
	for rows.Next() {
		var row domain.CohortRow
		if err := rows.Scan(&row.CohortMonth, &row.UsersCount, &row.TotalSpending); err != nil {
			return nil, fmt.Errorf("scan cohort: %w", err)
		}
		cohorts = append(cohorts, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cohort rows: %w", err)
	}

	// Calculate retention per week for each cohort
	for i := range cohorts {
		retention, err := r.getCohortRetention(ctx, cohorts[i].CohortMonth)
		if err != nil {
			continue
		}
		cohorts[i].RetentionWeeks = retention
	}

	return cohorts, nil
}

func (r *analyticsRepo) getCohortRetention(ctx context.Context, cohortMonth string) ([]float64, error) {
	query := `
		WITH cohort_users AS (
			SELECT id FROM users WHERE to_char(created_at, 'YYYY-MM') = $1
		),
		weeks AS (
			SELECT generate_series(0, 12) AS week_num
		)
		SELECT
			w.week_num,
			CASE WHEN (SELECT COUNT(*) FROM cohort_users) > 0
				THEN COUNT(DISTINCT b.user_id)::float / (SELECT COUNT(*) FROM cohort_users) * 100
				ELSE 0
			END AS retention_pct
		FROM weeks w
		LEFT JOIN bookings b ON b.user_id IN (SELECT id FROM cohort_users)
			AND b.status = 'completed'
			AND EXTRACT(WEEK FROM (b.created_at - ($1 || '-01')::date))::int = w.week_num
		GROUP BY w.week_num
		ORDER BY w.week_num`

	rows, err := r.pool.Query(ctx, query, cohortMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retention []float64
	for rows.Next() {
		var weekNum int
		var pct float64
		if err := rows.Scan(&weekNum, &pct); err != nil {
			return nil, err
		}
		retention = append(retention, pct)
	}
	return retention, rows.Err()
}

func (r *analyticsRepo) GetGeoSupplyDemand(ctx context.Context, from, to time.Time) ([]domain.GeoSupplyDemand, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT
			c.id AS city_id,
			c.name AS city_name,
			COALESCE(search_counts.cnt, 0) AS search_count,
			COALESCE(listing_counts.cnt, 0) AS listing_count,
			COALESCE(booking_counts.cnt, 0) AS booking_count
		FROM cities c
		LEFT JOIN (
			SELECT bh.city_id, COUNT(DISTINCT bv.ip_hash) AS cnt
			FROM bathhouse_views bv
			JOIN bathhouses bh ON bh.id = bv.bathhouse_id
			WHERE bv.viewed_at >= $1 AND bv.viewed_at <= $2 AND bv.source = 'search'
			GROUP BY bh.city_id
		) search_counts ON search_counts.city_id = c.id
		LEFT JOIN (
			SELECT city_id, COUNT(*) AS cnt
			FROM bathhouses
			WHERE status = 'active'
			GROUP BY city_id
		) listing_counts ON listing_counts.city_id = c.id
		LEFT JOIN (
			SELECT bh.city_id, COUNT(*) AS cnt
			FROM bookings b
			JOIN bathhouses bh ON bh.id = b.bathhouse_id
			WHERE b.created_at >= $1 AND b.created_at <= $2
			GROUP BY bh.city_id
		) booking_counts ON booking_counts.city_id = c.id
		ORDER BY c.name`

	rows, err := r.pool.Query(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("geo supply demand: %w", err)
	}
	defer rows.Close()

	var result []domain.GeoSupplyDemand
	for rows.Next() {
		var g domain.GeoSupplyDemand
		if err := rows.Scan(&g.CityID, &g.CityName, &g.SearchCount, &g.ListingCount, &g.BookingCount); err != nil {
			return nil, fmt.Errorf("scan geo: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (r *analyticsRepo) GetWalletMetrics(ctx context.Context, from, to time.Time) (*domain.WalletMetrics, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	m := &domain.WalletMetrics{}

	// Client balances (users with role 'client')
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(w.balance), 0)
		FROM wallets w
		JOIN users u ON u.id = w.user_id
		WHERE u.role = 'client' AND w.status = 'active'`).Scan(&m.TotalClientBalance)
	if err != nil {
		return nil, fmt.Errorf("wallet client balance: %w", err)
	}

	// Owner balances
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(w.balance), 0)
		FROM wallets w
		JOIN users u ON u.id = w.user_id
		WHERE u.role = 'owner' AND w.status = 'active'`).Scan(&m.TotalOwnerBalance)
	if err != nil {
		return nil, fmt.Errorf("wallet owner balance: %w", err)
	}

	// Escrow total
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM escrows WHERE status = 'held'`).Scan(&m.TotalEscrow)
	if err != nil {
		return nil, fmt.Errorf("wallet escrow: %w", err)
	}

	// Wallet payment share
	var totalBookings, walletBookings int64
	err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&totalBookings)
	if err != nil {
		return nil, fmt.Errorf("wallet total bookings: %w", err)
	}
	err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings WHERE created_at >= $1 AND created_at <= $2 AND payment_method IN ('wallet', 'combo')`,
		fromDate, toDate).Scan(&walletBookings)
	if err != nil {
		return nil, fmt.Errorf("wallet wallet bookings: %w", err)
	}
	if totalBookings > 0 {
		m.WalletPaymentShare = float64(walletBookings) / float64(totalBookings) * 100
	}

	// Expired bonuses in period
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_transactions
		WHERE type = 'bonus_expiry' AND created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&m.ExpiredBonusVolume)
	if err != nil {
		return nil, fmt.Errorf("wallet expired bonuses: %w", err)
	}

	// Active wallets
	err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM wallets WHERE status = 'active'`).Scan(&m.ActiveWallets)
	if err != nil {
		return nil, fmt.Errorf("wallet active count: %w", err)
	}

	return m, nil
}

// --- Business metrics (FR-148, FR-149) ---

// GetChurnRate returns the percentage of users who had bookings before `inactiveDays` ago
// but have not made any bookings in the last `inactiveDays` days.
func (r *analyticsRepo) GetChurnRate(ctx context.Context, inactiveDays int) (float64, error) {
	query := `
		WITH active_before AS (
			SELECT DISTINCT user_id
			FROM bookings
			WHERE status = 'completed'
			  AND created_at < NOW() - ($1 || ' days')::interval
		),
		active_recently AS (
			SELECT DISTINCT user_id
			FROM bookings
			WHERE status IN ('confirmed', 'completed')
			  AND created_at >= NOW() - ($1 || ' days')::interval
		)
		SELECT
			CASE WHEN (SELECT COUNT(*) FROM active_before) = 0 THEN 0
			ELSE (
				SELECT COUNT(*) FROM active_before ab
				WHERE ab.user_id NOT IN (SELECT user_id FROM active_recently)
			)::float / (SELECT COUNT(*) FROM active_before) * 100
			END`

	var churnRate float64
	err := r.pool.QueryRow(ctx, query, inactiveDays).Scan(&churnRate)
	if err != nil {
		return 0, fmt.Errorf("get churn rate: %w", err)
	}
	return churnRate, nil
}

// GetLTV returns the average lifetime value per user (total revenue from completed bookings / distinct users)
func (r *analyticsRepo) GetLTV(ctx context.Context) (int64, error) {
	query := `
		SELECT CASE WHEN COUNT(DISTINCT user_id) = 0 THEN 0
		ELSE COALESCE(SUM(total_price), 0) / COUNT(DISTINCT user_id)
		END
		FROM bookings
		WHERE status = 'completed'`

	var ltv int64
	err := r.pool.QueryRow(ctx, query).Scan(&ltv)
	if err != nil {
		return 0, fmt.Errorf("get ltv: %w", err)
	}
	return ltv, nil
}

// GetARPU returns the average revenue per user for the given period
func (r *analyticsRepo) GetARPU(ctx context.Context, from, to time.Time) (int64, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT CASE WHEN COUNT(DISTINCT user_id) = 0 THEN 0
		ELSE COALESCE(SUM(total_price), 0) / COUNT(DISTINCT user_id)
		END
		FROM bookings
		WHERE status = 'completed'
		  AND created_at >= $1 AND created_at <= $2`

	var arpu int64
	err := r.pool.QueryRow(ctx, query, fromDate, toDate).Scan(&arpu)
	if err != nil {
		return 0, fmt.Errorf("get arpu: %w", err)
	}
	return arpu, nil
}

// --- P&L metrics (FR-150) ---

// GetGMV returns the gross merchandise value (sum of total_price) and booking count for the period.
func (r *analyticsRepo) GetGMV(ctx context.Context, from, to time.Time) (int64, int64, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		SELECT COALESCE(SUM(total_price), 0), COUNT(*)
		FROM bookings
		WHERE status = 'completed'
		  AND created_at >= $1 AND created_at <= $2`

	var gmv, count int64
	err := r.pool.QueryRow(ctx, query, fromDate, toDate).Scan(&gmv, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("get gmv: %w", err)
	}
	return gmv, count, nil
}

// GetPlatformRevenue returns platform revenue broken down by source for the period.
func (r *analyticsRepo) GetPlatformRevenue(ctx context.Context, from, to time.Time) (int64, int64, int64, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	// Service fees from completed bookings
	var serviceFees int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(service_fee), 0)
		FROM bookings
		WHERE status = 'completed' AND created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&serviceFees)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("pnl service fees: %w", err)
	}

	// Subscriptions revenue
	var subscriptions int64
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE status = 'active' AND created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&subscriptions)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("pnl subscriptions: %w", err)
	}

	// Promotions revenue (promoted listings spending)
	var promotions int64
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE status = 'active' AND tier = 'promoted' AND created_at >= $1 AND created_at <= $2`,
		fromDate, toDate).Scan(&promotions)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("pnl promotions: %w", err)
	}

	return serviceFees, subscriptions, promotions, nil
}

// GetHeatmapData aggregates bathhouses and bookings/search views into grid cells for a heatmap.
// cellSize controls grid granularity in degrees (e.g., 0.01 ~ 1km).
func (r *analyticsRepo) GetHeatmapData(ctx context.Context, from, to time.Time, cellSize float64) ([]domain.HeatmapCell, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	query := `
		WITH grid AS (
			SELECT
				ROUND(bh.latitude / $3) * $3 + $3 / 2.0 AS cell_lat,
				ROUND(bh.longitude / $3) * $3 + $3 / 2.0 AS cell_lon,
				bh.id AS bathhouse_id
			FROM bathhouses bh
			WHERE bh.status = 'active' AND bh.latitude != 0 AND bh.longitude != 0
		),
		supply AS (
			SELECT cell_lat, cell_lon, COUNT(DISTINCT bathhouse_id) AS listing_count
			FROM grid
			GROUP BY cell_lat, cell_lon
		),
		demand AS (
			SELECT
				ROUND(bh.latitude / $3) * $3 + $3 / 2.0 AS cell_lat,
				ROUND(bh.longitude / $3) * $3 + $3 / 2.0 AS cell_lon,
				COUNT(DISTINCT bv.ip_hash) AS search_count
			FROM bathhouse_views bv
			JOIN bathhouses bh ON bh.id = bv.bathhouse_id
			WHERE bv.viewed_at >= $1 AND bv.viewed_at <= $2
			  AND bv.source = 'search'
			  AND bh.latitude != 0 AND bh.longitude != 0
			GROUP BY cell_lat, cell_lon
		),
		bookings_agg AS (
			SELECT
				ROUND(bh.latitude / $3) * $3 + $3 / 2.0 AS cell_lat,
				ROUND(bh.longitude / $3) * $3 + $3 / 2.0 AS cell_lon,
				COUNT(*) AS booking_count
			FROM bookings b
			JOIN bathhouses bh ON bh.id = b.bathhouse_id
			WHERE b.created_at >= $1 AND b.created_at <= $2
			  AND bh.latitude != 0 AND bh.longitude != 0
			GROUP BY cell_lat, cell_lon
		)
		SELECT
			COALESCE(s.cell_lat, d.cell_lat, ba.cell_lat) AS latitude,
			COALESCE(s.cell_lon, d.cell_lon, ba.cell_lon) AS longitude,
			COALESCE(s.listing_count, 0) AS listing_count,
			COALESCE(ba.booking_count, 0) AS booking_count,
			COALESCE(d.search_count, 0) AS search_count
		FROM supply s
		FULL OUTER JOIN demand d ON s.cell_lat = d.cell_lat AND s.cell_lon = d.cell_lon
		FULL OUTER JOIN bookings_agg ba ON COALESCE(s.cell_lat, d.cell_lat) = ba.cell_lat
			AND COALESCE(s.cell_lon, d.cell_lon) = ba.cell_lon
		ORDER BY listing_count DESC, booking_count DESC`

	rows, err := r.pool.Query(ctx, query, fromDate, toDate, cellSize)
	if err != nil {
		return nil, fmt.Errorf("heatmap data: %w", err)
	}
	defer rows.Close()

	var cells []domain.HeatmapCell
	for rows.Next() {
		var c domain.HeatmapCell
		if err := rows.Scan(&c.Latitude, &c.Longitude, &c.ListingCount, &c.BookingCount, &c.SearchCount); err != nil {
			return nil, fmt.Errorf("scan heatmap cell: %w", err)
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}

func (r *analyticsRepo) GetOwnerPerformance(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) (*domain.OwnerPerformance, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	perf := &domain.OwnerPerformance{BathhouseID: bathhouseID}

	// Bathhouse name and city
	var cityID sql.NullInt64
	err := r.pool.QueryRow(ctx,
		`SELECT name, city_id FROM bathhouses WHERE id = $1`, bathhouseID).
		Scan(&perf.BathhouseName, &cityID)
	if err != nil {
		return nil, fmt.Errorf("owner perf bathhouse: %w", err)
	}

	// Views and bookings for conversion rate
	var views, bookings int64
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(views), 0), COALESCE(SUM(bookings), 0) FROM analytics_snapshots WHERE bathhouse_id = $1 AND date >= $2 AND date <= $3`,
		bathhouseID, fromDate, toDate).Scan(&views, &bookings)
	if err != nil {
		return nil, fmt.Errorf("owner perf views: %w", err)
	}
	if views > 0 {
		perf.ConversionRate = float64(bookings) / float64(views)
	}

	// Revenue
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(revenue), 0) FROM analytics_snapshots WHERE bathhouse_id = $1 AND date >= $2 AND date <= $3`,
		bathhouseID, fromDate, toDate).Scan(&perf.Revenue)
	if err != nil {
		return nil, fmt.Errorf("owner perf revenue: %w", err)
	}

	// Occupancy rate (booked hours / available hours)
	var bookedHours sql.NullFloat64
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (end_time - start_time)) / 3600), 0)
		 FROM bookings WHERE bathhouse_id = $1 AND status IN ('confirmed', 'completed') AND start_time >= $2 AND end_time <= $3`,
		bathhouseID, fromDate, toDate).Scan(&bookedHours)
	if err != nil {
		return nil, fmt.Errorf("owner perf occupancy: %w", err)
	}
	days := toDate.Sub(fromDate).Hours() / 24
	availableHours := days * 12 // assume 12 working hours/day
	if availableHours > 0 && bookedHours.Valid {
		perf.OccupancyRate = bookedHours.Float64 / availableHours
		if perf.OccupancyRate > 1.0 {
			perf.OccupancyRate = 1.0
		}
	}

	// Average rating
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE bathhouse_id = $1`,
		bathhouseID).Scan(&perf.AvgRating)
	if err != nil {
		return nil, fmt.Errorf("owner perf rating: %w", err)
	}

	// City benchmarks (anonymous averages)
	if cityID.Valid {
		_ = r.pool.QueryRow(ctx, `
			SELECT
				COALESCE(AVG(CASE WHEN s.views > 0 THEN s.bookings::float / s.views ELSE 0 END), 0),
				COALESCE(AVG(bh.occupancy_rate), 0),
				COALESCE(AVG(bh.bayesian_rating), 0)
			FROM bathhouses bh
			LEFT JOIN (
				SELECT bathhouse_id, SUM(views) AS views, SUM(bookings) AS bookings
				FROM analytics_snapshots WHERE date >= $1 AND date <= $2
				GROUP BY bathhouse_id
			) s ON s.bathhouse_id = bh.id
			WHERE bh.city_id = $3 AND bh.status = 'active' AND bh.id != $4`,
			fromDate, toDate, cityID.Int64, bathhouseID).
			Scan(&perf.AvgCityConversionRate, &perf.AvgCityOccupancyRate, &perf.AvgCityRating)
	}

	return perf, nil
}

func (r *analyticsRepo) CountDistinctActiveUsers(ctx context.Context, from, to time.Time) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM bookings WHERE status = 'completed' AND start_time >= $1 AND start_time <= $2`,
		from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count distinct active users: %w", err)
	}
	return count, nil
}
