package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
}

type CityRepository interface {
	Create(ctx context.Context, city *domain.City) error
	GetAll(ctx context.Context) ([]domain.City, error)
	GetBySlug(ctx context.Context, slug string) (*domain.City, error)
	GetByID(ctx context.Context, id int64) (*domain.City, error)
	Update(ctx context.Context, city *domain.City) error
	Delete(ctx context.Context, id int64) error
}

type BathhouseRepository interface {
	Create(ctx context.Context, bh *domain.Bathhouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	Update(ctx context.Context, bh *domain.Bathhouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error
}

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
	CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)
}

type RepresentativeRepository interface {
	Create(ctx context.Context, rep *domain.Representative) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
	ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
