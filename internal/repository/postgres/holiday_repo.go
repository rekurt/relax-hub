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

type holidayRepo struct {
	pool *pgxpool.Pool
}

func NewHolidayRepository(pool *pgxpool.Pool) repository.HolidayRepository {
	return &holidayRepo{pool: pool}
}

var holidayColumns = `id, name, date, region, is_recurring, created_at, updated_at`

func scanHoliday(row pgx.Row) (*domain.Holiday, error) {
	var h domain.Holiday
	err := row.Scan(&h.ID, &h.Name, &h.Date, &h.Region, &h.IsRecurring, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &h, nil
}

func (r *holidayRepo) Create(ctx context.Context, holiday *domain.Holiday) error {
	if holiday.ID == uuid.Nil {
		holiday.ID = uuid.New()
	}
	now := time.Now()
	holiday.CreatedAt = now
	holiday.UpdatedAt = now

	query := `INSERT INTO holidays (id, name, date, region, is_recurring, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		holiday.ID, holiday.Name, holiday.Date, holiday.Region, holiday.IsRecurring,
		holiday.CreatedAt, holiday.UpdatedAt)
	return err
}

func (r *holidayRepo) Update(ctx context.Context, holiday *domain.Holiday) error {
	holiday.UpdatedAt = time.Now()
	query := `UPDATE holidays SET name = $1, date = $2, region = $3, is_recurring = $4, updated_at = $5 WHERE id = $6`
	tag, err := r.pool.Exec(ctx, query,
		holiday.Name, holiday.Date, holiday.Region, holiday.IsRecurring, holiday.UpdatedAt, holiday.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *holidayRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM holidays WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *holidayRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error) {
	query := `SELECT ` + holidayColumns + ` FROM holidays WHERE id = $1`
	return scanHoliday(r.pool.QueryRow(ctx, query, id))
}

func (r *holidayRepo) ListAll(ctx context.Context) ([]domain.Holiday, error) {
	query := `SELECT ` + holidayColumns + ` FROM holidays ORDER BY date`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holidays []domain.Holiday
	for rows.Next() {
		var h domain.Holiday
		if err := rows.Scan(&h.ID, &h.Name, &h.Date, &h.Region, &h.IsRecurring, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, rows.Err()
}

func (r *holidayRepo) ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error) {
	query := `SELECT ` + holidayColumns + ` FROM holidays WHERE region = $1 ORDER BY date`
	rows, err := r.pool.Query(ctx, query, region)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holidays []domain.Holiday
	for rows.Next() {
		var h domain.Holiday
		if err := rows.Scan(&h.ID, &h.Name, &h.Date, &h.Region, &h.IsRecurring, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, rows.Err()
}

func (r *holidayRepo) IsHoliday(ctx context.Context, date time.Time, region string) (*domain.Holiday, error) {
	query := `SELECT ` + holidayColumns + ` FROM holidays
		WHERE region = $1
		AND (
			date = $2::date
			OR (is_recurring = true AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM $2::date) AND EXTRACT(DAY FROM date) = EXTRACT(DAY FROM $2::date))
		)
		LIMIT 1`
	return scanHoliday(r.pool.QueryRow(ctx, query, region, date))
}

func (r *holidayRepo) GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error) {
	var multiplier float64
	query := `SELECT multiplier FROM bathhouse_holiday_prices WHERE bathhouse_id = $1`
	err := r.pool.QueryRow(ctx, query, bathhouseID).Scan(&multiplier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DefaultHolidayMultiplier, nil
		}
		return 0, err
	}
	return multiplier, nil
}

func (r *holidayRepo) SetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID, multiplier float64) error {
	query := `INSERT INTO bathhouse_holiday_prices (bathhouse_id, multiplier) VALUES ($1, $2)
		ON CONFLICT (bathhouse_id) DO UPDATE SET multiplier = $2`
	_, err := r.pool.Exec(ctx, query, bathhouseID, multiplier)
	return err
}
