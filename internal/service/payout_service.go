package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/repository"
)

type PayoutService interface {
	RequestPayout(ctx context.Context, userID uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error)
	GetPayoutHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error)
	SetAutoPayoutThreshold(ctx context.Context, userID uuid.UUID, threshold int64) error
	GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error)
	CalculateAvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error)
	ProcessPayout(ctx context.Context, payoutID uuid.UUID) error
}

type payoutService struct {
	payoutRepo     repository.PayoutRepository
	walletRepo     repository.WalletRepository
	userRepo       repository.UserRepository
	payoutProvider payment.PayoutProvider // nil = manual processing only
	logger         *logger.Logger
}

func NewPayoutService(
	payoutRepo repository.PayoutRepository,
	walletRepo repository.WalletRepository,
	userRepo repository.UserRepository,
	payoutProvider payment.PayoutProvider,
	log *logger.Logger,
) PayoutService {
	return &payoutService{
		payoutRepo:     payoutRepo,
		walletRepo:     walletRepo,
		userRepo:       userRepo,
		payoutProvider: payoutProvider,
		logger:         log,
	}
}

// DeterminePayoutMethod selects the payout method based on user region and phone availability.
// RU users with a verified phone get SBP (instant), everyone else gets bank transfer.
func DeterminePayoutMethod(user *domain.User) domain.PayoutMethod {
	if user.Region == domain.RegionRU && user.Phone != "" && user.PhoneVerified {
		return domain.PayoutMethodSBP
	}
	return domain.PayoutMethodBankTransfer
}

func (s *payoutService) RequestPayout(ctx context.Context, userID uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
	// Determine currency-appropriate limits from user's wallet
	minAmount := domain.PayoutMinAmountRUB
	dailyLimit := domain.PayoutDailyLimitRUB
	monthlyLimit := domain.PayoutMonthlyLimitRUB
	if wallet, wErr := s.walletRepo.GetByUserID(ctx, userID); wErr == nil && wallet.Currency == domain.WalletCurrencyBYN {
		minAmount = domain.PayoutMinAmountBYN
		dailyLimit = domain.PayoutDailyLimitBYN
		monthlyLimit = domain.PayoutMonthlyLimitBYN
	}

	if amount < minAmount {
		return nil, domain.ErrPayoutBelowMinimum
	}

	// Check available balance (wallet balance - held - pending payouts)
	available, err := s.CalculateAvailableBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if available < amount {
		return nil, domain.ErrInsufficientWalletBalance
	}

	// Check daily limit
	now := time.Now()
	dailyTotal, err := s.payoutRepo.GetDailyTotal(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if dailyTotal+amount > dailyLimit {
		return nil, domain.ErrPayoutDailyLimitExceeded
	}

	// Check monthly limit
	monthlyTotal, err := s.payoutRepo.GetMonthlyTotal(ctx, userID, now.Year(), now.Month())
	if err != nil {
		return nil, err
	}
	if monthlyTotal+amount > monthlyLimit {
		return nil, domain.ErrPayoutMonthlyLimitExceeded
	}

	// Determine payout method based on user profile
	payoutMethod := domain.PayoutMethodBankTransfer
	if s.userRepo != nil {
		user, userErr := s.userRepo.GetByID(ctx, userID)
		if userErr == nil {
			payoutMethod = DeterminePayoutMethod(user)
		}
	}

	payout := &domain.Payout{
		ID:           uuid.New(),
		UserID:       userID,
		Amount:       amount,
		Status:       domain.PayoutStatusPending,
		PayoutMethod: payoutMethod,
		BankDetails:  bankDetails,
		RequestedAt:  now,
		CreatedAt:    now,
	}

	if err := s.payoutRepo.Create(ctx, payout); err != nil {
		return nil, err
	}

	s.logger.Info("payout requested", "user_id", userID, "amount", amount, "payout_id", payout.ID, "method", payoutMethod)
	return payout, nil
}

func (s *payoutService) GetPayoutHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
	return s.payoutRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *payoutService) SetAutoPayoutThreshold(ctx context.Context, userID uuid.UUID, threshold int64) error {
	if threshold < 0 {
		return domain.ErrInvalidInput
	}
	// threshold of 0 means disabled, any positive value is a minimum threshold
	minAmount := domain.PayoutMinAmountRUB
	if wallet, wErr := s.walletRepo.GetByUserID(ctx, userID); wErr == nil && wallet.Currency == domain.WalletCurrencyBYN {
		minAmount = domain.PayoutMinAmountBYN
	}
	if threshold > 0 && threshold < minAmount {
		return domain.ErrPayoutBelowMinimum
	}

	settings := &domain.AutoPayoutSettings{
		UserID:    userID,
		Threshold: threshold,
	}
	return s.payoutRepo.UpsertAutoPayoutSettings(ctx, settings)
}

