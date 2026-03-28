package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type customSegmentRepo struct {
	pool *pgxpool.Pool
}

func NewCustomSegmentRepository(pool *pgxpool.Pool) repository.CustomSegmentRepository {
	return &customSegmentRepo{pool: pool}
}

func (r *customSegmentRepo) Create(ctx context.Context, segment *domain.CustomSegment) error {
	now := time.Now()
	if segment.ID == uuid.Nil {
		segment.ID = uuid.New()
	}
	segment.CreatedAt = now
	segment.UpdatedAt = now

	condJSON, err := json.Marshal(segment.Conditions)
	if err != nil {
		return fmt.Errorf("marshal conditions: %w", err)
	}

	query := `INSERT INTO custom_segments (id, owner_id, bathhouse_id, name, conditions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = r.pool.Exec(ctx, query,
		segment.ID, segment.OwnerID, segment.BathhouseID, segment.Name,
		condJSON, segment.CreatedAt, segment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create custom segment: %w", err)
	}
	return nil
}

func (r *customSegmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomSegment, error) {
	query := `SELECT id, owner_id, bathhouse_id, name, conditions, created_at, updated_at FROM custom_segments WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanCustomSegment(row)
}

func (r *customSegmentRepo) Update(ctx context.Context, segment *domain.CustomSegment) error {
	segment.UpdatedAt = time.Now()

	condJSON, err := json.Marshal(segment.Conditions)
	if err != nil {
		return fmt.Errorf("marshal conditions: %w", err)
	}

	query := `UPDATE custom_segments SET name = $1, conditions = $2, bathhouse_id = $3, updated_at = $4 WHERE id = $5`
	ct, err := r.pool.Exec(ctx, query, segment.Name, condJSON, segment.BathhouseID, segment.UpdatedAt, segment.ID)
	if err != nil {
		return fmt.Errorf("update custom segment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *customSegmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM custom_segments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete custom segment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *customSegmentRepo) ListByOwner(ctx context.Context, filter domain.CustomSegmentFilter) ([]domain.CustomSegment, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argIdx))
	args = append(args, filter.OwnerID)
	argIdx++

	if filter.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("(bathhouse_id = $%d OR bathhouse_id IS NULL)", argIdx))
		args = append(args, *filter.BathhouseID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`SELECT id, owner_id, bathhouse_id, name, conditions, created_at, updated_at FROM custom_segments WHERE %s ORDER BY name ASC`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list custom segments: %w", err)
	}
	defer rows.Close()

	var segments []domain.CustomSegment
	for rows.Next() {
		s, err := scanCustomSegment(rows)
		if err != nil {
			return nil, err
		}
		segments = append(segments, *s)
	}
	return segments, rows.Err()
}

func (r *customSegmentRepo) EvaluateSegment(ctx context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter, page, pageSize int) (*domain.PaginatedResult[domain.GuestCard], error) {
	where, args, argIdx := buildSegmentEvalCondition(segment, ownerFilter)

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM guest_cards WHERE %s`, where)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count segment guests: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	offset := (page - 1) * pageSize

	dataQuery := fmt.Sprintf(`SELECT %s FROM guest_cards WHERE %s ORDER BY last_visit_at DESC LIMIT $%d OFFSET $%d`,
		guestCardColumns, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("evaluate segment guests: %w", err)
	}
	defer rows.Close()

	cards, err := scanGuestCards(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.GuestCard]{
		Items:      cards,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *customSegmentRepo) CountSegmentGuests(ctx context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter) (int64, error) {
	where, args, _ := buildSegmentEvalCondition(segment, ownerFilter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM guest_cards WHERE %s`, where)
	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count segment guests: %w", err)
	}
	return count, nil
}

