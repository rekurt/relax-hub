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

const bookingColumns = `id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, modification_count, deposit_amount, deposit_status, deposit_external_id, deposit_released_at, checked_in_at, checked_out_at, hold_id, rejection_reason, cancelled_by_owner, points_spent, referral_bonus_used, status, comment, created_at, updated_at`

const bookingColumnsAliased = `b.id, b.user_id, b.bathhouse_id, b.start_time, b.end_time, b.guest_count, b.total_price, b.addon_total, b.base_price, b.long_session_discount, b.extra_guest_surcharge, b.last_minute_discount, b.service_fee_amount, b.modification_count, b.deposit_amount, b.deposit_status, b.deposit_external_id, b.deposit_released_at, b.checked_in_at, b.checked_out_at, b.hold_id, b.rejection_reason, b.cancelled_by_owner, b.points_spent, b.referral_bonus_used, b.status, b.comment, b.created_at, b.updated_at`

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
		&b.ModificationCount,
		&b.DepositAmount, &b.DepositStatus, &b.DepositExternalID, &b.DepositReleasedAt,
		&b.CheckedInAt, &b.CheckedOutAt,
		&b.HoldID, &b.RejectionReason, &b.CancelledByOwner,
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)`

	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		booking.ID, booking.UserID, booking.BathhouseID,
		booking.StartTime, booking.EndTime, booking.GuestCount,
		booking.TotalPrice, booking.AddOnTotal, booking.BasePrice, booking.LongSessionDiscount, booking.ExtraGuestSurcharge, booking.LastMinuteDiscount, booking.ServiceFeeAmount,
		booking.ModificationCount,
		booking.DepositAmount, booking.DepositStatus, booking.DepositExternalID, booking.DepositReleasedAt,
		booking.CheckedInAt, booking.CheckedOutAt,
		booking.HoldID, booking.RejectionReason, booking.CancelledByOwner,
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

func (r *bookingRepo) CheckAvailabilityExcluding(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time, excludeBookingID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE bathhouse_id = $1
			AND id != $4
			AND status IN ('pending', 'pending_owner', 'confirmed')
			AND start_time < $3
			AND end_time > $2`

	var count int64
	err := r.pool.QueryRow(ctx, query, bathhouseID, startTime, endTime, excludeBookingID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check availability excluding: %w", err)
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
	query := `SELECT ` + bookingColumnsAliased + `
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
	// Only consider bookings whose start_time is between noShowCutoff and 24h before it.
	// This prevents marking very old confirmed bookings (e.g., from days ago) as no-show.
	// Those should be handled by the Complete flow or manual intervention.
	upperBound := noShowCutoff.Add(-24 * time.Hour)
	query := `SELECT ` + bookingColumns + ` FROM bookings
		WHERE status = 'confirmed'
			AND start_time < $1
			AND start_time > $2
			AND checked_in_at IS NULL`

	rows, err := r.pool.Query(ctx, query, noShowCutoff, upperBound)
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

func (r *bookingRepo) UpdateCancelledByOwner(ctx context.Context, bookingID uuid.UUID) error {
	query := `UPDATE bookings SET cancelled_by_owner = true, updated_at = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, bookingID, time.Now())
	if err != nil {
		return fmt.Errorf("update cancelled_by_owner: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) CountOwnerCancellations(ctx context.Context, ownerID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM bookings b
		JOIN bathhouses bh ON bh.id = b.bathhouse_id
		WHERE bh.owner_id = $1
			AND b.cancelled_by_owner = true
			AND b.status = 'cancelled'
			AND b.updated_at >= $2`
	var count int
	err := r.pool.QueryRow(ctx, query, ownerID, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count owner cancellations: %w", err)
	}
	return count, nil
}

func (r *bookingRepo) GetResponseStats(ctx context.Context, bathhouseID uuid.UUID, since time.Time) (totalRequests int, respondedInTime int, avgResponseMinutes int, err error) {
	// Total request-based bookings (those that were ever pending_owner)
	// We identify them by: booking_mode='request' on the bathhouse AND booking was created after 'since'
	// Since pending_owner bookings transition to confirmed/rejected/cancelled, we count all bookings
	// that have hold_id (indicating request mode was used) or were explicitly request-based
	query := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE b.status IN ('confirmed', 'rejected')
				AND (b.rejection_reason IS NULL OR b.rejection_reason != 'Время ожидания ответа истекло')) as responded,
			COALESCE(AVG(EXTRACT(EPOCH FROM (b.updated_at - b.created_at)) / 60)
				FILTER (WHERE b.status IN ('confirmed', 'rejected')
				AND (b.rejection_reason IS NULL OR b.rejection_reason != 'Время ожидания ответа истекло')), 0)::INT as avg_response_minutes
		FROM bookings b
		JOIN bathhouses bh ON bh.id = b.bathhouse_id
		WHERE b.bathhouse_id = $1
			AND b.created_at >= $2
			AND bh.booking_mode = 'request'`

	err = r.pool.QueryRow(ctx, query, bathhouseID, since).Scan(&totalRequests, &respondedInTime, &avgResponseMinutes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get response stats: %w", err)
	}
	return totalRequests, respondedInTime, avgResponseMinutes, nil
}

func (r *bookingRepo) ListCompletedForReviewRequests(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error) {
	query := `
		SELECT ` + bookingColumns + `
		FROM bookings
		WHERE status = 'completed'
		  AND checked_out_at IS NOT NULL
		  AND checked_out_at <= $1
		  AND NOT EXISTS (SELECT 1 FROM reviews WHERE reviews.booking_id = bookings.id)
		ORDER BY checked_out_at ASC
		LIMIT 1000`

	rows, err := r.pool.Query(ctx, query, checkedOutBefore)
	if err != nil {
		return nil, fmt.Errorf("list completed for review requests: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking for review request: %w", err)
		}
		bookings = append(bookings, *b)
	}
	return bookings, nil
}

