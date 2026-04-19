package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email        string
	Password     string
	Name         string
	Phone        string
	Role         domain.UserRole   // client or owner only
	Region       domain.UserRegion // optional: if empty, detected from phone or defaults to RU
	ReferralCode string            // optional referral code from inviter
	AgeConfirmed bool              // required: user confirms they are 18+
}

type RegisterPhoneInput struct {
	Phone        string
	Name         string
	AgeConfirmed bool
}

type LoginResult struct {
	User        *domain.User
	Token       string
	Requires2FA bool
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
	ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error)
	ParsePartialToken(ctx context.Context, token string) (uuid.UUID, error)
	RegisterPhone(ctx context.Context, input RegisterPhoneInput) error
	LoginPhone(ctx context.Context, phone string) error
	VerifyPhone(ctx context.Context, phone string, code string, name string, ageConfirmed bool) (*LoginResult, error)
	Complete2FALogin(ctx context.Context, userID uuid.UUID) (*domain.User, string, error)
}

type authService struct {
	userRepo           repository.UserRepository
	referralSvc        ReferralService
	otpSvc             OTPService
	walletSvc          WalletService
	sessionSvc         SessionService
	logger             *logger.Logger
	jwtSecret          []byte
	tokenTTL           time.Duration
	welcomeBonusAmount   int64
	welcomeBonusAmountBY int64
	welcomeBonusExpiry   int
}

func NewAuthService(userRepo repository.UserRepository, referralSvc ReferralService, otpSvc OTPService, walletSvc WalletService, sessionSvc SessionService, cfg *config.Config, log *logger.Logger) AuthService {
	return &authService{
		userRepo:           userRepo,
		referralSvc:        referralSvc,
		otpSvc:             otpSvc,
		walletSvc:          walletSvc,
		sessionSvc:         sessionSvc,
		logger:             log,
		jwtSecret:          []byte(cfg.JWT.Secret),
		tokenTTL:           cfg.JWT.TokenTTL,
		welcomeBonusAmount:   cfg.WelcomeBonus.Amount,
		welcomeBonusAmountBY: cfg.WelcomeBonus.AmountBY,
		welcomeBonusExpiry:   cfg.WelcomeBonus.ExpiryDays,
	}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*domain.User, string, error) {
	if input.Role != domain.RoleClient && input.Role != domain.RoleOwner {
		return nil, "", fmt.Errorf("%w: can only register as client or owner", domain.ErrInvalidInput)
	}

	if !input.AgeConfirmed {
		return nil, "", fmt.Errorf("%w: age confirmation is required", domain.ErrInvalidInput)
	}

	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, "", domain.ErrInvalidInput
	}

	if !isValidEmail(input.Email) {
		return nil, "", fmt.Errorf("%w: invalid email format", domain.ErrInvalidInput)
	}

	if len(input.Password) < 8 {
		return nil, "", fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}
	if len(input.Password) > 72 {
		return nil, "", fmt.Errorf("%w: password must be at most 72 characters", domain.ErrInvalidInput)
	}

	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", domain.ErrAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine region: explicit input > phone detection > default RU
	region := input.Region
	if region == "" && input.Phone != "" {
		region = detectRegionFromPhone(input.Phone)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Phone:        input.Phone,
		Role:         input.Role,
		Region:       region,
		IsActive:     true,
		AgeConfirmed: true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Register referral if code provided (best-effort, don't fail registration)
	if input.ReferralCode != "" {
		if err := s.referralSvc.RegisterReferral(ctx, input.ReferralCode, user.ID); err != nil {
			s.logger.Warn("failed to register referral", "user_id", user.ID, "referral_code", input.ReferralCode, "error", err)
		}
	}

	// Create wallet and credit welcome bonus (best-effort, don't fail registration)
	s.createWalletWithBonus(ctx, user.ID, user.Region)

	token, err := s.generateTokenWithSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *authService) createWalletWithBonus(ctx context.Context, userID uuid.UUID, region domain.UserRegion) {
	if s.walletSvc == nil {
		return
	}

	currency := domain.CurrencyForRegion(region)
	wallet, err := s.walletSvc.CreateWallet(ctx, userID, currency)
	if err != nil {
		s.logger.Warn("failed to create wallet on registration", "user_id", userID, "error", err)
		return
	}

	bonusAmount := s.welcomeBonusAmount
	if region == domain.RegionBY {
		bonusAmount = s.welcomeBonusAmountBY
	}

	if bonusAmount > 0 {
		expiresAt := time.Now().AddDate(0, 0, s.welcomeBonusExpiry)
		_, err = s.walletSvc.AddBonus(ctx, wallet.ID, bonusAmount, domain.WalletTxWelcomeBonus, &expiresAt, "Приветственный бонус")
		if err != nil {
			s.logger.Warn("failed to credit welcome bonus", "user_id", userID, "wallet_id", wallet.ID, "error", err)
		}
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Verify password BEFORE revealing account status to prevent enumeration.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !user.IsActive {
		return nil, domain.ErrUserBlocked
	}

	if user.DeletionScheduledAt != nil {
		return nil, domain.ErrAccountDeletionPending
	}

	if user.TwoFAMethod != domain.TwoFANone && user.TwoFAMethod != "" {
		// For SMS 2FA, auto-send the OTP code so user can verify
		if user.TwoFAMethod == domain.TwoFASMS && user.Phone != "" {
			if err := s.otpSvc.SendOTP(ctx, user.Phone); err != nil {
				s.logger.Error("Failed to send SMS 2FA code during login", "user_id", user.ID, "error", err)
			}
		}
		partialToken, err := s.generatePartialToken(user.ID)
		if err != nil {
			return nil, err
		}
		return &LoginResult{
			User:        user,
			Token:       partialToken,
			Requires2FA: true,
		}, nil
	}

	token, err := s.generateTokenWithSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:  user,
		Token: token,
	}, nil
}

func (s *authService) ParseToken(_ context.Context, tokenString string) (uuid.UUID, domain.UserRole, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	role := domain.UserRole(roleStr)
	if !role.IsValid() {
		return uuid.Nil, "", domain.ErrUnauthorized
	}

	return userID, role, nil
}

func (s *authService) ParseTokenWithSession(_ context.Context, tokenString string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	role := domain.UserRole(roleStr)
	if !role.IsValid() {
		return uuid.Nil, "", uuid.Nil, domain.ErrUnauthorized
	}

	var sessionID uuid.UUID
	if sidStr, ok := claims["session_id"].(string); ok {
		if sid, err := uuid.Parse(sidStr); err == nil {
			sessionID = sid
		}
	}

	return userID, role, sessionID, nil
}

func (s *authService) RegisterPhone(ctx context.Context, input RegisterPhoneInput) error {
	phone := normalizePhone(input.Phone)
	if !isValidPhone(phone) {
		return domain.ErrPhoneInvalid
	}

	if input.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}

	if !input.AgeConfirmed {
		return fmt.Errorf("%w: age confirmation is required", domain.ErrInvalidInput)
	}

	existing, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if existing != nil {
		return domain.ErrAlreadyExists
	}

	return s.otpSvc.SendOTP(ctx, phone)
}

