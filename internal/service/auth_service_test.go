package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:   "test-secret-key-for-testing",
			TokenTTL: 3600_000_000_000, // 1 hour in nanoseconds
		},
		WelcomeBonus: config.WelcomeBonusConfig{
			Amount:     50000, // 500 RUB
			AmountBY:   1500,  // 15 BYN
			ExpiryDays: 30,
		},
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user, token, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Phone:    "+7900000000",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "owner@example.com",
		Password: "password123",
		Name:     "Owner User",
		Role:     domain.RoleOwner,
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "dup@example.com",
		Password: "password123",
		Name:     "First User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	_, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "dup@example.com",
		Password: "password456",
		Name:     "Second User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("should return ErrAlreadyExists, got: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, _, _ = svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	_, err := svc.Login(context.Background(), "login@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, err := svc.Login(context.Background(), "nonexistent@example.com", "password")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_Login_BlockedUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user, _, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "blocked@example.com",
		Password: "password",
		Name:     "Blocked User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user, token, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "parse@example.com",
		Password: "password",
		Name:     "Parse User",
		Role:     domain.RoleOwner,
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, _, err := svc.ParseToken(context.Background(), "invalid-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("should return ErrUnauthorized for invalid token, got: %v", err)
	}
}

func TestAuthService_ParseToken_WrongSecret(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, token, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	cfg2 := &config.Config{
		JWT: config.JWTConfig{
			Secret:   "different-secret-key",
			TokenTTL: 3600_000_000_000,
		},
	}
	svc2 := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg2, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, referralSvc, &noopOTPService{}, nil, nil, cfg, log)

	// First, register a referrer and generate a referral code
	referrer, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "referrer@example.com",
		Password: "password123",
		Name:     "Referrer",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
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
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, referralSvc, &noopOTPService{}, nil, nil, cfg, log)

	// Register with an invalid referral code - should still succeed (best-effort)
	user, token, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "test@example.com",
		Password:     "password123",
		Name:         "Test User",
		Role:         domain.RoleClient,
		ReferralCode: "invalidcode",
		AgeConfirmed: true,
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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, otpSvc, nil, nil, cfg, logger.New(logger.LevelWarn))

	err := svc.RegisterPhone(context.Background(), service.RegisterPhoneInput{
		Phone:        "+79001234567",
		Name:         "Phone User",
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_StartPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	if err := svc.StartPhone(context.Background(), "+79001234567"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_StartPhone_InvalidPhone(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	err := svc.StartPhone(context.Background(), "invalid")
	if !errors.Is(err, domain.ErrPhoneInvalid) {
		t.Errorf("expected ErrPhoneInvalid, got: %v", err)
	}
}

func TestAuthService_RegisterPhone_InvalidPhone(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	// Returns nil for non-existent phone to prevent phone enumeration
	err := svc.LoginPhone(context.Background(), "+79001234567")
	if err != nil {
		t.Errorf("expected nil (anti-enumeration), got: %v", err)
	}
}

func TestAuthService_LoginPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

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
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	userRepo.Create(context.Background(), &domain.User{
		Name:     "Blocked User",
		Phone:    "+79001234567",
		Email:    "blocked@test.com",
		Role:     domain.RoleClient,
		IsActive: false,
	})

	// Returns nil for blocked users to prevent phone enumeration
	// (no OTP is sent, but caller sees success)
	err := svc.LoginPhone(context.Background(), "+79001234567")
	if err != nil {
		t.Errorf("expected nil (anti-enumeration), got: %v", err)
	}
}

func TestAuthService_VerifyPhone_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	userRepo.Create(context.Background(), &domain.User{
		Name:     "Phone User",
		Phone:    "+79001234567",
		Email:    "phone@test.com",
		Role:     domain.RoleClient,
		IsActive: true,
	})

	result, err := svc.VerifyPhone(context.Background(), "+79001234567", "123456", "", false)
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

func TestAuthService_VerifyPhone_NewUserCreated(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	// VerifyPhone for non-existent user should create a new user (phone registration)
	result, err := svc.VerifyPhone(context.Background(), "+79001234567", "123456", "New User", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.User == nil {
		t.Fatal("expected user to be created")
	}
	if result.User.Phone != "+79001234567" {
		t.Errorf("expected phone +79001234567, got %s", result.User.Phone)
	}
	if !result.User.PhoneVerified {
		t.Error("phone should be verified")
	}
	if result.User.Role != domain.RoleClient {
		t.Errorf("expected role client, got %s", result.User.Role)
	}
	if result.Token == "" {
		t.Error("expected token to be generated")
	}
}

func TestAuthService_PhoneNormalization(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

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

func TestAuthService_Register_AgeNotConfirmed(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "test@example.com",
		Password:     "password123",
		Name:         "Test User",
		Role:         domain.RoleClient,
		AgeConfirmed: false,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput when age not confirmed, got: %v", err)
	}
}

func TestAuthService_Register_AgeConfirmedFlag(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "test@example.com",
		Password:     "password123",
		Name:         "Test User",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !user.AgeConfirmed {
		t.Error("expected AgeConfirmed=true on created user")
	}
}

func TestAuthService_Register_WelcomeBonus(t *testing.T) {
	userRepo := mock.NewUserRepo()
	walletRepo := mock.NewWalletRepo()
	cfg := newTestConfig()
	log := logger.New(logger.LevelWarn)
	walletSvc := service.NewWalletService(walletRepo, log)
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, walletSvc, nil, cfg, log)

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "bonus@example.com",
		Password:     "password123",
		Name:         "Bonus User",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify wallet was created
	wallet, err := walletSvc.GetWallet(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected wallet to be created, got error: %v", err)
	}
	if wallet.Currency != domain.WalletCurrencyRUB {
		t.Errorf("wallet currency = %s, want RUB", wallet.Currency)
	}

	// Verify welcome bonus was credited
	if wallet.Balance != 50000 {
		t.Errorf("wallet balance = %d, want 50000 (500 RUB welcome bonus)", wallet.Balance)
	}

	// Verify transaction was recorded
	txs, err := walletSvc.ListTransactions(context.Background(), user.ID, domain.WalletTransactionFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if txs.TotalCount != 1 {
		t.Fatalf("expected 1 transaction, got %d", txs.TotalCount)
	}
	tx := txs.Items[0]
	if tx.Type != domain.WalletTxWelcomeBonus {
		t.Errorf("transaction type = %s, want welcome_bonus", tx.Type)
	}
	if tx.Amount != 50000 {
		t.Errorf("transaction amount = %d, want 50000", tx.Amount)
	}
	if !tx.IsBonus {
		t.Error("expected transaction to be marked as bonus")
	}
	if tx.ExpiresAt == nil {
		t.Error("expected welcome bonus to have expiry date")
	} else {
		expectedExpiry := time.Now().AddDate(0, 0, 30)
		diff := tx.ExpiresAt.Sub(expectedExpiry)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("welcome bonus expiry too far from expected: got %v, want ~%v", tx.ExpiresAt, expectedExpiry)
		}
	}
}

func TestAuthService_Register_WelcomeBonusDisabled(t *testing.T) {
	userRepo := mock.NewUserRepo()
	walletRepo := mock.NewWalletRepo()
	cfg := newTestConfig()
	cfg.WelcomeBonus.Amount = 0 // disable welcome bonus
	log := logger.New(logger.LevelWarn)
	walletSvc := service.NewWalletService(walletRepo, log)
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, walletSvc, nil, cfg, log)

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "nobonus@example.com",
		Password:     "password123",
		Name:         "No Bonus User",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wallet should be created but with zero balance
	wallet, err := walletSvc.GetWallet(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected wallet to be created, got error: %v", err)
	}
	if wallet.Balance != 0 {
		t.Errorf("wallet balance = %d, want 0 (no welcome bonus)", wallet.Balance)
	}
}

func TestAuthService_Register_WelcomeBonus_BY_Region(t *testing.T) {
	userRepo := mock.NewUserRepo()
	walletRepo := mock.NewWalletRepo()
	cfg := newTestConfig()
	log := logger.New(logger.LevelWarn)
	walletSvc := service.NewWalletService(walletRepo, log)
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, walletSvc, nil, cfg, log)

	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "by-user@example.com",
		Password:     "password123",
		Name:         "BY User",
		Phone:        "+375291234567",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Region != domain.RegionBY {
		t.Errorf("user region = %s, want BY (detected from phone)", user.Region)
	}

	wallet, err := walletSvc.GetWallet(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected wallet to be created, got error: %v", err)
	}
	if wallet.Currency != domain.WalletCurrencyBYN {
		t.Errorf("wallet currency = %s, want BYN", wallet.Currency)
	}
	if wallet.Balance != 1500 {
		t.Errorf("wallet balance = %d, want 1500 (15 BYN welcome bonus)", wallet.Balance)
	}
}

