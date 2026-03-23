package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/sms"
	"github.com/redis/go-redis/v9"
)

const (
	otpTTL         = 5 * time.Minute
	otpMaxAttempts = 3
	otpRateLimit   = 3
	otpRateWindow  = 15 * time.Minute
)

type OTPService interface {
	SendOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone string, code string) (bool, error)
}

type otpData struct {
	Code     string `json:"code"`
	Attempts int    `json:"attempts"`
}

type otpService struct {
	redis       *redis.Client
	smsProvider sms.Provider
	logger      *logger.Logger
}

func NewOTPService(redisClient *redis.Client, smsProvider sms.Provider, log *logger.Logger) OTPService {
	return &otpService{
		redis:       redisClient,
		smsProvider: smsProvider,
		logger:      log,
	}
}

func (s *otpService) SendOTP(ctx context.Context, phone string) error {
	rateKey := fmt.Sprintf("otp_rate:%s", phone)
	count, err := s.redis.Get(ctx, rateKey).Int64()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("check OTP rate limit: %w", err)
	}
	if count >= int64(otpRateLimit) {
		return domain.ErrOTPRateLimited
	}

	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("generate OTP code: %w", err)
	}

	data := otpData{Code: code, Attempts: 0}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal OTP data: %w", err)
	}

	otpKey := fmt.Sprintf("otp:%s", phone)
	if err := s.redis.Set(ctx, otpKey, dataBytes, otpTTL).Err(); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}

	// Use Incr + ExpireNX so the window TTL is only set on first request (fixed window)
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, rateKey)
	pipe.ExpireNX(ctx, rateKey, otpRateWindow)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("update OTP rate limit: %w", err)
	}

	message := fmt.Sprintf("Ваш код подтверждения: %s", code)
	if err := s.smsProvider.SendSMS(ctx, phone, message); err != nil {
		maskedPhone := phone
		if len(phone) > 4 {
			maskedPhone = "***" + phone[len(phone)-4:]
		}
		s.logger.Error("Failed to send OTP SMS", "phone", maskedPhone, "error", err)
		return fmt.Errorf("send OTP SMS: %w", err)
	}

	return nil
}

func (s *otpService) VerifyOTP(ctx context.Context, phone string, code string) (bool, error) {
	otpKey := fmt.Sprintf("otp:%s", phone)
	dataBytes, err := s.redis.Get(ctx, otpKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, domain.ErrOTPExpired
		}
		return false, fmt.Errorf("get OTP data: %w", err)
	}

	var data otpData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return false, fmt.Errorf("unmarshal OTP data: %w", err)
	}

	if data.Attempts >= otpMaxAttempts {
		s.redis.Del(ctx, otpKey)
		return false, domain.ErrOTPMaxAttempts
	}

	if subtle.ConstantTimeCompare([]byte(data.Code), []byte(code)) != 1 {
		data.Attempts++
		updated, _ := json.Marshal(data)
		ttl, _ := s.redis.TTL(ctx, otpKey).Result()
		if ttl > 0 {
			s.redis.Set(ctx, otpKey, updated, ttl)
		}
		return false, domain.ErrOTPInvalid
	}

	s.redis.Del(ctx, otpKey)
	return true, nil
}

func generateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
