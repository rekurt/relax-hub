package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// BookingModificationService handles the two-party modification request flow.
type BookingModificationService interface {
	// RequestModification creates a pending modification request that the owner must approve.
	RequestModification(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input ModifyBookingInput) (*domain.BookingModificationRequest, error)
	// ApproveModification approves a pending modification request and applies the changes.
	ApproveModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*ModifyBookingResult, error)
	// RejectModification rejects a pending modification request.
	RejectModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error
	// ListByBooking returns all modification requests for a booking.
	ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error)
	// GetByID returns a modification request by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error)
	// ExpireTimedOutRequests expires pending requests that have passed their deadline.
	ExpireTimedOutRequests(ctx context.Context) (int, error)
}

type bookingModificationService struct {
	modReqRepo  repository.BookingModificationRequestRepository
	bookingRepo repository.BookingRepository
	bhRepo      repository.BathhouseRepository
	bookingSvc  BookingService
	access      *AccessChecker
	notifSvc    NotificationService
	logger      *logger.Logger
}

func NewBookingModificationService(
	modReqRepo repository.BookingModificationRequestRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	bookingSvc BookingService,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) BookingModificationService {
	return &bookingModificationService{
		modReqRepo:  modReqRepo,
		bookingRepo: bookingRepo,
		bhRepo:      bhRepo,
		bookingSvc:  bookingSvc,
		access:      access,
		notifSvc:    notifSvc,
		logger:      log,
	}
}

func (s *bookingModificationService) RequestModification(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input ModifyBookingInput) (*domain.BookingModificationRequest, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingConfirmed && booking.Status != domain.BookingPendingOwner {
		return nil, domain.ErrBookingNotModifiable
	}

	if booking.ModificationCount >= domain.MaxBookingModifications {
		return nil, domain.ErrBookingModificationLimit
	}

	if time.Now().After(booking.StartTime) {
		return nil, fmt.Errorf("%w: cannot modify a booking that has already started", domain.ErrBookingNotModifiable)
	}

	// Check for existing pending request
	_, err = s.modReqRepo.GetPendingByBookingID(ctx, bookingID)
	if err == nil {
		return nil, domain.ErrModificationRequestPending
	}

	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, err
	}

	// Calculate proposed price (reuse pricing logic via a dry-run)
	proposedPrice, err := s.calculateProposedPrice(ctx, booking, bh, input)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	req := &domain.BookingModificationRequest{
		ID:                 uuid.New(),
		BookingID:          bookingID,
		UserID:             userID,
		BathhouseID:        booking.BathhouseID,
		Status:             domain.ModReqPending,
		OldStartTime:       booking.StartTime,
		OldEndTime:         booking.EndTime,
		OldGuestCount:      booking.GuestCount,
		OldTotalPrice:      booking.TotalPrice,
		ProposedStartTime:  input.StartTime,
		ProposedEndTime:    input.EndTime,
		ProposedGuestCount: input.GuestCount,
		ProposedTotalPrice: proposedPrice,
		CreatedAt:          now,
		ExpiresAt:          now.Add(domain.ModificationRequestTimeout),
	}

	if err := s.modReqRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	// Notify the bathhouse owner
	go s.notifyOwner(context.Background(), bh.OwnerID, booking, req)

	s.logger.Info("booking modification requested",
		"request_id", req.ID,
		"booking_id", bookingID,
		"user_id", userID,
		"proposed_price", proposedPrice,
	)

	return req, nil
}

func (s *bookingModificationService) ApproveModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*ModifyBookingResult, error) {
	req, err := s.modReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.Status != domain.ModReqPending {
		return nil, domain.ErrModificationRequestNotFound
	}

	if time.Now().After(req.ExpiresAt) {
		_ = s.modReqRepo.UpdateStatus(ctx, requestID, domain.ModReqExpired, "")
		return nil, domain.ErrModificationRequestExpired
	}

	// Verify caller can manage this bathhouse
	if err := s.access.CanManageBathhouse(ctx, ownerID, role, req.BathhouseID); err != nil {
		return nil, err
	}

	// Apply the modification via the existing Modify method
	result, err := s.bookingSvc.Modify(ctx, req.UserID, req.BookingID, ModifyBookingInput{
		StartTime:  req.ProposedStartTime,
		EndTime:    req.ProposedEndTime,
		GuestCount: req.ProposedGuestCount,
	})
	if err != nil {
		return nil, fmt.Errorf("apply modification: %w", err)
	}

	if err := s.modReqRepo.UpdateStatus(ctx, requestID, domain.ModReqApproved, ""); err != nil {
		s.logger.Warn("failed to update modification request status after approval",
			"request_id", requestID, "error", err)
	}

	// Notify client
	go s.notifyClient(context.Background(), req.UserID, req, domain.NotifBookingModificationApproved, "")

	s.logger.Info("booking modification approved",
		"request_id", requestID,
		"booking_id", req.BookingID,
		"approved_by", ownerID,
	)

	return result, nil
}

