package postgres

import (
	"context"
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

type guestCardRepo struct {
	pool *pgxpool.Pool
}

func NewGuestCardRepository(pool *pgxpool.Pool) repository.GuestCardRepository {
	return &guestCardRepo{pool: pool}
}

var guestCardColumns = `id, owner_id, client_id, bathhouse_id, first_visit_at, last_visit_at, visit_count, total_spent, avg_check, notes, tags, created_at, updated_at`

func scanGuestCard(row pgx.Row) (*domain.GuestCard, error) {
	var c domain.GuestCard
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.ClientID, &c.BathhouseID,
		&c.FirstVisitAt, &c.LastVisitAt, &c.VisitCount,
		&c.TotalSpent, &c.AvgCheck, &c.Notes, &c.Tags,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanGuestCards(rows pgx.Rows) ([]domain.GuestCard, error) {
	var cards []domain.GuestCard
	for rows.Next() {
		c, err := scanGuestCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan guest card: %w", err)
		}
		cards = append(cards, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guest card rows: %w", err)
	}
	return cards, nil
}

func (r *guestCardRepo) Upsert(ctx context.Context, card *domain.GuestCard) error {
	now := time.Now()
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	if card.UpdatedAt.IsZero() {
		card.UpdatedAt = now
	}
	if card.FirstVisitAt.IsZero() {
		card.FirstVisitAt = now
	}
	if card.LastVisitAt.IsZero() {
		card.LastVisitAt = now
	}

	query := `
		INSERT INTO guest_cards (id, owner_id, client_id, bathhouse_id, first_visit_at, last_visit_at, visit_count, total_spent, avg_check, notes, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (owner_id, client_id, bathhouse_id) DO UPDATE SET
			last_visit_at = EXCLUDED.last_visit_at,
			visit_count = guest_cards.visit_count + 1,
			total_spent = guest_cards.total_spent + EXCLUDED.total_spent,
			avg_check = (guest_cards.total_spent + EXCLUDED.total_spent) / (guest_cards.visit_count + 1),
			updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query,
		card.ID, card.OwnerID, card.ClientID, card.BathhouseID,
		card.FirstVisitAt, card.LastVisitAt, card.VisitCount,
		card.TotalSpent, card.AvgCheck, card.Notes, card.Tags,
		card.CreatedAt, card.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert guest card: %w", err)
	}
	return nil
}

func (r *guestCardRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestCard, error) {
	query := `SELECT ` + guestCardColumns + ` FROM guest_cards WHERE id = $1`
	card, err := scanGuestCard(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get guest card by id: %w", err)
	}
	return card, nil
}

func (r *guestCardRepo) GetByOwnerAndClient(ctx context.Context, ownerID, clientID, bathhouseID uuid.UUID) (*domain.GuestCard, error) {
	query := `SELECT ` + guestCardColumns + ` FROM guest_cards WHERE owner_id = $1 AND client_id = $2 AND bathhouse_id = $3`
	card, err := scanGuestCard(r.pool.QueryRow(ctx, query, ownerID, clientID, bathhouseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get guest card: %w", err)
	}
	return card, nil
}

func (r *guestCardRepo) ListByOwner(ctx context.Context, filter domain.GuestCardFilter) (*domain.PaginatedResult[domain.GuestCard], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.NoOwnerFilter {
		// Admin: no owner/bathhouse scoping
	} else if len(filter.BathhouseIDs) > 0 {
		placeholders := make([]string, len(filter.BathhouseIDs))
		for i, id := range filter.BathhouseIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("gc.bathhouse_id IN (%s)", strings.Join(placeholders, ",")))
	} else {
		conditions = append(conditions, fmt.Sprintf("gc.owner_id = $%d", argIdx))
		args = append(args, filter.OwnerID)
		argIdx++
	}

	if filter.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("gc.bathhouse_id = $%d", argIdx))
		args = append(args, *filter.BathhouseID)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(*filter.Search)
		conditions = append(conditions, fmt.Sprintf("(u.name ILIKE $%d ESCAPE '\\' OR u.email ILIKE $%d ESCAPE '\\' OR u.phone ILIKE $%d ESCAPE '\\')", argIdx, argIdx, argIdx))
		args = append(args, "%"+escaped+"%")
		argIdx++
	}

	if filter.Tag != nil && *filter.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(gc.tags)", argIdx))
		args = append(args, *filter.Tag)
		argIdx++
	}

	if filter.Segment != nil {
		cond := segmentCondition(*filter.Segment)
		if cond != "" {
			// Prefix "gc." for segment conditions that use column names
			gcCond := strings.ReplaceAll(cond, "visit_count", "gc.visit_count")
			gcCond = strings.ReplaceAll(gcCond, "last_visit_at", "gc.last_visit_at")
			gcCond = strings.ReplaceAll(gcCond, "total_spent", "gc.total_spent")
			conditions = append(conditions, gcCond)
		}
	}

	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("gc.last_visit_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}

	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("gc.last_visit_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM guest_cards gc LEFT JOIN users u ON gc.client_id = u.id WHERE %s`, where)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count guest cards: %w", err)
	}

	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	orderBy := "gc.last_visit_at DESC"
	switch filter.SortBy {
	case "total_spent":
		orderBy = "gc.total_spent DESC"
	case "visit_count":
		orderBy = "gc.visit_count DESC"
	case "avg_check":
		orderBy = "gc.avg_check DESC"
	case "last_visit":
		orderBy = "gc.last_visit_at DESC"
	}

	offset := (page - 1) * pageSize

	// Prefix each column with gc. for the join query
	prefixedColumns := prefixColumns("gc", guestCardColumns)

	dataQuery := fmt.Sprintf(`SELECT %s FROM guest_cards gc LEFT JOIN users u ON gc.client_id = u.id WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		prefixedColumns, where, orderBy, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list guest cards: %w", err)
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

func (r *guestCardRepo) UpdateNotes(ctx context.Context, id uuid.UUID, notes string, tags []string) error {
	query := `UPDATE guest_cards SET notes = $1, tags = $2, updated_at = NOW() WHERE id = $3`
	ct, err := r.pool.Exec(ctx, query, notes, tags, id)
	if err != nil {
		return fmt.Errorf("update guest card notes: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *guestCardRepo) GetStats(ctx context.Context, filter domain.GuestCardFilter) (*domain.GuestCardStats, error) {
	where, args := buildGuestCardOwnerCondition(filter)
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) AS total_guests,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('month', NOW())) AS new_this_month,
			COALESCE(AVG(visit_count), 0)::BIGINT AS avg_visit_count,
			COALESCE(AVG(total_spent), 0)::BIGINT AS avg_spent
		FROM guest_cards
		WHERE %s`, where)

	var stats domain.GuestCardStats
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalGuests, &stats.NewThisMonth, &stats.AvgVisitCount, &stats.AvgSpent,
	)
	if err != nil {
		return nil, fmt.Errorf("get guest card stats: %w", err)
	}
	return &stats, nil
}

