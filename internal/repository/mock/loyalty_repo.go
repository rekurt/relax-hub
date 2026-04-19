package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// LoyaltyRepo is an in-memory mock implementation of repository.LoyaltyRepository.
type LoyaltyRepo struct {
	mu           sync.RWMutex
	accounts     map[uuid.UUID]*domain.LoyaltyAccount
	transactions []domain.LoyaltyTransaction
}

func NewLoyaltyRepo() *LoyaltyRepo {
	return &LoyaltyRepo{
		accounts:     make(map[uuid.UUID]*domain.LoyaltyAccount),
		transactions: make([]domain.LoyaltyTransaction, 0),
	}
}

func (r *LoyaltyRepo) GetAccount(_ context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *acc
	return &cp, nil
}

func (r *LoyaltyRepo) CreateAccount(_ context.Context, account *domain.LoyaltyAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[account.UserID]; ok {
		return domain.ErrAlreadyExists
	}

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now
	cp := *account
	r.accounts[account.UserID] = &cp
	return nil
}

func (r *LoyaltyRepo) AddPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.Points += amount
	acc.TotalEarned += amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) SpendPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	if acc.Points < amount {
		return domain.ErrInsufficientPoints
	}
	acc.Points -= amount
	acc.TotalSpent += amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) RefundPoints(_ context.Context, userID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	if acc.TotalSpent < amount {
		return fmt.Errorf("%w: refund amount exceeds total spent", domain.ErrInvalidInput)
	}
	acc.Points += amount
	acc.TotalSpent -= amount
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) IncrementVisitCount(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.VisitCount++
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) UpdateLevel(_ context.Context, userID uuid.UUID, level domain.LoyaltyLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, ok := r.accounts[userID]
	if !ok {
		return domain.ErrNotFound
	}
	acc.Level = level
	acc.UpdatedAt = time.Now()
	return nil
}

func (r *LoyaltyRepo) ListTransactions(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.LoyaltyTransaction
	for i := len(r.transactions) - 1; i >= 0; i-- {
		if r.transactions[i].UserID == userID {
			items = append(items, r.transactions[i])
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *LoyaltyRepo) CreateTransaction(_ context.Context, tx *domain.LoyaltyTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.CreatedAt = time.Now()
	r.transactions = append(r.transactions, *tx)
	return nil
}


