package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	resetTokenTTL    = 30 * time.Minute
	// Per-email reset cap. Was 3/15min; relaxed 5x in 2026-04 because
	// operators were tripping it during normal password resets.
	resetRateLimit  = 15
	resetRateWindow = 15 * time.Minute
	resetTokenBytes  = 32
	resetRedisPrefix = "password_reset:"
	resetRatePrefix  = "password_reset_rate:"
)

type PasswordResetService interface {
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
}

type passwordResetService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	redis       *redis.Client
	emailSender notification.EmailSender
	logger      *logger.Logger
	frontendURL string
}

func NewPasswordResetService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	redisClient *redis.Client,
	emailSender notification.EmailSender,
	log *logger.Logger,
	frontendURL string,
) PasswordResetService {
	return &passwordResetService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		redis:       redisClient,
		emailSender: emailSender,
		logger:      log,
		frontendURL: frontendURL,
	}
}

func (s *passwordResetService) ForgotPassword(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}

	// Rate limiting per email
	rateKey := resetRatePrefix + email
	count, err := s.redis.Get(ctx, rateKey).Int64()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("check reset rate limit: %w", err)
	}
	if count >= int64(resetRateLimit) {
		return domain.ErrResetRateLimited
	}

	// Look up user - always return success to prevent email enumeration
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Increment rate limit even for non-existent emails to prevent
		// unlimited probing and reduce timing differences
		pipe := s.redis.Pipeline()
		pipe.Incr(ctx, rateKey)
		pipe.ExpireNX(ctx, rateKey, resetRateWindow)
		_, _ = pipe.Exec(ctx) //nolint:errcheck // best-effort rate limit for non-existent email
		s.logger.Debug("password reset requested for non-existent email", "email", email)
		return nil
	}

	// Generate secure token
	tokenBytes := make([]byte, resetTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Store token in Redis
	redisKey := resetRedisPrefix + token
	if err := s.redis.Set(ctx, redisKey, user.ID.String(), resetTokenTTL).Err(); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	// Increment rate limit counter
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, rateKey)
	pipe.ExpireNX(ctx, rateKey, resetRateWindow)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("update reset rate limit: %w", err)
	}

	// Send reset email
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)
	subject := "Сброс пароля"
	body := fmt.Sprintf(
		"Здравствуйте, %s!\n\nВы запросили сброс пароля. Перейдите по ссылке для установки нового пароля:\n\n%s\n\nСсылка действительна 30 минут.\n\nЕсли вы не запрашивали сброс пароля, проигнорируйте это письмо.",
		user.Name, resetLink,
	)

	if err := s.emailSender.Send(ctx, email, subject, body); err != nil {
		s.logger.Error("failed to send password reset email", "email", email, "error", err)
		// Don't return error to user - token is already stored, email delivery is best-effort
	}

	return nil
}

func (s *passwordResetService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	if token == "" {
		return domain.ErrResetTokenInvalid
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}
	if len(newPassword) > 72 {
		return fmt.Errorf("%w: password must be at most 72 characters", domain.ErrInvalidInput)
	}

	// Atomically retrieve and delete token from Redis to prevent replay attacks.
	// GetDel ensures no concurrent request can read the same token.
	redisKey := resetRedisPrefix + token
	userIDStr, err := s.redis.GetDel(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return domain.ErrResetTokenInvalid
		}
		return fmt.Errorf("get reset token: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return domain.ErrResetTokenInvalid
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrResetTokenInvalid
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Terminate all active sessions
	if err := s.sessionRepo.DeleteAllByUser(ctx, userID); err != nil {
		s.logger.Error("failed to terminate sessions after password reset", "user_id", userID, "error", err)
	}

	return nil
}
