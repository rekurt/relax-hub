package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type cityRepo struct {
	pool *pgxpool.Pool
}

func NewCityRepository(pool *pgxpool.Pool) repository.CityRepository {
	return &cityRepo{pool: pool}
}

func (r *cityRepo) Create(ctx context.Context, city *domain.City) error {
	query := `
		INSERT INTO cities (name, slug, region, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		city.Name, city.Slug, city.Region, city.Latitude, city.Longitude,
	).Scan(&city.ID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create city: %w", err)
	}
	return nil
}

func (r *cityRepo) GetAll(ctx context.Context) ([]domain.City, error) {
	query := `SELECT id, name, slug, region, latitude, longitude FROM cities ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all cities: %w", err)
	}
	defer rows.Close()

	var cities []domain.City
	for rows.Next() {
		var c domain.City
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Region, &c.Latitude, &c.Longitude); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		cities = append(cities, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate city rows: %w", err)
	}
	return cities, nil
}

func (r *cityRepo) GetBySlug(ctx context.Context, slug string) (*domain.City, error) {
	query := `SELECT id, name, slug, region, latitude, longitude FROM cities WHERE slug = $1`

	var city domain.City
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&city.ID, &city.Name, &city.Slug, &city.Region, &city.Latitude, &city.Longitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get city by slug: %w", err)
	}
	return &city, nil
}

func (r *cityRepo) GetByID(ctx context.Context, id int64) (*domain.City, error) {
	query := `SELECT id, name, slug, region, latitude, longitude FROM cities WHERE id = $1`

	var city domain.City
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&city.ID, &city.Name, &city.Slug, &city.Region, &city.Latitude, &city.Longitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get city by id: %w", err)
	}
	return &city, nil
}

func (r *cityRepo) Update(ctx context.Context, city *domain.City) error {
	query := `UPDATE cities SET name = $2, slug = $3, region = $4, latitude = $5, longitude = $6 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query,
		city.ID, city.Name, city.Slug, city.Region, city.Latitude, city.Longitude,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("update city: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *cityRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM cities WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete city: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
