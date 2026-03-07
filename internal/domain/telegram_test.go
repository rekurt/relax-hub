package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestTelegramLinkValidate(t *testing.T) {
	tests := []struct {
		name    string
		link    *TelegramLink
		wantErr bool
	}{
		{
			name: "valid telegram link",
			link: &TelegramLink{
				ID:               uuid.New(),
				UserID:           uuid.New(),
				TelegramID:       123456789,
				TelegramUsername: "testuser",
			},
			wantErr: false,
		},
		{
			name: "missing user id",
			link: &TelegramLink{
				ID:               uuid.New(),
				TelegramID:       123456789,
				TelegramUsername: "testuser",
			},
			wantErr: true,
		},
		{
			name: "invalid telegram id",
			link: &TelegramLink{
				ID:               uuid.New(),
				UserID:           uuid.New(),
				TelegramID:       0,
				TelegramUsername: "testuser",
			},
			wantErr: true,
		},
		{
			name: "missing telegram username",
			link: &TelegramLink{
				ID:         uuid.New(),
				UserID:     uuid.New(),
				TelegramID: 123456789,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.link.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
