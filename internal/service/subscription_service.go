package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID, plan domain.SubscriptionPlan) (*domain.Subscription, error)
	Cancel(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error
	GetActive(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error)
	ListByOwner(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error)
}

type subscriptionService struct {
	subRepo   repository.SubscriptionRepository
	bhRepo    repository.BathhouseRepository
	access    *AccessChecker
	logger    *logger.Logger
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	bhRepo repository.BathhouseRepository,
	access *AccessChecker,
	log *logger.Logger,
) SubscriptionService {
	return &subscriptionService{
		subRepo: subRepo,
		bhRepo:  bhRepo,
		access:  access,
		logger:  log,
	}
}

// Subscribe creates a new subscription for a bathhouse
func (s *subscriptionService) Subscribe(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID, plan domain.SubscriptionPlan) (*domain.Subscription, error) {
	// Verify user has access to manage the bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, domain.RoleOwner, bathhouseID); err != nil {
		return nil, err
	}

	// Verify bathhouse exists
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	// Check if there's already an active subscription
	existing, err := s.subRepo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrSubscriptionAlreadyActive
	}

	now := time.Now()
	var endDate *time.Time
	var priceKopecks int64
	autoRenew := true

	// Set plan-specific parameters
	switch plan {
	case domain.PlanFree:
		endDate = nil
		priceKopecks = 0
		autoRenew = false
	case domain.PlanPremium:
		// Premium: 1 month, 5000 kopecks (50 rubles)
		end := now.AddDate(0, 1, 0)
		endDate = &end
		priceKopecks = 5000 // 50 rubles
	case domain.PlanPromoted:
		// Promoted: 1 month, 10000 kopecks (100 rubles)
		end := now.AddDate(0, 1, 0)
		endDate = &end
		priceKopecks = 10000 // 100 rubles
	default:
		return nil, domain.ErrInvalidInput
	}

	sub := &domain.Subscription{
		ID:           uuid.New(),
		BathhouseID:  bathhouseID,
		OwnerID:      bh.OwnerID,
		Plan:         plan,
		Status:       domain.SubscriptionActive,
		StartDate:    now,
		EndDate:      endDate,
		AutoRenew:    autoRenew,
		PriceKopecks: priceKopecks,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := sub.Validate(); err != nil {
		return nil, err
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("subscription created", "bathhouse_id", bathhouseID, "plan", plan)

	return sub, nil
}

// Cancel cancels auto-renewal for a subscription
func (s *subscriptionService) Cancel(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error {
	sub, err := s.subRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Verify user has access to manage the bathhouse (owns it)
	if err := s.access.CanManageBathhouse(ctx, userID, domain.RoleOwner, sub.BathhouseID); err != nil {
		return err
	}

	if sub.Status != domain.SubscriptionActive {
		return domain.ErrInvalidInput
	}

	sub.AutoRenew = false
	sub.Status = domain.SubscriptionCancelled
	sub.UpdatedAt = time.Now()

	if err := sub.Validate(); err != nil {
		return err
	}

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return err
	}

	s.logger.Info("subscription cancelled", "subscription_id", subscriptionID, "bathhouse_id", sub.BathhouseID)

	return nil
}

// GetActive returns the active subscription for a bathhouse
func (s *subscriptionService) GetActive(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error) {
	return s.subRepo.GetActiveBybathhouse(ctx, bathhouseID)
}

// ListByOwner returns all subscriptions for an owner
func (s *subscriptionService) ListByOwner(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
	return s.subRepo.ListByOwner(ctx, userID, page, pageSize)
}
