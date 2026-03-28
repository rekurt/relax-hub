package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// RegionService handles region switching for users.
type RegionService interface {
	SwitchRegion(ctx context.Context, userID uuid.UUID, newRegion domain.UserRegion) error
	GetRegion(ctx context.Context, userID uuid.UUID) (domain.UserRegion, error)
}

type regionService struct {
	userRepo    repository.UserRepository
	walletRepo  repository.WalletRepository
	bookingRepo repository.BookingRepository
	disputeRepo repository.DisputeRepository
	loyaltyRepo repository.LoyaltyRepository
	log         *logger.Logger
}

func NewRegionService(
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	bookingRepo repository.BookingRepository,
	disputeRepo repository.DisputeRepository,
	loyaltyRepo repository.LoyaltyRepository,
	log *logger.Logger,
) RegionService {
	return &regionService{
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		bookingRepo: bookingRepo,
		disputeRepo: disputeRepo,
		loyaltyRepo: loyaltyRepo,
		log:         log,
	}
}

func (s *regionService) GetRegion(ctx context.Context, userID uuid.UUID) (domain.UserRegion, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user.Region == "" {
		return domain.RegionRU, nil
	}
	return user.Region, nil
}

func (s *regionService) SwitchRegion(ctx context.Context, userID uuid.UUID, newRegion domain.UserRegion) error {
	if !newRegion.IsValid() {
		return domain.ErrRegionInvalid
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	currentRegion := user.Region
	if currentRegion == "" {
		currentRegion = domain.RegionRU
	}
	if currentRegion == newRegion {
		return domain.ErrRegionSameAsCurrent
	}

	// Check blocking conditions
	if err := s.validateRegionSwitch(ctx, userID); err != nil {
		return err
	}

	// Archive old wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil && err != domain.ErrNotFound && err != domain.ErrWalletNotFound {
		return fmt.Errorf("get wallet: %w", err)
	}
	if wallet != nil {
		if err := s.walletRepo.UpdateStatus(ctx, wallet.ID, domain.WalletStatusArchived); err != nil {
			return fmt.Errorf("archive wallet: %w", err)
		}
		s.log.Info("archived wallet on region switch",
			"user_id", userID, "wallet_id", wallet.ID, "old_region", currentRegion, "new_region", newRegion)
	}

	// Create new wallet in new currency
	newCurrency := domain.CurrencyForRegion(newRegion)
	newWallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Currency: newCurrency,
		Status:   domain.WalletStatusActive,
	}
	if err := s.walletRepo.Create(ctx, newWallet); err != nil {
		return fmt.Errorf("create new wallet: %w", err)
	}

	// Reset loyalty status
	loyaltyAccount, err := s.loyaltyRepo.GetAccount(ctx, userID)
	if err == nil && loyaltyAccount != nil {
		if err := s.loyaltyRepo.UpdateLevel(ctx, userID, domain.LoyaltyBronze); err != nil {
			s.log.Warn("failed to reset loyalty level on region switch", "user_id", userID, "error", err)
		}
	}

	// Update user region
	user.Region = newRegion
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user region: %w", err)
	}

	s.log.Info("region switched successfully",
		"user_id", userID, "from", currentRegion, "to", newRegion)

	return nil
}

func (s *regionService) validateRegionSwitch(ctx context.Context, userID uuid.UUID) error {
	// 1. Check wallet balance
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil && err != domain.ErrNotFound && err != domain.ErrWalletNotFound {
		return fmt.Errorf("check wallet: %w", err)
	}
	if wallet != nil && (wallet.Balance > 0 || wallet.HeldAmount > 0) {
		return domain.ErrRegionSwitchBlocked
	}

	// 2. Check active bookings
	activeBookings, err := s.bookingRepo.CountActiveByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("check active bookings: %w", err)
	}
	if activeBookings > 0 {
		return domain.ErrRegionSwitchBlocked
	}

	// 3. Check open disputes
	openDisputes, err := s.disputeRepo.CountOpenByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("check open disputes: %w", err)
	}
	if openDisputes > 0 {
		return domain.ErrRegionSwitchBlocked
	}

	return nil
}
