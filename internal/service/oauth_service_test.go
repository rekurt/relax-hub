package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/nikitaaldaev/bani/internal/auth"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

// mockOAuthProvider is a test double for auth.OAuthProvider.
type mockOAuthProvider struct {
	authURL  string
	userInfo *auth.OAuthUserInfo
	err      error
}

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) Exchange(_ context.Context, _ string) (*auth.OAuthUserInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	cp := *m.userInfo
	return &cp, nil
}

func newTestOAuthService(t *testing.T, userRepo *mock.UserRepo, socialRepo *mock.SocialAccountRepo, providers map[domain.OAuthProvider]auth.OAuthProvider) (service.OAuthService, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Pre-populate Redis with test state values for backward compatibility
	// Tests can use uuid.New().String() as the state parameter if needed
	testState := "test-state-12345"
	for provider := range providers {
		if err := mr.Set("oauth_state:"+testState, string(provider)); err != nil {
			panic(err) // fail test setup
		}
	}

	return service.NewOAuthServiceWithProviders(
		userRepo, socialRepo, redisClient, providers,
		nil,
		"test-secret-key-for-testing",
		time.Hour,
	), mr
}

func TestOAuthService_GetOAuthURL_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{authURL: "https://oauth.vk.com/authorize"},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	url, err := svc.GetOAuthURL(domain.OAuthProviderVK, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("URL should not be empty")
	}
}

func TestOAuthService_GetOAuthURL_InvalidProvider(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, nil)
	defer mr.Close()

	_, err := svc.GetOAuthURL("invalid", "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestOAuthService_GetOAuthURL_ProviderNotConfigured(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	providers := map[domain.OAuthProvider]auth.OAuthProvider{}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	_, err := svc.GetOAuthURL(domain.OAuthProviderGoogle, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestOAuthService_OAuthCallback_FirstLogin_CreatesUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "12345",
				Email:      "vk@example.com",
				Name:       "VK User",
				AvatarURL:  "https://vk.com/photo.jpg",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	user, token, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderVK, "test-code", "test-state-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if user.Email != "vk@example.com" {
		t.Errorf("email = %q, want %q", user.Email, "vk@example.com")
	}
	if user.Name != "VK User" {
		t.Errorf("name = %q, want %q", user.Name, "VK User")
	}
	if user.Role != domain.RoleClient {
		t.Errorf("role = %q, want %q", user.Role, domain.RoleClient)
	}

	// Verify social account was created
	accounts, err := socialRepo.ListByUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("list social accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 social account, got %d", len(accounts))
	}
	if accounts[0].ProviderID != "12345" {
		t.Errorf("provider_id = %q, want %q", accounts[0].ProviderID, "12345")
	}
}

func TestOAuthService_OAuthCallback_ReturningUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	// Pre-create user and social account
	existingUser := &domain.User{
		ID:       uuid.New(),
		Email:    "existing@example.com",
		Name:     "Existing User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), existingUser)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     existingUser.ID,
		Provider:   domain.OAuthProviderGoogle,
		ProviderID: "google-id-123",
		Email:      "existing@example.com",
		LinkedAt:   time.Now(),
	})

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderGoogle: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "google-id-123",
				Email:      "existing@example.com",
				Name:       "Existing User",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	user, token, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderGoogle, "test-code", "test-state-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != existingUser.ID {
		t.Errorf("user ID = %v, want %v", user.ID, existingUser.ID)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestOAuthService_OAuthCallback_BlockedUser(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	blockedUser := &domain.User{
		ID:       uuid.New(),
		Email:    "blocked@example.com",
		Name:     "Blocked User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), blockedUser)
	_ = userRepo.SetActive(context.Background(), blockedUser.ID, false)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     blockedUser.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "blocked-vk-id",
		LinkedAt:   time.Now(),
	})

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "blocked-vk-id",
				Email:      "blocked@example.com",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	_, _, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderVK, "test-code", "test-state-12345")
	if !errors.Is(err, domain.ErrUserBlocked) {
		t.Errorf("expected ErrUserBlocked, got: %v", err)
	}
}

func TestOAuthService_OAuthCallback_ExistingEmailLinksAccount(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	// User exists with email but no social account
	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "shared@example.com",
		PasswordHash: "hashed-password",
		Name:         "Shared Email User",
		Role:         domain.RoleClient,
		IsActive:     true,
	}
	_ = userRepo.Create(context.Background(), existingUser)

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderYandex: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "yandex-id-456",
				Email:      "shared@example.com",
				Name:       "Yandex User",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	user, _, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderYandex, "test-code", "test-state-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should link to existing user, not create new one
	if user.ID != existingUser.ID {
		t.Errorf("should link to existing user, got different ID: %v vs %v", user.ID, existingUser.ID)
	}
}

