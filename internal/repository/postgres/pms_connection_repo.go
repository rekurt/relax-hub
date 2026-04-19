package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type pmsConnectionRepo struct {
	pool *pgxpool.Pool
}

func NewPMSConnectionRepository(pool *pgxpool.Pool) repository.PMSConnectionRepository {
	return &pmsConnectionRepo{pool: pool}
}

var pmsConnectionColumns = `id, owner_id, bathhouse_id, provider, credentials_encrypted, sync_direction,
	sync_interval_minutes, status, last_sync_at, last_sync_error, external_id, created_at, updated_at`

func scanPMSConnection(row pgx.Row) (*domain.PMSConnection, error) {
	var c domain.PMSConnection
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.BathhouseID, &c.Provider, &c.CredentialsEncrypted,
		&c.SyncDirection, &c.SyncIntervalMinutes, &c.Status,
		&c.LastSyncAt, &c.LastSyncError, &c.ExternalID,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPMSConnectionNotFound
		}
		return nil, fmt.Errorf("scan pms connection: %w", err)
	}
	return &c, nil
}

func (r *pmsConnectionRepo) Create(ctx context.Context, conn *domain.PMSConnection) error {
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	_, err := r.pool.Exec(ctx,
		`INSERT INTO pms_connections (id, owner_id, bathhouse_id, provider, credentials_encrypted,
			sync_direction, sync_interval_minutes, status, last_sync_error, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		conn.ID, conn.OwnerID, conn.BathhouseID, conn.Provider, conn.CredentialsEncrypted,
		conn.SyncDirection, conn.SyncIntervalMinutes, conn.Status,
		conn.LastSyncError, conn.ExternalID, conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create pms connection: %w", err)
	}
	return nil
}

func (r *pmsConnectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PMSConnection, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM pms_connections WHERE id = $1", pmsConnectionColumns), id)
	return scanPMSConnection(row)
}

func (r *pmsConnectionRepo) Update(ctx context.Context, conn *domain.PMSConnection) error {
	conn.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE pms_connections SET
			provider = $2, credentials_encrypted = $3, sync_direction = $4,
			sync_interval_minutes = $5, status = $6, external_id = $7, updated_at = $8
		WHERE id = $1`,
		conn.ID, conn.Provider, conn.CredentialsEncrypted, conn.SyncDirection,
		conn.SyncIntervalMinutes, conn.Status, conn.ExternalID, conn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update pms connection: %w", err)
	}
	return nil
}

func (r *pmsConnectionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM pms_connections WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete pms connection: %w", err)
	}
	return nil
}

func (r *pmsConnectionRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSConnection], error) {
	offset := (page - 1) * pageSize

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM pms_connections WHERE owner_id = $1", ownerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count pms connections: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		fmt.Sprintf("SELECT %s FROM pms_connections WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", pmsConnectionColumns),
		ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list pms connections: %w", err)
	}
	defer rows.Close()

	var items []domain.PMSConnection
	for rows.Next() {
		var c domain.PMSConnection
		if err := rows.Scan(
			&c.ID, &c.OwnerID, &c.BathhouseID, &c.Provider, &c.CredentialsEncrypted,
			&c.SyncDirection, &c.SyncIntervalMinutes, &c.Status,
			&c.LastSyncAt, &c.LastSyncError, &c.ExternalID,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pms connection row: %w", err)
		}
		items = append(items, c)
	}

	return &domain.PaginatedResult[domain.PMSConnection]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *pmsConnectionRepo) GetByBathhouseID(ctx context.Context, bathhouseID uuid.UUID) (*domain.PMSConnection, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM pms_connections WHERE bathhouse_id = $1", pmsConnectionColumns), bathhouseID)
	return scanPMSConnection(row)
}

func (r *pmsConnectionRepo) ListActive(ctx context.Context) ([]domain.PMSConnection, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf("SELECT %s FROM pms_connections WHERE status = $1", pmsConnectionColumns),
		domain.PMSConnectionActive)
	if err != nil {
		return nil, fmt.Errorf("list active pms connections: %w", err)
	}
	defer rows.Close()

	var items []domain.PMSConnection
	for rows.Next() {
		var c domain.PMSConnection
		if err := rows.Scan(
			&c.ID, &c.OwnerID, &c.BathhouseID, &c.Provider, &c.CredentialsEncrypted,
			&c.SyncDirection, &c.SyncIntervalMinutes, &c.Status,
			&c.LastSyncAt, &c.LastSyncError, &c.ExternalID,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active pms connection: %w", err)
		}
		items = append(items, c)
	}

	return items, nil
}

func (r *pmsConnectionRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, lastSyncAt time.Time, lastSyncError string, status domain.PMSConnectionStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE pms_connections SET last_sync_at = $2, last_sync_error = $3, status = $4, updated_at = $5 WHERE id = $1`,
		id, lastSyncAt, lastSyncError, status, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update pms connection sync status: %w", err)
	}
	return nil
}

// PMSSyncLogRepo

type pmsSyncLogRepo struct {
	pool *pgxpool.Pool
}

func NewPMSSyncLogRepository(pool *pgxpool.Pool) repository.PMSSyncLogRepository {
	return &pmsSyncLogRepo{pool: pool}
}

func (r *pmsSyncLogRepo) Create(ctx context.Context, log *domain.PMSSyncLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO pms_sync_logs (id, connection_id, direction, status, items_synced, error_message, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.ConnectionID, log.Direction, log.Status,
		log.ItemsSynced, log.ErrorMessage, log.StartedAt, log.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("create pms sync log: %w", err)
	}
	return nil
}

func (r *pmsSyncLogRepo) ListByConnection(ctx context.Context, connectionID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSSyncLog], error) {
	offset := (page - 1) * pageSize

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM pms_sync_logs WHERE connection_id = $1", connectionID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count pms sync logs: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, connection_id, direction, status, items_synced, error_message, started_at, completed_at
		FROM pms_sync_logs WHERE connection_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3`,
		connectionID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list pms sync logs: %w", err)
	}
	defer rows.Close()

	var items []domain.PMSSyncLog
	for rows.Next() {
		var l domain.PMSSyncLog
		if err := rows.Scan(&l.ID, &l.ConnectionID, &l.Direction, &l.Status,
			&l.ItemsSynced, &l.ErrorMessage, &l.StartedAt, &l.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan pms sync log: %w", err)
		}
		items = append(items, l)
	}

	return &domain.PaginatedResult[domain.PMSSyncLog]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}, nil
}
