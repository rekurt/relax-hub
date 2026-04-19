package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// ForceMajeureRepo is an in-memory mock implementation of repository.ForceMajeureRepository.
type ForceMajeureRepo struct {
	mu     sync.RWMutex
	events []domain.ForceMajeureEvent
}

func NewForceMajeureRepo() *ForceMajeureRepo {
	return &ForceMajeureRepo{}
}

func (r *ForceMajeureRepo) Create(_ context.Context, event *domain.ForceMajeureEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	event.CreatedAt = time.Now()
	r.events = append(r.events, *event)
	return nil
}

func (r *ForceMajeureRepo) List(_ context.Context) ([]domain.ForceMajeureEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.ForceMajeureEvent, len(r.events))
	copy(result, r.events)
	// Return in reverse order (newest first)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}


