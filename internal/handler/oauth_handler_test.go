package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

// --- Mock OAuth Service ---

type mockOAuthService struct {
	getOAuthURLFn         func(provider domain.OAuthProvider, referralCode string) (string, error)
	oauthCallbackFn       func(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, string, error)
	linkSocialAccountFn   func(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider, code string) error
	unlinkSocialAccountFn func(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
	listSocialAccountsFn  func(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error)
}

func (m *mockOAuthService) GetOAuthURL(provider domain.OAuthProvider, referralCode string) (string, error) {
	if m.getOAuthURLFn != nil {
		return m.getOAuthURLFn(provider, referralCode)
	}
	return "", nil
}

func (m *mockOAuthService) OAuthCallback(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, string, error) {
	if m.oauthCallbackFn != nil {
		return m.oauthCallbackFn(ctx, provider, code, state)
	}
	return nil, "", nil
}

func (m *mockOAuthService) LinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider, code string) error {
	if m.linkSocialAccountFn != nil {
		return m.linkSocialAccountFn(ctx, userID, provider, code)
	}
	return nil
}

func (m *mockOAuthService) UnlinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	if m.unlinkSocialAccountFn != nil {
		return m.unlinkSocialAccountFn(ctx, userID, provider)
	}
	return nil
}

func (m *mockOAuthService) ListSocialAccounts(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error) {
	if m.listSocialAccountsFn != nil {
		return m.listSocialAccountsFn(ctx, userID)
	}
	return nil, nil
}

// Compile-time check
var _ service.OAuthService = (*mockOAuthService)(nil)

// --- OAuth Handler Tests ---

func TestOAuthHandler_OAuthRedirect(t *testing.T) {
	oauthSvc := &mockOAuthService{
		getOAuthURLFn: func(provider domain.OAuthProvider, referralCode string) (string, error) {
			return "https://oauth.example.com/authorize?state=abc", nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.Get("/auth/oauth/{provider}", h.OAuthRedirect)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/vk", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", rec.Code)
	}

	location := rec.Result().Header.Get("Location")
	if location != "https://oauth.example.com/authorize?state=abc" {
		t.Errorf("expected redirect to oauth URL, got %q", location)
	}
}

