package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type PhotoStatus string

const (
	PhotoStatusPending  PhotoStatus = "pending"
	PhotoStatusVerified PhotoStatus = "verified"
	PhotoStatusRejected PhotoStatus = "rejected"
)

func (s PhotoStatus) IsValid() bool {
	switch s {
	case PhotoStatusPending, PhotoStatusVerified, PhotoStatusRejected:
		return true
	}
	return false
}

type BathhousePhoto struct {
	ID              uuid.UUID
	BathhouseID     uuid.UUID
	URL             string
	ThumbnailURL    string
	Position        int
	Status          PhotoStatus
	VerifiedByID    *uuid.UUID
	VerifiedAt      *time.Time
	RejectionReason string
	UploadedAt      time.Time
}

func (p *BathhousePhoto) Validate() error {
	if p.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if p.URL == "" || (!strings.HasPrefix(p.URL, "https://") && !strings.HasPrefix(p.URL, "http://")) {
		return ErrInvalidInput
	}
	if p.ThumbnailURL != "" && !strings.HasPrefix(p.ThumbnailURL, "https://") && !strings.HasPrefix(p.ThumbnailURL, "http://") {
		return ErrInvalidInput
	}
	if p.Position < 0 {
		return ErrInvalidInput
	}
	if p.Status != "" && !p.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