func (s *payoutService) GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error) {
	return s.payoutRepo.GetAutoPayoutSettings(ctx, userID)
}

func (s *payoutService) CalculateAvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	pendingPayouts, err := s.payoutRepo.GetPendingTotal(ctx, userID)
	if err != nil {
		return 0, err
	}

	available := wallet.AvailableBalance() - pendingPayouts
	if available < 0 {
		available = 0
	}
	return available, nil
}

func (s *payoutService) ProcessPayout(ctx context.Context, payoutID uuid.UUID) error {
	payout, err := s.payoutRepo.GetByID(ctx, payoutID)
	if err != nil {
		return err
	}

	// Allow retry from "processing" state (e.g., after SBP provider failure)
	isRetry := payout.Status == domain.PayoutStatusProcessing
	if payout.Status != domain.PayoutStatusPending && !isRetry {
		return domain.ErrPayoutAlreadyProcessed
	}

	// Mark as processing (skip if already processing on retry)
	if !isRetry {
		if err := s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusProcessing, nil, ""); err != nil {
			return err
		}
	}

	// Deduct from wallet (skip on retry — wallet was already deducted)
	if !isRetry {
		wallet, err := s.walletRepo.GetByUserID(ctx, payout.UserID)
		if err != nil {
			now := time.Now()
			_ = s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusFailed, &now, "wallet not found")
			return err
		}

		if wallet.AvailableBalance() < payout.Amount {
			now := time.Now()
			_ = s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusFailed, &now, "insufficient balance")
			return domain.ErrInsufficientWalletBalance
		}

		newBalance := wallet.Balance - payout.Amount
		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
			now := time.Now()
			_ = s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusFailed, &now, "balance update failed")
			return err
		}

		// Record payout transaction in wallet
		tx := &domain.WalletTransaction{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			Type:          domain.WalletTxPayout,
			Amount:        payout.Amount,
			BalanceAfter:  newBalance,
			Status:        domain.WalletTxStatusCompleted,
			Description:   "Вывод средств",
			ReferenceType: "payout",
			ReferenceID:   &payout.ID,
		}
		if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
			// Balance was already deducted but transaction record failed - rollback balance
			s.logger.Error("failed to create payout transaction, rolling back balance",
				"error", err, "payout_id", payoutID, "wallet_id", wallet.ID, "amount", payout.Amount)
			if rbErr := s.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance, wallet.Balance, wallet.HeldAmount, wallet.HeldAmount); rbErr != nil {
				s.logger.Error("CRITICAL: balance rollback failed after payout transaction failure, requires manual reconciliation",
					"error", rbErr, "payout_id", payoutID, "wallet_id", wallet.ID, "amount", payout.Amount,
					"deducted_balance", newBalance, "original_balance", wallet.Balance)
			}
			failedAt := time.Now()
			_ = s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusFailed, &failedAt, "transaction record failed")
			return fmt.Errorf("create payout transaction: %w", err)
		}
	}

	// Execute payout via payment provider if available
	if s.payoutProvider != nil && payout.PayoutMethod == domain.PayoutMethodSBP {
		user, userErr := s.userRepo.GetByID(ctx, payout.UserID)
		if userErr != nil {
			s.logger.Error("SBP payout failed: cannot load user",
				"error", userErr, "payout_id", payoutID, "user_id", payout.UserID)
			s.failPayoutAndRefundWallet(ctx, payout, "user not found")
			return fmt.Errorf("SBP payout: load user: %w", userErr)
		}
		if user.Phone == "" {
			s.logger.Error("SBP payout failed: user has no phone",
				"payout_id", payoutID, "user_id", payout.UserID)
			s.failPayoutAndRefundWallet(ctx, payout, "user has no phone")
			return fmt.Errorf("SBP payout: user has no phone number")
		}
		currency := string(domain.CurrencyForRegion(user.Region))
		result, providerErr := s.payoutProvider.CreatePayout(ctx, payment.CreatePayoutRequest{
			Amount:      payout.Amount,
			Currency:    currency,
			Method:      "sbp",
			Phone:       user.Phone,
			Description: fmt.Sprintf("Вывод средств #%s", payout.ID.String()[:8]),
		})
		if providerErr != nil {
			s.logger.Error("SBP payout provider failed",
				"error", providerErr, "payout_id", payoutID)
			s.failPayoutAndRefundWallet(ctx, payout, providerErr.Error())
			return fmt.Errorf("SBP payout provider failed: %w", providerErr)
		}
		_ = s.payoutRepo.UpdateExternalID(ctx, payoutID, result.ExternalID)
	}

	now := time.Now()
	if err := s.payoutRepo.UpdateStatus(ctx, payoutID, domain.PayoutStatusCompleted, &now, ""); err != nil {
		return err
	}

	s.logger.Info("payout processed", "payout_id", payoutID, "amount", payout.Amount, "method", payout.PayoutMethod)
	return nil
}

