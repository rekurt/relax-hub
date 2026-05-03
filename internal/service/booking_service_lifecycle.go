package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// GetByID returns a booking if the caller is allowed to see it.
// Access rules: admin sees any booking; client sees their own; owner/representative
// sees bookings of their bathhouses.
func (s *bookingService) GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	switch role {
	case domain.RoleAdmin:
		return booking, nil
	case domain.RoleClient:
		if booking.UserID != userID {
			return nil, domain.ErrForbidden
		}
		return booking, nil
	case domain.RoleOwner, domain.RoleRepresentative:
		bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
		if err != nil {
			return nil, err
		}
		if bh.OwnerID != userID {
			if s.access == nil {
				return nil, domain.ErrForbidden
			}
			if err := s.access.CanManageBathhouse(ctx, userID, role, bh.ID); err != nil {
				return nil, err
			}
		}
		return booking, nil
	default:
		return nil, domain.ErrForbidden
	}
}

func (s *bookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, refundTo string) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner && booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidInput
	}

	// Client cancels their own booking. Refund amount determined by bathhouse cancellation policy.
	if role == domain.RoleClient {
		if booking.UserID != userID {
			return domain.ErrForbidden
		}
		// Release payment holds and wallet hold if pending_owner
		if booking.Status == domain.BookingPendingOwner {
			if err := s.paymentSvc.ReleaseHoldPayment(ctx, bookingID); err != nil {
				s.logger.Error("failed to release payment hold on client cancel", "booking_id", bookingID, "error", err)
			}
			if booking.HoldID != nil && s.walletSvc != nil {
				if err := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); err != nil {
					s.logger.Warn("failed to release booking wallet hold on client cancel", "booking_id", bookingID, "hold_id", booking.HoldID, "error", err)
				}
			}
		}
		if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); err != nil {
			return err
		}
		s.refundBookingPoints(ctx, booking)
		s.refundReferralBonus(ctx, booking)
		s.refundPromoUsage(ctx, booking)
		s.refundCertificateUsage(ctx, booking)
		if booking.Status != domain.BookingPendingOwner {
			s.refundPayment(ctx, booking, false, refundTo)
		}
		s.releaseDepositOnCancel(ctx, booking)
		s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)
		return nil
	}

	// Owner/representative cancels booking for their bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	// Release payment holds and wallet hold if pending_owner
	if booking.Status == domain.BookingPendingOwner {
		if err := s.paymentSvc.ReleaseHoldPayment(ctx, bookingID); err != nil {
			s.logger.Error("failed to release payment hold on owner cancel", "booking_id", bookingID, "error", err)
		}
		if booking.HoldID != nil && s.walletSvc != nil {
			if err := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); err != nil {
				s.logger.Warn("failed to release booking wallet hold on owner cancel", "booking_id", bookingID, "hold_id", booking.HoldID, "error", err)
			}
		}
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); err != nil {
		return err
	}

	// Mark as cancelled by owner for tracking
	if err := s.bookingRepo.UpdateCancelledByOwner(ctx, bookingID); err != nil {
		s.logger.Warn("failed to mark booking as cancelled by owner", "booking_id", bookingID, "error", err)
	}

	s.refundBookingPoints(ctx, booking)
	s.refundReferralBonus(ctx, booking)
	s.refundPromoUsage(ctx, booking)
	s.refundCertificateUsage(ctx, booking)
	if booking.Status != domain.BookingPendingOwner {
		s.refundPayment(ctx, booking, true, "")
	}
	s.releaseDepositOnCancel(ctx, booking)
	s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)

	// Owner cancellation penalty: credit 10% to client wallet as compensation
	// Skip penalty when admin cancels — only penalize actual owner/representative
	if role != domain.RoleAdmin {
		s.handleOwnerCancellationPenalty(ctx, booking)
	}

	return nil
}

func (s *bookingService) Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending {
		return domain.ErrInvalidInput
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed); err != nil {
		return err
	}
	s.sendBookingNotification(ctx, booking, domain.NotifBookingConfirmed)
	return nil
}

