package service

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PriceCalculationInput struct {
	BathhouseID                uuid.UUID
	BasePrice                  int64
	StartTime                  time.Time
	EndTime                    time.Time
	GuestCount                 int
	BaseCapacity               int
	ExtraGuestSurcharge        int64
	LongSessionThresholdHours  int
	LongSessionDiscountPercent int
}

type PriceBreakdown struct {
	BasePrice               int64   // price after dynamic rules (before discount/surcharge)
	LongSessionDiscount     int64   // discount amount (positive value)
	ExtraGuestSurcharge     int64   // surcharge amount
	IsHolidayPrice          bool    // whether holiday pricing was applied
	HolidayName             string  // holiday name if applicable
	HolidayMultiplier       float64 // holiday multiplier used (0 if not holiday)
	IsSeasonalPrice         bool    // whether seasonal tariff was applied
	SeasonalTariffName      string  // seasonal tariff name if applicable
	SeasonalTariffMultiplier float64 // seasonal tariff multiplier used (0 if not seasonal)
}

type PricingService interface {
	CalculatePrice(ctx context.Context, bathhouseID uuid.UUID, basePrice int64, startTime, endTime time.Time) (int64, error)
	CalculateFullPrice(ctx context.Context, input PriceCalculationInput) (int64, *PriceBreakdown, error)
	CreateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error)
	UpdateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error
	DeleteRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, ruleID uuid.UUID) error
	ListRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	// Seasonal tariff CRUD
	CreateSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariff *domain.SeasonalTariff) (*domain.SeasonalTariff, error)
	UpdateSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariff *domain.SeasonalTariff) error
	DeleteSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariffID uuid.UUID) error
	ListSeasonalTariffs(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error)
}

type pricingService struct {
	priceRuleRepo      repository.PricingRuleRepository
	seasonalTariffRepo repository.SeasonalTariffRepository
	bhRepo             repository.BathhouseRepository
	holidaySvc         HolidayService
	access             *AccessChecker
	logger             *logger.Logger
}

func NewPricingService(
	priceRuleRepo repository.PricingRuleRepository,
	seasonalTariffRepo repository.SeasonalTariffRepository,
	bhRepo repository.BathhouseRepository,
	holidaySvc HolidayService,
	access *AccessChecker,
	log *logger.Logger,
) PricingService {
	return &pricingService{
		priceRuleRepo:      priceRuleRepo,
		seasonalTariffRepo: seasonalTariffRepo,
		bhRepo:             bhRepo,
		holidaySvc:         holidaySvc,
		access:             access,
		logger:             log,
	}
}

// CalculatePrice calculates the total price for a booking by applying pricing rules
// Algorithm: split interval into hourly slots, find highest-priority rule for each hour, apply multiplier
func (s *pricingService) CalculatePrice(ctx context.Context, bathhouseID uuid.UUID, basePrice int64, startTime, endTime time.Time) (int64, error) {
	if startTime.After(endTime) {
		return 0, domain.ErrInvalidInput
	}

	// Get all active rules for this bathhouse
	rules, err := s.priceRuleRepo.GetActiveRules(ctx, bathhouseID)
	if err != nil {
		return 0, err
	}

	totalPrice := int64(0)
	currentTime := startTime

	// Process each hour in the interval using consistent algorithm
	for currentTime.Before(endTime) {
		hourEnd := currentTime.Add(1 * time.Hour)
		if hourEnd.After(endTime) {
			hourEnd = endTime
		}

		// Find the best matching rule for this hour
		multiplier := 1.0
		applicableRules := s.findApplicableRules(currentTime, rules)
		if len(applicableRules) > 0 {
			// Pick the rule with highest priority, using ID as tiebreaker for deterministic ordering
			slices.SortFunc(applicableRules, func(a, b *domain.PricingRule) int {
				if a.Priority != b.Priority {
					return b.Priority - a.Priority
				}
				// Deterministic tiebreaker by rule ID
				return strings.Compare(a.ID.String(), b.ID.String())
			})
			multiplier = applicableRules[0].Multiplier
		}

		// Calculate price for this hour
		hourDuration := hourEnd.Sub(currentTime).Hours()
		hourPrice := int64(math.Round(float64(basePrice) * hourDuration * multiplier))
		totalPrice += hourPrice

		currentTime = hourEnd
	}

	return totalPrice, nil
}

