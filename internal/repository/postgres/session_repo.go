package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type sessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) repository.SessionRepository {
	return &sessionRepo{pool: pool}
}

func (r *sessionRepo) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.LastActiveAt.IsZero() {
		session.LastActiveAt = now
	}
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = now.Add(domain.SessionMaxAge)
	}

	query := `
		INSERT INTO sessions (id, user_id, device_info, browser, ip, last_active_at, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.DeviceInfo, session.Browser,
		session.IP, session.LastActiveAt, session.CreatedAt, session.ExpiresAt,
	)
	return err
}

func (r *sessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, device_info, browser, ip, last_active_at, created_at, expires_at
		FROM sessions WHERE id = $1`

	var s domain.Session
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.DeviceInfo, &s.Browser,
		&s.IP, &s.LastActiveAt, &s.CreatedAt, &s.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *sessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	query := `
		SELECT id, user_id, device_info, browser, ip, last_active_at, created_at, expires_at
		FROM sessions WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_active_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.DeviceInfo, &s.Browser,
			&s.IP, &s.LastActiveAt, &s.CreatedAt, &s.ExpiresAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepo) DeleteAllExcept(ctx context.Context, userID uuid.UUID, exceptID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1 AND id != $2`
	_, err := r.pool.Exec(ctx, query, userID, exceptID)
	return err
}

func (r *sessionRepo) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *sessionRepo) UpdateLastActive(ctx context.Context, id uuid.UUID, lastActiveAt time.Time) error {
	expiresAt := lastActiveAt.Add(domain.SessionMaxAge)
	query := `UPDATE sessions SET last_active_at = $2, expires_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, lastActiveAt, expiresAt)
	return err
}

func (r *sessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
