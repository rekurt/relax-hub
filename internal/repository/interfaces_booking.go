package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
	CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
	CheckAvailabilityExcluding(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time, excludeBookingID uuid.UUID) (bool, error)
	CreateWithAvailabilityCheck(ctx context.Context, booking *domain.Booking, checkStart, checkEnd time.Time) error
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error)
	CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserBookingStats, error)
	ListTimedOutRequests(ctx context.Context) ([]domain.Booking, error)
	UpdateCheckin(ctx context.Context, bookingID uuid.UUID, checkedInAt *time.Time) error
	UpdateCheckout(ctx context.Context, bookingID uuid.UUID, checkedOutAt *time.Time, status domain.BookingStatus) error
	ListConfirmedWithoutCheckin(ctx context.Context, noShowCutoff time.Time) ([]domain.Booking, error)
	ListUpcoming(ctx context.Context, from, to time.Time) ([]domain.Booking, error)
	UpdateModification(ctx context.Context, bookingID uuid.UUID, startTime, endTime time.Time, guestCount int, totalPrice, addOnTotal, basePrice, longSessionDiscount, extraGuestSurcharge, lastMinuteDiscount, serviceFeeAmount int64, modificationCount int) error
	UpdateEndTime(ctx context.Context, bookingID uuid.UUID, oldEndTime, newEndTime time.Time, newTotalPrice int64) error
	UpdateCancelledByOwner(ctx context.Context, bookingID uuid.UUID) error
	CountOwnerCancellations(ctx context.Context, ownerID uuid.UUID, since time.Time) (int, error)
	GetResponseStats(ctx context.Context, bathhouseID uuid.UUID, since time.Time) (totalRequests int, respondedInTime int, avgResponseMinutes int, err error)
	ListCompletedForReviewRequests(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error)
	GetLastBookingDateByUser(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	ListConfirmedByRegionAndDateRange(ctx context.Context, region string, dateFrom, dateTo time.Time) ([]domain.Booking, error)
	UpdateDeposit(ctx context.Context, bookingID uuid.UUID, depositAmount int64, depositStatus domain.DepositStatus, depositExternalID string) error
	UpdateDepositStatus(ctx context.Context, bookingID uuid.UUID, depositStatus domain.DepositStatus, releasedAt *time.Time) error
	ListHeldDepositsReadyForRelease(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error)
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	ListAll(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error)
}


type BookingModificationRequestRepository interface {
	Create(ctx context.Context, req *domain.BookingModificationRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error)
	GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingModificationRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ModificationRequestStatus, reason string) error
	ListExpired(ctx context.Context) ([]domain.BookingModificationRequest, error)
	ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error)
}


type ExtensionRequestRepository interface {
	Create(ctx context.Context, req *domain.BookingExtensionRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error)
	GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingExtensionRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ExtensionRequestStatus, reason string) error
	ListExpired(ctx context.Context) ([]domain.BookingExtensionRequest, error)
	ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error)
}


type AddOnRepository interface {
	Create(ctx context.Context, addon *domain.AddOn) error
	Update(ctx context.Context, addon *domain.AddOn) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AddOn, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	ListActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	CountByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	CreateBookingAddOn(ctx context.Context, ba *domain.BookingAddOn) error
	ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingAddOn, error)
}


type SavedSearchRepository interface {
	Create(ctx context.Context, search *domain.SavedSearch) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedSearch, error)
	ListWithNotifications(ctx context.Context) ([]domain.SavedSearch, error)
}


type EscrowRepository interface {
	Create(ctx context.Context, escrow *domain.Escrow) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Escrow, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Escrow, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EscrowStatus, releasedAt *time.Time) error
	ListMatured(ctx context.Context) ([]domain.Escrow, error) // status=held AND claim_period_ends_at < now
}


// BookingShareRepository manages shareable booking links.
type BookingShareRepository interface {
	Create(ctx context.Context, share *domain.BookingShare) error
	GetByToken(ctx context.Context, token string) (*domain.BookingShare, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

