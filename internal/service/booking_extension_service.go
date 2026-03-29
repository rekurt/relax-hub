package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// BookingExtensionService handles the two-party extension request flow.
type BookingExtensionService interface {
	// RequestExtension creates a pending extension request, holds funds, and notifies the owner.
	RequestExtension(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*domain.BookingExtensionRequest, error)
	// ApproveExtension approves a pending extension and applies it to the booking.
	ApproveExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*ExtendResult, error)
	// RejectExtension rejects a pending extension and releases the wallet hold.
	RejectExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error
	// AuthorizeListAccess verifies the caller is the booking client or can manage the bathhouse.
	AuthorizeListAccess(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	// ListByBooking returns all extension requests for a booking.
	ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error)
	// GetByID returns an extension request by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error)
	// ExpireTimedOutRequests expires pending requests past their 30-minute deadline.
	ExpireTimedOutRequests(ctx context.Context) (int, error)
}

type bookingExtensionService struct {
	extReqRepo  repository.ExtensionRequestRepository
	bookingRepo repository.BookingRepository
	bhRepo      repository.BathhouseRepository
	bookingSvc  BookingService
	walletSvc   WalletService
	access      *AccessChecker
	notifSvc    NotificationService
	logger      *logger.Logger
}

func NewBookingExtensionService(
	extReqRepo repository.ExtensionRequestRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	bookingSvc BookingService,
	walletSvc WalletService,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) BookingExtensionService {
	return &bookingExtensionService{
		extReqRepo:  extReqRepo,
		bookingRepo: bookingRepo,
		bhRepo:      bhRepo,
		bookingSvc:  bookingSvc,
		walletSvc:   walletSvc,
		access:      access,
		notifSvc:    notifSvc,
		logger:      log,
	}
}

func (s *bookingExtensionService) RequestExtension(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*domain.BookingExtensionRequest, error) {
	if extraHours < 1 || extraHours > 2 {
		return nil, fmt.Errorf("%w: extra_hours must be 1 or 2", domain.ErrInvalidInput)
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if booking.Status != domain.BookingConfirmed {
		return nil, fmt.Errorf("%w: only confirmed bookings can be extended", domain.ErrInvalidInput)
	}

	if time.Now().After(booking.EndTime) {
		return nil, fmt.Errorf("%w: cannot extend a session that has already ended", domain.ErrInvalidInput)
	}

	// Check for existing pending request
	_, err = s.extReqRepo.GetPendingByBookingID(ctx, bookingID)
	if err == nil {
		return nil, domain.ErrExtensionRequestPending
	}
	if !errors.Is(err, domain.ErrExtensionRequestNotFound) {
		return nil, err
	}

	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	newEndTime := booking.EndTime.Add(time.Duration(extraHours) * time.Hour)

	// Validate within working hours
	if err := validateWithinWorkingHours(bh, booking.StartTime, newEndTime); err != nil {
		return nil, fmt.Errorf("%w: extension exceeds working hours", domain.ErrInvalidInput)
	}

	// Check availability for extension period
	available, err := s.bookingRepo.CheckAvailability(ctx, booking.BathhouseID, booking.EndTime, newEndTime)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, fmt.Errorf("%w: extension period is not available", domain.ErrSlotUnavailable)
	}

	// Check buffer time conflicts
	bufferDuration := time.Duration(bh.BufferMinutes) * time.Minute
	if bufferDuration > 0 {
		overlapping, err := s.bookingRepo.GetOverlapping(ctx, booking.BathhouseID, booking.EndTime, newEndTime.Add(bufferDuration))
		if err != nil {
			return nil, err
		}
		for _, b := range overlapping {
			if b.ID == booking.ID {
				continue
			}
			return nil, fmt.Errorf("%w: extension conflicts with buffer time", domain.ErrSlotUnavailable)
		}
	}

	// Calculate extension price
	extensionPrice := int64(extraHours) * bh.PricePerHour

	// Hold funds in wallet
	var holdID *uuid.UUID
	if s.walletSvc != nil && extensionPrice > 0 {
		wallet, wErr := s.walletSvc.GetWallet(ctx, userID)
		if wErr != nil {
			return nil, fmt.Errorf("%w: wallet required for session extension payment", domain.ErrInvalidInput)
		}
		bookingIDRef := bookingID
		hold, hErr := s.walletSvc.Hold(ctx, wallet.ID, extensionPrice, "booking_extension", &bookingIDRef,
			fmt.Sprintf("Холд: продление сессии %s на %dч", bookingID.String()[:8], extraHours),
			time.Now().Add(domain.ExtensionRequestTimeout+5*time.Minute))
		if hErr != nil {
			return nil, hErr
		}
		holdID = &hold.ID
	}

	now := time.Now()
	req := &domain.BookingExtensionRequest{
		ID:             uuid.New(),
		BookingID:      bookingID,
		UserID:         userID,
		BathhouseID:    booking.BathhouseID,
		Status:         domain.ExtReqPending,
		ExtraHours:     extraHours,
		ExtensionPrice: extensionPrice,
		NewEndTime:     newEndTime,
		HoldID:         holdID,
		CreatedAt:      now,
		ExpiresAt:      now.Add(domain.ExtensionRequestTimeout),
	}

	if err := s.extReqRepo.Create(ctx, req); err != nil {
		// Release hold on create failure
		if holdID != nil {
			if rErr := s.walletSvc.ReleaseHold(ctx, *holdID); rErr != nil {
				s.logger.Error("failed to release hold after create failure", "hold_id", holdID, "error", rErr)
			}
		}
		return nil, err
	}

	// Notify owner
	go s.notifyOwner(context.Background(), bh.OwnerID, booking, req)

	s.logger.Info("booking extension requested",
		"request_id", req.ID,
		"booking_id", bookingID,
		"user_id", userID,
		"extra_hours", extraHours,
		"extension_price", extensionPrice,
	)

	return req, nil
}