func (s *bookingModificationService) RejectModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error {
	req, err := s.modReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.Status != domain.ModReqPending {
		return domain.ErrModificationRequestNotFound
	}

	if err := s.access.CanManageBathhouse(ctx, ownerID, role, req.BathhouseID); err != nil {
		return err
	}

	if err := s.modReqRepo.UpdateStatus(ctx, requestID, domain.ModReqRejected, reason); err != nil {
		return err
	}

	go s.notifyClient(context.Background(), req.UserID, req, domain.NotifBookingModificationRejected, reason)

	s.logger.Info("booking modification rejected",
		"request_id", requestID,
		"booking_id", req.BookingID,
		"rejected_by", ownerID,
		"reason", reason,
	)

	return nil
}

func (s *bookingModificationService) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error) {
	return s.modReqRepo.ListByBookingID(ctx, bookingID)
}

func (s *bookingModificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error) {
	return s.modReqRepo.GetByID(ctx, id)
}

func (s *bookingModificationService) ExpireTimedOutRequests(ctx context.Context) (int, error) {
	expired, err := s.modReqRepo.ListExpired(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, req := range expired {
		if err := s.modReqRepo.UpdateStatus(ctx, req.ID, domain.ModReqExpired, "auto-expired after 24h timeout"); err != nil {
			s.logger.Warn("failed to expire modification request", "request_id", req.ID, "error", err)
			continue
		}

		go s.notifyClient(context.Background(), req.UserID, &req, domain.NotifBookingModificationExpired, "")

		count++
	}

	return count, nil
}

// calculateProposedPrice estimates the new total price for the proposed changes.
func (s *bookingModificationService) calculateProposedPrice(ctx context.Context, booking *domain.Booking, bh *domain.Bathhouse, input ModifyBookingInput) (int64, error) {
	_ = ctx
	_ = booking

	// Simple price estimation based on duration and base price.
	// The actual Modify() call will do the full recalculation.
	duration := input.EndTime.Sub(input.StartTime)
	durationHours := int(duration / time.Hour)
	if duration%time.Hour != 0 || durationHours < 1 {
		return 0, fmt.Errorf("%w: invalid duration", domain.ErrInvalidInput)
	}

	estimatedPrice := int64(durationHours) * bh.PricePerHour
	return estimatedPrice, nil
}

func (s *bookingModificationService) notifyOwner(ctx context.Context, ownerID uuid.UUID, booking *domain.Booking, req *domain.BookingModificationRequest) {
	_ = req
	s.notifSvc.Send(ctx, ownerID, domain.NotifBookingModificationRequested,
		"Запрос на изменение бронирования",
		fmt.Sprintf("Клиент запросил изменение бронирования на %s", booking.StartTime.Format("02.01.2006 15:04")),
		map[string]string{
			"booking_id": booking.ID.String(),
			"request_id": req.ID.String(),
		},
	)
}

func (s *bookingModificationService) notifyClient(ctx context.Context, clientID uuid.UUID, req *domain.BookingModificationRequest, notifType domain.NotificationType, reason string) {
	var title, body string
	switch notifType {
	case domain.NotifBookingModificationApproved:
		title = "Изменение бронирования одобрено"
		body = "Владелец одобрил ваш запрос на изменение бронирования"
	case domain.NotifBookingModificationRejected:
		title = "Изменение бронирования отклонено"
		body = "Владелец отклонил ваш запрос на изменение бронирования"
		if reason != "" {
			body += ": " + reason
		}
	case domain.NotifBookingModificationExpired:
		title = "Запрос на изменение истёк"
		body = "Владелец не ответил на ваш запрос на изменение бронирования в течение 24 часов"
	}

	s.notifSvc.Send(ctx, clientID, notifType, title, body, map[string]string{
		"booking_id": req.BookingID.String(),
		"request_id": req.ID.String(),
	})
}
