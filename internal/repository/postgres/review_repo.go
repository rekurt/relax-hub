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

type reviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) repository.ReviewRepository {
	return &reviewRepo{pool: pool}
}

func (r *reviewRepo) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, bathhouse_id, booking_id, rating, text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	review.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		review.ID, review.UserID, review.BathhouseID, review.BookingID,
		review.Rating, review.Text, review.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *reviewRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE bathhouse_id = $1`, bathhouseID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count reviews: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, bathhouse_id, booking_id, rating, text, created_at
		FROM reviews WHERE bathhouse_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		if err := rows.Scan(
			&rev.ID, &rev.UserID, &rev.BathhouseID, &rev.BookingID,
			&rev.Rating, &rev.Text, &rev.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Review]{
		Items:      reviews,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *reviewRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	query := `
		SELECT id, user_id, bathhouse_id, booking_id, rating, text, created_at
		FROM reviews WHERE booking_id = $1`

	var rev domain.Review
	err := r.pool.QueryRow(ctx, query, bookingID).Scan(
		&rev.ID, &rev.UserID, &rev.BathhouseID, &rev.BookingID,
		&rev.Rating, &rev.Text, &rev.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get review by booking id: %w", err)
	}
	return &rev, nil
}
