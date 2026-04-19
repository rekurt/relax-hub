package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/service"
)

func (b *Bot) doStartBookingWizard(ctx context.Context, chatID int64, sid string) {
	bhID, ok := b.resolveID(sid)
	if !ok {
		b.sendPlainMessage(chatID, "Баня не найдена. Попробуйте поиск заново.")
		return
	}

	_, ok = b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	bh, err := b.bathhouseService.GetByID(ctx, bhID)
	if err != nil {
		b.logger.Error("failed to get bathhouse", "error", err, "id", bhID)
		b.sendPlainMessage(chatID, "Ошибка при загрузке данных бани.")
		return
	}

	// Store wizard state
	b.wizardsMu.Lock()
	b.wizards[chatID] = &BookingWizardState{
		BathhouseID: bhID,
		Bathhouse:   bh,
		Step:        "date",
		CreatedAt:   time.Now(),
	}
	b.wizardsMu.Unlock()

	kb := buildDatePickerKeyboard()
	b.sendMessageWithKeyboard(chatID, fmt.Sprintf("*Бронирование: %s*\n\nВыберите дату:", escapeMD(bh.Name)), kb)
}

func (b *Bot) doSelectDate(ctx context.Context, chatID int64, dateStr string) {
	b.wizardsMu.RLock()
	wizard := b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.RUnlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}
	bathhouseID := wizard.BathhouseID
	b.wizardsMu.RUnlock()

	date, err := time.ParseInLocation("20060102", dateStr, time.Local)
	if err != nil {
		b.sendPlainMessage(chatID, "Неверная дата.")
		return
	}

	slots, err := b.bookingService.GetAvailableSlots(ctx, bathhouseID, date)
	if err != nil {
		b.logger.Error("failed to get slots", "error", err, "bathhouse_id", bathhouseID)
		b.sendPlainMessage(chatID, "Ошибка при загрузке слотов.")
		return
	}

	availableCount := 0
	for _, s := range slots {
		if s.Available {
			availableCount++
		}
	}

	if availableCount == 0 {
		b.sendPlainMessage(chatID, "На эту дату нет свободных слотов. Выберите другую дату.")
		kb := buildDatePickerKeyboard()
		b.sendMessageWithKeyboard(chatID, "Выберите дату:", kb)
		return
	}

	b.wizardsMu.Lock()
	// Re-check wizard after acquiring write lock to prevent TOCTOU race
	wizard = b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}
	wizard.Date = date
	wizard.Slots = slots
	wizard.Step = "time"
	b.wizardsMu.Unlock()

	kb := buildTimeSlotsKeyboard(slots)
	b.sendMessageWithKeyboard(chatID,
		fmt.Sprintf("Дата: *%s*\nВыберите время:", date.Format("02\\.01\\.2006")),
		kb,
	)
}

func (b *Bot) doSelectSlot(ctx context.Context, chatID int64, slotIdx int) {
	b.wizardsMu.Lock()
	wizard := b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}

	if slotIdx < 0 || slotIdx >= len(wizard.Slots) {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Неверный слот. Попробуйте ещё раз.")
		return
	}

	slot := wizard.Slots[slotIdx]
	if !slot.Available {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Этот слот уже занят. Выберите другой.")
		return
	}

	wizard.StartTime = slot.StartTime
	wizard.EndTime = slot.EndTime
	wizard.Step = "guests"

	maxGuests := 10
	if wizard.Bathhouse != nil {
		maxGuests = wizard.Bathhouse.MaxGuests
	}
	b.wizardsMu.Unlock()

	kb := buildGuestCountKeyboard(maxGuests)
	b.sendMessageWithKeyboard(chatID,
		fmt.Sprintf("Время: *%s\\-%s*\nВыберите количество гостей:",
			slot.StartTime.Format("15:04"),
			slot.EndTime.Format("15:04"),
		),
		kb,
	)
}

func (b *Bot) doSelectGuests(ctx context.Context, chatID int64, count int) {
	b.wizardsMu.Lock()
	wizard := b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}

	if wizard.Bathhouse != nil && count > wizard.Bathhouse.MaxGuests {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, fmt.Sprintf("Максимум гостей: %d", wizard.Bathhouse.MaxGuests))
		return
	}

	wizard.GuestCount = count
	wizard.Step = "promo"
	b.wizardsMu.Unlock()

	kb := buildPromoStepKeyboard()
	b.sendMessageWithKeyboard(chatID, "Есть промокод или подарочный сертификат? Введите код или нажмите Пропустить\\.", kb)
}

