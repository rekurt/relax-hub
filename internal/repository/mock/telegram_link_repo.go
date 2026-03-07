package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// TelegramLinkRepo is an in-memory mock implementation of repository.TelegramLinkRepository
type TelegramLinkRepo struct {
	mu    sync.RWMutex
	links map[int64]*domain.TelegramLink
}

func NewTelegramLinkRepo() repository.TelegramLinkRepository {
	return &TelegramLinkRepo{
		links: make(map[int64]*domain.TelegramLink),
	}
}

func (r *TelegramLinkRepo) Create(_ context.Context, link *domain.TelegramLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}

	// Check if telegram ID already exists
	if _, exists := r.links[link.TelegramID]; exists {
		return domain.ErrAlreadyExists
	}

	// Check if user already has a link
	for _, existing := range r.links {
		if existing.UserID == link.UserID {
			return domain.ErrAlreadyExists
		}
	}

	link.LinkedAt = time.Now()
	cp := *link
	r.links[link.TelegramID] = &cp
	return nil
}

func (r *TelegramLinkRepo) GetByTelegramID(_ context.Context, telegramID int64) (*domain.TelegramLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.links[telegramID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *link
	return &cp, nil
}

func (r *TelegramLinkRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.TelegramLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, link := range r.links {
		if link.UserID == userID {
			cp := *link
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *TelegramLinkRepo) Delete(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for telegramID, link := range r.links {
		if link.UserID == userID {
			delete(r.links, telegramID)
			return nil
		}
	}
	return domain.ErrNotFound
}
