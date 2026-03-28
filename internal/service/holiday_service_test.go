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

var testHolidayLogger = logger.New(logger.LevelError)

func newHolidayService() (HolidayService, *mock.HolidayRepo) {
	repo := mock.NewHolidayRepo().(*mock.HolidayRepo)
	access := NewAccessChecker(mock.NewRepresentativeRepo(), mock.NewBathhouseRepo())
	svc := NewHolidayService(repo, access, testHolidayLogger)
	return svc, repo
}

func TestHolidayService_IsHolidayDate_Recurring(t *testing.T) {
	svc, repo := newHolidayService()
	ctx := context.Background()

	// Create a recurring holiday (New Year on Jan 1)
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Новый год",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := repo.Create(ctx, holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	// Test recurring match - different year should still match
	h, mult, err := svc.IsHolidayDate(ctx, time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), "RU", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected holiday, got nil")
	}
	if h.Name != "Новый год" {
		t.Errorf("expected name 'Новый год', got '%s'", h.Name)
	}
	if mult != domain.DefaultHolidayMultiplier {
		t.Errorf("expected default multiplier %f, got %f", domain.DefaultHolidayMultiplier, mult)
	}

	// Test non-holiday date
	h, _, err = svc.IsHolidayDate(ctx, time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC), "RU", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h != nil {
		t.Errorf("expected nil for non-holiday date, got %v", h)
	}
}

func TestHolidayService_IsHolidayDate_NonRecurring(t *testing.T) {
	svc, repo := newHolidayService()
	ctx := context.Background()

	// Create a non-recurring holiday
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Особый день",
		Date:        time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: false,
	}
	if err := repo.Create(ctx, holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	// Exact date match should work
	h, _, err := svc.IsHolidayDate(ctx, time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC), "RU", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected holiday, got nil")
	}

	// Different year should NOT match (not recurring)
	h, _, err = svc.IsHolidayDate(ctx, time.Date(2027, 6, 15, 14, 0, 0, 0, time.UTC), "RU", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h != nil {
		t.Errorf("expected nil for non-recurring holiday in different year, got %v", h)
	}
}

func TestHolidayService_IsHolidayDate_RegionFilter(t *testing.T) {
	svc, repo := newHolidayService()
	ctx := context.Background()

	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "День России",
		Date:        time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := repo.Create(ctx, holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	// Should match for RU
	h, _, err := svc.IsHolidayDate(ctx, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC), "RU", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected holiday for RU region, got nil")
	}

	// Should NOT match for BY
	h, _, err = svc.IsHolidayDate(ctx, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC), "BY", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h != nil {
		t.Errorf("expected nil for BY region, got %v", h)
	}
}

func TestHolidayService_BathhouseMultiplierOverride(t *testing.T) {
	svc, repo := newHolidayService()
	ctx := context.Background()

	// Create holiday
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Новый год",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := repo.Create(ctx, holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	bathhouseID := uuid.New()

	// Default multiplier
	_, mult, err := svc.IsHolidayDate(ctx, time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), "RU", bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mult != 1.5 {
		t.Errorf("expected default multiplier 1.5, got %f", mult)
	}

	// Set custom multiplier
	if err := repo.SetBathhouseMultiplier(ctx, bathhouseID, 1.3); err != nil {
		t.Fatalf("failed to set multiplier: %v", err)
	}

	// Custom multiplier should be returned
	_, mult, err = svc.IsHolidayDate(ctx, time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), "RU", bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mult != 1.3 {
		t.Errorf("expected custom multiplier 1.3, got %f", mult)
	}
}

