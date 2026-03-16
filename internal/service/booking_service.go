package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const cancelDeadline = 2 * time.Hour

type CreateBookingInput struct {
	BathhouseID      uuid.UUID
	StartTime        time.Time
	EndTime          time.Time
	GuestCount       int
	Comment          string
	UsePoints        int64  // Optional: loyalty points to spend (reduces total price)
	UseReferralBonus int64  // Optional: referral bonus to spend (reduces total price)
	PromoCode        string // Optional: promo code to apply for a discount
	CertificateCode  string // Optional: gift certificate code to apply
}

type BookingResult struct {
	Booking           *domain.Booking
	EarnedPoints      int64 // Points earned (only on complete)
	LoyaltyDiscount   int64 // Discount from loyalty level in kopecks
	PointsSpent       int64 // Points spent on this booking
	ReferralBonusUsed int64 // Referral bonus used on this booking
	OriginalPrice     int64 // Price before promo code discount
	PromoDiscount     int64 // Discount from promo code in kopecks
	CertificateDiscount int64 // Discount from gift certificate in kopecks
}

type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Available bool      `json:"available"`
	Price     int64     `json:"price"` // Price in kopecks for this hour slot
}

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*BookingResult, error)
	Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*BookingResult, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error)
}

type bookingService struct {
	bookingRepo   repository.BookingRepository
	bhRepo        repository.BathhouseRepository
	slotBlockRepo repository.SlotBlockRepository
	pricingSvc    PricingService
	loyaltySvc    LoyaltyService
	referralSvc   ReferralService
	promoSvc      PromoService
	certSvc       CertificateService
	paymentSvc    PaymentService
	access        *AccessChecker
	notifSvc      NotificationService
	logger        *logger.Logger
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	slotBlockRepo repository.SlotBlockRepository,
	pricingSvc PricingService,
	loyaltySvc LoyaltyService,
	referralSvc ReferralService,
	promoSvc PromoService,
	certSvc CertificateService,
	paymentSvc PaymentService,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) BookingService {
	return &bookingService{
		bookingRepo:   bookingRepo,
		bhRepo:        bhRepo,
		slotBlockRepo: slotBlockRepo,
		pricingSvc:    pricingSvc,
		loyaltySvc:    loyaltySvc,
		referralSvc:   referralSvc,
		promoSvc:      promoSvc,
		certSvc:       certSvc,
		paymentSvc:    paymentSvc,
		access:        access,
		notifSvc:      notifSvc,
		logger:        log,
	}
}

