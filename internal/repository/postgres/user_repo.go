package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, role, is_active, avatar_url, bio, city_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Phone,
		user.Role, user.IsActive, user.AvatarURL, user.Bio, user.CityID,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, role, is_active, avatar_url, bio, city_id, created_at, updated_at
		FROM users WHERE id = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone,
		&user.Role, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, role, is_active, avatar_url, bio, city_id, created_at, updated_at
		FROM users WHERE email = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone,
		&user.Role, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET email = $2, name = $3, phone = $4, role = $5, avatar_url = $6, bio = $7, city_id = $8, updated_at = $9
		WHERE id = $1`

	user.UpdatedAt = time.Now()
	tag, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.Phone, user.Role,
		user.AvatarURL, user.Bio, user.CityID, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	query := `
		SELECT id, email, name, phone, role, is_active, avatar_url, bio, city_id, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Name, &u.Phone,
			&u.Role, &u.IsActive, &u.AvatarURL, &u.Bio, &u.CityID,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return &domain.PaginatedResult[domain.User]{
		Items:      users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *userRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	query := `UPDATE users SET is_active = $2, updated_at = $3 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, active, time.Now())
	if err != nil {
		return fmt.Errorf("set user active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	query := `
		SELECT
			u.id, u.name, u.avatar_url, u.bio,
			COALESCE(c.name, '') AS city_name,
			u.created_at,
			COUNT(DISTINCT rev.id) AS review_count,
			COUNT(DISTINCT CASE WHEN b.status = 'completed' THEN b.id END) AS visit_count,
			COALESCE(AVG(rev.rating), 0) AS avg_rating
		FROM users u
		LEFT JOIN cities c ON u.city_id = c.id
		LEFT JOIN reviews rev ON rev.user_id = u.id AND rev.status = 'approved'
		LEFT JOIN bookings b ON b.user_id = u.id
		WHERE u.id = $1 AND u.is_active = true
		GROUP BY u.id, u.name, u.avatar_url, u.bio, c.name, u.created_at`

	var p domain.UserProfile
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.AvatarURL, &p.Bio,
		&p.CityName, &p.MemberSince,
		&p.ReviewCount, &p.VisitCount, &p.AvgRating,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get public profile: %w", err)
	}
	return &p, nil
}
