package bot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/service"
)

// BookingWizardState tracks the booking wizard progress per chat
type BookingWizardState struct {
	BathhouseID     uuid.UUID
	Bathhouse       *domain.Bathhouse
	Date            time.Time
	Slots           []service.TimeSlot
	StartTime       time.Time
	EndTime         time.Time
	GuestCount      int
	PromoCode       string // applied promo code
	CertificateCode string // applied certificate code
	Step            string // "date", "time", "guests", "promo", "confirm"
	CreatedAt       time.Time
}

const wizardTTL = 30 * time.Minute

type Bot struct {
	client              *tgbotapi.BotAPI
	config              *config.TelegramConfig
	logger              *logger.Logger
	bathhouseService    service.BathhouseService
	bookingService      service.BookingService
	userService         service.UserService
	notificationService service.NotificationService
	telegramLinkService service.TelegramLinkService
	favoriteService     service.FavoriteService
	cityService         service.CityService
	promoService        service.PromoService
	certificateService  service.CertificateService
	paymentService      service.PaymentService

	// In-memory booking wizard state per chatID
	wizards   map[int64]*BookingWizardState
	wizardsMu sync.RWMutex

	// Cache: short ID -> full UUID mapping for callback data
	idCache   map[string]uuid.UUID
	idCacheMu sync.RWMutex
}

func NewBot(
	cfg *config.TelegramConfig,
	log *logger.Logger,
	bathhouseService service.BathhouseService,
	bookingService service.BookingService,
	userService service.UserService,
	notificationService service.NotificationService,
	telegramLinkService service.TelegramLinkService,
	favoriteService service.FavoriteService,
	cityService service.CityService,
	promoService service.PromoService,
	certificateService service.CertificateService,
	paymentService service.PaymentService,
) (*Bot, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	// Validate required service dependencies
	if bathhouseService == nil || bookingService == nil || userService == nil ||
		telegramLinkService == nil || favoriteService == nil || cityService == nil {
		return nil, fmt.Errorf("all service dependencies are required for the bot")
	}

	client, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
	}

	bot := &Bot{
		client:              client,
		config:              cfg,
		logger:              log,
		bathhouseService:    bathhouseService,
		bookingService:      bookingService,
		userService:         userService,
		notificationService: notificationService,
		telegramLinkService: telegramLinkService,
		favoriteService:     favoriteService,
		cityService:         cityService,
		promoService:        promoService,
		certificateService:  certificateService,
		paymentService:      paymentService,
		wizards:             make(map[int64]*BookingWizardState),
		idCache:             make(map[string]uuid.UUID),
	}

	log.Info("Telegram bot initialized", "username", client.Self.UserName)

	return bot, nil
}

// Start begins the bot in either polling or webhook mode
func (b *Bot) Start(ctx context.Context) error {
	go b.cleanupExpiredWizards(ctx)

	if b.config.Mode == "webhook" {
		return b.startWebhook(ctx)
	}
	return b.startPolling(ctx)
}

// cleanupExpiredWizards periodically removes stale wizard states
func (b *Bot) cleanupExpiredWizards(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.wizardsMu.Lock()
			now := time.Now()
			for chatID, w := range b.wizards {
				if now.Sub(w.CreatedAt) > wizardTTL {
					delete(b.wizards, chatID)
				}
			}
			b.wizardsMu.Unlock()
		}
	}
}

// startPolling runs the bot using long polling (suitable for development)
func (b *Bot) startPolling(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot in polling mode")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.client.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Shutting down Telegram bot")
			b.client.StopReceivingUpdates()
			return ctx.Err()
		case update := <-updates:
			go func(u tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						b.logger.Error("panic in handleUpdate", "recover", r)
					}
				}()
				b.handleUpdate(ctx, u)
			}(update)
		}
	}
}

// startWebhook runs the bot using webhook mode (suitable for production)
func (b *Bot) startWebhook(ctx context.Context) error {
	if b.config.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required when mode is 'webhook'")
	}

	// Derive a secret path component from the bot token for webhook verification
	webhookPath := "/telegram/webhook/" + deriveWebhookSecret(b.config.BotToken)

	// The full webhook URL must include the secret path so Telegram sends updates to the correct endpoint
	fullWebhookURL := strings.TrimRight(b.config.WebhookURL, "/") + webhookPath
	b.logger.Info("Starting Telegram bot in webhook mode", "webhook_url", fullWebhookURL)

	whConfig, err := tgbotapi.NewWebhook(fullWebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}

	if _, err := b.client.Request(whConfig); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	// Start HTTP server to receive webhook callbacks from Telegram
	// The path includes a secret derived from the bot token to prevent unauthorized access
	mux := http.NewServeMux()
	mux.HandleFunc(webhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			b.logger.Error("failed to decode webhook update", "error", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		go func(u tgbotapi.Update) {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Error("panic in handleUpdate", "recover", r)
				}
			}()
			b.handleUpdate(ctx, u)
		}(update)

		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              ":8443",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			b.logger.Error("webhook server shutdown error", "error", err)
		}
	}()

	b.logger.Info("Webhook server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("webhook server error: %w", err)
	}
	return nil
}

