package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/bot"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Telegram bot",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		log := logger.New(logger.ParseLogLevel(cfg.Logger.Level))

		// Validate telegram config
		if cfg.Telegram.BotToken == "" {
			return fmt.Errorf("telegram bot token is required (set BANI_TELEGRAM_BOT_TOKEN)")
		}

		log.Info("Starting Telegram bot", "mode", cfg.Telegram.Mode)

		// Create a minimal bot for infrastructure testing
		// In production, this would use the full app with all services
		botInstance, err := bot.NewBot(
			&cfg.Telegram,
			log,
			nil, // bathhouseService - will be provided by DI
			nil, // bookingService - will be provided by DI
			nil, // userService - will be provided by DI
			nil, // notificationService - will be provided by DI
			nil, // telegramLinkService - will be provided by DI
			nil, // favoriteService - will be provided by DI
			nil, // cityService - will be provided by DI
		)
		if err != nil {
			return fmt.Errorf("failed to create bot: %w", err)
		}

		// Create a cancellable context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create a channel to listen for interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start the bot in a goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- botInstance.Start(ctx)
		}()

		// Wait for shutdown signal or bot error
		select {
		case sig := <-sigChan:
			log.Info("Received shutdown signal", "signal", sig.String())
			cancel()
		case err := <-errChan:
			if err != nil && err != context.Canceled {
				return fmt.Errorf("bot error: %w", err)
			}
		}

		log.Info("Graceful shutdown completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
