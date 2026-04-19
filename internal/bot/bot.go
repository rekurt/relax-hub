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
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/service"
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