func TestOAuthService_OAuthCallback_BlockedUser_EmailMatch(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	// User exists with email but is blocked, no social account linked
	blockedUser := &domain.User{
		ID:       uuid.New(),
		Email:    "blocked@example.com",
		Name:     "Blocked User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), blockedUser)
	_ = userRepo.SetActive(context.Background(), blockedUser.ID, false)

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderGoogle: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "new-google-id",
				Email:      "blocked@example.com",
				Name:       "Google User",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	_, _, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderGoogle, "test-code", "test-state-12345")
	if !errors.Is(err, domain.ErrUserBlocked) {
		t.Errorf("expected ErrUserBlocked for blocked user via email match, got: %v", err)
	}
}

func TestOAuthService_LinkSocialAccount_Success(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		Role:         domain.RoleClient,
		IsActive:     true,
	}
	_ = userRepo.Create(context.Background(), user)

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderGoogle: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "new-google-id",
				Email:      "user@google.com",
				Name:       "Google User",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	err := svc.LinkSocialAccount(context.Background(), user.ID, domain.OAuthProviderGoogle, "test-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accounts, _ := socialRepo.ListByUser(context.Background(), user.ID)
	if len(accounts) != 1 {
		t.Fatalf("expected 1 social account, got %d", len(accounts))
	}
}

func TestOAuthService_LinkSocialAccount_AlreadyLinked(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "existing-vk-id",
		LinkedAt:   time.Now(),
	})

	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "existing-vk-id",
				Email:      "vk@example.com",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	err := svc.LinkSocialAccount(context.Background(), user.ID, domain.OAuthProviderVK, "test-code")
	if !errors.Is(err, domain.ErrSocialAccountAlreadyLinked) {
		t.Errorf("expected ErrSocialAccountAlreadyLinked, got: %v", err)
	}
}

func TestOAuthService_UnlinkSocialAccount_WithPassword(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hashed-password",
		Name:         "Test User",
		Role:         domain.RoleClient,
		IsActive:     true,
	}
	_ = userRepo.Create(context.Background(), user)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-id",
		LinkedAt:   time.Now(),
	})

	svc, mr := newTestOAuthService(t, userRepo, socialRepo, nil)
	defer mr.Close()

	err := svc.UnlinkSocialAccount(context.Background(), user.ID, domain.OAuthProviderVK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accounts, _ := socialRepo.ListByUser(context.Background(), user.ID)
	if len(accounts) != 0 {
		t.Errorf("expected 0 social accounts, got %d", len(accounts))
	}
}

func TestOAuthService_UnlinkSocialAccount_WithOtherProvider(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	// User without password but with two social accounts
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-id",
		LinkedAt:   time.Now(),
	})
	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderGoogle,
		ProviderID: "google-id",
		LinkedAt:   time.Now(),
	})

	svc, mr := newTestOAuthService(t, userRepo, socialRepo, nil)
	defer mr.Close()

	err := svc.UnlinkSocialAccount(context.Background(), user.ID, domain.OAuthProviderVK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOAuthService_UnlinkSocialAccount_LastAuthMethod(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	// User without password and only one social account
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-id",
		LinkedAt:   time.Now(),
	})

	svc, mr := newTestOAuthService(t, userRepo, socialRepo, nil)
	defer mr.Close()

	err := svc.UnlinkSocialAccount(context.Background(), user.ID, domain.OAuthProviderVK)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for last auth method, got: %v", err)
	}
}

func TestOAuthService_ListSocialAccounts(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
	}
	_ = userRepo.Create(context.Background(), user)

	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderVK,
		ProviderID: "vk-id",
		LinkedAt:   time.Now(),
	})
	_ = socialRepo.Create(context.Background(), &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   domain.OAuthProviderGoogle,
		ProviderID: "google-id",
		LinkedAt:   time.Now(),
	})

	svc, mr := newTestOAuthService(t, userRepo, socialRepo, nil)
	defer mr.Close()

	accounts, err := svc.ListSocialAccounts(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 social accounts, got %d", len(accounts))
	}
}

func TestOAuthService_OAuthCallback_EmptyName_UsesFallback(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{
			userInfo: &auth.OAuthUserInfo{
				ProviderID: "vk-no-name",
				Email:      "noname@example.com",
				Name:       "",
			},
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	user, token, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderVK, "test-code", "test-state-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if user.Name != "User" {
		t.Errorf("name = %q, want fallback %q", user.Name, "User")
	}
}

func TestOAuthService_OAuthCallback_ExchangeError(t *testing.T) {
	userRepo := mock.NewUserRepo()
	socialRepo := mock.NewSocialAccountRepo()
	providers := map[domain.OAuthProvider]auth.OAuthProvider{
		domain.OAuthProviderVK: &mockOAuthProvider{
			err: errors.New("exchange failed"),
		},
	}
	svc, mr := newTestOAuthService(t, userRepo, socialRepo, providers)
	defer mr.Close()

	_, _, err := svc.OAuthCallback(context.Background(), domain.OAuthProviderVK, "bad-code", "test-state-12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
