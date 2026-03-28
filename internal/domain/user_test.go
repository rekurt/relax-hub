package domain_test

import (
	"testing"

	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestUser_ProfileCompleteness_AllEmpty(t *testing.T) {
	u := &domain.User{}
	pct, items := u.ProfileCompleteness(false)
	if pct != 0 {
		t.Errorf("expected 0%%, got %d%%", pct)
	}
	if len(items) != 6 {
		t.Errorf("expected 6 items, got %d", len(items))
	}
	for _, item := range items {
		if item.Complete {
			t.Errorf("field %q should not be complete", item.Field)
		}
	}
}

func TestUser_ProfileCompleteness_AllFilled(t *testing.T) {
	cityID := int64(1)
	u := &domain.User{
		Name:      "Test",
		AvatarURL: "http://example.com/avatar.jpg",
		Phone:     "+79001234567",
		Bio:       "Bio text",
		CityID:    &cityID,
	}
	pct, items := u.ProfileCompleteness(true)
	if pct != 100 {
		t.Errorf("expected 100%%, got %d%%", pct)
	}
	for _, item := range items {
		if !item.Complete {
			t.Errorf("field %q should be complete", item.Field)
		}
	}
}

func TestUser_ProfileCompleteness_Partial(t *testing.T) {
	u := &domain.User{
		Name:  "Test",
		Phone: "+79001234567",
	}
	pct, _ := u.ProfileCompleteness(false)
	// 2 of 6 = 33%
	if pct != 33 {
		t.Errorf("expected 33%%, got %d%%", pct)
	}
}

func TestUser_ProfileCompleteness_WithPreferences(t *testing.T) {
	u := &domain.User{
		Name:  "Test",
		Phone: "+79001234567",
		Bio:   "some bio",
	}
	pct, _ := u.ProfileCompleteness(true)
	// 4 of 6 = 66%
	if pct != 66 {
		t.Errorf("expected 66%%, got %d%%", pct)
	}
}

func TestUser_Validate(t *testing.T) {
	// Valid user
	u := &domain.User{Email: "test@example.com", Name: "Test", Role: domain.RoleClient}
	if err := u.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// No email or phone
	u2 := &domain.User{Name: "Test", Role: domain.RoleClient}
	if err := u2.Validate(); err == nil {
		t.Error("expected error for missing email and phone")
	}

	// No name
	u3 := &domain.User{Email: "test@example.com", Role: domain.RoleClient}
	if err := u3.Validate(); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestUserRegion_IsValid(t *testing.T) {
	if !domain.RegionRU.IsValid() {
		t.Error("RU should be valid")
	}
	if !domain.RegionBY.IsValid() {
		t.Error("BY should be valid")
	}
	if domain.UserRegion("XX").IsValid() {
		t.Error("XX should not be valid")
	}
}
