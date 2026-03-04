package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenServer creates a test server that simulates OAuth2 token exchange.
// It returns both the token endpoint URL and the server (for cleanup).
func mockTokenServer(t *testing.T, extraFields map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"access_token": "mock-access-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		}
		for k, v := range extraFields {
			resp[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// --- VK Provider Tests ---

func TestVKProvider_GetAuthURL(t *testing.T) {
	p := NewVKProvider("client-id", "client-secret", "http://localhost/callback")
	url := p.GetAuthURL("test-state")

	assert.Contains(t, url, "client_id=client-id")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "redirect_uri=")
}

func TestVKProvider_Exchange(t *testing.T) {
	// Mock VK API for users.get
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := vkUsersGetResponse{
			Response: []vkUser{
				{
					ID:        12345,
					FirstName: "Иван",
					LastName:  "Петров",
					Photo200:  "https://vk.com/photo.jpg",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	// Mock token server (VK returns email and user_id in token response)
	tokenServer := mockTokenServer(t, map[string]any{
		"email":   "ivan@vk.com",
		"user_id": float64(12345),
	})
	defer tokenServer.Close()

	p := NewVKProvider("client-id", "client-secret", "http://localhost/callback",
		WithVKHTTPClient(tokenServer.Client()),
		WithVKAPIBase(apiServer.URL),
	)
	// Override the endpoint to point to our mock token server
	p.config.Endpoint.TokenURL = tokenServer.URL

	info, err := p.Exchange(context.Background(), "mock-code")
	require.NoError(t, err)

	assert.Equal(t, "12345", info.ProviderID)
	assert.Equal(t, "ivan@vk.com", info.Email)
	assert.Equal(t, "Иван Петров", info.Name)
	assert.Equal(t, "https://vk.com/photo.jpg", info.AvatarURL)
}

func TestVKProvider_Exchange_NoUserID(t *testing.T) {
	tokenServer := mockTokenServer(t, map[string]any{
		"email": "ivan@vk.com",
		// no user_id
	})
	defer tokenServer.Close()

	p := NewVKProvider("client-id", "client-secret", "http://localhost/callback",
		WithVKHTTPClient(tokenServer.Client()),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	_, err := p.Exchange(context.Background(), "mock-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id not found")
}

func TestVKProvider_Exchange_EmptyResponse(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := vkUsersGetResponse{Response: []vkUser{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	tokenServer := mockTokenServer(t, map[string]any{
		"email":   "ivan@vk.com",
		"user_id": float64(12345),
	})
	defer tokenServer.Close()

	p := NewVKProvider("client-id", "client-secret", "http://localhost/callback",
		WithVKHTTPClient(tokenServer.Client()),
		WithVKAPIBase(apiServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	_, err := p.Exchange(context.Background(), "mock-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty user response")
}

// --- Yandex Provider Tests ---

func TestYandexProvider_GetAuthURL(t *testing.T) {
	p := NewYandexProvider("client-id", "client-secret", "http://localhost/callback")
	url := p.GetAuthURL("test-state")

	assert.Contains(t, url, "client_id=client-id")
	assert.Contains(t, url, "state=test-state")
}

func TestYandexProvider_Exchange(t *testing.T) {
	// Mock Yandex userinfo API
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "OAuth mock-access-token", r.Header.Get("Authorization"))
		resp := yandexUserInfo{
			ID:              "67890",
			DisplayName:     "Мария Иванова",
			DefaultEmail:    "maria@yandex.ru",
			DefaultAvatarID: "abc123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer userinfoServer.Close()

	tokenServer := mockTokenServer(t, nil)
	defer tokenServer.Close()

	p := NewYandexProvider("client-id", "client-secret", "http://localhost/callback",
		WithYandexHTTPClient(tokenServer.Client()),
		WithYandexUserinfoURL(userinfoServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	info, err := p.Exchange(context.Background(), "mock-code")
	require.NoError(t, err)

	assert.Equal(t, "67890", info.ProviderID)
	assert.Equal(t, "maria@yandex.ru", info.Email)
	assert.Equal(t, "Мария Иванова", info.Name)
	assert.Equal(t, "https://avatars.yandex.net/get-yapic/abc123/islands-200", info.AvatarURL)
}

func TestYandexProvider_Exchange_NoAvatar(t *testing.T) {
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := yandexUserInfo{
			ID:            "67890",
			RealName:      "Мария Иванова",
			DefaultEmail:  "maria@yandex.ru",
			IsAvatarEmpty: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer userinfoServer.Close()

	tokenServer := mockTokenServer(t, nil)
	defer tokenServer.Close()

	p := NewYandexProvider("client-id", "client-secret", "http://localhost/callback",
		WithYandexHTTPClient(tokenServer.Client()),
		WithYandexUserinfoURL(userinfoServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	info, err := p.Exchange(context.Background(), "mock-code")
	require.NoError(t, err)

	assert.Equal(t, "67890", info.ProviderID)
	assert.Equal(t, "Мария Иванова", info.Name)
	assert.Empty(t, info.AvatarURL)
}

func TestYandexProvider_Exchange_EmptyID(t *testing.T) {
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := yandexUserInfo{DisplayName: "Test"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer userinfoServer.Close()

	tokenServer := mockTokenServer(t, nil)
	defer tokenServer.Close()

	p := NewYandexProvider("client-id", "client-secret", "http://localhost/callback",
		WithYandexHTTPClient(tokenServer.Client()),
		WithYandexUserinfoURL(userinfoServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	_, err := p.Exchange(context.Background(), "mock-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty user id")
}

// --- Google Provider Tests ---

func TestGoogleProvider_GetAuthURL(t *testing.T) {
	p := NewGoogleProvider("client-id", "client-secret", "http://localhost/callback")
	url := p.GetAuthURL("test-state")

	assert.Contains(t, url, "client_id=client-id")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "scope=")
}

func TestGoogleProvider_Exchange(t *testing.T) {
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mock-access-token", r.Header.Get("Authorization"))
		resp := googleUserInfo{
			ID:      "google-id-123",
			Email:   "user@gmail.com",
			Name:    "John Doe",
			Picture: "https://google.com/photo.jpg",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer userinfoServer.Close()

	tokenServer := mockTokenServer(t, nil)
	defer tokenServer.Close()

	p := NewGoogleProvider("client-id", "client-secret", "http://localhost/callback",
		WithGoogleHTTPClient(tokenServer.Client()),
		WithGoogleUserinfoURL(userinfoServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	info, err := p.Exchange(context.Background(), "mock-code")
	require.NoError(t, err)

	assert.Equal(t, "google-id-123", info.ProviderID)
	assert.Equal(t, "user@gmail.com", info.Email)
	assert.Equal(t, "John Doe", info.Name)
	assert.Equal(t, "https://google.com/photo.jpg", info.AvatarURL)
}

func TestGoogleProvider_Exchange_EmptyID(t *testing.T) {
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := googleUserInfo{Email: "user@gmail.com", Name: "John"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer userinfoServer.Close()

	tokenServer := mockTokenServer(t, nil)
	defer tokenServer.Close()

	p := NewGoogleProvider("client-id", "client-secret", "http://localhost/callback",
		WithGoogleHTTPClient(tokenServer.Client()),
		WithGoogleUserinfoURL(userinfoServer.URL),
	)
	p.config.Endpoint.TokenURL = tokenServer.URL

	_, err := p.Exchange(context.Background(), "mock-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty user id")
}

// --- Interface compliance ---

func TestProviders_ImplementInterface(t *testing.T) {
	var _ OAuthProvider = (*VKProvider)(nil)
	var _ OAuthProvider = (*YandexProvider)(nil)
	var _ OAuthProvider = (*GoogleProvider)(nil)
}
