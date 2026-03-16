package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nikitaaldaev/bani/config"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration up failed: %w", err)
		}

		log.Println("Migrations applied successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()

		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration down failed: %w", err)
		}

		log.Println("Migration rolled back successfully")
		return nil
	},
}

func newMigrate() (*migrate.Migrate, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	m, err := migrate.New("file://migrations", cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateForceCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	rootCmd.AddCommand(migrateCmd)
}

var migrateForceCmd = &cobra.Command{
	Use:   "force VERSION",
	Short: "Set migration version without running migration (fixes dirty state)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}

		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()

		if err := m.Force(version); err != nil {
			return fmt.Errorf("force version failed: %w", err)
		}

		log.Printf("Forced version to %d\n", version)
		return nil
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Rollback all and re-apply all migrations (dev only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.Environment == "production" {
			return fmt.Errorf("migrate reset is not allowed in production environment")
		}

		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()

		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("rollback failed: %w", err)
		}
		m.Close()

		m, err = newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration up failed: %w", err)
		}

		log.Println("Database reset and migrations applied successfully")
		return nil
	},
}
