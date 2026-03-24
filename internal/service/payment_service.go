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

// ComboPaymentRequest describes how a booking payment should be split between wallet and card/SBP.
type ComboPaymentRequest struct {
	WalletAmount  int64             // amount to pay from wallet (0 = card only)
	CardAmount    int64             // amount to pay by card/SBP (0 = wallet only)
	PaymentMethod domain.PaymentMethod // "card" or "sbp" (for card portion)
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, paymentMethod domain.PaymentMethod) (confirmationURL string, err error)
	InitiateComboPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req ComboPaymentRequest) (confirmationURL string, err error)
	HandleWebhook(ctx context.Context, event WebhookEvent) error
	RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool) error
	GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error)
	ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
	bookingRepo repository.BookingRepository
	provider    payment.PaymentProvider
	walletSvc   WalletService
	notifSvc    NotificationService
	returnURL   string
	logger      *logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	provider payment.PaymentProvider,
	walletSvc WalletService,
	notifSvc NotificationService,
	returnURL string,
	log *logger.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
		provider:    provider,
		walletSvc:   walletSvc,
		notifSvc:    notifSvc,
		returnURL:   returnURL,
		logger:      log,
	}
}

func (s *paymentService) InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, paymentMethod domain.PaymentMethod) (string, error) {
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

	if paymentMethod == "" {
		paymentMethod = domain.PaymentMethodCard
	}

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      "RUB",
		Status:        domain.PaymentPending,
		Provider:      "yookassa",
		PaymentMethod: paymentMethod,
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
	result, err := s.provider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:      p.Amount,
		Currency:    p.Currency,
		Description: description,
		ReturnURL:   s.returnURL,
		Metadata:    p.Metadata,
		Method:      string(paymentMethod),
		Capture:     true,
	})
	if err != nil {
		// Update payment status to failed
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		s.logger.Error("payment provider error", "payment_id", p.ID, "error", err)
		return "", domain.ErrPaymentFailed
	}

	// Update with external ID and processing status
	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, result.ExternalID); err != nil {
		return "", err
	}

	return result.ConfirmationURL, nil
}

func (s *paymentService) InitiateComboPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req ComboPaymentRequest) (string, error) {
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

	if req.WalletAmount < 0 || req.CardAmount < 0 {
		return "", fmt.Errorf("%w: payment amounts must be non-negative", domain.ErrInvalidInput)
	}

	if req.WalletAmount+req.CardAmount != booking.TotalPrice {
		return "", fmt.Errorf("%w: wallet_amount + card_amount must equal total price", domain.ErrInvalidInput)
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
		if existing.Status == domain.PaymentFailed {
			if err := s.paymentRepo.Delete(ctx, existing.ID); err != nil {
				return "", fmt.Errorf("failed to delete failed payment: %w", err)
			}
		}
	}

	// Determine the effective payment method
	paymentMethod := req.PaymentMethod
	if req.WalletAmount > 0 && req.CardAmount > 0 {
		paymentMethod = domain.PaymentMethodCombo
	} else if req.WalletAmount > 0 && req.CardAmount == 0 {
		paymentMethod = domain.PaymentMethodWallet
	}
	// If CardAmount only, keep the provided method (card/sbp)

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      "RUB",
		Status:        domain.PaymentPending,
		Provider:      "yookassa",
		PaymentMethod: paymentMethod,
		WalletAmount:  req.WalletAmount,
		CardAmount:    req.CardAmount,
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

	// Step 1: Debit wallet if wallet portion exists
	var walletID uuid.UUID
	if req.WalletAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, userID)
		if err != nil {
			return "", err
		}
		walletID = wallet.ID

		bookingIDRef := bookingID
		_, err = s.walletSvc.Spend(ctx, walletID, req.WalletAmount, "booking_payment", &bookingIDRef, fmt.Sprintf("Оплата бронирования %s", bookingID.String()[:8]))
		if err != nil {
			return "", err
		}
	}

	if err := s.paymentRepo.Create(ctx, p); err != nil {
		// Rollback wallet debit
		if req.WalletAmount > 0 {
			bookingIDRef := bookingID
			if _, refundErr := s.walletSvc.Refund(ctx, walletID, req.WalletAmount, "booking_payment_rollback", &bookingIDRef, "Возврат: ошибка создания платежа"); refundErr != nil {
				s.logger.Error("failed to refund wallet after payment creation error",
					"user_id", userID, "wallet_id", walletID, "amount", req.WalletAmount, "error", refundErr)
			}
		}
		return "", err
	}

	// Step 2: If full wallet payment, complete immediately
	if req.CardAmount == 0 {
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentSucceeded, ""); err != nil {
			return "", err
		}
		if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed); err != nil {
			return "", fmt.Errorf("failed to confirm booking after wallet payment: %w", err)
		}
		s.sendPaymentConfirmationNotification(ctx, booking)
		return "", nil
	}

	// Step 3: Create card/SBP payment for the card portion
	cardMethod := req.PaymentMethod
	if cardMethod == "" || cardMethod == domain.PaymentMethodWallet || cardMethod == domain.PaymentMethodCombo {
		cardMethod = domain.PaymentMethodCard
	}

	description := fmt.Sprintf("Оплата бронирования %s", bookingID.String()[:8])
	result, err := s.provider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:      req.CardAmount,
		Currency:    p.Currency,
		Description: description,
		ReturnURL:   s.returnURL,
		Metadata:    p.Metadata,
		Method:      string(cardMethod),
		Capture:     true,
	})
	if err != nil {
		// Rollback wallet debit on card failure
		if req.WalletAmount > 0 {
			bookingIDRef := bookingID
			if _, refundErr := s.walletSvc.Refund(ctx, walletID, req.WalletAmount, "booking_payment_rollback", &bookingIDRef, "Возврат: ошибка оплаты картой"); refundErr != nil {
				s.logger.Error("failed to refund wallet after card payment error",
					"user_id", userID, "wallet_id", walletID, "amount", req.WalletAmount, "error", refundErr)
			}
		}
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		s.logger.Error("payment provider error", "payment_id", p.ID, "error", err)
		return "", domain.ErrPaymentFailed
	}

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
		// Already in terminal state - ensure downstream actions completed
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
			if booking.Status == domain.BookingCancelled {
				// Payment succeeded but booking was cancelled - retry auto-refund
				s.logger.Warn("retrying auto-refund for cancelled booking",
					"booking_id", p.BookingID, "payment_id", p.ID)
				if refundErr := s.provider.CreateRefund(ctx, p.ExternalID, p.Amount); refundErr != nil {
					return fmt.Errorf("failed to auto-refund cancelled booking on retry: %w", refundErr)
				}
				now := time.Now()
				return s.paymentRepo.UpdateRefund(ctx, p.ID, p.Amount, now, domain.PaymentRefunded)
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
		// Send notification about successful payment and booking confirmation
		s.sendPaymentConfirmationNotification(ctx, booking)
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

func (s *paymentService) sendPaymentConfirmationNotification(ctx context.Context, booking *domain.Booking) {
	title := "Оплата прошла успешно"
	body := fmt.Sprintf("Бронирование на %s оплачено и подтверждено", booking.StartTime.Format("02.01.2006 15:04"))
	data := map[string]string{
		"booking_id":   booking.ID.String(),
		"bathhouse_id": booking.BathhouseID.String(),
	}
	if err := s.notifSvc.Send(ctx, booking.UserID, domain.NotifBookingConfirmed, title, body, data); err != nil {
		s.logger.Warn("failed to send payment confirmation notification",
			"booking_id", booking.ID, "error", err)
	}
}
