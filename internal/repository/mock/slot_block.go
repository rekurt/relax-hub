package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// SlotBlockRepo is an in-memory mock implementation of repository.SlotBlockRepository.
type SlotBlockRepo struct {
	mu     sync.RWMutex
	blocks map[uuid.UUID]*domain.SlotBlock
}

func NewSlotBlockRepo() *SlotBlockRepo {
	return &SlotBlockRepo{blocks: make(map[uuid.UUID]*domain.SlotBlock)}
}

func (r *SlotBlockRepo) Create(_ context.Context, block *domain.SlotBlock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if block.ID == uuid.Nil {
		block.ID = uuid.New()
	}
	block.CreatedAt = time.Now()
	cp := *block
	r.blocks[block.ID] = &cp
	return nil
}

func (r *SlotBlockRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.blocks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.blocks, id)
	return nil
}

func (r *SlotBlockRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.SlotBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.blocks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *SlotBlockRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.SlotBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SlotBlock
	for _, b := range r.blocks {
		if b.BathhouseID == bathhouseID {
			cp := *b
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *SlotBlockRepo) GetByExternalID(_ context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource, externalID string) (*domain.SlotBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.blocks {
		if b.BathhouseID == bathhouseID && b.Source == source && b.ExternalID == externalID {
			cp := *b
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *SlotBlockRepo) DeleteBySource(_ context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, b := range r.blocks {
		if b.BathhouseID == bathhouseID && b.Source == source {
			delete(r.blocks, id)
		}
	}
	return nil
}

func (r *SlotBlockRepo) GetOverlapping(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.SlotBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SlotBlock
	for _, b := range r.blocks {
		if b.BathhouseID == bathhouseID && b.StartTime.Before(endTime) && b.EndTime.After(startTime) {
			cp := *b
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *SlotBlockRepo) HasOverlapping(_ context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.blocks {
		if b.BathhouseID == bathhouseID && b.StartTime.Before(endTime) && b.EndTime.After(startTime) {
			return true, nil
		}
	}
	return false, nil
}
