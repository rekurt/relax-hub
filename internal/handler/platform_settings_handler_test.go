package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPlatformSettingsService struct {
	getStringFn func(key string) (string, error)
}

func (m *mockPlatformSettingsService) GetString(_ context.Context, key string) (string, error) {
	return m.getStringFn(key)
}

func (m *mockPlatformSettingsService) GetInt(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (m *mockPlatformSettingsService) GetFloat(_ context.Context, _ string) (float64, error) {
	return 0, nil
}

func (m *mockPlatformSettingsService) GetBool(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockPlatformSettingsService) Set(_ context.Context, _, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockPlatformSettingsService) GetAll(_ context.Context) ([]domain.PlatformSetting, error) {
	return nil, nil
}

func TestPlatformSettingsHandler_GetPublic(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		mockValue  string
		mockErr    error
		wantStatus int
		wantValue  string
	}{
		{
			name:       "returns whitelisted setting",
			key:        "listing_wizard_video_url",
			mockValue:  "https://www.youtube.com/watch?v=test123",
			wantStatus: http.StatusOK,
			wantValue:  "https://www.youtube.com/watch?v=test123",
		},
		{
			name:       "returns empty whitelisted setting",
			key:        "listing_wizard_video_url",
			mockValue:  "",
			wantStatus: http.StatusOK,
			wantValue:  "",
		},
		{
			name:       "rejects non-whitelisted key",
			key:        "jwt_secret",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "rejects unknown key",
			key:        "unknown_setting",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPlatformSettingsService{
				getStringFn: func(key string) (string, error) {
					if tt.mockErr != nil {
						return "", tt.mockErr
					}
					return tt.mockValue, nil
				},
			}

			h := handler.NewPlatformSettingsHandler(mock)

			r := chi.NewRouter()
			r.Get("/settings/{key}", h.GetPublic)

			req := httptest.NewRequest(http.MethodGet, "/settings/"+tt.key, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Success bool `json:"success"`
					Data    struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"data"`
				}
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.key, resp.Data.Key)
				assert.Equal(t, tt.wantValue, resp.Data.Value)
			}
		})
	}
}