func TestAuthService_Register_WelcomeBonus_ExplicitRegion(t *testing.T) {
	userRepo := mock.NewUserRepo()
	walletRepo := mock.NewWalletRepo()
	cfg := newTestConfig()
	log := logger.New(logger.LevelWarn)
	walletSvc := service.NewWalletService(walletRepo, log)
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, walletSvc, nil, cfg, log)

	// Explicit region BY overrides phone detection
	user, _, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "explicit-by@example.com",
		Password:     "password123",
		Name:         "Explicit BY",
		Phone:        "+79001234567", // RU phone, but explicit BY region
		Role:         domain.RoleClient,
		Region:       domain.RegionBY,
		AgeConfirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Region != domain.RegionBY {
		t.Errorf("user region = %s, want BY (explicit)", user.Region)
	}

	wallet, err := walletSvc.GetWallet(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected wallet to be created, got error: %v", err)
	}
	if wallet.Currency != domain.WalletCurrencyBYN {
		t.Errorf("wallet currency = %s, want BYN", wallet.Currency)
	}
	if wallet.Balance != 1500 {
		t.Errorf("wallet balance = %d, want 1500 (15 BYN)", wallet.Balance)
	}
}

func TestAuthService_Register_NilWalletService(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	// Registration should succeed even without wallet service
	user, token, err := svc.Register(context.Background(), service.RegisterInput{
		Email:        "nowallet@example.com",
		Password:     "password123",
		Name:         "No Wallet User",
		Role:         domain.RoleClient,
		AgeConfirmed: true,
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
}
