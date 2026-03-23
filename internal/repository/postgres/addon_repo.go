package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type addonRepo struct {
	pool *pgxpool.Pool
}

func NewAddOnRepository(pool *pgxpool.Pool) repository.AddOnRepository {
	return &addonRepo{pool: pool}
}

var addonColumns = `id, bathhouse_id, name, description, price, unit, is_active, sort_order, created_at, updated_at`

func scanAddOn(row pgx.Row) (*domain.AddOn, error) {
	var a domain.AddOn
	err := row.Scan(
		&a.ID, &a.BathhouseID, &a.Name, &a.Description,
		&a.Price, &a.Unit, &a.IsActive, &a.SortOrder,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func scanAddOns(rows pgx.Rows) ([]domain.AddOn, error) {
	var addons []domain.AddOn
	for rows.Next() {
		a, err := scanAddOn(rows)
		if err != nil {
			return nil, fmt.Errorf("scan addon: %w", err)
		}
		addons = append(addons, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate addon rows: %w", err)
	}
	return addons, nil
}

func (r *addonRepo) Create(ctx context.Context, addon *domain.AddOn) error {
	if addon.ID == uuid.Nil {
		addon.ID = uuid.New()
	}

	query := `
		INSERT INTO addons (id, bathhouse_id, name, description, price, unit, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		addon.ID, addon.BathhouseID, addon.Name, addon.Description,
		addon.Price, string(addon.Unit), addon.IsActive, addon.SortOrder,
		addon.CreatedAt, addon.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create addon: %w", err)
	}
	return nil
}

func (r *addonRepo) Update(ctx context.Context, addon *domain.AddOn) error {
	query := `
		UPDATE addons SET name=$1, description=$2, price=$3, unit=$4, is_active=$5, sort_order=$6, updated_at=$7
		WHERE id=$8`

	ct, err := r.pool.Exec(ctx, query,
		addon.Name, addon.Description, addon.Price, string(addon.Unit),
		addon.IsActive, addon.SortOrder, addon.UpdatedAt, addon.ID,
	)
	if err != nil {
		return fmt.Errorf("update addon: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrAddOnNotFound
	}
	return nil
}

func (r *addonRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM addons WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete addon: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrAddOnNotFound
	}
	return nil
}

func (r *addonRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.AddOn, error) {
	query := fmt.Sprintf(`SELECT %s FROM addons WHERE id=$1`, addonColumns)
	a, err := scanAddOn(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAddOnNotFound
		}
		return nil, fmt.Errorf("get addon by id: %w", err)
	}
	return a, nil
}

func (r *addonRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	query := fmt.Sprintf(`SELECT %s FROM addons WHERE bathhouse_id=$1 ORDER BY sort_order, created_at`, addonColumns)
	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list addons by bathhouse: %w", err)
	}
	defer rows.Close()
	return scanAddOns(rows)
}

func (r *addonRepo) CountByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM addons WHERE bathhouse_id=$1`, bathhouseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count addons: %w", err)
	}
	return count, nil
}

var bookingAddOnColumns = `id, booking_id, addon_id, name, quantity, unit_price, total_price, created_at`

func scanBookingAddOn(row pgx.Row) (*domain.BookingAddOn, error) {
	var ba domain.BookingAddOn
	err := row.Scan(
		&ba.ID, &ba.BookingID, &ba.AddOnID, &ba.Name,
		&ba.Quantity, &ba.UnitPrice, &ba.TotalPrice, &ba.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ba, nil
}

func (r *addonRepo) CreateBookingAddOn(ctx context.Context, ba *domain.BookingAddOn) error {
	if ba.ID == uuid.Nil {
		ba.ID = uuid.New()
	}

	query := `
		INSERT INTO booking_addons (id, booking_id, addon_id, name, quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		ba.ID, ba.BookingID, ba.AddOnID, ba.Name,
		ba.Quantity, ba.UnitPrice, ba.TotalPrice, ba.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create booking addon: %w", err)
	}
	return nil
}

func (r *addonRepo) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingAddOn, error) {
	query := fmt.Sprintf(`SELECT %s FROM booking_addons WHERE booking_id=$1 ORDER BY created_at`, bookingAddOnColumns)
	rows, err := r.pool.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("list booking addons: %w", err)
	}
	defer rows.Close()

	var items []domain.BookingAddOn
	for rows.Next() {
		ba, err := scanBookingAddOn(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking addon: %w", err)
		}
		items = append(items, *ba)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking addon rows: %w", err)
	}
	return items, nil
}
