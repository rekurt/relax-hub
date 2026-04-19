package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// PricingRuleRepo is an in-memory mock implementation of repository.PricingRuleRepository.
type PricingRuleRepo struct {
	mu    sync.RWMutex
	rules map[uuid.UUID]*domain.PricingRule
}

func NewPricingRuleRepo() *PricingRuleRepo {
	return &PricingRuleRepo{rules: make(map[uuid.UUID]*domain.PricingRule)}
}

func (r *PricingRuleRepo) Create(_ context.Context, rule *domain.PricingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	rule.CreatedAt = time.Now()
	cp := *rule
	r.rules[rule.ID] = &cp
	return nil
}

func (r *PricingRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rule
	return &cp, nil
}

func (r *PricingRuleRepo) Update(_ context.Context, rule *domain.PricingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rules[rule.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *rule
	r.rules[rule.ID] = &cp
	return nil
}

func (r *PricingRuleRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rules[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.rules, id)
	return nil
}

func (r *PricingRuleRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PricingRule
	for _, rule := range r.rules {
		if rule.BathhouseID == bathhouseID {
			cp := *rule
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *PricingRuleRepo) GetActiveRules(_ context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PricingRule
	for _, rule := range r.rules {
		if rule.BathhouseID == bathhouseID && rule.IsActive {
			cp := *rule
			result = append(result, cp)
		}
	}
	return result, nil
}


