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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create pricing rule: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), weekendRule); err != nil {
		t.Fatalf("failed to create weekend rule: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), happyHourRule); err != nil {
		t.Fatalf("failed to create happy hour rule: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
		if err := priceRepo.Create(context.Background(), rule); err != nil {
			t.Fatalf("failed to create rule: %v", err)
		}
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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), holidayRule); err != nil {
		t.Fatalf("failed to create holiday rule: %v", err)
	}

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
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

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
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

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

func TestPricingService_CalculatePrice_WraparoundTimeRange(t *testing.T) {
	// Test overnight wraparound times (e.g., 22:00-06:00)
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Overnight rule: 22:00-06:00 with 1.5x multiplier
	from := "22:00"
	to := "06:00"
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Overnight",
		Type:        domain.RuleTypeTimeRange,
		Multiplier:  1.5,
		TimeFrom:    &from,
		TimeTo:      &to,
		Priority:    10,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create overnight rule: %v", err)
	}

	// Test 1: Booking at 23:00-01:00 (inside wraparound window)
	startTime := time.Date(2024, 3, 15, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 16, 1, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error in wraparound test: %v", err)
	}

	expected := int64(3000) // 1000 * 2 * 1.5
	if price != expected {
		t.Errorf("expected price %d for 23:00-01:00 booking, got %d", expected, price)
	}

	// Test 2: Booking at 05:00-06:00 (end of wraparound window)
	startTime = time.Date(2024, 3, 15, 5, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, 3, 15, 6, 0, 0, 0, time.UTC)

	price, err = service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error in wraparound end test: %v", err)
	}

	expected = int64(1500) // 1000 * 1 * 1.5
	if price != expected {
		t.Errorf("expected price %d for 05:00-06:00 booking, got %d", expected, price)
	}

	// Test 3: Booking at 20:00-21:00 (outside wraparound window)
	startTime = time.Date(2024, 3, 15, 20, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, 3, 15, 21, 0, 0, 0, time.UTC)

	price, err = service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error in non-wraparound test: %v", err)
	}

	expected = int64(1000) // 1000 * 1 * 1.0 (no multiplier)
	if price != expected {
		t.Errorf("expected price %d for 20:00-21:00 booking, got %d", expected, price)
	}
}

