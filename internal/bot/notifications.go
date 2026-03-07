package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
)

type telegramSender struct {
	client *tgbotapi.BotAPI
	logger *logger.Logger
}

// NewTelegramSender creates a notification.TelegramSender using the given bot API client.
func NewTelegramSender(client *tgbotapi.BotAPI, log *logger.Logger) notification.TelegramSender {
	return &telegramSender{client: client, logger: log}
}

func (s *telegramSender) Send(_ context.Context, chatID int64, title, body string) error {
	text := fmt.Sprintf("*%s*\n\n%s", escapeMD(title), escapeMD(body))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	if _, err := s.client.Send(msg); err != nil {
		s.logger.Error("failed to send telegram notification", "chat_id", chatID, "error", err)
		return fmt.Errorf("send telegram notification: %w", err)
	}

	s.logger.Debug("telegram notification sent", "chat_id", chatID, "title", title)
	return nil
}
