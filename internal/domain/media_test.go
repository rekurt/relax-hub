package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestMediaType_IsValid(t *testing.T) {
	tests := []struct {
		mediaType MediaType
		valid     bool
	}{
		{MediaTypeImage, true},
		{MediaTypeVideo, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.mediaType.IsValid(); got != tt.valid {
			t.Errorf("MediaType(%q).IsValid() = %v, want %v", tt.mediaType, got, tt.valid)
		}
	}
}

func TestMediaStatus_IsValid(t *testing.T) {
	tests := []struct {
		status MediaStatus
		valid  bool
	}{
		{MediaStatusPending, true},
		{MediaStatusApproved, true},
		{MediaStatusRejected, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("MediaStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestMediaOwnerType_IsValid(t *testing.T) {
	tests := []struct {
		ownerType MediaOwnerType
		valid     bool
	}{
		{MediaOwnerReview, true},
		{MediaOwnerBathhouse, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.ownerType.IsValid(); got != tt.valid {
			t.Errorf("MediaOwnerType(%q).IsValid() = %v, want %v", tt.ownerType, got, tt.valid)
		}
	}
}

func TestMedia_Validate(t *testing.T) {
	valid := &Media{
		OwnerType:    MediaOwnerReview,
		OwnerID:      uuid.New(),
		UserID:       uuid.New(),
		Type:         MediaTypeImage,
		URL:          "https://s3.example.com/photo.jpg",
		OriginalName: "photo.jpg",
		Size:         1024,
		MimeType:     "image/jpeg",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid media returned error: %v", err)
	}

	validWithStatus := &Media{
		OwnerType:    MediaOwnerBathhouse,
		OwnerID:      uuid.New(),
		UserID:       uuid.New(),
		Type:         MediaTypeVideo,
		URL:          "https://s3.example.com/video.mp4",
		OriginalName: "video.mp4",
		Size:         5000,
		MimeType:     "video/mp4",
		Status:       MediaStatusPending,
	}
	if err := validWithStatus.Validate(); err != nil {
		t.Errorf("valid media with status returned error: %v", err)
	}

	tests := []struct {
		name  string
		media Media
	}{
		{"invalid owner type", Media{OwnerType: "bad", OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg"}},
		{"nil owner id", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.Nil, UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg"}},
		{"nil user id", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.Nil, Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg"}},
		{"invalid type", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: "bad", URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg"}},
		{"empty url", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg"}},
		{"empty original name", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "", Size: 1, MimeType: "image/jpeg"}},
		{"zero size", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 0, MimeType: "image/jpeg"}},
		{"empty mime type", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: ""}},
		{"invalid status", Media{OwnerType: MediaOwnerReview, OwnerID: uuid.New(), UserID: uuid.New(), Type: MediaTypeImage, URL: "url", OriginalName: "f.jpg", Size: 1, MimeType: "image/jpeg", Status: "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.media.Validate(); err == nil {
				t.Error("expected error for invalid media")
			}
		})
	}
}

func TestMedia_ValidateFileSize(t *testing.T) {
	tests := []struct {
		name    string
		media   Media
		wantErr bool
	}{
		{"image under limit", Media{Type: MediaTypeImage, Size: MaxImageSizeBytes - 1}, false},
		{"image at limit", Media{Type: MediaTypeImage, Size: MaxImageSizeBytes}, false},
		{"image over limit", Media{Type: MediaTypeImage, Size: MaxImageSizeBytes + 1}, true},
		{"video under limit", Media{Type: MediaTypeVideo, Size: MaxVideoSizeBytes - 1}, false},
		{"video at limit", Media{Type: MediaTypeVideo, Size: MaxVideoSizeBytes}, false},
		{"video over limit", Media{Type: MediaTypeVideo, Size: MaxVideoSizeBytes + 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.media.ValidateFileSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMedia_ValidateMimeType(t *testing.T) {
	tests := []struct {
		name    string
		media   Media
		wantErr bool
	}{
		{"jpeg", Media{Type: MediaTypeImage, MimeType: "image/jpeg"}, false},
		{"png", Media{Type: MediaTypeImage, MimeType: "image/png"}, false},
		{"webp", Media{Type: MediaTypeImage, MimeType: "image/webp"}, true},
		{"invalid image mime", Media{Type: MediaTypeImage, MimeType: "image/gif"}, true},
		{"mp4", Media{Type: MediaTypeVideo, MimeType: "video/mp4"}, false},
		{"webm", Media{Type: MediaTypeVideo, MimeType: "video/webm"}, false},
		{"invalid video mime", Media{Type: MediaTypeVideo, MimeType: "video/avi"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.media.ValidateMimeType()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMimeType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
