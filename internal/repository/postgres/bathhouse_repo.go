package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type bathhouseRepo struct {
	pool *pgxpool.Pool
}

func NewBathhouseRepository(pool *pgxpool.Pool) repository.BathhouseRepository {
	return &bathhouseRepo{pool: pool}
}

func (r *bathhouseRepo) Create(ctx context.Context, bh *domain.Bathhouse) error {
	query := `
		INSERT INTO bathhouses (
			id, owner_id, name, slug, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, review_count, images, working_hours, status, api_key,
			long_session_threshold_hours, long_session_discount_percent, base_capacity, extra_guest_surcharge,
			last_minute_enabled, last_minute_discount_percent, last_minute_hours_threshold,
			buffer_minutes, lead_time_hours, max_advance_days,
			booking_mode, request_timeout,
			cancellation_policy, security_deposit_percent,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28,
			$29, $30, $31,
			$32, $33, $34,
			$35, $36,
			$37, $38,
			$39, $40
		)`

	if bh.ID == uuid.Nil {
		bh.ID = uuid.New()
	}

	imagesJSON, err := json.Marshal(bh.Images)
	if err != nil {
		return fmt.Errorf("marshal images: %w", err)
	}
	whJSON, err := json.Marshal(bh.WorkingHours)
	if err != nil {
		return fmt.Errorf("marshal working hours: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		bh.ID, bh.OwnerID, bh.Name, bh.Slug, bh.Description, bh.Address, bh.CityID,
		bh.Latitude, bh.Longitude, bh.PricePerHour, bh.MinDuration, bh.MaxGuests,
		bh.HasPool, bh.HasSauna, bh.HasSteamRoom, bh.HasHotTub, bh.HasBBQ, bh.HasKaraoke,
		bh.Rating, bh.ReviewCount, imagesJSON, whJSON, bh.Status, bh.ApiKey,
		bh.LongSessionThresholdHours, bh.LongSessionDiscountPercent, bh.BaseCapacity, bh.ExtraGuestSurcharge,
		bh.LastMinuteEnabled, bh.LastMinuteDiscountPercent, bh.LastMinuteHoursThreshold,
		bh.BufferMinutes, bh.LeadTimeHours, bh.MaxAdvanceDays,
		bh.BookingMode, bh.RequestTimeout,
		bh.CancellationPolicy, bh.SecurityDepositPercent,
		bh.CreatedAt, bh.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create bathhouse: %w", err)
	}
	return nil
}

func (r *bathhouseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	query := `
		SELECT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.bayesian_rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			bathhouses.long_session_threshold_hours, bathhouses.long_session_discount_percent, bathhouses.base_capacity, bathhouses.extra_guest_surcharge,
			bathhouses.last_minute_enabled, bathhouses.last_minute_discount_percent, bathhouses.last_minute_hours_threshold,
			bathhouses.buffer_minutes, bathhouses.lead_time_hours, bathhouses.max_advance_days,
			bathhouses.booking_mode, bathhouses.request_timeout,
			bathhouses.response_rate, bathhouses.avg_response_time_minutes,
			bathhouses.cancellation_policy, bathhouses.security_deposit_percent,
			EXISTS (SELECT 1 FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active') as is_promoted,
			NULL::bigint as promo_rank
		FROM bathhouses
		WHERE bathhouses.id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse by id: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return r.scanBathhouseFromRowWithSubscription(rows)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get bathhouse by id: %w", err)
	}

	return nil, domain.ErrNotFound
}

func (r *bathhouseRepo) GetBySlug(ctx context.Context, slug string) (*domain.Bathhouse, error) {
	query := `
		SELECT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.bayesian_rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			bathhouses.long_session_threshold_hours, bathhouses.long_session_discount_percent, bathhouses.base_capacity, bathhouses.extra_guest_surcharge,
			bathhouses.last_minute_enabled, bathhouses.last_minute_discount_percent, bathhouses.last_minute_hours_threshold,
			bathhouses.buffer_minutes, bathhouses.lead_time_hours, bathhouses.max_advance_days,
			bathhouses.booking_mode, bathhouses.request_timeout,
			bathhouses.response_rate, bathhouses.avg_response_time_minutes,
			bathhouses.cancellation_policy, bathhouses.security_deposit_percent,
			EXISTS (SELECT 1 FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active') as is_promoted,
			NULL::bigint as promo_rank
		FROM bathhouses
		WHERE bathhouses.slug = $1`

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse by slug: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return r.scanBathhouseFromRowWithSubscription(rows)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get bathhouse by slug: %w", err)
	}

	return nil, domain.ErrNotFound
}

func (r *bathhouseRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bathhouses WHERE slug = $1)`, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return exists, nil
}

func (r *bathhouseRepo) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error) {
	query := `
		SELECT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.bayesian_rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified, bathhouses.api_key,
			bathhouses.long_session_threshold_hours, bathhouses.long_session_discount_percent, bathhouses.base_capacity, bathhouses.extra_guest_surcharge,
			bathhouses.last_minute_enabled, bathhouses.last_minute_discount_percent, bathhouses.last_minute_hours_threshold,
			bathhouses.buffer_minutes, bathhouses.lead_time_hours, bathhouses.max_advance_days,
			bathhouses.booking_mode, bathhouses.request_timeout,
			bathhouses.response_rate, bathhouses.avg_response_time_minutes,
			bathhouses.cancellation_policy, bathhouses.security_deposit_percent,
			EXISTS (SELECT 1 FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active') as is_promoted
		FROM bathhouses
		WHERE bathhouses.api_key = $1`

	rows, err := r.pool.Query(ctx, query, apiKey)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse by api_key: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return r.scanBathhouseFromRowWithAPIKey(rows)
	}

	return nil, domain.ErrNotFound
}

