package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type BathhouseRepository interface {
	Create(ctx context.Context, bh *domain.Bathhouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Bathhouse, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error)
	Update(ctx context.Context, bh *domain.Bathhouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListIDsByOwner(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error
	UpdateBayesianRating(ctx context.Context, bathhouseID uuid.UUID, bayesianRating float64) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error
	UpdatePhotoVerified(ctx context.Context, id uuid.UUID, verified bool) error
	GetCalendarToken(ctx context.Context, bathhouseID uuid.UUID) (string, error)
	SetCalendarToken(ctx context.Context, bathhouseID uuid.UUID, token string) error
	GetByCalendarToken(ctx context.Context, token string) (*domain.Bathhouse, error)
	SuggestNames(ctx context.Context, filter SuggestionFilter) ([]string, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	UpdateRankingFields(ctx context.Context, id uuid.UUID, conversionRate, occupancyRate float64) error
	UpdateResponseRate(ctx context.Context, id uuid.UUID, responseRate float64, avgResponseMinutes int, lowResponseRateSince *time.Time) error
	ListRequestModeBathhouses(ctx context.Context) ([]domain.Bathhouse, error)
	GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error)
}

// SuggestionFilter specifies parameters for name-based suggestions.
type SuggestionFilter struct {
	Query string
	Limit int
}


type CityRepository interface {
	Create(ctx context.Context, city *domain.City) error
	GetAll(ctx context.Context) ([]domain.City, error)
	GetBySlug(ctx context.Context, slug string) (*domain.City, error)
	GetByID(ctx context.Context, id int64) (*domain.City, error)
	Update(ctx context.Context, city *domain.City) error
	Delete(ctx context.Context, id int64) error
}


type AmenityRepository interface {
	Create(ctx context.Context, amenity *domain.Amenity) error
	Update(ctx context.Context, amenity *domain.Amenity) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Amenity, error)
	ListAll(ctx context.Context) ([]domain.Amenity, error)
}


type ObjectTypeRepository interface {
	Create(ctx context.Context, objType *domain.ObjectType) error
	Update(ctx context.Context, objType *domain.ObjectType) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ObjectType, error)
	ListAll(ctx context.Context) ([]domain.ObjectType, error)
}


type BathhousePhotoRepository interface {
	Create(ctx context.Context, photo *domain.BathhousePhoto) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BathhousePhoto, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	ListVerifiedByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PhotoStatus, verifiedByID *uuid.UUID, rejectionReason string) error
	Reorder(ctx context.Context, bathhouseID uuid.UUID, photoIDs []uuid.UUID) error
	ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error)
}


type HolidayRepository interface {
	Create(ctx context.Context, holiday *domain.Holiday) error
	Update(ctx context.Context, holiday *domain.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error)
	ListAll(ctx context.Context) ([]domain.Holiday, error)
	ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error)
	IsHoliday(ctx context.Context, date time.Time, region string) (*domain.Holiday, error)
	GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error)
	SetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID, multiplier float64) error
}


type SlotBlockRepository interface {
	Create(ctx context.Context, block *domain.SlotBlock) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SlotBlock, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SlotBlock, error)
	GetByExternalID(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource, externalID string) (*domain.SlotBlock, error)
	DeleteBySource(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource) error
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.SlotBlock, error)
	HasOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
}


type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	List(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
}


type ExternalCalendarRepository interface {
	Create(ctx context.Context, cal *domain.ExternalCalendar) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ExternalCalendar, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error)
	ListAll(ctx context.Context) ([]domain.ExternalCalendar, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, syncedAt time.Time, lastError string) error
}

