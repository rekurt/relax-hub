package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type TwoFAService interface {
	GenerateTOTPSecret(ctx context.Context, userID uuid.UUID) (secret string, qrURL string, err error)
	EnableTOTP(ctx context.Context, userID uuid.UUID, code string) error
	DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error
	VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	EnableSMS2FA(ctx context.Context, userID uuid.UUID) error
	SendSMS2FA(ctx context.Context, userID uuid.UUID) error
	VerifySMS2FA(ctx context.Context, userID uuid.UUID, code string) (bool, error)
}

type twoFAService struct {
	userRepo repository.UserRepository
	otpSvc   OTPService
	logger   *logger.Logger
	issuer   string
}

func NewTwoFAService(userRepo repository.UserRepository, otpSvc OTPService, log *logger.Logger) TwoFAService {
	return &twoFAService{
		userRepo: userRepo,
		otpSvc:   otpSvc,
		logger:   log,
		issuer:   "Bani",
	}
}

func (s *twoFAService) GenerateTOTPSecret(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if user.TwoFAMethod == domain.TwoFATOTP {
		return "", "", domain.Err2FAAlreadyEnabled
	}

	accountName := user.Email
	if accountName == "" {
		accountName = user.Phone
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountName,
		Period:      30,
		Digits:      6,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP key: %w", err)
	}

	// Store secret temporarily on user (not yet enabled)
	user.TOTPSecret = key.Secret()
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *twoFAService) EnableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFAMethod == domain.TwoFATOTP {
		return domain.Err2FAAlreadyEnabled
	}

	if user.TOTPSecret == "" {
		return fmt.Errorf("%w: generate secret first", domain.ErrInvalidInput)
	}

	if !totp.Validate(code, user.TOTPSecret) {
		return domain.Err2FAInvalidCode
	}

	user.TwoFAMethod = domain.TwoFATOTP
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *twoFAService) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFAMethod != domain.TwoFATOTP {
		return domain.Err2FANotEnabled
	}

	if !totp.Validate(code, user.TOTPSecret) {
		return domain.Err2FAInvalidCode
	}

	user.TwoFAMethod = domain.TwoFANone
	user.TOTPSecret = ""
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *twoFAService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.TwoFAMethod != domain.TwoFATOTP {
		return false, domain.Err2FANotEnabled
	}

	return totp.Validate(code, user.TOTPSecret), nil
}

func (s *twoFAService) EnableSMS2FA(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFAMethod != domain.TwoFANone && user.TwoFAMethod != "" {
		return domain.Err2FAAlreadyEnabled
	}

	if !user.PhoneVerified || user.Phone == "" {
		return domain.Err2FAPhoneRequired
	}

	user.TwoFAMethod = domain.TwoFASMS
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *twoFAService) SendSMS2FA(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFAMethod != domain.TwoFASMS {
		return domain.Err2FANotEnabled
	}

	return s.otpSvc.SendOTP(ctx, user.Phone)
}

func (s *twoFAService) VerifySMS2FA(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.TwoFAMethod != domain.TwoFASMS {
		return false, domain.Err2FANotEnabled
	}

	return s.otpSvc.VerifyOTP(ctx, user.Phone, code)
}