// handleUpdate processes incoming Telegram updates
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.InlineQuery != nil {
		b.handleInlineQuery(ctx, update.InlineQuery)
		return
	}

	if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	if update.Message.IsCommand() {
		b.handleCommand(ctx, update.Message)
	} else {
		b.handleMessage(ctx, update.Message)
	}
}

// handleInlineQuery processes inline queries (@bot_name <query>)
func (b *Bot) handleInlineQuery(ctx context.Context, iq *tgbotapi.InlineQuery) {
	query := strings.TrimSpace(iq.Query)
	if query == "" {
		return
	}

	filter := domain.BathhouseFilter{
		SearchQuery: &query,
		Page:        1,
		PageSize:    10,
		SortBy:      "rating",
		SortOrder:   "desc",
	}

	status := domain.BathhouseStatusActive
	filter.Status = &status

	result, err := b.bathhouseService.Search(ctx, filter)
	if err != nil {
		b.logger.Error("inline query search failed", "error", err, "query", query)
		return
	}

	articles := make([]interface{}, 0, len(result.Items))
	for _, bh := range result.Items {
		priceRub := float64(bh.PricePerHour) / 100
		description := fmt.Sprintf("%.0f ₽/ч | ⭐ %.1f (%d отзывов) | до %d гостей",
			priceRub, bh.Rating, bh.ReviewCount, bh.MaxGuests)

		if bh.Address != "" {
			description = bh.Address + "\n" + description
		}

		msgText := formatBathhouseDetailPlain(bh)

		article := tgbotapi.NewInlineQueryResultArticle(bh.ID.String(), bh.Name, msgText)
		article.Description = description

		articles = append(articles, article)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: iq.ID,
		Results:       articles,
		CacheTime:     60,
	}

	if _, err := b.client.Request(inlineConf); err != nil {
		b.logger.Error("failed to answer inline query", "error", err, "query", query)
	}
}

// handleCommand processes Telegram commands
func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	command := msg.Command()
	chatID := msg.Chat.ID

	switch command {
	case "start":
		b.commandStart(ctx, chatID)
	case "help":
		b.commandHelp(ctx, chatID)
	case "link":
		b.commandLink(ctx, chatID, msg)
	case "search":
		b.commandSearch(ctx, chatID, msg.CommandArguments())
	case "book":
		b.commandBook(ctx, chatID, msg.CommandArguments())
	case "mybookings":
		b.commandMyBookings(ctx, chatID, 1)
	case "cancel":
		b.commandCancel(ctx, chatID, msg.CommandArguments())
	case "favorites":
		b.commandFavorites(ctx, chatID, 1)
	default:
		b.sendPlainMessage(chatID, "Неизвестная команда. Используйте /help для справки.")
	}
}

// handleMessage processes regular text messages
func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	// Check if wizard is active and expecting promo code input
	b.wizardsMu.RLock()
	wizard := b.wizards[msg.Chat.ID]
	b.wizardsMu.RUnlock()

	if wizard != nil && wizard.Step == "promo" {
		b.doApplyCode(ctx, msg.Chat.ID, strings.TrimSpace(msg.Text))
		return
	}

	b.sendPlainMessage(msg.Chat.ID, "Используйте /help для списка команд.")
}

