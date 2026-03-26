package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// UserRepo is an in-memory mock implementation of repository.UserRepository.
type UserRepo struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (r *UserRepo) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	for _, u := range r.users {
		if u.Email != "" && u.Email == user.Email {
			return domain.ErrAlreadyExists
		}
		if u.Phone != "" && u.Phone == user.Phone {
			return domain.ErrAlreadyExists
		}
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	cp := *user
	r.users[user.ID] = &cp
	return nil
}

func (r *UserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepo) GetByPhone(_ context.Context, phone string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Phone == phone && u.Phone != "" {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepo) Update(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return domain.ErrNotFound
	}
	user.UpdatedAt = time.Now()
	cp := *user
	r.users[user.ID] = &cp
	return nil
}

func (r *UserRepo) List(_ context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		all = append(all, *u)
	}

	return paginate(all, page, pageSize), nil
}

func (r *UserRepo) SetActive(_ context.Context, id uuid.UUID, active bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.IsActive = active
	u.UpdatedAt = time.Now()
	return nil
}

func (r *UserRepo) GetPublicProfile(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if !u.IsActive {
		return nil, domain.ErrNotFound
	}
	return &domain.UserProfile{
		ID:          u.ID,
		Name:        u.Name,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		CityName:    "", // Mock doesn't compute city name
		MemberSince: u.CreatedAt,
		ReviewCount: 0, // Mock doesn't compute review count - use integration tests for validation
		VisitCount:  0, // Mock doesn't compute visit count - use integration tests for validation
		AvgRating:   0, // Mock doesn't compute average rating - use integration tests for validation
	}, nil
}

func (r *UserRepo) GetByReferralCode(_ context.Context, code string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ReferralCode == code {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *UserRepo) UpdateReferralCode(_ context.Context, userID uuid.UUID, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return domain.ErrNotFound
	}
	// Check uniqueness
	for _, other := range r.users {
		if other.ID != userID && other.ReferralCode == code {
			return domain.ErrAlreadyExists
		}
	}
	u.ReferralCode = code
	u.UpdatedAt = time.Now()
	return nil
}

func (r *UserRepo) SetDeletionSchedule(_ context.Context, userID uuid.UUID, requestedAt, scheduledAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return domain.ErrNotFound
	}
	u.DeletionRequestedAt = requestedAt
	u.DeletionScheduledAt = scheduledAt
	u.UpdatedAt = time.Now()
	return nil
}

func (r *UserRepo) ListPendingDeletions(_ context.Context, before time.Time) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.User
	for _, u := range r.users {
		if u.DeletionScheduledAt != nil && !u.DeletionScheduledAt.After(before) {
			cp := *u
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *UserRepo) AnonymizeUser(_ context.Context, userID uuid.UUID, anonEmail string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return domain.ErrNotFound
	}
	u.Name = "Deleted User"
	u.Email = anonEmail
	u.Phone = ""
	u.PhoneVerified = false
	u.PasswordHash = ""
	u.AvatarURL = ""
	u.Bio = ""
	u.TOTPSecret = ""
	u.TwoFAMethod = domain.TwoFANone
	u.ReferralCode = ""
	u.IsActive = false
	u.DeletionRequestedAt = nil
	u.DeletionScheduledAt = nil
	u.UpdatedAt = time.Now()
	return nil
}

// CityRepo is an in-memory mock implementation of repository.CityRepository.
type CityRepo struct {
	mu     sync.RWMutex
	cities map[int64]*domain.City
	nextID int64
}

func NewCityRepo() *CityRepo {
	return &CityRepo{cities: make(map[int64]*domain.City), nextID: 1}
}

func (r *CityRepo) Create(_ context.Context, city *domain.City) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.cities {
		if c.Slug == city.Slug {
			return domain.ErrAlreadyExists
		}
	}
	city.ID = r.nextID
	r.nextID++
	cp := *city
	r.cities[city.ID] = &cp
	return nil
}

func (r *CityRepo) GetAll(_ context.Context) ([]domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.City, 0, len(r.cities))
	for _, c := range r.cities {
		result = append(result, *c)
	}
	return result, nil
}