func (r *guestCardRepo) CountBySegment(ctx context.Context, filter domain.GuestCardFilter, segment domain.GuestSegmentSlug) (int64, error) {
	cond := segmentCondition(segment)
	if cond == "" {
		return 0, nil
	}

	where, args := buildGuestCardOwnerCondition(filter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM guest_cards WHERE %s AND %s`, where, cond)
	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count by segment %s: %w", segment, err)
	}
	return count, nil
}

// buildGuestCardOwnerCondition returns the owner/bathhouse WHERE clause and args for guest card queries.
func buildGuestCardOwnerCondition(filter domain.GuestCardFilter) (string, []interface{}) {
	if filter.NoOwnerFilter {
		return "TRUE", nil
	}
	if len(filter.BathhouseIDs) > 0 {
		placeholders := make([]string, len(filter.BathhouseIDs))
		args := make([]interface{}, len(filter.BathhouseIDs))
		for i, id := range filter.BathhouseIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}
		return fmt.Sprintf("bathhouse_id IN (%s)", strings.Join(placeholders, ",")), args
	}
	return "owner_id = $1", []interface{}{filter.OwnerID}
}

// segmentCondition returns SQL WHERE condition for a segment slug (without the owner_id filter).
func segmentCondition(segment domain.GuestSegmentSlug) string {
	switch segment {
	case domain.SegmentNew:
		return "visit_count = 1"
	case domain.SegmentRegular:
		return "visit_count >= 3"
	case domain.SegmentLost:
		return "last_visit_at < NOW() - INTERVAL '90 days'"
	case domain.SegmentVIP:
		return "total_spent > 5000000" // 50,000 RUB in kopecks
	case domain.SegmentBirthdaySoon:
		// Requires birthday field on users table; currently always returns 0
		return "FALSE"
	default:
		return ""
	}
}

// prefixColumns adds a table alias prefix to each column in a comma-separated column list.
func prefixColumns(alias, columns string) string {
	cols := strings.Split(columns, ", ")
	for i, col := range cols {
		cols[i] = alias + "." + strings.TrimSpace(col)
	}
	return strings.Join(cols, ", ")
}