// CalculateFullPrice calculates the total price including long session discount and extra guest surcharge
func (s *pricingService) CalculateFullPrice(ctx context.Context, input PriceCalculationInput) (int64, *PriceBreakdown, error) {
	// Check if the booking date is a holiday and apply multiplier to base price before hourly slot iteration
	effectiveBasePrice := input.BasePrice
	breakdown := &PriceBreakdown{}

	if s.holidaySvc != nil {
		holiday, multiplier, err := s.holidaySvc.IsHolidayDate(ctx, input.StartTime, "RU", input.BathhouseID)
		if err != nil {
			s.logger.Warn("failed to check holiday pricing, continuing without", "error", err)
		} else if holiday != nil && multiplier > 0 {
			effectiveBasePrice = int64(math.Round(float64(input.BasePrice) * multiplier))
			breakdown.IsHolidayPrice = true
			breakdown.HolidayName = holiday.Name
			breakdown.HolidayMultiplier = multiplier
		}
	}

	// Apply seasonal tariff multiplier (after holiday, before dynamic rules)
	if s.seasonalTariffRepo != nil {
		tariffs, err := s.seasonalTariffRepo.GetActiveTariffs(ctx, input.BathhouseID, input.StartTime)
		if err != nil {
			s.logger.Warn("failed to check seasonal tariffs, continuing without", "error", err)
		} else if len(tariffs) > 0 {
			// When multiple tariffs overlap, pick the highest multiplier
			bestTariff := tariffs[0]
			for _, t := range tariffs[1:] {
				if t.Multiplier > bestTariff.Multiplier {
					bestTariff = t
				}
			}
			effectiveBasePrice = int64(math.Round(float64(effectiveBasePrice) * bestTariff.Multiplier))
			breakdown.IsSeasonalPrice = true
			breakdown.SeasonalTariffName = bestTariff.Name
			breakdown.SeasonalTariffMultiplier = bestTariff.Multiplier
		}
	}

	// Calculate base price using dynamic pricing rules (with holiday-adjusted base price)
	basePrice, err := s.CalculatePrice(ctx, input.BathhouseID, effectiveBasePrice, input.StartTime, input.EndTime)
	if err != nil {
		return 0, nil, err
	}

	breakdown.BasePrice = basePrice

	durationHours := int(input.EndTime.Sub(input.StartTime) / time.Hour)

	// Long session discount: discount hours beyond the threshold
	if durationHours > 0 && durationHours >= input.LongSessionThresholdHours && input.LongSessionDiscountPercent > 0 {
		discountableHours := durationHours - input.LongSessionThresholdHours
		if discountableHours > 0 {
			// Calculate average hourly rate from the base price
			avgHourlyRate := basePrice / int64(durationHours)
			breakdown.LongSessionDiscount = avgHourlyRate * int64(discountableHours) * int64(input.LongSessionDiscountPercent) / 100
		}
	}

	// Extra guest surcharge
	if input.GuestCount > input.BaseCapacity && input.ExtraGuestSurcharge > 0 {
		extraGuests := input.GuestCount - input.BaseCapacity
		breakdown.ExtraGuestSurcharge = input.ExtraGuestSurcharge * int64(extraGuests) * int64(durationHours)
	}

	totalPrice := basePrice - breakdown.LongSessionDiscount + breakdown.ExtraGuestSurcharge
	if totalPrice <= 0 {
		totalPrice = 1
	}

	return totalPrice, breakdown, nil
}

// CreateRule creates a new pricing rule for a bathhouse
func (s *pricingService) CreateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error) {
	// Verify user has access to manage the bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, rule.BathhouseID); err != nil {
		return nil, err
	}

	// Verify bathhouse exists
	_, err := s.bhRepo.GetByID(ctx, rule.BathhouseID)
	if err != nil {
		return nil, err
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	// Create the rule
	if err := s.priceRuleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	s.logger.Info("pricing rule created", "rule_id", rule.ID, "bathhouse_id", rule.BathhouseID)
	return rule, nil
}

