package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type DeviceTokenRepo struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*domain.DeviceToken
}

func NewDeviceTokenRepo() repository.DeviceTokenRepository {
	return &DeviceTokenRepo{
		tokens: make(map[uuid.UUID]*domain.DeviceToken),
	}
}

func (r *DeviceTokenRepo) Create(_ context.Context, token *domain.DeviceToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}

	// Upsert: if token string already exists, update it
	for _, existing := range r.tokens {
		if existing.Token == token.Token {
			existing.UserID = token.UserID
			existing.Platform = token.Platform
			existing.UpdatedAt = now
			return nil
		}
	}

	cp := *token
	r.tokens[token.ID] = &cp
	return nil
}

func (r *DeviceTokenRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tokens[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.tokens, id)
	return nil
}

func (r *DeviceTokenRepo) DeleteByToken(_ context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, dt := range r.tokens {
		if dt.Token == token {
			delete(r.tokens, id)
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *DeviceTokenRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.DeviceToken
	for _, dt := range r.tokens {
		if dt.UserID == userID {
			cp := *dt
			result = append(result, cp)
		}
	}
	return result, nil
}
