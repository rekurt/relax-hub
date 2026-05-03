package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type clientReviewRepo struct {
	pool *pgxpool.Pool
}

func NewClientReviewRepository(pool *pgxpool.Pool) repository.ClientReviewRepository {
	return &clientReviewRepo{pool: pool}
}

func scanClientReview(row pgx.Row) (*domain.ClientReview, error) {
	var r domain.ClientReview
	err := row.Scan(
		&r.ID, &r.OwnerID, &r.ClientID, &r.BookingID, &r.BathhouseID,
		&r.Punctuality, &r.Cleanliness, &r.RuleCompliance,
		&r.Rating, &r.Text, &r.RevealAt, &r.IsRevealed,
		&r.ModerationScore, &r.ModerationFlags,
		&r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClientReviewNotFound
		}
		return nil, err
	}
	return &r, nil
}

const clientReviewColumns = `id, owner_id, client_id, booking_id, bathhouse_id,
	punctuality, cleanliness, rule_compliance,
	rating, text, reveal_at, is_revealed,
	moderation_score, moderation_flags,
	status, created_at, updated_at`

func (r *clientReviewRepo) Create(ctx context.Context, review *domain.ClientReview) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO client_reviews (id, owner_id, client_id, booking_id, bathhouse_id,
			punctuality, cleanliness, rule_compliance,
			rating, text, reveal_at, is_revealed,
			moderation_score, moderation_flags,
			status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		review.ID, review.OwnerID, review.ClientID, review.BookingID, review.BathhouseID,
		review.Punctuality, review.Cleanliness, review.RuleCompliance,
		review.Rating, review.Text, review.RevealAt, review.IsRevealed,
		review.ModerationScore, review.ModerationFlags,
		review.Status, review.CreatedAt, review.UpdatedAt,
	)
	return err
}

func (r *clientReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientReview, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+clientReviewColumns+` FROM client_reviews WHERE id = $1`, id)
	return scanClientReview(row)
}

func (r *clientReviewRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.ClientReview, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+clientReviewColumns+` FROM client_reviews WHERE booking_id = $1`, bookingID)
	return scanClientReview(row)
}

func (r *clientReviewRepo) Update(ctx context.Context, review *domain.ClientReview) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE client_reviews SET
			punctuality = $2, cleanliness = $3, rule_compliance = $4,
			rating = $5, text = $6, reveal_at = $7, is_revealed = $8,
			moderation_score = $9, moderation_flags = $10,
			status = $11, updated_at = $12
		WHERE id = $1`,
		review.ID,
		review.Punctuality, review.Cleanliness, review.RuleCompliance,
		review.Rating, review.Text, review.RevealAt, review.IsRevealed,
		review.ModerationScore, review.ModerationFlags,
		review.Status, review.UpdatedAt,
	)
	return err
}

func (r *clientReviewRepo) ListByClient(ctx context.Context, clientID uuid.UUID, onlyRevealed bool, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error) {
	where := "WHERE client_id = $1"
	args := []interface{}{clientID}
	if onlyRevealed {
		where += " AND is_revealed = true"
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM client_reviews "+where, args...).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM client_reviews %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		clientReviewColumns, where, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ClientReview
	for rows.Next() {
		var cr domain.ClientReview
		if err := rows.Scan(
			&cr.ID, &cr.OwnerID, &cr.ClientID, &cr.BookingID, &cr.BathhouseID,
			&cr.Punctuality, &cr.Cleanliness, &cr.RuleCompliance,
			&cr.Rating, &cr.Text, &cr.RevealAt, &cr.IsRevealed,
			&cr.ModerationScore, &cr.ModerationFlags,
			&cr.Status, &cr.CreatedAt, &cr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, cr)
	}

	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.ClientReview]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func (r *clientReviewRepo) ListUnrevealedPastDeadline(ctx context.Context, now time.Time) ([]domain.ClientReview, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+clientReviewColumns+`
		FROM client_reviews
		WHERE is_revealed = false AND reveal_at <= $1
		ORDER BY reveal_at ASC
		LIMIT 500`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ClientReview
	for rows.Next() {
		var cr domain.ClientReview
		if err := rows.Scan(
			&cr.ID, &cr.OwnerID, &cr.ClientID, &cr.BookingID, &cr.BathhouseID,
			&cr.Punctuality, &cr.Cleanliness, &cr.RuleCompliance,
			&cr.Rating, &cr.Text, &cr.RevealAt, &cr.IsRevealed,
			&cr.ModerationScore, &cr.ModerationFlags,
			&cr.Status, &cr.CreatedAt, &cr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, cr)
	}
	return items, nil
}

func (r *clientReviewRepo) RevealByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE client_reviews SET is_revealed = true, updated_at = now() WHERE id = $1`, id)
	return err
}
