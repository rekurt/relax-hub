package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/app"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/spf13/cobra"
)

var withAdmin bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if withAdmin {
			cfg.Admin.Enabled = true
		}

		log := logger.New(logger.ParseLogLevel(cfg.Logger.Level))
		log.Info("Starting server", "host", cfg.Server.Host, "port", cfg.Server.Port)

		if cfg.Admin.Enabled {
			adminURL := fmt.Sprintf("http://%s:%d%s", cfg.Server.Host, cfg.Server.Port, cfg.Admin.Prefix)
			log.Info("Admin panel enabled", "url", adminURL)
		}

		fxApp := app.New(cfg)

		// Create a channel to listen for interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start the app in a goroutine
		go func() {
			if err := fxApp.Start(context.Background()); err != nil {
				log.Error("Failed to start application", "error", err)
				sigChan <- syscall.SIGINT
			}
		}()

		// Wait for shutdown signal
		sig := <-sigChan
		log.Info("Received shutdown signal", "signal", sig.String())

		// Create a context with 30-second timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info("Starting graceful shutdown", "timeout_seconds", 30)

		// Stop the application with the timeout context
		if err := fxApp.Stop(shutdownCtx); err != nil {
			log.Error("Error during graceful shutdown", "error", err)
			return err
		}

		log.Info("Graceful shutdown completed successfully")
		return nil
	},
}

func init() {
	serveCmd.Flags().BoolVar(&withAdmin, "with-admin", false, "Enable admin panel")
	rootCmd.AddCommand(serveCmd)
}
