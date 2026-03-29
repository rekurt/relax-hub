package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/fiscal"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// Default cancellation policy constants kept for backward compatibility.
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
	WalletAmount  int64                // amount to pay from wallet (0 = card only)
	CardAmount    int64                // amount to pay by card/SBP (0 = wallet only)
	PaymentMethod domain.PaymentMethod // "card" or "sbp" (for card portion)
}

// TokenPaymentRequest extends a standard payment with a client-side payment token (Apple Pay / Google Pay).
type TokenPaymentRequest struct {
	PaymentMethod domain.PaymentMethod
	PaymentToken  string // token from Apple Pay JS or Google Pay API
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, paymentMethod domain.PaymentMethod) (confirmationURL string, err error)
	InitiateTokenPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req TokenPaymentRequest) (confirmationURL string, err error)
	InitiateComboPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req ComboPaymentRequest) (confirmationURL string, err error)
	HandleWebhook(ctx context.Context, event WebhookEvent) error
	RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool, refundTo string, policy domain.CancellationPolicy) error
	AdminRefund(ctx context.Context, adminUserID uuid.UUID, bookingID uuid.UUID, amount int64, reason string, refundTo string) error
	CaptureHoldPayment(ctx context.Context, bookingID uuid.UUID) error
	ReleaseHoldPayment(ctx context.Context, bookingID uuid.UUID) error
	GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error)
	ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

const defaultWalletRefundBonusPercent = 5

type paymentService struct {
	paymentRepo              repository.PaymentRepository
	bookingRepo              repository.BookingRepository
	bathhouseRepo            repository.BathhouseRepository
	kycRepo                  repository.KYCRepository
	auditLogRepo             repository.AuditLogRepository
	provider                 payment.PaymentProvider
	fiscalProvider           fiscal.FiscalProvider
	walletSvc                WalletService
	notifSvc                 NotificationService
	returnURL                string
	walletRefundBonusPercent int
	logger                   *logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	kycRepo repository.KYCRepository,
	auditLogRepo repository.AuditLogRepository,
	provider payment.PaymentProvider,
	fiscalProvider fiscal.FiscalProvider,
	walletSvc WalletService,
	notifSvc NotificationService,
	returnURL string,
	walletRefundBonusPercent int,
	log *logger.Logger,
) PaymentService {
	if walletRefundBonusPercent < 0 || walletRefundBonusPercent > 15 {
		walletRefundBonusPercent = defaultWalletRefundBonusPercent
	}
	return &paymentService{
		paymentRepo:              paymentRepo,
		bookingRepo:              bookingRepo,
		bathhouseRepo:            bathhouseRepo,
		kycRepo:                  kycRepo,
		auditLogRepo:             auditLogRepo,
		provider:                 provider,
		fiscalProvider:           fiscalProvider,
		walletSvc:                walletSvc,
		notifSvc:                 notifSvc,
		returnURL:                returnURL,
		walletRefundBonusPercent: walletRefundBonusPercent,
		logger:                   log,
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

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner && booking.Status != domain.BookingConfirmed {
		return "", fmt.Errorf("%w: booking must be pending, pending_owner, or confirmed to pay", domain.ErrInvalidInput)
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

	// For request-based bookings (pending_owner), use authorization hold instead of immediate capture
	isHold := booking.Status == domain.BookingPendingOwner
	capture := !isHold

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
		IsHold:        isHold,
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
		Capture:     capture,
	})
	if err != nil {
		// Update payment status to failed
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		s.logger.Error("payment provider error", "payment_id", p.ID, "error", err)
		info := payment.ClassifyError(err)
		return "", &domain.PaymentFailedError{
			Code:       info.Code,
			MessageRU:  info.MessageRU,
			Suggestion: info.Suggestion,
		}
	}

	// Update with external ID and processing status
	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, result.ExternalID); err != nil {
		return "", err
	}

	return result.ConfirmationURL, nil
}

