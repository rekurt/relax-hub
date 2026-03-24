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

const bookingColumns = `id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, checked_in_at, checked_out_at, hold_id, rejection_reason, points_spent, referral_bonus_used, status, comment, created_at, updated_at`

type bookingRepo struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) repository.BookingRepository {
	return &bookingRepo{pool: pool}
}

func scanBooking(row interface{ Scan(dest ...any) error }) (*domain.Booking, error) {
	var b domain.Booking
	err := row.Scan(
		&b.ID, &b.UserID, &b.BathhouseID,
		&b.StartTime, &b.EndTime, &b.GuestCount,
		&b.TotalPrice, &b.AddOnTotal, &b.BasePrice, &b.LongSessionDiscount, &b.ExtraGuestSurcharge, &b.LastMinuteDiscount, &b.ServiceFeeAmount,
		&b.CheckedInAt, &b.CheckedOutAt,
		&b.HoldID, &b.RejectionReason,
		&b.PointsSpent, &b.ReferralBonusUsed, &b.Status, &b.Comment,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	query := `
		INSERT INTO bookings (` + bookingColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		booking.ID, booking.UserID, booking.BathhouseID,
		booking.StartTime, booking.EndTime, booking.GuestCount,
		booking.TotalPrice, booking.AddOnTotal, booking.BasePrice, booking.LongSessionDiscount, booking.ExtraGuestSurcharge, booking.LastMinuteDiscount, booking.ServiceFeeAmount,
		booking.CheckedInAt, booking.CheckedOutAt,
		booking.HoldID, booking.RejectionReason,
		booking.PointsSpent, booking.ReferralBonusUsed, booking.Status, booking.Comment,
		booking.CreatedAt, booking.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (r *bookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE id = $1`

	b, err := scanBooking(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return b, nil
}

func (r *bookingRepo) Update(ctx context.Context, booking *domain.Booking) error {
	query := `
		UPDATE bookings SET
			total_price = $2, status = $3, hold_id = $4, rejection_reason = $5, updated_at = $6
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query,
		booking.ID, booking.TotalPrice, booking.Status, booking.HoldID, booking.RejectionReason, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bookings by user: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bookings by user: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Booking]{
		Items:      bookings,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *bookingRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings WHERE bathhouse_id = $1`, bathhouseID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bookings by bathhouse: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE bathhouse_id = $1 ORDER BY start_time DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bookings by bathhouse: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Booking]{
		Items:      bookings,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *bookingRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error {
	query := `UPDATE bookings SET status = $2, updated_at = $3 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE bathhouse_id = $1
			AND status IN ('pending', 'pending_owner', 'confirmed')
			AND start_time < $3
			AND end_time > $2`

	var count int64
	err := r.pool.QueryRow(ctx, query, bathhouseID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check availability: %w", err)
	}
	return count == 0, nil
}

func (r *bookingRepo) GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + `
		FROM bookings
		WHERE bathhouse_id = $1
			AND status IN ('pending', 'pending_owner', 'confirmed')
			AND start_time < $3
			AND end_time > $2
		ORDER BY start_time`

	rows, err := r.pool.Query(ctx, query, bathhouseID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get overlapping bookings: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan overlapping booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overlapping booking rows: %w", err)
	}
	return bookings, nil
}

func (r *bookingRepo) CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE bathhouse_id = $1 AND status IN ('pending', 'pending_owner', 'confirmed')`

	var count int64
	err := r.pool.QueryRow(ctx, query, bathhouseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active bookings: %w", err)
	}
	return count, nil
}

func (r *bookingRepo) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserBookingStats, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(total_price), 0), COALESCE(AVG(total_price), 0)
		FROM bookings
		WHERE user_id = $1 AND status = 'completed'`

	var stats domain.UserBookingStats
	var avgCheck float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&stats.TotalVisits, &stats.TotalSpent, &avgCheck)
	if err != nil {
		return nil, fmt.Errorf("get user booking stats: %w", err)
	}
	stats.AvgCheck = int64(math.Round(avgCheck))
	return &stats, nil
}

func (r *bookingRepo) ListTimedOutRequests(ctx context.Context) ([]domain.Booking, error) {
	query := `SELECT b.` + bookingColumns + `
		FROM bookings b
		JOIN bathhouses bh ON b.bathhouse_id = bh.id
		WHERE b.status = 'pending_owner'
			AND b.created_at + (bh.request_timeout || ' hours')::interval < now()`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list timed out requests: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan timed out booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timed out booking rows: %w", err)
	}
	return bookings, nil
}

func (r *bookingRepo) UpdateCheckin(ctx context.Context, bookingID uuid.UUID, checkedInAt *time.Time) error {
	query := `UPDATE bookings SET checked_in_at = $2, updated_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, bookingID, checkedInAt, time.Now())
	if err != nil {
		return fmt.Errorf("update checkin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) UpdateCheckout(ctx context.Context, bookingID uuid.UUID, checkedOutAt *time.Time, status domain.BookingStatus) error {
	query := `UPDATE bookings SET checked_out_at = $2, status = $3, updated_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, bookingID, checkedOutAt, status, time.Now())
	if err != nil {
		return fmt.Errorf("update checkout: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) ListConfirmedWithoutCheckin(ctx context.Context, noShowCutoff time.Time) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings
		WHERE status = 'confirmed'
			AND start_time < $1
			AND checked_in_at IS NULL`

	rows, err := r.pool.Query(ctx, query, noShowCutoff)
	if err != nil {
		return nil, fmt.Errorf("list confirmed without checkin: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan no-show booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate no-show booking rows: %w", err)
	}
	return bookings, nil
}

func (r *bookingRepo) ListUpcoming(ctx context.Context, from, to time.Time) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings
		WHERE status = 'confirmed'
			AND start_time >= $1
			AND start_time <= $2`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("list upcoming bookings: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan upcoming booking: %w", err)
		}
		bookings = append(bookings, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate upcoming booking rows: %w", err)
	}
	return bookings, nil
}
