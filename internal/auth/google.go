package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config      *oauth2.Config
	client      *http.Client
	userinfoURL string
}

type GoogleProviderOption func(*GoogleProvider)

func WithGoogleHTTPClient(client *http.Client) GoogleProviderOption {
	return func(p *GoogleProvider) {
		p.client = client
	}
}

func WithGoogleUserinfoURL(url string) GoogleProviderOption {
	return func(p *GoogleProvider) {
		p.userinfoURL = url
	}
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string, opts ...GoogleProviderOption) *GoogleProvider {
	p := &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		client:      http.DefaultClient,
		userinfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*OAuthUserInfo, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google token exchange: %w", err)
	}

	return p.fetchUserInfo(ctx, token.AccessToken)
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (p *GoogleProvider) fetchUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("google: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google: fetch user info: %w", err)
	}
	defer resp.Body.Close()

	var user googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("google: decode response: %w", err)
	}

	if user.ID == "" {
		return nil, fmt.Errorf("google: empty user id")
	}

	return &OAuthUserInfo{
		ProviderID: user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.Picture,
	}, nil
}
