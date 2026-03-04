package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role  UserRole
		valid bool
	}{
		{RoleClient, true},
		{RoleOwner, true},
		{RoleRepresentative, true},
		{RoleAdmin, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.valid {
			t.Errorf("UserRole(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

func TestUser_Validate(t *testing.T) {
	valid := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  RoleClient,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid user returned error: %v", err)
	}

	cityID := int64(1)
	validWithProfile := &User{
		ID:        uuid.New(),
		Email:     "profile@example.com",
		Name:      "Profile User",
		Role:      RoleClient,
		AvatarURL: "https://example.com/avatar.jpg",
		Bio:       "Hello world",
		CityID:    &cityID,
	}
	if err := validWithProfile.Validate(); err != nil {
		t.Errorf("valid user with profile fields returned error: %v", err)
	}

	tests := []struct {
		name string
		user User
	}{
		{"empty email", User{Email: "", Name: "Test", Role: RoleClient}},
		{"empty name", User{Email: "test@example.com", Name: "", Role: RoleClient}},
		{"invalid role", User{Email: "test@example.com", Name: "Test", Role: "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.user.Validate(); err == nil {
				t.Error("expected error for invalid user")
			}
		})
	}
}

func TestUser_ProfileFields(t *testing.T) {
	cityID := int64(42)
	user := User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      RoleClient,
		AvatarURL: "https://example.com/avatar.jpg",
		Bio:       "I love saunas",
		CityID:    &cityID,
	}

	if user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("AvatarURL = %q, want %q", user.AvatarURL, "https://example.com/avatar.jpg")
	}
	if user.Bio != "I love saunas" {
		t.Errorf("Bio = %q, want %q", user.Bio, "I love saunas")
	}
	if user.CityID == nil || *user.CityID != 42 {
		t.Errorf("CityID = %v, want 42", user.CityID)
	}

	// CityID can be nil (optional)
	userNilCity := User{
		Email: "test2@example.com",
		Name:  "Test",
		Role:  RoleClient,
	}
	if userNilCity.CityID != nil {
		t.Errorf("CityID should be nil by default, got %v", userNilCity.CityID)
	}
}

func TestUserProfile(t *testing.T) {
	profile := UserProfile{
		ID:          uuid.New(),
		Name:        "Test User",
		AvatarURL:   "https://example.com/avatar.jpg",
		Bio:         "Bio text",
		CityName:    "Moscow",
		MemberSince: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ReviewCount: 5,
		VisitCount:  10,
		AvgRating:   4.2,
	}

	if profile.Name != "Test User" {
		t.Errorf("Name = %q, want %q", profile.Name, "Test User")
	}
	if profile.ReviewCount != 5 {
		t.Errorf("ReviewCount = %d, want 5", profile.ReviewCount)
	}
	if profile.VisitCount != 10 {
		t.Errorf("VisitCount = %d, want 10", profile.VisitCount)
	}
	if profile.AvgRating != 4.2 {
		t.Errorf("AvgRating = %f, want 4.2", profile.AvgRating)
	}
	if profile.CityName != "Moscow" {
		t.Errorf("CityName = %q, want %q", profile.CityName, "Moscow")
	}
}

