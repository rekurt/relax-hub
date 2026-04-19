package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/payment"
)

// currencyAndProvider returns the payment currency, provider name, and the actual PaymentProvider
// for the user's region. Falls back to RUB/yookassa/default provider if the wallet is not found.
func (s *paymentService) currencyAndProvider(ctx context.Context, userID uuid.UUID) (string, string, payment.PaymentProvider) {
	if s.walletSvc != nil {
		if w, err := s.walletSvc.GetWallet(ctx, userID); err == nil {
			switch w.Currency {
			case domain.WalletCurrencyBYN:
				if factory, ok := s.provider.(*payment.ProviderFactory); ok {
					if p, fErr := factory.ProviderForRegion("BY"); fErr == nil {
						return "BYN", "bepaid", p
					}
				}
				return "BYN", "bepaid", s.provider
			}
		}
	}
	return "RUB", "yookassa", s.provider
}

// providerForPayment returns the PaymentProvider for an existing payment record,
// based on its stored Provider field.
func (s *paymentService) providerForPayment(p *domain.Payment) payment.PaymentProvider {
	if factory, ok := s.provider.(*payment.ProviderFactory); ok {
		region := "RU"
		if p.Provider == "bepaid" {
			region = "BY"
		}
		if prov, err := factory.ProviderForRegion(region); err == nil {
			return prov
		}
	}
	return s.provider
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

	if paymentMethod == "" {
		paymentMethod = domain.PaymentMethodCard
	}

	// For request-based bookings (pending_owner), use authorization hold instead of immediate capture
	isHold := booking.Status == domain.BookingPendingOwner
	capture := !isHold

	currency, providerName, regionProvider := s.currencyAndProvider(ctx, userID)

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      currency,
		Status:        domain.PaymentPending,
		Provider:      providerName,
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
	result, err := regionProvider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:      p.Amount,
		Currency:    p.Currency,
		Description: description,
		ReturnURL:   s.returnURL,
		Metadata:    p.Metadata,
		Method:      string(paymentMethod),
		Capture:     capture,
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

	currency, providerName, regionProvider := s.currencyAndProvider(ctx, userID)

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      currency,
		Status:        domain.PaymentPending,
		Provider:      providerName,
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
	result, err := regionProvider.CreatePayment(ctx, payment.CreatePaymentRequest{
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

	paymentMethod := req.PaymentMethod
	if req.WalletAmount > 0 && req.CardAmount > 0 {
		paymentMethod = domain.PaymentMethodCombo
	} else if req.WalletAmount > 0 && req.CardAmount == 0 {
		paymentMethod = domain.PaymentMethodWallet
	}

	currency, providerName, regionProvider := s.currencyAndProvider(ctx, userID)

	now := time.Now()
	p := &domain.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		UserID:        booking.UserID,
		Amount:        booking.TotalPrice,
		Currency:      currency,
		Status:        domain.PaymentPending,
		Provider:      providerName,
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

	var walletID uuid.UUID
	if req.WalletAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, userID)
		if err != nil {
			return "", err
		}
		walletID = wallet.ID

		if isHold {
			holdExpiry := now.Add(72 * time.Hour)
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
		if req.WalletAmount > 0 {
			if isHold {
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

	if req.CardAmount == 0 {
		if isHold {
			if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentProcessing, ""); err != nil {
				return "", err
			}
			return "", nil
		}
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentSucceeded, ""); err != nil {
			return "", err
		}
		if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed); err != nil {
			return "", fmt.Errorf("failed to confirm booking after wallet payment: %w", err)
		}
		s.sendPaymentConfirmationNotification(ctx, booking)
		return "", nil
	}

	cardMethod := req.PaymentMethod
	if cardMethod == "" || cardMethod == domain.PaymentMethodWallet || cardMethod == domain.PaymentMethodCombo {
		cardMethod = domain.PaymentMethodCard
	}

	description := fmt.Sprintf("Оплата бронирования %s", bookingID.String()[:8])
	result, err := regionProvider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:      req.CardAmount,
		Currency:    p.Currency,
		Description: description,
		ReturnURL:   s.returnURL,
		Metadata:    p.Metadata,
		Method:      string(cardMethod),
		Capture:     capture,
	})
	if err != nil {
		if req.WalletAmount > 0 {
			if isHold {
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