func (s *paymentService) InitiateTokenPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req TokenPaymentRequest) (string, error) {
	if req.PaymentToken == "" {
		return "", fmt.Errorf("%w: payment_token is required for token-based payments", domain.ErrInvalidInput)
	}
	if !req.PaymentMethod.IsTokenBased() {
		return "", fmt.Errorf("%w: payment_method must be apple_pay or google_pay for token payments", domain.ErrInvalidInput)
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return "", err
	}

	if booking.UserID != userID {
		return "", domain.ErrForbidden
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner && booking.Status != domain.BookingConfirmed {
		return "", fmt.Errorf("%w: booking must be pending, pending_owner, or confirmed to pay", domain.ErrInvalidInput)
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

	isHold := booking.Status == domain.BookingPendingOwner
	capture := !isHold

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      "RUB",
		Status:        domain.PaymentPending,
		Provider:      "yookassa",
		PaymentMethod: req.PaymentMethod,
		IsHold:        isHold,
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
		Amount:       p.Amount,
		Currency:     p.Currency,
		Description:  description,
		ReturnURL:    s.returnURL,
		Metadata:     p.Metadata,
		Method:       string(req.PaymentMethod),
		Capture:      capture,
		PaymentToken: req.PaymentToken,
	})
	if err != nil {
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		s.logger.Error("payment provider error", "payment_id", p.ID, "error", err)
		info := payment.ClassifyError(err)
		return "", &domain.PaymentFailedError{
			Code:       info.Code,
			MessageRU:  info.MessageRU,
			Suggestion: info.Suggestion,
		}
	}

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

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner && booking.Status != domain.BookingConfirmed {
		return "", fmt.Errorf("%w: booking must be pending, pending_owner, or confirmed to pay", domain.ErrInvalidInput)
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

	// For request-based bookings (pending_owner), use holds instead of immediate debit/capture
	isHold := booking.Status == domain.BookingPendingOwner
	capture := !isHold

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
		IsHold:        isHold,
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

	// Step 1: Debit/hold wallet if wallet portion exists
	var walletID uuid.UUID
	if req.WalletAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, userID)
		if err != nil {
			return "", err
		}
		walletID = wallet.ID

		if isHold {
			// For request-based bookings, create a wallet hold instead of spending
			holdExpiry := now.Add(72 * time.Hour) // default 72h
			bookingIDRef := bookingID
			_, err = s.walletSvc.Hold(ctx, walletID, req.WalletAmount, "booking_payment", &bookingIDRef, fmt.Sprintf("Hold for booking payment %s", bookingID.String()[:8]), holdExpiry)
			if err != nil {
				return "", err
			}
		} else {
			bookingIDRef := bookingID
			_, err = s.walletSvc.Spend(ctx, walletID, req.WalletAmount, "booking_payment", &bookingIDRef, fmt.Sprintf("Оплата бронирования %s", bookingID.String()[:8]))
			if err != nil {
				return "", err
			}
		}
	}

	if err := s.paymentRepo.Create(ctx, p); err != nil {
		// Rollback wallet debit/hold
		if req.WalletAmount > 0 {
			if isHold {
				// Release wallet hold - find active holds for booking
				holds, holdErr := s.walletSvc.GetActiveHolds(ctx, walletID)
				if holdErr == nil {
					for _, h := range holds {
						if h.ReferenceID != nil && *h.ReferenceID == bookingID && h.ReferenceType == "booking_payment" {
							if releaseErr := s.walletSvc.ReleaseHold(ctx, h.ID); releaseErr != nil {
								s.logger.Error("failed to release wallet hold after payment creation error",
									"hold_id", h.ID, "error", releaseErr)
							}
							break
						}
					}
				} else {
					s.logger.Error("failed to find wallet holds for rollback",
						"user_id", userID, "wallet_id", walletID, "error", holdErr)
				}
			} else {
				bookingIDRef := bookingID
				if _, refundErr := s.walletSvc.Refund(ctx, walletID, req.WalletAmount, "booking_payment_rollback", &bookingIDRef, "Возврат: ошибка создания платежа"); refundErr != nil {
					s.logger.Error("failed to refund wallet after payment creation error",
						"user_id", userID, "wallet_id", walletID, "amount", req.WalletAmount, "error", refundErr)
				}
			}
		}
		return "", err
	}

	// Step 2: If full wallet payment
	if req.CardAmount == 0 {
		if isHold {
			// For request-based wallet-only: hold is already created, payment stays in processing
			if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, ""); err != nil {
				return "", err
			}
			return "", nil
		}
		// Instant wallet payment: complete immediately
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
		Capture:     capture,
	})
	if err != nil {
		// Rollback wallet debit/hold on card failure
		if req.WalletAmount > 0 {
			if isHold {
				// Release wallet hold - find active holds for booking
				holds, holdErr := s.walletSvc.GetActiveHolds(ctx, walletID)
				if holdErr == nil {
					for _, h := range holds {
						if h.ReferenceID != nil && *h.ReferenceID == bookingID && h.ReferenceType == "booking_payment" {
							if releaseErr := s.walletSvc.ReleaseHold(ctx, h.ID); releaseErr != nil {
								s.logger.Error("failed to release wallet hold after card failure",
									"hold_id", h.ID, "error", releaseErr)
							}
							break
						}
					}
				}
			} else {
				bookingIDRef := bookingID
				if _, refundErr := s.walletSvc.Refund(ctx, walletID, req.WalletAmount, "booking_payment_rollback", &bookingIDRef, "Возврат: ошибка оплаты картой"); refundErr != nil {
					s.logger.Error("failed to refund wallet after card payment error",
						"user_id", userID, "wallet_id", walletID, "amount", req.WalletAmount, "error", refundErr)
				}
			}
		}
		if updErr := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, ""); updErr != nil {
			s.logger.Error("failed to update payment status after provider error",
				"payment_id", p.ID, "error", updErr)
		}
		s.logger.Error("payment provider error", "payment_id", p.ID, "error", err)
		info := payment.ClassifyError(err)
		return "", &domain.PaymentFailedError{
			Code:       info.Code,
			MessageRU:  info.MessageRU,
			Suggestion: info.Suggestion,
		}
	}

	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, result.ExternalID); err != nil {
		return "", err
	}

	return result.ConfirmationURL, nil
}

