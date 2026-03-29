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
		INSERT INTO users (id, email, password_hash, name, phone, phone_verified, role, admin_sub_role, is_active, avatar_url, bio, city_id, region, totp_secret, two_fa_method, onboarding_completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	twoFAMethod := string(user.TwoFAMethod)
	if twoFAMethod == "" {
		twoFAMethod = string(domain.TwoFANone)
	}

	region := string(user.Region)
	if region == "" {
		region = string(domain.RegionRU)
		user.Region = domain.RegionRU
	}

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Phone, user.PhoneVerified,
		user.Role, string(user.AdminSubRole), user.IsActive, user.AvatarURL, user.Bio, user.CityID,
		region, user.TOTPSecret, twoFAMethod, user.OnboardingCompleted,
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
		SELECT id, email, password_hash, name, phone, phone_verified, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), COALESCE(referral_code, ''), COALESCE(totp_secret, ''), COALESCE(two_fa_method, 'none'), COALESCE(onboarding_completed, false), deletion_requested_at, deletion_scheduled_at, created_at, updated_at
		FROM users WHERE id = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone, &user.PhoneVerified,
		&user.Role, &user.AdminSubRole, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.Region, &user.ReferralCode, &user.TOTPSecret, &user.TwoFAMethod, &user.OnboardingCompleted, &user.DeletionRequestedAt, &user.DeletionScheduledAt, &user.CreatedAt, &user.UpdatedAt,
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
		SELECT id, email, password_hash, name, phone, phone_verified, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), COALESCE(referral_code, ''), COALESCE(totp_secret, ''), COALESCE(two_fa_method, 'none'), COALESCE(onboarding_completed, false), deletion_requested_at, deletion_scheduled_at, created_at, updated_at
		FROM users WHERE email = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone, &user.PhoneVerified,
		&user.Role, &user.AdminSubRole, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.Region, &user.ReferralCode, &user.TOTPSecret, &user.TwoFAMethod, &user.OnboardingCompleted, &user.DeletionRequestedAt, &user.DeletionScheduledAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, phone_verified, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), COALESCE(referral_code, ''), COALESCE(totp_secret, ''), COALESCE(two_fa_method, 'none'), COALESCE(onboarding_completed, false), deletion_requested_at, deletion_scheduled_at, created_at, updated_at
		FROM users WHERE phone = $1 AND phone != ''`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone, &user.PhoneVerified,
		&user.Role, &user.AdminSubRole, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.Region, &user.ReferralCode, &user.TOTPSecret, &user.TwoFAMethod, &user.OnboardingCompleted, &user.DeletionRequestedAt, &user.DeletionScheduledAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET email = $2, name = $3, phone = $4, phone_verified = $5, role = $6, admin_sub_role = $7, avatar_url = $8, bio = $9, city_id = $10, region = $11, totp_secret = $12, two_fa_method = $13, onboarding_completed = $14, updated_at = $15
		WHERE id = $1`

	user.UpdatedAt = time.Now()
	region := string(user.Region)
	if region == "" {
		region = string(domain.RegionRU)
	}
	tag, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.Phone, user.PhoneVerified, user.Role, string(user.AdminSubRole),
		user.AvatarURL, user.Bio, user.CityID, region, user.TOTPSecret, user.TwoFAMethod, user.OnboardingCompleted, user.UpdatedAt,
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
		SELECT id, email, name, phone, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), created_at, updated_at
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
			&u.Role, &u.AdminSubRole, &u.IsActive, &u.AvatarURL, &u.Bio, &u.CityID,
			&u.Region, &u.CreatedAt, &u.UpdatedAt,
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
			COALESCE(rev_stats.review_count, 0) AS review_count,
			COALESCE(booking_stats.visit_count, 0) AS visit_count,
			COALESCE(rev_stats.avg_rating, 0) AS avg_rating
		FROM users u
		LEFT JOIN cities c ON u.city_id = c.id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS review_count, AVG(rating) AS avg_rating
			FROM reviews WHERE status = 'approved' AND user_id = $1 GROUP BY user_id
		) rev_stats ON rev_stats.user_id = u.id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS visit_count
			FROM bookings WHERE status = 'completed' AND user_id = $1 GROUP BY user_id
		) booking_stats ON booking_stats.user_id = u.id
		WHERE u.id = $1 AND u.is_active = true`

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

func (r *userRepo) GetByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, phone_verified, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), COALESCE(referral_code, ''), COALESCE(totp_secret, ''), COALESCE(two_fa_method, 'none'), COALESCE(onboarding_completed, false), deletion_requested_at, deletion_scheduled_at, created_at, updated_at
		FROM users WHERE referral_code = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone, &user.PhoneVerified,
		&user.Role, &user.AdminSubRole, &user.IsActive, &user.AvatarURL, &user.Bio, &user.CityID,
		&user.Region, &user.ReferralCode, &user.TOTPSecret, &user.TwoFAMethod, &user.OnboardingCompleted, &user.DeletionRequestedAt, &user.DeletionScheduledAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by referral code: %w", err)
	}
	return &user, nil
}

func (r *userRepo) UpdateReferralCode(ctx context.Context, userID uuid.UUID, code string) error {
	query := `UPDATE users SET referral_code = $2, updated_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, userID, code, time.Now())
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("update referral code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) SetDeletionSchedule(ctx context.Context, userID uuid.UUID, requestedAt, scheduledAt *time.Time) error {
	query := `UPDATE users SET deletion_requested_at = $2, deletion_scheduled_at = $3, updated_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, userID, requestedAt, scheduledAt, time.Now())
	if err != nil {
		return fmt.Errorf("set deletion schedule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) ListPendingDeletions(ctx context.Context, before time.Time) ([]domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, phone_verified, role, COALESCE(admin_sub_role, ''), is_active, avatar_url, bio, city_id, COALESCE(region, 'RU'), COALESCE(referral_code, ''), COALESCE(totp_secret, ''), COALESCE(two_fa_method, 'none'), COALESCE(onboarding_completed, false), deletion_requested_at, deletion_scheduled_at, created_at, updated_at
		FROM users WHERE deletion_scheduled_at IS NOT NULL AND deletion_scheduled_at <= $1`

	rows, err := r.pool.Query(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("list pending deletions: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.PhoneVerified,
			&u.Role, &u.AdminSubRole, &u.IsActive, &u.AvatarURL, &u.Bio, &u.CityID,
			&u.Region, &u.ReferralCode, &u.TOTPSecret, &u.TwoFAMethod, &u.OnboardingCompleted, &u.DeletionRequestedAt, &u.DeletionScheduledAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pending deletion user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepo) AnonymizeUser(ctx context.Context, userID uuid.UUID, anonEmail string) error {
	query := `
		UPDATE users SET
			name = 'Deleted User',
			email = $2,
			phone = '',
			phone_verified = false,
			password_hash = '',
			avatar_url = '',
			bio = '',
			totp_secret = '',
			two_fa_method = 'none',
			referral_code = NULL,
			is_active = false,
			deletion_requested_at = NULL,
			deletion_scheduled_at = NULL,
			updated_at = $3
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, userID, anonEmail, time.Now())
	if err != nil {
		return fmt.Errorf("anonymize user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) CountByCreatedAtRange(ctx context.Context, from, to time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at <= $2`
	var count int64
	err := r.pool.QueryRow(ctx, query, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users by created_at range: %w", err)
	}
	return count, nil
}