func (s *bookingService) Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, reason string) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner {
		return domain.ErrInvalidInput
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	// For pending_owner bookings, release payment holds (card + wallet) and standalone wallet hold
	wasRequestBased := booking.Status == domain.BookingPendingOwner
	if wasRequestBased {
		if err := s.paymentSvc.ReleaseHoldPayment(ctx, bookingID); err != nil {
			s.logger.Error("failed to release payment hold on reject", "booking_id", bookingID, "error", err)
		}
		if booking.HoldID != nil && s.walletSvc != nil {
			if err := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); err != nil {
				s.logger.Warn("failed to release booking wallet hold on reject (may already be released by payment hold)", "booking_id", bookingID, "hold_id", booking.HoldID, "error", err)
			}
		}
	}

	booking.Status = domain.BookingRejected
	booking.RejectionReason = reason
	if err := s.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	s.refundBookingPoints(ctx, booking)
	s.refundReferralBonus(ctx, booking)
	s.refundPromoUsage(ctx, booking)
	s.refundCertificateUsage(ctx, booking)
	if !wasRequestBased {
		// Only refund non-hold payments; holds were already released above
		s.refundPayment(ctx, booking, true, "")
	}

	if reason != "" {
		s.sendBookingNotificationWithReason(ctx, booking, domain.NotifBookingRejected, reason)
	} else {
		s.sendBookingNotification(ctx, booking, domain.NotifBookingRejected)
	}
	return nil
}

func (s *bookingService) Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*BookingResult, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.Status != domain.BookingConfirmed {
		return nil, fmt.Errorf("%w: only confirmed bookings can be completed", domain.ErrInvalidInput)
	}

	if time.Now().Before(booking.EndTime) {
		return nil, fmt.Errorf("%w: booking can only be completed after end time", domain.ErrInvalidInput)
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return nil, err
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCompleted); err != nil {
		return nil, err
	}

	// Earn loyalty points and recalculate level
	var earnedPoints int64
	earned, err := s.loyaltySvc.EarnPoints(ctx, booking.UserID, bookingID, booking.TotalPrice)
	if err != nil {
		s.logger.Warn("failed to earn loyalty points", "booking_id", bookingID, "error", err)
	} else {
		earnedPoints = earned
		levelChange, err := s.loyaltySvc.RecalculateLevel(ctx, booking.UserID)
		if err != nil {
			s.logger.Warn("failed to recalculate loyalty level", "booking_id", bookingID, "error", err)
		} else if levelChange != nil && levelChange.Changed {
			s.sendLoyaltyUpgradeNotification(ctx, booking.UserID, levelChange)
		}
	}

	// Credit cashback to wallet based on loyalty level (cashback is on base price per BRD)
	s.creditCashback(ctx, booking.UserID, bookingID, booking.BasePrice)

	// Complete referral if this is the referee's first completed booking
	referralResult, err := s.referralSvc.CompleteReferral(ctx, booking.UserID)
	if err != nil {
		s.logger.Warn("failed to complete referral", "booking_id", bookingID, "user_id", booking.UserID, "error", err)
	} else if referralResult != nil && referralResult.Completed {
		s.sendReferralBonusNotifications(ctx, referralResult)
	}

	// Create escrow to hold funds during claim period before releasing to owner.
	// Only create escrow if a succeeded payment exists -- no funds to escrow otherwise.
	// Use booking.TotalPrice (not payment.Amount) to include session extension charges.
	if s.escrowSvc != nil {
		payment, payErr := s.paymentSvc.GetPaymentByBooking(ctx, booking.UserID, bookingID)
		if payErr != nil {
			s.logger.Warn("no payment found for escrow on complete", "booking_id", bookingID, "error", payErr)
		} else if payment.Status == domain.PaymentSucceeded {
			_, escrowErr := s.escrowSvc.CreateEscrow(ctx, bookingID, booking.TotalPrice, booking.ServiceFeeAmount)
			if escrowErr != nil {
				return nil, fmt.Errorf("failed to create escrow on complete: %w", escrowErr)
			}
		}
	}

	// Record guest visit in CRM
	if s.guestCardSvc != nil {
		bh, bhErr := s.bhRepo.GetByID(ctx, booking.BathhouseID)
		if bhErr != nil {
			s.logger.Warn("failed to get bathhouse for guest card", "booking_id", bookingID, "error", bhErr)
		} else {
			if gcErr := s.guestCardSvc.RecordVisit(ctx, bh.OwnerID, booking.UserID, booking.BathhouseID, booking.TotalPrice); gcErr != nil {
				s.logger.Warn("failed to record guest visit", "booking_id", bookingID, "error", gcErr)
			}
		}
	}

	return &BookingResult{
		Booking:           booking,
		EarnedPoints:      earnedPoints,
		ReferralBonusUsed: booking.ReferralBonusUsed,
	}, nil
}