func TestBathhouseStatus_IsValid(t *testing.T) {
	tests := []struct {
		status BathhouseStatus
		valid  bool
	}{
		{BathhouseStatusActive, true},
		{BathhouseStatusInactive, true},
		{BathhouseStatusPending, true},
		{BathhouseStatusRejected, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("BathhouseStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestBathhouse_Validate(t *testing.T) {
	valid := &Bathhouse{
		Name:         "Test Bathhouse",
		Address:      "123 Street",
		CityID:       1,
		PricePerHour: 5000,
		MaxGuests:    10,
		MinDuration:  1,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid bathhouse returned error: %v", err)
	}

	tests := []struct {
		name string
		bh   Bathhouse
	}{
		{"empty name", Bathhouse{Name: "", Address: "addr", CityID: 1, PricePerHour: 100, MaxGuests: 5, MinDuration: 1}},
		{"empty address", Bathhouse{Name: "name", Address: "", CityID: 1, PricePerHour: 100, MaxGuests: 5, MinDuration: 1}},
		{"zero city", Bathhouse{Name: "name", Address: "addr", CityID: 0, PricePerHour: 100, MaxGuests: 5, MinDuration: 1}},
		{"zero price", Bathhouse{Name: "name", Address: "addr", CityID: 1, PricePerHour: 0, MaxGuests: 5, MinDuration: 1}},
		{"zero guests", Bathhouse{Name: "name", Address: "addr", CityID: 1, PricePerHour: 100, MaxGuests: 0, MinDuration: 1}},
		{"zero duration", Bathhouse{Name: "name", Address: "addr", CityID: 1, PricePerHour: 100, MaxGuests: 5, MinDuration: 0}},
		{"invalid latitude", Bathhouse{Name: "name", Address: "addr", CityID: 1, PricePerHour: 100, MaxGuests: 5, MinDuration: 1, Latitude: 91}},
		{"invalid longitude", Bathhouse{Name: "name", Address: "addr", CityID: 1, PricePerHour: 100, MaxGuests: 5, MinDuration: 1, Longitude: 181}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bh.Validate(); err == nil {
				t.Error("expected error for invalid bathhouse")
			}
		})
	}
}

func TestBookingStatus_IsValid(t *testing.T) {
	tests := []struct {
		status BookingStatus
		valid  bool
	}{
		{BookingPending, true},
		{BookingConfirmed, true},
		{BookingCancelled, true},
		{BookingRejected, true},
		{BookingCompleted, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("BookingStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestBooking_Validate(t *testing.T) {
	now := time.Now()
	valid := &Booking{
		BathhouseID: uuid.New(),
		StartTime:   now,
		EndTime:     now.Add(2 * time.Hour),
		GuestCount:  3,
		TotalPrice:  10000,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid booking returned error: %v", err)
	}

	tests := []struct {
		name    string
		booking Booking
	}{
		{"nil bathhouse", Booking{BathhouseID: uuid.Nil, StartTime: now, EndTime: now.Add(time.Hour), GuestCount: 1, TotalPrice: 100}},
		{"zero guests", Booking{BathhouseID: uuid.New(), StartTime: now, EndTime: now.Add(time.Hour), GuestCount: 0, TotalPrice: 100}},
		{"zero price", Booking{BathhouseID: uuid.New(), StartTime: now, EndTime: now.Add(time.Hour), GuestCount: 1, TotalPrice: 0}},
		{"end before start", Booking{BathhouseID: uuid.New(), StartTime: now, EndTime: now.Add(-time.Hour), GuestCount: 1, TotalPrice: 100}},
		{"end equals start", Booking{BathhouseID: uuid.New(), StartTime: now, EndTime: now, GuestCount: 1, TotalPrice: 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.booking.Validate(); err == nil {
				t.Error("expected error for invalid booking")
			}
		})
	}
}

func TestReviewStatus_IsValid(t *testing.T) {
	tests := []struct {
		status ReviewStatus
		valid  bool
	}{
		{ReviewStatusPending, true},
		{ReviewStatusApproved, true},
		{ReviewStatusRejected, true},
		{ReviewStatusHidden, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("ReviewStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestReview_Validate(t *testing.T) {
	valid := &Review{
		BookingID: uuid.New(),
		Rating:    3,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid review returned error: %v", err)
	}

	validWithStatus := &Review{
		BookingID: uuid.New(),
		Rating:    4,
		Status:    ReviewStatusApproved,
	}
	if err := validWithStatus.Validate(); err != nil {
		t.Errorf("valid review with status returned error: %v", err)
	}

	tests := []struct {
		name   string
		review Review
	}{
		{"nil booking", Review{BookingID: uuid.Nil, Rating: 3}},
		{"zero rating", Review{BookingID: uuid.New(), Rating: 0}},
		{"rating too high", Review{BookingID: uuid.New(), Rating: 6}},
		{"negative rating", Review{BookingID: uuid.New(), Rating: -1}},
		{"invalid status", Review{BookingID: uuid.New(), Rating: 3, Status: "invalid"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.review.Validate(); err == nil {
				t.Error("expected error for invalid review")
			}
		})
	}
}

func TestReviewFilter(t *testing.T) {
	bathhouseID := uuid.New()
	status := ReviewStatusApproved
	minRating := 3

	filter := ReviewFilter{
		BathhouseID: &bathhouseID,
		Status:      &status,
		MinRating:   &minRating,
		Page:        1,
		PageSize:    20,
	}

	if *filter.BathhouseID != bathhouseID {
		t.Error("BathhouseID mismatch")
	}
	if *filter.Status != ReviewStatusApproved {
		t.Error("Status mismatch")
	}
	if *filter.MinRating != 3 {
		t.Error("MinRating mismatch")
	}
	if filter.Page != 1 {
		t.Error("Page mismatch")
	}
	if filter.PageSize != 20 {
		t.Error("PageSize mismatch")
	}
}

func TestCity_Validate(t *testing.T) {
	valid := &City{
		Name: "Moscow",
		Slug: "moscow",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid city returned error: %v", err)
	}

	tests := []struct {
		name string
		city City
	}{
		{"empty name", City{Name: "", Slug: "slug"}},
		{"empty slug", City{Name: "name", Slug: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.city.Validate(); err == nil {
				t.Error("expected error for invalid city")
			}
		})
	}
}

func TestDomainErrors(t *testing.T) {
	errors := []error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidInput,
		ErrUnauthorized,
		ErrForbidden,
		ErrSlotUnavailable,
		ErrBookingCancelLate,
		ErrUserBlocked,
		ErrBathhouseNotActive,
		ErrReviewAlreadyResponded,
		ErrSocialAccountAlreadyLinked,
		ErrSocialAccountNotFound,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("domain error should not be nil")
		}
		if err.Error() == "" {
			t.Error("domain error message should not be empty")
		}
	}
}
