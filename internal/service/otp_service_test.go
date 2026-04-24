package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/redis/go-redis/v9"
)

func newTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a separate DB for testing
	})
}

type mockSMSProvider struct {
	sentMessages []struct {
		Phone   string
		Message string
	}
	err error
}

func (m *mockSMSProvider) SendSMS(_ context.Context, phone string, message string) error {
	if m.err != nil {
		return m.err
	}
	m.sentMessages = append(m.sentMessages, struct {
		Phone   string
		Message string
	}{phone, message})
	return nil
}

func TestOTPService_SendOTP_Success(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	err := otpSvc.SendOTP(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(smsMock.sentMessages) != 1 {
		t.Fatalf("expected 1 SMS sent, got %d", len(smsMock.sentMessages))
	}
	if smsMock.sentMessages[0].Phone != "+79001234567" {
		t.Errorf("SMS sent to wrong phone: %s", smsMock.sentMessages[0].Phone)
	}
}

func TestOTPService_VerifyOTP_Success(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	phone := "+79001234567"
	_ = otpSvc.SendOTP(context.Background(), phone)

	// Extract the code from the SMS message
	if len(smsMock.sentMessages) != 1 {
		t.Fatal("expected 1 SMS sent")
	}
	msg := smsMock.sentMessages[0].Message
	// Format: "Ваш код подтверждения: XXXXXX"
	code := msg[len(msg)-6:]

	valid, err := otpSvc.VerifyOTP(context.Background(), phone, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected OTP to be valid")
	}
}

func TestOTPService_VerifyOTP_WrongCode(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	phone := "+79001234567"
	_ = otpSvc.SendOTP(context.Background(), phone)

	_, err := otpSvc.VerifyOTP(context.Background(), phone, "000000")
	if !errors.Is(err, domain.ErrOTPInvalid) {
		t.Errorf("expected ErrOTPInvalid, got: %v", err)
	}
}

func TestOTPService_VerifyOTP_Expired(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	// Try to verify without sending OTP
	_, err := otpSvc.VerifyOTP(context.Background(), "+79001234567", "123456")
	if !errors.Is(err, domain.ErrOTPExpired) {
		t.Errorf("expected ErrOTPExpired, got: %v", err)
	}
}

func TestOTPService_VerifyOTP_MaxAttempts(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	phone := "+79001234567"
	_ = otpSvc.SendOTP(context.Background(), phone)

	// Use wrong code 3 times
	for i := 0; i < 3; i++ {
		otpSvc.VerifyOTP(context.Background(), phone, "000000")
	}

	// 4th attempt should return max attempts error
	_, err := otpSvc.VerifyOTP(context.Background(), phone, "000000")
	if !errors.Is(err, domain.ErrOTPMaxAttempts) {
		t.Errorf("expected ErrOTPMaxAttempts, got: %v", err)
	}
}

func TestOTPService_RateLimit(t *testing.T) {
	rdb := newTestRedis()
	defer rdb.FlushDB(context.Background())

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping OTP integration test")
	}

	smsMock := &mockSMSProvider{}
	otpSvc := service.NewOTPService(rdb, smsMock, logger.New(logger.LevelWarn))

	phone := "+79001234567"

	// Send up to the rate limit (otpRateLimit = 15 in the service package).
	const limit = 15
	for i := 0; i < limit; i++ {
		if err := otpSvc.SendOTP(context.Background(), phone); err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i+1, err)
		}
	}

	// Next call should be rate limited.
	err := otpSvc.SendOTP(context.Background(), phone)
	if !errors.Is(err, domain.ErrOTPRateLimited) {
		t.Errorf("expected ErrOTPRateLimited, got: %v", err)
	}
}
