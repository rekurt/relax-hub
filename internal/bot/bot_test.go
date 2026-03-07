package bot

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortID(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sid := shortID(id)
	assert.Equal(t, "550e8400", sid)
	assert.Len(t, sid, 8)
}

func TestEscapeMD(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello_world", "hello\\_world"},
		{"*bold*", "\\*bold\\*"},
		{"[link]", "\\[link\\]"},
		{"`code`", "\\`code\\`"},
		{"mixed_*[`]", "mixed\\_\\*\\[\\`\\]"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, escapeMD(tt.input))
	}
}

func TestTranslateBookingStatus(t *testing.T) {
	tests := []struct {
		status   domain.BookingStatus
		expected string
	}{
		{domain.BookingPending, "Ожидает"},
		{domain.BookingConfirmed, "Подтверждено"},
		{domain.BookingCancelled, "Отменено"},
		{domain.BookingRejected, "Отклонено"},
		{domain.BookingCompleted, "Завершено"},
		{domain.BookingStatus("unknown"), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, translateBookingStatus(tt.status))
	}
}

func TestBuildSearchResultsKeyboard(t *testing.T) {
	bathhouses := []domain.Bathhouse{
		{
			ID:           uuid.New(),
			Name:         "Баня 1",
			PricePerHour: 500000, // 5000 rub
			Rating:       4.5,
		},
		{
			ID:           uuid.New(),
			Name:         "Баня 2",
			PricePerHour: 300000, // 3000 rub
			Rating:       4.0,
		},
	}

	t.Run("single page", func(t *testing.T) {
		kb := buildSearchResultsKeyboard(bathhouses, "moscow", 1, 1)
		// 2 bathhouse rows, no nav
		assert.Len(t, kb.InlineKeyboard, 2)
	})

	t.Run("first of multiple pages", func(t *testing.T) {
		kb := buildSearchResultsKeyboard(bathhouses, "moscow", 1, 3)
		// 2 bathhouse rows + 1 nav row
		assert.Len(t, kb.InlineKeyboard, 3)
		navRow := kb.InlineKeyboard[2]
		assert.Len(t, navRow, 1) // only "Далее"
		require.NotNil(t, navRow[0].CallbackData)
		assert.Contains(t, *navRow[0].CallbackData, "s:moscow:2")
	})

	t.Run("middle page", func(t *testing.T) {
		kb := buildSearchResultsKeyboard(bathhouses, "moscow", 2, 3)
		assert.Len(t, kb.InlineKeyboard, 3)
		navRow := kb.InlineKeyboard[2]
		assert.Len(t, navRow, 2) // "Назад" + "Далее"
	})

	t.Run("last page", func(t *testing.T) {
		kb := buildSearchResultsKeyboard(bathhouses, "moscow", 3, 3)
		assert.Len(t, kb.InlineKeyboard, 3)
		navRow := kb.InlineKeyboard[2]
		assert.Len(t, navRow, 1) // only "Назад"
		require.NotNil(t, navRow[0].CallbackData)
		assert.Contains(t, *navRow[0].CallbackData, "s:moscow:2")
	})

	t.Run("empty results", func(t *testing.T) {
		kb := buildSearchResultsKeyboard(nil, "moscow", 1, 0)
		assert.Len(t, kb.InlineKeyboard, 0)
	})
}

func TestBuildBathhouseDetailKeyboard(t *testing.T) {
	bh := &domain.Bathhouse{ID: uuid.New()}
	kb := buildBathhouseDetailKeyboard(bh)
	require.Len(t, kb.InlineKeyboard, 1)
	assert.Len(t, kb.InlineKeyboard[0], 2) // Book + Favorite
}

func TestBuildDatePickerKeyboard(t *testing.T) {
	kb := buildDatePickerKeyboard()
	assert.Len(t, kb.InlineKeyboard, 7) // 7 days
	for _, row := range kb.InlineKeyboard {
		assert.Len(t, row, 1)
		assert.NotNil(t, row[0].CallbackData)
		assert.Contains(t, *row[0].CallbackData, cbDate+":")
	}
}

func TestBuildTimeSlotsKeyboard(t *testing.T) {
	now := time.Now()
	slots := []service.TimeSlot{
		{StartTime: now, EndTime: now.Add(time.Hour), Available: true, Price: 500000},
		{StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour), Available: false, Price: 500000},
		{StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour), Available: true, Price: 600000},
	}

	kb := buildTimeSlotsKeyboard(slots)
	// Only 2 available slots shown
	assert.Len(t, kb.InlineKeyboard, 2)
}

func TestBuildTimeSlotsKeyboard_AllUnavailable(t *testing.T) {
	slots := []service.TimeSlot{
		{Available: false},
		{Available: false},
	}
	kb := buildTimeSlotsKeyboard(slots)
	assert.Len(t, kb.InlineKeyboard, 0)
}

