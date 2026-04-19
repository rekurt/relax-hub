package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

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
	SavedCardID      uuid.UUID        // Optional: saved card to use for payment
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
	DepositAmount       int64                 // Security deposit amount (card hold, not charged)
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
	Price         int64     `json:"price"`                    // Price in kopecks for this hour slot
	IsLastMinute  bool      `json:"is_last_minute,omitempty"` // Whether last-minute discount is applied
	OriginalPrice int64     `json:"original_price,omitempty"` // Original price before last-minute discount
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

// ModifyBookingInput contains the fields a client can change when modifying a booking.
type ModifyBookingInput struct {
	StartTime  time.Time
	EndTime    time.Time
	GuestCount int
	AddOns     []AddOnSelection
}

// ModifyBookingResult holds the outcome of a booking modification.
type ModifyBookingResult struct {
	Booking   *domain.Booking
	OldPrice  int64 // previous total price
	NewPrice  int64 // recalculated total price
	PriceDiff int64 // positive = client owes more, negative = refund due
}

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*BookingResult, error)
	Modify(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input ModifyBookingInput) (*ModifyBookingResult, error)
	Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, refundTo string) error
	Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, reason string) error
	Approve(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*BookingResult, error)
	GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*domain.Booking, error)
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
	AdminCancel(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, reason string) error
	AdminChangeStatus(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, status domain.BookingStatus, reason string) error
	AdminListBookings(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error)
}

type bookingService struct {
	bookingRepo   repository.BookingRepository
	bhRepo        repository.BathhouseRepository
	slotBlockRepo repository.SlotBlockRepository
	addonRepo     repository.AddOnRepository
	userRepo      repository.UserRepository
	cityRepo      repository.CityRepository
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
	depositSvc    SecurityDepositService
	guestCardSvc  GuestCardService
	access        *AccessChecker
	notifSvc      NotificationService
	logger        *logger.Logger
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	slotBlockRepo repository.SlotBlockRepository,
	addonRepo repository.AddOnRepository,
	userRepo repository.UserRepository,
	cityRepo repository.CityRepository,
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
	depositSvc SecurityDepositService,
	guestCardSvc GuestCardService,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) BookingService {
	return &bookingService{
		bookingRepo:   bookingRepo,
		bhRepo:        bhRepo,
		slotBlockRepo: slotBlockRepo,
		addonRepo:     addonRepo,
		userRepo:      userRepo,
		cityRepo:      cityRepo,
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
		depositSvc:    depositSvc,
		guestCardSvc:  guestCardSvc,
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
	// Resolve bathhouse region from city for per-region fee configs
	feeRegion := "*"
	if s.cityRepo != nil {
		if city, cityErr := s.cityRepo.GetByID(ctx, bh.CityID); cityErr == nil && city.Region != "" {
			feeRegion = city.Region
		}
	}
	var serviceFeeAmount int64
	serviceFeeAmount, err = s.serviceFeeSvc.CalculateFee(ctx, priceBreakdown.BasePrice, feeRegion, nil)
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
		promo, discount, promoErr := s.promoSvc.Validate(ctx, input.PromoCode, input.BathhouseID, totalPrice)
		if promoErr != nil {
			return nil, promoErr
		}
		// For free_addon promo: verify the target add-on is included and use actual line item total
		if promo.Type == domain.PromoTypeFreeAddon && promo.TargetAddOnID != nil {
			addonIncluded := false
			for _, item := range addOnLineItems {
				if item.AddOnID == *promo.TargetAddOnID {
					addonIncluded = true
					// Use actual computed line item total (accounts for per_hour/per_person multipliers)
					discount = item.TotalPrice
					break
				}
			}
			if !addonIncluded {
				return nil, domain.ErrPromoInvalid
			}
		}
		promoDiscount = discount
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
	// Calculate security deposit amount if bathhouse requires it
	depositAmount := CalculateDepositAmount(priceBreakdown.BasePrice, bh.SecurityDepositPercent)

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
		DepositAmount:       depositAmount,
		DepositStatus:       domain.DepositNone,
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

	// Atomically check availability and create booking in a single transaction
	// with advisory lock to prevent double-booking race conditions (FR-061).
	// The earlier CheckAvailability call serves as a fast pre-check to avoid
	// unnecessary lock acquisition.
	if err := s.bookingRepo.CreateWithAvailabilityCheck(ctx, booking, checkStart, checkEnd); err != nil {
		return nil, err
	}

	// For request-based bookings, create wallet hold and notify owner
	if bh.BookingMode == domain.BookingModeRequest {
		// Wallet hold is required for request-based bookings to guarantee funds.
		// Without a hold, the owner cannot approve the booking (see Approve method).
		if s.walletSvc != nil {
			wallet, walletErr := s.walletSvc.GetWallet(ctx, userID)
			if walletErr != nil || wallet == nil {
				s.logger.Warn("no wallet found for request booking, proceeding without hold", "booking_id", bookingID, "error", walletErr)
			} else {
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
		}
		s.sendBookingRequestNotification(ctx, booking, bh)
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
			releaseHoldOnRollback()
			if cancelErr := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled); cancelErr != nil {
				s.logger.Error("failed to cancel booking after add-on failure",
					"booking_id", bookingID, "error", cancelErr)
			}
			return nil, fmt.Errorf("store booking add-on: %w", err)
		}
		bookingAddOns = append(bookingAddOns, *ba)
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

	// Hold security deposit if required (best-effort: booking proceeds even if deposit hold fails)
	if depositAmount > 0 && s.depositSvc != nil {
		if depositErr := s.depositSvc.HoldDeposit(ctx, bookingID, depositAmount, "card"); depositErr != nil {
			s.logger.Warn("failed to hold security deposit, proceeding without deposit",
				"booking_id", bookingID, "deposit_amount", depositAmount, "error", depositErr)
			booking.DepositAmount = 0
			booking.DepositStatus = domain.DepositNone
			// Persist the zeroed deposit to DB so it matches in-memory state
			if updateErr := s.bookingRepo.UpdateDeposit(ctx, bookingID, 0, domain.DepositNone, ""); updateErr != nil {
				s.logger.Error("failed to reset deposit amount after hold failure",
					"booking_id", bookingID, "error", updateErr)
			}
		} else {
			booking.DepositStatus = domain.DepositHeld
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
		DepositAmount:       booking.DepositAmount,
		IsHolidayPrice:      priceBreakdown.IsHolidayPrice,
		HolidayName:         priceBreakdown.HolidayName,
		HolidayMultiplier:   priceBreakdown.HolidayMultiplier,
	}, nil
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
	leadTime := time.Duration(bh.LeadTimeHours) * time.Hour
	if leadTime < 5*time.Minute {
		leadTime = 5 * time.Minute
	}
	leadTimeCutoff := now.Add(leadTime)
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
			BathhouseID:  bathhouseID,
			BasePrice:    bh.PricePerHour,
			StartTime:    t,
			EndTime:      slotEnd,
			GuestCount:   1,
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