func (s *paymentService) CaptureHoldPayment(ctx context.Context, bookingID uuid.UUID) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if !p.IsHold {
		// Not a hold payment - nothing to capture
		return nil
	}

	if p.Status != domain.PaymentProcessing {
		return fmt.Errorf("%w: only processing hold payments can be captured", domain.ErrInvalidInput)
	}

	// Capture card portion if external payment exists
	if p.ExternalID != "" {
		// For non-combo payments, CardAmount is 0 — use full Amount instead
		captureAmount := p.CardAmount
		if captureAmount == 0 {
			captureAmount = p.Amount
		}
		if err := s.provider.CapturePayment(ctx, p.ExternalID, captureAmount); err != nil {
			return fmt.Errorf("failed to capture card hold: %w", err)
		}
	}

	// Capture wallet hold if wallet portion exists
	if p.WalletAmount > 0 && s.walletSvc != nil {
		wallet, walletErr := s.walletSvc.GetWallet(ctx, p.UserID)
		if walletErr != nil {
			return fmt.Errorf("failed to get wallet for hold capture: %w", walletErr)
		}
		holds, holdErr := s.walletSvc.GetActiveHolds(ctx, wallet.ID)
		if holdErr != nil {
			return fmt.Errorf("failed to get active holds for capture: %w", holdErr)
		}
		walletHoldCaptured := false
		for _, h := range holds {
			if h.ReferenceID != nil && *h.ReferenceID == bookingID && h.ReferenceType == "booking_payment" {
				if _, captureErr := s.walletSvc.CaptureHold(ctx, h.ID); captureErr != nil {
					return fmt.Errorf("failed to capture wallet hold after card capture: %w", captureErr)
				}
				walletHoldCaptured = true
				break
			}
		}
		if !walletHoldCaptured {
			s.logger.Warn("wallet hold not found for capture, wallet portion may not be debited",
				"booking_id", bookingID, "user_id", p.UserID)
		}
	}

	now := time.Now()
	if err := s.paymentRepo.UpdateCapture(ctx, p.ID, now, domain.PaymentSucceeded); err != nil {
		return err
	}

	return nil
}

