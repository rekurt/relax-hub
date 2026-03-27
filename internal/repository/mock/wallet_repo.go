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

type WalletRepo struct {
	mu           sync.RWMutex
	wallets      map[uuid.UUID]*domain.Wallet
	transactions map[uuid.UUID]*domain.WalletTransaction
	holds        map[uuid.UUID]*domain.WalletHold
}

func NewWalletRepo() repository.WalletRepository {
	return &WalletRepo{
		wallets:      make(map[uuid.UUID]*domain.Wallet),
		transactions: make(map[uuid.UUID]*domain.WalletTransaction),
		holds:        make(map[uuid.UUID]*domain.WalletHold),
	}
}

func (r *WalletRepo) Create(_ context.Context, wallet *domain.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if wallet.ID == uuid.Nil {
		wallet.ID = uuid.New()
	}
	now := time.Now()
	if wallet.CreatedAt.IsZero() {
		wallet.CreatedAt = now
	}
	if wallet.UpdatedAt.IsZero() {
		wallet.UpdatedAt = now
	}
	if wallet.Status == "" {
		wallet.Status = domain.WalletStatusActive
	}
	if wallet.Currency == "" {
		wallet.Currency = domain.WalletCurrencyRUB
	}

	for _, w := range r.wallets {
		if w.UserID == wallet.UserID {
			return domain.ErrAlreadyExists
		}
	}

	cp := *wallet
	r.wallets[wallet.ID] = &cp
	return nil
}

func (r *WalletRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	w, ok := r.wallets[id]
	if !ok {
		return nil, domain.ErrWalletNotFound
	}
	cp := *w
	return &cp, nil
}

func (r *WalletRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, w := range r.wallets {
		if w.UserID == userID {
			cp := *w
			return &cp, nil
		}
	}
	return nil, domain.ErrWalletNotFound
}

