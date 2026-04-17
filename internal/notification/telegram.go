package notification

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// TelegramSender sends notifications via Telegram bot.
type TelegramSender interface {
	Send(ctx context.Context, chatID int64, title, body string) error
}

type telegramSender struct {
	client *tgbotapi.BotAPI
	logger *logger.Logger
}

// NewTelegramSender creates a TelegramSender using a lightweight bot API client (no polling).
func NewTelegramSender(botToken string, log *logger.Logger) (TelegramSender, error) {
	client, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot api: %w", err)
	}
	client.Debug = false

	return &telegramSender{client: client, logger: log}, nil
}

func (s *telegramSender) Send(_ context.Context, chatID int64, title, body string) error {
	text := fmt.Sprintf("*%s*\n\n%s", escapeTelegramMD(title), escapeTelegramMD(body))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	if _, err := s.client.Send(msg); err != nil {
		s.logger.Error("failed to send telegram notification", "chat_id", chatID, "error", err)
		return fmt.Errorf("send telegram notification: %w", err)
	}

	s.logger.Debug("telegram notification sent", "chat_id", chatID, "title", title)
	return nil
}

func escapeTelegramMD(s string) string {
	special := []string{"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	r := s
	for _, ch := range special {
		r = strings.ReplaceAll(r, ch, "\\"+ch)
	}
	return r
}
