package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestSeasonalTariff_CRUD(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create
	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Летний сезон",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}
	created, err := svc.CreateSeasonalTariff(context.Background(), ownerID, domain.RoleOwner, tariff)
	if err != nil {
		t.Fatalf("unexpected error creating tariff: %v", err)
	}
	if created.Name != "Летний сезон" {
		t.Errorf("expected name 'Летний сезон', got '%s'", created.Name)
	}

	// List
	tariffs, err := svc.ListSeasonalTariffs(context.Background(), bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error listing tariffs: %v", err)
	}
	if len(tariffs) != 1 {
		t.Errorf("expected 1 tariff, got %d", len(tariffs))
	}

	// Update
	tariff.Name = "Летний сезон (обновлено)"
	tariff.Multiplier = 1.5
	err = svc.UpdateSeasonalTariff(context.Background(), ownerID, domain.RoleOwner, tariff)
	if err != nil {
		t.Fatalf("unexpected error updating tariff: %v", err)
	}

	// Delete
	err = svc.DeleteSeasonalTariff(context.Background(), ownerID, domain.RoleOwner, tariff.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting tariff: %v", err)
	}

	tariffs, err = svc.ListSeasonalTariffs(context.Background(), bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error listing tariffs: %v", err)
	}
	if len(tariffs) != 0 {
		t.Errorf("expected 0 tariffs after delete, got %d", len(tariffs))
	}
}

func TestSeasonalTariff_UnauthorizedUser(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Unauthorized",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}

	_, err := svc.CreateSeasonalTariff(context.Background(), otherUserID, domain.RoleOwner, tariff)
	if err == nil {
		t.Error("expected authorization error, got nil")
	}
}

func TestSeasonalTariff_InvalidInput(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Multiplier 0 = invalid
	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Invalid",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  0,
		IsActive:    true,
	}

	_, err := svc.CreateSeasonalTariff(context.Background(), ownerID, domain.RoleOwner, tariff)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestSeasonalTariff_PriceCalculation_SingleTariff(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create a summer tariff: 1.3x
	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Summer",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}
	if err := tariffRepo.Create(context.Background(), tariff); err != nil {
		t.Fatalf("failed to create tariff: %v", err)
	}

	// Booking in summer: 3 hours at 1000/h
	start := time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 7, 15, 13, 0, 0, 0, time.UTC)

	total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 2,
		BaseCapacity:               5,
		ExtraGuestSurcharge:        0,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base: 1000 * 1.3 = 1300/h, 3 hours = 3900
	if total != 3900 {
		t.Errorf("expected total 3900, got %d", total)
	}
	if !breakdown.IsSeasonalPrice {
		t.Error("expected IsSeasonalPrice to be true")
	}
	if breakdown.SeasonalTariffName != "Summer" {
		t.Errorf("expected tariff name 'Summer', got '%s'", breakdown.SeasonalTariffName)
	}
	if breakdown.SeasonalTariffMultiplier != 1.3 {
		t.Errorf("expected multiplier 1.3, got %f", breakdown.SeasonalTariffMultiplier)
	}
}

func TestSeasonalTariff_PriceCalculation_NoTariffApplied(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create summer tariff, but book in winter
	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Summer",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}
	if err := tariffRepo.Create(context.Background(), tariff); err != nil {
		t.Fatalf("failed to create tariff: %v", err)
	}

	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)

	total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 2,
		BaseCapacity:               5,
		ExtraGuestSurcharge:        0,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No tariff applied: 1000/h * 3h = 3000
	if total != 3000 {
		t.Errorf("expected total 3000, got %d", total)
	}
	if breakdown.IsSeasonalPrice {
		t.Error("expected IsSeasonalPrice to be false")
	}
}

