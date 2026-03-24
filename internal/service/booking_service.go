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
	UsePoints        int64            // Optional: loyalty points to spend (reduces total price)
	UseReferralBonus int64            // Optional: referral bonus to spend (reduces total price)
	PromoCode        string           // Optional: promo code to apply for a discount
	CertificateCode  string           // Optional: gift certificate code to apply
	AddOns           []AddOnSelection // Optional: add-ons to include in booking
}

type BookingResult struct {
	Booking             *domain.Booking
	AddOns              []domain.BookingAddOn // Add-ons included in the booking
	EarnedPoints        int64                 // Points earned (only on complete)
	LoyaltyDiscount     int64                 // Discount from loyalty level in kopecks
	PointsSpent         int64                 // Points spent on this booking
	ReferralBonusUsed   int64                 // Referral bonus used on this booking
	OriginalPrice       int64                 // Price before promo code discount
	PromoDiscount       int64                 // Discount from promo code in kopecks
	CertificateDiscount int64                 // Discount from gift certificate in kopecks
	ServiceFeeAmount    int64                 // Platform service fee in kopecks
	BasePrice           int64                 // Base price after dynamic rules
	LongSessionDiscount int64                 // Long session discount amount
	ExtraGuestSurcharge int64                 // Extra guest surcharge amount
	LastMinuteDiscount  int64                 // Last-minute discount amount
	IsHolidayPrice      bool                  // Whether holiday pricing was applied
	HolidayName         string                // Holiday name if applicable
	HolidayMultiplier   float64               // Holiday multiplier used
}

type ExtendResult struct {
	Booking        *domain.Booking
	ExtensionPrice int64
	NewEndTime     time.Time
	NewTotalPrice  int64
}

type TimeSlot struct {
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Available     bool      `json:"available"`
	Price         int64     `json:"price"`                        // Price in kopecks for this hour slot
	IsLastMinute  bool      `json:"is_last_minute,omitempty"`     // Whether last-minute discount is applied
	OriginalPrice int64     `json:"original_price,omitempty"`     // Original price before last-minute discount
}

// RebookData contains pre-filled booking parameters extracted from a historical booking.
type RebookData struct {
	BathhouseID   uuid.UUID     `json:"bathhouse_id"`
	DurationHours int           `json:"duration_hours"`
	TimeFrom      string        `json:"time_from"`
	TimeTo        string        `json:"time_to"`
	GuestCount    int           `json:"guest_count"`
	AddOns        []RebookAddOn `json:"addons,omitempty"`
}

// RebookAddOn represents an add-on from the original booking for re-booking.
type RebookAddOn struct {
	AddOnID  uuid.UUID `json:"addon_id"`
	Quantity int       `json:"quantity"`
}

