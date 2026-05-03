package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/geo"
)

func (s *bookingService) CheckIn(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingConfirmed {
		return fmt.Errorf("%w: only confirmed bookings can be checked in", domain.ErrInvalidInput)
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	now := time.Now()
	earliest := booking.StartTime.Add(-checkinEarlyWindow)
	latest := booking.StartTime.Add(checkinLateWindow)

	if now.Before(earliest) {
		return domain.ErrCheckinTooEarly
	}
	if now.After(latest) {
		return domain.ErrCheckinTooLate
	}

	if err := s.bookingRepo.UpdateCheckin(ctx, bookingID, &now); err != nil {
		return err
	}

	s.sendBookingNotification(ctx, booking, domain.NotifBookingCheckedIn)
	return nil
}

func (s *bookingService) CheckOut(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingConfirmed {
		return fmt.Errorf("%w: only confirmed bookings can be checked out", domain.ErrInvalidInput)
	}

	if booking.CheckedInAt == nil {
		return domain.ErrNotCheckedIn
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	now := time.Now()
	if err := s.bookingRepo.UpdateCheckout(ctx, bookingID, &now, domain.BookingCompleted); err != nil {
		return err
	}

	// Earn loyalty points
	earned, err := s.loyaltySvc.EarnPoints(ctx, booking.UserID, bookingID, booking.TotalPrice)
	if err != nil {
		s.logger.Warn("failed to earn loyalty points on checkout", "booking_id", bookingID, "error", err)
	} else if earned > 0 {
		levelChange, err := s.loyaltySvc.RecalculateLevel(ctx, booking.UserID)
		if err != nil {
			s.logger.Warn("failed to recalculate loyalty level on checkout", "booking_id", bookingID, "error", err)
		} else if levelChange != nil && levelChange.Changed {
			s.sendLoyaltyUpgradeNotification(ctx, booking.UserID, levelChange)
		}
	}

	// Credit cashback to wallet based on loyalty level (cashback is on base price per BRD)
	s.creditCashback(ctx, booking.UserID, bookingID, booking.BasePrice)

	// Complete referral if applicable
	referralResult, err := s.referralSvc.CompleteReferral(ctx, booking.UserID)
	if err != nil {
		s.logger.Warn("failed to complete referral on checkout", "booking_id", bookingID, "error", err)
	} else if referralResult != nil && referralResult.Completed {
		s.sendReferralBonusNotifications(ctx, referralResult)
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

	// Create escrow to hold funds during claim period before releasing to owner.
	// Only create escrow if a succeeded payment exists -- no funds to escrow otherwise.
	// Use booking.TotalPrice (not payment.Amount) to include session extension charges.
	if s.escrowSvc != nil {
		payment, payErr := s.paymentSvc.GetPaymentByBooking(ctx, booking.UserID, bookingID)
		if payErr != nil {
			s.logger.Warn("no payment found for escrow on checkout", "booking_id", bookingID, "error", payErr)
		} else if payment.Status == domain.PaymentSucceeded {
			_, escrowErr := s.escrowSvc.CreateEscrow(ctx, bookingID, booking.TotalPrice, booking.ServiceFeeAmount)
			if escrowErr != nil {
				return fmt.Errorf("failed to create escrow on checkout: %w", escrowErr)
			}
		}
	}

	return nil
}

func (s *bookingService) MarkNoShows(ctx context.Context) (int, error) {
	cutoff := time.Now().Add(-noShowGracePeriod)
	bookings, err := s.bookingRepo.ListConfirmedWithoutCheckin(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("list no-show candidates: %w", err)
	}

	marked := 0
	for _, booking := range bookings {
		b := booking
		if err := s.bookingRepo.UpdateStatus(ctx, b.ID, domain.BookingNoShow); err != nil {
			s.logger.Error("failed to mark booking as no-show", "booking_id", b.ID, "error", err)
			continue
		}

		// Notify owner: payment is kept
		bh, err := s.bhRepo.GetByID(ctx, b.BathhouseID)
		if err == nil {
			ownerBody := fmt.Sprintf("Гость не прибыл на бронирование %s %s, оплата сохранена",
				b.StartTime.Format("02.01.2006"), b.StartTime.Format("15:04"))
			if err := s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifBookingNoShowOwner,
				"Гость не прибыл", ownerBody,
				map[string]string{"booking_id": b.ID.String()}); err != nil {
				s.logger.Warn("failed to send no-show owner notification", "booking_id", b.ID, "error", err)
			}
		}

		// Create escrow so funds flow to owner after claim period.
		// Only if payment was actually collected.
		if s.escrowSvc != nil {
			payment, payErr := s.paymentSvc.GetPaymentByBooking(ctx, b.UserID, b.ID)
			if payErr != nil {
				s.logger.Warn("no payment found for escrow on no-show", "booking_id", b.ID, "error", payErr)
			} else if payment.Status == domain.PaymentSucceeded {
				if _, escrowErr := s.escrowSvc.CreateEscrow(ctx, b.ID, b.TotalPrice, b.ServiceFeeAmount); escrowErr != nil {
					s.logger.Warn("failed to create escrow for no-show booking", "booking_id", b.ID, "error", escrowErr)
				}
			}
		}

		// Notify client
		clientBody := "Вы не прибыли на бронирование. Если это ошибка, оспорьте в течение 2 часов"
		if err := s.notifSvc.Send(ctx, b.UserID, domain.NotifBookingNoShow,
			"Неявка на бронирование", clientBody,
			map[string]string{"booking_id": b.ID.String()}); err != nil {
			s.logger.Warn("failed to send no-show client notification", "booking_id", b.ID, "error", err)
		}

		marked++
	}

	return marked, nil
}

func (s *bookingService) DisputeNoShow(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, gpsLat, gpsLon float64, comment string) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.UserID != userID {
		return domain.ErrForbidden
	}

	if booking.Status != domain.BookingNoShow {
		return fmt.Errorf("%w: only no-show bookings can be disputed", domain.ErrInvalidInput)
	}

	if time.Since(booking.UpdatedAt) > noShowDisputeWindow {
		return domain.ErrNoShowDisputeExpired
	}

	// GPS validation: check distance from bathhouse
	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return fmt.Errorf("get bathhouse for GPS validation: %w", err)
	}

	const maxGPSDistanceMeters = 200
	gpsDistanceMeters := geo.HaversineDistance(gpsLat, gpsLon, bh.Latitude, bh.Longitude)
	gpsValid := gpsDistanceMeters <= maxGPSDistanceMeters

	description := fmt.Sprintf("GPS: %.6f, %.6f (расстояние: %d м)", gpsLat, gpsLon, gpsDistanceMeters)
	if !gpsValid {
		description += fmt.Sprintf("\n⚠️ GPS_WEAK_EVIDENCE: клиент находился в %d м от объекта (порог: %d м)", gpsDistanceMeters, maxGPSDistanceMeters)
	}
	if comment != "" {
		description += "\n" + comment
	}

	if s.complaintSvc != nil {
		_, err = s.complaintSvc.Report(ctx, userID, CreateComplaintInput{
			TargetType:  domain.ComplaintTargetNoShowDispute,
			TargetID:    bookingID,
			Reason:      domain.ComplaintReasonOther,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("create no-show dispute: %w", err)
		}
	}

	// Mark escrow as disputed to prevent automatic release to owner during dispute resolution
	if s.escrowSvc != nil {
		if err := s.escrowSvc.MarkDisputedByBookingID(ctx, bookingID); err != nil {
			s.logger.Warn("failed to mark escrow as disputed for no-show dispute", "booking_id", bookingID, "error", err)
		}
	}

	return nil
}

func (s *bookingService) sendReferralBonusNotifications(ctx context.Context, result *ReferralCompletionResult) {
	amountRub := result.BonusAmount / 100
	title := "Реферальный бонус начислен!"

	// Notify referrer
	referrerBody := fmt.Sprintf("Вам начислен реферальный бонус %d руб. за приглашённого друга", amountRub)
	if err := s.notifSvc.Send(ctx, result.ReferrerID, domain.NotifReferralBonus, title, referrerBody, nil); err != nil {
		s.logger.Warn("failed to send referral bonus notification to referrer",
			"referrer_id", result.ReferrerID, "error", err)
	}

	// Notify referee
	refereeBody := fmt.Sprintf("Вам начислен приветственный бонус %d руб. по реферальной программе", amountRub)
	if err := s.notifSvc.Send(ctx, result.RefereeID, domain.NotifReferralBonus, title, refereeBody, nil); err != nil {
		s.logger.Warn("failed to send referral bonus notification to referee",
			"referee_id", result.RefereeID, "error", err)
	}
}

func (s *bookingService) ListUpcomingWithBathhouse(ctx context.Context, from, to time.Time) ([]UpcomingBookingInfo, error) {
	bookings, err := s.bookingRepo.ListUpcoming(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("list upcoming bookings: %w", err)
	}

	bhCache := make(map[uuid.UUID]*domain.Bathhouse)
	var result []UpcomingBookingInfo
	for _, b := range bookings {
		bh, ok := bhCache[b.BathhouseID]
		if !ok {
			bh, err = s.bhRepo.GetByID(ctx, b.BathhouseID)
			if err != nil {
				s.logger.Warn("failed to get bathhouse for reminder", "bathhouse_id", b.BathhouseID, "error", err)
				continue
			}
			bhCache[b.BathhouseID] = bh
		}
		result = append(result, UpcomingBookingInfo{
			Booking:       b,
			BathhouseName: bh.Name,
			Address:       bh.Address,
			Latitude:      bh.Latitude,
			Longitude:     bh.Longitude,
			OwnerID:       bh.OwnerID,
		})
	}
	return result, nil
}
