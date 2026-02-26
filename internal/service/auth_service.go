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
	"github.com/nikitaaldaev/bani/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
	Role     domain.UserRole // client or owner only
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(cfg.JWT.Secret),
		tokenTTL:  cfg.JWT.TokenTTL,
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
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func isValidEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}
