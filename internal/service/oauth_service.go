package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/auth"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type OAuthService interface {
	GetOAuthURL(provider domain.OAuthProvider) (string, error)
	OAuthCallback(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, string, error)
	LinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider, code string) error
	UnlinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
	ListSocialAccounts(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error)
}

type oauthService struct {
	userRepo       repository.UserRepository
	socialRepo     repository.SocialAccountRepository
	redisClient    *redis.Client
	providers      map[domain.OAuthProvider]auth.OAuthProvider
	jwtSecret      []byte
	tokenTTL       time.Duration
}

func NewOAuthService(
	userRepo repository.UserRepository,
	socialRepo repository.SocialAccountRepository,
	redisClient *redis.Client,
	cfg *config.Config,
) OAuthService {
	providers := make(map[domain.OAuthProvider]auth.OAuthProvider)

	if cfg.OAuth.VK.ClientID != "" {
		providers[domain.OAuthProviderVK] = auth.NewVKProvider(
			cfg.OAuth.VK.ClientID, cfg.OAuth.VK.ClientSecret, cfg.OAuth.VK.RedirectURL,
		)
	}
	if cfg.OAuth.Yandex.ClientID != "" {
		providers[domain.OAuthProviderYandex] = auth.NewYandexProvider(
			cfg.OAuth.Yandex.ClientID, cfg.OAuth.Yandex.ClientSecret, cfg.OAuth.Yandex.RedirectURL,
		)
	}
	if cfg.OAuth.Google.ClientID != "" {
		providers[domain.OAuthProviderGoogle] = auth.NewGoogleProvider(
			cfg.OAuth.Google.ClientID, cfg.OAuth.Google.ClientSecret, cfg.OAuth.Google.RedirectURL,
		)
	}

	return &oauthService{
		userRepo:     userRepo,
		socialRepo:   socialRepo,
		redisClient:  redisClient,
		providers:    providers,
		jwtSecret:    []byte(cfg.JWT.Secret),
		tokenTTL:     cfg.JWT.TokenTTL,
	}
}

// NewOAuthServiceWithProviders creates an OAuthService with explicit provider map (for testing).
func NewOAuthServiceWithProviders(
	userRepo repository.UserRepository,
	socialRepo repository.SocialAccountRepository,
	redisClient *redis.Client,
	providers map[domain.OAuthProvider]auth.OAuthProvider,
	jwtSecret string,
	tokenTTL time.Duration,
) OAuthService {
	return &oauthService{
		userRepo:    userRepo,
		socialRepo:  socialRepo,
		redisClient: redisClient,
		providers:   providers,
		jwtSecret:   []byte(jwtSecret),
		tokenTTL:    tokenTTL,
	}
}

func (s *oauthService) GetOAuthURL(provider domain.OAuthProvider) (string, error) {
	p, err := s.getProvider(provider)
	if err != nil {
		return "", err
	}

	state := uuid.New().String()
	// Store state in Redis with 10-minute expiration for CSRF protection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.redisClient.Set(ctx, "oauth_state:"+state, string(provider), 10*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store OAuth state: %w", err)
	}
	return p.GetAuthURL(state), nil
}