// UpdateRule updates an existing pricing rule
func (s *pricingService) UpdateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error {
	// Get the rule first
	existing, err := s.priceRuleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		return err
	}

	// Verify user has access to manage the bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return err
	}

	// Preserve bathhouse ID from the existing rule
	rule.BathhouseID = existing.BathhouseID

	// Validate rule
	if err := rule.Validate(); err != nil {
		return err
	}

	// Update the rule
	if err := s.priceRuleRepo.Update(ctx, rule); err != nil {
		return err
	}

	s.logger.Info("pricing rule updated", "rule_id", rule.ID, "bathhouse_id", rule.BathhouseID)
	return nil
}

// DeleteRule deletes a pricing rule
func (s *pricingService) DeleteRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, ruleID uuid.UUID) error {
	// Get the rule first
	existing, err := s.priceRuleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return err
	}

	// Verify user has access to manage the bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return err
	}

	// Delete the rule
	if err := s.priceRuleRepo.Delete(ctx, ruleID); err != nil {
		return err
	}

	s.logger.Info("pricing rule deleted", "rule_id", ruleID, "bathhouse_id", existing.BathhouseID)
	return nil
}

// ListRules lists all pricing rules for a bathhouse
func (s *pricingService) ListRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	return s.priceRuleRepo.ListByBathhouse(ctx, bathhouseID)
}

// GetActiveRules returns all active rules for a bathhouse
func (s *pricingService) GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	return s.priceRuleRepo.GetActiveRules(ctx, bathhouseID)
}

// findApplicableRules finds all rules that apply to the given time
func (s *pricingService) findApplicableRules(t time.Time, rules []domain.PricingRule) []*domain.PricingRule {
	var applicable []*domain.PricingRule

	for i := range rules {
		if s.ruleAppliesAt(t, &rules[i]) {
			applicable = append(applicable, &rules[i])
		}
	}

	return applicable
}

// ruleAppliesAt checks if a rule applies at the given time
func (s *pricingService) ruleAppliesAt(t time.Time, rule *domain.PricingRule) bool {
	if !rule.IsActive {
		return false
	}

	switch rule.Type {
	case domain.RuleTypeWeekday, domain.RuleTypeWeekend:
		// Check day of week using the app's convention (0=Monday, 6=Sunday)
		dayOfWeek := s.getDayOfWeek(t.Weekday())
		return slices.Contains(rule.DaysOfWeek, dayOfWeek)

	case domain.RuleTypeHoliday:
		// Check date range (comparing only dates, ignoring time)
		if rule.DateFrom != nil && rule.DateTo != nil {
			tDate := t.Truncate(24 * time.Hour)
			fromDate := rule.DateFrom.Truncate(24 * time.Hour)
			toDate := rule.DateTo.Truncate(24 * time.Hour)
			return !tDate.Before(fromDate) && !tDate.After(toDate)
		}
		return false

	case domain.RuleTypeTimeRange:
		// Check time of day (independent of date)
		if rule.TimeFrom != nil && rule.TimeTo != nil {
			timeStr := t.Format("15:04")
			fromStr := *rule.TimeFrom
			toStr := *rule.TimeTo

			// Handle wraparound times (e.g., 22:00-06:00 for overnight)
			if fromStr > toStr {
				// Wraparound: rule applies if timeStr >= from OR timeStr < to
				return timeStr >= fromStr || timeStr < toStr
			}
			// Non-wraparound: rule applies if from <= timeStr < to
			return timeStr >= fromStr && timeStr < toStr
		}
		return false

	case domain.RuleTypeSeason:
		// Check date range (comparing only dates, ignoring time)
		if rule.DateFrom != nil && rule.DateTo != nil {
			tDate := t.Truncate(24 * time.Hour)
			fromDate := rule.DateFrom.Truncate(24 * time.Hour)
			toDate := rule.DateTo.Truncate(24 * time.Hour)
			return !tDate.Before(fromDate) && !tDate.After(toDate)
		}
		return false

	default:
		return false
	}
}

