package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestPhotoStatus_IsValid(t *testing.T) {
	tests := []struct {
		status PhotoStatus
		valid  bool
	}{
		{PhotoStatusPending, true},
		{PhotoStatusVerified, true},
		{PhotoStatusRejected, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("PhotoStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestBathhousePhoto_Validate(t *testing.T) {
	valid := &BathhousePhoto{
		BathhouseID: uuid.New(),
		URL:         "https://example.com/photo.jpg",
		Position:    0,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid photo returned error: %v", err)
	}

	validWithStatus := &BathhousePhoto{
		BathhouseID: uuid.New(),
		URL:         "https://example.com/photo.jpg",
		Position:    1,
		Status:      PhotoStatusPending,
	}
	if err := validWithStatus.Validate(); err != nil {
		t.Errorf("valid photo with status returned error: %v", err)
	}

	tests := []struct {
		name  string
		photo BathhousePhoto
	}{
		{"nil bathhouse ID", BathhousePhoto{BathhouseID: uuid.Nil, URL: "https://example.com/photo.jpg"}},
		{"empty URL", BathhousePhoto{BathhouseID: uuid.New(), URL: ""}},
		{"invalid URL scheme", BathhousePhoto{BathhouseID: uuid.New(), URL: "javascript:alert(1)"}},
		{"no URL scheme", BathhousePhoto{BathhouseID: uuid.New(), URL: "example.com/photo.jpg"}},
		{"invalid thumbnail URL scheme", BathhousePhoto{BathhouseID: uuid.New(), URL: "https://example.com/photo.jpg", ThumbnailURL: "ftp://example.com/thumb.jpg"}},
		{"negative position", BathhousePhoto{BathhouseID: uuid.New(), URL: "https://example.com/photo.jpg", Position: -1}},
		{"invalid status", BathhousePhoto{BathhouseID: uuid.New(), URL: "https://example.com/photo.jpg", Status: "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.photo.Validate(); err == nil {
				t.Error("expected error for invalid photo")
			}
		})
	}
}