// UpcomingBookingInfo contains booking data enriched with bathhouse info for reminders.
type UpcomingBookingInfo struct {
	Booking       domain.Booking
	BathhouseName string
	Address       string
	Latitude      float64
	Longitude     float64
	OwnerID       uuid.UUID
}

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*BookingResult, error)
	Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, refundTo string) error
	Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, reason string) error
	Approve(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*BookingResult, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error)
	AutoRejectTimedOutRequests(ctx context.Context) (int, error)
	CheckIn(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	CheckOut(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	MarkNoShows(ctx context.Context) (int, error)
	DisputeNoShow(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, gpsLat, gpsLon float64, comment string) error
	ListUpcomingWithBathhouse(ctx context.Context, from, to time.Time) ([]UpcomingBookingInfo, error)
	Extend(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*ExtendResult, error)
	GetRebookData(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*RebookData, error)
	RecalculateResponseRates(ctx context.Context) (int, error)
}

type bookingService struct {
	bookingRepo   repository.BookingRepository
	bhRepo        repository.BathhouseRepository
	slotBlockRepo repository.SlotBlockRepository
	addonRepo     repository.AddOnRepository
	pricingSvc    PricingService
	addonSvc      AddOnService
	loyaltySvc    LoyaltyService
	referralSvc   ReferralService
	promoSvc      PromoService
	certSvc       CertificateService
	paymentSvc    PaymentService
	serviceFeeSvc ServiceFeeService
	walletSvc     WalletService
	complaintSvc  ComplaintService
	escrowSvc     EscrowService
	access        *AccessChecker
	notifSvc      NotificationService
	logger        *logger.Logger
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	slotBlockRepo repository.SlotBlockRepository,
	addonRepo repository.AddOnRepository,
	pricingSvc PricingService,
	addonSvc AddOnService,
	loyaltySvc LoyaltyService,
	referralSvc ReferralService,
	promoSvc PromoService,
	certSvc CertificateService,
	paymentSvc PaymentService,
	serviceFeeSvc ServiceFeeService,
	walletSvc WalletService,
	complaintSvc ComplaintService,
	escrowSvc EscrowService,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) BookingService {
	return &bookingService{
		bookingRepo:   bookingRepo,
		bhRepo:        bhRepo,
		slotBlockRepo: slotBlockRepo,
		addonRepo:     addonRepo,
		pricingSvc:    pricingSvc,
		addonSvc:      addonSvc,
		loyaltySvc:    loyaltySvc,
		referralSvc:   referralSvc,
		promoSvc:      promoSvc,
		certSvc:       certSvc,
		paymentSvc:    paymentSvc,
		serviceFeeSvc: serviceFeeSvc,
		walletSvc:     walletSvc,
		complaintSvc:  complaintSvc,
		escrowSvc:     escrowSvc,
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

	// Enforce lead time (configurable per bathhouse, minimum 5 minutes fallback)
	leadTime := time.Duration(bh.LeadTimeHours) * time.Hour
	if leadTime < 5*time.Minute {
		leadTime = 5 * time.Minute
	}
	if input.StartTime.Before(time.Now().Add(leadTime)) {
		return nil, fmt.Errorf("%w: start time must be at least %v in the future", domain.ErrInvalidInput, leadTime)
	}

	// Enforce max advance days (default 90 if not set)
	maxAdvanceDays := bh.MaxAdvanceDays
	if maxAdvanceDays <= 0 {
		maxAdvanceDays = 90
	}
	if input.StartTime.After(time.Now().AddDate(0, 0, maxAdvanceDays)) {
		return nil, fmt.Errorf("%w: booking exceeds maximum advance days (%d)", domain.ErrInvalidInput, maxAdvanceDays)
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

	// Check availability with buffer zone: expand the time range by buffer duration
	bufferDuration := time.Duration(bh.BufferMinutes) * time.Minute
	checkStart := input.StartTime
	checkEnd := input.EndTime
	if bufferDuration > 0 {
		checkStart = input.StartTime.Add(-bufferDuration)
		checkEnd = input.EndTime.Add(bufferDuration)
	}

	available, err := s.bookingRepo.CheckAvailability(ctx, input.BathhouseID, checkStart, checkEnd)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domain.ErrSlotUnavailable
	}

	// Check buffer zone conflicts with adjacent bookings
	if bufferDuration > 0 {
		overlapping, err := s.bookingRepo.GetOverlapping(ctx, input.BathhouseID, checkStart, checkEnd)
		if err != nil {
			return nil, err
		}
		for _, b := range overlapping {
			// Skip if the overlapping booking is exactly our time range (already checked above)
			if b.StartTime.Equal(input.StartTime) && b.EndTime.Equal(input.EndTime) {
				continue
			}
			// If a booking ends within buffer before our start, or starts within buffer after our end
			if b.EndTime.After(input.StartTime.Add(-bufferDuration)) && b.StartTime.Before(input.EndTime.Add(bufferDuration)) {
				return nil, fmt.Errorf("%w: conflicts with buffer time between bookings", domain.ErrSlotUnavailable)
			}
		}
	}

	// Check for slot blocks (external calendar events, manual blocks)
	blocked, err := s.slotBlockRepo.HasOverlapping(ctx, input.BathhouseID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrSlotUnavailable
	}

	// Calculate price using pricing service (which applies dynamic pricing, long session discount, extra guest surcharge)
	totalPrice, priceBreakdown, err := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
		BathhouseID:                input.BathhouseID,
		BasePrice:                  bh.PricePerHour,
		StartTime:                  input.StartTime,
		EndTime:                    input.EndTime,
		GuestCount:                 input.GuestCount,
		BaseCapacity:               bh.BaseCapacity,
		ExtraGuestSurcharge:        bh.ExtraGuestSurcharge,
		LongSessionThresholdHours:  bh.LongSessionThresholdHours,
		LongSessionDiscountPercent: bh.LongSessionDiscountPercent,
	})
	if err != nil {
		return nil, err
	}

	// Calculate last-minute discount per slot
	var lastMinuteDiscount int64
	if bh.LastMinuteEnabled && bh.LastMinuteDiscountPercent > 0 {
		now := time.Now()
		slotStart := input.StartTime
		for slotStart.Before(input.EndTime) {
			slotEnd := slotStart.Add(time.Hour)
			if slotEnd.After(input.EndTime) {
				slotEnd = input.EndTime
			}
			hoursUntilStart := slotStart.Sub(now).Hours()
			if hoursUntilStart >= 0 && hoursUntilStart <= float64(bh.LastMinuteHoursThreshold) {
				// Calculate slot price using full pricing (includes holiday multiplier) for accurate per-slot discount
				slotPrice, _, calcErr := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
					BathhouseID:  input.BathhouseID,
					BasePrice:    bh.PricePerHour,
					StartTime:    slotStart,
					EndTime:      slotEnd,
					GuestCount:   1,
					BaseCapacity: 1,
				})
				if calcErr == nil {
					lastMinuteDiscount += slotPrice * int64(bh.LastMinuteDiscountPercent) / 100
				}
			}
			slotStart = slotEnd
		}
		if lastMinuteDiscount > 0 {
			totalPrice -= lastMinuteDiscount
			if totalPrice <= 0 {
				totalPrice = 1
			}
		}
	}

	// Calculate service fee on base price (before discounts and add-ons)
	var serviceFeeAmount int64
	serviceFeeAmount, err = s.serviceFeeSvc.CalculateFee(ctx, priceBreakdown.BasePrice, "*", nil)
	if err != nil {
		s.logger.Warn("failed to calculate service fee, defaulting to 0", "error", err)
		serviceFeeAmount = 0
	}
	totalPrice += serviceFeeAmount

	// Calculate add-on totals
	var addOnTotal int64
	var addOnLineItems []AddOnLineItem
	if len(input.AddOns) > 0 {
		addOnTotal, addOnLineItems, err = s.addonSvc.CalculateAddOnTotal(ctx, input.AddOns, input.BathhouseID, durationHours, input.GuestCount)
		if err != nil {
			return nil, err
		}
		totalPrice += addOnTotal
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
		ID:                  bookingID,
		UserID:              userID,
		BathhouseID:         input.BathhouseID,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		GuestCount:          input.GuestCount,
		TotalPrice:          totalPrice,
		AddOnTotal:          addOnTotal,
		BasePrice:           priceBreakdown.BasePrice,
		LongSessionDiscount: priceBreakdown.LongSessionDiscount,
		ExtraGuestSurcharge: priceBreakdown.ExtraGuestSurcharge,
		LastMinuteDiscount:  lastMinuteDiscount,
		ServiceFeeAmount:    serviceFeeAmount,
		PointsSpent:         pointsSpent,
		ReferralBonusUsed:   referralBonusUsed,
		Status:              domain.BookingPending,
		Comment:             input.Comment,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// For request-based booking mode, set status to pending_owner and create wallet hold
	if bh.BookingMode == domain.BookingModeRequest {
		booking.Status = domain.BookingPendingOwner
	}

	if err := booking.Validate(); err != nil {
		return nil, err
	}

	// Create booking first so loyalty_transactions FK on booking_id is valid
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	// For request-based bookings, create wallet hold and notify owner
	if bh.BookingMode == domain.BookingModeRequest && s.walletSvc != nil {
		wallet, walletErr := s.walletSvc.GetWallet(ctx, userID)
		if walletErr == nil && wallet != nil {
			holdExpiry := now.Add(time.Duration(bh.RequestTimeout) * time.Hour)
			hold, holdErr := s.walletSvc.Hold(ctx, wallet.ID, totalPrice, "booking", &bookingID, fmt.Sprintf("Hold for booking request %s", bookingID), holdExpiry)
			if holdErr != nil {
				s.logger.Warn("failed to create wallet hold for request booking", "booking_id", bookingID, "error", holdErr)
			} else {
				booking.HoldID = &hold.ID
				if updateErr := s.bookingRepo.Update(ctx, booking); updateErr != nil {
					s.logger.Warn("failed to update booking with hold_id", "booking_id", bookingID, "error", updateErr)
				}
			}
		}
		s.sendBookingRequestNotification(ctx, booking, bh)
	}

	// Store booking add-ons
	var bookingAddOns []domain.BookingAddOn
	for _, item := range addOnLineItems {
		ba := &domain.BookingAddOn{
			ID:         uuid.New(),
			BookingID:  bookingID,
			AddOnID:    item.AddOnID,
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
			CreatedAt:  now,
		}
		if err := s.addonRepo.CreateBookingAddOn(ctx, ba); err != nil {
			s.logger.Error("failed to store booking add-on", "booking_id", bookingID, "addon_id", item.AddOnID, "error", err)
			return nil, fmt.Errorf("store booking add-on: %w", err)
		}
		bookingAddOns = append(bookingAddOns, *ba)
	}

	// Helper to release wallet hold during rollback
	releaseHoldOnRollback := func() {
		if booking.HoldID != nil && s.walletSvc != nil {
			if releaseErr := s.walletSvc.ReleaseHold(ctx, *booking.HoldID); releaseErr != nil {
				s.logger.Error("failed to release wallet hold during booking rollback",
					"booking_id", bookingID, "hold_id", *booking.HoldID, "error", releaseErr)
			}
		}
	}

	// Spend loyalty points after booking exists in DB
	if pointsSpent > 0 {
		if err := s.loyaltySvc.SpendPoints(ctx, userID, pointsSpent, bookingID); err != nil {
			releaseHoldOnRollback()
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
			releaseHoldOnRollback()
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
			releaseHoldOnRollback()
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
			releaseHoldOnRollback()
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
		AddOns:              bookingAddOns,
		LoyaltyDiscount:     loyaltyDiscount,
		PointsSpent:         pointsSpent,
		ReferralBonusUsed:   referralBonusUsed,
		OriginalPrice:       originalPriceBeforePromo,
		PromoDiscount:       promoDiscount,
		CertificateDiscount: certificateDiscount,
		ServiceFeeAmount:    serviceFeeAmount,
		BasePrice:           priceBreakdown.BasePrice,
		LongSessionDiscount: priceBreakdown.LongSessionDiscount,
		ExtraGuestSurcharge: priceBreakdown.ExtraGuestSurcharge,
		LastMinuteDiscount:  lastMinuteDiscount,
		IsHolidayPrice:      priceBreakdown.IsHolidayPrice,
		HolidayName:         priceBreakdown.HolidayName,
		HolidayMultiplier:   priceBreakdown.HolidayMultiplier,
	}, nil
}

func (s *bookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, refundTo string) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingPendingOwner && booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidInput
	}

	// Client cancels their own booking (at least 2 hours before start, unless pending_owner)
	if role == domain.RoleClient {
		if booking.UserID != userID {
			return domain.ErrForbidden
		}
		// pending_owner bookings can be cancelled anytime by client (request not yet approved)
		if booking.Status != domain.BookingPendingOwner && time.Until(booking.StartTime) < cancelDeadline {
			return domain.ErrBookingCancelLate
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
	s.sendBookingNotification(ctx, booking, domain.NotifBookingCancelled)

	// Owner cancellation penalty: credit 10% to client wallet as compensation
	s.handleOwnerCancellationPenalty(ctx, booking)

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
	bufferDuration := time.Duration(bh.BufferMinutes) * time.Minute
	leadTimeCutoff := now.Add(time.Duration(bh.LeadTimeHours) * time.Hour)
	maxAdvDays := bh.MaxAdvanceDays
	if maxAdvDays <= 0 {
		maxAdvDays = 90
	}
	maxAdvanceCutoff := now.AddDate(0, 0, maxAdvDays)

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
		// Enforce lead time: skip slots starting before the minimum lead time
		if t.Before(leadTimeCutoff) {
			continue
		}
		// Enforce max advance days: skip slots beyond the max booking horizon
		if t.After(maxAdvanceCutoff) || t.Equal(maxAdvanceCutoff) {
			continue
		}
		avail := true
		for _, b := range overlapping {
			// Check booking overlap including buffer zone both after and before the booking
			bookingEndWithBuffer := b.EndTime.Add(bufferDuration)
			bookingStartWithBuffer := b.StartTime.Add(-bufferDuration)
			if t.Before(bookingEndWithBuffer) && slotEnd.After(bookingStartWithBuffer) {
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

		// Calculate price for this hour slot using full pricing (includes holiday multiplier)
		slotPrice, _, err := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
			BathhouseID: bathhouseID,
			BasePrice:   bh.PricePerHour,
			StartTime:   t,
			EndTime:     slotEnd,
			GuestCount:  1,
			BaseCapacity: 1,
		})
		if err != nil {
			return nil, err
		}

		slot := TimeSlot{
			StartTime: t,
			EndTime:   slotEnd,
			Available: avail,
			Price:     slotPrice,
		}

		// Apply last-minute discount if enabled and slot starts within threshold
		if bh.LastMinuteEnabled && bh.LastMinuteDiscountPercent > 0 {
			hoursUntilStart := t.Sub(now).Hours()
			if hoursUntilStart >= 0 && hoursUntilStart <= float64(bh.LastMinuteHoursThreshold) {
				slot.OriginalPrice = slotPrice
				slot.Price = slotPrice - (slotPrice * int64(bh.LastMinuteDiscountPercent) / 100)
				slot.IsLastMinute = true
			}
		}

		slots = append(slots, slot)
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

func (s *bookingService) refundCertificateUsage(ctx context.Context, booking *domain.Booking) {
	if err := s.certSvc.RefundUsage(ctx, booking.ID); err != nil {
		s.logger.Error("failed to refund certificate usage on booking cancellation",
			"booking_id", booking.ID, "user_id", booking.UserID, "error", err)
	}
}

func (s *bookingService) refundPayment(ctx context.Context, booking *domain.Booking, forceFullRefund bool, refundTo string) {
	if err := s.paymentSvc.RefundPayment(ctx, booking.ID, forceFullRefund, refundTo); err != nil {
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
		return fmt.Errorf("%w: no payment or wallet hold found for booking", domain.ErrInvalidInput)
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

	// Credit 10% compensation to client wallet
	compensationAmount := booking.TotalPrice / 10
	if compensationAmount > 0 && s.walletSvc != nil {
		wallet, err := s.walletSvc.GetWallet(ctx, booking.UserID)
		if err != nil {
			s.logger.Warn("failed to get client wallet for compensation", "user_id", booking.UserID, "error", err)
		} else {
			bookingID := booking.ID
			_, err := s.walletSvc.Refund(ctx, wallet.ID, compensationAmount, "owner_cancellation", &bookingID, "Компенсация за отмену владельцем")
			if err != nil {
				s.logger.Warn("failed to credit owner cancellation compensation", "booking_id", booking.ID, "amount", compensationAmount, "error", err)
			} else {
				s.logger.Info("credited owner cancellation compensation", "booking_id", booking.ID, "user_id", booking.UserID, "amount", compensationAmount)
				// Notify client about compensation
				if s.notifSvc != nil {
					body := fmt.Sprintf("Вам начислена компенсация %d₽ за отмену бронирования владельцем", compensationAmount/100)
					_ = s.notifSvc.Send(ctx, booking.UserID, domain.NotifOwnerCancellationCompensation,
						"Компенсация за отмену", body, map[string]string{"booking_id": booking.ID.String()})
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
	checkinEarlyWindow = 15 * time.Minute
	checkinLateWindow  = 30 * time.Minute
	noShowGracePeriod  = 30 * time.Minute
	noShowDisputeWindow = 2 * time.Hour
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

	// Complete referral if applicable
	referralResult, err := s.referralSvc.CompleteReferral(ctx, booking.UserID)
	if err != nil {
		s.logger.Warn("failed to complete referral on checkout", "booking_id", bookingID, "error", err)
	} else if referralResult != nil && referralResult.Completed {
		s.sendReferralBonusNotifications(ctx, referralResult)
	}

	// Create escrow to hold funds during claim period before releasing to owner
	if s.escrowSvc != nil {
		_, escrowErr := s.escrowSvc.CreateEscrow(ctx, bookingID, booking.TotalPrice, booking.ServiceFeeAmount)
		if escrowErr != nil {
			s.logger.Warn("failed to create escrow on checkout", "booking_id", bookingID, "error", escrowErr)
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

		// Create escrow so funds flow to owner after claim period
		if s.escrowSvc != nil {
			if _, escrowErr := s.escrowSvc.CreateEscrow(ctx, b.ID, b.TotalPrice, b.ServiceFeeAmount); escrowErr != nil {
				s.logger.Warn("failed to create escrow for no-show booking", "booking_id", b.ID, "error", escrowErr)
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

	description := fmt.Sprintf("GPS: %.6f, %.6f", gpsLat, gpsLon)
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

func (s *bookingService) Extend(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*ExtendResult, error) {
	if extraHours < 1 || extraHours > 2 {
		return nil, fmt.Errorf("%w: extra_hours must be 1 or 2", domain.ErrInvalidInput)
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Only the booking owner can extend
	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Must be confirmed (possibly checked-in, which keeps confirmed status)
	if booking.Status != domain.BookingConfirmed {
		return nil, fmt.Errorf("%w: only confirmed bookings can be extended", domain.ErrInvalidInput)
	}

	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	newEndTime := booking.EndTime.Add(time.Duration(extraHours) * time.Hour)

	// Validate extended time is within working hours
	if err := validateWithinWorkingHours(bh, booking.StartTime, newEndTime); err != nil {
		return nil, fmt.Errorf("%w: extension exceeds working hours", domain.ErrInvalidInput)
	}

	// Check availability for the extension period (from current end to new end)
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

	// Check for slot blocks in the extension period
	blocked, err := s.slotBlockRepo.HasOverlapping(ctx, booking.BathhouseID, booking.EndTime, newEndTime)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, fmt.Errorf("%w: extension period is blocked", domain.ErrSlotUnavailable)
	}

	// Calculate extension price using full pricing (includes holiday multiplier and extra guest surcharge)
	extensionPrice, _, err := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
		BathhouseID:         booking.BathhouseID,
		BasePrice:           bh.PricePerHour,
		StartTime:           booking.EndTime,
		EndTime:             newEndTime,
		GuestCount:          booking.GuestCount,
		BaseCapacity:        bh.BaseCapacity,
		ExtraGuestSurcharge: bh.ExtraGuestSurcharge,
	})
	if err != nil {
		return nil, fmt.Errorf("calculate extension price: %w", err)
	}

	newTotalPrice := booking.TotalPrice + extensionPrice

	// Charge extension price via wallet
	if s.walletSvc != nil && extensionPrice > 0 {
		wallet, wErr := s.walletSvc.GetWallet(ctx, userID)
		if wErr != nil {
			return nil, fmt.Errorf("%w: wallet required for session extension payment", domain.ErrInvalidInput)
		}
		bookingIDRef := bookingID
		_, spendErr := s.walletSvc.Spend(ctx, wallet.ID, extensionPrice, "booking_extension", &bookingIDRef,
			fmt.Sprintf("Продление сессии %s на %dч", bookingID.String()[:8], extraHours))
		if spendErr != nil {
			return nil, spendErr
		}
	}

	// Update booking end time and total price
	if err := s.bookingRepo.UpdateEndTime(ctx, bookingID, newEndTime, newTotalPrice); err != nil {
		// Rollback wallet debit
		if s.walletSvc != nil && extensionPrice > 0 {
			wallet, wErr := s.walletSvc.GetWallet(ctx, userID)
			if wErr == nil {
				bookingIDRef := bookingID
				if _, rErr := s.walletSvc.Refund(ctx, wallet.ID, extensionPrice, "extension_rollback", &bookingIDRef,
					"Возврат: ошибка продления"); rErr != nil {
					s.logger.Error("failed to rollback extension wallet debit", "booking_id", bookingID, "error", rErr)
				}
			}
		}
		return nil, fmt.Errorf("update booking end time: %w", err)
	}

	booking.EndTime = newEndTime
	booking.TotalPrice = newTotalPrice

	// Notify owner about extension
	if ownerBh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID); err == nil {
		ownerBody := fmt.Sprintf("Гость продлил сессию на %dч до %s", extraHours, newEndTime.Format("15:04"))
		data := map[string]string{
			"booking_id":   booking.ID.String(),
			"bathhouse_id": booking.BathhouseID.String(),
		}
		if err := s.notifSvc.Send(ctx, ownerBh.OwnerID, domain.NotifBookingExtendedOwner, "Сессия продлена", ownerBody, data); err != nil {
			s.logger.Warn("failed to send extension notification to owner", "booking_id", bookingID, "error", err)
		}
	}

	// Notify client
	clientBody := fmt.Sprintf("Сессия продлена до %s", newEndTime.Format("15:04"))
	data := map[string]string{
		"booking_id":   booking.ID.String(),
		"bathhouse_id": booking.BathhouseID.String(),
	}
	if err := s.notifSvc.Send(ctx, booking.UserID, domain.NotifBookingExtended, "Сессия продлена", clientBody, data); err != nil {
		s.logger.Warn("failed to send extension notification to client", "booking_id", bookingID, "error", err)
	}

	return &ExtendResult{
		Booking:        booking,
		ExtensionPrice: extensionPrice,
		NewEndTime:     newEndTime,
		NewTotalPrice:  newTotalPrice,
	}, nil
}

func (s *bookingService) GetRebookData(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*RebookData, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Only the booking owner can get rebook data
	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Only completed or cancelled bookings can be rebooked
	if booking.Status != domain.BookingCompleted && booking.Status != domain.BookingCancelled {
		return nil, fmt.Errorf("%w: only completed or cancelled bookings can be rebooked", domain.ErrInvalidInput)
	}

	// Check that the bathhouse still exists and is active
	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, err
	}
	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	durationHours := int(booking.EndTime.Sub(booking.StartTime) / time.Hour)
	timeFrom := booking.StartTime.Format("15:04")
	timeTo := booking.EndTime.Format("15:04")

	// Fetch add-ons from the original booking
	bookingAddOns, err := s.addonRepo.ListByBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	var rebookAddOns []RebookAddOn
	for _, a := range bookingAddOns {
		rebookAddOns = append(rebookAddOns, RebookAddOn{
			AddOnID:  a.AddOnID,
			Quantity: a.Quantity,
		})
	}

	return &RebookData{
		BathhouseID:   booking.BathhouseID,
		DurationHours: durationHours,
		TimeFrom:      timeFrom,
		TimeTo:        timeTo,
		GuestCount:    booking.GuestCount,
		AddOns:        rebookAddOns,
	}, nil
}

func (s *bookingService) RecalculateResponseRates(ctx context.Context) (int, error) {
	bathhouses, err := s.bhRepo.ListRequestModeBathhouses(ctx)
	if err != nil {
		return 0, fmt.Errorf("list request mode bathhouses: %w", err)
	}

	since := time.Now().AddDate(0, 0, -90)
	updated := 0

	for _, bh := range bathhouses {
		totalRequests, respondedInTime, avgResponseMinutes, err := s.bookingRepo.GetResponseStats(ctx, bh.ID, since)
		if err != nil {
			s.logger.Warn("failed to get response stats", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		var responseRate float64
		if totalRequests > 0 {
			responseRate = float64(respondedInTime) / float64(totalRequests)
		} else {
			responseRate = 1.0 // no requests = perfect rate
		}

		if err := s.bhRepo.UpdateResponseRate(ctx, bh.ID, responseRate, avgResponseMinutes); err != nil {
			s.logger.Warn("failed to update response rate", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		updated++

		// Enforcement: warn at <0.5, deactivate at <0.3 for 60+ days
		if responseRate < 0.3 && bh.ResponseRate < 0.3 {
			// Both current and previous rate below 0.3 — assume it's been low for a while
			// In production, we'd track consecutive days, but for now we use the stored rate
			s.logger.Warn("bathhouse response rate critically low, forcing to instant mode",
				"bathhouse_id", bh.ID, "response_rate", responseRate)
			// Force booking_mode to instant
			bh.BookingMode = domain.BookingModeInstant
			if err := s.bhRepo.Update(ctx, &bh); err != nil {
				s.logger.Error("failed to force bathhouse to instant mode", "bathhouse_id", bh.ID, "error", err)
			}
			if s.notifSvc != nil {
				body := fmt.Sprintf("Ваш процент ответов %.0f%%. Режим бронирования изменён на мгновенный.", responseRate*100)
				_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerResponseRateWarning,
					"Низкий процент ответов", body, nil)
			}
		} else if responseRate < 0.5 {
			// Send warning
			if s.notifSvc != nil {
				body := fmt.Sprintf("Ваш процент ответов %.0f%%, рекомендуем отвечать быстрее", responseRate*100)
				_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerResponseRateWarning,
					"Низкий процент ответов", body, nil)
			}
		}
	}

	return updated, nil
}