func (s *bookingService) Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*BookingResult, error) {
	bh, err := s.bhRepo.GetByID(ctx, input.BathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	if input.StartTime.Before(time.Now().Add(5 * time.Minute)) {
		return nil, fmt.Errorf("%w: start time must be at least 5 minutes in the future", domain.ErrInvalidInput)
	}

	if input.GuestCount <= 0 || input.GuestCount > bh.MaxGuests {
		return nil, domain.ErrInvalidInput
	}

	duration := input.EndTime.Sub(input.StartTime)
	durationHours := int(duration / time.Hour)
	if duration%time.Hour != 0 || durationHours < bh.MinDuration {
		return nil, domain.ErrInvalidInput
	}

	// Validate price is positive and won't overflow
	if bh.PricePerHour <= 0 {
		return nil, fmt.Errorf("%w: invalid bathhouse price", domain.ErrInvalidInput)
	}
	if durationHours > 0 && bh.PricePerHour > math.MaxInt64/int64(durationHours) {
		return nil, fmt.Errorf("%w: price calculation overflow", domain.ErrInvalidInput)
	}

	// Validate booking falls within working hours
	if err := validateWithinWorkingHours(bh, input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	available, err := s.bookingRepo.CheckAvailability(ctx, input.BathhouseID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domain.ErrSlotUnavailable
	}

	// Check for slot blocks (external calendar events, manual blocks)
	blocked, err := s.slotBlockRepo.HasOverlapping(ctx, input.BathhouseID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrSlotUnavailable
	}

	// Calculate price using pricing service (which applies any dynamic pricing rules)
	totalPrice, err := s.pricingSvc.CalculatePrice(ctx, input.BathhouseID, bh.PricePerHour, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}

	// Apply loyalty discount
	var loyaltyDiscount int64
	discount, err := s.loyaltySvc.GetDiscount(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to get loyalty discount", "user_id", userID, "error", err)
	} else if discount > 0 {
		loyaltyDiscount = totalPrice * int64(discount) / 100
		totalPrice -= loyaltyDiscount
	}

	// Validate and calculate promo code discount (before points/referral)
	var promoDiscount int64
	originalPriceBeforePromo := totalPrice
	if input.PromoCode != "" {
		_, promoDiscount, err = s.promoSvc.Validate(ctx, input.PromoCode, input.BathhouseID, totalPrice)
		if err != nil {
			return nil, err
		}
		totalPrice -= promoDiscount
		if totalPrice <= 0 {
			totalPrice = 1
		}
	}

	// Apply gift certificate discount (after promo, before points/referral)
	var certificateDiscount int64
	var certificateForApply *domain.GiftCertificate
	if input.CertificateCode != "" {
		cert, err := s.certSvc.GetBalance(ctx, input.CertificateCode)
		if err != nil {
			return nil, err
		}
		if cert.Balance <= 0 || !cert.IsUsable() {
			return nil, domain.ErrCertificateInsufficientBalance
		}
		certificateDiscount = cert.Balance
		if certificateDiscount > totalPrice-1 {
			certificateDiscount = totalPrice - 1 // keep at least 1 kopeck
		}
		totalPrice -= certificateDiscount
		certificateForApply = cert
	}

	// Validate points before applying
	var pointsSpent int64
	if input.UsePoints > 0 {
		if input.UsePoints > totalPrice {
			return nil, fmt.Errorf("%w: points exceed total price", domain.ErrInvalidInput)
		}
		pointsSpent = input.UsePoints
		totalPrice -= input.UsePoints
	}

	// Apply referral bonus
	var referralBonusUsed int64
	if input.UseReferralBonus > 0 {
		if input.UseReferralBonus > totalPrice {
			return nil, fmt.Errorf("%w: referral bonus exceeds total price", domain.ErrInvalidInput)
		}
		// Verify user has sufficient referral balance before creating booking
		balance, err := s.referralSvc.GetBalance(ctx, userID)
		if err != nil {
			return nil, err
		}
		if balance.Balance < input.UseReferralBonus {
			return nil, domain.ErrInsufficientReferralBalance
		}
		referralBonusUsed = input.UseReferralBonus
		totalPrice -= input.UseReferralBonus
	}

	if totalPrice <= 0 {
		totalPrice = 1 // Minimum price 1 kopeck
	}

	bookingID := uuid.New()
	now := time.Now()
	booking := &domain.Booking{
		ID:                bookingID,
		UserID:            userID,
		BathhouseID:       input.BathhouseID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		GuestCount:        input.GuestCount,
		TotalPrice:        totalPrice,
		PointsSpent:       pointsSpent,
		ReferralBonusUsed: referralBonusUsed,
		Status:            domain.BookingPending,
		Comment:           input.Comment,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := booking.Validate(); err != nil {
		return nil, err
	}

	// Create booking first so loyalty_transactions FK on booking_id is valid
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	// Spend loyalty points after booking exists in DB
	if pointsSpent > 0 {
		if err := s.loyaltySvc.SpendPoints(ctx, userID, pointsSpent, bookingID); err != nil {
			// Roll back the booking since points couldn't be spent
			if delErr := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); delErr != nil {
				s.logger.Error("failed to cancel booking after loyalty spend failure",
					"booking_id", bookingID, "spend_error", err, "cancel_error", delErr)
			}
			return nil, err
		}
	}

	// Spend referral bonus after booking exists in DB
	if referralBonusUsed > 0 {
		if err := s.referralSvc.UseBalance(ctx, userID, referralBonusUsed, bookingID); err != nil {
			// Refund loyalty points if they were spent
			if pointsSpent > 0 {
				if refundErr := s.loyaltySvc.RefundPoints(ctx, userID, pointsSpent, bookingID); refundErr != nil {
					s.logger.Error("failed to refund loyalty points after referral spend failure",
						"booking_id", bookingID, "spend_error", err, "refund_error", refundErr)
				}
			}
			// Roll back the booking since referral bonus couldn't be spent
			if delErr := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); delErr != nil {
				s.logger.Error("failed to cancel booking after referral spend failure",
					"booking_id", bookingID, "spend_error", err, "cancel_error", delErr)
			}
			return nil, err
		}
	}

	// Apply promo code after booking exists in DB
	if input.PromoCode != "" && promoDiscount > 0 {
		if _, err := s.promoSvc.Apply(ctx, userID, input.PromoCode, bookingID, input.BathhouseID, originalPriceBeforePromo); err != nil {
			// Refund loyalty points if they were spent
			if pointsSpent > 0 {
				if refundErr := s.loyaltySvc.RefundPoints(ctx, userID, pointsSpent, bookingID); refundErr != nil {
					s.logger.Error("failed to refund loyalty points after promo apply failure",
						"booking_id", bookingID, "apply_error", err, "refund_error", refundErr)
				}
			}
			// Refund referral bonus if it was used
			if referralBonusUsed > 0 {
				if refundErr := s.referralSvc.RefundBalance(ctx, userID, referralBonusUsed, bookingID); refundErr != nil {
					s.logger.Error("failed to refund referral bonus after promo apply failure",
						"booking_id", bookingID, "apply_error", err, "refund_error", refundErr)
				}
			}
			// Roll back the booking
			if delErr := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); delErr != nil {
				s.logger.Error("failed to cancel booking after promo apply failure",
					"booking_id", bookingID, "apply_error", err, "cancel_error", delErr)
			}
			return nil, err
		}
	}

	// Apply gift certificate after booking exists in DB
	if certificateForApply != nil && certificateDiscount > 0 {
		if err := s.certSvc.Apply(ctx, certificateForApply.ID, bookingID, certificateDiscount); err != nil {
			// Refund promo usage
			if input.PromoCode != "" && promoDiscount > 0 {
				if refundErr := s.promoSvc.RefundUsage(ctx, bookingID); refundErr != nil {
					s.logger.Error("failed to refund promo after certificate apply failure",
						"booking_id", bookingID, "apply_error", err, "refund_error", refundErr)
				}
			}
			// Refund loyalty points
			if pointsSpent > 0 {
				if refundErr := s.loyaltySvc.RefundPoints(ctx, userID, pointsSpent, bookingID); refundErr != nil {
					s.logger.Error("failed to refund loyalty points after certificate apply failure",
						"booking_id", bookingID, "apply_error", err, "refund_error", refundErr)
				}
			}
			// Refund referral bonus
			if referralBonusUsed > 0 {
				if refundErr := s.referralSvc.RefundBalance(ctx, userID, referralBonusUsed, bookingID); refundErr != nil {
					s.logger.Error("failed to refund referral bonus after certificate apply failure",
						"booking_id", bookingID, "apply_error", err, "refund_error", refundErr)
				}
			}
			// Roll back the booking
			if delErr := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); delErr != nil {
				s.logger.Error("failed to cancel booking after certificate apply failure",
					"booking_id", bookingID, "apply_error", err, "cancel_error", delErr)
			}
			return nil, err
		}
	}

	return &BookingResult{
		Booking:             booking,
		LoyaltyDiscount:     loyaltyDiscount,
		PointsSpent:         pointsSpent,
		ReferralBonusUsed:   referralBonusUsed,
		OriginalPrice:       originalPriceBeforePromo,
		PromoDiscount:       promoDiscount,
		CertificateDiscount: certificateDiscount,
	}, nil
}

