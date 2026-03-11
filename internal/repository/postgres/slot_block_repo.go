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

type slotBlockRepo struct {
	pool *pgxpool.Pool
}

func NewSlotBlockRepository(pool *pgxpool.Pool) repository.SlotBlockRepository {
	return &slotBlockRepo{pool: pool}
}

func (r *slotBlockRepo) Create(ctx context.Context, block *domain.SlotBlock) error {
	if block.ID == uuid.Nil {
		block.ID = uuid.New()
	}
	block.CreatedAt = time.Now()

	query := `INSERT INTO slot_blocks (id, bathhouse_id, start_time, end_time, source, external_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		block.ID, block.BathhouseID, block.StartTime, block.EndTime,
		block.Source, block.ExternalID, block.Description, block.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create slot block: %w", err)
	}
	return nil
}

func (r *slotBlockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM slot_blocks WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete slot block: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *slotBlockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SlotBlock, error) {
	query := `SELECT id, bathhouse_id, start_time, end_time, source, external_id, description, created_at
		FROM slot_blocks WHERE id = $1`

	var block domain.SlotBlock
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&block.ID, &block.BathhouseID, &block.StartTime, &block.EndTime,
		&block.Source, &block.ExternalID, &block.Description, &block.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get slot block: %w", err)
	}
	return &block, nil
}

func (r *slotBlockRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SlotBlock, error) {
	query := `SELECT id, bathhouse_id, start_time, end_time, source, external_id, description, created_at
		FROM slot_blocks WHERE bathhouse_id = $1 ORDER BY start_time`

	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list slot blocks: %w", err)
	}
	defer rows.Close()

	var blocks []domain.SlotBlock
	for rows.Next() {
		var b domain.SlotBlock
		if err := rows.Scan(&b.ID, &b.BathhouseID, &b.StartTime, &b.EndTime, &b.Source, &b.ExternalID, &b.Description, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan slot block: %w", err)
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *slotBlockRepo) GetByExternalID(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource, externalID string) (*domain.SlotBlock, error) {
	query := `SELECT id, bathhouse_id, start_time, end_time, source, external_id, description, created_at
		FROM slot_blocks WHERE bathhouse_id = $1 AND source = $2 AND external_id = $3`

	var block domain.SlotBlock
	err := r.pool.QueryRow(ctx, query, bathhouseID, source, externalID).Scan(
		&block.ID, &block.BathhouseID, &block.StartTime, &block.EndTime,
		&block.Source, &block.ExternalID, &block.Description, &block.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get slot block by external id: %w", err)
	}
	return &block, nil
}

func (r *slotBlockRepo) DeleteBySource(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource) error {
	query := `DELETE FROM slot_blocks WHERE bathhouse_id = $1 AND source = $2`
	_, err := r.pool.Exec(ctx, query, bathhouseID, source)
	if err != nil {
		return fmt.Errorf("delete slot blocks by source: %w", err)
	}
	return nil
}

func (r *slotBlockRepo) GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.SlotBlock, error) {
	query := `SELECT id, bathhouse_id, start_time, end_time, source, external_id, description, created_at
		FROM slot_blocks
		WHERE bathhouse_id = $1 AND start_time < $3 AND end_time > $2
		ORDER BY start_time`

	rows, err := r.pool.Query(ctx, query, bathhouseID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get overlapping slot blocks: %w", err)
	}
	defer rows.Close()

	var blocks []domain.SlotBlock
	for rows.Next() {
		var b domain.SlotBlock
		if err := rows.Scan(&b.ID, &b.BathhouseID, &b.StartTime, &b.EndTime, &b.Source, &b.ExternalID, &b.Description, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan slot block: %w", err)
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *slotBlockRepo) HasOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM slot_blocks
		WHERE bathhouse_id = $1 AND start_time < $3 AND end_time > $2`

	var count int64
	err := r.pool.QueryRow(ctx, query, bathhouseID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check overlapping slot blocks: %w", err)
	}
	return count > 0, nil
}
