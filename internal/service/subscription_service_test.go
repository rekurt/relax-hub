package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newSubscriptionService() (service.SubscriptionService, *mock.BathhouseRepo, *mock.SubscriptionRepo) {
	bhRepo := mock.NewBathhouseRepo()
	subRepo := mock.NewSubscriptionRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	svc := service.NewSubscriptionService(subRepo, bhRepo, access, log)
	return svc, bhRepo, subRepo
}

func TestSubscriptionService_Subscribe_Success(t *testing.T) {
	svc, bhRepo, subRepo := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sub.Plan != domain.PlanPremium {
		t.Errorf("plan = %q, want %q", sub.Plan, domain.PlanPremium)
	}
	if sub.Status != domain.SubscriptionActive {
		t.Errorf("status = %q, want %q", sub.Status, domain.SubscriptionActive)
	}
	if sub.BathhouseID != bh.ID {
		t.Errorf("bathhouse_id mismatch")
	}
	if sub.OwnerID != ownerID {
		t.Errorf("owner_id mismatch")
	}

	// Verify it was saved in repo
	saved, err := subRepo.GetByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("failed to get subscription from repo: %v", err)
	}
	if saved.Plan != domain.PlanPremium {
		t.Errorf("saved subscription plan = %q, want %q", saved.Plan, domain.PlanPremium)
	}
}

func TestSubscriptionService_Subscribe_FreeVersion(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanFree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sub.Plan != domain.PlanFree {
		t.Errorf("plan = %q, want %q", sub.Plan, domain.PlanFree)
	}
	if sub.AutoRenew {
		t.Errorf("free subscription should not auto-renew")
	}
	if sub.PriceKopecks != 0 {
		t.Errorf("free subscription price = %d, want 0", sub.PriceKopecks)
	}
	if sub.EndDate != nil {
		t.Errorf("free subscription should have nil EndDate")
	}
}

func TestSubscriptionService_Subscribe_AlreadyActive(t *testing.T) {
	svc, bhRepo, subRepo := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create first subscription
	_, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("first subscription failed: %v", err)
	}

	// Try to create second subscription
	_, err = svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPromoted)
	if !errors.Is(err, domain.ErrSubscriptionAlreadyActive) {
		t.Errorf("expected ErrSubscriptionAlreadyActive, got: %v", err)
	}

	// Verify only one subscription exists
	result, err := subRepo.ListByOwner(context.Background(), ownerID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list subscriptions: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("total count = %d, want 1", result.TotalCount)
	}
}

func TestSubscriptionService_Subscribe_NotOwner(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Different user tries to subscribe
	_, err := svc.Subscribe(context.Background(), otherUserID, bh.ID, domain.PlanPremium)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestSubscriptionService_Subscribe_BathhouseNotFound(t *testing.T) {
	svc, _, _ := newSubscriptionService()
	ownerID := uuid.New()

	_, err := svc.Subscribe(context.Background(), ownerID, uuid.New(), domain.PlanPremium)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSubscriptionService_Cancel_Success(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("subscription creation failed: %v", err)
	}

	err = svc.Cancel(context.Background(), ownerID, sub.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptionService_Cancel_NotOwner(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("subscription creation failed: %v", err)
	}

	// Different user tries to cancel
	err = svc.Cancel(context.Background(), otherUserID, sub.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestSubscriptionService_GetActive_Success(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("subscription creation failed: %v", err)
	}

	active, err := svc.GetActive(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if active.ID != sub.ID {
		t.Errorf("subscription id mismatch")
	}
	if active.Plan != domain.PlanPremium {
		t.Errorf("plan = %q, want %q", active.Plan, domain.PlanPremium)
	}
}

func TestSubscriptionService_GetActive_NotFound(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.GetActive(context.Background(), bh.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSubscriptionService_ListByOwner_Success(t *testing.T) {
	svc, bhRepo, _ := newSubscriptionService()
	ownerID := uuid.New()

	bh1 := createBathhouse(t, bhRepo, ownerID)
	bh2 := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.Subscribe(context.Background(), ownerID, bh1.ID, domain.PlanPremium)
	if err != nil {
		t.Fatalf("first subscription failed: %v", err)
	}

	_, err = svc.Subscribe(context.Background(), ownerID, bh2.ID, domain.PlanPromoted)
	if err != nil {
		t.Fatalf("second subscription failed: %v", err)
	}

	// Now cancel the first subscription's auto-renewal to test listing both
	result, err := svc.ListByOwner(context.Background(), ownerID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("total count = %d, want 2", result.TotalCount)
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
}

func TestSubscriptionService_ListByOwner_Empty(t *testing.T) {
	svc, _, _ := newSubscriptionService()
	ownerID := uuid.New()

	result, err := svc.ListByOwner(context.Background(), ownerID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 0 {
		t.Errorf("total count = %d, want 0", result.TotalCount)
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestSubscriptionService_PricingByPlan(t *testing.T) {
	tests := []struct {
		plan          domain.SubscriptionPlan
		expectedPrice int64
		expectedRenew bool
	}{
		{domain.PlanFree, 0, false},
		{domain.PlanPremium, 500000, true},
		{domain.PlanPromoted, 1000000, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			svc, bhRepo, _ := newSubscriptionService()
			ownerID := uuid.New()
			bh := createBathhouse(t, bhRepo, ownerID)

			sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, tt.plan)
			if err != nil {
				t.Fatalf("subscription creation failed: %v", err)
			}

			if sub.PriceKopecks != tt.expectedPrice {
				t.Errorf("price = %d, want %d", sub.PriceKopecks, tt.expectedPrice)
			}
			if sub.AutoRenew != tt.expectedRenew {
				t.Errorf("auto_renew = %v, want %v", sub.AutoRenew, tt.expectedRenew)
			}
		})
	}
}

func TestSubscriptionService_EndDateForPaidPlans(t *testing.T) {
	tests := []struct {
		plan         domain.SubscriptionPlan
		hasEndDate   bool
	}{
		{domain.PlanFree, false},
		{domain.PlanPremium, true},
		{domain.PlanPromoted, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			svc, bhRepo, _ := newSubscriptionService()
			ownerID := uuid.New()
			bh := createBathhouse(t, bhRepo, ownerID)

			sub, err := svc.Subscribe(context.Background(), ownerID, bh.ID, tt.plan)
			if err != nil {
				t.Fatalf("subscription creation failed: %v", err)
			}

			hasEndDate := sub.EndDate != nil
			if hasEndDate != tt.hasEndDate {
				t.Errorf("has_end_date = %v, want %v", hasEndDate, tt.hasEndDate)
			}

			if tt.hasEndDate && sub.EndDate != nil {
				// End date should be approximately 1 month from now
				duration := sub.EndDate.Sub(sub.StartDate)
				if duration < 28*24*time.Hour || duration > 32*24*time.Hour {
					t.Errorf("end date duration = %v, want ~1 month", duration)
				}
			}
		})
	}
}
