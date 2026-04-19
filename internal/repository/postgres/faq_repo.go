package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type faqRepo struct {
	pool *pgxpool.Pool
}

func NewFAQRepository(pool *pgxpool.Pool) repository.FAQRepository {
	return &faqRepo{pool: pool}
}

var faqColumns = `id, category, question, answer, keywords, sort_order, active, created_at, updated_at`

func scanFAQ(row pgx.Row) (*domain.FAQ, error) {
	var f domain.FAQ
	err := row.Scan(
		&f.ID, &f.Category, &f.Question, &f.Answer, &f.Keywords,
		&f.SortOrder, &f.Active, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *faqRepo) Create(ctx context.Context, faq *domain.FAQ) error {
	if faq.ID == uuid.Nil {
		faq.ID = uuid.New()
	}
	query := fmt.Sprintf(`INSERT INTO faq (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, faqColumns)
	_, err := r.pool.Exec(ctx, query,
		faq.ID, string(faq.Category), faq.Question, faq.Answer, faq.Keywords,
		faq.SortOrder, faq.Active, faq.CreatedAt, faq.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create faq: %w", err)
	}
	return nil
}

func (r *faqRepo) Update(ctx context.Context, faq *domain.FAQ) error {
	query := `UPDATE faq SET category=$1, question=$2, answer=$3, keywords=$4, sort_order=$5, active=$6, updated_at=$7 WHERE id=$8`
	ct, err := r.pool.Exec(ctx, query,
		string(faq.Category), faq.Question, faq.Answer, faq.Keywords,
		faq.SortOrder, faq.Active, faq.UpdatedAt, faq.ID,
	)
	if err != nil {
		return fmt.Errorf("update faq: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrFAQNotFound
	}
	return nil
}

func (r *faqRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM faq WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete faq: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrFAQNotFound
	}
	return nil
}

func (r *faqRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.FAQ, error) {
	query := fmt.Sprintf(`SELECT %s FROM faq WHERE id = $1`, faqColumns)
	f, err := scanFAQ(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrFAQNotFound
		}
		return nil, fmt.Errorf("get faq by id: %w", err)
	}
	return f, nil
}

func (r *faqRepo) List(ctx context.Context, filter domain.FAQFilter) (*domain.PaginatedResult[domain.FAQ], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, string(*filter.Category))
		argIdx++
	}
	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIdx))
		args = append(args, *filter.Active)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM faq %s`, where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count faq: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM faq %s ORDER BY sort_order ASC, created_at DESC LIMIT $%d OFFSET $%d`, faqColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list faq: %w", err)
	}
	defer rows.Close()

	var items []domain.FAQ
	for rows.Next() {
		f, err := scanFAQ(rows)
		if err != nil {
			return nil, fmt.Errorf("scan faq: %w", err)
		}
		items = append(items, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate faq rows: %w", err)
	}

	return &domain.PaginatedResult[domain.FAQ]{
		Items:      items,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int((totalCount + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
	}, nil
}

func (r *faqRepo) SearchByKeywords(ctx context.Context, query string, limit int) ([]domain.FAQMatch, error) {
	if limit <= 0 {
		limit = 3
	}
	if query == "" {
		return nil, nil
	}

	// Use trigram similarity on question + keyword array overlap for ranking.
	// The query combines trigram similarity on the question text with
	// keyword array matching for comprehensive relevance scoring.
	sqlQuery := `
		SELECT id, category, question, answer, keywords, sort_order, active, created_at, updated_at,
			GREATEST(
				similarity(lower(question), lower($1)),
				(SELECT COALESCE(MAX(similarity(lower(kw), lower($1))), 0) FROM unnest(keywords) AS kw)
			) AS score
		FROM faq
		WHERE active = true
			AND (
				similarity(lower(question), lower($1)) > 0.1
				OR EXISTS (SELECT 1 FROM unnest(keywords) AS kw WHERE similarity(lower(kw), lower($1)) > 0.2)
			)
		ORDER BY score DESC, sort_order ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search faq by keywords: %w", err)
	}
	defer rows.Close()

	var matches []domain.FAQMatch
	for rows.Next() {
		var f domain.FAQ
		var score float64
		err := rows.Scan(
			&f.ID, &f.Category, &f.Question, &f.Answer, &f.Keywords,
			&f.SortOrder, &f.Active, &f.CreatedAt, &f.UpdatedAt,
			&score,
		)
		if err != nil {
			return nil, fmt.Errorf("scan faq match: %w", err)
		}
		matches = append(matches, domain.FAQMatch{FAQ: f, Score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate faq match rows: %w", err)
	}

	return matches, nil
}

func (r *faqRepo) ListActiveByCategory(ctx context.Context, category domain.FAQCategory) ([]domain.FAQ, error) {
	query := fmt.Sprintf(`SELECT %s FROM faq WHERE active = true AND category = $1 ORDER BY sort_order ASC`, faqColumns)
	rows, err := r.pool.Query(ctx, query, string(category))
	if err != nil {
		return nil, fmt.Errorf("list active faq by category: %w", err)
	}
	defer rows.Close()

	var items []domain.FAQ
	for rows.Next() {
		f, err := scanFAQ(rows)
		if err != nil {
			return nil, fmt.Errorf("scan faq: %w", err)
		}
		items = append(items, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate faq rows: %w", err)
	}

	return items, nil
}