func (r *bookingRepo) UpdateModification(ctx context.Context, bookingID uuid.UUID, startTime, endTime time.Time, guestCount int, totalPrice, addOnTotal, basePrice, longSessionDiscount, extraGuestSurcharge, lastMinuteDiscount, serviceFeeAmount int64, modificationCount int) error {
	query := `UPDATE bookings SET
		start_time = $2, end_time = $3, guest_count = $4,
		total_price = $5, addon_total = $6, base_price = $7,
		long_session_discount = $8, extra_guest_surcharge = $9,
		last_minute_discount = $10, service_fee_amount = $11,
		modification_count = $12, updated_at = $13
	WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		bookingID, startTime, endTime, guestCount,
		totalPrice, addOnTotal, basePrice,
		longSessionDiscount, extraGuestSurcharge,
		lastMinuteDiscount, serviceFeeAmount,
		modificationCount, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update booking modification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) UpdateEndTime(ctx context.Context, bookingID uuid.UUID, oldEndTime, newEndTime time.Time, newTotalPrice int64) error {
	query := `UPDATE bookings SET end_time = $2, total_price = $3, updated_at = $4 WHERE id = $1 AND end_time = $5`
	tag, err := r.pool.Exec(ctx, query, bookingID, newEndTime, newTotalPrice, time.Now(), oldEndTime)
	if err != nil {
		return fmt.Errorf("update end time: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWalletConcurrentUpdate
	}
	return nil
}

func (r *bookingRepo) GetLastBookingDateByUser(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	var lastDate *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT MAX(created_at) FROM bookings WHERE user_id = $1 AND status IN ('completed', 'confirmed')`,
		userID,
	).Scan(&lastDate)
	if err != nil {
		return nil, fmt.Errorf("get last booking date: %w", err)
	}
	return lastDate, nil
}

func (r *bookingRepo) ListConfirmedByRegionAndDateRange(ctx context.Context, region string, dateFrom, dateTo time.Time) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumnsAliased + `
		FROM bookings b
		JOIN bathhouses bh ON b.bathhouse_id = bh.id
		JOIN cities c ON bh.city_id = c.id
		WHERE b.status = 'confirmed'
			AND c.region = $1
			AND b.start_time <= $3
			AND b.end_time >= $2`

	rows, err := r.pool.Query(ctx, query, region, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("list confirmed by region and date range: %w", err)
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
	return bookings, nil
}

func (r *bookingRepo) UpdateDeposit(ctx context.Context, bookingID uuid.UUID, depositAmount int64, depositStatus domain.DepositStatus, depositExternalID string) error {
	query := `UPDATE bookings SET deposit_amount = $2, deposit_status = $3, deposit_external_id = $4, updated_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, bookingID, depositAmount, depositStatus, depositExternalID)
	if err != nil {
		return fmt.Errorf("update deposit: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) UpdateDepositStatus(ctx context.Context, bookingID uuid.UUID, depositStatus domain.DepositStatus, releasedAt *time.Time) error {
	query := `UPDATE bookings SET deposit_status = $2, deposit_released_at = $3, updated_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, bookingID, depositStatus, releasedAt)
	if err != nil {
		return fmt.Errorf("update deposit status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bookingRepo) ListHeldDepositsReadyForRelease(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings
		WHERE deposit_status = 'held'
		AND checked_out_at IS NOT NULL
		AND checked_out_at < $1
		ORDER BY checked_out_at ASC`

	rows, err := r.pool.Query(ctx, query, checkedOutBefore)
	if err != nil {
		return nil, fmt.Errorf("list held deposits ready for release: %w", err)
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
	return bookings, nil
}
