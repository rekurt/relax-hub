package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type WalletBalanceSummary struct {
	Balance        int64              `json:"balance"`
	HeldAmount     int64              `json:"held_amount"`
	Available      int64              `json:"available"`
	Currency       domain.WalletCurrency `json:"currency"`
	ExpiringSoon   int64              `json:"expiring_soon"`
	EarliestExpiry *time.Time         `json:"earliest_expiry,omitempty"`
}

type WalletService interface {
	CreateWallet(ctx context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error)
	GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	TopUp(ctx context.Context, userID uuid.UUID, amount int64) (*domain.WalletTransaction, error)
	Spend(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error)
	Hold(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string, expiresAt time.Time) (*domain.WalletHold, error)
	CaptureHold(ctx context.Context, holdID uuid.UUID) (*domain.WalletTransaction, error)
	ReleaseHold(ctx context.Context, holdID uuid.UUID) error
	Refund(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error)
	AddBonus(ctx context.Context, walletID uuid.UUID, amount int64, bonusType domain.WalletTransactionType, expiresAt *time.Time, description string) (*domain.WalletTransaction, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (*WalletBalanceSummary, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error)
	GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error)
	ExpireBonuses(ctx context.Context) (int, error)
	ExpireBonusesForWallet(ctx context.Context, walletID uuid.UUID) (int, error)
	FreezeAndZeroBalance(ctx context.Context, userID uuid.UUID) error
}

type walletService struct {
	walletRepo repository.WalletRepository
	logger     *logger.Logger
}

func NewWalletService(
	walletRepo repository.WalletRepository,
	log *logger.Logger,
) WalletService {
	return &walletService{
		walletRepo: walletRepo,
		logger:     log,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
	if !currency.IsValid() {
		currency = domain.WalletCurrencyRUB
	}

	wallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  0,
		HeldAmount: 0,
		Currency: currency,
		Status:   domain.WalletStatusActive,
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *walletService) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

func (s *walletService) TopUp(ctx context.Context, userID uuid.UUID, amount int64) (*domain.WalletTransaction, error) {
	if amount < domain.WalletTopUpMinRUB {
		return nil, domain.ErrTopUpBelowMinimum
	}
	if amount > domain.WalletTopUpMaxRUB {
		return nil, domain.ErrTopUpAboveMaximum
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if wallet.IsFrozen() {
		return nil, domain.ErrWalletFrozen
	}

	newBalance := wallet.Balance + amount
	maxBalance := domain.MaxBalanceForCurrency(wallet.Currency)
	if newBalance > maxBalance {
		return nil, domain.ErrWalletLimitExceeded
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		return nil, err
	}

	tx := &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     wallet.ID,
		Type:         domain.WalletTxTopUp,
		Amount:       amount,
		BalanceAfter: newBalance,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Пополнение кошелька",
	}

	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *walletService) Spend(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidInput
	}

	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.IsFrozen() {
		return nil, domain.ErrWalletFrozen
	}

	if wallet.AvailableBalance() < amount {
		return nil, domain.ErrInsufficientWalletBalance
	}

	newBalance := wallet.Balance - amount
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		return nil, err
	}

	tx := &domain.WalletTransaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		Type:          domain.WalletTxSpend,
		Amount:        amount,
		BalanceAfter:  newBalance,
		Status:        domain.WalletTxStatusCompleted,
		Description:   description,
		ReferenceType: refType,
		ReferenceID:   refID,
	}

	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *walletService) Hold(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string, expiresAt time.Time) (*domain.WalletHold, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidInput
	}

	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.IsFrozen() {
		return nil, domain.ErrWalletFrozen
	}

	if wallet.AvailableBalance() < amount {
		return nil, domain.ErrInsufficientWalletBalance
	}

	newHeldAmount := wallet.HeldAmount + amount
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, wallet.Balance, wallet.HeldAmount, newHeldAmount); err != nil {
		return nil, err
	}

	hold := &domain.WalletHold{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		Amount:        amount,
		Status:        domain.WalletHoldStatusActive,
		Description:   description,
		ReferenceType: refType,
		ReferenceID:   refID,
		ExpiresAt:     expiresAt,
	}

	if err := s.walletRepo.CreateHold(ctx, hold); err != nil {
		return nil, err
	}

	return hold, nil
}

