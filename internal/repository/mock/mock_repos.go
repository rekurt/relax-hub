package mock

import (
	"context"
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
	review.CreatedAt = time.Now()
	cp := *review
	r.reviews[review.ID] = &cp
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
