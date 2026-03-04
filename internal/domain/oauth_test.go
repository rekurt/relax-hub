package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestOAuthProvider_IsValid(t *testing.T) {
	tests := []struct {
		provider OAuthProvider
		valid    bool
	}{
		{OAuthProviderVK, true},
		{OAuthProviderYandex, true},
		{OAuthProviderGoogle, true},
		{"facebook", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.provider.IsValid(); got != tt.valid {
			t.Errorf("OAuthProvider(%q).IsValid() = %v, want %v", tt.provider, got, tt.valid)
		}
	}
}

func TestSocialAccount_Validate(t *testing.T) {
	valid := &SocialAccount{
		UserID:     uuid.New(),
		Provider:   OAuthProviderVK,
		ProviderID: "123456",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid social account returned error: %v", err)
	}

	tests := []struct {
		name    string
		account SocialAccount
	}{
		{"nil user", SocialAccount{UserID: uuid.Nil, Provider: OAuthProviderVK, ProviderID: "123"}},
		{"invalid provider", SocialAccount{UserID: uuid.New(), Provider: "facebook", ProviderID: "123"}},
		{"empty provider", SocialAccount{UserID: uuid.New(), Provider: "", ProviderID: "123"}},
		{"empty provider_id", SocialAccount{UserID: uuid.New(), Provider: OAuthProviderGoogle, ProviderID: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.account.Validate(); err == nil {
				t.Error("expected error for invalid social account")
			}
		})
	}
}

func TestSocialAccount_Fields(t *testing.T) {
	account := SocialAccount{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Provider:    OAuthProviderYandex,
		ProviderID:  "yandex-user-42",
		Email:       "user@yandex.ru",
		Name:        "Yandex User",
		AvatarURL:   "https://avatars.yandex.net/photo.jpg",
		AccessToken: "secret-token",
	}

	if account.Provider != OAuthProviderYandex {
		t.Errorf("Provider = %q, want %q", account.Provider, OAuthProviderYandex)
	}
	if account.ProviderID != "yandex-user-42" {
		t.Errorf("ProviderID = %q, want %q", account.ProviderID, "yandex-user-42")
	}
	if account.Email != "user@yandex.ru" {
		t.Errorf("Email = %q, want %q", account.Email, "user@yandex.ru")
	}
	if account.Name != "Yandex User" {
		t.Errorf("Name = %q, want %q", account.Name, "Yandex User")
	}
	if account.AvatarURL != "https://avatars.yandex.net/photo.jpg" {
		t.Errorf("AvatarURL = %q, want %q", account.AvatarURL, "https://avatars.yandex.net/photo.jpg")
	}
	if account.AccessToken != "secret-token" {
		t.Errorf("AccessToken = %q, want %q", account.AccessToken, "secret-token")
	}
}