func (r *CityRepo) GetBySlug(_ context.Context, slug string) (*domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.cities {
		if c.Slug == slug {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *CityRepo) GetByID(_ context.Context, id int64) (*domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cities[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *CityRepo) Update(_ context.Context, city *domain.City) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cities[city.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *city
	r.cities[city.ID] = &cp
	return nil
}

func (r *CityRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cities[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.cities, id)
	return nil
}

// BookingRepo is an in-memory mock implementation of repository.BookingRepository.
type BookingRepo struct {
	mu       sync.RWMutex
	bookings map[uuid.UUID]*domain.Booking
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

// RepresentativeRepo is an in-memory mock implementation of repository.RepresentativeRepository.
type RepresentativeRepo struct {
	mu   sync.RWMutex
	reps map[uuid.UUID]*domain.Representative
}

func NewRepresentativeRepo() *RepresentativeRepo {
	return &RepresentativeRepo{reps: make(map[uuid.UUID]*domain.Representative)}
}

func (r *RepresentativeRepo) Create(_ context.Context, rep *domain.Representative) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rep.ID == uuid.Nil {
		rep.ID = uuid.New()
	}
	for _, existing := range r.reps {
		if existing.UserID == rep.UserID && existing.BathhouseID == rep.BathhouseID {
			return domain.ErrAlreadyExists
		}
	}
	rep.CreatedAt = time.Now()
	cp := *rep
	r.reps[rep.ID] = &cp
	return nil
}

func (r *RepresentativeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Representative, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rep, ok := r.reps[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rep
	return &cp, nil
}

func (r *RepresentativeRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reps[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.reps, id)
	return nil
}

func (r *RepresentativeRepo) GetByUserAndBathhouse(_ context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rep := range r.reps {
		if rep.UserID == userID && rep.BathhouseID == bathhouseID {
			cp := *rep
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *RepresentativeRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Representative
	for _, rep := range r.reps {
		if rep.BathhouseID == bathhouseID {
			result = append(result, *rep)
		}
	}
	return result, nil
}

func (r *RepresentativeRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.Representative, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Representative
	for _, rep := range r.reps {
		if rep.UserID == userID {
			result = append(result, *rep)
		}
	}
	return result, nil
}

func (r *RepresentativeRepo) ListBathhouseIDsByUser(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []uuid.UUID
	for _, rep := range r.reps {
		if rep.UserID == userID {
			ids = append(ids, rep.BathhouseID)
		}
	}
	return ids, nil
}

// ReviewRepo is an in-memory mock implementation of repository.ReviewRepository.
type ReviewRepo struct {
	mu      sync.RWMutex
	reviews map[uuid.UUID]*domain.Review
}

func NewReviewRepo() *ReviewRepo {
	return &ReviewRepo{reviews: make(map[uuid.UUID]*domain.Review)}
}

func (r *ReviewRepo) Create(_ context.Context, review *domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	for _, existing := range r.reviews {
		if existing.UserID == review.UserID && existing.BookingID == review.BookingID {
			return domain.ErrAlreadyExists
		}
	}
	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now
	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ReviewRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rev, ok := r.reviews[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rev
	return &cp, nil
}

func (r *ReviewRepo) Update(_ context.Context, review *domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[review.ID]; !ok {
		return domain.ErrNotFound
	}
	review.UpdatedAt = time.Now()
	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ReviewRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.reviews, id)
	return nil
}

func (r *ReviewRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if rev.BathhouseID == bathhouseID && rev.Status == domain.ReviewStatusApproved {
			items = append(items, *rev)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *ReviewRepo) ListByBathhouseFiltered(_ context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if filter.BathhouseID != nil && rev.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && rev.Status != *filter.Status {
			continue
		}
		if filter.MinRating != nil && rev.Rating < *filter.MinRating {
			continue
		}
		items = append(items, *rev)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *ReviewRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rev := range r.reviews {
		if rev.BookingID == bookingID {
			cp := *rev
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ReviewRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ReviewStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.Status = status
	rev.UpdatedAt = time.Now()
	return nil
}

func (r *ReviewRepo) AddOwnerResponse(_ context.Context, id uuid.UUID, response string, respondedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.OwnerResponse = response
	rev.OwnerResponseAt = &respondedAt
	rev.UpdatedAt = respondedAt
	return nil
}

func (r *ReviewRepo) GetUserReviewStats(_ context.Context, userID uuid.UUID) (*domain.UserReviewStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var stats domain.UserReviewStats
	var totalRating int
	for _, rev := range r.reviews {
		if rev.UserID == userID && rev.Status == domain.ReviewStatusApproved {
			stats.ReviewCount++
			totalRating += rev.Rating
		}
	}
	if stats.ReviewCount > 0 {
		stats.AvgRating = float64(totalRating) / float64(stats.ReviewCount)
	}
	return &stats, nil
}

func (r *ReviewRepo) UpdateStatusWithReasons(_ context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.Status = status
	rev.RejectionReasons = reasons
	rev.UpdatedAt = time.Now()
	return nil
}

func (r *ReviewRepo) CountPendingReviews(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, rev := range r.reviews {
		if rev.Status == domain.ReviewStatusPending {
			count++
		}
	}
	return count, nil
}

func (r *ReviewRepo) ListAllReviews(_ context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if filter.BathhouseID != nil && rev.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && rev.Status != *filter.Status {
			continue
		}
		if filter.MinRating != nil && rev.Rating < *filter.MinRating {
			continue
		}
		if filter.MaxRating != nil && rev.Rating > *filter.MaxRating {
			continue
		}
		if filter.FromDate != nil && rev.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && rev.CreatedAt.After(*filter.ToDate) {
			continue
		}
		items = append(items, *rev)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *ReviewRepo) GetCriteriaAverages(_ context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var avgs domain.ReviewCriteriaAverages
	var count float64
	for _, rev := range r.reviews {
		if rev.BathhouseID == bathhouseID && rev.Status == domain.ReviewStatusApproved && rev.Cleanliness != nil {
			avgs.AvgCleanliness += *rev.Cleanliness
			avgs.AvgAccuracy += *rev.Accuracy
			avgs.AvgCommunication += *rev.Communication
			avgs.AvgValueForMoney += *rev.ValueForMoney
			count++
		}
	}
	if count > 0 {
		avgs.AvgCleanliness /= count
		avgs.AvgAccuracy /= count
		avgs.AvgCommunication /= count
		avgs.AvgValueForMoney /= count
	}
	return &avgs, nil
}

func (r *ReviewRepo) GetPlatformAverageRating(_ context.Context) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total float64
	var count float64
	for _, rev := range r.reviews {
		if rev.Status == domain.ReviewStatusApproved {
			total += float64(rev.Rating)
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / count, nil
}

// FavoriteRepo is an in-memory mock implementation of repository.FavoriteRepository.
type FavoriteRepo struct {
	mu        sync.RWMutex
	favorites map[uuid.UUID]*domain.Favorite
}

func NewFavoriteRepo() *FavoriteRepo {
	return &FavoriteRepo{favorites: make(map[uuid.UUID]*domain.Favorite)}
}

func (r *FavoriteRepo) Add(_ context.Context, favorite *domain.Favorite) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if favorite.ID == uuid.Nil {
		favorite.ID = uuid.New()
	}
	for _, f := range r.favorites {
		if f.UserID == favorite.UserID && f.BathhouseID == favorite.BathhouseID {
			return domain.ErrAlreadyExists
		}
	}
	now := time.Now()
	favorite.CreatedAt = now
	cp := *favorite
	r.favorites[favorite.ID] = &cp
	return nil
}

func (r *FavoriteRepo) Remove(_ context.Context, userID, bathhouseID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, f := range r.favorites {
		if f.UserID == userID && f.BathhouseID == bathhouseID {
			delete(r.favorites, id)
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *FavoriteRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Favorite
	for _, f := range r.favorites {
		if f.UserID == userID {
			items = append(items, *f)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *FavoriteRepo) IsFavorite(_ context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, f := range r.favorites {
		if f.UserID == userID && f.BathhouseID == bathhouseID {
			return true, nil
		}
	}
	return false, nil
}

func (r *FavoriteRepo) CountByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, f := range r.favorites {
		if f.UserID == userID {
			count++
		}
	}
	return count, nil
}

// BathhouseRepo is an in-memory mock implementation of repository.BathhouseRepository.
type BathhouseRepo struct {
	mu         sync.RWMutex
	bathhouses map[uuid.UUID]*domain.Bathhouse
}

func NewBathhouseRepo() *BathhouseRepo {
	return &BathhouseRepo{bathhouses: make(map[uuid.UUID]*domain.Bathhouse)}
}

func (r *BathhouseRepo) Create(_ context.Context, bh *domain.Bathhouse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bh.ID == uuid.Nil {
		bh.ID = uuid.New()
	}
	now := time.Now()
	bh.CreatedAt = now
	bh.UpdatedAt = now
	cp := *bh
	r.bathhouses[bh.ID] = &cp
	return nil
}

func (r *BathhouseRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *bh
	return &cp, nil
}

func (r *BathhouseRepo) GetByAPIKey(_ context.Context, apiKey string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.ApiKey == apiKey {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) GetBySlug(_ context.Context, slug string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.Slug == slug {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) SlugExists(_ context.Context, slug string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (r *BathhouseRepo) Update(_ context.Context, bh *domain.Bathhouse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bathhouses[bh.ID]; !ok {
		return domain.ErrNotFound
	}
	bh.UpdatedAt = time.Now()
	cp := *bh
	r.bathhouses[bh.ID] = &cp
	return nil
}

func (r *BathhouseRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bathhouses[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.bathhouses, id)
	return nil
}

func (r *BathhouseRepo) List(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if filter.Status != nil {
			if bh.Status != *filter.Status {
				continue
			}
		} else if !filter.ShowAllStatuses && bh.Status != domain.BathhouseStatusActive {
			continue
		}
		if filter.CityID != nil && bh.CityID != *filter.CityID {
			continue
		}
		if filter.PriceMin != nil && bh.PricePerHour < *filter.PriceMin {
			continue
		}
		if filter.PriceMax != nil && bh.PricePerHour > *filter.PriceMax {
			continue
		}
		if filter.GuestCount != nil && bh.MaxGuests < *filter.GuestCount {
			continue
		}
		if filter.SearchQuery != nil && *filter.SearchQuery != "" {
			q := strings.ToLower(*filter.SearchQuery)
			name := strings.ToLower(bh.Name)
			desc := strings.ToLower(bh.Description)
			addr := strings.ToLower(bh.Address)
			// Exact substring match
			matched := strings.Contains(name, q) || strings.Contains(desc, q) || strings.Contains(addr, q)
			// Prefix match: check if any word in name/description starts with any query word
			if !matched {
				queryWords := strings.Fields(q)
				for _, qw := range queryWords {
					for _, field := range []string{name, desc, addr} {
						for _, fw := range strings.Fields(field) {
							if strings.HasPrefix(fw, qw) {
								matched = true
								break
							}
						}
						if matched {
							break
						}
					}
					if matched {
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		if filter.OpenNow != nil && *filter.OpenNow {
			if !isBathhouseOpenNow(bh) {
				continue
			}
		}
		items = append(items, *bh)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *BathhouseRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if bh.OwnerID == ownerID {
			items = append(items, *bh)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *BathhouseRepo) UpdateRating(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *BathhouseRepo) UpdateBayesianRating(_ context.Context, id uuid.UUID, bayesianRating float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.BayesianRating = bayesianRating
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.BathhouseStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.Status = status
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) UpdatePhotoVerified(_ context.Context, id uuid.UUID, verified bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.IsPhotoVerified = verified
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) GetCalendarToken(_ context.Context, bathhouseID uuid.UUID) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bh, ok := r.bathhouses[bathhouseID]
	if !ok {
		return "", domain.ErrNotFound
	}
	return bh.CalendarToken, nil
}

func (r *BathhouseRepo) SetCalendarToken(_ context.Context, bathhouseID uuid.UUID, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[bathhouseID]
	if !ok {
		return domain.ErrNotFound
	}
	bh.CalendarToken = token
	return nil
}

func (r *BathhouseRepo) GetByCalendarToken(_ context.Context, token string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.CalendarToken == token {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) SuggestNames(_ context.Context, filter repository.SuggestionFilter) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(filter.Query)
	var names []string
	for _, bh := range r.bathhouses {
		if bh.Status == domain.BathhouseStatusActive && strings.Contains(strings.ToLower(bh.Name), q) {
			names = append(names, bh.Name)
			if len(names) >= filter.Limit {
				break
			}
		}
	}
	return names, nil
}

func (r *BathhouseRepo) IncrementViewCount(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ViewCount++
	return nil
}

func (r *BathhouseRepo) UpdateRankingFields(_ context.Context, id uuid.UUID, conversionRate, occupancyRate float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ConversionRate = conversionRate
	bh.OccupancyRate = occupancyRate
	return nil
}

func (r *BathhouseRepo) ListIDsByOwner(_ context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ids []uuid.UUID
	for _, bh := range r.bathhouses {
		if bh.OwnerID == ownerID {
			ids = append(ids, bh.ID)
		}
	}
	return ids, nil
}

func (r *BathhouseRepo) UpdateResponseRate(_ context.Context, id uuid.UUID, responseRate float64, avgResponseMinutes int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ResponseRate = responseRate
	bh.AvgResponseTimeMinutes = avgResponseMinutes
	return nil
}

func (r *BathhouseRepo) ListRequestModeBathhouses(_ context.Context) ([]domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if bh.BookingMode == domain.BookingModeRequest {
			cp := *bh
			result = append(result, cp)
		}
	}
	return result, nil
}

func isBathhouseOpenNow(bh *domain.Bathhouse) bool {
	now := time.Now()
	d := now.Weekday()
	dayOfWeek := int(d) - 1
	if d == time.Sunday {
		dayOfWeek = 6
	}
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	for _, wh := range bh.WorkingHours {
		if wh.DayOfWeek != dayOfWeek {
			continue
		}
		if wh.OpenTime <= wh.CloseTime {
			// Normal schedule: e.g. 09:00 - 22:00
			if wh.OpenTime <= currentTime && wh.CloseTime > currentTime {
				return true
			}
		} else {
			// Overnight schedule: e.g. 20:00 - 06:00
			if wh.OpenTime <= currentTime || wh.CloseTime > currentTime {
				return true
			}
		}
	}
	return false
}

// NotificationRepo is an in-memory mock implementation of repository.NotificationRepository.
type NotificationRepo struct {
	mu            sync.RWMutex
	notifications map[uuid.UUID]*domain.Notification
	preferences   map[uuid.UUID]*domain.NotificationPreferences
}

func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{
		notifications: make(map[uuid.UUID]*domain.Notification),
		preferences:   make(map[uuid.UUID]*domain.NotificationPreferences),
	}
}

func (r *NotificationRepo) Create(_ context.Context, notification *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	if notification.Data == nil {
		notification.Data = make(map[string]string)
	}

	cp := *notification
	cpData := make(map[string]string, len(notification.Data))
	for k, v := range notification.Data {
		cpData[k] = v
	}
	cp.Data = cpData
	r.notifications[notification.ID] = &cp
	return nil
}

func (r *NotificationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.notifications[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	cpData := make(map[string]string, len(n.Data))
	for k, v := range n.Data {
		cpData[k] = v
	}
	cp.Data = cpData
	return &cp, nil
}

func (r *NotificationRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			items = append(items, *n)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *NotificationRepo) MarkAsRead(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	if n.IsRead {
		return domain.ErrNotFound
	}
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
	return nil
}

func (r *NotificationRepo) MarkAllAsRead(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, n := range r.notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			n.ReadAt = &now
		}
	}
	return nil
}

func (r *NotificationRepo) CountUnread(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, n := range r.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *NotificationRepo) GetPreferences(_ context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.preferences[userID]
	if !ok {
		defaults := domain.DefaultNotificationPreferences(userID)
		return &defaults, nil
	}
	cp := *p
	return &cp, nil
}

func (r *NotificationRepo) UpdatePreferences(_ context.Context, prefs *domain.NotificationPreferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *prefs
	r.preferences[prefs.UserID] = &cp
	return nil
}

func (r *NotificationRepo) HasRecentByType(_ context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, n := range r.notifications {
		if n.UserID == userID && n.Type == notifType && !n.CreatedAt.Before(since) {
			return true, nil
		}
	}
	return false, nil
}

// SocialAccountRepo is an in-memory mock implementation of repository.SocialAccountRepository.
type SocialAccountRepo struct {
	mu       sync.RWMutex
	accounts map[uuid.UUID]*domain.SocialAccount
}

func NewSocialAccountRepo() *SocialAccountRepo {
	return &SocialAccountRepo{accounts: make(map[uuid.UUID]*domain.SocialAccount)}
}

func (r *SocialAccountRepo) Create(_ context.Context, account *domain.SocialAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	for _, a := range r.accounts {
		if a.Provider == account.Provider && a.ProviderID == account.ProviderID {
			return domain.ErrSocialAccountAlreadyLinked
		}
	}
	cp := *account
	r.accounts[account.ID] = &cp
	return nil
}

func (r *SocialAccountRepo) GetByProviderAndID(_ context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, a := range r.accounts {
		if a.Provider == provider && a.ProviderID == providerID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrSocialAccountNotFound
}

func (r *SocialAccountRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.SocialAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SocialAccount
	for _, a := range r.accounts {
		if a.UserID == userID {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (r *SocialAccountRepo) Delete(_ context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, a := range r.accounts {
		if a.UserID == userID && a.Provider == provider {
			delete(r.accounts, id)
			return nil
		}
	}
	return domain.ErrSocialAccountNotFound
}

// RecommendationRepo is an in-memory mock implementation of repository.RecommendationRepository.
type RecommendationRepo struct {
	mu         sync.RWMutex
	preferences map[uuid.UUID]*domain.UserPreferences
	activities  []domain.UserActivity
	bookings    map[uuid.UUID][]uuid.UUID // userID -> list of bathhouse IDs
}

func NewRecommendationRepo() *RecommendationRepo {
	return &RecommendationRepo{
		preferences: make(map[uuid.UUID]*domain.UserPreferences),
		activities:  make([]domain.UserActivity, 0),
		bookings:    make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *RecommendationRepo) GetUserPreferences(_ context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefs, ok := r.preferences[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *prefs
	return &cp, nil
}

func (r *RecommendationRepo) SaveUserPreferences(_ context.Context, prefs *domain.UserPreferences) error {
	if err := prefs.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if prefs.UpdatedAt.IsZero() {
		prefs.UpdatedAt = time.Now()
	}
	cp := *prefs
	r.preferences[prefs.UserID] = &cp
	return nil
}

func (r *RecommendationRepo) RecordActivity(_ context.Context, activity *domain.UserActivity) error {
	if err := activity.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	r.activities = append(r.activities, *activity)
	return nil
}

func (r *RecommendationRepo) GetUserBookedBathhouses(_ context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	booked, ok := r.bookings[userID]
	if !ok {
		return []uuid.UUID{}, nil
	}

	if limit <= 0 {
		limit = 50
	}
	if len(booked) > limit {
		return booked[:limit], nil
	}
	return booked, nil
}

func (r *RecommendationRepo) GetUserBookedBathhousesWithDates(_ context.Context, userID uuid.UUID, limit int) ([]domain.BookedBathhouseWithDate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	booked, ok := r.bookings[userID]
	if !ok {
		return []domain.BookedBathhouseWithDate{}, nil
	}

	if limit <= 0 {
		limit = 50
	}

	// For mock, we return with current time as booking date
	// In reality, this would come from booking timestamps
	var results []domain.BookedBathhouseWithDate
	end := limit
	if len(booked) < limit {
		end = len(booked)
	}
	for i := 0; i < end; i++ {
		results = append(results, domain.BookedBathhouseWithDate{
			BathhouseID: booked[i],
			BookedAt:    time.Now().AddDate(0, 0, -i), // Mock data: each booking is 1 day older
		})
	}
	return results, nil
}

func (r *RecommendationRepo) GetSimilarUsers(_ context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	r.mu.RLock()

	if limit <= 0 {
		limit = 20
	}

	// Find users with overlapping bookings
	userBookings := make(map[uuid.UUID]bool)
	if booked, ok := r.bookings[userID]; ok {
		for _, bhID := range booked {
			userBookings[bhID] = true
		}
	}

	type userMatch struct {
		userID uuid.UUID
		score  int
	}
	matches := make([]userMatch, 0)

	for otherUserID, otherBookings := range r.bookings {
		if otherUserID == userID {
			continue
		}
		score := 0
		for _, bhID := range otherBookings {
			if userBookings[bhID] {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, userMatch{otherUserID, score})
		}
	}

	r.mu.RUnlock()

	// Sort by score descending (outside the lock)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	var result []uuid.UUID
	for i := 0; i < len(matches) && i < limit; i++ {
		result = append(result, matches[i].userID)
	}
	return result, nil
}

func (r *RecommendationRepo) GetPopularBathhouses(_ context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
	// Mock implementation returns empty list
	// In real implementation would require bathhouse repo access
	return []uuid.UUID{}, nil
}

func (r *RecommendationRepo) GetSimilarBathhouses(_ context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error) {
	// Mock implementation returns empty list
	// In real implementation would require bathhouse repo access
	return []uuid.UUID{}, nil
}

// SubscriptionRepo is an in-memory mock implementation of repository.SubscriptionRepository.
type SubscriptionRepo struct {
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]*domain.Subscription
}

func NewSubscriptionRepo() *SubscriptionRepo {
	return &SubscriptionRepo{subscriptions: make(map[uuid.UUID]*domain.Subscription)}
}

func (r *SubscriptionRepo) Create(_ context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	// Check if another active subscription exists for this bathhouse
	for _, existing := range r.subscriptions {
		if existing.BathhouseID == sub.BathhouseID && existing.Status == domain.SubscriptionActive {
			return domain.ErrAlreadyExists
		}
	}

	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now
	cp := *sub
	r.subscriptions[sub.ID] = &cp
	return nil
}

func (r *SubscriptionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sub, ok := r.subscriptions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *sub
	return &cp, nil
}

func (r *SubscriptionRepo) GetActiveBybathhouse(_ context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.BathhouseID == bathhouseID && sub.Status == domain.SubscriptionActive {
			if latest == nil || sub.CreatedAt.After(latest.CreatedAt) {
				cp := *sub
				latest = &cp
			}
		}
	}

	if latest == nil {
		return nil, domain.ErrNotFound
	}
	return latest, nil
}

func (r *SubscriptionRepo) Update(_ context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subscriptions[sub.ID]; !ok {
		return domain.ErrNotFound
	}

	sub.UpdatedAt = time.Now()
	cp := *sub
	r.subscriptions[sub.ID] = &cp
	return nil
}

func (r *SubscriptionRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.OwnerID == ownerID {
			items = append(items, *sub)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *SubscriptionRepo) GetExpiring(_ context.Context, before time.Time) ([]domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.Status == domain.SubscriptionActive && sub.EndDate != nil && sub.EndDate.Before(before) || sub.EndDate.Equal(before) {
			result = append(result, *sub)
		}
	}
	return result, nil
}

// PromotionRepo is an in-memory mock implementation of repository.PromotionRepository.
type PromotionRepo struct {
	mu         sync.RWMutex
	promotions map[uuid.UUID]*domain.Promotion
}

func NewPromotionRepo() *PromotionRepo {
	return &PromotionRepo{promotions: make(map[uuid.UUID]*domain.Promotion)}
}

func (r *PromotionRepo) Create(_ context.Context, promo *domain.Promotion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}

	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now
	cp := *promo
	r.promotions[promo.ID] = &cp
	return nil
}

func (r *PromotionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	promo, ok := r.promotions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *promo
	return &cp, nil
}

func (r *PromotionRepo) GetActiveBybathhouse(_ context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.Promotion
	for _, promo := range r.promotions {
		if promo.BathhouseID == bathhouseID && promo.Status == domain.PromotionActive {
			if latest == nil || promo.CreatedAt.After(latest.CreatedAt) {
				cp := *promo
				latest = &cp
			}
		}
	}

	if latest == nil {
		return nil, domain.ErrNotFound
	}
	return latest, nil
}

func (r *PromotionRepo) Update(_ context.Context, promo *domain.Promotion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.promotions[promo.ID]; !ok {
		return domain.ErrNotFound
	}

	cp := *promo
	r.promotions[promo.ID] = &cp
	return nil
}

func (r *PromotionRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Promotion
	// In a real app, we'd join with bathhouses to find owner
	// For mock, we return empty since we don't have bathhouse context
	// This would be populated by test setup

	return paginate(items, page, pageSize), nil
}

func (r *PromotionRepo) RecordImpression(_ context.Context, promotionID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promotions[promotionID]
	if !ok {
		return domain.ErrNotFound
	}
	promo.ImpressionCount++
	return nil
}

func (r *PromotionRepo) RecordClick(_ context.Context, promotionID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promotions[promotionID]
	if !ok {
		return domain.ErrNotFound
	}
	promo.ClickCount++
	return nil
}

// PricingRuleRepo is an in-memory mock implementation of repository.PricingRuleRepository.
type PricingRuleRepo struct {
	mu    sync.RWMutex
	rules map[uuid.UUID]*domain.PricingRule
}

func NewPricingRuleRepo() *PricingRuleRepo {
	return &PricingRuleRepo{rules: make(map[uuid.UUID]*domain.PricingRule)}
}

func (r *PricingRuleRepo) Create(_ context.Context, rule *domain.PricingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	rule.CreatedAt = time.Now()
	cp := *rule
	r.rules[rule.ID] = &cp
	return nil
}

func (r *PricingRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rule
	return &cp, nil
}

func (r *PricingRuleRepo) Update(_ context.Context, rule *domain.PricingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rules[rule.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *rule
	r.rules[rule.ID] = &cp
	return nil
}

func (r *PricingRuleRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rules[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.rules, id)
	return nil
}

func (r *PricingRuleRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PricingRule
	for _, rule := range r.rules {
		if rule.BathhouseID == bathhouseID {
			cp := *rule
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *PricingRuleRepo) GetActiveRules(_ context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PricingRule
	for _, rule := range r.rules {
		if rule.BathhouseID == bathhouseID && rule.IsActive {
			cp := *rule
			result = append(result, cp)
		}
	}
	return result, nil
}

// LoyaltyRepo is an in-memory mock implementation of repository.LoyaltyRepository.
type LoyaltyRepo struct {
	mu           sync.RWMutex
	accounts     map[uuid.UUID]*domain.LoyaltyAccount
	transactions []domain.LoyaltyTransaction
}

func NewLoyaltyRepo() *LoyaltyRepo {
	return &LoyaltyRepo{
		accounts:     make(map[uuid.UUID]*domain.LoyaltyAccount),
		transactions: make([]domain.LoyaltyTransaction, 0),
	}
}

func (r *LoyaltyRepo) GetAccount(_ context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *acc
	return &cp, nil
}

func (r *LoyaltyRepo) CreateAccount(_ context.Context, account *domain.LoyaltyAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[account.UserID]; ok {
		return domain.ErrAlreadyExists
	}

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now
	cp := *account
	r.accounts[account.UserID] = &cp
	return nil
}

func (r *LoyaltyRepo) AddPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.Points += amount
	acc.TotalEarned += amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) SpendPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	if acc.Points < amount {
		return domain.ErrInsufficientPoints
	}
	acc.Points -= amount
	acc.TotalSpent += amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) RefundPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	if acc.TotalSpent < amount {
		return fmt.Errorf("%w: refund amount exceeds total spent", domain.ErrInvalidInput)
	}
	acc.Points += amount
	acc.TotalSpent -= amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) IncrementVisitCount(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.VisitCount++
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) UpdateLevel(_ context.Context, userID uuid.UUID, level domain.LoyaltyLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.Level = level
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) ListTransactions(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.LoyaltyTransaction
	for i := len(r.transactions) - 1; i >= 0; i-- {
		if r.transactions[i].UserID == userID {
			items = append(items, r.transactions[i])
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *LoyaltyRepo) CreateTransaction(_ context.Context, tx *domain.LoyaltyTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.CreatedAt = time.Now()
	r.transactions = append(r.transactions, *tx)
	return nil
}