func (s *paymentService) ReleaseHoldPayment(ctx context.Context, bookingID uuid.UUID) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return nil // no payment to release
		}
		return err
	}

	if !p.IsHold {
		return nil
	}

	if p.Status != domain.PaymentProcessing && p.Status != domain.PaymentPending {
		return nil // already in terminal state
	}

	// Cancel card authorization if external payment exists
	if p.ExternalID != "" {
		if err := s.provider.CancelPayment(ctx, p.ExternalID); err != nil {
			s.logger.Error("failed to cancel card hold", "booking_id", bookingID, "external_id", p.ExternalID, "error", err)
		}
	}

	// Release wallet hold if wallet portion exists
	if p.WalletAmount > 0 && s.walletSvc != nil {
		wallet, walletErr := s.walletSvc.GetWallet(ctx, p.UserID)
		if walletErr == nil {
			holds, holdErr := s.walletSvc.GetActiveHolds(ctx, wallet.ID)
			if holdErr == nil {
				for _, h := range holds {
					if h.ReferenceID != nil && *h.ReferenceID == bookingID && h.ReferenceType == "booking_payment" {
						if releaseErr := s.walletSvc.ReleaseHold(ctx, h.ID); releaseErr != nil {
							s.logger.Error("failed to release wallet hold on payment release",
								"booking_id", bookingID, "hold_id", h.ID, "error", releaseErr)
						}
						break
					}
				}
			}
		}
	}

	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, p.ExternalID); err != nil {
		return err
	}

	return nil
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
				cardRefundAmount := p.CardAmount
				if cardRefundAmount == 0 {
					cardRefundAmount = p.Amount
				}
				if cardRefundAmount > 0 && p.ExternalID != "" {
					if refundErr := s.provider.CreateRefund(ctx, p.ExternalID, cardRefundAmount); refundErr != nil {
						return fmt.Errorf("failed to auto-refund cancelled booking on retry: %w", refundErr)
					}
				}
				if p.WalletAmount > 0 && s.walletSvc != nil {
					wallet, wErr := s.walletSvc.GetWallet(ctx, booking.UserID)
					if wErr == nil {
						bookingIDRef := p.BookingID
						if _, rErr := s.walletSvc.Refund(ctx, wallet.ID, p.WalletAmount, "auto_refund_cancelled_retry", &bookingIDRef,
							fmt.Sprintf("Автовозврат (повтор): бронирование отменено %s", p.BookingID.String()[:8])); rErr != nil {
							s.logger.Error("failed to auto-refund wallet on retry",
								"payment_id", p.ID, "amount", p.WalletAmount, "error", rErr)
						}
					}
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
	case "waiting_for_capture":
		// Hold payment authorized - don't confirm booking yet, wait for owner approval
		if p.IsHold {
			// Just update the status, booking stays in pending_owner
			return nil
		}
		// Non-hold payment shouldn't get waiting_for_capture, log and ignore
		s.logger.Warn("unexpected waiting_for_capture for non-hold payment",
			"external_id", event.ExternalID, "payment_id", p.ID)
	case "succeeded":
		// For hold payments, "succeeded" means the hold was captured (by us calling CapturePayment)
		// The capture flow already updates the payment status, so this is just idempotency
		if p.IsHold && p.CapturedAt == nil {
			// This can happen if capture webhook arrives - just record it
			now := time.Now()
			if err := s.paymentRepo.UpdateCapture(ctx, p.ID, now, domain.PaymentSucceeded); err != nil {
				return err
			}
			return nil
		}
		if !p.IsHold {
			if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentSucceeded, event.ExternalID); err != nil {
				return err
			}
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
			// For combo payments, only refund the card portion via provider
			cardRefundAmount := p.CardAmount
			if cardRefundAmount == 0 {
				cardRefundAmount = p.Amount
			}
			if cardRefundAmount > 0 && p.ExternalID != "" {
				if refundErr := s.provider.CreateRefund(ctx, event.ExternalID, cardRefundAmount); refundErr != nil {
					s.logger.Error("failed to auto-refund cancelled booking payment",
						"booking_id", p.BookingID, "payment_id", p.ID, "error", refundErr)
					return fmt.Errorf("failed to auto-refund cancelled booking: %w", refundErr)
				}
			}
			// Refund wallet portion for combo payments
			actualRefunded := cardRefundAmount
			if p.WalletAmount > 0 && s.walletSvc != nil {
				wallet, wErr := s.walletSvc.GetWallet(ctx, booking.UserID)
				if wErr == nil {
					bookingIDRef := p.BookingID
					if _, rErr := s.walletSvc.Refund(ctx, wallet.ID, p.WalletAmount, "auto_refund_cancelled", &bookingIDRef,
						fmt.Sprintf("Автовозврат: бронирование отменено %s", p.BookingID.String()[:8])); rErr != nil {
						s.logger.Error("failed to auto-refund wallet for cancelled booking",
							"payment_id", p.ID, "wallet_id", wallet.ID, "amount", p.WalletAmount, "error", rErr)
					} else {
						actualRefunded += p.WalletAmount
					}
				}
			}
			now := time.Now()
			return s.paymentRepo.UpdateRefund(ctx, p.ID, actualRefunded, now, domain.PaymentRefunded)
		}
		// For hold payments, booking confirmation is handled by BookingService.Approve
		if p.IsHold {
			return nil
		}
		// Confirm the booking
		if err := s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingConfirmed); err != nil {
			return fmt.Errorf("failed to confirm booking after payment: %w", err)
		}
		// Send notification about successful payment and booking confirmation
		s.sendPaymentConfirmationNotification(ctx, booking)
		// Create fiscal receipt (best-effort)
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptAdvance)
	case "canceled":
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, event.ExternalID); err != nil {
			return err
		}
		// Roll back wallet for combo payments where wallet was already debited/held
		if p.WalletAmount > 0 && s.walletSvc != nil {
			booking, bErr := s.bookingRepo.GetByID(ctx, p.BookingID)
			if bErr != nil {
				s.logger.Error("failed to get booking for wallet rollback on canceled payment",
					"payment_id", p.ID, "booking_id", p.BookingID, "error", bErr)
			} else {
				wallet, wErr := s.walletSvc.GetWallet(ctx, booking.UserID)
				if wErr != nil {
					s.logger.Error("failed to get wallet for rollback on canceled payment",
						"payment_id", p.ID, "user_id", booking.UserID, "error", wErr)
				} else if p.IsHold {
					// Release wallet hold
					holds, hErr := s.walletSvc.GetActiveHolds(ctx, wallet.ID)
					if hErr == nil {
						for _, h := range holds {
							if h.ReferenceID != nil && *h.ReferenceID == p.BookingID && h.ReferenceType == "booking_payment" {
								if rErr := s.walletSvc.ReleaseHold(ctx, h.ID); rErr != nil {
									s.logger.Error("failed to release wallet hold on canceled payment",
										"hold_id", h.ID, "error", rErr)
								}
								break
							}
						}
					}
				} else {
					// Refund wallet debit
					bookingIDRef := p.BookingID
					if _, rErr := s.walletSvc.Refund(ctx, wallet.ID, p.WalletAmount, "payment_canceled_rollback", &bookingIDRef,
						fmt.Sprintf("Возврат: оплата отменена %s", p.BookingID.String()[:8])); rErr != nil {
						s.logger.Error("failed to refund wallet on canceled payment",
							"payment_id", p.ID, "wallet_id", wallet.ID, "amount", p.WalletAmount, "error", rErr)
					}
				}
			}
		}
	default:
		s.logger.Warn("unknown webhook status", "external_id", event.ExternalID, "status", event.Status)
	}

	return nil
}

