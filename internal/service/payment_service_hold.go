package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/fiscal"
)

func (s *paymentService) CaptureHoldPayment(ctx context.Context, bookingID uuid.UUID) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	if !p.IsHold {
		return nil
	}

	if p.Status != domain.PaymentProcessing {
		return fmt.Errorf("%w: only processing hold payments can be captured", domain.ErrInvalidInput)
	}

	if p.ExternalID != "" {
		captureAmount := p.CardAmount
		if captureAmount == 0 {
			captureAmount = p.Amount
		}
		if err := s.providerForPayment(p).CapturePayment(ctx, p.ExternalID, captureAmount); err != nil {
			return fmt.Errorf("failed to capture card hold: %w", err)
		}
	}

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
	return s.paymentRepo.UpdateCapture(ctx, p.ID, now, domain.PaymentSucceeded)
}

func (s *paymentService) ReleaseHoldPayment(ctx context.Context, bookingID uuid.UUID) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return nil
		}
		return err
	}

	if !p.IsHold {
		return nil
	}

	if p.Status != domain.PaymentProcessing && p.Status != domain.PaymentPending {
		return nil
	}

	if p.ExternalID != "" {
		if err := s.providerForPayment(p).CancelPayment(ctx, p.ExternalID); err != nil {
			s.logger.Error("failed to cancel card hold", "booking_id", bookingID, "external_id", p.ExternalID, "error", err)
		}
	}

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

	return s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, p.ExternalID)
}

func (s *paymentService) HandleWebhook(ctx context.Context, event WebhookEvent) error {
	p, err := s.paymentRepo.GetByExternalID(ctx, event.ExternalID)
	if err != nil {
		return err
	}

	if p.Status == domain.PaymentSucceeded || p.Status == domain.PaymentRefunded || p.Status == domain.PaymentPartiallyRefunded {
		if p.Status == domain.PaymentSucceeded {
			booking, err := s.bookingRepo.GetByID(ctx, p.BookingID)
			if err != nil {
				return fmt.Errorf("failed to get booking for idempotency check: %w", err)
			}
			if booking.Status == domain.BookingPending {
				if err := s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingConfirmed); err != nil {
					return fmt.Errorf("failed to confirm booking on retry: %w", err)
				}
			}
			if booking.Status == domain.BookingCancelled {
				s.logger.Warn("retrying auto-refund for cancelled booking",
					"booking_id", p.BookingID, "payment_id", p.ID)
				cardRefundAmount := p.CardAmount
				if cardRefundAmount == 0 {
					cardRefundAmount = p.Amount
				}
				if cardRefundAmount > 0 && p.ExternalID != "" {
					if refundErr := s.providerForPayment(p).CreateRefund(ctx, p.ExternalID, cardRefundAmount); refundErr != nil {
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

	providerStatus, err := s.providerForPayment(p).GetPaymentStatus(ctx, event.ExternalID)
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
		if p.IsHold {
			return nil
		}
		s.logger.Warn("unexpected waiting_for_capture for non-hold payment",
			"external_id", event.ExternalID, "payment_id", p.ID)

	case "succeeded":
		if p.IsHold && p.CapturedAt == nil {
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
		booking, err := s.bookingRepo.GetByID(ctx, p.BookingID)
		if err != nil {
			return fmt.Errorf("failed to get booking for confirmation: %w", err)
		}
		if booking.Status == domain.BookingCancelled {
			s.logger.Warn("booking already cancelled, issuing automatic refund",
				"booking_id", p.BookingID, "payment_id", p.ID)
			cardRefundAmount := p.CardAmount
			if cardRefundAmount == 0 {
				cardRefundAmount = p.Amount
			}
			if cardRefundAmount > 0 && p.ExternalID != "" {
				if refundErr := s.providerForPayment(p).CreateRefund(ctx, event.ExternalID, cardRefundAmount); refundErr != nil {
					s.logger.Error("failed to auto-refund cancelled booking payment",
						"booking_id", p.BookingID, "payment_id", p.ID, "error", refundErr)
					return fmt.Errorf("failed to auto-refund cancelled booking: %w", refundErr)
				}
			}
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
		if p.IsHold {
			return nil
		}
		if err := s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingConfirmed); err != nil {
			return fmt.Errorf("failed to confirm booking after payment: %w", err)
		}
		s.sendPaymentConfirmationNotification(ctx, booking)
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptAdvance)

	case "canceled":
		if err := s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentFailed, event.ExternalID); err != nil {
			return err
		}
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
