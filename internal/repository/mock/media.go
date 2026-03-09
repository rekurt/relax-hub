package mock

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// MediaRepo is an in-memory mock implementation of repository.MediaRepository
type MediaRepo struct {
	mu    sync.RWMutex
	media map[uuid.UUID]*domain.Media
}

func NewMediaRepo() *MediaRepo {
	return &MediaRepo{
		media: make(map[uuid.UUID]*domain.Media),
	}
}

// Ensure MediaRepo implements repository.MediaRepository
var _ repository.MediaRepository = (*MediaRepo)(nil)

func (r *MediaRepo) Create(_ context.Context, media *domain.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	if media.Status == "" {
		media.Status = domain.MediaStatusPending
	}
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now()
	}

	cp := *media
	r.media[media.ID] = &cp
	return nil
}

func (r *MediaRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Media, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.media[id]
	if !ok {
		return nil, domain.ErrMediaNotFound
	}
	cp := *m
	return &cp, nil
}

func (r *MediaRepo) ListByOwner(_ context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var filtered []domain.Media
	for _, m := range r.media {
		if m.OwnerType == ownerType && m.OwnerID == ownerID {
			cp := *m
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	totalCount := int64(len(filtered))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > int(totalCount) {
		offset = int(totalCount)
	}
	if end > int(totalCount) {
		end = int(totalCount)
	}

	return &domain.PaginatedResult[domain.Media]{
		Items:      filtered[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *MediaRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.media[id]; !ok {
		return domain.ErrMediaNotFound
	}
	delete(r.media, id)
	return nil
}

func (r *MediaRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.MediaStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.media[id]
	if !ok {
		return domain.ErrMediaNotFound
	}
	m.Status = status
	return nil
}

func (r *MediaRepo) CountByOwner(_ context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType *domain.MediaType) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, m := range r.media {
		if m.OwnerType == ownerType && m.OwnerID == ownerID {
			if mediaType != nil && m.Type != *mediaType {
				continue
			}
			count++
		}
	}
	return count, nil
}