func (s *bookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidInput
	}

	// Client cancels their own booking (at least 2 hours before start)
	if role == domain.RoleClient {
		if booking.UserID != userID {
			return domain.ErrForbidden
		}
		if time.Until(booking.StartTime) < cancelDeadline {
			return domain.ErrBookingCancelLate
		}
		if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); err != nil {
			return err
		}
		s.refundBookingPoints(ctx, booking)
		s.refundReferralBonus(ctx, booking)
		s.refundPromoUsage(ctx, booking)
		s.refundPayment(ctx, booking, false)
		s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)
		return nil
	}

	// Owner/representative cancels booking for their bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); err != nil {
		return err
	}
	s.refundBookingPoints(ctx, booking)
	s.refundReferralBonus(ctx, booking)
	s.refundPromoUsage(ctx, booking)
	s.refundPayment(ctx, booking, true)
	s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)
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

func (s *bookingService) Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
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

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingRejected); err != nil {
		return err
	}
	s.refundBookingPoints(ctx, booking)
	s.refundReferralBonus(ctx, booking)
	s.refundPromoUsage(ctx, booking)
	s.refundPayment(ctx, booking, true)
	s.sendBookingNotification(ctx, booking, domain.NotifBookingRejected)
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

	// Complete referral if this is the referee's first completed booking
	referralResult, err := s.referralSvc.CompleteReferral(ctx, booking.UserID)
	if err != nil {
		s.logger.Warn("failed to complete referral", "booking_id", bookingID, "user_id", booking.UserID, "error", err)
	} else if referralResult != nil && referralResult.Completed {
		s.sendReferralBonusNotifications(ctx, referralResult)
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

func (s *bookingService) GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error) {
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	dayOfWeek := toDayOfWeek(date.Weekday())

	var wh *domain.WorkingHours
	for i := range bh.WorkingHours {
		if bh.WorkingHours[i].DayOfWeek == dayOfWeek {
			wh = &bh.WorkingHours[i]
			break
		}
	}

	if wh == nil {
		return nil, nil // closed on this day
	}

	openHour, openMin, err := parseTime(wh.OpenTime)
	if err != nil {
		return nil, fmt.Errorf("invalid open time: %w", err)
	}
	closeHour, closeMin, err := parseTime(wh.CloseTime)
	if err != nil {
		return nil, fmt.Errorf("invalid close time: %w", err)
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), openHour, openMin, 0, 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), closeHour, closeMin, 0, 0, date.Location())

	// Handle overnight working hours (e.g., 18:00 - 06:00)
	if !dayEnd.After(dayStart) {
		dayEnd = dayEnd.Add(24 * time.Hour)
	}

	overlapping, err := s.bookingRepo.GetOverlapping(ctx, bathhouseID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	// Get slot blocks for the day
	slotBlocks, err := s.slotBlockRepo.GetOverlapping(ctx, bathhouseID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var slots []TimeSlot
	for t := dayStart; t.Before(dayEnd); t = t.Add(time.Hour) {
		slotEnd := t.Add(time.Hour)
		if slotEnd.After(dayEnd) {
			break
		}
		// Skip slots that have already passed
		if slotEnd.Before(now) {
			continue
		}
		avail := true
		for _, b := range overlapping {
			if t.Before(b.EndTime) && slotEnd.After(b.StartTime) {
				avail = false
				break
			}
		}
		// Check slot blocks
		if avail {
			for _, sb := range slotBlocks {
				if t.Before(sb.EndTime) && slotEnd.After(sb.StartTime) {
					avail = false
					break
				}
			}
		}

		// Calculate price for this hour slot using pricing service
		slotPrice, err := s.pricingSvc.CalculatePrice(ctx, bathhouseID, bh.PricePerHour, t, slotEnd)
		if err != nil {
			return nil, err
		}

		slots = append(slots, TimeSlot{
			StartTime: t,
			EndTime:   slotEnd,
			Available: avail,
			Price:     slotPrice,
		})
	}

	return slots, nil
}

func parseTime(s string) (int, int, error) {
	var h, m int
	n, _ := fmt.Sscanf(s, "%d:%d", &h, &m)
	if n != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %q", s)
	}
	// Validate time ranges (hours: 0-23, minutes: 0-59)
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid time values: hour=%d, minute=%d", h, m)
	}
	return h, m, nil
}

