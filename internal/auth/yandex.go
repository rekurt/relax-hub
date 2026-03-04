package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type YandexProvider struct {
	config      *oauth2.Config
	client      *http.Client
	userinfoURL string
}

type YandexProviderOption func(*YandexProvider)

func WithYandexHTTPClient(client *http.Client) YandexProviderOption {
	return func(p *YandexProvider) {
		p.client = client
	}
}

func WithYandexUserinfoURL(url string) YandexProviderOption {
	return func(p *YandexProvider) {
		p.userinfoURL = url
	}
}

func NewYandexProvider(clientID, clientSecret, redirectURL string, opts ...YandexProviderOption) *YandexProvider {
	p := &YandexProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     yandex.Endpoint,
		},
		client:      http.DefaultClient,
		userinfoURL: "https://login.yandex.ru/info",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *YandexProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *YandexProvider) Exchange(ctx context.Context, code string) (*OAuthUserInfo, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("yandex token exchange: %w", err)
	}

	return p.fetchUserInfo(ctx, token.AccessToken)
}

type yandexUserInfo struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	DisplayName  string `json:"display_name"`
	RealName     string `json:"real_name"`
	DefaultEmail string `json:"default_email"`
	IsAvatarEmpty bool  `json:"is_avatar_empty"`
	DefaultAvatarID string `json:"default_avatar_id"`
}

func (p *YandexProvider) fetchUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userinfoURL+"?format=json", nil)
	if err != nil {
		return nil, fmt.Errorf("yandex: create request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yandex: fetch user info: %w", err)
	}
	defer resp.Body.Close()

	var user yandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("yandex: decode response: %w", err)
	}

	if user.ID == "" {
		return nil, fmt.Errorf("yandex: empty user id")
	}

	name := user.DisplayName
	if name == "" {
		name = user.RealName
	}

	var avatarURL string
	if !user.IsAvatarEmpty && user.DefaultAvatarID != "" {
		avatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", user.DefaultAvatarID)
	}

	return &OAuthUserInfo{
		ProviderID: user.ID,
		Email:      user.DefaultEmail,
		Name:       name,
		AvatarURL:  avatarURL,
	}, nil
}