func (s *bookingExtensionService) ApproveExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*ExtendResult, error) {
	req, err := s.extReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.Status != domain.ExtReqPending {
		return nil, fmt.Errorf("%w: request is already %s", domain.ErrBookingNotModifiable, req.Status)
	}

	if time.Now().After(req.ExpiresAt) {
		_ = s.extReqRepo.UpdateStatus(ctx, requestID, domain.ExtReqExpired, "")
		s.releaseHoldQuietly(ctx, req)
		return nil, domain.ErrExtensionRequestExpired
	}

	// Verify caller can manage this bathhouse
	if err := s.access.CanManageBathhouse(ctx, ownerID, role, req.BathhouseID); err != nil {
		return nil, err
	}

	// Re-check slot availability (may have changed since the request was created)
	booking, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}

	available, avErr := s.bookingRepo.CheckAvailabilityExcluding(ctx, req.BathhouseID, booking.EndTime, req.NewEndTime, req.BookingID)
	if avErr != nil {
		return nil, fmt.Errorf("re-check availability: %w", avErr)
	}
	if !available {
		s.releaseHoldQuietly(ctx, req)
		_ = s.extReqRepo.UpdateStatus(ctx, requestID, domain.ExtReqExpired, "slot no longer available")
		return nil, domain.ErrSlotUnavailable
	}

	// Update booking first, then capture the hold (safer ordering: if update fails, hold is not lost)
	newTotalPrice := booking.TotalPrice + req.ExtensionPrice

	if err := s.bookingRepo.UpdateEndTime(ctx, req.BookingID, booking.EndTime, req.NewEndTime, newTotalPrice); err != nil {
		return nil, fmt.Errorf("update booking end time: %w", err)
	}

	// Capture the wallet hold after successful booking update
	if req.HoldID != nil {
		if _, cErr := s.walletSvc.CaptureHold(ctx, *req.HoldID); cErr != nil {
			s.logger.Error("failed to capture hold after booking update, manual intervention needed",
				"hold_id", req.HoldID, "request_id", requestID, "error", cErr)
			return nil, fmt.Errorf("capture extension hold: %w", cErr)
		}
	}

	if err := s.extReqRepo.UpdateStatus(ctx, requestID, domain.ExtReqApproved, ""); err != nil {
		s.logger.Warn("failed to update extension request status after approval",
			"request_id", requestID, "error", err)
	}

	booking.EndTime = req.NewEndTime
	booking.TotalPrice = newTotalPrice

	// Notify client
	go s.notifyClient(context.Background(), req.UserID, req, domain.NotifBookingExtensionApproved, "")

	s.logger.Info("booking extension approved",
		"request_id", requestID,
		"booking_id", req.BookingID,
		"approved_by", ownerID,
	)

	return &ExtendResult{
		Booking:        booking,
		ExtensionPrice: req.ExtensionPrice,
		NewEndTime:     req.NewEndTime,
		NewTotalPrice:  newTotalPrice,
	}, nil
}

