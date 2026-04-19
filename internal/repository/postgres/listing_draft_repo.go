package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type listingDraftRepo struct {
	pool *pgxpool.Pool
}

func NewListingDraftRepository(pool *pgxpool.Pool) repository.ListingDraftRepository {
	return &listingDraftRepo{pool: pool}
}

func (r *listingDraftRepo) Create(ctx context.Context, draft *domain.ListingDraft) error {
	stepDataJSON, err := json.Marshal(draft.StepData)
	if err != nil {
		return fmt.Errorf("marshal step_data: %w", err)
	}

	query := `INSERT INTO listing_drafts (id, user_id, status, current_step, step_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = r.pool.Exec(ctx, query,
		draft.ID, draft.UserID, draft.Status, draft.CurrentStep,
		stepDataJSON, draft.CreatedAt, draft.UpdatedAt,
	)
	return err
}

func (r *listingDraftRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ListingDraft, error) {
	query := `SELECT id, user_id, status, current_step, step_data, created_at, updated_at FROM listing_drafts WHERE id = $1`
	return r.scanDraft(r.pool.QueryRow(ctx, query, id))
}

func (r *listingDraftRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.ListingDraft, error) {
	query := `SELECT id, user_id, status, current_step, step_data, created_at, updated_at FROM listing_drafts WHERE user_id = $1 ORDER BY updated_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []domain.ListingDraft
	for rows.Next() {
		d, err := r.scanDraft(rows)
		if err != nil {
			return nil, fmt.Errorf("scan draft: %w", err)
		}
		drafts = append(drafts, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return drafts, nil
}

func (r *listingDraftRepo) UpdateStep(ctx context.Context, id uuid.UUID, step int, data json.RawMessage, currentStep int) error {
	query := `UPDATE listing_drafts
		SET step_data = jsonb_set(step_data, $2::text[], $3::jsonb),
		    current_step = $4,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'draft'`
	stepKey := fmt.Sprintf("{%d}", step)
	result, err := r.pool.Exec(ctx, query, id, stepKey, data, currentStep)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrListingDraftNotFound
	}
	return nil
}

func (r *listingDraftRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ListingDraftStatus) error {
	query := `UPDATE listing_drafts SET status = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrListingDraftNotFound
	}
	return nil
}

func (r *listingDraftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM listing_drafts WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrListingDraftNotFound
	}
	return nil
}

func (r *listingDraftRepo) scanDraft(row pgx.Row) (*domain.ListingDraft, error) {
	var d domain.ListingDraft
	var stepDataJSON []byte
	err := row.Scan(&d.ID, &d.UserID, &d.Status, &d.CurrentStep, &stepDataJSON, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrListingDraftNotFound
		}
		return nil, err
	}

	d.StepData = make(map[int]json.RawMessage)
	if len(stepDataJSON) > 0 {
		// JSONB stores keys as strings, so we need to parse as map[string]json.RawMessage
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(stepDataJSON, &raw); err != nil {
			return nil, fmt.Errorf("unmarshal step_data: %w", err)
		}
		for k, v := range raw {
			var step int
			if _, err := fmt.Sscanf(k, "%d", &step); err == nil {
				d.StepData[step] = v
			}
		}
	}

	return &d, nil
}