func TestHolidayService_CRUD(t *testing.T) {
	svc, _ := newHolidayService()
	ctx := context.Background()

	// Create
	holiday := &domain.Holiday{
		Name:        "Тестовый праздник",
		Date:        time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: false,
	}
	if err := svc.Create(ctx, holiday); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if holiday.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	// GetByID
	got, err := svc.GetByID(ctx, holiday.ID)
	if err != nil {
		t.Fatalf("getByID failed: %v", err)
	}
	if got.Name != "Тестовый праздник" {
		t.Errorf("expected name 'Тестовый праздник', got '%s'", got.Name)
	}

	// Update
	holiday.Name = "Обновленный праздник"
	if err := svc.Update(ctx, holiday); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ = svc.GetByID(ctx, holiday.ID)
	if got.Name != "Обновленный праздник" {
		t.Errorf("expected updated name, got '%s'", got.Name)
	}

	// ListAll
	list, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("listAll failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 holiday, got %d", len(list))
	}

	// Delete
	if err := svc.Delete(ctx, holiday.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = svc.GetByID(ctx, holiday.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestHolidayService_Validate(t *testing.T) {
	svc, _ := newHolidayService()
	ctx := context.Background()

	tests := []struct {
		name    string
		holiday *domain.Holiday
		wantErr bool
	}{
		{
			name:    "empty name",
			holiday: &domain.Holiday{Name: "", Date: time.Now(), Region: "RU"},
			wantErr: true,
		},
		{
			name:    "invalid region",
			holiday: &domain.Holiday{Name: "Test", Date: time.Now(), Region: "XX"},
			wantErr: true,
		},
		{
			name:    "valid holiday",
			holiday: &domain.Holiday{Name: "Test", Date: time.Now(), Region: "RU"},
			wantErr: false,
		},
		{
			name:    "valid BY region",
			holiday: &domain.Holiday{Name: "Test", Date: time.Now(), Region: "BY"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(ctx, tt.holiday)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test holiday pricing integration with PricingService
func TestPricingService_CalculateFullPrice_HolidayMultiplier(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	holidayRepo := mock.NewHolidayRepo().(*mock.HolidayRepo)
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	// Create holiday service with a holiday
	holidaySvc := NewHolidayService(holidayRepo, access, log)
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Новый год",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := holidayRepo.Create(context.Background(), holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	// Create pricing service with holiday service
	svc := NewPricingService(priceRepo, nil, bhRepo, holidaySvc, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Booking on Jan 1 (holiday) with base price 1000 kopecks/h, 3 hours
	// Default multiplier = 1.5
	// Holiday-adjusted base = 1000 * 1.5 = 1500/h
	// Total = 1500 * 3 = 4500
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
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

	if !breakdown.IsHolidayPrice {
		t.Error("expected IsHolidayPrice to be true")
	}
	if breakdown.HolidayName != "Новый год" {
		t.Errorf("expected holiday name 'Новый год', got '%s'", breakdown.HolidayName)
	}
	if breakdown.HolidayMultiplier != 1.5 {
		t.Errorf("expected multiplier 1.5, got %f", breakdown.HolidayMultiplier)
	}
	expectedBase := int64(4500) // 1500/h * 3h
	if breakdown.BasePrice != expectedBase {
		t.Errorf("expected base price %d, got %d", expectedBase, breakdown.BasePrice)
	}
	if total != expectedBase {
		t.Errorf("expected total %d, got %d", expectedBase, total)
	}
}

func TestPricingService_CalculateFullPrice_HolidayWithDynamicRule(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	holidayRepo := mock.NewHolidayRepo().(*mock.HolidayRepo)
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	holidaySvc := NewHolidayService(holidayRepo, access, log)
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Новый год",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := holidayRepo.Create(context.Background(), holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	svc := NewPricingService(priceRepo, nil, bhRepo, holidaySvc, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Add a holiday pricing rule with 2.0x (this stacks with the holiday multiplier)
	dateFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 1, 8, 23, 59, 59, 0, time.UTC)
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "New Year pricing",
		Type:        domain.RuleTypeHoliday,
		Multiplier:  2.0,
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		Priority:    20,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Booking on Jan 1: holiday multiplier 1.5 applied to base price first
	// Holiday-adjusted base = 1000 * 1.5 = 1500/h
	// Then dynamic rule 2.0x applied per hour: 1500 * 2.0 = 3000/h
	// Total = 3000 * 3 = 9000
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
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

	if !breakdown.IsHolidayPrice {
		t.Error("expected IsHolidayPrice to be true")
	}
	expectedBase := int64(9000) // 1500/h * 2.0 rule * 3h
	if breakdown.BasePrice != expectedBase {
		t.Errorf("expected base price %d, got %d", expectedBase, breakdown.BasePrice)
	}
	if total != expectedBase {
		t.Errorf("expected total %d, got %d", expectedBase, total)
	}
}

func TestPricingService_CalculateFullPrice_NonHolidayDate(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	holidayRepo := mock.NewHolidayRepo().(*mock.HolidayRepo)
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	holidaySvc := NewHolidayService(holidayRepo, access, log)
	// Create a holiday but check a non-holiday date
	holiday := &domain.Holiday{
		ID:          uuid.New(),
		Name:        "Новый год",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Region:      "RU",
		IsRecurring: true,
	}
	if err := holidayRepo.Create(context.Background(), holiday); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	svc := NewPricingService(priceRepo, nil, bhRepo, holidaySvc, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Booking on March 15 (not a holiday)
	start := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
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

	if breakdown.IsHolidayPrice {
		t.Error("expected IsHolidayPrice to be false for non-holiday date")
	}
	expectedBase := int64(3000) // 1000/h * 3h, no multiplier
	if breakdown.BasePrice != expectedBase {
		t.Errorf("expected base price %d, got %d", expectedBase, breakdown.BasePrice)
	}
	if total != expectedBase {
		t.Errorf("expected total %d, got %d", expectedBase, total)
	}
}