func TestBuildGuestCountKeyboard(t *testing.T) {
	t.Run("small max", func(t *testing.T) {
		kb := buildGuestCountKeyboard(3)
		assert.Len(t, kb.InlineKeyboard, 1) // one row of 3
		assert.Len(t, kb.InlineKeyboard[0], 3)
	})

	t.Run("large max capped at 10", func(t *testing.T) {
		kb := buildGuestCountKeyboard(20)
		assert.Len(t, kb.InlineKeyboard, 2) // 5 + 5
		assert.Len(t, kb.InlineKeyboard[0], 5)
		assert.Len(t, kb.InlineKeyboard[1], 5)
	})

	t.Run("exactly 5", func(t *testing.T) {
		kb := buildGuestCountKeyboard(5)
		assert.Len(t, kb.InlineKeyboard, 1)
		assert.Len(t, kb.InlineKeyboard[0], 5)
	})
}

func TestBuildConfirmBookingKeyboard(t *testing.T) {
	kb := buildConfirmBookingKeyboard()
	require.Len(t, kb.InlineKeyboard, 1)
	assert.Len(t, kb.InlineKeyboard[0], 2) // Confirm + Cancel
}

func TestBuildMyBookingsKeyboard(t *testing.T) {
	now := time.Now()
	bookings := []domain.Booking{
		{
			ID:          uuid.New(),
			BathhouseID: uuid.New(),
			StartTime:   now,
			EndTime:     now.Add(2 * time.Hour),
			Status:      domain.BookingPending,
		},
		{
			ID:          uuid.New(),
			BathhouseID: uuid.New(),
			StartTime:   now.Add(24 * time.Hour),
			EndTime:     now.Add(26 * time.Hour),
			Status:      domain.BookingCompleted,
		},
	}

	t.Run("with pagination", func(t *testing.T) {
		kb := buildMyBookingsKeyboard(bookings, 1, 3)
		// 2 booking rows + 1 nav row
		assert.Len(t, kb.InlineKeyboard, 3)

		// First booking (pending) should have cancel button
		assert.Len(t, kb.InlineKeyboard[0], 2) // view + cancel

		// Second booking (completed) should NOT have cancel button
		assert.Len(t, kb.InlineKeyboard[1], 1) // view only
	})

	t.Run("empty", func(t *testing.T) {
		kb := buildMyBookingsKeyboard(nil, 1, 0)
		assert.Len(t, kb.InlineKeyboard, 0)
	})
}

func TestBuildFavoritesKeyboard(t *testing.T) {
	bhID1 := uuid.New()
	bhID2 := uuid.New()
	favorites := []domain.Favorite{
		{BathhouseID: bhID1},
		{BathhouseID: bhID2},
	}
	names := map[uuid.UUID]string{
		bhID1: "Баня Люкс",
		bhID2: "Русская Баня",
	}

	t.Run("with names", func(t *testing.T) {
		kb := buildFavoritesKeyboard(favorites, names, 1, 1)
		assert.Len(t, kb.InlineKeyboard, 2)
		assert.Contains(t, kb.InlineKeyboard[0][0].Text, "Баня Люкс")
		assert.Contains(t, kb.InlineKeyboard[1][0].Text, "Русская Баня")
	})

	t.Run("missing name falls back", func(t *testing.T) {
		kb := buildFavoritesKeyboard(favorites, map[uuid.UUID]string{}, 1, 1)
		assert.Len(t, kb.InlineKeyboard, 2)
		assert.Equal(t, "Баня", kb.InlineKeyboard[0][0].Text)
	})

	t.Run("empty", func(t *testing.T) {
		kb := buildFavoritesKeyboard(nil, nil, 1, 0)
		assert.Len(t, kb.InlineKeyboard, 0)
	})
}

func TestFormatBathhouseDetail(t *testing.T) {
	bh := &domain.Bathhouse{
		Name:         "Русская Баня",
		Address:      "ул. Ленина, 1",
		PricePerHour: 500000, // 5000 rub
		Rating:       4.5,
		ReviewCount:  10,
		MaxGuests:    8,
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: false,
		HasBBQ:       true,
		Description:  "Лучшая баня в городе",
	}

	text := formatBathhouseDetail(bh)
	assert.Contains(t, text, "Русская Баня")
	assert.Contains(t, text, "ул. Ленина, 1")
	assert.Contains(t, text, "5000")
	assert.Contains(t, text, "4.5")
	assert.Contains(t, text, "10 отзывов")
	assert.Contains(t, text, "8 гостей")
	assert.Contains(t, text, "бассейн")
	assert.Contains(t, text, "сауна")
	assert.NotContains(t, text, "парная") // HasSteamRoom = false
	assert.Contains(t, text, "мангал")
	assert.Contains(t, text, "Лучшая баня в городе")
}

