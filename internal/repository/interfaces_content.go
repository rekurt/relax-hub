package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Update(ctx context.Context, review *domain.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	ListByBathhouseFiltered(ctx context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error
	UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error
	AddOwnerResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error
	GetUserReviewStats(ctx context.Context, userID uuid.UUID) (*domain.UserReviewStats, error)
	CountPendingReviews(ctx context.Context) (int64, error)
	ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	GetCriteriaAverages(ctx context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error)
	GetPlatformAverageRating(ctx context.Context) (float64, error)
	ListUnrevealedPastDeadline(ctx context.Context, now time.Time) ([]domain.Review, error)
}


type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	ListByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error)
	ListByOwnerIDs(ctx context.Context, ownerType domain.MediaOwnerType, ownerIDs []uuid.UUID) (map[uuid.UUID][]domain.Media, error)
	ListByBathhouseReviews(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MediaStatus) error
	CountByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType *domain.MediaType) (int64, error)
}


type ComplaintRepository interface {
	Create(ctx context.Context, complaint *domain.Complaint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error)
	List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, resolvedByID *uuid.UUID, resolution string) error
	CountByTarget(ctx context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) (int64, error)
	CheckExists(ctx context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID) (bool, error)
}


type FavoriteRepository interface {
	Add(ctx context.Context, favorite *domain.Favorite) error
	Remove(ctx context.Context, userID, bathhouseID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error)
	IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}


type RecommendationRepository interface {
	// User preferences
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	SaveUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error

	// Activity tracking
	RecordActivity(ctx context.Context, activity *domain.UserActivity) error

	// Booking history
	GetUserBookedBathhouses(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)
	// GetUserBookedBathhousesWithDates returns booked bathhouses with their booking dates for recency calculation
	GetUserBookedBathhousesWithDates(ctx context.Context, userID uuid.UUID, limit int) ([]domain.BookedBathhouseWithDate, error)

	// Collaborative filtering
	GetSimilarUsers(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)

	// Popularity
	GetPopularBathhouses(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error)

	// Similarity based on amenities and location
	GetSimilarBathhouses(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error)
}


type ClientReviewRepository interface {
	Create(ctx context.Context, review *domain.ClientReview) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientReview, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.ClientReview, error)
	Update(ctx context.Context, review *domain.ClientReview) error
	ListByClient(ctx context.Context, clientID uuid.UUID, onlyRevealed bool, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error)
	ListUnrevealedPastDeadline(ctx context.Context, now time.Time) ([]domain.ClientReview, error)
	RevealByID(ctx context.Context, id uuid.UUID) error
}

