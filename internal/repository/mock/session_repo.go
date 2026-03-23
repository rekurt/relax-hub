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

type SessionRepo struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*domain.Session
}

func NewSessionRepo() repository.SessionRepository {
	return &SessionRepo{
		sessions: make(map[uuid.UUID]*domain.Session),
	}
}

func (r *SessionRepo) Create(_ context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.LastActiveAt.IsZero() {
		session.LastActiveAt = now
	}
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = now.Add(domain.SessionMaxAge)
	}

	cp := *session
	r.sessions[session.ID] = &cp
	return nil
}

func (r *SessionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *SessionRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var result []domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID && s.ExpiresAt.After(now) {
			cp := *s
			result = append(result, cp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastActiveAt.After(result[j].LastActiveAt)
	})

	return result, nil
}

func (r *SessionRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[id]; !ok {
		return domain.ErrSessionNotFound
	}
	delete(r.sessions, id)
	return nil
}

func (r *SessionRepo) DeleteAllExcept(_ context.Context, userID uuid.UUID, exceptID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, s := range r.sessions {
		if s.UserID == userID && id != exceptID {
			delete(r.sessions, id)
		}
	}
	return nil
}

func (r *SessionRepo) UpdateLastActive(_ context.Context, id uuid.UUID, lastActiveAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[id]
	if !ok {
		return domain.ErrSessionNotFound
	}
	s.LastActiveAt = lastActiveAt
	s.ExpiresAt = lastActiveAt.Add(domain.SessionMaxAge)
	return nil
}

func (r *SessionRepo) DeleteExpired(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var count int64
	for id, s := range r.sessions {
		if !s.ExpiresAt.After(now) {
			delete(r.sessions, id)
			count++
		}
	}
	return count, nil
}