func (s *walletService) CaptureHold(ctx context.Context, holdID uuid.UUID) (*domain.WalletTransaction, error) {
	hold, err := s.walletRepo.GetHoldByID(ctx, holdID)
	if err != nil {
		return nil, err
	}

	if hold.Status != domain.WalletHoldStatusActive {
		return nil, domain.ErrHoldNotFound
	}

	if hold.IsExpired() {
		return nil, domain.ErrHoldExpired
	}

	wallet, err := s.walletRepo.GetByID(ctx, hold.WalletID)
	if err != nil {
		return nil, err
	}

	newBalance := wallet.Balance - hold.Amount
	if newBalance < 0 {
		return nil, domain.ErrInsufficientWalletBalance
	}

	newHeldAmount := wallet.HeldAmount - hold.Amount
	if newHeldAmount < 0 {
		newHeldAmount = 0
	}

	// Update balance first, then mark hold as captured.
	// If balance update succeeds but hold status update fails, the balance is deducted
	// but the hold remains active -- a subsequent retry can still capture it safely.
	// The reverse order (hold first, then balance) would leak money on partial failure.
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, newHeldAmount); err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.walletRepo.UpdateHoldStatus(ctx, holdID, domain.WalletHoldStatusCaptured, &now, nil); err != nil {
		s.logger.Error("hold captured but status update failed", "hold_id", holdID, "error", err)
	}

	tx := &domain.WalletTransaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		Type:          domain.WalletTxHoldCapture,
		Amount:        hold.Amount,
		BalanceAfter:  newBalance,
		Status:        domain.WalletTxStatusCompleted,
		Description:   hold.Description,
		ReferenceType: hold.ReferenceType,
		ReferenceID:   hold.ReferenceID,
	}

	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *walletService) ReleaseHold(ctx context.Context, holdID uuid.UUID) error {
	hold, err := s.walletRepo.GetHoldByID(ctx, holdID)
	if err != nil {
		return err
	}

	if hold.Status != domain.WalletHoldStatusActive {
		return domain.ErrHoldNotFound
	}

	wallet, err := s.walletRepo.GetByID(ctx, hold.WalletID)
	if err != nil {
		return err
	}

	newHeldAmount := wallet.HeldAmount - hold.Amount
	if newHeldAmount < 0 {
		newHeldAmount = 0
	}

	// Update wallet balance first, then mark hold as released.
	// If balance update succeeds but hold status update fails, the held amount is reduced
	// but hold remains active -- a subsequent retry can still release it safely.
	// The reverse order (hold first, then balance) would permanently freeze funds on partial failure.
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, wallet.Balance, wallet.HeldAmount, newHeldAmount); err != nil {
		return err
	}

	now := time.Now()
	if err := s.walletRepo.UpdateHoldStatus(ctx, holdID, domain.WalletHoldStatusReleased, nil, &now); err != nil {
		s.logger.Error("hold released but status update failed", "hold_id", holdID, "error", err)
	}

	return nil
}

func (s *walletService) Refund(ctx context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidInput
	}

	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	newBalance := wallet.Balance + amount
	maxBalance := domain.MaxBalanceForCurrency(wallet.Currency)
	if newBalance > maxBalance {
		newBalance = maxBalance
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		return nil, err
	}

	tx := &domain.WalletTransaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		Type:          domain.WalletTxRefund,
		Amount:        amount,
		BalanceAfter:  newBalance,
		Status:        domain.WalletTxStatusCompleted,
		Description:   description,
		ReferenceType: refType,
		ReferenceID:   refID,
	}

	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *walletService) AddBonus(ctx context.Context, walletID uuid.UUID, amount int64, bonusType domain.WalletTransactionType, expiresAt *time.Time, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidInput
	}

	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.IsFrozen() {
		return nil, domain.ErrWalletFrozen
	}

	newBalance := wallet.Balance + amount
	maxBalance := domain.MaxBalanceForCurrency(wallet.Currency)
	if newBalance > maxBalance {
		return nil, domain.ErrWalletLimitExceeded
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		return nil, err
	}

	tx := &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     wallet.ID,
		Type:         bonusType,
		Amount:       amount,
		BalanceAfter: newBalance,
		Status:       domain.WalletTxStatusCompleted,
		Description:  description,
		IsBonus:      true,
		ExpiresAt:    expiresAt,
	}

	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *walletService) GetBalance(ctx context.Context, userID uuid.UUID) (*WalletBalanceSummary, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &WalletBalanceSummary{
		Balance:    wallet.Balance,
		HeldAmount: wallet.HeldAmount,
		Available:  wallet.AvailableBalance(),
		Currency:   wallet.Currency,
	}

	// Check for bonuses expiring within 14 days
	soon := time.Now().Add(14 * 24 * time.Hour)
	expiring, err := s.walletRepo.GetExpiringBonusesSoon(ctx, wallet.ID, time.Now(), soon)
	if err != nil {
		s.logger.Error("failed to get expiring bonuses", "error", err)
		return summary, nil
	}

	for _, tx := range expiring {
		summary.ExpiringSoon += tx.Amount
		if summary.EarliestExpiry == nil || (tx.ExpiresAt != nil && tx.ExpiresAt.Before(*summary.EarliestExpiry)) {
			summary.EarliestExpiry = tx.ExpiresAt
		}
	}

	return summary, nil
}

