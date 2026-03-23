package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleClient         UserRole = "client"
	RoleOwner          UserRole = "owner"
	RoleRepresentative UserRole = "representative"
	RoleAdmin          UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleClient, RoleOwner, RoleRepresentative, RoleAdmin:
		return true
	}
	return false
}

type TwoFAMethod string

const (
	TwoFANone TwoFAMethod = "none"
	TwoFATOTP TwoFAMethod = "totp"
	TwoFASMS  TwoFAMethod = "sms"
)

type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string `json:"-"`
	Name          string
	Phone         string
	PhoneVerified bool
	Role          UserRole
	IsActive      bool
	AvatarURL     string
	Bio           string
	CityID        *int64
	ReferralCode  string
	TOTPSecret    string      `json:"-"`
	TwoFAMethod   TwoFAMethod
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UserProfile is a read-only aggregate for public user profiles.
type UserProfile struct {
	ID          uuid.UUID
	Name        string
	AvatarURL   string
	Bio         string
	CityName    string
	MemberSince time.Time
	ReviewCount int
	VisitCount  int
	AvgRating   float64
}

// UserBookingStats holds aggregate booking statistics for a user.
type UserBookingStats struct {
	TotalVisits int
	TotalSpent  int64   // in kopecks
	AvgCheck    int64   // in kopecks
}

// UserReviewStats holds aggregate review statistics for a user.
type UserReviewStats struct {
	ReviewCount int
	AvgRating   float64
}

func (u *User) Validate() error {
	if u.Email == "" && u.Phone == "" {
		return ErrInvalidInput
	}
	if u.Name == "" {
		return ErrInvalidInput
	}
	if !u.Role.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
