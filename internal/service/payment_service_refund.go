package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/fiscal"
)

// Default cancellation policy deadlines for legacy fallback (when no per-bathhouse policy is set).
const (
	fullRefundDeadline = 24 * time.Hour
	noRefundDeadline   = 2 * time.Hour
)

func (s *paymentService) RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool, refundTo string, policy domain.CancellationPolicy) error {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

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

	if p.WalletAmount > 0 && p.CardAmount > 0 {
		if err := s.executeComboRefund(ctx, p, userID, refundAmount, refundStatus, refundTo); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	if p.PaymentMethod == domain.PaymentMethodWallet {
		if err := s.refundToWallet(ctx, p, userID, refundAmount, refundStatus); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	if refundTo == "wallet" {
		if err := s.refundToWallet(ctx, p, userID, refundAmount, refundStatus); err != nil {
			return err
		}
		s.createFiscalReceipt(ctx, p, fiscal.ReceiptRefund, refundAmount)
		return nil
	}

	if p.ExternalID != "" {
		if err := s.providerForPayment(p).CreateRefund(ctx, p.ExternalID, refundAmount); err != nil {
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
	walletRefundRatio := float64(p.WalletAmount) / float64(p.Amount)
	walletRefund := int64(math.Round(float64(refundAmount) * walletRefundRatio))
	cardRefund := refundAmount - walletRefund

	wallet, err := s.walletSvc.GetWallet(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet for combo refund: %w", err)
	}

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

	if cardRefund > 0 {
		if refundTo == "wallet" {
			bookingIDRef := p.BookingID
			_, err = s.walletSvc.Refund(ctx, wallet.ID, cardRefund, "booking_refund_card_to_wallet", &bookingIDRef,
				fmt.Sprintf("Возврат карта→кошелёк за %s", p.BookingID.String()[:8]))
			if err != nil {
				return fmt.Errorf("failed to redirect card refund to wallet: %w", err)
			}
		} else if p.ExternalID != "" {
			if err := s.providerForPayment(p).CreateRefund(ctx, p.ExternalID, cardRefund); err != nil {
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

func (s *paymentService) createFiscalReceipt(ctx context.Context, p *domain.Payment, receiptType fiscal.ReceiptType, amounts ...int64) {
	if s.fiscalProvider == nil {
		return
	}
	amount := p.Amount
	if len(amounts) > 0 && amounts[0] > 0 {
		amount = amounts[0]
	}

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