func (s *walletService) ListTransactions(ctx context.Context, userID uuid.UUID, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	filter.WalletID = &wallet.ID
	return s.walletRepo.ListTransactions(ctx, filter)
}

func (s *walletService) GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error) {
	return s.walletRepo.GetActiveHolds(ctx, walletID)
}

func (s *walletService) ExpireBonuses(ctx context.Context) (int, error) {
	walletIDs, err := s.walletRepo.ListAllIDs(ctx)
	if err != nil {
		return 0, err
	}

	totalExpired := 0
	for _, wid := range walletIDs {
		expired, err := s.ExpireBonusesForWallet(ctx, wid)
		if err != nil {
			s.logger.Error("failed to expire bonuses for wallet", "wallet_id", wid, "error", err)
			continue
		}
		totalExpired += expired
	}

	return totalExpired, nil
}

func (s *walletService) ExpireBonusesForWallet(ctx context.Context, walletID uuid.UUID) (int, error) {
	now := time.Now()

	expiring, err := s.walletRepo.GetExpiringBonuses(ctx, walletID, now)
	if err != nil {
		return 0, err
	}

	if len(expiring) == 0 {
		return 0, nil
	}

	var totalExpired int64
	var ids []uuid.UUID
	for _, tx := range expiring {
		ids = append(ids, tx.ID)
		totalExpired += tx.Amount
	}

	if err := s.walletRepo.ExpireBonuses(ctx, ids); err != nil {
		return 0, err
	}

	// Deduct expired bonus amount from wallet balance
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return 0, err
	}

	newBalance := wallet.Balance - totalExpired
	if newBalance < 0 {
		newBalance = 0
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		return 0, err
	}

	// Record expiration transaction
	expiryTx := &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         domain.WalletTxBonusExpiry,
		Amount:       totalExpired,
		BalanceAfter: newBalance,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Истечение срока бонусов",
	}

	if err := s.walletRepo.CreateTransaction(ctx, expiryTx); err != nil {
		s.logger.Error("failed to create expiry transaction", "error", err)
	}

	return len(expiring), nil
}

func (s *walletService) FreezeAndZeroBalance(ctx context.Context, userID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Release all active holds before zeroing balance to avoid orphaned hold records
	activeHolds, err := s.walletRepo.GetActiveHolds(ctx, wallet.ID)
	if err != nil {
		s.logger.Error("failed to get active holds for freeze", "wallet_id", wallet.ID, "error", err)
	} else {
		now := time.Now()
		for _, hold := range activeHolds {
			if err := s.walletRepo.UpdateHoldStatus(ctx, hold.ID, domain.WalletHoldStatusReleased, nil, &now); err != nil {
				s.logger.Error("failed to release hold during freeze", "hold_id", hold.ID, "error", err)
			}
		}
	}

	// Freeze status first, then zero balance.
	// If freeze succeeds but zeroing fails, the wallet is frozen with remaining balance --
	// safe because no new operations are allowed on a frozen wallet.
	// The reverse order (zero first, then freeze) would leave a zero-balance active wallet,
	// allowing new top-ups/bonuses to be credited to a wallet pending deletion.
	if err := s.walletRepo.UpdateStatus(ctx, wallet.ID, domain.WalletStatusFrozen); err != nil {
		return fmt.Errorf("freeze wallet: %w", err)
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, 0, wallet.HeldAmount, 0); err != nil {
		return fmt.Errorf("zero wallet balance: %w", err)
	}

	return nil
}