func (s *paymentService) RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool, refundTo string, policy domain.CancellationPolicy) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	// If payment is a hold, release it instead of refunding
	if p.IsHold {
		return s.ReleaseHoldPayment(ctx, bookingID)
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
		refundAmount = p.Amount
		refundStatus = domain.PaymentRefunded
	} else {
		timeUntilStart := time.Until(booking.StartTime)

		// Use per-bathhouse cancellation policy if provided, otherwise fall back to legacy flat rules
		if policy.IsValid() {
			refundAmount = policy.CalculateRefundAmount(p.Amount, timeUntilStart)
		} else {
			if timeUntilStart < noRefundDeadline {
				return nil
			}
			if timeUntilStart >= fullRefundDeadline {
				refundAmount = p.Amount
			} else {
				refundAmount = p.Amount / 2
			}
		}

		if refundAmount == 0 {
			return nil
		}
		if refundAmount == p.Amount {
			refundStatus = domain.PaymentRefunded
		} else {
			refundStatus = domain.PaymentPartiallyRefunded
		}
	}

	return s.executeRefund(ctx, p, booking.UserID, refundAmount, refundStatus, refundTo)
}

func (s *paymentService) executeRefund(ctx context.Context, p *domain.Payment, userID uuid.UUID, refundAmount int64, refundStatus domain.PaymentStatus, refundTo string) error {
	if refundTo == "" {
		refundTo = "card"
	}

	// For combo payments, split refund proportionally
	if p.WalletAmount > 0 && p.CardAmount > 0 {
		if err := s.executeComboRefund(ctx, p, userID, refundAmount, refundStatus, refundTo); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	// Pure wallet payment — always refund to wallet
	if p.PaymentMethod == domain.PaymentMethodWallet {
		if err := s.refundToWallet(ctx, p, userID, refundAmount, refundStatus); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	// Pure card/SBP payment
	if refundTo == "wallet" {
		if err := s.refundToWallet(ctx, p, userID, refundAmount, refundStatus); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	// Default: refund to card
	if p.ExternalID != "" {
		if err := s.provider.CreateRefund(ctx, p.ExternalID, refundAmount); err != nil {
			return fmt.Errorf("failed to create refund: %w", err)
		}
	}

	now := time.Now()
	if err := s.paymentRepo.UpdateRefund(ctx, p.ID, p.RefundAmount+refundAmount, now, refundStatus); err != nil {
		return err
	}
	s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
	return nil
}

func (s *paymentService) refundToWallet(ctx context.Context, p *domain.Payment, userID uuid.UUID, refundAmount int64, refundStatus domain.PaymentStatus) error {
	wallet, err := s.walletSvc.GetWallet(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet for refund: %w", err)
	}

	bonusAmount := int64(math.Round(float64(refundAmount) * float64(s.walletRefundBonusPercent) / 100))
	totalCredit := refundAmount + bonusAmount

	bookingIDRef := p.BookingID
	_, err = s.walletSvc.Refund(ctx, wallet.ID, totalCredit, "booking_refund", &bookingIDRef,
		fmt.Sprintf("Возврат за бронирование %s (бонус %d%%)", p.BookingID.String()[:8], s.walletRefundBonusPercent))
	if err != nil {
		return fmt.Errorf("failed to refund to wallet: %w", err)
	}

	now := time.Now()
	return s.paymentRepo.UpdateRefund(ctx, p.ID, p.RefundAmount+refundAmount, now, refundStatus)
}

func (s *paymentService) executeComboRefund(ctx context.Context, p *domain.Payment, userID uuid.UUID, refundAmount int64, refundStatus domain.PaymentStatus, refundTo string) error {
	// Calculate proportional split
	walletRefundRatio := float64(p.WalletAmount) / float64(p.Amount)
	walletRefund := int64(math.Round(float64(refundAmount) * walletRefundRatio))
	cardRefund := refundAmount - walletRefund

	wallet, err := s.walletSvc.GetWallet(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet for combo refund: %w", err)
	}

	// Wallet portion always goes back to wallet with bonus
	if walletRefund > 0 {
		bonusAmount := int64(math.Round(float64(walletRefund) * float64(s.walletRefundBonusPercent) / 100))
		totalCredit := walletRefund + bonusAmount
		bookingIDRef := p.BookingID
		_, err = s.walletSvc.Refund(ctx, wallet.ID, totalCredit, "booking_refund", &bookingIDRef,
			fmt.Sprintf("Возврат кошелёк за %s (бонус %d%%)", p.BookingID.String()[:8], s.walletRefundBonusPercent))
		if err != nil {
			return fmt.Errorf("failed to refund wallet portion: %w", err)
		}
	}

	// Card portion follows refundTo preference
	if cardRefund > 0 {
		if refundTo == "wallet" {
			// Redirect card portion to wallet without bonus (bonus applies only to wallet portion)
			bookingIDRef := p.BookingID
			_, err = s.walletSvc.Refund(ctx, wallet.ID, cardRefund, "booking_refund_card_to_wallet", &bookingIDRef,
				fmt.Sprintf("Возврат карта→кошелёк за %s", p.BookingID.String()[:8]))
			if err != nil {
				return fmt.Errorf("failed to redirect card refund to wallet: %w", err)
			}
		} else if p.ExternalID != "" {
			if err := s.provider.CreateRefund(ctx, p.ExternalID, cardRefund); err != nil {
				return fmt.Errorf("failed to create card refund: %w", err)
			}
		}
	}

	now := time.Now()
	return s.paymentRepo.UpdateRefund(ctx, p.ID, p.RefundAmount+refundAmount, now, refundStatus)
}

func (s *paymentService) AdminRefund(ctx context.Context, adminUserID uuid.UUID, bookingID uuid.UUID, amount int64, reason string, refundTo string) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if p.Status != domain.PaymentSucceeded && p.Status != domain.PaymentPartiallyRefunded {
		return fmt.Errorf("%w: only succeeded or partially refunded payments can be refunded", domain.ErrInvalidInput)
	}

	if amount <= 0 {
		return fmt.Errorf("%w: refund amount must be positive", domain.ErrInvalidInput)
	}

	maxRefundable := p.Amount - p.RefundAmount
	if amount > maxRefundable {
		return domain.ErrRefundExceedsAmount
	}

	var refundStatus domain.PaymentStatus
	if amount == maxRefundable {
		refundStatus = domain.PaymentRefunded
	} else {
		refundStatus = domain.PaymentPartiallyRefunded
	}

	if err := s.executeRefund(ctx, p, p.UserID, amount, refundStatus, refundTo); err != nil {
		return err
	}

	// Create audit log entry
	if s.auditLogRepo != nil {
		changedFields, _ := json.Marshal(map[string]interface{}{
			"refund_amount": amount,
			"reason":        reason,
			"refund_to":     refundTo,
			"booking_id":    bookingID.String(),
		})
		auditLog := &domain.AuditLog{
			ID:            uuid.New(),
			EntityType:    "payment",
			EntityID:      p.ID,
			UserID:        adminUserID,
			Action:        domain.AuditActionUpdate,
			ChangedFields: changedFields,
			CreatedAt:     time.Now(),
		}
		if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
			s.logger.Error("failed to create audit log for admin refund",
				"payment_id", p.ID, "error", err)
		}
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

func (s *paymentService) createFiscalReceipt(ctx context.Context, p *domain.Payment, receiptType fiscal.ReceiptType, amounts ...int64) {
	if s.fiscalProvider == nil {
		return
	}
	// Use provided amount (e.g. actual refund amount) or fall back to full payment amount
	amount := p.Amount
	if len(amounts) > 0 && amounts[0] > 0 {
		amount = amounts[0]
	}

	// Resolve owner's entity type for correct VAT and tax system
	vat, taxSystem := s.resolveOwnerTaxInfo(ctx, p.BookingID)

	req := fiscal.ReceiptRequest{
		Type:      receiptType,
		Amount:    amount,
		TaxSystem: taxSystem,
		Items: []fiscal.ReceiptItem{
			{
				Name:     fmt.Sprintf("Бронирование %s", p.BookingID.String()[:8]),
				Quantity: 1,
				Price:    amount,
				VAT:      vat,
			},
		},
	}
	receipt, err := s.fiscalProvider.CreateReceipt(ctx, req)
	if err != nil {
		s.logger.Error("failed to create fiscal receipt",
			"payment_id", p.ID, "receipt_type", string(receiptType), "error", err)
		return
	}
	if receipt != nil {
		s.logger.Info("fiscal receipt created",
			"payment_id", p.ID, "receipt_id", receipt.ID, "receipt_type", string(receiptType),
			"tax_system", string(taxSystem), "vat", vat)
	}
}

// resolveOwnerTaxInfo looks up the bathhouse owner's KYC entity type to determine
// the correct VAT rate and tax system for fiscal receipts.
func (s *paymentService) resolveOwnerTaxInfo(ctx context.Context, bookingID uuid.UUID) (string, fiscal.TaxSystem) {
	defaultVAT, defaultTax := "none", fiscal.TaxSystemDefault

	if s.bathhouseRepo == nil || s.kycRepo == nil {
		return defaultVAT, defaultTax
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		s.logger.Warn("fiscal: failed to get booking for tax info", "booking_id", bookingID, "error", err)
		return defaultVAT, defaultTax
	}

	bh, err := s.bathhouseRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		s.logger.Warn("fiscal: failed to get bathhouse for tax info", "bathhouse_id", booking.BathhouseID, "error", err)
		return defaultVAT, defaultTax
	}

	kyc, err := s.kycRepo.GetByUserID(ctx, bh.OwnerID)
	if err != nil {
		s.logger.Debug("fiscal: no KYC for owner, using default tax info", "owner_id", bh.OwnerID)
		return defaultVAT, defaultTax
	}

	if kyc.Status != domain.KYCStatusApproved {
		return defaultVAT, defaultTax
	}

	return fiscal.TaxInfoForEntityType(kyc.EntityType)
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
