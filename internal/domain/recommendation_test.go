package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserPreferencesValidate(t *testing.T) {
	validUserID := uuid.New()
	cityID := int64(1)
	priceMin := int64(100000) // kopecks
	priceMax := int64(500000)

	tests := []struct {
		name    string
		prefs   UserPreferences
		wantErr bool
	}{
		{
			name: "valid preferences",
			prefs: UserPreferences{
				UserID:          validUserID,
				PreferredCityID: &cityID,
				PriceRangeMin:   &priceMin,
				PriceRangeMax:   &priceMax,
				PreferPool:      true,
				PreferSauna:     true,
			},
			wantErr: false,
		},
		{
			name: "valid preferences without optional fields",
			prefs: UserPreferences{
				UserID:      validUserID,
				PreferPool:  false,
				PreferSauna: false,
			},
			wantErr: false,
		},
		{
			name: "nil user id",
			prefs: UserPreferences{
				UserID: uuid.Nil,
			},
			wantErr: true,
		},
		{
			name: "negative city id",
			prefs: UserPreferences{
				UserID:          validUserID,
				PreferredCityID: &[]int64{-1}[0],
			},
			wantErr: true,
		},
		{
			name: "negative price range",
			prefs: UserPreferences{
				UserID:        validUserID,
				PriceRangeMin: &[]int64{-1000}[0],
				PriceRangeMax: &priceMax,
			},
			wantErr: true,
		},
		{
			name: "min price greater than max price",
			prefs: UserPreferences{
				UserID:        validUserID,
				PriceRangeMin: &[]int64{600000}[0],
				PriceRangeMax: &priceMax,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prefs.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserActivityValidate(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()

	tests := []struct {
		name    string
		activity UserActivity
		wantErr bool
	}{
		{
			name: "valid view activity",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      userID,
				BathhouseID: bathhouseID,
				Type:        ActivityTypeView,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid booking activity",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      userID,
				BathhouseID: bathhouseID,
				Type:        ActivityTypeBooking,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid favorite activity",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      userID,
				BathhouseID: bathhouseID,
				Type:        ActivityTypeFavorite,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "nil user id",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      uuid.Nil,
				BathhouseID: bathhouseID,
				Type:        ActivityTypeView,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "nil bathhouse id",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      userID,
				BathhouseID: uuid.Nil,
				Type:        ActivityTypeView,
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid activity type",
			activity: UserActivity{
				ID:          uuid.New(),
				UserID:      userID,
				BathhouseID: bathhouseID,
				Type:        UserActivityType("invalid"),
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.activity.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserActivityTypeIsValid(t *testing.T) {
	tests := []struct {
		name  string
		aType UserActivityType
		want  bool
	}{
		{
			name:  "view activity type",
			aType: ActivityTypeView,
			want:  true,
		},
		{
			name:  "booking activity type",
			aType: ActivityTypeBooking,
			want:  true,
		},
		{
			name:  "favorite activity type",
			aType: ActivityTypeFavorite,
			want:  true,
		},
		{
			name:  "invalid activity type",
			aType: UserActivityType("invalid"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aType.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
