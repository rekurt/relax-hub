package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

type phoneAuthMock struct {
	registerPhoneFn func(ctx context.Context, input service.RegisterPhoneInput) error
	loginPhoneFn    func(ctx context.Context, phone string) error
	verifyPhoneFn   func(ctx context.Context, phone, code string) (*service.LoginResult, error)
}

func (m *phoneAuthMock) Register(_ context.Context, _ service.RegisterInput) (*domain.User, string, error) {
	return nil, "", nil
}
func (m *phoneAuthMock) Login(_ context.Context, _, _ string) (*service.LoginResult, error) {
	return &service.LoginResult{}, nil
}
func (m *phoneAuthMock) ParseToken(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
	return uuid.Nil, "", domain.ErrUnauthorized
}
func (m *phoneAuthMock) ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	userID, role, err := m.ParseToken(ctx, token)
	return userID, role, uuid.Nil, err
}
func (m *phoneAuthMock) ParsePartialToken(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, domain.ErrUnauthorized
}
func (m *phoneAuthMock) Complete2FALogin(_ context.Context, _ uuid.UUID) (*domain.User, string, error) {
	return nil, "", nil
}
func (m *phoneAuthMock) RegisterPhone(ctx context.Context, input service.RegisterPhoneInput) error {
	if m.registerPhoneFn != nil {
		return m.registerPhoneFn(ctx, input)
	}
	return nil
}
func (m *phoneAuthMock) LoginPhone(ctx context.Context, phone string) error {
	if m.loginPhoneFn != nil {
		return m.loginPhoneFn(ctx, phone)
	}
	return nil
}
func (m *phoneAuthMock) VerifyPhone(ctx context.Context, phone, code string) (*service.LoginResult, error) {
	if m.verifyPhoneFn != nil {
		return m.verifyPhoneFn(ctx, phone, code)
	}
	return &service.LoginResult{}, nil
}

// noopTwoFAServicePhone is a no-op TwoFAService for phone auth tests.
type noopTwoFAServicePhone struct{}

func (n *noopTwoFAServicePhone) GenerateTOTPSecret(_ context.Context, _ uuid.UUID) (string, string, error) {
	return "", "", nil
}
func (n *noopTwoFAServicePhone) EnableTOTP(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (n *noopTwoFAServicePhone) DisableTOTP(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (n *noopTwoFAServicePhone) VerifyTOTP(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}
func (n *noopTwoFAServicePhone) EnableSMS2FA(_ context.Context, _ uuid.UUID) error  { return nil }
func (n *noopTwoFAServicePhone) SendSMS2FA(_ context.Context, _ uuid.UUID) error    { return nil }
func (n *noopTwoFAServicePhone) VerifySMS2FA(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}

func TestRegisterPhone_Success(t *testing.T) {
	authSvc := &phoneAuthMock{}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/register-phone", h.RegisterPhone)

	body := `{"phone":"+79001234567","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterPhone_AlreadyExists(t *testing.T) {
	authSvc := &phoneAuthMock{
		registerPhoneFn: func(_ context.Context, _ service.RegisterPhoneInput) error {
			return domain.ErrAlreadyExists
		},
	}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/register-phone", h.RegisterPhone)

	body := `{"phone":"+79001234567","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginPhone_Success(t *testing.T) {
	authSvc := &phoneAuthMock{}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/login-phone", h.LoginPhone)

	body := `{"phone":"+79001234567"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginPhone_Unauthorized(t *testing.T) {
	authSvc := &phoneAuthMock{
		loginPhoneFn: func(_ context.Context, _ string) error {
			return domain.ErrUnauthorized
		},
	}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/login-phone", h.LoginPhone)

	body := `{"phone":"+79001234567"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyPhone_Success(t *testing.T) {
	testUser := &domain.User{
		ID:            uuid.New(),
		Name:          "Test User",
		Phone:         "+79001234567",
		PhoneVerified: true,
		Role:          domain.RoleClient,
		IsActive:      true,
	}
	authSvc := &phoneAuthMock{
		verifyPhoneFn: func(_ context.Context, _, _ string) (*service.LoginResult, error) {
			return &service.LoginResult{User: testUser, Token: "test-jwt-token"}, nil
		},
	}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/verify-phone", h.VerifyPhone)

	body := `{"phone":"+79001234567","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyPhone_InvalidOTP(t *testing.T) {
	authSvc := &phoneAuthMock{
		verifyPhoneFn: func(_ context.Context, _, _ string) (*service.LoginResult, error) {
			return nil, domain.ErrOTPInvalid
		},
	}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/verify-phone", h.VerifyPhone)

	body := `{"phone":"+79001234567","code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyPhone_RateLimited(t *testing.T) {
	authSvc := &phoneAuthMock{
		verifyPhoneFn: func(_ context.Context, _, _ string) (*service.LoginResult, error) {
			return nil, domain.ErrOTPMaxAttempts
		},
	}
	h := NewAuthHandler(authSvc, nil, &noopTwoFAServicePhone{})

	r := chi.NewRouter()
	r.Post("/auth/verify-phone", h.VerifyPhone)

	body := `{"phone":"+79001234567","code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-phone", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d: %s", rec.Code, rec.Body.String())
	}
}