func (s *bookingExtensionService) RejectExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error {
	req, err := s.extReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.Status != domain.ExtReqPending {
		return fmt.Errorf("%w: request is already %s", domain.ErrBookingNotModifiable, req.Status)
	}

	if err := s.access.CanManageBathhouse(ctx, ownerID, role, req.BathhouseID); err != nil {
		return err
	}

	// Release wallet hold
	s.releaseHoldQuietly(ctx, req)

	if err := s.extReqRepo.UpdateStatus(ctx, requestID, domain.ExtReqRejected, reason); err != nil {
		return err
	}

	go s.notifyClient(context.Background(), req.UserID, req, domain.NotifBookingExtensionRejected, reason)

	s.logger.Info("booking extension rejected",
		"request_id", requestID,
		"booking_id", req.BookingID,
		"rejected_by", ownerID,
		"reason", reason,
	)

	return nil
}

func (s *bookingExtensionService) AuthorizeListAccess(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking.UserID == userID {
		return nil
	}
	if role == domain.RoleAdmin {
		return nil
	}
	return s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID)
}

func (s *bookingExtensionService) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error) {
	return s.extReqRepo.ListByBookingID(ctx, bookingID)
}

func (s *bookingExtensionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error) {
	return s.extReqRepo.GetByID(ctx, id)
}

func (s *bookingExtensionService) ExpireTimedOutRequests(ctx context.Context) (int, error) {
	expired, err := s.extReqRepo.ListExpired(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, req := range expired {
		// Release wallet hold
		s.releaseHoldQuietly(ctx, &req)

		if err := s.extReqRepo.UpdateStatus(ctx, req.ID, domain.ExtReqExpired, "auto-expired after 30min timeout"); err != nil {
			s.logger.Warn("failed to expire extension request", "request_id", req.ID, "error", err)
			continue
		}

		go s.notifyClient(context.Background(), req.UserID, &req, domain.NotifBookingExtensionExpired, "")

		count++
	}

	return count, nil
}

func (s *bookingExtensionService) releaseHoldQuietly(ctx context.Context, req *domain.BookingExtensionRequest) {
	if req.HoldID != nil && s.walletSvc != nil {
		if rErr := s.walletSvc.ReleaseHold(ctx, *req.HoldID); rErr != nil {
			s.logger.Warn("failed to release extension hold", "hold_id", req.HoldID, "request_id", req.ID, "error", rErr)
		}
	}
}

func (s *bookingExtensionService) notifyOwner(ctx context.Context, ownerID uuid.UUID, booking *domain.Booking, req *domain.BookingExtensionRequest) {
	s.notifSvc.Send(ctx, ownerID, domain.NotifBookingExtensionRequested,
		"Запрос на продление сеанса",
		fmt.Sprintf("Гость запросил продление сеанса на %dч до %s", req.ExtraHours, req.NewEndTime.Format("15:04")),
		map[string]string{
			"booking_id": booking.ID.String(),
			"request_id": req.ID.String(),
		},
	)
}

func (s *bookingExtensionService) notifyClient(ctx context.Context, clientID uuid.UUID, req *domain.BookingExtensionRequest, notifType domain.NotificationType, reason string) {
	var title, body string
	switch notifType {
	case domain.NotifBookingExtensionApproved:
		title = "Продление сеанса одобрено"
		body = fmt.Sprintf("Владелец одобрил продление сеанса на %dч до %s", req.ExtraHours, req.NewEndTime.Format("15:04"))
	case domain.NotifBookingExtensionRejected:
		title = "Продление сеанса отклонено"
		body = "Владелец отклонил ваш запрос на продление сеанса"
		if reason != "" {
			body += ": " + reason
		}
	case domain.NotifBookingExtensionExpired:
		title = "Запрос на продление истёк"
		body = "Владелец не ответил на ваш запрос на продление сеанса в течение 30 минут"
	}

	s.notifSvc.Send(ctx, clientID, notifType, title, body, map[string]string{
		"booking_id": req.BookingID.String(),
		"request_id": req.ID.String(),
	})
}
