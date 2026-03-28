package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSeasonalTariffValidate(t *testing.T) {
	bathhouseID := uuid.New()
	from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		tariff  *SeasonalTariff
		wantErr bool
	}{
		{
			name: "valid tariff",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Летний сезон",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  1.3,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "valid same-day tariff",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Один день",
				DateFrom:    from,
				DateTo:      from,
				Multiplier:  2.0,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "missing bathhouse_id",
			tariff: &SeasonalTariff{
				Name:       "Test",
				DateFrom:   from,
				DateTo:     to,
				Multiplier: 1.0,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  1.0,
			},
			wantErr: true,
		},
		{
			name: "date_to before date_from",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Invalid",
				DateFrom:    to,
				DateTo:      from,
				Multiplier:  1.0,
			},
			wantErr: true,
		},
		{
			name: "zero multiplier",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Zero",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  0,
			},
			wantErr: true,
		},
		{
			name: "negative multiplier",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Negative",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  -1.0,
			},
			wantErr: true,
		},
		{
			name: "multiplier exceeds 10",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Too High",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  10.01,
			},
			wantErr: true,
		},
		{
			name: "multiplier at max boundary (10.0)",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "Max",
				DateFrom:    from,
				DateTo:      to,
				Multiplier:  10.0,
			},
			wantErr: false,
		},
		{
			name: "zero date_from",
			tariff: &SeasonalTariff{
				BathhouseID: bathhouseID,
				Name:        "No From",
				DateTo:      to,
				Multiplier:  1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tariff.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSeasonalTariffAppliesToDate(t *testing.T) {
	tariff := &SeasonalTariff{
		BathhouseID: uuid.New(),
		Name:        "Summer",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}

	tests := []struct {
		name     string
		date     time.Time
		isActive bool
		expected bool
	}{
		{
			name:     "date within range",
			date:     time.Date(2024, 7, 15, 14, 0, 0, 0, time.UTC),
			isActive: true,
			expected: true,
		},
		{
			name:     "date at start boundary",
			date:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			isActive: true,
			expected: true,
		},
		{
			name:     "date at end boundary",
			date:     time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC),
			isActive: true,
			expected: true,
		},
		{
			name:     "date before range",
			date:     time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC),
			isActive: true,
			expected: false,
		},
		{
			name:     "date after range",
			date:     time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
			isActive: true,
			expected: false,
		},
		{
			name:     "inactive tariff does not apply",
			date:     time.Date(2024, 7, 15, 14, 0, 0, 0, time.UTC),
			isActive: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tariff.IsActive = tt.isActive
			assert.Equal(t, tt.expected, tariff.AppliesToDate(tt.date))
		})
	}
}
