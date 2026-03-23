package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
	})

	result, err := svc.Login(context.Background(), "login@example.com", "mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User == nil || result.Token == "" {
		t.Fatal("user and token should not be nil/empty")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
	})

	_, err := svc.Login(context.Background(), "login@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	_, err := svc.Login(context.Background(), "nonexistent@example.com", "password")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_BlockedUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	user, _, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "blocked@example.com",
		Password: "password",
		Name:     "Blocked User",
		Role:     domain.RoleClient,
	})

	_ = userRepo.SetActive(context.Background(), user.ID, false)

	_, err := svc.Login(context.Background(), "blocked@example.com", "password")
	if !errors.Is(err, domain.ErrUserBlocked) {
		t.Errorf("should return ErrUserBlocked for blocked user, got: %v", err)
	}
}

func TestAuthService_ParseToken_Roundtrip(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	_, _, err := svc.ParseToken(context.Background(), "invalid-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized for invalid token, got: %v", err)
	}
}

func TestAuthService_ParseToken_WrongSecret(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

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
	svc2 := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg2, logger.New(logger.LevelWarn))

	_, _, err := svc2.ParseToken(context.Background(), token)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized for wrong secret, got: %v", err)
	}
}

func TestAuthService_Register_WithReferralCode(t *testing.T) {
	userRepo := mock.NewUserRepo()
	referralRepo := mock.NewReferralRepo()
	cfg := newTestConfig()
	log := logger.New(logger.LevelWarn)
	referralSvc := service.NewReferralService(referralRepo, userRepo, log)
	svc := service.NewAuthService(userRepo, referralSvc, &noopOTPService{}, cfg, log)

	// First, register a referrer and generate a referral code
	referrer, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "referrer@example.com",
		Password: "password123",
		Name:     "Referrer",
		Role:     domain.RoleClient,
	})
	if err != nil {
		t.Fatalf("unexpected error creating referrer: %v", err)
	}

	code, err := referralSvc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("unexpected error generating code: %v", err)
	}

	// Register a new user with the referral code
	referee, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "referee@example.com",
		Password:     "password123",
		Name:         "Referee",
		Role:         domain.RoleClient,
		ReferralCode: code,
	})
	if err != nil {
		t.Fatalf("unexpected error registering with referral: %v", err)
	}

	// Verify the referral was created
	stats, err := referralSvc.GetStats(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("unexpected error getting stats: %v", err)
	}
	if stats.TotalInvited != 1 {
		t.Errorf("totalInvited = %d, want 1", stats.TotalInvited)
	}
	_ = referee
}

func TestAuthService_Register_WithInvalidReferralCode(t *testing.T) {
	userRepo := mock.NewUserRepo()
	referralRepo := mock.NewReferralRepo()
	cfg := newTestConfig()
	log := logger.New(logger.LevelWarn)
	referralSvc := service.NewReferralService(referralRepo, userRepo, log)
	svc := service.NewAuthService(userRepo, referralSvc, &noopOTPService{}, cfg, log)

	// Register with an invalid referral code - should still succeed (best-effort)
	user, token, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "test@example.com",
		Password:     "password123",
		Name:         "Test User",
		Role:         domain.RoleClient,
		ReferralCode: "invalidcode",
	})
	if err != nil {
		t.Fatalf("registration should succeed even with invalid referral code, got: %v", err)
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestAuthService_RegisterPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	otpSvc := &noopOTPService{}
	svc := service.NewAuthService(userRepo, &noopReferralService{}, otpSvc, cfg, logger.New(logger.LevelWarn))

	err := svc.RegisterPhone(context.Background(), service.RegisterPhoneInput{
		Phone: "+79001234567",
		Name:  "Phone User",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_RegisterPhone_InvalidPhone(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	err := svc.RegisterPhone(context.Background(), service.RegisterPhoneInput{
		Phone: "invalid",
		Name:  "Phone User",
	})
	if !errors.Is(err, domain.ErrPhoneInvalid) {
		t.Errorf("expected ErrPhoneInvalid, got: %v", err)
	}
}

func TestAuthService_RegisterPhone_EmptyName(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	err := svc.RegisterPhone(context.Background(), service.RegisterPhoneInput{
		Phone: "+79001234567",
		Name:  "",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestAuthService_LoginPhone_NonExistent(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	err := svc.LoginPhone(context.Background(), "+79001234567")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_LoginPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	// Create user with phone
	userRepo.Create(context.Background(), &domain.User{
		Name:     "Phone User",
		Phone:    "+79001234567",
		Email:    "phone@test.com",
		Role:     domain.RoleClient,
		IsActive: true,
	})

	err := svc.LoginPhone(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_LoginPhone_BlockedUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	userRepo.Create(context.Background(), &domain.User{
		Name:     "Blocked User",
		Phone:    "+79001234567",
		Email:    "blocked@test.com",
		Role:     domain.RoleClient,
		IsActive: false,
	})

	err := svc.LoginPhone(context.Background(), "+79001234567")
	if !errors.Is(err, domain.ErrUserBlocked) {
		t.Errorf("expected ErrUserBlocked, got: %v", err)
	}
}

func TestAuthService_VerifyPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	userRepo.Create(context.Background(), &domain.User{
		Name:     "Phone User",
		Phone:    "+79001234567",
		Email:    "phone@test.com",
		Role:     domain.RoleClient,
		IsActive: true,
	})

	result, err := svc.VerifyPhone(context.Background(), "+79001234567", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User == nil {
		t.Fatal("user should not be nil")
	}
	if result.Token == "" {
		t.Fatal("token should not be empty")
	}
	if !result.User.PhoneVerified {
		t.Error("phone should be verified after successful OTP")
	}
}

func TestAuthService_VerifyPhone_NonExistentUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	_, err := svc.VerifyPhone(context.Background(), "+79001234567", "123456")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_PhoneNormalization(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, cfg, logger.New(logger.LevelWarn))

	// Create user with normalized phone
	userRepo.Create(context.Background(), &domain.User{
		Name:     "Phone User",
		Phone:    "+79001234567",
		Email:    "phone@test.com",
		Role:     domain.RoleClient,
		IsActive: true,
	})

	// Login with 8-prefix format - should normalize to +7
	err := svc.LoginPhone(context.Background(), "89001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
