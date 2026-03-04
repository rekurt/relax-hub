package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNotificationType_IsValid(t *testing.T) {
	tests := []struct {
		typ   NotificationType
		valid bool
	}{
		{NotifBookingConfirmed, true},
		{NotifBookingCancelled, true},
		{NotifNewReview, true},
		{NotifReviewResponse, true},
		{NotifPromo, true},
		{NotifReminder, true},
		{NotifSystem, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.typ.IsValid(); got != tt.valid {
			t.Errorf("NotificationType(%q).IsValid() = %v, want %v", tt.typ, got, tt.valid)
		}
	}
}

func TestNotification_Validate(t *testing.T) {
	valid := &Notification{
		UserID: uuid.New(),
		Type:   NotifBookingConfirmed,
		Title:  "Booking Confirmed",
		Body:   "Your booking has been confirmed.",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid notification returned error: %v", err)
	}

	tests := []struct {
		name string
		n    Notification
	}{
		{"nil user", Notification{UserID: uuid.Nil, Type: NotifSystem, Title: "t", Body: "b"}},
		{"invalid type", Notification{UserID: uuid.New(), Type: "bad", Title: "t", Body: "b"}},
		{"empty title", Notification{UserID: uuid.New(), Type: NotifSystem, Title: "", Body: "b"}},
		{"empty body", Notification{UserID: uuid.New(), Type: NotifSystem, Title: "t", Body: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Validate(); err == nil {
				t.Error("expected error for invalid notification")
			}
		})
	}
}

func TestNotificationPreferences_Validate(t *testing.T) {
	valid := &NotificationPreferences{UserID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid preferences returned error: %v", err)
	}

	invalid := &NotificationPreferences{UserID: uuid.Nil}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for nil user ID")
	}
}

func TestDefaultNotificationPreferences(t *testing.T) {
	userID := uuid.New()
	prefs := DefaultNotificationPreferences(userID)

	if prefs.UserID != userID {
		t.Errorf("UserID = %v, want %v", prefs.UserID, userID)
	}
	if !prefs.InApp {
		t.Error("InApp should default to true")
	}
	if !prefs.Email {
		t.Error("Email should default to true")
	}
	if prefs.Push {
		t.Error("Push should default to false")
	}
	if !prefs.BookingEvents {
		t.Error("BookingEvents should default to true")
	}
	if !prefs.ReviewEvents {
		t.Error("ReviewEvents should default to true")
	}
	if !prefs.PromoEvents {
		t.Error("PromoEvents should default to true")
	}
	if !prefs.Reminders {
		t.Error("Reminders should default to true")
	}
}

func TestNotificationPreferences_WantsEventType(t *testing.T) {
	prefs := &NotificationPreferences{
		UserID:        uuid.New(),
		BookingEvents: true,
		ReviewEvents:  false,
		PromoEvents:   true,
		Reminders:     false,
	}

	tests := []struct {
		typ  NotificationType
		want bool
	}{
		{NotifBookingConfirmed, true},
		{NotifBookingCancelled, true},
		{NotifNewReview, false},
		{NotifReviewResponse, false},
		{NotifPromo, true},
		{NotifReminder, false},
		{NotifSystem, true}, // system always true
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if got := prefs.WantsEventType(tt.typ); got != tt.want {
				t.Errorf("WantsEventType(%q) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}