func (r *bathhouseRepo) Update(ctx context.Context, bh *domain.Bathhouse) error {
	query := `
		UPDATE bathhouses SET
			name = $2, slug = $3, description = $4, address = $5, city_id = $6,
			latitude = $7, longitude = $8, price_per_hour = $9, min_duration = $10, max_guests = $11,
			has_pool = $12, has_sauna = $13, has_steam_room = $14, has_hot_tub = $15, has_bbq = $16, has_karaoke = $17,
			images = $18, working_hours = $19, api_key = $20, updated_at = $21, status = $22,
			long_session_threshold_hours = $23, long_session_discount_percent = $24, base_capacity = $25, extra_guest_surcharge = $26,
			last_minute_enabled = $27, last_minute_discount_percent = $28, last_minute_hours_threshold = $29,
			buffer_minutes = $30, lead_time_hours = $31, max_advance_days = $32,
			booking_mode = $33, request_timeout = $34,
			cancellation_policy = $35, security_deposit_percent = $36
		WHERE id = $1`

	bh.UpdatedAt = time.Now()

	imagesJSON, err := json.Marshal(bh.Images)
	if err != nil {
		return fmt.Errorf("marshal images: %w", err)
	}
	whJSON, err := json.Marshal(bh.WorkingHours)
	if err != nil {
		return fmt.Errorf("marshal working hours: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query,
		bh.ID, bh.Name, bh.Slug, bh.Description, bh.Address, bh.CityID,
		bh.Latitude, bh.Longitude, bh.PricePerHour, bh.MinDuration, bh.MaxGuests,
		bh.HasPool, bh.HasSauna, bh.HasSteamRoom, bh.HasHotTub, bh.HasBBQ, bh.HasKaraoke,
		imagesJSON, whJSON, bh.ApiKey, bh.UpdatedAt, bh.Status,
		bh.LongSessionThresholdHours, bh.LongSessionDiscountPercent, bh.BaseCapacity, bh.ExtraGuestSurcharge,
		bh.LastMinuteEnabled, bh.LastMinuteDiscountPercent, bh.LastMinuteHoursThreshold,
		bh.BufferMinutes, bh.LeadTimeHours, bh.MaxAdvanceDays,
		bh.BookingMode, bh.RequestTimeout,
		bh.CancellationPolicy, bh.SecurityDepositPercent,
	)
	if err != nil {
		return fmt.Errorf("update bathhouse: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bathhouses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete bathhouse: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bathhouses WHERE owner_id = $1`, ownerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bathhouses by owner: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, owner_id, name, slug, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, bayesian_rating, review_count, images, working_hours, status, api_key, is_photo_verified,
			long_session_threshold_hours, long_session_discount_percent, base_capacity, extra_guest_surcharge,
			last_minute_enabled, last_minute_discount_percent, last_minute_hours_threshold,
			buffer_minutes, lead_time_hours, max_advance_days,
			booking_mode, request_timeout,
			response_rate, avg_response_time_minutes,
			cancellation_policy, security_deposit_percent,
			created_at, updated_at
		FROM bathhouses WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bathhouses by owner: %w", err)
	}
	defer rows.Close()

	var bathhouses []domain.Bathhouse
	for rows.Next() {
		bh, err := r.scanBathhouseFromRowWithAPIKeyOnly(rows)
		if err != nil {
			return nil, err
		}
		bathhouses = append(bathhouses, *bh)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bathhouse rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Bathhouse]{
		Items:      bathhouses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *bathhouseRepo) ListIDsByOwner(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT id FROM bathhouses WHERE owner_id = $1`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list bathhouse IDs by owner: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan bathhouse ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bathhouse ID rows: %w", err)
	}
	return ids, nil
}

func (r *bathhouseRepo) UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error {
	query := `
		UPDATE bathhouses SET
			rating = COALESCE((SELECT AVG(rating)::double precision FROM reviews WHERE bathhouse_id = $1 AND status = 'approved'), 0),
			review_count = (SELECT COUNT(*) FROM reviews WHERE bathhouse_id = $1 AND status = 'approved'),
			updated_at = $2
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, bathhouseID, time.Now())
	if err != nil {
		return fmt.Errorf("update bathhouse rating: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) UpdateBayesianRating(ctx context.Context, bathhouseID uuid.UUID, bayesianRating float64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bathhouses SET bayesian_rating = $2, updated_at = $3 WHERE id = $1`,
		bathhouseID, bayesianRating, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update bayesian rating: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error {
	query := `UPDATE bathhouses SET status = $2, updated_at = $3 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("update bathhouse status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) UpdatePhotoVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	query := `UPDATE bathhouses SET is_photo_verified = $2, updated_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, verified, time.Now())
	if err != nil {
		return fmt.Errorf("update bathhouse photo verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) GetCalendarToken(ctx context.Context, bathhouseID uuid.UUID) (string, error) {
	var token *string
	err := r.pool.QueryRow(ctx, `SELECT calendar_token FROM bathhouses WHERE id = $1`, bathhouseID).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("get calendar token: %w", err)
	}
	if token == nil {
		return "", nil
	}
	return *token, nil
}

func (r *bathhouseRepo) SetCalendarToken(ctx context.Context, bathhouseID uuid.UUID, token string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE bathhouses SET calendar_token = $2 WHERE id = $1`, bathhouseID, token)
	if err != nil {
		return fmt.Errorf("set calendar token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) GetByCalendarToken(ctx context.Context, token string) (*domain.Bathhouse, error) {
	query := `
		SELECT id, owner_id, name, slug, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, bayesian_rating, review_count, images, working_hours, status,
			created_at, updated_at, is_photo_verified,
			long_session_threshold_hours, long_session_discount_percent, base_capacity, extra_guest_surcharge,
			last_minute_enabled, last_minute_discount_percent, last_minute_hours_threshold,
			buffer_minutes, lead_time_hours, max_advance_days,
			booking_mode, request_timeout,
			response_rate, avg_response_time_minutes,
			cancellation_policy, security_deposit_percent,
			low_response_rate_since
		FROM bathhouses
		WHERE calendar_token = $1`

	rows, err := r.pool.Query(ctx, query, token)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse by calendar token: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return r.scanBathhouseMinimal(rows)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get bathhouse by calendar token: %w", err)
	}
	return nil, domain.ErrNotFound
}

func (r *bathhouseRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `UPDATE bathhouses SET view_count = view_count + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("increment view count: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) UpdateRankingFields(ctx context.Context, id uuid.UUID, conversionRate, occupancyRate float64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bathhouses SET conversion_rate = $2, occupancy_rate = $3, updated_at = $4 WHERE id = $1`,
		id, conversionRate, occupancyRate, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update ranking fields: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) UpdateResponseRate(ctx context.Context, id uuid.UUID, responseRate float64, avgResponseMinutes int, lowResponseRateSince *time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bathhouses SET response_rate = $2, avg_response_time_minutes = $3, low_response_rate_since = $4, updated_at = $5 WHERE id = $1`,
		id, responseRate, avgResponseMinutes, lowResponseRateSince, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update response rate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bathhouseRepo) ListRequestModeBathhouses(ctx context.Context) ([]domain.Bathhouse, error) {
	query := `
		SELECT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.bayesian_rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			bathhouses.long_session_threshold_hours, bathhouses.long_session_discount_percent, bathhouses.base_capacity, bathhouses.extra_guest_surcharge,
			bathhouses.last_minute_enabled, bathhouses.last_minute_discount_percent, bathhouses.last_minute_hours_threshold,
			bathhouses.buffer_minutes, bathhouses.lead_time_hours, bathhouses.max_advance_days,
			bathhouses.booking_mode, bathhouses.request_timeout,
			bathhouses.response_rate, bathhouses.avg_response_time_minutes,
			bathhouses.cancellation_policy, bathhouses.security_deposit_percent,
			bathhouses.low_response_rate_since
		FROM bathhouses
		WHERE bathhouses.booking_mode = 'request'
			AND bathhouses.status = 'active'`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list request mode bathhouses: %w", err)
	}
	defer rows.Close()

	var result []domain.Bathhouse
	for rows.Next() {
		bh, err := r.scanBathhouseMinimal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *bh)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate request mode bathhouse rows: %w", err)
	}
	return result, nil
}