func (s *bookingService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	return s.bookingRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *bookingService) ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if err := s.access.CanManageBathhouse(ctx, userID, role, bathhouseID); err != nil {
		return nil, err
	}
	return s.bookingRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
}

func (s *bookingService) refundBookingPoints(ctx context.Context, booking *domain.Booking) {
	if booking.PointsSpent <= 0 {
		return
	}
	if err := s.loyaltySvc.RefundPoints(ctx, booking.UserID, booking.PointsSpent, booking.ID); err != nil {
		s.logger.Error("failed to refund loyalty points on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID,
			"points_spent", booking.PointsSpent, "error", err)
	}
}

func (s *bookingService) refundReferralBonus(ctx context.Context, booking *domain.Booking) {
	if booking.ReferralBonusUsed <= 0 {
		return
	}
	if err := s.referralSvc.RefundBalance(ctx, booking.UserID, booking.ReferralBonusUsed, booking.ID); err != nil {
		s.logger.Error("failed to refund referral bonus on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID,
			"referral_bonus_used", booking.ReferralBonusUsed, "error", err)
	}
}

func (s *bookingService) refundPromoUsage(ctx context.Context, booking *domain.Booking) {
	if err := s.promoSvc.RefundUsage(ctx, booking.ID); err != nil {
		s.logger.Error("failed to refund promo usage on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID, "error", err)
	}
}

func (s *bookingService) refundCertificateUsage(ctx context.Context, booking *domain.Booking) {
	if err := s.certSvc.RefundUsage(ctx, booking.ID); err != nil {
		s.logger.Error("failed to refund certificate usage on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID, "error", err)
	}
}

func (s *bookingService) refundPayment(ctx context.Context, booking *domain.Booking, forceFullRefund bool, refundTo string) {
	// Look up the bathhouse's cancellation policy for policy-based refund calculation
	var policy domain.CancellationPolicy
	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err == nil {
		policy = bh.CancellationPolicy
	} else {
		s.logger.Warn("failed to get bathhouse for cancellation policy, using default",
			"booking_id", booking.ID, "bathhouse_id", booking.BathhouseID, "error", err)
	}
	if err := s.paymentSvc.RefundPayment(ctx, booking.ID, forceFullRefund, refundTo, policy); err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return
		}
		s.logger.Error("failed to refund payment on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID, "error", err)
	}
}

func (s *bookingService) releaseDepositOnCancel(ctx context.Context, booking *domain.Booking) {
	if booking.DepositStatus != domain.DepositHeld || s.depositSvc == nil {
		return
	}
	if err := s.depositSvc.ReleaseDeposit(ctx, booking.ID); err != nil {
		s.logger.Error("failed to release deposit on booking cancellation",
			"booking_id", booking.ID, "error", err)
	}
}

func (s *bookingService) sendBookingNotification(ctx context.Context, booking *domain.Booking, notifType domain.NotificationType) {
	var title, body string
	switch notifType {
	case domain.NotifBookingConfirmed:
		title = "Бронирование подтверждено"
		body = fmt.Sprintf("Ваше бронирование на %s подтверждено", booking.StartTime.Format("02.01.2006 15:04"))
	case domain.NotifBookingCancelled:
		title = "Бронирование отменено"
		body = fmt.Sprintf("Ваше бронирование на %s отменено", booking.StartTime.Format("02.01.2006 15:04"))
	case domain.NotifBookingRejected:
		title = "Бронирование отклонено"
		body = fmt.Sprintf("Ваше бронирование на %s отклонено", booking.StartTime.Format("02.01.2006 15:04"))
	case domain.NotifBookingCheckedIn:
		title = "Вы отмечены как прибывший"
		body = fmt.Sprintf("Отмечено прибытие на бронирование %s", booking.StartTime.Format("02.01.2006 15:04"))
	default:
		return
	}

	data := map[string]string{
		"booking_id":   booking.ID.String(),
		"bathhouse_id": booking.BathhouseID.String(),
	}

	if err := s.notifSvc.Send(ctx, booking.UserID, notifType, title, body, data); err != nil {
		s.logger.Warn("failed to send booking notification", "booking_id", booking.ID, "type", notifType, "error", err)
	}
}

func (s *bookingService) Approve(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPendingOwner {
		return fmt.Errorf("%w: only pending_owner bookings can be approved", domain.ErrInvalidInput)
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	// Capture payment hold (card + wallet portions) if exists
	paymentHoldCaptured := false
	if err := s.paymentSvc.CaptureHoldPayment(ctx, bookingID); err != nil {
		if !errors.Is(err, domain.ErrPaymentNotFound) {
			s.logger.Error("failed to capture payment hold on approve", "booking_id", bookingID, "error", err)
			return fmt.Errorf("failed to capture payment hold: %w", err)
		}
		// No payment record exists — booking relies on standalone wallet hold
	} else {
		paymentHoldCaptured = true
	}

	// Capture standalone wallet hold only if no payment hold was captured
	// (payment hold already handles wallet via its own hold; capturing both would double-charge)
	if !paymentHoldCaptured && booking.HoldID != nil && s.walletSvc != nil {
		if _, err := s.walletSvc.CaptureHold(ctx, *booking.HoldID); err != nil {
			s.logger.Warn("failed to capture booking wallet hold on approve", "booking_id", bookingID, "hold_id", booking.HoldID, "error", err)
		}
	} else if paymentHoldCaptured && booking.HoldID != nil && s.walletSvc != nil {
		// Payment hold was captured — release the duplicate booking-level hold
		if err := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); err != nil {
			s.logger.Warn("failed to release duplicate booking wallet hold on approve", "booking_id", bookingID, "hold_id", booking.HoldID, "error", err)
		}
	} else if !paymentHoldCaptured && booking.HoldID == nil {
		// No hold exists — user may have had no wallet or insufficient funds at booking time.
		// Booking will proceed without pre-captured funds; payment must be initiated separately.
		s.logger.Warn("approving booking without payment or wallet hold", "booking_id", bookingID)
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed); err != nil {
		return err
	}

	s.sendBookingNotification(ctx, booking, domain.NotifBookingConfirmed)
	return nil
}

func (s *bookingService) AutoRejectTimedOutRequests(ctx context.Context) (int, error) {
	bookings, err := s.bookingRepo.ListTimedOutRequests(ctx)
	if err != nil {
		return 0, fmt.Errorf("list timed out requests: %w", err)
	}

	rejected := 0
	for _, booking := range bookings {
		b := booking // copy for pointer stability

		// Release payment hold (card + wallet portions)
		if err := s.paymentSvc.ReleaseHoldPayment(ctx, b.ID); err != nil {
			s.logger.Error("failed to release payment hold on auto-reject", "booking_id", b.ID, "error", err)
		}

		// Release standalone wallet hold
		if b.HoldID != nil && s.walletSvc != nil {
			if err := s.walletSvc.ReleaseHold(ctx, *b.HoldID); err != nil {
				s.logger.Warn("failed to release booking wallet hold on auto-reject", "booking_id", b.ID, "hold_id", b.HoldID, "error", err)
			}
		}

		b.Status = domain.BookingRejected
		b.RejectionReason = "Время ожидания ответа истекло"
		if err := s.bookingRepo.Update(ctx, &b); err != nil {
			s.logger.Error("failed to auto-reject booking", "booking_id", b.ID, "error", err)
			continue
		}

		s.refundBookingPoints(ctx, &b)
		s.refundReferralBonus(ctx, &b)
		s.refundPromoUsage(ctx, &b)
		s.refundCertificateUsage(ctx, &b)

		s.sendBookingNotificationWithReason(ctx, &b, domain.NotifBookingRejected, b.RejectionReason)
		rejected++
	}

	return rejected, nil
}

func (s *bookingService) sendBookingRequestNotification(ctx context.Context, booking *domain.Booking, bh *domain.Bathhouse) {
	title := "Новая заявка на бронирование"
	body := fmt.Sprintf("Новая заявка на бронирование от %s", booking.StartTime.Format("02.01.2006 15:04"))
	data := map[string]string{
		"booking_id":   booking.ID.String(),
		"bathhouse_id": booking.BathhouseID.String(),
	}

	// Notify the bathhouse owner
	if err := s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifBookingRequest, title, body, data); err != nil {
		s.logger.Warn("failed to send booking request notification", "booking_id", booking.ID, "owner_id", bh.OwnerID, "error", err)
	}
}

