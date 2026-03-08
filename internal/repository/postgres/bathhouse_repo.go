package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

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
			rating, review_count, images, working_hours, status, api_key, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26
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
		bh.Rating, bh.ReviewCount, imagesJSON, whJSON, bh.Status, bh.ApiKey, bh.CreatedAt, bh.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create bathhouse: %w", err)
	}
	return nil
}

func (r *bathhouseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	query := `
		SELECT DISTINCT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			CASE WHEN p.id IS NOT NULL THEN true ELSE false END as is_promoted
		FROM bathhouses
		LEFT JOIN subscriptions s ON bathhouses.id = s.bathhouse_id AND s.status = 'active'
		LEFT JOIN promotions p ON bathhouses.id = p.bathhouse_id AND p.status = 'active'
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
		SELECT DISTINCT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			CASE WHEN p.id IS NOT NULL THEN true ELSE false END as is_promoted
		FROM bathhouses
		LEFT JOIN subscriptions s ON bathhouses.id = s.bathhouse_id AND s.status = 'active'
		LEFT JOIN promotions p ON bathhouses.id = p.bathhouse_id AND p.status = 'active'
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
		SELECT DISTINCT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified, bathhouses.api_key,
			CASE WHEN p.id IS NOT NULL THEN true ELSE false END as is_promoted
		FROM bathhouses
		LEFT JOIN subscriptions s ON bathhouses.id = s.bathhouse_id AND s.status = 'active'
		LEFT JOIN promotions p ON bathhouses.id = p.bathhouse_id AND p.status = 'active'
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
			images = $18, working_hours = $19, api_key = $20, updated_at = $21
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
		imagesJSON, whJSON, bh.ApiKey, bh.UpdatedAt,
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
		escaped := strings.ReplaceAll(*filter.SearchQuery, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "%", "\\%")
		escaped = strings.ReplaceAll(escaped, "_", "\\_")
		likePattern := "%" + escaped + "%"
		conditions = append(conditions, fmt.Sprintf("(name ILIKE %s OR description ILIKE %s)", addArg(likePattern), addArg(likePattern)))
	}
	if filter.OpenNow != nil && *filter.OpenNow {
		dayArg := addArg(currentDayOfWeek())
		timeArg1 := addArg(currentTimeHHMM())
		timeArg2 := addArg(currentTimeHHMM())
		timeArg3 := addArg(currentTimeHHMM())
		timeArg4 := addArg(currentTimeHHMM())
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM jsonb_array_elements(working_hours) wh WHERE (wh->>'day_of_week')::int = %s AND ((wh->>'open_time' <= wh->>'close_time' AND wh->>'open_time' <= %s AND wh->>'close_time' > %s) OR (wh->>'open_time' > wh->>'close_time' AND (wh->>'open_time' <= %s OR wh->>'close_time' > %s))))",
			dayArg, timeArg1, timeArg2, timeArg3, timeArg4,
		))
	}
	if filter.AvailableDate != nil {
		dateStr := filter.AvailableDate.Format("2006-01-02")
		subConditions := fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM bookings b WHERE b.bathhouse_id = bathhouses.id AND b.status IN ('pending','confirmed') AND b.start_time::date = %s",
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
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = %s", addArg(string(*filter.Status))))
	} else if !filter.ShowAllStatuses {
		conditions = append(conditions, fmt.Sprintf("status = %s", addArg(string(domain.BathhouseStatusActive))))
	}

	// Geo filter using PostGIS ST_DWithin
	if filter.Latitude != nil && filter.Longitude != nil && filter.RadiusKm != nil {
		radiusMeters := *filter.RadiusKm * 1000
		conditions = append(conditions, fmt.Sprintf(
			"ST_DWithin(location, ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography, %s)",
			addArg(*filter.Longitude), addArg(*filter.Latitude), addArg(radiusMeters),
		))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build joins for subscriptions and promotions
	joinClause := `
		LEFT JOIN subscriptions s ON bathhouses.id = s.bathhouse_id AND s.status = 'active'
		LEFT JOIN promotions p ON bathhouses.id = p.bathhouse_id AND p.status = 'active'
	`

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT bathhouses.id) FROM bathhouses %s %s", joinClause, whereClause)
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bathhouses: %w", err)
	}

	// Determine ORDER BY
	orderBy := "created_at DESC"
	sortOrder := "ASC"
	if filter.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	switch filter.SortBy {
	case "price":
		// Price sort: promoted first, then premium, then by price
		orderBy = fmt.Sprintf(
			"CASE WHEN p.id IS NOT NULL THEN 0 ELSE 1 END ASC, CASE WHEN s.plan = 'premium' THEN 10 ELSE 0 END DESC, CAST(price_per_hour AS FLOAT) %s",
			sortOrder,
		)
	case "rating":
		// Rating sort: promoted first, then premium, then by rating
		orderBy = fmt.Sprintf(
			"CASE WHEN p.id IS NOT NULL THEN 0 ELSE 1 END ASC, CASE WHEN s.plan = 'premium' THEN 10 ELSE 0 END DESC, rating %s",
			sortOrder,
		)
	case "distance":
		if filter.Latitude != nil && filter.Longitude != nil {
			orderBy = fmt.Sprintf(
				"CASE WHEN p.id IS NOT NULL THEN 0 ELSE 1 END ASC, ST_Distance(location, ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography) %s",
				addArg(*filter.Longitude), addArg(*filter.Latitude), sortOrder,
			)
		}
	default:
		// Default sort: promoted first, then premium, then by created_at
		orderBy = fmt.Sprintf(
			"CASE WHEN p.id IS NOT NULL THEN 0 ELSE 1 END ASC, CASE WHEN s.plan = 'premium' THEN 10 ELSE 0 END DESC, created_at DESC",
		)
	}

	offset := (filter.Page - 1) * filter.PageSize

	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT bathhouses.id, bathhouses.owner_id, bathhouses.name, bathhouses.slug, bathhouses.description, bathhouses.address, bathhouses.city_id,
			bathhouses.latitude, bathhouses.longitude, bathhouses.price_per_hour, bathhouses.min_duration, bathhouses.max_guests,
			bathhouses.has_pool, bathhouses.has_sauna, bathhouses.has_steam_room, bathhouses.has_hot_tub, bathhouses.has_bbq, bathhouses.has_karaoke,
			bathhouses.rating, bathhouses.review_count, bathhouses.images, bathhouses.working_hours, bathhouses.status,
			bathhouses.created_at, bathhouses.updated_at, bathhouses.is_photo_verified,
			CASE WHEN p.id IS NOT NULL THEN true ELSE false END as is_promoted
		FROM bathhouses %s %s ORDER BY %s LIMIT %s OFFSET %s`,
		joinClause, whereClause, orderBy, addArg(filter.PageSize), addArg(offset),
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
			rating, review_count, images, working_hours, status, api_key, is_photo_verified, created_at, updated_at
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

func (r *bathhouseRepo) scanBathhouseFromRowWithSubscription(rows pgx.Rows) (*domain.Bathhouse, error) {
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
		&bh.Rating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified, &isPromoted,
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
		&bh.Rating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified, &bh.ApiKey, &isPromoted,
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
		&bh.Rating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.ApiKey, &bh.IsPhotoVerified, &bh.CreatedAt, &bh.UpdatedAt,
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