// failPayoutAndRefundWallet marks a payout as failed and refunds the wallet.
// Wallet was deducted in the first ProcessPayout call (before SBP provider), so refund is always needed.
func (s *payoutService) failPayoutAndRefundWallet(ctx context.Context, payout *domain.Payout, reason string) {
	failedAt := time.Now()
	_ = s.payoutRepo.UpdateStatus(ctx, payout.ID, domain.PayoutStatusFailed, &failedAt, reason)

	wallet, err := s.walletRepo.GetByUserID(ctx, payout.UserID)
	if err != nil {
		s.logger.Error("CRITICAL: cannot refund wallet for failed payout, manual reconciliation required",
			"error", err, "payout_id", payout.ID, "user_id", payout.UserID, "amount", payout.Amount)
		return
	}
	newBalance := wallet.Balance + payout.Amount
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Balance, newBalance, wallet.HeldAmount, wallet.HeldAmount); err != nil {
		s.logger.Error("CRITICAL: wallet refund failed for failed payout, manual reconciliation required",
			"error", err, "payout_id", payout.ID, "wallet_id", wallet.ID, "amount", payout.Amount)
		return
	}

	// Record refund transaction for audit trail
	tx := &domain.WalletTransaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		Type:          domain.WalletTxRefund,
		Amount:        payout.Amount,
		BalanceAfter:  newBalance,
		Status:        domain.WalletTxStatusCompleted,
		Description:   "Возврат средств: неудачный вывод",
		ReferenceType: "payout",
		ReferenceID:   &payout.ID,
	}
	if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("failed to create refund transaction for failed payout, balance updated but no record",
			"error", err, "payout_id", payout.ID, "wallet_id", wallet.ID, "amount", payout.Amount)
	}

	s.logger.Info("wallet refunded for permanently failed SBP payout",
		"payout_id", payout.ID, "wallet_id", wallet.ID, "amount", payout.Amount)
}
