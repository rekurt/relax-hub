package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/pquerna/otp/totp"
)

func TestTwoFAService_GenerateTOTPSecret(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	secret, qrURL, err := svc.GenerateTOTPSecret(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret == "" {
		t.Error("secret should not be empty")
	}
	if qrURL == "" {
		t.Error("qrURL should not be empty")
	}

	// Verify secret was stored on user
	updated, _ := userRepo.GetByID(context.Background(), user.ID)
	if updated.TOTPSecret == "" {
		t.Error("TOTP secret should be stored on user")
	}
}

func TestTwoFAService_GenerateTOTPSecret_AlreadyEnabled(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:          uuid.New(),
		Email:       "test@example.com",
		Name:        "Test",
		Role:        domain.RoleClient,
		IsActive:    true,
		TwoFAMethod: domain.TwoFATOTP,
	}
	userRepo.Create(context.Background(), user)

	_, _, err := svc.GenerateTOTPSecret(context.Background(), user.ID)
	if !errors.Is(err, domain.Err2FAAlreadyEnabled) {
		t.Errorf("expected Err2FAAlreadyEnabled, got: %v", err)
	}
}

func TestTwoFAService_EnableTOTP(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	// Generate secret first
	secret, _, err := svc.GenerateTOTPSecret(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error generating secret: %v", err)
	}

	// Generate a valid TOTP code
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	// Enable TOTP
	err = svc.EnableTOTP(context.Background(), user.ID, code)
	if err != nil {
		t.Fatalf("unexpected error enabling TOTP: %v", err)
	}

	// Verify 2FA method is set
	updated, _ := userRepo.GetByID(context.Background(), user.ID)
	if updated.TwoFAMethod != domain.TwoFATOTP {
		t.Errorf("expected TwoFAMethod=totp, got %s", updated.TwoFAMethod)
	}
}

func TestTwoFAService_EnableTOTP_InvalidCode(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	// Generate secret
	svc.GenerateTOTPSecret(context.Background(), user.ID)

	// Try enabling with wrong code
	err := svc.EnableTOTP(context.Background(), user.ID, "000000")
	if !errors.Is(err, domain.Err2FAInvalidCode) {
		t.Errorf("expected Err2FAInvalidCode, got: %v", err)
	}
}

