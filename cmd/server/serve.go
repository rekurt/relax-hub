package main

import (
	"fmt"

	"github.com/nikitaaldaev/bani/config"
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

		fmt.Printf("Starting server on %s:%d\n", cfg.Server.Host, cfg.Server.Port)

		// TODO: start fx app with HTTP server
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
