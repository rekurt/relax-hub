package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/service"
)

type Bot struct {
	client                *tgbotapi.BotAPI
	config                *config.TelegramConfig
	logger                *logger.Logger
	bathhouseService      service.BathhouseService
	bookingService        service.BookingService
	userService           service.UserService
	notificationService   service.NotificationService
	telegramLinkService   service.TelegramLinkService
}

func NewBot(
	cfg *config.TelegramConfig,
	log *logger.Logger,
	bathhouseService service.BathhouseService,
	bookingService service.BookingService,
	userService service.UserService,
	notificationService service.NotificationService,
	telegramLinkService service.TelegramLinkService,
) (*Bot, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
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
	}

	log.Info("Telegram bot initialized", "username", client.Self.UserName)

	return bot, nil
}

// Start begins the bot in either polling or webhook mode
func (b *Bot) Start(ctx context.Context) error {
	if b.config.Mode == "webhook" {
		return b.startWebhook(ctx)
	}
	return b.startPolling(ctx)
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
			go b.handleUpdate(ctx, update)
		}
	}
}

// startWebhook runs the bot using webhook mode (suitable for production)
func (b *Bot) startWebhook(ctx context.Context) error {
	if b.config.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required when mode is 'webhook'")
	}

	b.logger.Info("Starting Telegram bot in webhook mode", "webhook_url", b.config.WebhookURL)

	// For webhook mode, we set the webhook URL via the Telegram Bot API
	whConfig, err := tgbotapi.NewWebhook(b.config.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}

	if _, err := b.client.Request(whConfig); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	// For webhook mode, the HTTP server will handle updates
	// This is a placeholder - actual webhook handling happens in the HTTP server
	<-ctx.Done()
	return nil
}

// HandleUpdate processes incoming Telegram updates
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// Route message to appropriate handler
	if update.Message.IsCommand() {
		b.handleCommand(ctx, update.Message)
	} else if update.Message.Text != "" {
		b.handleMessage(ctx, update.Message)
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
		b.commandLink(ctx, chatID, msg.CommandArguments())
	default:
		b.sendMessage(chatID, "Неизвестная команда. Используйте /help для справки.")
	}
}

// handleMessage processes regular text messages
func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	b.sendMessage(msg.Chat.ID, "Я получил ваше сообщение. Используйте /help для справки.")
}

// commandStart handles the /start command
func (b *Bot) commandStart(ctx context.Context, chatID int64) {
	text := "Добро пожаловать в Bани! 🛁\n\n" +
		"Я помогу вам найти идеальную баню и забронировать сеанс.\n\n" +
		"Доступные команды:\n" +
		"/search <город> - поиск бань в городе\n" +
		"/mybookings - мои бронирования\n" +
		"/favorites - мое избранное\n" +
		"/help - справка\n"
	b.sendMessage(chatID, text)
}

// commandHelp handles the /help command
func (b *Bot) commandHelp(ctx context.Context, chatID int64) {
	text := "Справка по командам:\n\n" +
		"/start - начать работу с ботом\n" +
		"/search <город> - найти бани в городе\n" +
		"/mybookings - посмотреть мои бронирования\n" +
		"/favorites - мое избранное\n" +
		"/link <токен> - привязать Telegram-аккаунт\n"
	b.sendMessage(chatID, text)
}

// commandLink handles the /link command for account linking
func (b *Bot) commandLink(ctx context.Context, chatID int64, token string) {
	if token == "" {
		b.sendMessage(chatID, "Пожалуйста, укажите токен: /link <токен>")
		return
	}

	// TODO: Implement token validation and linking logic
	// This requires generating and validating one-time tokens on the web API side
	// For now, provide placeholder feedback
	b.sendMessage(chatID, "Функция привязки аккаунта находится в разработке. Пожалуйста, используйте веб-интерфейс.")
}

// sendMessage sends a text message to the user
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := b.client.Send(msg)
	if err != nil {
		b.logger.Error("failed to send message", "error", err, "chat_id", chatID)
	}
}
