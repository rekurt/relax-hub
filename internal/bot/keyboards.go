package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

// Callback data prefixes
const (
	cbSearch    = "s"   // s:<slug>:<page>
	cbView      = "v"   // v:<bathhouse_id>
	cbBook      = "bk"  // bk:<bathhouse_id>
	cbDate      = "dt"  // dt:<YYYYMMDD>
	cbSlot      = "tm"  // tm:<index>
	cbGuests    = "gs"  // gs:<count>
	cbConfirm   = "cf"  // cf
	cbCancelBk  = "xb"  // xb:<booking_id>
	cbFavPage   = "fp"  // fp:<page>
	cbBkPage    = "mp"  // mp:<page>
	cbFavToggle = "ft"  // ft:<bathhouse_id>
)

// shortID returns first 8 characters of UUID for compact callback data
func shortID(id uuid.UUID) string {
	return strings.ReplaceAll(id.String(), "-", "")[:8]
}

// buildSearchResultsKeyboard builds inline keyboard for bathhouse search results
func buildSearchResultsKeyboard(items []domain.Bathhouse, citySlug string, page, totalPages int) tgbotapi.InlineKeyboardMarkup {
	// Truncate citySlug to fit within Telegram's 64-byte callback data limit
	// Format: "s:<slug>:<page>" - reserve 10 bytes for prefix, separator, and page number
	const maxSlugLen = 54
	if len(citySlug) > maxSlugLen {
		citySlug = citySlug[:maxSlugLen]
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, bh := range items {
		priceRub := float64(bh.PricePerHour) / 100
		label := fmt.Sprintf("%s - %.0f \u20bd/ч (%.1f\u2b50)", bh.Name, priceRub, bh.Rating)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s", cbView, shortID(bh.ID))),
		))
	}

	// Pagination
	var navRow []tgbotapi.InlineKeyboardButton
	if page > 1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("\u25c0 Назад", fmt.Sprintf("%s:%s:%d", cbSearch, citySlug, page-1)))
	}
	if page < totalPages {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Далее \u25b6", fmt.Sprintf("%s:%s:%d", cbSearch, citySlug, page+1)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildBathhouseDetailKeyboard builds keyboard for bathhouse detail view
func buildBathhouseDetailKeyboard(bh *domain.Bathhouse) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001f4c5 Забронировать", fmt.Sprintf("%s:%s", cbBook, shortID(bh.ID))),
			tgbotapi.NewInlineKeyboardButtonData("\u2764\ufe0f В избранное", fmt.Sprintf("%s:%s", cbFavToggle, shortID(bh.ID))),
		),
	)
}

// buildDatePickerKeyboard builds keyboard for date selection (next 7 days)
func buildDatePickerKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	now := time.Now()

	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, i)
		label := date.Format("02.01 (Mon)")
		data := fmt.Sprintf("%s:%s", cbDate, date.Format("20060102"))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, data),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildTimeSlotsKeyboard builds keyboard for time slot selection
func buildTimeSlotsKeyboard(slots []service.TimeSlot) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, slot := range slots {
		if !slot.Available {
			continue
		}
		priceRub := float64(slot.Price) / 100
		label := fmt.Sprintf("%s-%s (%.0f \u20bd)",
			slot.StartTime.Format("15:04"),
			slot.EndTime.Format("15:04"),
			priceRub,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%d", cbSlot, i)),
		))
	}

	if len(rows) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildGuestCountKeyboard builds keyboard for guest count selection
func buildGuestCountKeyboard(maxGuests int) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton
	for i := 1; i <= maxGuests && i <= 10; i++ {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%s:%d", cbGuests, i),
		))
	}

	// Split into rows of 5
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, buttons[i:end])
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildConfirmBookingKeyboard builds confirmation keyboard
func buildConfirmBookingKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\u2705 Подтвердить", cbConfirm),
			tgbotapi.NewInlineKeyboardButtonData("\u274c Отменить", "cancel_wizard"),
		),
	)
}

