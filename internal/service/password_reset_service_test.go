package service_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type recordingEmailSender struct {
	sent []sentEmail
}

type sentEmail struct {
	To      string
	Subject string
	Body    string
}

func (r *recordingEmailSender) Send(_ context.Context, to, subject, body string) error {
	r.sent = append(r.sent, sentEmail{To: to, Subject: subject, Body: body})
	return nil
}

func newPasswordResetService(t *testing.T) (service.PasswordResetService, *mock.UserRepo, *mock.SessionRepo, *recordingEmailSender, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	userRepo := mock.NewUserRepo()
	sessionRepo := mock.NewSessionRepo().(*mock.SessionRepo)
	emailSender := &recordingEmailSender{}
	log := logger.New(logger.LevelWarn)

	svc := service.NewPasswordResetService(userRepo, sessionRepo, redisClient, emailSender, log, "http://localhost:3000")
	return svc, userRepo, sessionRepo, emailSender, mr
}

func createResetTestUser(t *testing.T, userRepo *mock.UserRepo, email string) *domain.User {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         "Test User",
		Role:         domain.RoleClient,
		IsActive:     true,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestPasswordResetService_ForgotPassword_Success(t *testing.T) {
	svc, userRepo, _, emailSender, _ := newPasswordResetService(t)
	createResetTestUser(t, userRepo, "test@example.com")

	err := svc.ForgotPassword(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emailSender.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(emailSender.sent))
	}
	if emailSender.sent[0].To != "test@example.com" {
		t.Errorf("email sent to %q, want %q", emailSender.sent[0].To, "test@example.com")
	}
}

func TestPasswordResetService_ForgotPassword_NonExistentEmail(t *testing.T) {
	svc, _, _, emailSender, _ := newPasswordResetService(t)

	// Should not return error (prevents email enumeration)
	err := svc.ForgotPassword(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("should not return error for non-existent email, got: %v", err)
	}

	if len(emailSender.sent) != 0 {
		t.Errorf("should not send email for non-existent user, got %d emails", len(emailSender.sent))
	}
}

func TestPasswordResetService_ForgotPassword_EmptyEmail(t *testing.T) {
	svc, _, _, _, _ := newPasswordResetService(t)

	err := svc.ForgotPassword(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestPasswordResetService_ForgotPassword_RateLimit(t *testing.T) {
	svc, userRepo, _, _, _ := newPasswordResetService(t)
	createResetTestUser(t, userRepo, "test@example.com")

	// Exhaust rate limit (3 requests)
	for i := 0; i < 3; i++ {
		if err := svc.ForgotPassword(context.Background(), "test@example.com"); err != nil {
			t.Fatalf("request %d should succeed, got: %v", i+1, err)
		}
	}

	// 4th request should be rate limited
	err := svc.ForgotPassword(context.Background(), "test@example.com")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if err != domain.ErrResetRateLimited {
		t.Errorf("expected ErrResetRateLimited, got: %v", err)
	}
}

func TestPasswordResetService_ResetPassword_Success(t *testing.T) {
	svc, userRepo, sessionRepo, _, mr := newPasswordResetService(t)
	user := createResetTestUser(t, userRepo, "test@example.com")

	// Create a session for this user
	sessionRepo.Create(context.Background(), &domain.Session{
		ID:     uuid.New(),
		UserID: user.ID,
	})

	// Simulate a stored reset token in Redis
	token := "abc123token"
	mr.Set("password_reset:"+token, user.ID.String())
	mr.SetTTL("password_reset:"+token, 1800) // 30 min

	err := svc.ResetPassword(context.Background(), token, "newpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify password was changed
	updatedUser, _ := userRepo.GetByID(context.Background(), user.ID)
	if err := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword123")); err != nil {
		t.Error("password was not updated correctly")
	}

	// Verify old password no longer works
	if err := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("oldpassword")); err == nil {
		t.Error("old password should not work anymore")
	}

	// Verify sessions were terminated
	sessions, _ := sessionRepo.ListByUser(context.Background(), user.ID)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after reset, got %d", len(sessions))
	}

	// Verify token was consumed (can't reuse)
	if mr.Exists("password_reset:" + token) {
		t.Error("reset token should be deleted after use")
	}
}

func TestPasswordResetService_ResetPassword_InvalidToken(t *testing.T) {
	svc, _, _, _, _ := newPasswordResetService(t)

	err := svc.ResetPassword(context.Background(), "invalid-token", "newpassword123")
	if err != domain.ErrResetTokenInvalid {
		t.Errorf("expected ErrResetTokenInvalid, got: %v", err)
	}
}

func TestPasswordResetService_ResetPassword_EmptyToken(t *testing.T) {
	svc, _, _, _, _ := newPasswordResetService(t)

	err := svc.ResetPassword(context.Background(), "", "newpassword123")
	if err != domain.ErrResetTokenInvalid {
		t.Errorf("expected ErrResetTokenInvalid, got: %v", err)
	}
}

func TestPasswordResetService_ResetPassword_ShortPassword(t *testing.T) {
	svc, userRepo, _, _, mr := newPasswordResetService(t)
	user := createResetTestUser(t, userRepo, "test@example.com")

	token := "abc123token"
	mr.Set("password_reset:"+token, user.ID.String())

	err := svc.ResetPassword(context.Background(), token, "short")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestPasswordResetService_ResetPassword_TokenCannotBeReused(t *testing.T) {
	svc, userRepo, _, _, mr := newPasswordResetService(t)
	user := createResetTestUser(t, userRepo, "test@example.com")

	token := "abc123token"
	mr.Set("password_reset:"+token, user.ID.String())

	// First reset should succeed
	err := svc.ResetPassword(context.Background(), token, "newpassword123")
	if err != nil {
		t.Fatalf("first reset should succeed, got: %v", err)
	}

	// Second reset with same token should fail
	err = svc.ResetPassword(context.Background(), token, "anotherpassword")
	if err != domain.ErrResetTokenInvalid {
		t.Errorf("expected ErrResetTokenInvalid on reuse, got: %v", err)
	}
}

func TestPasswordResetService_ForgotPassword_ThenReset_Integration(t *testing.T) {
	svc, userRepo, _, emailSender, mr := newPasswordResetService(t)
	createResetTestUser(t, userRepo, "integration@example.com")

	// Step 1: Request password reset
	err := svc.ForgotPassword(context.Background(), "integration@example.com")
	if err != nil {
		t.Fatalf("forgot password failed: %v", err)
	}

	if len(emailSender.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(emailSender.sent))
	}

	// Extract token from Redis (find the key)
	keys := mr.Keys()
	var resetToken string
	for _, k := range keys {
		if len(k) > len("password_reset:") && k[:15] == "password_reset:" {
			resetToken = k[15:]
			break
		}
	}
	if resetToken == "" {
		t.Fatal("no reset token found in Redis")
	}

	// Step 2: Reset password with extracted token
	err = svc.ResetPassword(context.Background(), resetToken, "brandnewpassword")
	if err != nil {
		t.Fatalf("reset password failed: %v", err)
	}
}