func buildSegmentEvalCondition(segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter) (string, []interface{}, int) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Owner scoping
	ownerWhere, ownerArgs := buildGuestCardOwnerCondition(ownerFilter)
	conditions = append(conditions, ownerWhere)
	args = append(args, ownerArgs...)
	argIdx += len(ownerArgs)

	// Bathhouse scoping
	if segment.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("bathhouse_id = $%d", argIdx))
		args = append(args, *segment.BathhouseID)
		argIdx++
	}

	c := segment.Conditions

	if c.VisitCountMin != nil {
		conditions = append(conditions, fmt.Sprintf("visit_count >= $%d", argIdx))
		args = append(args, *c.VisitCountMin)
		argIdx++
	}
	if c.VisitCountMax != nil {
		conditions = append(conditions, fmt.Sprintf("visit_count <= $%d", argIdx))
		args = append(args, *c.VisitCountMax)
		argIdx++
	}
	if c.AvgCheckMin != nil {
		conditions = append(conditions, fmt.Sprintf("avg_check >= $%d", argIdx))
		args = append(args, *c.AvgCheckMin)
		argIdx++
	}
	if c.AvgCheckMax != nil {
		conditions = append(conditions, fmt.Sprintf("avg_check <= $%d", argIdx))
		args = append(args, *c.AvgCheckMax)
		argIdx++
	}
	if c.TotalSpentMin != nil {
		conditions = append(conditions, fmt.Sprintf("total_spent >= $%d", argIdx))
		args = append(args, *c.TotalSpentMin)
		argIdx++
	}
	if c.TotalSpentMax != nil {
		conditions = append(conditions, fmt.Sprintf("total_spent <= $%d", argIdx))
		args = append(args, *c.TotalSpentMax)
		argIdx++
	}
	if c.LastVisitDaysMin != nil {
		conditions = append(conditions, fmt.Sprintf("last_visit_at <= NOW() - INTERVAL '%d days'", *c.LastVisitDaysMin))
	}
	if c.LastVisitDaysMax != nil {
		conditions = append(conditions, fmt.Sprintf("last_visit_at >= NOW() - INTERVAL '%d days'", *c.LastVisitDaysMax))
	}
	if len(c.TagsInclude) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags @> $%d::text[]", argIdx))
		args = append(args, c.TagsInclude)
		argIdx++
	}
	if len(c.TagsExclude) > 0 {
		conditions = append(conditions, fmt.Sprintf("NOT (tags && $%d::text[])", argIdx))
		args = append(args, c.TagsExclude)
		argIdx++
	}

	// RFM conditions require a subquery with NTILE
	rfmConditions := buildRFMConditions(c)
	if len(rfmConditions) > 0 {
		// Wrap in a subquery that computes RFM scores
		rfmWhere := strings.Join(rfmConditions, " AND ")
		// Replace the entire condition set with a CTE approach
		ctePrefix := fmt.Sprintf(`id IN (
			SELECT id FROM (
				SELECT id,
					NTILE(5) OVER (ORDER BY last_visit_at ASC) AS r_score,
					NTILE(5) OVER (ORDER BY visit_count ASC) AS f_score,
					NTILE(5) OVER (ORDER BY total_spent ASC) AS m_score
				FROM guest_cards WHERE %s
			) rfm WHERE %s
		)`, strings.Join(conditions, " AND "), rfmWhere)
		conditions = append(conditions, ctePrefix)
	}

	where := strings.Join(conditions, " AND ")
	return where, args, argIdx
}

func buildRFMConditions(c domain.CustomSegmentCondition) []string {
	var conds []string
	if c.RFMRecencyMin != nil {
		conds = append(conds, fmt.Sprintf("r_score >= %d", *c.RFMRecencyMin))
	}
	if c.RFMRecencyMax != nil {
		conds = append(conds, fmt.Sprintf("r_score <= %d", *c.RFMRecencyMax))
	}
	if c.RFMFrequencyMin != nil {
		conds = append(conds, fmt.Sprintf("f_score >= %d", *c.RFMFrequencyMin))
	}
	if c.RFMFrequencyMax != nil {
		conds = append(conds, fmt.Sprintf("f_score <= %d", *c.RFMFrequencyMax))
	}
	if c.RFMMonetaryMin != nil {
		conds = append(conds, fmt.Sprintf("m_score >= %d", *c.RFMMonetaryMin))
	}
	if c.RFMMonetaryMax != nil {
		conds = append(conds, fmt.Sprintf("m_score <= %d", *c.RFMMonetaryMax))
	}
	return conds
}

func scanCustomSegment(row pgx.Row) (*domain.CustomSegment, error) {
	var s domain.CustomSegment
	var condJSON []byte
	err := row.Scan(&s.ID, &s.OwnerID, &s.BathhouseID, &s.Name, &condJSON, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan custom segment: %w", err)
	}
	if err := json.Unmarshal(condJSON, &s.Conditions); err != nil {
		return nil, fmt.Errorf("unmarshal conditions: %w", err)
	}
	return &s, nil
}
