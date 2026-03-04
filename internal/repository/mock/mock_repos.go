package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
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
		if u.Email == user.Email {
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	all := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		all = append(all, *u)
	}

	total := int64(len(all))
	start := (page - 1) * pageSize
	if start >= len(all) {
		return &domain.PaginatedResult[domain.User]{
			Items:      nil,
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}, nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}

	return &domain.PaginatedResult[domain.User]{
		Items:      all[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
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
		MemberSince: u.CreatedAt,
	}, nil
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
	booking.CreatedAt = now
	booking.UpdatedAt = now
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

	return paginateBookings(items, page, pageSize), nil
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

	return paginateBookings(items, page, pageSize), nil
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

func (r *BookingRepo) CheckAvailability(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.bookings {
		if b.BathhouseID == bathhouseID &&
			(b.Status == domain.BookingPending || b.Status == domain.BookingConfirmed) &&
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
			(b.Status == domain.BookingPending || b.Status == domain.BookingConfirmed) &&
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
		if b.BathhouseID == bathhouseID &&
			(b.Status == domain.BookingPending || b.Status == domain.BookingConfirmed) {
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

func paginateBookings(items []domain.Booking, page, pageSize int) *domain.PaginatedResult[domain.Booking] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Booking]{
			Items: nil, TotalCount: total, Page: page, PageSize: pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return &domain.PaginatedResult[domain.Booking]{
		Items: items[start:end], TotalCount: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var items []domain.Review
	for _, rev := range r.reviews {
		if rev.BathhouseID == bathhouseID {
			items = append(items, *rev)
		}
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Review]{
			Items: nil, TotalCount: total, Page: page, PageSize: pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Review]{
		Items: items[start:end], TotalCount: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (r *ReviewRepo) ListByBathhouseFiltered(_ context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

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

	total := int64(len(items))
	start := (filter.Page - 1) * filter.PageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Review]{
			Items: nil, TotalCount: total, Page: filter.Page, PageSize: filter.PageSize,
			TotalPages: int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
		}, nil
	}
	end := start + filter.PageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Review]{
		Items: items[start:end], TotalCount: total, Page: filter.Page, PageSize: filter.PageSize,
		TotalPages: int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
	}, nil
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var items []domain.Favorite
	for _, f := range r.favorites {
		if f.UserID == userID {
			items = append(items, *f)
		}
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Favorite]{
			Items: nil, TotalCount: total, Page: page, PageSize: pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Favorite]{
		Items: items[start:end], TotalCount: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
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

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

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
			if !strings.Contains(strings.ToLower(bh.Name), q) && !strings.Contains(strings.ToLower(bh.Description), q) {
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

	total := int64(len(items))
	start := (filter.Page - 1) * filter.PageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Bathhouse]{
			Items: nil, TotalCount: total, Page: filter.Page, PageSize: filter.PageSize,
			TotalPages: int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
		}, nil
	}
	end := start + filter.PageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Bathhouse]{
		Items: items[start:end], TotalCount: total, Page: filter.Page, PageSize: filter.PageSize,
		TotalPages: int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
	}, nil
}

func (r *BathhouseRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var items []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if bh.OwnerID == ownerID {
			items = append(items, *bh)
		}
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Bathhouse]{
			Items: nil, TotalCount: total, Page: page, PageSize: pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Bathhouse]{
		Items: items[start:end], TotalCount: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (r *BathhouseRepo) UpdateRating(_ context.Context, _ uuid.UUID) error {
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var items []domain.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			items = append(items, *n)
		}
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[domain.Notification]{
			Items: nil, TotalCount: total, Page: page, PageSize: pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[domain.Notification]{
		Items: items[start:end], TotalCount: total, Page: page, PageSize: pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
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
