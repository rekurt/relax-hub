package mock

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// SocialAccountRepo is an in-memory mock implementation of repository.SocialAccountRepository.
type SocialAccountRepo struct {
	mu       sync.RWMutex
	accounts map[uuid.UUID]*domain.SocialAccount
}

func NewSocialAccountRepo() *SocialAccountRepo {
	return &SocialAccountRepo{accounts: make(map[uuid.UUID]*domain.SocialAccount)}
}

func (r *SocialAccountRepo) Create(_ context.Context, account *domain.SocialAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	for _, a := range r.accounts {
		if a.Provider == account.Provider && a.ProviderID == account.ProviderID {
			return domain.ErrSocialAccountAlreadyLinked
		}
	}
	cp := *account
	r.accounts[account.ID] = &cp
	return nil
}

func (r *SocialAccountRepo) GetByProviderAndID(_ context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, a := range r.accounts {
		if a.Provider == provider && a.ProviderID == providerID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrSocialAccountNotFound
}

func (r *SocialAccountRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.SocialAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SocialAccount
	for _, a := range r.accounts {
		if a.UserID == userID {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (r *SocialAccountRepo) Delete(_ context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, a := range r.accounts {
		if a.UserID == userID && a.Provider == provider {
			delete(r.accounts, id)
			return nil
		}
	}
	return domain.ErrSocialAccountNotFound
}
