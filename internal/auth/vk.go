package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const vkAPIVersion = "5.131"

type VKProvider struct {
	config  *oauth2.Config
	client  *http.Client
	apiBase string
}

type VKProviderOption func(*VKProvider)

func WithVKHTTPClient(client *http.Client) VKProviderOption {
	return func(p *VKProvider) {
		p.client = client
	}
}

func WithVKAPIBase(base string) VKProviderOption {
	return func(p *VKProvider) {
		p.apiBase = base
	}
}

func NewVKProvider(clientID, clientSecret, redirectURL string, opts ...VKProviderOption) *VKProvider {
	p := &VKProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"email"},
			Endpoint:     vk.Endpoint,
		},
		client:  http.DefaultClient,
		apiBase: "https://api.vk.com",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *VKProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *VKProvider) Exchange(ctx context.Context, code string) (*OAuthUserInfo, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("vk token exchange: %w", err)
	}

	// VK returns email in the token response extra fields
	var email string
	if emailVal := token.Extra("email"); emailVal != nil {
		email, _ = emailVal.(string)
	}

	userIDRaw := token.Extra("user_id")
	if userIDRaw == nil {
		return nil, fmt.Errorf("vk: user_id not found in token response")
	}

	userID, ok := userIDRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("vk: user_id is not a number: %T", userIDRaw)
	}

	if userID == 0 {
		return nil, fmt.Errorf("vk: user_id is zero")
	}

	info, err := p.fetchUserInfo(ctx, token.AccessToken, int64(userID))
	if err != nil {
		return nil, err
	}

	info.Email = email
	return info, nil
}

type vkUsersGetResponse struct {
	Response []vkUser `json:"response"`
}

type vkUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Photo200  string `json:"photo_200"`
}

func (p *VKProvider) fetchUserInfo(ctx context.Context, accessToken string, userID int64) (*OAuthUserInfo, error) {
	url := fmt.Sprintf(
		"%s/method/users.get?user_ids=%d&fields=photo_200&v=%s",
		p.apiBase, userID, vkAPIVersion,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("vk: create request: %w", err)
	}

	// Include access token in Authorization header instead of URL query
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vk: fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vk: API returned status %d", resp.StatusCode)
	}

	var result vkUsersGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("vk: decode response: %w", err)
	}

	if len(result.Response) == 0 {
		return nil, fmt.Errorf("vk: empty user response")
	}

	user := result.Response[0]
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}

	return &OAuthUserInfo{
		ProviderID: strconv.FormatInt(user.ID, 10),
		Name:       name,
		AvatarURL:  user.Photo200,
	}, nil
}
