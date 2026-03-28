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

// UserRegion represents the geographic region the user belongs to.
type UserRegion string

const (
	RegionRU UserRegion = "RU"
	RegionBY UserRegion = "BY"
)

func (r UserRegion) IsValid() bool {
	switch r {
	case RegionRU, RegionBY:
		return true
	}
	return false
}

// CurrencyForRegion returns the wallet currency for a given region.
func CurrencyForRegion(region UserRegion) WalletCurrency {
	switch region {
	case RegionBY:
		return WalletCurrencyBYN
	default:
		return WalletCurrencyRUB
	}
}

type User struct {
	ID                  uuid.UUID
	Email               string
	PasswordHash        string `json:"-"`
	Name                string
	Phone               string
	PhoneVerified       bool
	Role                UserRole
	AdminSubRole        AdminSubRole
	IsActive            bool
	AvatarURL           string
	Bio                 string
	CityID              *int64
	Region              UserRegion
	ReferralCode        string
	TOTPSecret          string `json:"-"`
	TwoFAMethod         TwoFAMethod
	AgeConfirmed        bool
	OnboardingCompleted bool
	DeletionRequestedAt *time.Time
	DeletionScheduledAt *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ProfileCompletenessItem represents a single profile field and whether it's filled.
type ProfileCompletenessItem struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	Complete bool   `json:"complete"`
}

// ProfileCompleteness returns the completeness percentage (0-100) and individual field statuses.
func (u *User) ProfileCompleteness(hasPreferences bool) (int, []ProfileCompletenessItem) {
	items := []ProfileCompletenessItem{
		{Field: "name", Label: "Имя", Complete: u.Name != ""},
		{Field: "avatar", Label: "Фото профиля", Complete: u.AvatarURL != ""},
		{Field: "phone", Label: "Телефон", Complete: u.Phone != ""},
		{Field: "bio", Label: "О себе", Complete: u.Bio != ""},
		{Field: "city", Label: "Город", Complete: u.CityID != nil},
		{Field: "preferences", Label: "Предпочтения", Complete: hasPreferences},
	}
	filled := 0
	for _, item := range items {
		if item.Complete {
			filled++
		}
	}
	pct := filled * 100 / len(items)
	return pct, items
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
	TotalSpent  int64 // in kopecks
	AvgCheck    int64 // in kopecks
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
