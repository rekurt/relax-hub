package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type OfferRepo struct {
	mu          sync.RWMutex
	acceptances map[uuid.UUID]*domain.OfferAcceptance
}

func NewOfferRepo() *OfferRepo {
	return &OfferRepo{
		acceptances: make(map[uuid.UUID]*domain.OfferAcceptance),
	}
}

var _ repository.OfferRepository = (*OfferRepo)(nil)

func (r *OfferRepo) Create(_ context.Context, acceptance *domain.OfferAcceptance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if acceptance.ID == uuid.Nil {
		acceptance.ID = uuid.New()
	}
	now := time.Now()
	if acceptance.CreatedAt.IsZero() {
		acceptance.CreatedAt = now
	}
	if acceptance.AcceptedAt.IsZero() {
		acceptance.AcceptedAt = now
	}

	// Check unique constraint (user_id, offer_version)
	for _, a := range r.acceptances {
		if a.UserID == acceptance.UserID && a.OfferVersion == acceptance.OfferVersion {
			return domain.ErrOfferAlreadyAccepted
		}
	}

	cp := *acceptance
	r.acceptances[acceptance.ID] = &cp
	return nil
}

func (r *OfferRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.OfferAcceptance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.OfferAcceptance
	for _, a := range r.acceptances {
		if a.UserID == userID {
			if latest == nil || a.AcceptedAt.After(latest.AcceptedAt) {
				cp := *a
				latest = &cp
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrOfferNotFound
	}
	return latest, nil
}

func (r *OfferRepo) GetByUserAndVersion(_ context.Context, userID uuid.UUID, version string) (*domain.OfferAcceptance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, a := range r.acceptances {
		if a.UserID == userID && a.OfferVersion == version {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrOfferNotFound
}

func (r *OfferRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.OfferAcceptance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.OfferAcceptance
	for _, a := range r.acceptances {
		if a.UserID == userID {
			result = append(result, *a)
		}
	}
	return result, nil
}
