package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
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

var reviewColumns = `id, user_id, bathhouse_id, booking_id, rating, cleanliness, accuracy, communication, value_for_money, text, status, rejection_reasons, owner_response, owner_response_at, moderation_score, moderation_flags, images, created_at, updated_at`

func scanReview(row pgx.Row) (*domain.Review, error) {
	var rev domain.Review
	var images []string
	var rejectionReasons []string
	var moderationFlags []string
	err := row.Scan(
		&rev.ID, &rev.UserID, &rev.BathhouseID, &rev.BookingID,
		&rev.Rating, &rev.Cleanliness, &rev.Accuracy, &rev.Communication, &rev.ValueForMoney,
		&rev.Text, &rev.Status, &rejectionReasons, &rev.OwnerResponse,
		&rev.OwnerResponseAt, &rev.ModerationScore, &moderationFlags, &images, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	rev.Images = images
	rev.RejectionReasons = rejectionReasons
	rev.ModerationFlags = moderationFlags
	return &rev, nil
}

func scanReviews(rows pgx.Rows) ([]domain.Review, error) {
	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		var images []string
		var rejectionReasons []string
		var moderationFlags []string
		if err := rows.Scan(
			&rev.ID, &rev.UserID, &rev.BathhouseID, &rev.BookingID,
			&rev.Rating, &rev.Cleanliness, &rev.Accuracy, &rev.Communication, &rev.ValueForMoney,
			&rev.Text, &rev.Status, &rejectionReasons, &rev.OwnerResponse,
			&rev.OwnerResponseAt, &rev.ModerationScore, &moderationFlags, &images, &rev.CreatedAt, &rev.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		rev.Images = images
		rev.RejectionReasons = rejectionReasons
		rev.ModerationFlags = moderationFlags
		reviews = append(reviews, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}
	return reviews, nil
}

func (r *reviewRepo) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, bathhouse_id, booking_id, rating, cleanliness, accuracy, communication, value_for_money, text, status, rejection_reasons, moderation_score, moderation_flags, images, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		review.ID, review.UserID, review.BathhouseID, review.BookingID,
		review.Rating, review.Cleanliness, review.Accuracy, review.Communication, review.ValueForMoney,
		review.Text, review.Status, review.RejectionReasons,
		review.ModerationScore, review.ModerationFlags, review.Images,
		review.CreatedAt, review.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *reviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	query := `SELECT ` + reviewColumns + ` FROM reviews WHERE id = $1`

	rev, err := scanReview(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get review by id: %w", err)
	}
	return rev, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *domain.Review) error {
	query := `
		UPDATE reviews SET rating = $2, cleanliness = $3, accuracy = $4, communication = $5, value_for_money = $6,
		text = $7, images = $8, updated_at = $9, status = $10, moderation_score = $11, moderation_flags = $12
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		review.ID, review.Rating, review.Cleanliness, review.Accuracy, review.Communication, review.ValueForMoney,
		review.Text, review.Images, review.UpdatedAt, review.Status, review.ModerationScore, review.ModerationFlags,
	)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *reviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
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
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE bathhouse_id = $1 AND status = 'approved'`, bathhouseID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count reviews: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + reviewColumns + ` FROM reviews WHERE bathhouse_id = $1 AND status = 'approved' ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	reviews, err := scanReviews(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Review]{
		Items:      reviews,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *reviewRepo) ListByBathhouseFiltered(ctx context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("bathhouse_id = $%d", argIdx))
		args = append(args, *filter.BathhouseID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", argIdx))
		args = append(args, *filter.MinRating)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM reviews` + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count filtered reviews: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM reviews%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		reviewColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list filtered reviews: %w", err)
	}
	defer rows.Close()

	reviews, err := scanReviews(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Review]{
		Items:      reviews,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}

func (r *reviewRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	query := `SELECT ` + reviewColumns + ` FROM reviews WHERE booking_id = $1`

	rev, err := scanReview(r.pool.QueryRow(ctx, query, bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get review by booking id: %w", err)
	}
	return rev, nil
}

func (r *reviewRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error {
	query := `UPDATE reviews SET status = $2, updated_at = $3 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status), time.Now())
	if err != nil {
		return fmt.Errorf("update review status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *reviewRepo) AddOwnerResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error {
	query := `UPDATE reviews SET owner_response = $2, owner_response_at = $3, updated_at = $4 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, response, respondedAt, respondedAt)
	if err != nil {
		return fmt.Errorf("add owner response: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *reviewRepo) GetUserReviewStats(ctx context.Context, userID uuid.UUID) (*domain.UserReviewStats, error) {
	query := `
		SELECT COUNT(*), COALESCE(AVG(rating), 0)
		FROM reviews
		WHERE user_id = $1 AND status = 'approved'`

	var stats domain.UserReviewStats
	err := r.pool.QueryRow(ctx, query, userID).Scan(&stats.ReviewCount, &stats.AvgRating)
	if err != nil {
		return nil, fmt.Errorf("get user review stats: %w", err)
	}
	return &stats, nil
}

func (r *reviewRepo) UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error {
	query := `UPDATE reviews SET status = $2, rejection_reasons = $3, updated_at = $4 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status), reasons, time.Now())
	if err != nil {
		return fmt.Errorf("update review status with reasons: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *reviewRepo) CountPendingReviews(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE status = 'pending'`
	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count pending reviews: %w", err)
	}
	return count, nil
}

func (r *reviewRepo) GetCriteriaAverages(ctx context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error) {
	query := `
		SELECT
			COALESCE(AVG(cleanliness), 0),
			COALESCE(AVG(accuracy), 0),
			COALESCE(AVG(communication), 0),
			COALESCE(AVG(value_for_money), 0)
		FROM reviews
		WHERE bathhouse_id = $1 AND status = 'approved' AND cleanliness IS NOT NULL`

	var avgs domain.ReviewCriteriaAverages
	err := r.pool.QueryRow(ctx, query, bathhouseID).Scan(
		&avgs.AvgCleanliness, &avgs.AvgAccuracy, &avgs.AvgCommunication, &avgs.AvgValueForMoney,
	)
	if err != nil {
		return nil, fmt.Errorf("get criteria averages: %w", err)
	}
	return &avgs, nil
}

func (r *reviewRepo) GetPlatformAverageRating(ctx context.Context) (float64, error) {
	var avg float64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(AVG(rating)::double precision, 0) FROM reviews WHERE status = 'approved'`).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("get platform average rating: %w", err)
	}
	return avg, nil
}

func (r *reviewRepo) ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("bathhouse_id = $%d", argIdx))
		args = append(args, *filter.BathhouseID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", argIdx))
		args = append(args, *filter.MinRating)
		argIdx++
	}
	if filter.MaxRating != nil {
		conditions = append(conditions, fmt.Sprintf("rating <= $%d", argIdx))
		args = append(args, *filter.MaxRating)
		argIdx++
	}
	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.FromDate)
		argIdx++
	}
	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.ToDate)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM reviews` + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count all reviews: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM reviews%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		reviewColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all reviews: %w", err)
	}
	defer rows.Close()

	reviews, err := scanReviews(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Review]{
		Items:      reviews,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}