// handleCallbackQuery processes inline keyboard button presses
func (b *Bot) handleCallbackQuery(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	if cq.Message == nil {
		return
	}
	chatID := cq.Message.Chat.ID
	data := cq.Data

	// Acknowledge the callback
	callback := tgbotapi.NewCallback(cq.ID, "")
	if _, err := b.client.Request(callback); err != nil {
		b.logger.Error("failed to answer callback query", "error", err)
	}

	parts := strings.SplitN(data, ":", 3)
	prefix := parts[0]

	switch prefix {
	case cbSearch:
		if len(parts) >= 3 {
			page, err := strconv.Atoi(parts[2])
			if err != nil || page < 1 {
				page = 1
			}
			b.doSearch(ctx, chatID, parts[1], page)
		}
	case cbView:
		if len(parts) >= 2 {
			b.doViewBathhouse(ctx, chatID, parts[1])
		}
	case cbBook:
		if len(parts) >= 2 {
			b.doStartBookingWizard(ctx, chatID, parts[1])
		}
	case cbDate:
		if len(parts) >= 2 {
			b.doSelectDate(ctx, chatID, parts[1])
		}
	case cbSlot:
		if len(parts) >= 2 {
			idx, err := strconv.Atoi(parts[1])
			if err != nil {
				b.sendPlainMessage(chatID, "Неверный слот.")
				return
			}
			b.doSelectSlot(ctx, chatID, idx)
		}
	case cbGuests:
		if len(parts) >= 2 {
			count, err := strconv.Atoi(parts[1])
			if err != nil {
				b.sendPlainMessage(chatID, "Неверное количество гостей.")
				return
			}
			b.doSelectGuests(ctx, chatID, count)
		}
	case cbConfirm:
		b.doConfirmBooking(ctx, chatID)
	case "cancel_wizard":
		b.doCancelWizard(chatID)
	case cbCancelBk:
		if len(parts) >= 2 {
			b.doCancelBooking(ctx, chatID, parts[1])
		}
	case cbFavPage:
		if len(parts) >= 2 {
			page, _ := strconv.Atoi(parts[1])
			b.commandFavorites(ctx, chatID, page)
		}
	case cbBkPage:
		if len(parts) >= 2 {
			page, _ := strconv.Atoi(parts[1])
			b.commandMyBookings(ctx, chatID, page)
		}
	case cbFavToggle:
		if len(parts) >= 2 {
			b.doToggleFavorite(ctx, chatID, parts[1])
		}
	case cbPromoSkip:
		b.doSkipPromo(ctx, chatID)
	}
}

// --- Commands ---

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
		payURL, err := b.paymentService.InitiatePayment(ctx, userID, result.Booking.ID)
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

	err := b.bookingService.Cancel(ctx, userID, domain.RoleClient, bkID)
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
func (b *Bot) getUserID(ctx context.Context, chatID int64) (uuid.UUID, bool) {
	link, err := b.telegramLinkService.GetByTelegramID(ctx, chatID)
	if err != nil {
		b.sendPlainMessage(chatID, "Ваш Telegram не привязан к аккаунту.\nИспользуйте /link <токен> для привязки.")
		return uuid.Nil, false
	}
	return link.UserID, true
}

func (b *Bot) getWizard(chatID int64) *BookingWizardState {
	b.wizardsMu.RLock()
	defer b.wizardsMu.RUnlock()
	return b.wizards[chatID]
}

func (b *Bot) clearWizard(chatID int64) {
	b.wizardsMu.Lock()
	delete(b.wizards, chatID)
	b.wizardsMu.Unlock()
}

const maxIDCacheSize = 10000

func (b *Bot) cacheID(short string, full uuid.UUID) {
	b.idCacheMu.Lock()
	// Evict half the entries when cache exceeds max size to preserve recent mappings
	if len(b.idCache) >= maxIDCacheSize {
		evictCount := 0
		for k := range b.idCache {
			delete(b.idCache, k)
			evictCount++
			if evictCount >= maxIDCacheSize/2 {
				break
			}
		}
	}
	b.idCache[short] = full
	b.idCacheMu.Unlock()
}

func (b *Bot) resolveID(short string) (uuid.UUID, bool) {
	b.idCacheMu.RLock()
	id, ok := b.idCache[short]
	b.idCacheMu.RUnlock()
	return id, ok
}

// sendMessage sends a text message with MarkdownV2 parsing
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := b.client.Send(msg)
	if err != nil {
		b.logger.Error("failed to send message", "error", err, "chat_id", chatID)
	}
}

// sendPlainMessage sends a text message without any parse mode
func (b *Bot) sendPlainMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := b.client.Send(msg)
	if err != nil {
		b.logger.Error("failed to send message", "error", err, "chat_id", chatID)
	}
}

// sendMessageWithKeyboard sends a message with inline keyboard
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = keyboard

	_, err := b.client.Send(msg)
	if err != nil {
		b.logger.Error("failed to send message with keyboard", "error", err, "chat_id", chatID)
	}
}

// deriveWebhookSecret creates a deterministic secret path component from the bot token.
// This prevents unauthorized parties from sending fake updates to the webhook endpoint.
func deriveWebhookSecret(botToken string) string {
	h := sha256.Sum256([]byte(botToken))
	return hex.EncodeToString(h[:16])
}
