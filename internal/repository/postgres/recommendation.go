package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type recommendationRepo struct {
	pool *pgxpool.Pool
}

func NewRecommendationRepository(pool *pgxpool.Pool) repository.RecommendationRepository {
	return &recommendationRepo{pool: pool}
}

func (r *recommendationRepo) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	query := `
		SELECT user_id, preferred_city_id, price_range_min, price_range_max,
		       prefer_pool, prefer_sauna, prefer_steam_room, prefer_hot_tub,
		       prefer_bbq, prefer_karaoke, updated_at
		FROM user_preferences
		WHERE user_id = $1`

	var prefs domain.UserPreferences
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.PreferredCityID, &prefs.PriceRangeMin, &prefs.PriceRangeMax,
		&prefs.PreferPool, &prefs.PreferSauna, &prefs.PreferSteamRoom, &prefs.PreferHotTub,
		&prefs.PreferBBQ, &prefs.PreferKaraoke, &prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user preferences: %w", err)
	}
	return &prefs, nil
}

func (r *recommendationRepo) SaveUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	if err := prefs.Validate(); err != nil {
		return err
	}

	if prefs.UpdatedAt.IsZero() {
		prefs.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO user_preferences (user_id, preferred_city_id, price_range_min, price_range_max,
		                              prefer_pool, prefer_sauna, prefer_steam_room, prefer_hot_tub,
		                              prefer_bbq, prefer_karaoke, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id) DO UPDATE SET
			preferred_city_id = $2,
			price_range_min = $3,
			price_range_max = $4,
			prefer_pool = $5,
			prefer_sauna = $6,
			prefer_steam_room = $7,
			prefer_hot_tub = $8,
			prefer_bbq = $9,
			prefer_karaoke = $10,
			updated_at = $11`

	_, err := r.pool.Exec(ctx, query,
		prefs.UserID, prefs.PreferredCityID, prefs.PriceRangeMin, prefs.PriceRangeMax,
		prefs.PreferPool, prefs.PreferSauna, prefs.PreferSteamRoom, prefs.PreferHotTub,
		prefs.PreferBBQ, prefs.PreferKaraoke, prefs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save user preferences: %w", err)
	}
	return nil
}

func (r *recommendationRepo) RecordActivity(ctx context.Context, activity *domain.UserActivity) error {
	if err := activity.Validate(); err != nil {
		return err
	}

	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO user_activity (id, user_id, bathhouse_id, type, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query,
		activity.ID, activity.UserID, activity.BathhouseID, string(activity.Type), activity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("record activity: %w", err)
	}
	return nil
}

func (r *recommendationRepo) GetUserBookedBathhouses(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT DISTINCT bathhouse_id
		FROM bookings
		WHERE user_id = $1 AND status IN ('confirmed', 'completed')
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get user booked bathhouses: %w", err)
	}
	defer rows.Close()

	var bathhouseIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan bathhouse id: %w", err)
		}
		bathhouseIDs = append(bathhouseIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booked bathhouses: %w", err)
	}

	return bathhouseIDs, nil
}

func (r *recommendationRepo) GetSimilarUsers(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 20
	}

	// Find users who have booked the same bathhouses
	query := `
		SELECT DISTINCT u.user_id
		FROM (
			SELECT bathhouse_id FROM bookings WHERE user_id = $1 AND status IN ('confirmed', 'completed')
		) my_bookings
		JOIN bookings b ON my_bookings.bathhouse_id = b.bathhouse_id
		JOIN (SELECT DISTINCT user_id FROM bookings WHERE user_id != $1) u ON b.user_id = u.user_id
		WHERE b.status IN ('confirmed', 'completed')
		GROUP BY u.user_id
		ORDER BY COUNT(*) DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get similar users: %w", err)
	}
	defer rows.Close()

	var similarUserIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan similar user id: %w", err)
		}
		similarUserIDs = append(similarUserIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate similar users: %w", err)
	}

	return similarUserIDs, nil
}

func (r *recommendationRepo) GetPopularBathhouses(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 20
	}

	// Popular bathhouses by number of confirmed bookings and rating
	query := `
		SELECT b.id
		FROM bathhouses b
		WHERE b.city_id = $1
		  AND b.status = 'active'
		ORDER BY b.rating DESC, (
			SELECT COUNT(*) FROM bookings
			WHERE bathhouse_id = b.id AND status IN ('confirmed', 'completed')
		) DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, cityID, limit)
	if err != nil {
		return nil, fmt.Errorf("get popular bathhouses: %w", err)
	}
	defer rows.Close()

	var bathhouseIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan popular bathhouse id: %w", err)
		}
		bathhouseIDs = append(bathhouseIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate popular bathhouses: %w", err)
	}

	return bathhouseIDs, nil
}

func (r *recommendationRepo) GetSimilarBathhouses(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get reference bathhouse
	query := `SELECT city_id FROM bathhouses WHERE id = $1`
	var cityID int64
	err := r.pool.QueryRow(ctx, query, bathhouseID).Scan(&cityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get similar bathhouses - get city: %w", err)
	}

	// Find similar bathhouses: same city, similar price, similar amenities
	similarQuery := `
		SELECT id
		FROM bathhouses
		WHERE id != $1 AND city_id = $2 AND status = 'active'
		ORDER BY rating DESC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, similarQuery, bathhouseID, cityID, limit)
	if err != nil {
		return nil, fmt.Errorf("get similar bathhouses: %w", err)
	}
	defer rows.Close()

	var similarIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan similar bathhouse id: %w", err)
		}
		similarIDs = append(similarIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate similar bathhouses: %w", err)
	}

	return similarIDs, nil
}