// toDayOfWeek converts Go's time.Weekday to the app's convention (0=Monday, 6=Sunday).
// This matches the convention used in BookingService and WorkingHours.
func (s *pricingService) getDayOfWeek(wd time.Weekday) int {
	if wd == time.Sunday {
		return 6
	}
	return int(wd) - 1
}

// CreateSeasonalTariff creates a new seasonal tariff for a bathhouse.
func (s *pricingService) CreateSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariff *domain.SeasonalTariff) (*domain.SeasonalTariff, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, tariff.BathhouseID); err != nil {
		return nil, err
	}
	if _, err := s.bhRepo.GetByID(ctx, tariff.BathhouseID); err != nil {
		return nil, err
	}
	if err := tariff.Validate(); err != nil {
		return nil, err
	}
	// Check for overlapping tariffs
	existing, err := s.seasonalTariffRepo.ListByBathhouse(ctx, tariff.BathhouseID)
	if err != nil {
		return nil, fmt.Errorf("check tariff overlap: %w", err)
	}
	for _, e := range existing {
		if tariffDatesOverlap(tariff, &e) {
			return nil, domain.ErrSeasonalTariffOverlap
		}
	}
	if err := s.seasonalTariffRepo.Create(ctx, tariff); err != nil {
		return nil, err
	}
	s.logger.Info("seasonal tariff created", "tariff_id", tariff.ID, "bathhouse_id", tariff.BathhouseID)
	return tariff, nil
}

// truncateToDate returns midnight in the time's own location, unlike Truncate(24h) which is UTC-relative.
func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// tariffDatesOverlap checks if two seasonal tariff date ranges overlap.
func tariffDatesOverlap(a *domain.SeasonalTariff, b *domain.SeasonalTariff) bool {
	aFrom := truncateToDate(a.DateFrom)
	aTo := truncateToDate(a.DateTo)
	bFrom := truncateToDate(b.DateFrom)
	bTo := truncateToDate(b.DateTo)
	return !aFrom.After(bTo) && !bFrom.After(aTo)
}

// UpdateSeasonalTariff updates an existing seasonal tariff.
func (s *pricingService) UpdateSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariff *domain.SeasonalTariff) error {
	existing, err := s.seasonalTariffRepo.GetByID(ctx, tariff.ID)
	if err != nil {
		return err
	}
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return err
	}
	tariff.BathhouseID = existing.BathhouseID
	if err := tariff.Validate(); err != nil {
		return err
	}
	// Check for overlapping tariffs (exclude self)
	allTariffs, err := s.seasonalTariffRepo.ListByBathhouse(ctx, tariff.BathhouseID)
	if err != nil {
		return fmt.Errorf("check tariff overlap: %w", err)
	}
	for _, e := range allTariffs {
		if e.ID != tariff.ID && tariffDatesOverlap(tariff, &e) {
			return domain.ErrSeasonalTariffOverlap
		}
	}
	if err := s.seasonalTariffRepo.Update(ctx, tariff); err != nil {
		return err
	}
	s.logger.Info("seasonal tariff updated", "tariff_id", tariff.ID, "bathhouse_id", tariff.BathhouseID)
	return nil
}

// DeleteSeasonalTariff deletes a seasonal tariff.
func (s *pricingService) DeleteSeasonalTariff(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, tariffID uuid.UUID) error {
	existing, err := s.seasonalTariffRepo.GetByID(ctx, tariffID)
	if err != nil {
		return err
	}
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return err
	}
	if err := s.seasonalTariffRepo.Delete(ctx, tariffID); err != nil {
		return err
	}
	s.logger.Info("seasonal tariff deleted", "tariff_id", tariffID, "bathhouse_id", existing.BathhouseID)
	return nil
}

// ListSeasonalTariffs lists all seasonal tariffs for a bathhouse.
func (s *pricingService) ListSeasonalTariffs(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error) {
	return s.seasonalTariffRepo.ListByBathhouse(ctx, bathhouseID)
}