func (s *authService) LoginPhone(ctx context.Context, phone string) error {
	phone = normalizePhone(phone)
	if !isValidPhone(phone) {
		return domain.ErrPhoneInvalid
	}

	existing, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Return nil to prevent phone enumeration — caller sees success
			// but no OTP is sent for non-existent phones
			s.logger.Debug("login attempt for non-existent phone", "phone", phone)
			return nil
		}
		return err
	}
	if !existing.IsActive {
		// Return nil to prevent leaking that a blocked user exists at this phone
		s.logger.Debug("login attempt for blocked user", "phone", phone)
		return nil
	}

	return s.otpSvc.SendOTP(ctx, phone)
}

func (s *authService) VerifyPhone(ctx context.Context, phone string, code string, name string, ageConfirmed bool) (*LoginResult, error) {
	phone = normalizePhone(phone)
	if !isValidPhone(phone) {
		return nil, domain.ErrPhoneInvalid
	}

	valid, err := s.otpSvc.VerifyOTP(ctx, phone, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, domain.ErrOTPInvalid
	}

	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// New user registration via phone - create user account
			if name == "" {
				return nil, fmt.Errorf("%w: name is required for registration", domain.ErrInvalidInput)
			}
			if !ageConfirmed {
				return nil, fmt.Errorf("%w: age confirmation is required", domain.ErrInvalidInput)
			}
			now := time.Now()
			user = &domain.User{
				ID:            uuid.New(),
				Name:          name,
				Phone:         phone,
				PhoneVerified: true,
				Role:          domain.RoleClient,
				Region:        detectRegionFromPhone(phone),
				IsActive:      true,
				AgeConfirmed:  true,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, err
			}
			// Create wallet and credit welcome bonus (best-effort)
			s.createWalletWithBonus(ctx, user.ID, user.Region)
		} else {
			return nil, err
		}
	} else {
		// Existing user - check status
		if !user.IsActive {
			return nil, domain.ErrUserBlocked
		}

		if user.DeletionScheduledAt != nil {
			return nil, domain.ErrAccountDeletionPending
		}

		if !user.PhoneVerified {
			user.PhoneVerified = true
			if err := s.userRepo.Update(ctx, user); err != nil {
				s.logger.Error("Failed to update phone_verified", "user_id", user.ID, "error", err)
			}
		}
	}

	// For TOTP 2FA, require the second factor even for phone-based login.
	// For SMS 2FA, skip it -- phone possession was already proven by the OTP above.
	if user.TwoFAMethod == domain.TwoFATOTP {
		partialToken, err := s.generatePartialToken(user.ID)
		if err != nil {
			return nil, err
		}
		return &LoginResult{
			User:        user,
			Token:       partialToken,
			Requires2FA: true,
		}, nil
	}

	token, err := s.generateTokenWithSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:  user,
		Token: token,
	}, nil
}