func (b *Bot) doConfirmBooking(ctx context.Context, chatID int64) {
	b.wizardsMu.Lock()
	wizard := b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}
	input := service.CreateBookingInput{
		BathhouseID:     wizard.BathhouseID,
		StartTime:       wizard.StartTime,
		EndTime:         wizard.EndTime,
		GuestCount:      wizard.GuestCount,
		PromoCode:       wizard.PromoCode,
		CertificateCode: wizard.CertificateCode,
	}
	// Delete wizard before releasing lock to prevent double-booking from concurrent confirms
	delete(b.wizards, chatID)
	b.wizardsMu.Unlock()

	userID, ok := b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	result, err := b.bookingService.Create(ctx, userID, input)
	if err != nil {
		b.logger.Error("failed to create booking", "error", err, "chat_id", chatID)
		b.sendPlainMessage(chatID, "Ошибка при бронировании. Попробуйте позже.")
		return
	}

	text := fmt.Sprintf("\u2705 *Бронирование создано\\!*\n\n"+
		"ID: `%s`\n"+
		"Статус: %s",
		result.Booking.ID.String(),
		translateBookingStatus(result.Booking.Status),
	)

	// Try to get a payment link
	if b.paymentService != nil {
		payURL, err := b.paymentService.InitiatePayment(ctx, userID, result.Booking.ID, domain.PaymentMethodCard)
		if err == nil && payURL != "" {
			text += fmt.Sprintf("\n\n[Оплатить онлайн](%s)", escapeMD(payURL))
		}
	}

	text += "\n\nВы получите уведомление при подтверждении\\."
	b.sendMessage(chatID, text)
}

// doApplyCode tries to apply a promo code or certificate code
func (b *Bot) doApplyCode(ctx context.Context, chatID int64, code string) {
	if code == "" {
		b.sendPlainMessage(chatID, "Пожалуйста, введите код или нажмите Пропустить.")
		return
	}

	b.wizardsMu.RLock()
	wizard := b.wizards[chatID]
	b.wizardsMu.RUnlock()

	if wizard == nil || wizard.Step != "promo" {
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}

	// Try as promo code first
	if b.promoService != nil {
		bathhouseID := wizard.BathhouseID
		// Estimate amount for validation (price per hour)
		estimatedAmount := int64(0)
		if wizard.Bathhouse != nil {
			hours := wizard.EndTime.Sub(wizard.StartTime).Hours()
			if hours < 1 {
				hours = 1
			}
			estimatedAmount = wizard.Bathhouse.PricePerHour * int64(hours)
		}

		_, _, err := b.promoService.Validate(ctx, code, bathhouseID, estimatedAmount)
		if err == nil {
			b.wizardsMu.Lock()
			w := b.wizards[chatID]
			if w != nil {
				w.PromoCode = code
			}
			b.wizardsMu.Unlock()
			b.sendPlainMessage(chatID, fmt.Sprintf("Промокод %s применён!", code))
			b.doShowConfirmation(ctx, chatID)
			return
		}
	}

	// Try as certificate code
	if b.certificateService != nil {
		cert, err := b.certificateService.GetBalance(ctx, code)
		if err == nil && cert != nil && cert.Balance > 0 {
			b.wizardsMu.Lock()
			w := b.wizards[chatID]
			if w != nil {
				w.CertificateCode = code
			}
			b.wizardsMu.Unlock()
			balanceRub := float64(cert.Balance) / 100
			b.sendPlainMessage(chatID, fmt.Sprintf("Сертификат применён! Баланс: %.0f ₽", balanceRub))
			b.doShowConfirmation(ctx, chatID)
			return
		}
	}

	// Code not found
	kb := buildPromoStepKeyboard()
	b.sendMessageWithKeyboard(chatID, "Код не найден или недействителен\\. Попробуйте другой или нажмите Пропустить\\.", kb)
}

// doSkipPromo skips the promo step and proceeds to confirmation
func (b *Bot) doSkipPromo(ctx context.Context, chatID int64) {
	b.wizardsMu.RLock()
	wizard := b.wizards[chatID]
	b.wizardsMu.RUnlock()

	if wizard == nil {
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}

	b.doShowConfirmation(ctx, chatID)
}

// doShowConfirmation displays booking confirmation with applied discount info
func (b *Bot) doShowConfirmation(_ context.Context, chatID int64) {
	b.wizardsMu.Lock()
	wizard := b.wizards[chatID]
	if wizard == nil {
		b.wizardsMu.Unlock()
		b.sendPlainMessage(chatID, "Сессия бронирования истекла. Начните заново через /book")
		return
	}

	wizard.Step = "confirm"

	bhName := "Баня"
	if wizard.Bathhouse != nil {
		bhName = wizard.Bathhouse.Name
	}

	text := fmt.Sprintf("*Подтверждение бронирования:*\n\n"+
		"\U0001f6c1 %s\n"+
		"\U0001f4c5 %s\n"+
		"\U0001f552 %s \\- %s\n"+
		"\U0001f465 %d гостей",
		escapeMD(bhName),
		wizard.Date.Format("02\\.01\\.2006"),
		wizard.StartTime.Format("15:04"),
		wizard.EndTime.Format("15:04"),
		wizard.GuestCount,
	)

	if wizard.PromoCode != "" {
		text += fmt.Sprintf("\n\U0001f3ab Промокод: %s", escapeMD(wizard.PromoCode))
	}
	if wizard.CertificateCode != "" {
		text += fmt.Sprintf("\n\U0001f381 Сертификат: %s", escapeMD(wizard.CertificateCode))
	}

	text += "\n\nВсё верно?"
	b.wizardsMu.Unlock()

	kb := buildConfirmBookingKeyboard()
	b.sendMessageWithKeyboard(chatID, text, kb)
}

func (b *Bot) doCancelWizard(chatID int64) {
	b.clearWizard(chatID)
	b.sendPlainMessage(chatID, "Бронирование отменено.")
}

