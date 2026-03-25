package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AutoScenarioRepo struct {
	mu         sync.RWMutex
	scenarios  map[uuid.UUID]*domain.AutoScenario
	executions map[string]bool // key: scenarioID:guestCardID
}

func NewAutoScenarioRepo() repository.AutoScenarioRepository {
	return &AutoScenarioRepo{
		scenarios:  make(map[uuid.UUID]*domain.AutoScenario),
		executions: make(map[string]bool),
	}
}

func execKey(scenarioID, guestCardID uuid.UUID) string {
	return scenarioID.String() + ":" + guestCardID.String()
}

func (r *AutoScenarioRepo) Upsert(_ context.Context, scenario *domain.AutoScenario) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check if exists by owner+type
	for _, s := range r.scenarios {
		if s.OwnerID == scenario.OwnerID && s.Type == scenario.Type {
			s.Enabled = scenario.Enabled
			s.CustomText = scenario.CustomText
			s.Channel = scenario.Channel
			s.DelayHours = scenario.DelayHours
			s.PromoCodeID = scenario.PromoCodeID
			s.UpdatedAt = now
			scenario.ID = s.ID
			scenario.CreatedAt = s.CreatedAt
			scenario.UpdatedAt = s.UpdatedAt
			return nil
		}
	}

	if scenario.ID == uuid.Nil {
		scenario.ID = uuid.New()
	}
	if scenario.CreatedAt.IsZero() {
		scenario.CreatedAt = now
	}
	scenario.UpdatedAt = now

	cp := *scenario
	r.scenarios[scenario.ID] = &cp
	return nil
}

func (r *AutoScenarioRepo) ListByOwner(_ context.Context, ownerID uuid.UUID) ([]domain.AutoScenario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AutoScenario
	for _, s := range r.scenarios {
		if s.OwnerID == ownerID {
			cp := *s
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *AutoScenarioRepo) GetByOwnerAndType(_ context.Context, ownerID uuid.UUID, scenarioType domain.AutoScenarioType) (*domain.AutoScenario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.scenarios {
		if s.OwnerID == ownerID && s.Type == scenarioType {
			cp := *s
			return &cp, nil
		}
	}
	return nil, domain.ErrAutoScenarioNotFound
}

func (r *AutoScenarioRepo) ListEnabled(_ context.Context) ([]domain.AutoScenario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AutoScenario
	for _, s := range r.scenarios {
		if s.Enabled {
			cp := *s
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *AutoScenarioRepo) RecordExecution(_ context.Context, scenarioID, guestCardID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.executions[execKey(scenarioID, guestCardID)] = true
	return nil
}

func (r *AutoScenarioRepo) HasBeenExecuted(_ context.Context, scenarioID, guestCardID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.executions[execKey(scenarioID, guestCardID)], nil
}