func toDayOfWeek(wd time.Weekday) int {
	if wd == time.Sunday {
		return 6
	}
	return int(wd) - 1
}

func validateWithinWorkingHours(bh *domain.Bathhouse, startTime, endTime time.Time) error {
	dayOfWeek := toDayOfWeek(startTime.Weekday())

	var wh *domain.WorkingHours
	for i := range bh.WorkingHours {
		if bh.WorkingHours[i].DayOfWeek == dayOfWeek {
			wh = &bh.WorkingHours[i]
			break
		}
	}

	if wh == nil {
		return fmt.Errorf("%w: bathhouse is closed on this day", domain.ErrInvalidInput)
	}

	openH, openM, err := parseTime(wh.OpenTime)
	if err != nil {
		return fmt.Errorf("invalid open time: %w", err)
	}
	closeH, closeM, err := parseTime(wh.CloseTime)
	if err != nil {
		return fmt.Errorf("invalid close time: %w", err)
	}

	loc := startTime.Location()
	dayOpen := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), openH, openM, 0, 0, loc)
	dayClose := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), closeH, closeM, 0, 0, loc)

	// Handle overnight working hours (e.g., 18:00 - 06:00)
	overnight := false
	if !dayClose.After(dayOpen) {
		dayClose = dayClose.Add(24 * time.Hour)
		overnight = true
	}

	if startTime.Before(dayOpen) || endTime.After(dayClose) {
		return fmt.Errorf("%w: booking must be within working hours (%s-%s)", domain.ErrInvalidInput, wh.OpenTime, wh.CloseTime)
	}

	// For multi-day bookings, validate end date doesn't exceed the end day's closing hours.
	// Skip this check for overnight schedules since dayClose already covers the next day.
	if endTime.Day() != startTime.Day() && !overnight {
		endDayOfWeek := toDayOfWeek(endTime.Weekday())
		var endWh *domain.WorkingHours
		for i := range bh.WorkingHours {
			if bh.WorkingHours[i].DayOfWeek == endDayOfWeek {
				endWh = &bh.WorkingHours[i]
				break
			}
		}

		if endWh == nil {
			return fmt.Errorf("%w: bathhouse is closed on end day", domain.ErrInvalidInput)
		}

		endOpenH, endOpenM, err := parseTime(endWh.OpenTime)
		if err != nil {
			return fmt.Errorf("invalid open time: %w", err)
		}
		endCloseH, endCloseM, err := parseTime(endWh.CloseTime)
		if err != nil {
			return fmt.Errorf("invalid close time: %w", err)
		}

		endDayOpen := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endOpenH, endOpenM, 0, 0, loc)
		endDayClose := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endCloseH, endCloseM, 0, 0, loc)

		// Handle overnight hours on end day
		if !endDayClose.After(endDayOpen) {
			endDayClose = endDayClose.Add(24 * time.Hour)
		}

		if endTime.After(endDayClose) {
			return fmt.Errorf("%w: booking must be within working hours (%s-%s)", domain.ErrInvalidInput, endWh.OpenTime, endWh.CloseTime)
		}
	}

	return nil
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

func (s *bookingService) refundPayment(ctx context.Context, booking *domain.Booking, forceFullRefund bool) {
	if err := s.paymentSvc.RefundPayment(ctx, booking.ID, forceFullRefund); err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return
		}
		s.logger.Error("failed to refund payment on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID, "error", err)
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