func (r *WalletRepo) UpdateBalance(_ context.Context, walletID uuid.UUID, oldBalance, newBalance int64, oldHeldAmount, newHeldAmount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	w, ok := r.wallets[walletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	if w.Balance != oldBalance || w.HeldAmount != oldHeldAmount {
		return domain.ErrWalletConcurrentUpdate
	}
	w.Balance = newBalance
	w.HeldAmount = newHeldAmount
	w.UpdatedAt = time.Now()
	return nil
}

func (r *WalletRepo) UpdateStatus(_ context.Context, walletID uuid.UUID, status domain.WalletStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	w, ok := r.wallets[walletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	w.Status = status
	w.UpdatedAt = time.Now()
	return nil
}

func (r *WalletRepo) ListAllIDs(_ context.Context) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(r.wallets))
	for id, w := range r.wallets {
		if w.Status == domain.WalletStatusActive {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// --- Transactions ---

func (r *WalletRepo) CreateTransaction(_ context.Context, tx *domain.WalletTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	if tx.Status == "" {
		tx.Status = domain.WalletTxStatusCompleted
	}

	cp := r.copyTransaction(tx)
	r.transactions[tx.ID] = &cp
	return nil
}

func (r *WalletRepo) ListTransactions(_ context.Context, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.WalletTransaction
	for _, tx := range r.transactions {
		if filter.WalletID != nil && tx.WalletID != *filter.WalletID {
			continue
		}
		if filter.Type != nil && tx.Type != *filter.Type {
			continue
		}
		if filter.IsBonus != nil && tx.IsBonus != *filter.IsBonus {
			continue
		}
		if filter.DateFrom != nil && tx.CreatedAt.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && tx.CreatedAt.After(*filter.DateTo) {
			continue
		}
		filtered = append(filtered, r.copyTransaction(tx))
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *WalletRepo) GetExpiringBonuses(_ context.Context, walletID uuid.UUID, before time.Time) ([]domain.WalletTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.WalletTransaction
	for _, tx := range r.transactions {
		if tx.WalletID == walletID && tx.IsBonus && tx.ExpiresAt != nil &&
			!tx.ExpiresAt.After(before) && tx.Status == domain.WalletTxStatusCompleted {
			result = append(result, r.copyTransaction(tx))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpiresAt.Before(*result[j].ExpiresAt)
	})
	return result, nil
}

func (r *WalletRepo) GetBonusTransactionsForSpending(_ context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var result []domain.WalletTransaction
	for _, tx := range r.transactions {
		if tx.WalletID == walletID && tx.IsBonus && tx.Status == domain.WalletTxStatusCompleted {
			if tx.ExpiresAt != nil && !tx.ExpiresAt.After(now) {
				continue
			}
			result = append(result, r.copyTransaction(tx))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].ExpiresAt == nil && result[j].ExpiresAt == nil {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		}
		if result[i].ExpiresAt == nil {
			return false
		}
		if result[j].ExpiresAt == nil {
			return true
		}
		return result[i].ExpiresAt.Before(*result[j].ExpiresAt)
	})
	return result, nil
}

func (r *WalletRepo) ExpireBonuses(_ context.Context, transactionIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idSet := make(map[uuid.UUID]bool, len(transactionIDs))
	for _, id := range transactionIDs {
		idSet[id] = true
	}

	for _, tx := range r.transactions {
		if idSet[tx.ID] && tx.Status == domain.WalletTxStatusCompleted {
			tx.Status = domain.WalletTxStatusCancelled
		}
	}
	return nil
}

func (r *WalletRepo) GetExpiringBonusesSoon(_ context.Context, walletID uuid.UUID, from, to time.Time) ([]domain.WalletTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.WalletTransaction
	for _, tx := range r.transactions {
		if tx.WalletID == walletID && tx.IsBonus && tx.ExpiresAt != nil &&
			!tx.ExpiresAt.Before(from) && !tx.ExpiresAt.After(to) &&
			tx.Status == domain.WalletTxStatusCompleted {
			result = append(result, r.copyTransaction(tx))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpiresAt.Before(*result[j].ExpiresAt)
	})
	return result, nil
}

func (r *WalletRepo) copyTransaction(tx *domain.WalletTransaction) domain.WalletTransaction {
	cp := *tx
	if tx.ReferenceID != nil {
		id := *tx.ReferenceID
		cp.ReferenceID = &id
	}
	if tx.ExpiresAt != nil {
		t := *tx.ExpiresAt
		cp.ExpiresAt = &t
	}
	return cp
}

// --- Holds ---

func (r *WalletRepo) CreateHold(_ context.Context, hold *domain.WalletHold) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hold.ID == uuid.Nil {
		hold.ID = uuid.New()
	}
	if hold.CreatedAt.IsZero() {
		hold.CreatedAt = time.Now()
	}
	if hold.Status == "" {
		hold.Status = domain.WalletHoldStatusActive
	}

	cp := r.copyHold(hold)
	r.holds[hold.ID] = &cp
	return nil
}

func (r *WalletRepo) GetHoldByID(_ context.Context, holdID uuid.UUID) (*domain.WalletHold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.holds[holdID]
	if !ok {
		return nil, domain.ErrHoldNotFound
	}
	cp := r.copyHold(h)
	return &cp, nil
}

func (r *WalletRepo) UpdateHoldStatus(_ context.Context, holdID uuid.UUID, status domain.WalletHoldStatus, capturedAt, releasedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	h, ok := r.holds[holdID]
	if !ok {
		return domain.ErrHoldNotFound
	}
	h.Status = status
	h.CapturedAt = nil
	if capturedAt != nil {
		t := *capturedAt
		h.CapturedAt = &t
	}
	h.ReleasedAt = nil
	if releasedAt != nil {
		t := *releasedAt
		h.ReleasedAt = &t
	}
	return nil
}

func (r *WalletRepo) GetActiveHolds(_ context.Context, walletID uuid.UUID) ([]domain.WalletHold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.WalletHold
	for _, h := range r.holds {
		if h.WalletID == walletID && h.Status == domain.WalletHoldStatusActive {
			result = append(result, r.copyHold(h))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

func (r *WalletRepo) GetExpiredHolds(_ context.Context, before time.Time) ([]domain.WalletHold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.WalletHold
	for _, h := range r.holds {
		if h.Status == domain.WalletHoldStatusActive && !h.ExpiresAt.After(before) {
			result = append(result, r.copyHold(h))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpiresAt.Before(result[j].ExpiresAt)
	})
	return result, nil
}

func (r *WalletRepo) copyHold(h *domain.WalletHold) domain.WalletHold {
	cp := *h
	if h.ReferenceID != nil {
		id := *h.ReferenceID
		cp.ReferenceID = &id
	}
	if h.CapturedAt != nil {
		t := *h.CapturedAt
		cp.CapturedAt = &t
	}
	if h.ReleasedAt != nil {
		t := *h.ReleasedAt
		cp.ReleasedAt = &t
	}
	return cp
}
