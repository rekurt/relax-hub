package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler functions for telegram bot commands and messages
// These will be expanded in later tasks

type CommandHandler func(ctx context.Context, bot *Bot, msg *tgbotapi.Message) error
type MessageHandler func(ctx context.Context, bot *Bot, msg *tgbotapi.Message) error