// buildMyBookingsKeyboard builds keyboard for bookings list
func buildMyBookingsKeyboard(bookings []domain.Booking, page, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, bk := range bookings {
		label := fmt.Sprintf("%s %s-%s [%s]",
			bk.StartTime.Format("02.01"),
			bk.StartTime.Format("15:04"),
			bk.EndTime.Format("15:04"),
			translateBookingStatus(bk.Status),
		)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s", cbView, shortID(bk.BathhouseID))),
		}
		if bk.Status == domain.BookingPending || bk.Status == domain.BookingConfirmed {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData("\u274c", fmt.Sprintf("%s:%s", cbCancelBk, shortID(bk.ID))))
		}
		rows = append(rows, row)
	}

	// Pagination
	var navRow []tgbotapi.InlineKeyboardButton
	if page > 1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("\u25c0 Назад", fmt.Sprintf("%s:%d", cbBkPage, page-1)))
	}
	if page < totalPages {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Далее \u25b6", fmt.Sprintf("%s:%d", cbBkPage, page+1)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	if len(rows) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildFavoritesKeyboard builds keyboard for favorites list
func buildFavoritesKeyboard(favorites []domain.Favorite, bathhouseNames map[uuid.UUID]string, page, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, fav := range favorites {
		name := bathhouseNames[fav.BathhouseID]
		if name == "" {
			name = "Баня"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(name, fmt.Sprintf("%s:%s", cbView, shortID(fav.BathhouseID))),
		))
	}

	// Pagination
	var navRow []tgbotapi.InlineKeyboardButton
	if page > 1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("\u25c0 Назад", fmt.Sprintf("%s:%d", cbFavPage, page-1)))
	}
	if page < totalPages {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Далее \u25b6", fmt.Sprintf("%s:%d", cbFavPage, page+1)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	if len(rows) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// translateBookingStatus translates booking status to Russian
func translateBookingStatus(status domain.BookingStatus) string {
	switch status {
	case domain.BookingPending:
		return "Ожидает"
	case domain.BookingConfirmed:
		return "Подтверждено"
	case domain.BookingCancelled:
		return "Отменено"
	case domain.BookingRejected:
		return "Отклонено"
	case domain.BookingCompleted:
		return "Завершено"
	default:
		return string(status)
	}
}

// formatBathhouseDetail formats bathhouse info for display
func formatBathhouseDetail(bh *domain.Bathhouse) string {
	priceRub := float64(bh.PricePerHour) / 100

	var amenities []string
	if bh.HasPool {
		amenities = append(amenities, "\U0001f3ca бассейн")
	}
	if bh.HasSauna {
		amenities = append(amenities, "\U0001f9d6 сауна")
	}
	if bh.HasSteamRoom {
		amenities = append(amenities, "\u2668\ufe0f парная")
	}
	if bh.HasHotTub {
		amenities = append(amenities, "\U0001f6c1 джакузи")
	}
	if bh.HasBBQ {
		amenities = append(amenities, "\U0001f356 мангал")
	}
	if bh.HasKaraoke {
		amenities = append(amenities, "\U0001f3a4 караоке")
	}

	text := fmt.Sprintf("*%s*\n\n", escapeMD(bh.Name))
	text += fmt.Sprintf("\U0001f4cd %s\n", escapeMD(bh.Address))
	text += fmt.Sprintf("\U0001f4b0 %s \u20bd/ч\n", escapeMD(fmt.Sprintf("%.0f", priceRub)))
	text += fmt.Sprintf("\u2b50 %s \\(%d отзывов\\)\n", escapeMD(fmt.Sprintf("%.1f", bh.Rating)), bh.ReviewCount)
	text += fmt.Sprintf("\U0001f465 до %d гостей\n", bh.MaxGuests)

	if len(amenities) > 0 {
		text += fmt.Sprintf("\n%s\n", escapeMD(strings.Join(amenities, " | ")))
	}

	if bh.Description != "" {
		text += fmt.Sprintf("\n%s", escapeMD(bh.Description))
	}

	return text
}

// formatBathhouseDetailPlain formats bathhouse info as plain text (for inline results)
func formatBathhouseDetailPlain(bh domain.Bathhouse) string {
	priceRub := float64(bh.PricePerHour) / 100

	var amenities []string
	if bh.HasPool {
		amenities = append(amenities, "бассейн")
	}
	if bh.HasSauna {
		amenities = append(amenities, "сауна")
	}
	if bh.HasSteamRoom {
		amenities = append(amenities, "парная")
	}
	if bh.HasHotTub {
		amenities = append(amenities, "джакузи")
	}
	if bh.HasBBQ {
		amenities = append(amenities, "мангал")
	}
	if bh.HasKaraoke {
		amenities = append(amenities, "караоке")
	}

	text := bh.Name + "\n\n"
	if bh.Address != "" {
		text += "📍 " + bh.Address + "\n"
	}
	text += fmt.Sprintf("💰 %.0f ₽/ч\n", priceRub)
	text += fmt.Sprintf("⭐ %.1f (%d отзывов)\n", bh.Rating, bh.ReviewCount)
	text += fmt.Sprintf("👥 до %d гостей\n", bh.MaxGuests)

	if len(amenities) > 0 {
		text += "\n" + strings.Join(amenities, " | ") + "\n"
	}

	if bh.Description != "" {
		text += "\n" + bh.Description
	}

	return text
}

// escapeMD escapes all MarkdownV2 special characters for Telegram
func escapeMD(s string) string {
	special := []string{"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	r := s
	for _, ch := range special {
		r = strings.ReplaceAll(r, ch, "\\"+ch)
	}
	return r
}
