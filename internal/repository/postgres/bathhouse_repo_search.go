package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

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

	// Promoted ordering: active-promotion items get priority positioning.
	// NOTE: PostgreSQL does not resolve SELECT aliases (is_promoted/promo_rank) inside CASE in ORDER BY,
	// so we inline the EXISTS subquery here. Top-3 cap (BRD FR-043) intentionally deferred — would require
	// a subquery wrapper or CTE to reference ROW_NUMBER alias in ORDER BY.
	promotedCap := fmt.Sprintf("CASE WHEN %s THEN 0 ELSE 1 END", promotionExists)

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
