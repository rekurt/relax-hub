package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email        string
	Password     string
	Name         string
	Phone        string
	Role         domain.UserRole // client or owner only
	ReferralCode string          // optional referral code from inviter
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

type authService struct {
	userRepo    repository.UserRepository
	referralSvc ReferralService
	logger      *logger.Logger
	jwtSecret   []byte
	tokenTTL    time.Duration
}

func NewAuthService(userRepo repository.UserRepository, referralSvc ReferralService, cfg *config.Config, log *logger.Logger) AuthService {
	return &authService{
		userRepo:    userRepo,
		referralSvc: referralSvc,
		logger:      log,
		jwtSecret:   []byte(cfg.JWT.Secret),
		tokenTTL:    cfg.JWT.TokenTTL,
	}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*domain.User, string, error) {
	if input.Role != domain.RoleClient && input.Role != domain.RoleOwner {
		return nil, "", fmt.Errorf("%w: can only register as client or owner", domain.ErrInvalidInput)
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

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Phone:        input.Phone,
		Role:         input.Role,
		IsActive:     true,
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

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrUnauthorized
	}

	if !user.IsActive {
		return nil, "", domain.ErrUserBlocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", domain.ErrUnauthorized
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
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

func (s *authService) generateToken(userID uuid.UUID, role domain.UserRole) (string, error) {
	return generateJWT(userID, role, s.jwtSecret, s.tokenTTL)
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
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}

	return true
}
