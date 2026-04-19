package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func (s *bookingService) AdminCancel(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, reason string) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status == domain.BookingCancelled || booking.Status == domain.BookingRejected ||
		booking.Status == domain.BookingForceMajeure || booking.Status == domain.BookingCompleted ||
		booking.Status == domain.BookingNoShow {
		return domain.ErrInvalidInput
	}

	// Release holds if pending_owner
	if booking.Status == domain.BookingPendingOwner {
		if err := s.paymentSvc.ReleaseHoldPayment(ctx, bookingID); err != nil {
			s.logger.Error("failed to release payment hold on admin cancel", "booking_id", bookingID, "error", err)
		}
		if booking.HoldID != nil && s.walletSvc != nil {
			if err := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); err != nil {
				s.logger.Warn("failed to release wallet hold on admin cancel", "booking_id", bookingID, "error", err)
			}
		}
	}

	origStatus := booking.Status

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); err != nil {
		return err
	}

	if reason != "" {
		booking.Status = domain.BookingCancelled
		booking.RejectionReason = reason
		_ = s.bookingRepo.Update(ctx, booking)
	}

	s.refundBookingPoints(ctx, booking)
	s.refundReferralBonus(ctx, booking)
	s.refundPromoUsage(ctx, booking)
	s.refundCertificateUsage(ctx, booking)
	if origStatus != domain.BookingPendingOwner {
		s.refundPayment(ctx, booking, false, "wallet")
	}
	s.releaseDepositOnCancel(ctx, booking)
	s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)

	return nil
}

func (s *bookingService) AdminChangeStatus(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, status domain.BookingStatus, reason string) error {
	if !status.IsValid() {
		return domain.ErrInvalidInput
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	// Prevent no-op transitions
	if booking.Status == status {
		return domain.ErrInvalidInput
	}

	// Block transitions out of terminal states
	switch booking.Status {
	case domain.BookingCompleted, domain.BookingCancelled, domain.BookingRejected,
		domain.BookingForceMajeure, domain.BookingNoShow:
		return domain.ErrInvalidInput
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, status); err != nil {
		return err
	}

	if reason != "" {
		booking.Status = status
		booking.RejectionReason = reason
		_ = s.bookingRepo.Update(ctx, booking)
	}

	return nil
}

func (s *bookingService) AdminListBookings(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
	return s.bookingRepo.ListAll(ctx, filter)
}
