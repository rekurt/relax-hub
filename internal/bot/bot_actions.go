package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/rekurt/relax-hub/internal/domain"
)

func (b *Bot) doSearch(ctx context.Context, chatID int64, query string, page int) {
	// Try to find city by slug or search by name
	filter := domain.BathhouseFilter{
		SearchQuery: &query,
		Page:        page,
		PageSize:    5,
		SortBy:      "rating",
		SortOrder:   "desc",
	}

	// Try matching city by slug
	city, err := b.cityService.GetBySlug(ctx, strings.ToLower(query))
	if err == nil {
		filter.CityID = &city.ID
		filter.SearchQuery = nil
	}

	status := domain.BathhouseStatusActive
	filter.Status = &status

	result, err := b.bathhouseService.Search(ctx, filter)
	if err != nil {
		b.logger.Error("failed to search bathhouses", "error", err, "query", query)
		b.sendPlainMessage(chatID, "Ошибка при поиске. Попробуйте позже.")
		return
	}

	if result.TotalCount == 0 {
		b.sendPlainMessage(chatID, fmt.Sprintf("По запросу \"%s\" ничего не найдено.", query))
		return
	}

	// Cache IDs
	for _, bh := range result.Items {
		b.cacheID(shortID(bh.ID), bh.ID)
	}

	slug := strings.ReplaceAll(strings.ToLower(query), ":", "")
	kb := buildSearchResultsKeyboard(result.Items, slug, page, result.TotalPages)
	text := fmt.Sprintf("Найдено %d бань:", result.TotalCount)
	b.sendMessageWithKeyboard(chatID, escapeMD(text), kb)
}

func (b *Bot) doViewBathhouse(ctx context.Context, chatID int64, sid string) {
	bhID, ok := b.resolveID(sid)
	if !ok {
		b.sendPlainMessage(chatID, "Баня не найдена. Попробуйте поиск заново.")
		return
	}

	bh, err := b.bathhouseService.GetByID(ctx, bhID)
	if err != nil {
		b.logger.Error("failed to get bathhouse", "error", err, "id", bhID)
		b.sendPlainMessage(chatID, "Ошибка при загрузке информации о бане.")
		return
	}

	kb := buildBathhouseDetailKeyboard(bh)
	b.sendMessageWithKeyboard(chatID, formatBathhouseDetail(bh), kb)
}


func (b *Bot) doCancelBooking(ctx context.Context, chatID int64, sid string) {
	bkID, ok := b.resolveID(sid)
	if !ok {
		b.sendPlainMessage(chatID, "Бронирование не найдено.")
		return
	}

	userID, ok := b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	err := b.bookingService.Cancel(ctx, userID, domain.RoleClient, bkID, "")
	if err != nil {
		b.logger.Error("failed to cancel booking", "error", err, "booking_id", bkID)
		b.sendPlainMessage(chatID, "Ошибка при отмене бронирования. Попробуйте позже.")
		return
	}

	b.sendPlainMessage(chatID, "Бронирование отменено.")
}

func (b *Bot) doToggleFavorite(ctx context.Context, chatID int64, sid string) {
	bhID, ok := b.resolveID(sid)
	if !ok {
		b.sendPlainMessage(chatID, "Баня не найдена.")
		return
	}

	userID, ok := b.getUserID(ctx, chatID)
	if !ok {
		return
	}

	isFav, err := b.favoriteService.Toggle(ctx, userID, bhID)
	if err != nil {
		b.logger.Error("failed to toggle favorite", "error", err, "bathhouse_id", bhID)
		b.sendPlainMessage(chatID, "Ошибка при обновлении избранного.")
		return
	}

	if isFav {
		b.sendPlainMessage(chatID, "Добавлено в избранное")
	} else {
		b.sendPlainMessage(chatID, "Удалено из избранного")
	}
}

// --- Helpers ---

// getUserID resolves the Telegram chat ID to a platform user ID via TelegramLink
