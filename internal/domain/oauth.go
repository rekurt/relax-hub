package domain

import (
	"time"

	"github.com/google/uuid"
)

type OAuthProvider string

const (
	OAuthProviderVK     OAuthProvider = "vk"
	OAuthProviderYandex OAuthProvider = "yandex"
	OAuthProviderGoogle OAuthProvider = "google"
)

func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderVK, OAuthProviderYandex, OAuthProviderGoogle:
		return true
	}
	return false
}

type SocialAccount struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Provider    OAuthProvider
	ProviderID  string
	Email       string
	Name        string
	AvatarURL   string
	AccessToken string `json:"-"`
	LinkedAt    time.Time
}

func (s *SocialAccount) Validate() error {
	if s.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !s.Provider.IsValid() {
		return ErrInvalidInput
	}
	if s.ProviderID == "" {
		return ErrInvalidInput
	}
	return nil
}
