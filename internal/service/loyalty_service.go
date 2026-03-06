package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type LoyaltyService interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error)
	EarnPoints(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, totalPrice int64) (int64, error)
	SpendPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	RefundPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	GetDiscount(ctx context.Context, userID uuid.UUID) (int, error)
	RecalculateLevel(ctx context.Context, userID uuid.UUID) error
	ListTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error)
}

type loyaltyService struct {
	loyaltyRepo repository.LoyaltyRepository
	logger      *logger.Logger
}

func NewLoyaltyService(
	loyaltyRepo repository.LoyaltyRepository,
	log *logger.Logger,
) LoyaltyService {
	return &loyaltyService{
		loyaltyRepo: loyaltyRepo,
		logger:      log,
	}
}

// GetAccount returns the loyalty account for a user, creating one if it doesn't exist.
func (s *loyaltyService) GetAccount(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error) {
	account, err := s.loyaltyRepo.GetAccount(ctx, userID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		// Auto-create account at Bronze level
		account = &domain.LoyaltyAccount{
			UserID:      userID,
			Level:       domain.LoyaltyBronze,
			Points:      0,
			TotalEarned: 0,
			TotalSpent:  0,
			VisitCount:  0,
		}
		if err := s.loyaltyRepo.CreateAccount(ctx, account); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				// Another concurrent request created the account first, fetch it
				return s.loyaltyRepo.GetAccount(ctx, userID)
			}
			return nil, err
		}
		s.logger.Info("created loyalty account", "user_id", userID)
	}
	return account, nil
}

// EarnPoints awards points to a user for a completed booking.
// Formula: (totalPrice / 100) * multiplier, where multiplier depends on loyalty level.
// Returns the number of points awarded.
func (s *loyaltyService) EarnPoints(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, totalPrice int64) (int64, error) {
	if totalPrice < 0 {
		return 0, fmt.Errorf("%w: total price must not be negative", domain.ErrInvalidInput)
	}

	account, err := s.GetAccount(ctx, userID)
	if err != nil {
		return 0, err
	}

	levelInfo := domain.GetLoyaltyLevelInfo(account.Level)
	points := int64(math.Round(float64(totalPrice) / 100.0 * levelInfo.PointMultiplier))

	// Always increment visit count for completed bookings, even if no points earned
	if err := s.loyaltyRepo.IncrementVisitCount(ctx, userID); err != nil {
		return 0, err
	}

	if points <= 0 {
		return 0, nil
	}

	if err := s.loyaltyRepo.AddPoints(ctx, userID, points); err != nil {
		return 0, err
	}

	tx := &domain.LoyaltyTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        domain.LoyaltyTransactionEarn,
		Amount:      points,
		BookingID:   &bookingID,
		Description: fmt.Sprintf("Начисление за бронирование (%d коп.)", totalPrice),
		CreatedAt:   time.Now(),
	}
	if err := s.loyaltyRepo.CreateTransaction(ctx, tx); err != nil {
		return 0, err
	}

	s.logger.Info("earned loyalty points", "user_id", userID, "points", points, "booking_id", bookingID)
	return points, nil
}

// SpendPoints deducts points from a user's balance for a booking payment.
func (s *loyaltyService) SpendPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}

	account, err := s.GetAccount(ctx, userID)
	if err != nil {
		return err
	}

	if account.Points < amount {
		return domain.ErrInsufficientPoints
	}

	if err := s.loyaltyRepo.SpendPoints(ctx, userID, amount); err != nil {
		return err
	}

	tx := &domain.LoyaltyTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        domain.LoyaltyTransactionSpend,
		Amount:      amount,
		BookingID:   &bookingID,
		Description: "Списание при бронировании",
		CreatedAt:   time.Now(),
	}
	if err := s.loyaltyRepo.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	s.logger.Info("spent loyalty points", "user_id", userID, "points", amount, "booking_id", bookingID)
	return nil
}

// RefundPoints restores previously spent points to the user's balance.
func (s *loyaltyService) RefundPoints(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}

	if err := s.loyaltyRepo.RefundPoints(ctx, userID, amount); err != nil {
		return err
	}

	tx := &domain.LoyaltyTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        domain.LoyaltyTransactionRefund,
		Amount:      amount,
		BookingID:   &bookingID,
		Description: "Возврат баллов за отменённое бронирование",
		CreatedAt:   time.Now(),
	}
	if err := s.loyaltyRepo.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	s.logger.Info("refunded loyalty points", "user_id", userID, "points", amount, "booking_id", bookingID)
	return nil
}

// GetDiscount returns the current discount percentage based on the user's loyalty level.
func (s *loyaltyService) GetDiscount(ctx context.Context, userID uuid.UUID) (int, error) {
	account, err := s.GetAccount(ctx, userID)
	if err != nil {
		return 0, err
	}

	levelInfo := domain.GetLoyaltyLevelInfo(account.Level)
	return levelInfo.DiscountPercent, nil
}

// RecalculateLevel updates the user's loyalty level based on their visit count.
func (s *loyaltyService) RecalculateLevel(ctx context.Context, userID uuid.UUID) error {
	account, err := s.GetAccount(ctx, userID)
	if err != nil {
		return err
	}

	newLevel := domain.LevelForVisitCount(account.VisitCount)
	if newLevel != account.Level {
		if err := s.loyaltyRepo.UpdateLevel(ctx, userID, newLevel); err != nil {
			return err
		}
		s.logger.Info("loyalty level changed", "user_id", userID, "old_level", account.Level, "new_level", newLevel)
	}

	return nil
}

// ListTransactions returns paginated transaction history for a user.
func (s *loyaltyService) ListTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
	return s.loyaltyRepo.ListTransactions(ctx, userID, page, pageSize)
}
