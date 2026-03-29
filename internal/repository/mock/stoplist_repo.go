package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type StoplistRepo struct {
	mu             sync.RWMutex
	entries        map[uuid.UUID]*domain.StoplistEntry
	duplicateCount int
}

func NewStoplistRepo() repository.StoplistRepository {
	return &StoplistRepo{
		entries: make(map[uuid.UUID]*domain.StoplistEntry),
	}
}

func (r *StoplistRepo) Create(_ context.Context, entry *domain.StoplistEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.BlockedAt.IsZero() {
		entry.BlockedAt = time.Now()
	}

	cp := *entry
	r.entries[entry.ID] = &cp
	return nil
}

func (r *StoplistRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.entries[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.entries, id)
	return nil
}

func (r *StoplistRepo) List(_ context.Context, filter domain.StoplistFilter) (*domain.PaginatedResult[domain.StoplistEntry], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.StoplistEntry
	for _, e := range r.entries {
		if filter.Phone != "" && e.Phone != filter.Phone {
			continue
		}
		if filter.Email != "" && e.Email != filter.Email {
			continue
		}
		cp := *e
		filtered = append(filtered, cp)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].BlockedAt.After(filtered[j].BlockedAt)
	})

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return paginate(filtered, page, pageSize), nil
}

func (r *StoplistRepo) IsBlocked(_ context.Context, phone, email, inn, bankCardNumber string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.entries {
		if phone != "" && e.Phone == phone {
			return true, nil
		}
		if email != "" && e.Email == email {
			return true, nil
		}
		if inn != "" && e.INN == inn {
			return true, nil
		}
		if bankCardNumber != "" && e.BankCardNumber == bankCardNumber {
			return true, nil
		}
	}
	return false, nil
}

func (r *StoplistRepo) CountDuplicateOwners(_ context.Context, _, _, _ string, _ uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.duplicateCount, nil
}

// SetDuplicateCount is a test helper to set a fixed return value for CountDuplicateOwners.
func (r *StoplistRepo) SetDuplicateCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.duplicateCount = count
}
