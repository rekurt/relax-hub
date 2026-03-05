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

var testPricingLogger = logger.New(logger.LevelError)

func TestPricingService_CalculatePrice_NoRules(t *testing.T) {
	// When there are no pricing rules, price should be basePrice * hours
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)

	service := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	startTime := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 15, 13, 0, 0, 0, time.UTC) // 3 hours

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(3000) // 1000 * 3 hours
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}

func TestPricingService_CalculatePrice_SingleRule(t *testing.T) {
	// Apply a single rule with 1.5x multiplier
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)

	service := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Create a weekend rule (1.5x multiplier)
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Weekend",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  1.5,
		DaysOfWeek:  []int{5, 6}, // Saturday and Sunday
		Priority:    10,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), rule)

	// Saturday, 10:00-13:00 (3 hours)
	startTime := time.Date(2024, 3, 16, 10, 0, 0, 0, time.UTC) // Saturday
	endTime := time.Date(2024, 3, 16, 13, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(4500) // 1000 * 3 * 1.5
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}

func TestPricingService_CalculatePrice_PriorityConflict(t *testing.T) {
	// When multiple rules apply, highest priority wins
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)

	service := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Weekend rule: 1.5x (priority 10)
	weekendRule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Weekend",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  1.5,
		DaysOfWeek:  []int{5, 6},
		Priority:    10,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), weekendRule)

	// Saturday Happy Hour: 0.8x (priority 20) - should win
	from := "18:00"
	to := "20:00"
	happyHourRule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Happy Hour",
		Type:        domain.RuleTypeTimeRange,
		Multiplier:  0.8,
		TimeFrom:    &from,
		TimeTo:      &to,
		Priority:    20,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), happyHourRule)

	// Saturday 18:00-20:00 (2 hours, during happy hour)
	startTime := time.Date(2024, 3, 16, 18, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 16, 20, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use happy hour multiplier (0.8) because it has higher priority
	expected := int64(1600) // 1000 * 2 * 0.8
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}

func TestPricingService_CalculatePrice_PartialRuleApplication(t *testing.T) {
	// Test when a rule applies to only part of the booking period
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)

	service := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Happy hour 18:00-20:00 with 0.8x
	from := "18:00"
	to := "20:00"
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Happy Hour",
		Type:        domain.RuleTypeTimeRange,
		Multiplier:  0.8,
		TimeFrom:    &from,
		TimeTo:      &to,
		Priority:    10,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), rule)

	// Book from 17:00-21:00 (4 hours)
	// 17:00-18:00: base price (1000)
	// 18:00-20:00: happy hour (800 each, 2 hours)
	// 20:00-21:00: base price (1000)
	startTime := time.Date(2024, 3, 15, 17, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 15, 21, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(3600) // 1000 + 800 + 800 + 1000
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}

func TestPricingService_CreateRule_InvalidInput(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)

	service := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Create rule with invalid multiplier
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Invalid",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  0, // invalid
		DaysOfWeek:  []int{5, 6},
		Priority:    10,
		IsActive:    true,
	}

	_, err := service.CreateRule(context.Background(), bh.OwnerID, domain.RoleOwner, rule)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestPricingService_CreateRule_UnauthorizedUser(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	ownerID := uuid.New()
	unauthorizedID := uuid.New()

	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	bhRepo.Create(context.Background(), bh)

	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Unauthorized",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  1.5,
		DaysOfWeek:  []int{5, 6},
		Priority:    10,
		IsActive:    true,
	}

	_, err := service.CreateRule(context.Background(), unauthorizedID, domain.RoleOwner, rule)
	if err == nil {
		t.Error("expected authorization error, got nil")
	}
}

func TestPricingService_ListRules(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Create multiple rules
	for i := 0; i < 3; i++ {
		rule := &domain.PricingRule{
			ID:          uuid.New(),
			BathhouseID: bathhouseID,
			Name:        "Rule",
			Type:        domain.RuleTypeWeekend,
			Multiplier:  1.5,
			DaysOfWeek:  []int{5, 6},
			Priority:    10,
			IsActive:    true,
			CreatedAt:   time.Now(),
		}
		priceRepo.Create(context.Background(), rule)
	}

	rules, err := service.ListRules(context.Background(), bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rules))
	}
}

func TestPricingService_CalculatePrice_HolidayRule(t *testing.T) {
	// Test holiday pricing rules
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// New Year holidays: 2x multiplier
	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 8, 23, 59, 59, 0, time.UTC)
	holidayRule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "New Year",
		Type:        domain.RuleTypeHoliday,
		Multiplier:  2.0,
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		Priority:    20,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), holidayRule)

	// Book on Jan 3, 10:00-13:00
	startTime := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 3, 13, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(6000) // 1000 * 3 * 2.0
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}

func TestPricingService_CalculatePrice_InactiveRule(t *testing.T) {
	// Inactive rules should not be applied
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	bhRepo.Create(context.Background(), bh)

	// Create an inactive rule
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Inactive",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  2.0,
		DaysOfWeek:  []int{5, 6},
		Priority:    10,
		IsActive:    false, // inactive
		CreatedAt:   time.Now(),
	}
	priceRepo.Create(context.Background(), rule)

	// Saturday booking should use base price, not the inactive rule
	startTime := time.Date(2024, 3, 16, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 16, 13, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(3000) // 1000 * 3, no multiplier
	if price != expected {
		t.Errorf("expected price %d, got %d", expected, price)
	}
}
