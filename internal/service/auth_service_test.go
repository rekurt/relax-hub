package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:   "test-secret-key-for-testing",
			TokenTTL: 3600_000_000_000, // 1 hour in nanoseconds
		},
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	user, token, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Phone:    "+7900000000",
		Role:     domain.RoleClient,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if user.Email != "test@example.com" {
		t.Errorf("email = %q, want %q", user.Email, "test@example.com")
	}
	if user.Role != domain.RoleClient {
		t.Errorf("role = %q, want %q", user.Role, domain.RoleClient)
	}
}

func TestAuthService_Register_AsOwner(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner User",
		Role:     domain.RoleOwner,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != domain.RoleOwner {
		t.Errorf("role = %q, want %q", user.Role, domain.RoleOwner)
	}
}

func TestAuthService_Register_InvalidRole(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	tests := []struct {
		name string
		role domain.UserRole
	}{
		{"admin", domain.RoleAdmin},
		{"representative", domain.RoleRepresentative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), service.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test",
				Role:     tt.role,
			})
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("should reject registration as %s, got: %v", tt.role, err)
			}
		})
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "dup@example.com",
		Password: "password123",
		Name:     "First User",
		Role:     domain.RoleClient,
	})

	_, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "dup@example.com",
		Password: "password456",
		Name:     "Second User",
		Role:     domain.RoleClient,
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("should return ErrAlreadyExists, got: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
	})

	user, token, err := svc.Login(context.Background(), "login@example.com", "mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || token == "" {
		t.Fatal("user and token should not be nil/empty")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
	})

	_, _, err := svc.Login(context.Background(), "login@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, _, err := svc.Login(context.Background(), "nonexistent@example.com", "password")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_BlockedUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	user, _, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "blocked@example.com",
		Password: "password",
		Name:     "Blocked User",
		Role:     domain.RoleClient,
	})

	_ = userRepo.SetActive(context.Background(), user.ID, false)

	_, _, err := svc.Login(context.Background(), "blocked@example.com", "password")
	if !errors.Is(err, domain.ErrUserBlocked) {
		t.Errorf("should return ErrUserBlocked for blocked user, got: %v", err)
	}
}

func TestAuthService_ParseToken_Roundtrip(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	user, token, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "parse@example.com",
		Password: "password",
		Name:     "Parse User",
		Role:     domain.RoleOwner,
	})

	userID, role, err := svc.ParseToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != user.ID {
		t.Errorf("userID = %v, want %v", userID, user.ID)
	}
	if role != domain.RoleOwner {
		t.Errorf("role = %v, want %v", role, domain.RoleOwner)
	}
}

func TestAuthService_ParseToken_InvalidToken(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, _, err := svc.ParseToken(context.Background(), "invalid-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized for invalid token, got: %v", err)
	}
}

func TestAuthService_ParseToken_WrongSecret(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, cfg)

	_, token, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test",
		Role:     domain.RoleClient,
	})

	cfg2 := &config.Config{
		JWT: config.JWTConfig{
			Secret:   "different-secret-key",
			TokenTTL: 3600_000_000_000,
		},
	}
	svc2 := service.NewAuthService(userRepo, cfg2)

	_, _, err := svc2.ParseToken(context.Background(), token)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized for wrong secret, got: %v", err)
	}
}