func TestSeasonalTariff_PriceCalculation_OverlappingTariffs(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Summer: June-August, 1.3x
	summer := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Summer",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.3,
		IsActive:    true,
	}
	// Peak summer: July 1-31, 1.8x (overlaps with summer)
	peak := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Peak Summer",
		DateFrom:    time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 7, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  1.8,
		IsActive:    true,
	}

	for _, st := range []*domain.SeasonalTariff{summer, peak} {
		if err := tariffRepo.Create(context.Background(), st); err != nil {
			t.Fatalf("failed to create tariff: %v", err)
		}
	}

	// Book in July - should pick highest multiplier (1.8x)
	start := time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 7, 15, 13, 0, 0, 0, time.UTC)

	total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 2,
		BaseCapacity:               5,
		ExtraGuestSurcharge:        0,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1000 * 1.8 = 1800/h * 3h = 5400
	if total != 5400 {
		t.Errorf("expected total 5400, got %d", total)
	}
	if breakdown.SeasonalTariffName != "Peak Summer" {
		t.Errorf("expected 'Peak Summer', got '%s'", breakdown.SeasonalTariffName)
	}
	if breakdown.SeasonalTariffMultiplier != 1.8 {
		t.Errorf("expected multiplier 1.8, got %f", breakdown.SeasonalTariffMultiplier)
	}

	// Book in June (only base summer tariff applies: 1.3x)
	start = time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end = time.Date(2024, 6, 15, 13, 0, 0, 0, time.UTC)

	total, breakdown, err = svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 2,
		BaseCapacity:               5,
		ExtraGuestSurcharge:        0,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1000 * 1.3 = 1300/h * 3h = 3900
	if total != 3900 {
		t.Errorf("expected total 3900, got %d", total)
	}
	if breakdown.SeasonalTariffName != "Summer" {
		t.Errorf("expected 'Summer', got '%s'", breakdown.SeasonalTariffName)
	}
}

func TestSeasonalTariff_PriceCalculation_InactiveTariffIgnored(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Inactive Summer",
		DateFrom:    time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
		Multiplier:  2.0,
		IsActive:    false, // inactive
	}
	if err := tariffRepo.Create(context.Background(), tariff); err != nil {
		t.Fatalf("failed to create tariff: %v", err)
	}

	start := time.Date(2024, 7, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 7, 15, 13, 0, 0, 0, time.UTC)

	total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 2,
		BaseCapacity:               5,
		ExtraGuestSurcharge:        0,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Inactive tariff should be ignored
	if total != 3000 {
		t.Errorf("expected total 3000, got %d", total)
	}
	if breakdown.IsSeasonalPrice {
		t.Error("expected IsSeasonalPrice to be false for inactive tariff")
	}
}

func TestSeasonalTariff_PriceCalculation_BoundaryDates(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	tariffRepo := mock.NewSeasonalTariffRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	svc := NewPricingService(priceRepo, tariffRepo, bhRepo, nil, nil, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// New Year tariff: Dec 31 - Jan 8
	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "New Year",
		DateFrom:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		DateTo:      time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Multiplier:  2.0,
		IsActive:    true,
	}
	if err := tariffRepo.Create(context.Background(), tariff); err != nil {
		t.Fatalf("failed to create tariff: %v", err)
	}

	tests := []struct {
		name          string
		date          time.Time
		expectedTotal int64
		isSeasonal    bool
	}{
		{
			name:          "Dec 31 - first day applies",
			date:          time.Date(2024, 12, 31, 10, 0, 0, 0, time.UTC),
			expectedTotal: 6000, // 1000 * 2.0 * 3h
			isSeasonal:    true,
		},
		{
			name:          "Jan 8 - last day applies",
			date:          time.Date(2025, 1, 8, 10, 0, 0, 0, time.UTC),
			expectedTotal: 6000,
			isSeasonal:    true,
		},
		{
			name:          "Dec 30 - before range",
			date:          time.Date(2024, 12, 30, 10, 0, 0, 0, time.UTC),
			expectedTotal: 3000, // base price
			isSeasonal:    false,
		},
		{
			name:          "Jan 9 - after range",
			date:          time.Date(2025, 1, 9, 10, 0, 0, 0, time.UTC),
			expectedTotal: 3000,
			isSeasonal:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := tt.date
			end := start.Add(3 * time.Hour)

			total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
				BathhouseID:                bathhouseID,
				BasePrice:                  1000,
				StartTime:                  start,
				EndTime:                    end,
				GuestCount:                 2,
				BaseCapacity:               5,
				ExtraGuestSurcharge:        0,
				LongSessionThresholdHours:  4,
				LongSessionDiscountPercent: 0,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}
			if breakdown.IsSeasonalPrice != tt.isSeasonal {
				t.Errorf("expected IsSeasonalPrice %v, got %v", tt.isSeasonal, breakdown.IsSeasonalPrice)
			}
		})
	}
}