func (s *bookingService) sendBookingNotificationWithReason(ctx context.Context, booking *domain.Booking, notifType domain.NotificationType, reason string) {
	title := "Бронирование отклонено"
	body := fmt.Sprintf("К сожалению, ваша заявка на %s отклонена. Причина: %s", booking.StartTime.Format("02.01.2006 15:04"), reason)
	data := map[string]string{
		"booking_id":   booking.ID.String(),
		"bathhouse_id": booking.BathhouseID.String(),
	}

	if err := s.notifSvc.Send(ctx, booking.UserID, notifType, title, body, data); err != nil {
		s.logger.Warn("failed to send booking notification with reason", "booking_id", booking.ID, "type", notifType, "error", err)
	}
}

func (s *bookingService) handleOwnerCancellationPenalty(ctx context.Context, booking *domain.Booking) {
	if booking.TotalPrice <= 0 {
		return
	}

	// Credit 10% compensation to client wallet, funded by deducting from owner wallet.
	// Only credit client if owner debit succeeds -- platform must not absorb the cost.
	compensationAmount := booking.TotalPrice / 10
	if compensationAmount > 0 && s.walletSvc != nil {
		ownerDebited := false

		// First, deduct from owner's wallet
		bh2, bhErr := s.bhRepo.GetByID(ctx, booking.BathhouseID)
		if bhErr != nil {
			s.logger.Warn("failed to get bathhouse for owner penalty debit", "bathhouse_id", booking.BathhouseID, "error", bhErr)
		} else {
			ownerWallet, owErr := s.walletSvc.GetWallet(ctx, bh2.OwnerID)
			if owErr != nil {
				s.logger.Warn("failed to get owner wallet for penalty debit", "owner_id", bh2.OwnerID, "error", owErr)
			} else {
				bookingID := booking.ID
				_, spendErr := s.walletSvc.Spend(ctx, ownerWallet.ID, compensationAmount, "owner_cancellation_penalty", &bookingID, fmt.Sprintf("Штраф за отмену бронирования %s", bookingID.String()[:8]))
				if spendErr != nil {
					s.logger.Warn("failed to debit owner for cancellation penalty — skipping client credit", "owner_id", bh2.OwnerID, "amount", compensationAmount, "error", spendErr)
				} else {
					ownerDebited = true
				}
			}
		}

		// Credit compensation to client wallet only if owner was successfully debited
		if ownerDebited {
			wallet, err := s.walletSvc.GetWallet(ctx, booking.UserID)
			if err != nil {
				s.logger.Warn("failed to get client wallet for compensation", "user_id", booking.UserID, "error", err)
			} else {
				bookingID := booking.ID
				_, err := s.walletSvc.Refund(ctx, wallet.ID, compensationAmount, "owner_cancellation", &bookingID, "Компенсация за отмену владельцем")
				if err != nil {
					s.logger.Error("failed to credit owner cancellation compensation, refunding owner", "booking_id", booking.ID, "amount", compensationAmount, "error", err)
					// Refund the owner to avoid money loss
					ownerWallet2, ownerErr := s.walletSvc.GetWallet(ctx, bh2.OwnerID)
					if ownerErr == nil {
						_, refundErr := s.walletSvc.Refund(ctx, ownerWallet2.ID, compensationAmount, "owner_penalty_reversal", &bookingID, "Возврат штрафа: не удалось начислить компенсацию клиенту")
						if refundErr != nil {
							s.logger.Error("CRITICAL: failed to refund owner after client credit failure, funds lost", "booking_id", booking.ID, "owner_id", bh2.OwnerID, "amount", compensationAmount, "error", refundErr)
						}
					}
				} else {
					s.logger.Info("credited owner cancellation compensation", "booking_id", booking.ID, "user_id", booking.UserID, "amount", compensationAmount)
					if s.notifSvc != nil {
						body := fmt.Sprintf("Вам начислена компенсация %d₽ за отмену бронирования владельцем", compensationAmount/100)
						_ = s.notifSvc.Send(ctx, booking.UserID, domain.NotifOwnerCancellationCompensation,
							"Компенсация за отмену", body, map[string]string{"booking_id": booking.ID.String()})
					}
				}
			}
		}
	}

	// Get bathhouse to find owner
	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		s.logger.Warn("failed to get bathhouse for penalty check", "bathhouse_id", booking.BathhouseID, "error", err)
		return
	}

	// Count owner cancellations in the last 30 days
	since := time.Now().AddDate(0, 0, -30)
	count, err := s.bookingRepo.CountOwnerCancellations(ctx, bh.OwnerID, since)
	if err != nil {
		s.logger.Warn("failed to count owner cancellations", "owner_id", bh.OwnerID, "error", err)
		return
	}

	if count > 5 {
		// Deactivate ALL owner's bathhouses
		ownerBhIDs, err := s.bhRepo.ListIDsByOwner(ctx, bh.OwnerID)
		if err != nil {
			s.logger.Error("failed to list owner bathhouses for deactivation", "owner_id", bh.OwnerID, "error", err)
			return
		}
		for _, bhID := range ownerBhIDs {
			if err := s.bhRepo.UpdateStatus(ctx, bhID, domain.BathhouseStatusInactive); err != nil {
				s.logger.Error("failed to deactivate bathhouse", "bathhouse_id", bhID, "error", err)
			}
		}
		s.logger.Warn("owner bathhouses auto-deactivated due to excessive cancellations",
			"owner_id", bh.OwnerID, "cancellation_count", count)
		// Notify owner about penalty
		if s.notifSvc != nil {
			body := fmt.Sprintf("Ваши объекты деактивированы из-за %d отмен за 30 дней. Обратитесь в поддержку.", count)
			_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerCancellationPenalty,
				"Объекты деактивированы", body, nil)
		}
	} else if count > 3 {
		// Send warning notification
		if s.notifSvc != nil {
			body := fmt.Sprintf("Внимание: вы отменили %d бронирований за 30 дней. При более чем 5 отменах все ваши объекты будут деактивированы.", count)
			_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerCancellationWarning,
				"Предупреждение об отменах", body, nil)
		}
	}
}

