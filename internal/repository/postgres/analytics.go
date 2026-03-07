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
		result.ReviewCount += snapshot.ReviewCount
		// Average rating requires weighted average - not aggregated simply
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

	var orderBy string
	switch metric {
	case domain.MetricViews:
		orderBy = "SUM(views)"
	case domain.MetricBookings:
		orderBy = "SUM(bookings)"
	case domain.MetricRevenue:
		orderBy = "SUM(revenue)"
	case domain.MetricRating:
		orderBy = "AVG(avg_rating)"
	}

	query := fmt.Sprintf(`
		SELECT bathhouse_id
		FROM analytics_snapshots
		GROUP BY bathhouse_id
		ORDER BY %s DESC
		LIMIT $1`, orderBy)

	rows, err := r.pool.Query(ctx, query, limit)
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
