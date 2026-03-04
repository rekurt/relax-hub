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
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
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
	CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserBookingStats, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Update(ctx context.Context, review *domain.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	ListByBathhouseFiltered(ctx context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error
	AddOwnerResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error
	GetUserReviewStats(ctx context.Context, userID uuid.UUID) (*domain.UserReviewStats, error)
}

type FavoriteRepository interface {
	Add(ctx context.Context, favorite *domain.Favorite) error
	Remove(ctx context.Context, userID, bathhouseID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error)
	IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *domain.NotificationPreferences) error
}

type SocialAccountRepository interface {
	Create(ctx context.Context, account *domain.SocialAccount) error
	GetByProviderAndID(ctx context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error)
	Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
}

type RepresentativeRepository interface {
	Create(ctx context.Context, rep *domain.Representative) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Representative, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
	ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
