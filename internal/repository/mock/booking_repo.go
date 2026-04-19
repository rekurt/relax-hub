package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// BookingRepo is an in-memory mock implementation of repository.BookingRepository.
type BookingRepo struct {
	mu               sync.RWMutex
	bookings         map[uuid.UUID]*domain.Booking
	bathhouseRegions map[uuid.UUID]string // test helper: bathhouseID -> region
}

func NewBookingRepo() *BookingRepo {
	return &BookingRepo{bookings: make(map[uuid.UUID]*domain.Booking)}
}

func (r *BookingRepo) Create(_ context.Context, booking *domain.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}
	now := time.Now()
	if booking.CreatedAt.IsZero() {
		booking.CreatedAt = now
	}
	if booking.UpdatedAt.IsZero() {
		booking.UpdatedAt = now
	}
	cp := *booking
	r.bookings[booking.ID] = &cp
	return nil
}

func (r *BookingRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bookings[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *BookingRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Booking
	for _, b := range r.bookings {
		if b.UserID == userID {
			items = append(items, *b)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *BookingRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Booking
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID {
			items = append(items, *b)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *BookingRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.BookingStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[id]
	if !ok {
		return domain.ErrNotFound
	}
	b.Status = status
	b.UpdatedAt = time.Now()
	return nil
}

func isActiveBookingStatus(s domain.BookingStatus) bool {
	return s == domain.BookingPending || s == domain.BookingPendingOwner || s == domain.BookingConfirmed
}

func (r *BookingRepo) CheckAvailability(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID &&
			isActiveBookingStatus(b.Status) &&
			b.StartTime.Before(endTime) && b.EndTime.After(startTime) {
			return false, nil
		}
	}
	return true, nil
}

func (r *BookingRepo) CheckAvailabilityExcluding(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time, excludeBookingID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.bookings {
		if b.ID == excludeBookingID {
			continue
		}
		if b.BathhouseID == bathhouseID &&
			isActiveBookingStatus(b.Status) &&
			b.StartTime.Before(endTime) && b.EndTime.After(startTime) {
			return false, nil
		}
	}
	return true, nil
}

func (r *BookingRepo) CreateWithAvailabilityCheck(_ context.Context, booking *domain.Booking, checkStart, checkEnd time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check availability
	for _, b := range r.bookings {
		if b.BathhouseID == booking.BathhouseID &&
			isActiveBookingStatus(b.Status) &&
			b.StartTime.Before(checkEnd) && b.EndTime.After(checkStart) {
			return domain.ErrSlotUnavailable
		}
	}

	// Create booking
	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}
	now := time.Now()
	if booking.CreatedAt.IsZero() {
		booking.CreatedAt = now
	}
	if booking.UpdatedAt.IsZero() {
		booking.UpdatedAt = now
	}
	cp := *booking
	r.bookings[booking.ID] = &cp
	return nil
}

func (r *BookingRepo) GetOverlapping(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID &&
			isActiveBookingStatus(b.Status) &&
			b.StartTime.Before(endTime) && b.EndTime.After(startTime) {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *BookingRepo) CountActiveByBathhouse(_ context.Context, bathhouseID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID && isActiveBookingStatus(b.Status) {
			count++
		}
	}
	return count, nil
}

func (r *BookingRepo) CountActiveByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, b := range r.bookings {
		if b.UserID == userID && isActiveBookingStatus(b.Status) {
			count++
		}
	}
	return count, nil
}

func (r *BookingRepo) ListAll(_ context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var filtered []domain.Booking
	for _, b := range r.bookings {
		if filter.UserID != nil && b.UserID != *filter.UserID {
			continue
		}
		if filter.BathhouseID != nil && b.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && b.Status != *filter.Status {
			continue
		}
		if filter.FromDate != nil && b.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && b.CreatedAt.After(*filter.ToDate) {
			continue
		}
		filtered = append(filtered, *b)
	}
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	totalPages := int(total+int64(pageSize)-1) / pageSize
	return &domain.PaginatedResult[domain.Booking]{
		Items:      filtered[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *BookingRepo) GetUserStats(_ context.Context, userID uuid.UUID) (*domain.UserBookingStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var stats domain.UserBookingStats
	for _, b := range r.bookings {
		if b.UserID == userID && b.Status == domain.BookingCompleted {
			stats.TotalVisits++
			stats.TotalSpent += b.TotalPrice
		}
	}
	if stats.TotalVisits > 0 {
		stats.AvgCheck = stats.TotalSpent / int64(stats.TotalVisits)
	}
	return &stats, nil
}

func (r *BookingRepo) Update(_ context.Context, booking *domain.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bookings[booking.ID]; !ok {
		return domain.ErrNotFound
	}
	booking.UpdatedAt = time.Now()
	cp := *booking
	r.bookings[booking.ID] = &cp
	return nil
}

func (r *BookingRepo) ListTimedOutRequests(_ context.Context) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	now := time.Now()
	for _, b := range r.bookings {
		if b.Status != domain.BookingPendingOwner {
			continue
		}
		// We can't know the bathhouse's RequestTimeout here, so we use a default check.
		// In practice, the postgres impl joins with bathhouses table.
		// For mock testing, we consider bookings older than 24h as timed out.
		if now.Sub(b.CreatedAt).Hours() >= 24 {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *BookingRepo) UpdateCheckin(_ context.Context, bookingID uuid.UUID, checkedInAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.CheckedInAt = checkedInAt
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BookingRepo) UpdateCheckout(_ context.Context, bookingID uuid.UUID, checkedOutAt *time.Time, status domain.BookingStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.CheckedOutAt = checkedOutAt
	b.Status = status
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BookingRepo) ListConfirmedWithoutCheckin(_ context.Context, noShowCutoff time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	for _, b := range r.bookings {
		if b.Status == domain.BookingConfirmed && b.CheckedInAt == nil && b.StartTime.Before(noShowCutoff) {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *BookingRepo) ListUpcoming(_ context.Context, from, to time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	for _, b := range r.bookings {
		if b.Status == domain.BookingConfirmed &&
			!b.StartTime.Before(from) &&
			!b.StartTime.After(to) {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *BookingRepo) UpdateModification(_ context.Context, bookingID uuid.UUID, startTime, endTime time.Time, guestCount int, totalPrice, addOnTotal, basePrice, longSessionDiscount, extraGuestSurcharge, lastMinuteDiscount, serviceFeeAmount int64, modificationCount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.StartTime = startTime
	b.EndTime = endTime
	b.GuestCount = guestCount
	b.TotalPrice = totalPrice
	b.AddOnTotal = addOnTotal
	b.BasePrice = basePrice
	b.LongSessionDiscount = longSessionDiscount
	b.ExtraGuestSurcharge = extraGuestSurcharge
	b.LastMinuteDiscount = lastMinuteDiscount
	b.ServiceFeeAmount = serviceFeeAmount
	b.ModificationCount = modificationCount
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BookingRepo) UpdateEndTime(_ context.Context, bookingID uuid.UUID, oldEndTime, newEndTime time.Time, newTotalPrice int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	if !b.EndTime.Equal(oldEndTime) {
		return domain.ErrWalletConcurrentUpdate
	}
	b.EndTime = newEndTime
	b.TotalPrice = newTotalPrice
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BookingRepo) UpdateCancelledByOwner(_ context.Context, bookingID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.CancelledByOwner = true
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BookingRepo) CountOwnerCancellations(_ context.Context, ownerID uuid.UUID, since time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, b := range r.bookings {
		if b.CancelledByOwner && b.Status == domain.BookingCancelled && !b.UpdatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

func (r *BookingRepo) GetResponseStats(_ context.Context, bathhouseID uuid.UUID, since time.Time) (totalRequests int, respondedInTime int, avgResponseMinutes int, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var totalMinutes int
	responded := 0
	total := 0
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID && !b.CreatedAt.Before(since) && b.HoldID != nil {
			total++
			if b.Status == domain.BookingConfirmed || b.Status == domain.BookingRejected {
				responded++
				minutes := int(b.UpdatedAt.Sub(b.CreatedAt).Minutes())
				totalMinutes += minutes
			}
		}
	}
	avg := 0
	if responded > 0 {
		avg = totalMinutes / responded
	}
	return total, responded, avg, nil
}

func (r *BookingRepo) ListCompletedForReviewRequests(_ context.Context, checkedOutBefore time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Booking
	for _, b := range r.bookings {
		if b.Status == domain.BookingCompleted &&
			b.CheckedOutAt != nil &&
			!b.CheckedOutAt.After(checkedOutBefore) {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (r *BookingRepo) GetLastBookingDateByUser(_ context.Context, userID uuid.UUID) (*time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *time.Time
	for _, b := range r.bookings {
		if b.UserID == userID && (b.Status == domain.BookingCompleted || b.Status == domain.BookingConfirmed) {
			t := b.CreatedAt
			if latest == nil || t.After(*latest) {
				latest = &t
			}
		}
	}
	return latest, nil
}

func (r *BookingRepo) ListConfirmedByRegionAndDateRange(_ context.Context, region string, dateFrom, dateTo time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	for _, b := range r.bookings {
		if b.Status == domain.BookingConfirmed &&
			!b.StartTime.After(dateTo) &&
			!b.EndTime.Before(dateFrom) {
			if r.bathhouseRegions != nil {
				if r.bathhouseRegions[b.BathhouseID] == region {
					result = append(result, *b)
				}
			} else {
				result = append(result, *b)
			}
		}
	}
	return result, nil
}

// SetBathhouseRegion sets the region for a bathhouse ID (test helper for ListConfirmedByRegionAndDateRange).
func (r *BookingRepo) SetBathhouseRegion(bathhouseID uuid.UUID, region string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bathhouseRegions == nil {
		r.bathhouseRegions = make(map[uuid.UUID]string)
	}
	r.bathhouseRegions[bathhouseID] = region
}

func (r *BookingRepo) UpdateDeposit(_ context.Context, bookingID uuid.UUID, depositAmount int64, depositStatus domain.DepositStatus, depositExternalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.DepositAmount = depositAmount
	b.DepositStatus = depositStatus
	b.DepositExternalID = depositExternalID
	return nil
}

func (r *BookingRepo) UpdateDepositStatus(_ context.Context, bookingID uuid.UUID, depositStatus domain.DepositStatus, releasedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.ErrNotFound
	}
	b.DepositStatus = depositStatus
	b.DepositReleasedAt = releasedAt
	return nil
}

func (r *BookingRepo) ListHeldDepositsReadyForRelease(_ context.Context, checkedOutBefore time.Time) ([]domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Booking
	for _, b := range r.bookings {
		if b.DepositStatus == domain.DepositHeld &&
			b.CheckedOutAt != nil &&
			b.CheckedOutAt.Before(checkedOutBefore) {
			result = append(result, *b)
		}
	}
	return result, nil
}