func TestTwoFAService_DisableTOTP(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	// Setup TOTP
	secret, _, _ := svc.GenerateTOTPSecret(context.Background(), user.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.EnableTOTP(context.Background(), user.ID, code)

	// Disable with valid code
	code2, _ := totp.GenerateCode(secret, time.Now())
	err := svc.DisableTOTP(context.Background(), user.ID, code2)
	if err != nil {
		t.Fatalf("unexpected error disabling TOTP: %v", err)
	}

	updated, _ := userRepo.GetByID(context.Background(), user.ID)
	if updated.TwoFAMethod != domain.TwoFANone {
		t.Errorf("expected TwoFAMethod=none, got %s", updated.TwoFAMethod)
	}
	if updated.TOTPSecret != "" {
		t.Error("TOTP secret should be cleared")
	}
}

func TestTwoFAService_DisableTOTP_NotEnabled(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	err := svc.DisableTOTP(context.Background(), user.ID, "123456")
	if !errors.Is(err, domain.Err2FANotEnabled) {
		t.Errorf("expected Err2FANotEnabled, got: %v", err)
	}
}

func TestTwoFAService_VerifyTOTP(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	// Setup TOTP
	secret, _, _ := svc.GenerateTOTPSecret(context.Background(), user.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.EnableTOTP(context.Background(), user.ID, code)

	// Verify with valid code
	code2, _ := totp.GenerateCode(secret, time.Now())
	valid, err := svc.VerifyTOTP(context.Background(), user.ID, code2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid TOTP code")
	}

	// Verify with invalid code
	valid, err = svc.VerifyTOTP(context.Background(), user.ID, "000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected invalid TOTP code")
	}
}

func TestTwoFAService_EnableSMS2FA(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:            uuid.New(),
		Email:         "test@example.com",
		Name:          "Test",
		Phone:         "+79001234567",
		PhoneVerified: true,
		Role:          domain.RoleClient,
		IsActive:      true,
	}
	userRepo.Create(context.Background(), user)

	err := svc.EnableSMS2FA(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := userRepo.GetByID(context.Background(), user.ID)
	if updated.TwoFAMethod != domain.TwoFASMS {
		t.Errorf("expected TwoFAMethod=sms, got %s", updated.TwoFAMethod)
	}
}

func TestTwoFAService_EnableSMS2FA_NoPhone(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	userRepo.Create(context.Background(), user)

	err := svc.EnableSMS2FA(context.Background(), user.ID)
	if !errors.Is(err, domain.Err2FAPhoneRequired) {
		t.Errorf("expected Err2FAPhoneRequired, got: %v", err)
	}
}

func TestTwoFAService_EnableSMS2FA_AlreadyEnabled(t *testing.T) {
	userRepo := mock.NewUserRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewTwoFAService(userRepo, &noopOTPService{}, log)

	user := &domain.User{
		ID:            uuid.New(),
		Email:         "test@example.com",
		Name:          "Test",
		Phone:         "+79001234567",
		PhoneVerified: true,
		Role:          domain.RoleClient,
		IsActive:      true,
		TwoFAMethod:   domain.TwoFATOTP,
	}
	userRepo.Create(context.Background(), user)

	err := svc.EnableSMS2FA(context.Background(), user.ID)
	if !errors.Is(err, domain.Err2FAAlreadyEnabled) {
		t.Errorf("expected Err2FAAlreadyEnabled, got: %v", err)
	}
}

func TestAuthService_Login_With2FA(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	// Register user
	user, _, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "twofa@example.com",
		Password: "password123",
		Name:     "2FA User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	// Enable TOTP on user directly in repo
	user.TwoFAMethod = domain.TwoFATOTP
	user.TOTPSecret = "JBSWY3DPEHPK3PXP"
	userRepo.Update(context.Background(), user)

	// Login should return partial token
	result, err := svc.Login(context.Background(), "twofa@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Requires2FA {
		t.Error("expected Requires2FA=true")
	}
	if result.Token == "" {
		t.Error("partial token should not be empty")
	}

	// Parse partial token
	parsedID, err := svc.ParsePartialToken(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("failed to parse partial token: %v", err)
	}
	if parsedID != user.ID {
		t.Errorf("expected userID=%v, got=%v", user.ID, parsedID)
	}

	// Complete 2FA login
	finalUser, finalToken, err := svc.Complete2FALogin(context.Background(), parsedID)
	if err != nil {
		t.Fatalf("failed to complete 2FA login: %v", err)
	}
	if finalUser == nil {
		t.Fatal("user should not be nil")
	}
	if finalToken == "" {
		t.Fatal("token should not be empty")
	}

	// Full token should be parseable
	_, role, err := svc.ParseToken(context.Background(), finalToken)
	if err != nil {
		t.Fatalf("failed to parse full token: %v", err)
	}
	if role != domain.RoleClient {
		t.Errorf("expected role=client, got=%s", role)
	}
}

func TestAuthService_Login_Without2FA(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	svc.Register(context.Background(), service.RegisterInput{
		Email:    "normal@example.com",
		Password: "password123",
		Name:     "Normal User",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	result, err := svc.Login(context.Background(), "normal@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Requires2FA {
		t.Error("expected Requires2FA=false for user without 2FA")
	}
	if result.Token == "" {
		t.Error("should return full token")
	}
	if result.User == nil {
		t.Error("user should not be nil")
	}
}

func TestAuthService_ParsePartialToken_Invalid(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	_, err := svc.ParsePartialToken(context.Background(), "invalid-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestAuthService_ParsePartialToken_FullTokenRejected(t *testing.T) {
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	// Register and get a full token
	_, fullToken, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test",
		Role:     domain.RoleClient,
		AgeConfirmed: true,
	})

	// Full token should not be accepted as partial token
	_, err := svc.ParsePartialToken(context.Background(), fullToken)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized for full token used as partial, got: %v", err)
	}
}

func TestAuthService_VerifyPhone_With2FA_SMS_Skipped(t *testing.T) {
	// SMS 2FA is skipped for phone-based login because phone possession
	// was already proven by verifying the OTP code.
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user := &domain.User{
		ID:            uuid.New(),
		Name:          "Phone 2FA User",
		Phone:         "+79001234567",
		Email:         "phone2fa@test.com",
		PhoneVerified: true,
		Role:          domain.RoleClient,
		IsActive:      true,
		TwoFAMethod:   domain.TwoFASMS,
	}
	userRepo.Create(context.Background(), user)

	result, err := svc.VerifyPhone(context.Background(), "+79001234567", "123456", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Requires2FA {
		t.Error("expected Requires2FA=false for SMS 2FA on phone-based login (phone already verified)")
	}
}

func TestAuthService_VerifyPhone_With2FA_TOTP_Required(t *testing.T) {
	// TOTP 2FA must still be required for phone-based login.
	userRepo := mock.NewUserRepo()
	cfg := newTestConfig()
	svc := service.NewAuthService(userRepo, &noopReferralService{}, &noopOTPService{}, nil, nil, cfg, logger.New(logger.LevelWarn))

	user := &domain.User{
		ID:            uuid.New(),
		Name:          "Phone TOTP User",
		Phone:         "+79001234568",
		Email:         "phonetotp@test.com",
		PhoneVerified: true,
		Role:          domain.RoleClient,
		IsActive:      true,
		TwoFAMethod:   domain.TwoFATOTP,
	}
	userRepo.Create(context.Background(), user)

	result, err := svc.VerifyPhone(context.Background(), "+79001234568", "123456", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Requires2FA {
		t.Error("expected Requires2FA=true for TOTP 2FA on phone-based login")
	}
}