func (s *bookingService) creditCashback(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, basePrice int64) {
	cashbackAmount, err := s.loyaltySvc.CalculateCashback(ctx, userID, basePrice)
	if err != nil {
		s.logger.Warn("failed to calculate cashback", "booking_id", bookingID, "error", err)
		return
	}
	if cashbackAmount <= 0 {
		return
	}

	wallet, err := s.walletSvc.GetWallet(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to get wallet for cashback", "user_id", userID, "booking_id", bookingID, "error", err)
		return
	}

	description := fmt.Sprintf("Кэшбэк за бронирование (%d коп.)", basePrice)
	_, err = s.walletSvc.AddBonus(ctx, wallet.ID, cashbackAmount, domain.WalletTxCashback, nil, description)
	if err != nil {
		s.logger.Warn("failed to credit cashback", "user_id", userID, "booking_id", bookingID, "amount", cashbackAmount, "error", err)
		return
	}

	s.logger.Info("credited cashback", "user_id", userID, "booking_id", bookingID, "amount", cashbackAmount)
}

func (s *bookingService) sendLoyaltyUpgradeNotification(ctx context.Context, userID uuid.UUID, change *LevelChangeResult) {
	title := "Повышение уровня лояльности!"
	body := fmt.Sprintf("Поздравляем! Ваш уровень лояльности повышен: %s → %s",
		change.OldLevel, change.NewLevel)
	data := map[string]string{
		"old_level": string(change.OldLevel),
		"new_level": string(change.NewLevel),
	}
	if err := s.notifSvc.Send(ctx, userID, domain.NotifLoyaltyUpgrade, title, body, data); err != nil {
		s.logger.Warn("failed to send loyalty upgrade notification", "user_id", userID, "error", err)
	}
}

const (
	checkinEarlyWindow  = 15 * time.Minute
	checkinLateWindow   = 30 * time.Minute
	noShowGracePeriod   = 30 * time.Minute
	noShowDisputeWindow = 2 * time.Hour
)
