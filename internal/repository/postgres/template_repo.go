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

type templateRepo struct {
	pool *pgxpool.Pool
}

func NewResponseTemplateRepository(pool *pgxpool.Pool) repository.ResponseTemplateRepository {
	return &templateRepo{pool: pool}
}

var templateColumns = `id, owner_id, title, body, is_default, sort_order, created_at`

func scanTemplate(row pgx.Row) (*domain.ResponseTemplate, error) {
	var t domain.ResponseTemplate
	err := row.Scan(&t.ID, &t.OwnerID, &t.Title, &t.Body, &t.IsDefault, &t.SortOrder, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *templateRepo) Create(ctx context.Context, template *domain.ResponseTemplate) error {
	now := time.Now()
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}

	query := `INSERT INTO response_templates (id, owner_id, title, body, is_default, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		template.ID, template.OwnerID, template.Title, template.Body,
		template.IsDefault, template.SortOrder, template.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create response template: %w", err)
	}
	return nil
}

func (r *templateRepo) Update(ctx context.Context, template *domain.ResponseTemplate) error {
	query := `UPDATE response_templates SET title = $1, body = $2, sort_order = $3 WHERE id = $4`
	ct, err := r.pool.Exec(ctx, query, template.Title, template.Body, template.SortOrder, template.ID)
	if err != nil {
		return fmt.Errorf("update response template: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

func (r *templateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM response_templates WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete response template: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

func (r *templateRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.ResponseTemplate, error) {
	query := fmt.Sprintf(`SELECT %s FROM response_templates WHERE owner_id = $1 ORDER BY sort_order ASC, created_at ASC`, templateColumns)
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list response templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.ResponseTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, fmt.Errorf("scan response template: %w", err)
		}
		templates = append(templates, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate response template rows: %w", err)
	}
	return templates, nil
}

func (r *templateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ResponseTemplate, error) {
	query := fmt.Sprintf(`SELECT %s FROM response_templates WHERE id = $1`, templateColumns)
	t, err := scanTemplate(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get response template: %w", err)
	}
	return t, nil
}

func (r *templateRepo) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM response_templates WHERE owner_id = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count response templates: %w", err)
	}
	return count, nil
}
