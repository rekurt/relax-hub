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

const (
	fullRefundDeadline = 24 * time.Hour
	noRefundDeadline   = 2 * time.Hour
)

// WebhookEvent represents a payment webhook notification from the provider.
type WebhookEvent struct {
	ExternalID string
	Status     string // "succeeded", "canceled", etc.
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (confirmationURL string, err error)
	HandleWebhook(ctx context.Context, event WebhookEvent) error
	RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool) error
	GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error)
	ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
	bookingRepo repository.BookingRepository
	provider    payment.PaymentProvider
	returnURL   string
	logger      *logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	provider payment.PaymentProvider,
	returnURL string,
	log *logger.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
		provider:    provider,
		returnURL:   returnURL,
		logger:      log,
	}
}

func (s *paymentService) InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (string, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return "", err
	}

	if booking.UserID != userID {
		return "", domain.ErrForbidden
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingConfirmed {
		return "", fmt.Errorf("%w: booking must be pending or confirmed to pay", domain.ErrInvalidInput)
	}

	// Check if payment already exists for this booking
	existing, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err == nil && existing != nil {
		if existing.Status == domain.PaymentSucceeded {
			return "", domain.ErrPaymentAlreadyProcessed
		}
		if existing.Status == domain.PaymentPending || existing.Status == domain.PaymentProcessing {
			return "", domain.ErrPaymentAlreadyProcessed
		}
		// If previous payment failed, delete it to allow retry
		if existing.Status == domain.PaymentFailed {
			if err := s.paymentRepo.Delete(ctx, existing.ID); err != nil {
				return "", fmt.Errorf("failed to delete failed payment: %w", err)
			}
		}
	}

	now := time.Now()
	p := &domain.Payment{
		ID:        uuid.New(),
		BookingID: bookingID,
		UserID:    booking.UserID,
		Amount:    booking.TotalPrice,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
		Metadata: map[string]string{
			"booking_id": bookingID.String(),
			"user_id":    booking.UserID.String(),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := p.Validate(); err != nil {
		return "", err
	}

	if err := s.paymentRepo.Create(ctx, p); err != nil {
		return "", err
	}

	description := fmt.Sprintf("Оплата бронирования %s", bookingID.String()[:8])
	result, err := s.provider.CreatePayment(ctx, p.Amount, p.Currency, description, s.returnURL, p.Metadata)
	if err != nil {
		// Update payment status to failed
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		return "", fmt.Errorf("%w: %v", domain.ErrPaymentFailed, err)
	}

	// Update with external ID and processing status
	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, result.ExternalID); err != nil {
		return "", err
	}

	return result.ConfirmationURL, nil
}

func (s *paymentService) HandleWebhook(ctx context.Context, event WebhookEvent) error {
	p, err := s.paymentRepo.GetByExternalID(ctx, event.ExternalID)
	if err != nil {
		return err
	}

	if p.Status == domain.PaymentSucceeded || p.Status == domain.PaymentRefunded || p.Status == domain.PaymentPartiallyRefunded {
		// Already in terminal state - ensure downstream actions completed (booking confirmation)
		if p.Status == domain.PaymentSucceeded {
			booking, err := s.bookingRepo.GetByID(ctx, p.BookingID)
			if err != nil {
				return fmt.Errorf("failed to get booking for idempotency check: %w", err)
			}
			if booking.Status == domain.BookingPending {
				// Payment succeeded but booking was never confirmed (previous webhook partially failed)
				if err := s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingConfirmed); err != nil {
					return fmt.Errorf("failed to confirm booking on retry: %w", err)
				}
			}
		}
		return nil
	}

	// Verify the webhook by checking the actual payment status at the provider
	providerStatus, err := s.provider.GetPaymentStatus(ctx, event.ExternalID)
	if err != nil {
		s.logger.Error("failed to verify payment status with provider",
			"external_id", event.ExternalID, "error", err)
		return fmt.Errorf("failed to verify payment with provider: %w", err)
	}
	if providerStatus != event.Status {
		s.logger.Warn("webhook status mismatch with provider",
			"external_id", event.ExternalID, "webhook_status", event.Status, "provider_status", providerStatus)
		return fmt.Errorf("%w: webhook status does not match provider", domain.ErrInvalidInput)
	}

	switch event.Status {
	case "succeeded":
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentSucceeded, event.ExternalID); err != nil {
			return err
		}
		// Check booking status before confirming - don't re-confirm cancelled bookings
		booking, err := s.bookingRepo.GetByID(ctx, p.BookingID)
		if err != nil {
			return fmt.Errorf("failed to get booking for confirmation: %w", err)
		}
		if booking.Status == domain.BookingCancelled {
			// Booking was cancelled while payment was processing - issue automatic refund
			s.logger.Warn("booking already cancelled, issuing automatic refund",
				"booking_id", p.BookingID, "payment_id", p.ID)
			if refundErr := s.provider.CreateRefund(ctx, event.ExternalID, p.Amount); refundErr != nil {
				s.logger.Error("failed to auto-refund cancelled booking payment",
					"booking_id", p.BookingID, "payment_id", p.ID, "error", refundErr)
				return fmt.Errorf("failed to auto-refund cancelled booking: %w", refundErr)
			}
			now := time.Now()
			return s.paymentRepo.UpdateRefund(ctx, p.ID, p.Amount, now, domain.PaymentRefunded)
		}
		// Confirm the booking
		if err := s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingConfirmed); err != nil {
			return fmt.Errorf("failed to confirm booking after payment: %w", err)
		}
	case "canceled":
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, event.ExternalID); err != nil {
			return err
		}
	default:
		s.logger.Warn("unknown webhook status", "external_id", event.ExternalID, "status", event.Status)
	}

	return nil
}

func (s *paymentService) RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if p.Status != domain.PaymentSucceeded {
		return fmt.Errorf("%w: only succeeded payments can be refunded", domain.ErrInvalidInput)
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	var refundAmount int64
	var refundStatus domain.PaymentStatus

	if forceFullRefund {
		// Owner/representative-initiated cancellation: always full refund
		refundAmount = p.Amount
		refundStatus = domain.PaymentRefunded
	} else {
		timeUntilStart := time.Until(booking.StartTime)

		// No refund if less than 2 hours until start
		if timeUntilStart < noRefundDeadline {
			return nil
		}

		// Full refund if more than 24 hours until start
		if timeUntilStart >= fullRefundDeadline {
			refundAmount = p.Amount
			refundStatus = domain.PaymentRefunded
		} else {
			// Partial refund: 50% between 2h and 24h
			refundAmount = p.Amount / 2
			refundStatus = domain.PaymentPartiallyRefunded
		}
	}

	if err := s.provider.CreateRefund(ctx, p.ExternalID, refundAmount); err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	now := time.Now()
	if err := s.paymentRepo.UpdateRefund(ctx, p.ID, refundAmount, now, refundStatus); err != nil {
		return err
	}

	return nil
}

func (s *paymentService) GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error) {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return p, nil
}

func (s *paymentService) ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
	return s.paymentRepo.ListByUser(ctx, userID, page, pageSize)
}
