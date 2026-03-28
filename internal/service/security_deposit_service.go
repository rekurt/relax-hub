package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const defaultDepositClaimHours = 48

type SecurityDepositService interface {
	HoldDeposit(ctx context.Context, bookingID uuid.UUID, amount int64, paymentMethod string) error
	ReleaseDeposit(ctx context.Context, bookingID uuid.UUID) error
	ClaimDeposit(ctx context.Context, bookingID uuid.UUID) error
	FreezeDeposit(ctx context.Context, bookingID uuid.UUID) error
	ProcessMaturedDeposits(ctx context.Context) (int, error)
}

type securityDepositService struct {
	bookingRepo repository.BookingRepository
	provider    payment.PaymentProvider
	logger      *logger.Logger
	claimHours  int
	returnURL   string
}

func NewSecurityDepositService(
	bookingRepo repository.BookingRepository,
	provider payment.PaymentProvider,
	log *logger.Logger,
	claimHours int,
	returnURL string,
) SecurityDepositService {
	if claimHours < 24 || claimHours > 168 {
		claimHours = defaultDepositClaimHours
	}
	return &securityDepositService{
		bookingRepo: bookingRepo,
		provider:    provider,
		logger:      log,
		claimHours:  claimHours,
		returnURL:   returnURL,
	}
}

// CalculateDepositAmount computes the deposit based on bathhouse settings and base price.
func CalculateDepositAmount(basePrice int64, depositPercent int) int64 {
	if depositPercent <= 0 || depositPercent > 50 {
		return 0
	}
	return basePrice * int64(depositPercent) / 100
}

// HoldDeposit creates a card authorization hold for the security deposit.
func (s *securityDepositService) HoldDeposit(ctx context.Context, bookingID uuid.UUID, amount int64, paymentMethod string) error {
	if amount <= 0 {
		return nil
	}

	if paymentMethod == "" {
		paymentMethod = "card"
	}

	description := fmt.Sprintf("Залог за бронирование %s", bookingID.String()[:8])
	result, err := s.provider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:      amount,
		Currency:    "RUB",
		Description: description,
		ReturnURL:   s.returnURL,
		Metadata: map[string]string{
			"booking_id": bookingID.String(),
			"type":       "security_deposit",
		},
		Method:  paymentMethod,
		Capture: false, // authorization hold, not charge
	})
	if err != nil {
		s.logger.Error("failed to create deposit hold", "booking_id", bookingID, "amount", amount, "error", err)
		return fmt.Errorf("create deposit hold: %w", err)
	}

	if err := s.bookingRepo.UpdateDeposit(ctx, bookingID, amount, domain.DepositHeld, result.ExternalID); err != nil {
		// Best-effort cancel the hold since we failed to record it
		if cancelErr := s.provider.CancelPayment(ctx, result.ExternalID); cancelErr != nil {
			s.logger.Error("failed to cancel deposit hold after update failure",
				"booking_id", bookingID, "external_id", result.ExternalID, "error", cancelErr)
		}
		return fmt.Errorf("record deposit hold: %w", err)
	}

	s.logger.Info("Security deposit held",
		"booking_id", bookingID, "amount", amount, "external_id", result.ExternalID)
	return nil
}

// ReleaseDeposit cancels the card authorization hold, returning the deposit to the client.
func (s *securityDepositService) ReleaseDeposit(ctx context.Context, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.DepositStatus == domain.DepositReleased {
		return domain.ErrDepositAlreadyReleased
	}
	if booking.DepositStatus == domain.DepositClaimed {
		return domain.ErrDepositAlreadyClaimed
	}
	if booking.DepositStatus != domain.DepositHeld && booking.DepositStatus != domain.DepositDisputed {
		return domain.ErrDepositNotFound
	}

	// Cancel the card authorization hold
	if booking.DepositExternalID != "" {
		if err := s.provider.CancelPayment(ctx, booking.DepositExternalID); err != nil {
			s.logger.Error("failed to cancel deposit hold with provider",
				"booking_id", bookingID, "external_id", booking.DepositExternalID, "error", err)
			// Continue to update status - the hold will expire naturally
		}
	}

	now := time.Now()
	if err := s.bookingRepo.UpdateDepositStatus(ctx, bookingID, domain.DepositReleased, &now); err != nil {
		return fmt.Errorf("update deposit status to released: %w", err)
	}

	s.logger.Info("Security deposit released",
		"booking_id", bookingID, "amount", booking.DepositAmount)
	return nil
}

// ClaimDeposit captures the held deposit amount (charges the client's card).
func (s *securityDepositService) ClaimDeposit(ctx context.Context, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.DepositStatus == domain.DepositClaimed {
		return domain.ErrDepositAlreadyClaimed
	}
	if booking.DepositStatus == domain.DepositReleased {
		return domain.ErrDepositAlreadyReleased
	}
	if booking.DepositStatus != domain.DepositHeld && booking.DepositStatus != domain.DepositDisputed {
		return domain.ErrDepositNotFound
	}

	// Capture the held amount
	if booking.DepositExternalID != "" {
		if err := s.provider.CapturePayment(ctx, booking.DepositExternalID, booking.DepositAmount); err != nil {
			return fmt.Errorf("capture deposit: %w", err)
		}
	}

	now := time.Now()
	if err := s.bookingRepo.UpdateDepositStatus(ctx, bookingID, domain.DepositClaimed, &now); err != nil {
		return fmt.Errorf("update deposit status to claimed: %w", err)
	}

	s.logger.Info("Security deposit claimed",
		"booking_id", bookingID, "amount", booking.DepositAmount)
	return nil
}

// FreezeDeposit marks the deposit as disputed, preventing auto-release.
func (s *securityDepositService) FreezeDeposit(ctx context.Context, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.DepositStatus != domain.DepositHeld {
		// Only held deposits can be frozen; if already released/claimed/disputed, skip
		return nil
	}

	if err := s.bookingRepo.UpdateDepositStatus(ctx, bookingID, domain.DepositDisputed, nil); err != nil {
		return fmt.Errorf("freeze deposit: %w", err)
	}

	s.logger.Info("Security deposit frozen for dispute",
		"booking_id", bookingID, "amount", booking.DepositAmount)
	return nil
}

// ProcessMaturedDeposits auto-releases deposits for completed bookings
// where the claim period has elapsed and no dispute exists.
func (s *securityDepositService) ProcessMaturedDeposits(ctx context.Context) (int, error) {
	cutoff := time.Now().Add(-time.Duration(s.claimHours) * time.Hour)
	bookings, err := s.bookingRepo.ListHeldDepositsReadyForRelease(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("list matured deposits: %w", err)
	}

	released := 0
	for _, booking := range bookings {
		if err := s.ReleaseDeposit(ctx, booking.ID); err != nil {
			s.logger.Error("failed to auto-release deposit",
				"booking_id", booking.ID, "error", err)
			continue
		}
		released++
	}

	return released, nil
}
