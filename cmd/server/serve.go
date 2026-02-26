package main

import (
	"fmt"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/app"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		log := logger.New(logger.ParseLogLevel(cfg.Logger.Level))
		log.Info("Starting server", "host", cfg.Server.Host, "port", cfg.Server.Port)

		fxApp := app.New(cfg)
		fxApp.Run()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
