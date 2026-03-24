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

type bookingRepo struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) repository.BookingRepository {
	return &bookingRepo{pool: pool}
}

func (r *bookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	query := `
		INSERT INTO bookings (id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, points_spent, referral_bonus_used, status, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		booking.ID, booking.UserID, booking.BathhouseID,
		booking.StartTime, booking.EndTime, booking.GuestCount,
		booking.TotalPrice, booking.AddOnTotal, booking.BasePrice, booking.LongSessionDiscount, booking.ExtraGuestSurcharge, booking.LastMinuteDiscount, booking.ServiceFeeAmount, booking.PointsSpent, booking.ReferralBonusUsed, booking.Status, booking.Comment,
		booking.CreatedAt, booking.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (r *bookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `
		SELECT id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, points_spent, referral_bonus_used, status, comment, created_at, updated_at
		FROM bookings WHERE id = $1`

	var b domain.Booking
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.UserID, &b.BathhouseID,
		&b.StartTime, &b.EndTime, &b.GuestCount,
		&b.TotalPrice, &b.AddOnTotal, &b.BasePrice, &b.LongSessionDiscount, &b.ExtraGuestSurcharge, &b.LastMinuteDiscount, &b.ServiceFeeAmount, &b.PointsSpent, &b.ReferralBonusUsed, &b.Status, &b.Comment,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return &b, nil
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
	query := `
		SELECT id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, points_spent, referral_bonus_used, status, comment, created_at, updated_at
		FROM bookings WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bookings by user: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BathhouseID,
			&b.StartTime, &b.EndTime, &b.GuestCount,
			&b.TotalPrice, &b.AddOnTotal, &b.BasePrice, &b.LongSessionDiscount, &b.ExtraGuestSurcharge, &b.LastMinuteDiscount, &b.ServiceFeeAmount, &b.PointsSpent, &b.ReferralBonusUsed, &b.Status, &b.Comment,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, b)
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
	query := `
		SELECT id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, points_spent, referral_bonus_used, status, comment, created_at, updated_at
		FROM bookings WHERE bathhouse_id = $1 ORDER BY start_time DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list bookings by bathhouse: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BathhouseID,
			&b.StartTime, &b.EndTime, &b.GuestCount,
			&b.TotalPrice, &b.AddOnTotal, &b.BasePrice, &b.LongSessionDiscount, &b.ExtraGuestSurcharge, &b.LastMinuteDiscount, &b.ServiceFeeAmount, &b.PointsSpent, &b.ReferralBonusUsed, &b.Status, &b.Comment,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, b)
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
			AND status IN ('pending', 'confirmed')
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
	query := `
		SELECT id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, points_spent, referral_bonus_used, status, comment, created_at, updated_at
		FROM bookings
		WHERE bathhouse_id = $1
			AND status IN ('pending', 'confirmed')
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
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BathhouseID,
			&b.StartTime, &b.EndTime, &b.GuestCount,
			&b.TotalPrice, &b.AddOnTotal, &b.BasePrice, &b.LongSessionDiscount, &b.ExtraGuestSurcharge, &b.LastMinuteDiscount, &b.ServiceFeeAmount, &b.PointsSpent, &b.ReferralBonusUsed, &b.Status, &b.Comment,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan overlapping booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overlapping booking rows: %w", err)
	}
	return bookings, nil
}

func (r *bookingRepo) CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE bathhouse_id = $1 AND status IN ('pending', 'confirmed')`

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