func TestPricingService_UpdateRule(t *testing.T) {
	// Test updating a pricing rule
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	ownerID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create initial rule
	rule := &domain.PricingRule{
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
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Update the rule
	updatedRule := &domain.PricingRule{
		ID:          rule.ID,
		BathhouseID: bathhouseID,
		Name:        "Weekend Updated",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  2.0,
		DaysOfWeek:  []int{5, 6},
		Priority:    20,
		IsActive:    true,
	}

	err := service.UpdateRule(context.Background(), ownerID, domain.RoleOwner, updatedRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the update
	retrieved, err := priceRepo.GetByID(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("failed to retrieve rule: %v", err)
	}

	if retrieved.Name != "Weekend Updated" {
		t.Errorf("expected name 'Weekend Updated', got '%s'", retrieved.Name)
	}
	if retrieved.Multiplier != 2.0 {
		t.Errorf("expected multiplier 2.0, got %f", retrieved.Multiplier)
	}
	if retrieved.Priority != 20 {
		t.Errorf("expected priority 20, got %d", retrieved.Priority)
	}
}

func TestPricingService_DeleteRule(t *testing.T) {
	// Test deleting a pricing rule
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	ownerID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create rule
	rule := &domain.PricingRule{
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
	if err := priceRepo.Create(context.Background(), rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Delete the rule
	err := service.DeleteRule(context.Background(), ownerID, domain.RoleOwner, rule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion
	_, err = priceRepo.GetByID(context.Background(), rule.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPricingService_GetActiveRules(t *testing.T) {
	// Test getting only active rules
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Create active rules
	activeRule1 := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Active 1",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  1.5,
		DaysOfWeek:  []int{5},
		Priority:    10,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	activeRule2 := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Active 2",
		Type:        domain.RuleTypeWeekday,
		Multiplier:  1.2,
		DaysOfWeek:  []int{0, 1, 2, 3, 4},
		Priority:    5,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	// Create inactive rule
	inactiveRule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Inactive",
		Type:        domain.RuleTypeWeekend,
		Multiplier:  2.0,
		DaysOfWeek:  []int{6},
		Priority:    15,
		IsActive:    false,
		CreatedAt:   time.Now(),
	}

	for _, rule := range []*domain.PricingRule{activeRule1, activeRule2, inactiveRule} {
		if err := priceRepo.Create(context.Background(), rule); err != nil {
			t.Fatalf("failed to create rule: %v", err)
		}
	}

	// Get active rules
	activeRules, err := service.GetActiveRules(context.Background(), bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activeRules) != 2 {
		t.Errorf("expected 2 active rules, got %d", len(activeRules))
	}

	// Verify all returned rules are active
	for _, rule := range activeRules {
		if !rule.IsActive {
			t.Errorf("returned inactive rule: %v", rule)
		}
	}
}

func TestPricingService_CalculatePrice_DateRangeBoundary(t *testing.T) {
	// Test that date range rules compare dates only, not timestamps
	// This ensures rules with DateFrom/DateTo boundaries work correctly
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)
	log := logger.New(logger.LevelError)

	service := NewPricingService(priceRepo, bhRepo, access, log)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// Holiday rule: Jan 1 to Jan 8 with 2.0x multiplier
	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 8, 23, 59, 59, 0, time.UTC)
	holidayRule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        "Holiday",
		Type:        domain.RuleTypeHoliday,
		Multiplier:  2.0,
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		Priority:    20,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	if err := priceRepo.Create(context.Background(), holidayRule); err != nil {
		t.Fatalf("failed to create holiday rule: %v", err)
	}

	// Test 1: Booking on Jan 1 at 08:00 (within the range) - should apply rule
	startTime := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	price, err := service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int64(6000) // 1000 * 3 * 2.0
	if price != expected {
		t.Errorf("expected price %d for Jan 1 booking, got %d", expected, price)
	}

	// Test 2: Booking on Jan 8 at 20:00 (last day at evening) - should apply rule
	startTime = time.Date(2024, 1, 8, 20, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, 1, 8, 23, 0, 0, 0, time.UTC)

	price, err = service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = int64(6000) // 1000 * 3 * 2.0
	if price != expected {
		t.Errorf("expected price %d for Jan 8 evening booking, got %d", expected, price)
	}

	// Test 3: Booking on Jan 9 (after range) - should NOT apply rule
	startTime = time.Date(2024, 1, 9, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, 1, 9, 13, 0, 0, 0, time.UTC)

	price, err = service.CalculatePrice(context.Background(), bathhouseID, 1000, startTime, endTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = int64(3000) // 1000 * 3 * 1.0 (no multiplier)
	if price != expected {
		t.Errorf("expected price %d for Jan 9 booking, got %d", expected, price)
	}
}

func TestPricingService_CalculateFullPrice_LongSessionDiscount(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)

	svc := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	tests := []struct {
		name                string
		durationHours       int
		thresholdHours      int
		discountPercent     int
		basePrice           int64
		expectedDiscount    int64
		expectedTotal       int64
	}{
		{
			name:            "3h booking with 4h threshold = no discount",
			durationHours:   3,
			thresholdHours:  4,
			discountPercent: 10,
			basePrice:       1000,
			expectedDiscount: 0,
			expectedTotal:   3000,
		},
		{
			name:            "6h booking with 4h threshold and 10% = discount on 2h",
			durationHours:   6,
			thresholdHours:  4,
			discountPercent: 10,
			basePrice:       1000,
			expectedDiscount: 200, // avgRate=1000, 2 discountable hours * 1000 * 10% = 200
			expectedTotal:   5800,
		},
		{
			name:            "8h booking with 4h threshold and 20% = discount on 4h",
			durationHours:   8,
			thresholdHours:  4,
			discountPercent: 20,
			basePrice:       1000,
			expectedDiscount: 800, // avgRate=1000, 4 discountable hours * 1000 * 20% = 800
			expectedTotal:   7200,
		},
		{
			name:            "4h booking at exactly threshold = no discount (not beyond threshold)",
			durationHours:   4,
			thresholdHours:  4,
			discountPercent: 10,
			basePrice:       1000,
			expectedDiscount: 0,
			expectedTotal:   4000,
		},
		{
			name:            "5h booking with 0% discount = no discount",
			durationHours:   5,
			thresholdHours:  4,
			discountPercent: 0,
			basePrice:       1000,
			expectedDiscount: 0,
			expectedTotal:   5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
			end := start.Add(time.Duration(tt.durationHours) * time.Hour)

			total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
				BathhouseID:                bathhouseID,
				BasePrice:                  tt.basePrice,
				StartTime:                  start,
				EndTime:                    end,
				GuestCount:                 2,
				BaseCapacity:               5,
				ExtraGuestSurcharge:        0,
				LongSessionThresholdHours:  tt.thresholdHours,
				LongSessionDiscountPercent: tt.discountPercent,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if breakdown.LongSessionDiscount != tt.expectedDiscount {
				t.Errorf("expected discount %d, got %d", tt.expectedDiscount, breakdown.LongSessionDiscount)
			}
			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestPricingService_CalculateFullPrice_ExtraGuestSurcharge(t *testing.T) {
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)

	svc := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	tests := []struct {
		name               string
		guestCount         int
		baseCapacity       int
		surchargePerGuest  int64
		durationHours      int
		basePrice          int64
		expectedSurcharge  int64
		expectedTotal      int64
	}{
		{
			name:              "2 guests with base_capacity=2 = no surcharge",
			guestCount:        2,
			baseCapacity:      2,
			surchargePerGuest: 500,
			durationHours:     3,
			basePrice:         1000,
			expectedSurcharge: 0,
			expectedTotal:     3000,
		},
		{
			name:              "5 guests with base_capacity=3 = 2 extra * 500 * 3h",
			guestCount:        5,
			baseCapacity:      3,
			surchargePerGuest: 500,
			durationHours:     3,
			basePrice:         1000,
			expectedSurcharge: 3000, // 2 extra * 500 * 3 hours
			expectedTotal:     6000, // 3000 base + 3000 surcharge
		},
		{
			name:              "1 extra guest for 1 hour",
			guestCount:        4,
			baseCapacity:      3,
			surchargePerGuest: 200,
			durationHours:     1,
			basePrice:         1000,
			expectedSurcharge: 200, // 1 extra * 200 * 1 hour
			expectedTotal:     1200,
		},
		{
			name:              "no surcharge when surcharge is 0",
			guestCount:        10,
			baseCapacity:      3,
			surchargePerGuest: 0,
			durationHours:     3,
			basePrice:         1000,
			expectedSurcharge: 0,
			expectedTotal:     3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
			end := start.Add(time.Duration(tt.durationHours) * time.Hour)

			total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
				BathhouseID:                bathhouseID,
				BasePrice:                  tt.basePrice,
				StartTime:                  start,
				EndTime:                    end,
				GuestCount:                 tt.guestCount,
				BaseCapacity:               tt.baseCapacity,
				ExtraGuestSurcharge:        tt.surchargePerGuest,
				LongSessionThresholdHours:  4,
				LongSessionDiscountPercent: 0,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if breakdown.ExtraGuestSurcharge != tt.expectedSurcharge {
				t.Errorf("expected surcharge %d, got %d", tt.expectedSurcharge, breakdown.ExtraGuestSurcharge)
			}
			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestPricingService_CalculateFullPrice_Combined(t *testing.T) {
	// Test combined discount + surcharge
	priceRepo := mock.NewPricingRuleRepo()
	bhRepo := mock.NewBathhouseRepo()
	access := NewAccessChecker(mock.NewRepresentativeRepo(), bhRepo)

	svc := NewPricingService(priceRepo, bhRepo, access, testPricingLogger)

	bathhouseID := uuid.New()
	bh := &domain.Bathhouse{ID: bathhouseID, OwnerID: uuid.New()}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("failed to create bathhouse: %v", err)
	}

	// 6h booking, threshold=4, discount=10%, 5 guests, capacity=3, surcharge=500/guest/h
	start := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(6 * time.Hour)

	total, breakdown, err := svc.CalculateFullPrice(context.Background(), PriceCalculationInput{
		BathhouseID:                bathhouseID,
		BasePrice:                  1000,
		StartTime:                  start,
		EndTime:                    end,
		GuestCount:                 5,
		BaseCapacity:               3,
		ExtraGuestSurcharge:        500,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base: 6000 (1000 * 6h)
	// Discount: 200 (avgRate=1000, 2 discountable hours * 1000 * 10% = 200)
	// Surcharge: 6000 (2 extra * 500 * 6h)
	// Total: 6000 - 200 + 6000 = 11800
	if breakdown.BasePrice != 6000 {
		t.Errorf("expected base price 6000, got %d", breakdown.BasePrice)
	}
	if breakdown.LongSessionDiscount != 200 {
		t.Errorf("expected discount 200, got %d", breakdown.LongSessionDiscount)
	}
	if breakdown.ExtraGuestSurcharge != 6000 {
		t.Errorf("expected surcharge 6000, got %d", breakdown.ExtraGuestSurcharge)
	}
	if total != 11800 {
		t.Errorf("expected total 11800, got %d", total)
	}
}
