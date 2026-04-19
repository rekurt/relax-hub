package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
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

func (r *UserRepo) CountByCreatedAtRange(_ context.Context, from, to time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, u := range r.users {
		if !u.CreatedAt.Before(from) && !u.CreatedAt.After(to) {
			count++
		}
	}
	return count, nil
}


