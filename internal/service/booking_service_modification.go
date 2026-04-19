package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func (s *bookingService) Modify(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input ModifyBookingInput) (*ModifyBookingResult, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Only the booking owner can modify
	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Only pending or confirmed bookings can be modified
	if booking.Status != domain.BookingPending && booking.Status != domain.BookingConfirmed && booking.Status != domain.BookingPendingOwner {
		return nil, domain.ErrBookingNotModifiable
	}

	// Enforce modification limit
	if booking.ModificationCount >= domain.MaxBookingModifications {
		return nil, domain.ErrBookingModificationLimit
	}

	// Cannot modify a booking that already started
	if time.Now().After(booking.StartTime) {
		return nil, fmt.Errorf("%w: cannot modify a booking that has already started", domain.ErrBookingNotModifiable)
	}

	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	// Validate new time
	if !input.EndTime.After(input.StartTime) {
		return nil, fmt.Errorf("%w: end_time must be after start_time", domain.ErrInvalidInput)
	}

	duration := input.EndTime.Sub(input.StartTime)
	durationHours := int(duration / time.Hour)
	if duration%time.Hour != 0 || durationHours < bh.MinDuration {
		return nil, fmt.Errorf("%w: invalid duration", domain.ErrInvalidInput)
	}

	if input.GuestCount <= 0 || input.GuestCount > bh.MaxGuests {
		return nil, fmt.Errorf("%w: invalid guest count", domain.ErrInvalidInput)
	}

	// Enforce lead time
	leadTime := time.Duration(bh.LeadTimeHours) * time.Hour
	if leadTime < 5*time.Minute {
		leadTime = 5 * time.Minute
	}
	if input.StartTime.Before(time.Now().Add(leadTime)) {
		return nil, fmt.Errorf("%w: start time must be at least %v in the future", domain.ErrInvalidInput, leadTime)
	}

	// Enforce max advance days
	maxAdvanceDays := bh.MaxAdvanceDays
	if maxAdvanceDays <= 0 {
		maxAdvanceDays = 90
	}
	if input.StartTime.After(time.Now().AddDate(0, 0, maxAdvanceDays)) {
		return nil, fmt.Errorf("%w: booking exceeds maximum advance days", domain.ErrInvalidInput)
	}

	// Validate working hours
	if err := validateWithinWorkingHours(bh, input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	// Check availability excluding this booking
	bufferDuration := time.Duration(bh.BufferMinutes) * time.Minute
	checkStart := input.StartTime
	checkEnd := input.EndTime
	if bufferDuration > 0 {
		checkStart = input.StartTime.Add(-bufferDuration)
		checkEnd = input.EndTime.Add(bufferDuration)
	}

	available, err := s.bookingRepo.CheckAvailabilityExcluding(ctx, booking.BathhouseID, checkStart, checkEnd, bookingID)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domain.ErrSlotUnavailable
	}

	// Check for slot blocks
	blocked, err := s.slotBlockRepo.HasOverlapping(ctx, booking.BathhouseID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrSlotUnavailable
	}

	// Recalculate price
	totalPrice, priceBreakdown, err := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
		BathhouseID:                booking.BathhouseID,
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

	// Calculate last-minute discount
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
				slotPrice, _, calcErr := s.pricingSvc.CalculateFullPrice(ctx, PriceCalculationInput{
					BathhouseID:  booking.BathhouseID,
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

	// Calculate service fee — resolve region from bathhouse city
	modFeeRegion := "*"
	if s.cityRepo != nil {
		if modCity, cityErr := s.cityRepo.GetByID(ctx, bh.CityID); cityErr == nil && modCity.Region != "" {
			modFeeRegion = modCity.Region
		}
	}
	serviceFeeAmount, err := s.serviceFeeSvc.CalculateFee(ctx, priceBreakdown.BasePrice, modFeeRegion, nil)
	if err != nil {
		s.logger.Warn("failed to calculate service fee on modification, defaulting to 0", "error", err)
		serviceFeeAmount = 0
	}
	totalPrice += serviceFeeAmount

	// Calculate add-on totals
	var addOnTotal int64
	if len(input.AddOns) > 0 {
		addOnTotal, _, err = s.addonSvc.CalculateAddOnTotal(ctx, input.AddOns, booking.BathhouseID, durationHours, input.GuestCount)
		if err != nil {
			return nil, err
		}
		totalPrice += addOnTotal
	}

	if totalPrice <= 0 {
		totalPrice = 1
	}

	oldPrice := booking.TotalPrice
	priceDiff := totalPrice - oldPrice

	// Handle payment adjustments for price difference
	if priceDiff < 0 {
		// Price went down — issue partial refund for the difference only
		refundAmount := -priceDiff
		if s.paymentSvc != nil {
			// Use uuid.Nil as admin ID to indicate system-initiated refund (not a manual admin action)
			if err := s.paymentSvc.AdminRefund(ctx, uuid.Nil, bookingID, refundAmount, "booking_modification", ""); err != nil {
				s.logger.Warn("failed to auto-refund price difference on modification",
					"booking_id", bookingID, "refund_amount", refundAmount, "error", err)
				// Continue — the booking is still modified, refund can be handled manually
			}
		}
	}
	// For priceDiff > 0, the client will need to pay the difference via a separate payment initiation

	// Update booking in database
	newModificationCount := booking.ModificationCount + 1
	if err := s.bookingRepo.UpdateModification(ctx, bookingID,
		input.StartTime, input.EndTime, input.GuestCount,
		totalPrice, addOnTotal, priceBreakdown.BasePrice,
		priceBreakdown.LongSessionDiscount, priceBreakdown.ExtraGuestSurcharge,
		lastMinuteDiscount, serviceFeeAmount, newModificationCount,
	); err != nil {
		return nil, err
	}

	// Refresh booking from DB
	updatedBooking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("booking modified",
		"booking_id", bookingID,
		"user_id", userID,
		"modification_count", newModificationCount,
		"old_price", oldPrice,
		"new_price", totalPrice,
		"price_diff", priceDiff,
	)

	return &ModifyBookingResult{
		Booking:   updatedBooking,
		OldPrice:  oldPrice,
		NewPrice:  totalPrice,
		PriceDiff: priceDiff,
	}, nil
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

	// Prevent extending a booking whose end time has already passed
	if time.Now().After(booking.EndTime) {
		return nil, fmt.Errorf("%w: cannot extend a session that has already ended", domain.ErrInvalidInput)
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
	if err := s.bookingRepo.UpdateEndTime(ctx, bookingID, booking.EndTime, newEndTime, newTotalPrice); err != nil {
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

		// Track when response rate first dropped below 30%
		var lowSince *time.Time
		if responseRate < 0.3 {
			if bh.LowResponseRateSince != nil {
				lowSince = bh.LowResponseRateSince // keep existing timestamp
			} else {
				now := time.Now()
				lowSince = &now // start tracking
			}
		}
		// If rate >= 0.3, lowSince stays nil (reset)

		if err := s.bhRepo.UpdateResponseRate(ctx, bh.ID, responseRate, avgResponseMinutes, lowSince); err != nil {
			s.logger.Warn("failed to update response rate", "bathhouse_id", bh.ID, "error", err)
			continue
		}

		updated++

		// Enforcement: force to instant mode after 60 consecutive days below 30%
		if responseRate < 0.3 && lowSince != nil {
			daysBelowThreshold := time.Since(*lowSince).Hours() / 24
			if daysBelowThreshold >= 60 {
				s.logger.Warn("forcing bathhouse to instant mode due to prolonged low response rate",
					"bathhouse_id", bh.ID, "response_rate", responseRate, "low_since", lowSince)
				bh.BookingMode = domain.BookingModeInstant
				if err := s.bhRepo.Update(ctx, &bh); err != nil {
					s.logger.Warn("failed to force instant mode", "bathhouse_id", bh.ID, "error", err)
				} else if s.notifSvc != nil {
					body := fmt.Sprintf("Ваш процент ответов %.0f%% более 60 дней. Режим бронирования принудительно переключён на мгновенный.", responseRate*100)
					_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerResponseRateWarning,
						"Режим бронирования изменён", body, nil)
				}
				continue
			}

			s.logger.Warn("bathhouse response rate critically low",
				"bathhouse_id", bh.ID, "response_rate", responseRate, "days_below", int(daysBelowThreshold))
			if s.notifSvc != nil {
				body := fmt.Sprintf("Ваш процент ответов %.0f%%. При сохранении низкого показателя режим бронирования будет изменён на мгновенный.", responseRate*100)
				_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerResponseRateWarning,
					"Критически низкий процент ответов", body, nil)
			}
		} else if responseRate < 0.5 {
			if s.notifSvc != nil {
				body := fmt.Sprintf("Ваш процент ответов %.0f%%, рекомендуем отвечать быстрее", responseRate*100)
				_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifOwnerResponseRateWarning,
					"Низкий процент ответов", body, nil)
			}
		}
	}

	return updated, nil
}