func TestOAuthHandler_OAuthRedirect_InvalidProvider(t *testing.T) {
	oauthSvc := &mockOAuthService{
		getOAuthURLFn: func(provider domain.OAuthProvider, referralCode string) (string, error) {
			return "", domain.ErrInvalidInput
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.Get("/auth/oauth/{provider}", h.OAuthRedirect)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/invalid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestOAuthHandler_OAuthCallback(t *testing.T) {
	userID := uuid.New()
	oauthSvc := &mockOAuthService{
		oauthCallbackFn: func(_ context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, string, error) {
			if provider != domain.OAuthProviderVK {
				t.Errorf("expected provider vk, got %s", provider)
			}
			if code != "test-code" {
				t.Errorf("expected code test-code, got %s", code)
			}
			if state != "test-state" {
				t.Errorf("expected state test-state, got %s", state)
			}
			return &domain.User{
				ID:       userID,
				Email:    "user@example.com",
				Name:     "Test User",
				Role:     domain.RoleClient,
				IsActive: true,
			}, "jwt-token-123", nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.Get("/auth/oauth/{provider}/callback", h.OAuthCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/vk/callback?code=test-code&state=test-state", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if data["token"] != "jwt-token-123" {
		t.Errorf("expected token jwt-token-123, got %v", data["token"])
	}
}

func TestOAuthHandler_OAuthCallback_MissingCode(t *testing.T) {
	h := handler.NewOAuthHandler(&mockOAuthService{})

	router := chi.NewRouter()
	router.Get("/auth/oauth/{provider}/callback", h.OAuthCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/vk/callback", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_input" {
		t.Errorf("expected error code invalid_input, got %v", resp.Error)
	}
}

func TestOAuthHandler_OAuthCallback_ServiceError(t *testing.T) {
	oauthSvc := &mockOAuthService{
		oauthCallbackFn: func(_ context.Context, _ domain.OAuthProvider, _ string, _ string) (*domain.User, string, error) {
			return nil, "", domain.ErrUserBlocked
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.Get("/auth/oauth/{provider}/callback", h.OAuthCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/vk/callback?code=test-code&state=test-state", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestOAuthHandler_LinkSocialAccount(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	oauthSvc := &mockOAuthService{
		linkSocialAccountFn: func(_ context.Context, uid uuid.UUID, provider domain.OAuthProvider, code string) error {
			if uid != userID {
				t.Errorf("expected userID %s, got %s", userID, uid)
			}
			if provider != domain.OAuthProviderGoogle {
				t.Errorf("expected provider google, got %s", provider)
			}
			if code != "link-code" {
				t.Errorf("expected code link-code, got %s", code)
			}
			return nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/auth/link/{provider}", h.LinkSocialAccount)

	body := jsonBody(map[string]string{"code": "link-code"})
	req := httptest.NewRequest(http.MethodPost, "/auth/link/google", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestOAuthHandler_LinkSocialAccount_MissingCode(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)

	h := handler.NewOAuthHandler(&mockOAuthService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/auth/link/{provider}", h.LinkSocialAccount)

	body := jsonBody(map[string]string{"code": ""})
	req := httptest.NewRequest(http.MethodPost, "/auth/link/google", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestOAuthHandler_LinkSocialAccount_AlreadyLinked(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	oauthSvc := &mockOAuthService{
		linkSocialAccountFn: func(_ context.Context, _ uuid.UUID, _ domain.OAuthProvider, _ string) error {
			return domain.ErrSocialAccountAlreadyLinked
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/auth/link/{provider}", h.LinkSocialAccount)

	body := jsonBody(map[string]string{"code": "some-code"})
	req := httptest.NewRequest(http.MethodPost, "/auth/link/vk", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestOAuthHandler_UnlinkSocialAccount(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	oauthSvc := &mockOAuthService{
		unlinkSocialAccountFn: func(_ context.Context, uid uuid.UUID, provider domain.OAuthProvider) error {
			if uid != userID {
				t.Errorf("expected userID %s, got %s", userID, uid)
			}
			if provider != domain.OAuthProviderYandex {
				t.Errorf("expected provider yandex, got %s", provider)
			}
			return nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/auth/link/{provider}", h.UnlinkSocialAccount)

	req := httptest.NewRequest(http.MethodDelete, "/auth/link/yandex", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestOAuthHandler_UnlinkSocialAccount_LastMethod(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	oauthSvc := &mockOAuthService{
		unlinkSocialAccountFn: func(_ context.Context, _ uuid.UUID, _ domain.OAuthProvider) error {
			return domain.ErrInvalidInput
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/auth/link/{provider}", h.UnlinkSocialAccount)

	req := httptest.NewRequest(http.MethodDelete, "/auth/link/vk", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestOAuthHandler_ListSocialAccounts(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	accountID1 := uuid.New()
	accountID2 := uuid.New()
	linkedAt := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	oauthSvc := &mockOAuthService{
		listSocialAccountsFn: func(_ context.Context, uid uuid.UUID) ([]domain.SocialAccount, error) {
			if uid != userID {
				t.Errorf("expected userID %s, got %s", userID, uid)
			}
			return []domain.SocialAccount{
				{
					ID:         accountID1,
					UserID:     userID,
					Provider:   domain.OAuthProviderVK,
					ProviderID: "vk-123",
					Email:      "user@vk.com",
					Name:       "VK User",
					AvatarURL:  "https://vk.com/avatar.jpg",
					LinkedAt:   linkedAt,
				},
				{
					ID:         accountID2,
					UserID:     userID,
					Provider:   domain.OAuthProviderGoogle,
					ProviderID: "google-456",
					Email:      "user@gmail.com",
					Name:       "Google User",
					AvatarURL:  "https://google.com/avatar.jpg",
					LinkedAt:   linkedAt,
				},
			}, nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/auth/me/social-accounts", h.ListSocialAccounts)

	req := httptest.NewRequest(http.MethodGet, "/auth/me/social-accounts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	var accounts []map[string]interface{}
	if err := json.Unmarshal(resp.Data, &accounts); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
	if accounts[0]["provider"] != "vk" {
		t.Errorf("expected first provider vk, got %v", accounts[0]["provider"])
	}
	if accounts[1]["provider"] != "google" {
		t.Errorf("expected second provider google, got %v", accounts[1]["provider"])
	}
}

func TestOAuthHandler_ListSocialAccounts_Empty(t *testing.T) {
	userID := uuid.New()
	authSvc := makeAuthToken(userID, domain.RoleClient)
	oauthSvc := &mockOAuthService{
		listSocialAccountsFn: func(_ context.Context, _ uuid.UUID) ([]domain.SocialAccount, error) {
			return []domain.SocialAccount{}, nil
		},
	}

	h := handler.NewOAuthHandler(oauthSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/auth/me/social-accounts", h.ListSocialAccounts)

	req := httptest.NewRequest(http.MethodGet, "/auth/me/social-accounts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	var accounts []map[string]interface{}
	if err := json.Unmarshal(resp.Data, &accounts); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}
}
