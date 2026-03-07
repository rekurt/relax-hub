package main

import (
	"fmt"

	"github.com/nikitaaldaev/bani/config"
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

		// The standalone bot binary requires a full DI setup to provide services.
		// For now, use the main app binary which includes all service dependencies.
		return fmt.Errorf("standalone bot binary is not supported; run the bot through the main application with BANI_TELEGRAM_BOT_TOKEN configured")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
