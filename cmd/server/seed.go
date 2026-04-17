package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var (
	seedEmail    string
	seedPassword string
	seedName     string
)

var seedAdminCmd = &cobra.Command{
	Use:   "seed-admin",
	Short: "Create the initial admin user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if seedEmail == "" || seedPassword == "" || seedName == "" {
			return fmt.Errorf("email, password, and name are required")
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		now := time.Now()
		user := domain.User{
			ID:           uuid.New(),
			Email:        seedEmail,
			PasswordHash: string(hash),
			Name:         seedName,
			Role:         domain.RoleAdmin,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		tag, err := pool.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, name, phone, role, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (email) WHERE email::text <> ''::text DO NOTHING`,
			user.ID, user.Email, user.PasswordHash, user.Name, user.Phone,
			string(user.Role), user.IsActive, user.CreatedAt, user.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		if tag.RowsAffected() == 0 {
			log.Printf("Admin user with email %s already exists, skipping\n", user.Email)
		} else {
			log.Printf("Admin user created: %s (%s)\n", user.Email, user.ID)
		}
		return nil
	},
}

func init() {
	seedAdminCmd.Flags().StringVar(&seedEmail, "email", "", "admin email (required)")
	seedAdminCmd.Flags().StringVar(&seedPassword, "password", "", "admin password (required)")
	seedAdminCmd.Flags().StringVar(&seedName, "name", "Admin", "admin name")
	rootCmd.AddCommand(seedAdminCmd)
}
