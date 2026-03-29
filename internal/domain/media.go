package domain

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

func (t MediaType) IsValid() bool {
	switch t {
	case MediaTypeImage, MediaTypeVideo:
		return true
	}
	return false
}

type MediaStatus string

const (
	MediaStatusPending  MediaStatus = "pending"
	MediaStatusApproved MediaStatus = "approved"
	MediaStatusRejected MediaStatus = "rejected"
)

func (s MediaStatus) IsValid() bool {
	switch s {
	case MediaStatusPending, MediaStatusApproved, MediaStatusRejected:
		return true
	}
	return false
}

type MediaOwnerType string

const (
	MediaOwnerReview    MediaOwnerType = "review"
	MediaOwnerBathhouse MediaOwnerType = "bathhouse"
)

func (t MediaOwnerType) IsValid() bool {
	switch t {
	case MediaOwnerReview, MediaOwnerBathhouse:
		return true
	}
	return false
}

const (
	MaxImagesPerReview = 10
	MaxVideosPerReview = 1
	MaxImageSizeBytes  = 10 * 1024 * 1024 // 10MB
	MaxVideoSizeBytes  = 50 * 1024 * 1024 // 50MB
	MaxImageWidth      = 1920
	ThumbnailSize      = 300
)

var (
	AllowedImageMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	AllowedVideoMimeTypes = map[string]bool{
		"video/mp4":  true,
		"video/webm": true,
	}
)

type Media struct {
	ID           uuid.UUID
	OwnerType    MediaOwnerType
	OwnerID      uuid.UUID
	UserID       uuid.UUID
	Type         MediaType
	URL          string
	ThumbnailURL string
	OriginalName string
	Size         int64
	MimeType     string
	Width        int
	Height       int
	Status       MediaStatus
	CreatedAt    time.Time
}

func (m *Media) Validate() error {
	if !m.OwnerType.IsValid() {
		return ErrInvalidInput
	}
	if m.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if m.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !m.Type.IsValid() {
		return ErrInvalidInput
	}
	if m.URL == "" {
		return ErrInvalidInput
	}
	if m.OriginalName == "" {
		return ErrInvalidInput
	}
	if m.Size <= 0 {
		return ErrInvalidInput
	}
	if m.MimeType == "" {
		return ErrInvalidInput
	}
	if m.Status != "" && !m.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

func (m *Media) ValidateFileSize() error {
	switch m.Type {
	case MediaTypeImage:
		if m.Size > MaxImageSizeBytes {
			return ErrMediaFileTooLarge
		}
	case MediaTypeVideo:
		if m.Size > MaxVideoSizeBytes {
			return ErrMediaFileTooLarge
		}
	default:
		return ErrMediaInvalidType
	}
	return nil
}

func (m *Media) ValidateMimeType() error {
	switch m.Type {
	case MediaTypeImage:
		if !AllowedImageMimeTypes[m.MimeType] {
			return ErrMediaInvalidType
		}
	case MediaTypeVideo:
		if !AllowedVideoMimeTypes[m.MimeType] {
			return ErrMediaInvalidType
		}
	default:
		return ErrMediaInvalidType
	}
	return nil
}