func normalizePhone(phone string) string {
	var digits []byte
	for _, c := range []byte(phone) {
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		}
	}
	s := string(digits)
	if len(s) == 11 && s[0] == '8' {
		s = "7" + s[1:]
	}
	return "+" + s
}

// detectRegionFromPhone returns BY for Belarus phone numbers (+375), RU otherwise.
func detectRegionFromPhone(phone string) domain.UserRegion {
	if strings.HasPrefix(phone, "+375") {
		return domain.RegionBY
	}
	return domain.RegionRU
}

func isValidPhone(phone string) bool {
	if len(phone) < 2 || phone[0] != '+' {
		return false
	}
	digits := phone[1:]
	if len(digits) < 10 || len(digits) > 15 {
		return false
	}
	for _, c := range []byte(digits) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

const partialTokenTTL = 5 * time.Minute

func (s *authService) generateTokenWithSession(ctx context.Context, userID uuid.UUID, role domain.UserRole) (string, error) {
	if s.sessionSvc != nil {
		session, err := s.sessionSvc.CreateSession(ctx, userID, "", "", "")
		if err != nil {
			s.logger.Warn("failed to create session during token generation", "user_id", userID, "error", err)
			return generateJWT(userID, role, s.jwtSecret, s.tokenTTL)
		}
		return generateJWTWithSession(userID, role, session.ID, s.jwtSecret, s.tokenTTL)
	}
	return generateJWT(userID, role, s.jwtSecret, s.tokenTTL)
}

func (s *authService) generatePartialToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID.String(),
		"2fa_pending": true,
		"exp":         time.Now().Add(partialTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *authService) ParsePartialToken(_ context.Context, tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, domain.ErrUnauthorized
	}

	pending, ok := claims["2fa_pending"].(bool)
	if !ok || !pending {
		return uuid.Nil, domain.ErrUnauthorized
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	return userID, nil
}

func (s *authService) Complete2FALogin(ctx context.Context, userID uuid.UUID) (*domain.User, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	if !user.IsActive {
		return nil, "", domain.ErrUserBlocked
	}

	if user.DeletionScheduledAt != nil {
		return nil, "", domain.ErrAccountDeletionPending
	}

	token, err := s.generateTokenWithSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func isValidEmail(email string) bool {
	// Basic RFC 5322 compliant email validation
	// Must have at least one character, then @, then domain with at least one dot
	if len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}

	localPart := email[:at]
	domain := email[at+1:]

	// Local part validation (simplified - must not be empty or start/end with dot)
	if localPart == "" || localPart[0] == '.' || localPart[len(localPart)-1] == '.' {
		return false
	}

	// Domain validation (must contain at least one dot and end with valid TLD)
	dotIdx := strings.LastIndex(domain, ".")
	if dotIdx <= 0 || dotIdx >= len(domain)-1 {
		return false
	}

	// Check TLD length (at least 2 characters, max 63)
	tld := domain[dotIdx+1:]
	if len(tld) < 2 || len(tld) > 63 {
		return false
	}

	// Validate TLD contains only letters
	for _, c := range tld {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}

	return true
}