func TestFormatBathhouseDetail_NoAmenities(t *testing.T) {
	bh := &domain.Bathhouse{
		Name:         "Простая Баня",
		PricePerHour: 200000,
	}

	text := formatBathhouseDetail(bh)
	assert.Contains(t, text, "Простая Баня")
	assert.NotContains(t, text, "бассейн")
}

func TestIDCaching(t *testing.T) {
	b := &Bot{
		idCache: make(map[string]uuid.UUID),
	}

	id := uuid.New()
	sid := shortID(id)

	// Not cached yet
	_, ok := b.resolveID(sid)
	assert.False(t, ok)

	// Cache and resolve
	b.cacheID(sid, id)
	resolved, ok := b.resolveID(sid)
	assert.True(t, ok)
	assert.Equal(t, id, resolved)
}

func TestWizardState(t *testing.T) {
	b := &Bot{
		wizards: make(map[int64]*BookingWizardState),
	}

	chatID := int64(12345)

	// No wizard initially
	assert.Nil(t, b.getWizard(chatID))

	// Set wizard
	b.wizardsMu.Lock()
	b.wizards[chatID] = &BookingWizardState{
		BathhouseID: uuid.New(),
		Step:        "date",
	}
	b.wizardsMu.Unlock()

	wizard := b.getWizard(chatID)
	require.NotNil(t, wizard)
	assert.Equal(t, "date", wizard.Step)

	// Clear wizard
	b.clearWizard(chatID)
	assert.Nil(t, b.getWizard(chatID))
}

func TestCallbackDataParsing(t *testing.T) {
	// Test that callback data format is parseable
	tests := []struct {
		data     string
		prefix   string
		hasExtra bool
	}{
		{"s:moscow:1", cbSearch, true},
		{"v:550e8400", cbView, true},
		{"bk:550e8400", cbBook, true},
		{"dt:20260308", cbDate, true},
		{"tm:3", cbSlot, true},
		{"gs:5", cbGuests, true},
		{"cf", cbConfirm, false},
		{"xb:550e8400", cbCancelBk, true},
		{"fp:2", cbFavPage, true},
		{"mp:1", cbBkPage, true},
		{"ft:550e8400", cbFavToggle, true},
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			parts := splitCallback(tt.data)
			assert.Equal(t, tt.prefix, parts[0])
			if tt.hasExtra {
				assert.Greater(t, len(parts), 1)
			}
		})
	}
}

// splitCallback mimics the callback parsing logic in handleCallbackQuery
func splitCallback(data string) []string {
	parts := make([]string, 0, 3)
	idx := 0
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			parts = append(parts, data[idx:i])
			idx = i + 1
			if len(parts) == 2 {
				// Remaining data is the last part
				parts = append(parts, data[idx:])
				return parts
			}
		}
	}
	parts = append(parts, data[idx:])
	return parts
}

func TestFormatBathhouseDetailPlain(t *testing.T) {
	bh := domain.Bathhouse{
		Name:         "Русская Баня",
		Address:      "ул. Ленина, 1",
		PricePerHour: 500000,
		Rating:       4.5,
		ReviewCount:  10,
		MaxGuests:    8,
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: false,
		HasBBQ:       true,
		Description:  "Лучшая баня в городе",
	}

	text := formatBathhouseDetailPlain(bh)
	assert.Contains(t, text, "Русская Баня")
	assert.Contains(t, text, "ул. Ленина, 1")
	assert.Contains(t, text, "5000")
	assert.Contains(t, text, "4.5")
	assert.Contains(t, text, "10 отзывов")
	assert.Contains(t, text, "8 гостей")
	assert.Contains(t, text, "бассейн")
	assert.Contains(t, text, "сауна")
	assert.NotContains(t, text, "парная")
	assert.Contains(t, text, "мангал")
	assert.Contains(t, text, "Лучшая баня в городе")
	// Should NOT contain markdown escapes (plain text)
	assert.NotContains(t, text, "\\*")
	assert.NotContains(t, text, "\\_")
}

func TestFormatBathhouseDetailPlain_NoAmenities(t *testing.T) {
	bh := domain.Bathhouse{
		Name:         "Простая Баня",
		PricePerHour: 200000,
	}

	text := formatBathhouseDetailPlain(bh)
	assert.Contains(t, text, "Простая Баня")
	assert.NotContains(t, text, "бассейн")
	assert.NotContains(t, text, "сауна")
}

func TestFormatBathhouseDetailPlain_NoAddress(t *testing.T) {
	bh := domain.Bathhouse{
		Name:         "Баня",
		PricePerHour: 100000,
	}

	text := formatBathhouseDetailPlain(bh)
	assert.NotContains(t, text, "📍")
}

func TestNoopTelegramSender(t *testing.T) {
	sender := notification.NewNoopTelegramSender()
	err := sender.Send(t.Context(), 12345, "Title", "Body")
	assert.NoError(t, err)
}
