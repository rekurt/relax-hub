package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

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
