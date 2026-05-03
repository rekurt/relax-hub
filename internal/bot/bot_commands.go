package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func (b *Bot) commandStart(ctx context.Context, chatID int64) {
	text := "\U0001f6c1 *Добро пожаловать в Бани\\!*\n\n" +
		"Я помогу вам найти идеальную баню и забронировать сеанс\\.\n\n" +
		"*Команды:*\n" +
		"/search _город_ \\- поиск бань\n" +
		"/book _id_ \\- забронировать\n" +
		"/mybookings \\- мои бронирования\n" +
		"/favorites \\- избранное\n" +
		"/link _токен_ \\- привязать аккаунт\n" +
		"/help \\- справка"
	b.sendMessage(chatID, text)
}

func (b *Bot) commandHelp(ctx context.Context, chatID int64) {
	text := "*Справка по командам:*\n\n" +
		"/start \\- начать работу с ботом\n" +
		"/search _город_ \\- найти бани в городе\n" +
		"/book _id_ \\- забронировать баню\n" +
		"/mybookings \\- посмотреть мои бронирования\n" +
		"/cancel _id_ \\- отменить бронирование\n" +
		"/favorites \\- моё избранное\n" +
		"/link _токен_ \\- привязать Telegram\\-аккаунт\n" +
		"/help \\- эта справка"
	b.sendMessage(chatID, text)
}

func (b *Bot) commandLink(ctx context.Context, chatID int64, msg *tgbotapi.Message) {
	token := msg.CommandArguments()
	if token == "" {
		b.sendPlainMessage(chatID, "Пожалуйста, укажите токен: /link <токен>")
		return
	}

	b.sendPlainMessage(chatID, "Функция привязки аккаунта находится в разработке. Пожалуйста, используйте веб-интерфейс.")
}

func (b *Bot) commandSearch(ctx context.Context, chatID int64, query string) {
	if query == "" {
		b.sendPlainMessage(chatID, "Укажите город: /search <москва>")
		return
	}

	b.doSearch(ctx, chatID, query, 1)
}

func (b *Bot) commandBook(ctx context.Context, chatID int64, args string) {
	if args == "" {
		b.sendPlainMessage(chatID, "Укажите ID бани: /book <id>\nНайдите баню через /search")
		return
	}

	bhID, err := uuid.Parse(args)
	if err != nil {
		b.sendPlainMessage(chatID, "Неверный формат ID. Используйте /search для поиска бань.")
		return
	}

	sid := shortID(bhID)
	b.cacheID(sid, bhID)
	b.doStartBookingWizard(ctx, chatID, sid)
}

func (b *Bot) commandCancel(ctx context.Context, chatID int64, args string) {
	if args == "" {
		b.sendPlainMessage(chatID, "Укажите ID бронирования: /cancel <id>\nСписок бронирований: /mybookings")
		return
	}

	bkID, err := uuid.Parse(args)
	if err != nil {
		b.sendPlainMessage(chatID, "Неверный формат ID.")
		return
	}

	sid := shortID(bkID)
	b.cacheID(sid, bkID)
	b.doCancelBooking(ctx, chatID, sid)
}

func (b *Bot) commandMyBookings(ctx context.Context, chatID int64, page int) {
	userID, ok := b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	result, err := b.bookingService.ListByUser(ctx, userID, page, 5)
	if err != nil {
		b.logger.Error("failed to list bookings", "error", err, "chat_id", chatID)
		b.sendPlainMessage(chatID, "Ошибка при получении бронирований.")
		return
	}

	if result.TotalCount == 0 {
		b.sendPlainMessage(chatID, "У вас пока нет бронирований.\nИспользуйте /search для поиска бань.")
		return
	}

	// Cache IDs for callback data
	for _, bk := range result.Items {
		b.cacheID(shortID(bk.ID), bk.ID)
		b.cacheID(shortID(bk.BathhouseID), bk.BathhouseID)
	}

	kb := buildMyBookingsKeyboard(result.Items, page, result.TotalPages)
	b.sendMessageWithKeyboard(chatID, fmt.Sprintf("*Мои бронирования* \\(стр\\. %d/%d\\):", page, result.TotalPages), kb)
}

func (b *Bot) commandFavorites(ctx context.Context, chatID int64, page int) {
	userID, ok := b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	result, err := b.favoriteService.List(ctx, userID, page, 5)
	if err != nil {
		b.logger.Error("failed to list favorites", "error", err, "chat_id", chatID)
		b.sendPlainMessage(chatID, "Ошибка при получении избранного.")
		return
	}

	if result.TotalCount == 0 {
		b.sendPlainMessage(chatID, "Список избранного пуст.\nДобавляйте бани через поиск /search")
		return
	}

	// Fetch bathhouse names
	names := make(map[uuid.UUID]string)
	for _, fav := range result.Items {
		b.cacheID(shortID(fav.BathhouseID), fav.BathhouseID)
		bh, err := b.bathhouseService.GetByID(ctx, fav.BathhouseID)
		if err == nil {
			names[fav.BathhouseID] = bh.Name
		}
	}

	kb := buildFavoritesKeyboard(result.Items, names, page, result.TotalPages)
	b.sendMessageWithKeyboard(chatID, fmt.Sprintf("*Избранное* \\(стр\\. %d/%d\\):", page, result.TotalPages), kb)
}

// --- Callback actions ---