func (s *oauthService) OAuthCallback(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, string, error) {
	// Validate CSRF state token
	if state == "" {
		return nil, "", domain.ErrInvalidInput
	}

	// Check if state exists in Redis and retrieve the provider it was issued for
	redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	storedProvider, err := s.redisClient.Get(redisCtx, "oauth_state:"+state).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, "", fmt.Errorf("%w: invalid or expired OAuth state", domain.ErrUnauthorized)
		}
		return nil, "", fmt.Errorf("failed to validate OAuth state: %w", err)
	}

	// Delete the state from Redis (one-time use only)
	// Ignore delete errors - state is already validated so this is just cleanup
	_ = s.redisClient.Del(redisCtx, "oauth_state:"+state).Err()

	// Verify the state matches the requested provider
	if domain.OAuthProvider(storedProvider) != provider {
		return nil, "", fmt.Errorf("%w: provider mismatch in OAuth state", domain.ErrUnauthorized)
	}

	p, err := s.getProvider(provider)
	if err != nil {
		return nil, "", err
	}

	info, err := p.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}

	// Try to find existing social account
	existing, err := s.socialRepo.GetByProviderAndID(ctx, provider, info.ProviderID)
	if err == nil {
		// Returning user — find and return JWT
		user, err := s.userRepo.GetByID(ctx, existing.UserID)
		if err != nil {
			return nil, "", fmt.Errorf("get user: %w", err)
		}

		if !user.IsActive {
			return nil, "", domain.ErrUserBlocked
		}

		token, err := s.generateToken(user.ID, user.Role)
		if err != nil {
			return nil, "", err
		}

		return user, token, nil
	}

	if !errors.Is(err, domain.ErrSocialAccountNotFound) {
		return nil, "", fmt.Errorf("lookup social account: %w", err)
	}

	// First-time login — check if user with same email already exists
	var user *domain.User
	if info.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, info.Email)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, "", fmt.Errorf("lookup user by email: %w", err)
		}
		if user != nil && !user.IsActive {
			return nil, "", domain.ErrUserBlocked
		}
	}

	if user == nil {
		// Create new user without password
		now := time.Now()
		email := info.Email
		if email == "" {
			email = fmt.Sprintf("oauth_%s@noemail.local", uuid.New().String())
		}
		name := info.Name
		if name == "" {
			name = "User"
		}
		user = &domain.User{
			ID:        uuid.New(),
			Email:     email,
			Name:      name,
			AvatarURL: info.AvatarURL,
			Role:      domain.RoleClient,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := user.Validate(); err != nil {
			return nil, "", fmt.Errorf("validate user: %w", err)
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("create user: %w", err)
		}
	}

	// Link social account
	socialAccount := &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     user.ID,
		Provider:   provider,
		ProviderID: info.ProviderID,
		Email:      info.Email,
		Name:       info.Name,
		AvatarURL:  info.AvatarURL,
		LinkedAt:   time.Now(),
	}

	if err := socialAccount.Validate(); err != nil {
		return nil, "", fmt.Errorf("validate social account: %w", err)
	}

	if err := s.socialRepo.Create(ctx, socialAccount); err != nil {
		return nil, "", fmt.Errorf("link social account: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *oauthService) LinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider, code string) error {
	p, err := s.getProvider(provider)
	if err != nil {
		return err
	}

	info, err := p.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}

	// Check if this social account is already linked
	_, err = s.socialRepo.GetByProviderAndID(ctx, provider, info.ProviderID)
	if err == nil {
		return domain.ErrSocialAccountAlreadyLinked
	}
	if err != nil && !errors.Is(err, domain.ErrSocialAccountNotFound) {
		return fmt.Errorf("lookup social account: %w", err)
	}

	socialAccount := &domain.SocialAccount{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   provider,
		ProviderID: info.ProviderID,
		Email:      info.Email,
		Name:       info.Name,
		AvatarURL:  info.AvatarURL,
		LinkedAt:   time.Now(),
	}

	if err := socialAccount.Validate(); err != nil {
		return fmt.Errorf("validate social account: %w", err)
	}

	if err := s.socialRepo.Create(ctx, socialAccount); err != nil {
		return fmt.Errorf("link social account: %w", err)
	}

	return nil
}

func (s *oauthService) UnlinkSocialAccount(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	if !provider.IsValid() {
		return fmt.Errorf("%w: invalid provider", domain.ErrInvalidInput)
	}

	// Check that user has a password or another linked social account
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	accounts, err := s.socialRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list social accounts: %w", err)
	}

	hasPassword := user.PasswordHash != ""
	otherProviders := 0
	for _, a := range accounts {
		if a.Provider != provider {
			otherProviders++
		}
	}

	if !hasPassword && otherProviders == 0 {
		return fmt.Errorf("%w: cannot unlink last auth method", domain.ErrInvalidInput)
	}

	return s.socialRepo.Delete(ctx, userID, provider)
}

func (s *oauthService) ListSocialAccounts(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error) {
	return s.socialRepo.ListByUser(ctx, userID)
}

func (s *oauthService) getProvider(provider domain.OAuthProvider) (auth.OAuthProvider, error) {
	if !provider.IsValid() {
		return nil, fmt.Errorf("%w: invalid provider %q", domain.ErrInvalidInput, provider)
	}

	p, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("%w: provider %q not configured", domain.ErrInvalidInput, provider)
	}

	return p, nil
}

func (s *oauthService) generateToken(userID uuid.UUID, role domain.UserRole) (string, error) {
	return generateJWT(userID, role, s.jwtSecret, s.tokenTTL)
}
