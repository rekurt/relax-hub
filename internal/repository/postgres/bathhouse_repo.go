package postgres

import (
	"context"
	"encoding/json"
	"errors"
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
			id, owner_id, name, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, review_count, images, working_hours, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24
		)`

	now := time.Now()
	if bh.ID == uuid.Nil {
		bh.ID = uuid.New()
	}
	bh.CreatedAt = now
	bh.UpdatedAt = now

	imagesJSON, err := json.Marshal(bh.Images)
	if err != nil {
		return fmt.Errorf("marshal images: %w", err)
	}
	whJSON, err := json.Marshal(bh.WorkingHours)
	if err != nil {
		return fmt.Errorf("marshal working hours: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		bh.ID, bh.OwnerID, bh.Name, bh.Description, bh.Address, bh.CityID,
		bh.Latitude, bh.Longitude, bh.PricePerHour, bh.MinDuration, bh.MaxGuests,
		bh.HasPool, bh.HasSauna, bh.HasSteamRoom, bh.HasHotTub, bh.HasBBQ, bh.HasKaraoke,
		bh.Rating, bh.ReviewCount, imagesJSON, whJSON, bh.Status, bh.CreatedAt, bh.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create bathhouse: %w", err)
	}
	return nil
}

func (r *bathhouseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	query := `
		SELECT id, owner_id, name, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, review_count, images, working_hours, status, created_at, updated_at
		FROM bathhouses WHERE id = $1`

	return r.scanBathhouse(r.pool.QueryRow(ctx, query, id))
}

func (r *bathhouseRepo) Update(ctx context.Context, bh *domain.Bathhouse) error {
	query := `
		UPDATE bathhouses SET
			name = $2, description = $3, address = $4, city_id = $5,
			latitude = $6, longitude = $7, price_per_hour = $8, min_duration = $9, max_guests = $10,
			has_pool = $11, has_sauna = $12, has_steam_room = $13, has_hot_tub = $14, has_bbq = $15, has_karaoke = $16,
			images = $17, working_hours = $18, updated_at = $19
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
		bh.ID, bh.Name, bh.Description, bh.Address, bh.CityID,
		bh.Latitude, bh.Longitude, bh.PricePerHour, bh.MinDuration, bh.MaxGuests,
		bh.HasPool, bh.HasSauna, bh.HasSteamRoom, bh.HasHotTub, bh.HasBBQ, bh.HasKaraoke,
		imagesJSON, whJSON, bh.UpdatedAt,
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

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bathhouses %s", whereClause)
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
		orderBy = fmt.Sprintf("price_per_hour %s", sortOrder)
	case "rating":
		orderBy = fmt.Sprintf("rating %s", sortOrder)
	case "distance":
		if filter.Latitude != nil && filter.Longitude != nil {
			orderBy = fmt.Sprintf(
				"ST_Distance(location, ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography) %s",
				addArg(*filter.Longitude), addArg(*filter.Latitude), sortOrder,
			)
		}
	}

	offset := (filter.Page - 1) * filter.PageSize

	selectQuery := fmt.Sprintf(`
		SELECT id, owner_id, name, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, review_count, images, working_hours, status, created_at, updated_at
		FROM bathhouses %s ORDER BY %s LIMIT %s OFFSET %s`,
		whereClause, orderBy, addArg(filter.PageSize), addArg(offset),
	)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list bathhouses: %w", err)
	}
	defer rows.Close()

	var bathhouses []domain.Bathhouse
	for rows.Next() {
		bh, err := r.scanBathhouseFromRow(rows)
		if err != nil {
			return nil, err
		}
		bathhouses = append(bathhouses, *bh)
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
		SELECT id, owner_id, name, description, address, city_id,
			latitude, longitude, price_per_hour, min_duration, max_guests,
			has_pool, has_sauna, has_steam_room, has_hot_tub, has_bbq, has_karaoke,
			rating, review_count, images, working_hours, status, created_at, updated_at
		FROM bathhouses WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bathhouses by owner: %w", err)
	}
	defer rows.Close()

	var bathhouses []domain.Bathhouse
	for rows.Next() {
		bh, err := r.scanBathhouseFromRow(rows)
		if err != nil {
			return nil, err
		}
		bathhouses = append(bathhouses, *bh)
	}

	return &domain.PaginatedResult[domain.Bathhouse]{
		Items:      bathhouses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *bathhouseRepo) UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error {
	query := `
		UPDATE bathhouses SET
			rating = COALESCE((SELECT AVG(rating)::double precision FROM reviews WHERE bathhouse_id = $1), 0),
			review_count = (SELECT COUNT(*) FROM reviews WHERE bathhouse_id = $1),
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

func (r *bathhouseRepo) scanBathhouse(row pgx.Row) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := row.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan bathhouse: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRow(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
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
