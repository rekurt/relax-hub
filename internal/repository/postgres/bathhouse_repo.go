package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
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

// buildPrefixTsQuery sanitizes a user query and builds a tsquery string with
// prefix matching (:*) on each word, joined by & (AND). Returns empty string
// if no valid words remain after sanitization.
func buildPrefixTsQuery(q string) string {
	words := strings.Fields(q)
	var parts []string
	for _, w := range words {
		var cleaned strings.Builder
		for _, r := range w {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				cleaned.WriteRune(r)
			}
		}
		if cleaned.Len() > 0 {
			parts = append(parts, cleaned.String()+":*")
		}
	}
	return strings.Join(parts, " & ")
}

func (r *bathhouseRepo) List(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var (
		conditions []string
		args       []interface{}
		argIdx     = 1
	)

	addArg := func(val interface{}) string {
		args = append(args, val)
		placeholder := fmt.Sprintf("$%d", argIdx)
		argIdx++
		return placeholder
	}

	if filter.CityID != nil {
		conditions = append(conditions, fmt.Sprintf("city_id = %s", addArg(*filter.CityID)))
	}
	if filter.CitySlug != nil {
		conditions = append(conditions, fmt.Sprintf("city_id = (SELECT id FROM cities WHERE slug = %s)", addArg(*filter.CitySlug)))
	}
	if filter.PriceMin != nil {
		conditions = append(conditions, fmt.Sprintf("price_per_hour >= %s", addArg(*filter.PriceMin)))
	}
	if filter.PriceMax != nil {
		conditions = append(conditions, fmt.Sprintf("price_per_hour <= %s", addArg(*filter.PriceMax)))
	}
	if filter.MinGuests != nil {
		conditions = append(conditions, fmt.Sprintf("max_guests >= %s", addArg(*filter.MinGuests)))
	}
	if filter.HasPool != nil && *filter.HasPool {
		conditions = append(conditions, "has_pool = true")
	}
	if filter.HasSauna != nil && *filter.HasSauna {
		conditions = append(conditions, "has_sauna = true")
	}
	if filter.HasSteamRoom != nil && *filter.HasSteamRoom {
		conditions = append(conditions, "has_steam_room = true")
	}
	if filter.HasHotTub != nil && *filter.HasHotTub {
		conditions = append(conditions, "has_hot_tub = true")
	}
	if filter.HasBBQ != nil && *filter.HasBBQ {
		conditions = append(conditions, "has_bbq = true")
	}
	if filter.HasKaraoke != nil && *filter.HasKaraoke {
		conditions = append(conditions, "has_karaoke = true")
	}
	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating >= %s", addArg(*filter.MinRating)))
	}
	if filter.GuestCount != nil {
		conditions = append(conditions, fmt.Sprintf("max_guests >= %s", addArg(*filter.GuestCount)))
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		q := strings.TrimSpace(*filter.SearchQuery)
		prefixQ := buildPrefixTsQuery(q)
		// Full-text search: exact words via plainto_tsquery, prefix via to_tsquery(:*), trigram fallback
		if prefixQ != "" {
			conditions = append(conditions, fmt.Sprintf(
				"(search_vector @@ plainto_tsquery('russian', %s) OR search_vector @@ to_tsquery('russian', %s) OR similarity(name, %s) > 0.2 OR similarity(description, %s) > 0.1)",
				addArg(q), addArg(prefixQ), addArg(q), addArg(q),
			))
		} else {
			conditions = append(conditions, fmt.Sprintf(
				"(search_vector @@ plainto_tsquery('russian', %s) OR similarity(name, %s) > 0.2 OR similarity(description, %s) > 0.1)",
				addArg(q), addArg(q), addArg(q),
			))
		}
	}
	if filter.OpenNow != nil && *filter.OpenNow {
		dayArg := addArg(currentDayOfWeek())
		timeArg := addArg(currentTimeHHMM())
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM jsonb_array_elements(working_hours) wh WHERE (wh->>'day_of_week')::int = %s AND ((wh->>'open_time' <= wh->>'close_time' AND wh->>'open_time' <= %s AND wh->>'close_time' > %s) OR (wh->>'open_time' > wh->>'close_time' AND (wh->>'open_time' <= %s OR wh->>'close_time' > %s))))",
			dayArg, timeArg, timeArg, timeArg, timeArg,
		))
	}
	if filter.AvailableDate != nil {
		dateStr := filter.AvailableDate.Format("2006-01-02")
		subConditions := fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM bookings b WHERE b.bathhouse_id = bathhouses.id AND b.status IN ('pending','pending_owner','confirmed') AND b.start_time::date = %s",
			addArg(dateStr),
		)
		if filter.AvailableTimeFrom != nil {
			subConditions += fmt.Sprintf(" AND b.end_time::time > %s", addArg(*filter.AvailableTimeFrom+":00"))
		}
		if filter.AvailableTimeTo != nil {
			subConditions += fmt.Sprintf(" AND b.start_time::time < %s", addArg(*filter.AvailableTimeTo+":00"))
		}
		subConditions += ")"
		conditions = append(conditions, subConditions)
	}
	if filter.LastMinute != nil && *filter.LastMinute {
		conditions = append(conditions, "last_minute_enabled = true")
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("bathhouses.status = %s", addArg(string(*filter.Status))))
	} else if !filter.ShowAllStatuses {
		conditions = append(conditions, fmt.Sprintf("bathhouses.status = %s", addArg(string(domain.BathhouseStatusActive))))
	}

	// Geo filter using PostGIS ST_DWithin
	if filter.Latitude != nil && filter.Longitude != nil && filter.RadiusKm != nil {
		radiusMeters := *filter.RadiusKm * 1000
		conditions = append(conditions, fmt.Sprintf(
			"ST_DWithin(location, ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography, %s)",
			addArg(*filter.Longitude), addArg(*filter.Latitude), addArg(radiusMeters),
		))
	}

	// Isochrone polygon filter using PostGIS ST_Within
	if filter.IsochroneWKT != nil && *filter.IsochroneWKT != "" {
		conditions = append(conditions, fmt.Sprintf(
			"ST_Within(location::geometry, ST_GeomFromText(%s, 4326))",
			addArg(*filter.IsochroneWKT),
		))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Promotion check as EXISTS subquery (avoids JOIN duplicates and DISTINCT issues)
	promotionExists := `EXISTS (SELECT 1 FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active')`

	// Auction-weighted promotion boost: normalized daily_bid_kopecks (higher bid = more visibility).
	// LEAST caps at 1.0, reference bid is 10000 kopecks (100 rub/day).
	promotionBoost := `COALESCE((SELECT LEAST(daily_bid_kopecks::float / 10000.0, 1.0) FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active' ORDER BY daily_bid_kopecks DESC LIMIT 1), 0)`

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bathhouses %s", whereClause)
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bathhouses: %w", err)
	}

	// Determine ORDER BY
	// Composite ranking: relevance*0.30 + bayesian_rating*0.25 + conversion_rate*0.20 + occupancy_rate*0.15 + promotion_boost*0.10
	// Promotion boost is auction-weighted: daily_bid / 10000 (capped at 1.0)
	compositeRank := fmt.Sprintf(`(
		COALESCE(bathhouses.conversion_rate, 0) * 0.20 +
		COALESCE(bathhouses.bayesian_rating, 0) / 5.0 * 0.25 +
		COALESCE(bathhouses.occupancy_rate, 0) * 0.15 +
		%s * 0.10
	)`, promotionBoost)

	// Promoted ordering: only top 3 promoted items (by bid) get priority positioning per BRD FR-043
	// promo_rank is computed via ROW_NUMBER in the SELECT; this expression caps promoted priority at 3
	promotedCap := "CASE WHEN is_promoted AND promo_rank <= 3 THEN 0 ELSE 1 END"

	orderBy := "created_at DESC"

	switch filter.SortBy {
	case "relevance":
		if filter.SearchQuery != nil && *filter.SearchQuery != "" {
			q := strings.TrimSpace(*filter.SearchQuery)
			// When searching: composite = text relevance * 0.30 + other factors
			orderBy = fmt.Sprintf(
				"%s, (ts_rank(search_vector, plainto_tsquery('russian', %s)) * 0.30 + %s) DESC, similarity(name, %s) DESC",
				promotedCap, addArg(q), compositeRank, addArg(q),
			)
		} else {
			orderBy = promotedCap + ", " + compositeRank + " DESC, created_at DESC"
		}
	case "price_asc":
		orderBy = promotedCap + ", price_per_hour ASC"
	case "price_desc":
		orderBy = promotedCap + ", price_per_hour DESC"
	case "price":
		// Legacy support: use sort_order
		sortDir := "ASC"
		if filter.SortOrder == "desc" {
			sortDir = "DESC"
		}
		orderBy = fmt.Sprintf(
			"%s, price_per_hour %s",
			promotedCap, sortDir,
		)
	case "rating":
		orderBy = promotedCap + ", rating DESC, review_count DESC"
	case "distance":
		if filter.Latitude != nil && filter.Longitude != nil {
			sortDir := "ASC"
			if filter.SortOrder == "desc" {
				sortDir = "DESC"
			}
			orderBy = fmt.Sprintf(
				"%s, ST_Distance(location, ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography) %s",
				promotedCap, addArg(*filter.Longitude), addArg(*filter.Latitude), sortDir,
			)
		}
	case "newest":
		orderBy = promotedCap + ", created_at DESC"
	default:
		// Default to relevance-based ranking
		filter.SortBy = "relevance"
		if filter.SearchQuery != nil && *filter.SearchQuery != "" {
			q := strings.TrimSpace(*filter.SearchQuery)
			orderBy = fmt.Sprintf(
				"%s, (ts_rank(search_vector, plainto_tsquery('russian', %s)) * 0.30 + %s) DESC, similarity(name, %s) DESC",
				promotedCap, addArg(q), compositeRank, addArg(q),
			)
		} else {
			orderBy = promotedCap + ", " + compositeRank + " DESC, created_at DESC"
		}
	}

	offset := (filter.Page - 1) * filter.PageSize

	// Bid-based rank for promoted items: ROW_NUMBER() partitioned by promoted status, ordered by bid descending.
	// Only the top 3 promoted items (by bid) get priority positioning in search results.
	promoRankExpr := fmt.Sprintf(`CASE WHEN %s
			THEN ROW_NUMBER() OVER (PARTITION BY (%s) ORDER BY COALESCE((SELECT daily_bid_kopecks FROM promotions WHERE bathhouse_id = bathhouses.id AND status = 'active' ORDER BY daily_bid_kopecks DESC LIMIT 1), 0) DESC)
		END`, promotionExists, promotionExists)

	selectQuery := fmt.Sprintf(`
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
			%s as is_promoted,
			%s as promo_rank
		FROM bathhouses %s ORDER BY %s LIMIT %s OFFSET %s`,
		promotionExists, promoRankExpr, whereClause, orderBy, addArg(filter.PageSize), addArg(offset),
	)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list bathhouses: %w", err)
	}
	defer rows.Close()

	var bathhouses []domain.Bathhouse
	for rows.Next() {
		bh, err := r.scanBathhouseFromRowWithSubscription(rows)
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
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
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

func (r *bathhouseRepo) scanBathhouseMinimal(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status,
		&bh.CreatedAt, &bh.UpdatedAt, &bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&bh.LowResponseRateSince,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}
	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}
	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithSubscription(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
		isPromoted bool
		promoRank  *int64 // used for ORDER BY only, ignored in domain
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&isPromoted,
		&promoRank,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	bh.IsPromoted = isPromoted

	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithAPIKey(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
		isPromoted bool
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified, &bh.ApiKey,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&isPromoted,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	bh.IsPromoted = isPromoted

	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithAPIKeyOnly(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.ApiKey, &bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&bh.CreatedAt, &bh.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	return &bh, nil
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

func (r *bathhouseRepo) SuggestNames(ctx context.Context, filter repository.SuggestionFilter) ([]string, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	// Escape LIKE metacharacters to prevent pattern injection
	escapedQuery := strings.NewReplacer("%", `\%`, "_", `\_`).Replace(filter.Query)

	query := `
		SELECT DISTINCT name
		FROM bathhouses
		WHERE status = 'active'
		  AND (similarity(name, $1) > 0.2 OR name ILIKE '%' || $2 || '%' ESCAPE '\')
		ORDER BY similarity(name, $1) DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, filter.Query, escapedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("suggest names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan suggestion: %w", err)
		}
		names = append(names, name)
	}

	return names, rows.Err()
}

// GetAreaAvgPrice returns the average price_per_hour for active bathhouses in the same city
// within a 5 km radius. Returns 0 if no qualifying bathhouses exist.
func (r *bathhouseRepo) GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error) {
	query := `
		SELECT COALESCE(AVG(price_per_hour), 0)::bigint
		FROM bathhouses
		WHERE status = 'active'
			AND city_id = $1
			AND ST_DWithin(
				location,
				ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
				5000
			)`
	var avg int64
	err := r.pool.QueryRow(ctx, query, cityID, lng, lat).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("get area avg price: %w", err)
	}
	return avg, nil
}
