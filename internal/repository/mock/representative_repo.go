package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

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


