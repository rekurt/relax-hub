package service

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PricingService interface {
	CalculatePrice(ctx context.Context, bathhouseID uuid.UUID, basePrice int64, startTime, endTime time.Time) (int64, error)
	CreateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) (*domain.PricingRule, error)
	UpdateRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, rule *domain.PricingRule) error
	DeleteRule(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, ruleID uuid.UUID) error
	ListRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
}

type pricingService struct {
	priceRuleRepo repository.PricingRuleRepository
	bhRepo        repository.BathhouseRepository
	access        *AccessChecker
	logger        *logger.Logger
}

func NewPricingService(
	priceRuleRepo repository.PricingRuleRepository,
	bhRepo repository.BathhouseRepository,
	access *AccessChecker,
	log *logger.Logger,
) PricingService {
	return &pricingService{
		priceRuleRepo: priceRuleRepo,
		bhRepo:        bhRepo,
		access:        access,
		logger:        log,
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

	// If no rules, return base price * hours
	if len(rules) == 0 {
		hours := int64(math.Ceil(endTime.Sub(startTime).Hours()))
		return basePrice * hours, nil
	}

	totalPrice := int64(0)
	currentTime := startTime

	// Process each hour in the interval
	for currentTime.Before(endTime) {
		hourEnd := currentTime.Add(1 * time.Hour)
		if hourEnd.After(endTime) {
			hourEnd = endTime
		}

		// Find the best matching rule for this hour
		multiplier := 1.0
		applicableRules := s.findApplicableRules(currentTime, rules)
		if len(applicableRules) > 0 {
			// Pick the rule with highest priority
			slices.SortFunc(applicableRules, func(a, b *domain.PricingRule) int {
				if a.Priority != b.Priority {
					return b.Priority - a.Priority
				}
				return 0
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
