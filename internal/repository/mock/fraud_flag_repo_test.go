package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestFraudFlagRepo_CreateAndList(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()
	userID := uuid.New()

	flag := &domain.FraudFlag{
		UserID:   userID,
		Rule:     domain.FraudRuleMultiCardTopUp,
		Severity: domain.FraudSeverityHigh,
		Action:   domain.FraudActionFreezeWallet,
		Details:  []byte(`{"card_count":5}`),
	}

	err := repo.Create(ctx, flag)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if flag.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	result, err := repo.ListByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleMultiCardTopUp {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleMultiCardTopUp, result.Items[0].Rule)
	}
}

func TestFraudFlagRepo_ListPending(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()
	userID := uuid.New()

	// Create pending flag
	err := repo.Create(ctx, &domain.FraudFlag{
		UserID:   userID,
		Rule:     domain.FraudRuleRapidBookings,
		Severity: domain.FraudSeverityMedium,
		Status:   domain.FraudFlagStatusPending,
		Action:   domain.FraudActionBlock,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Create reviewed flag
	err = repo.Create(ctx, &domain.FraudFlag{
		UserID:   userID,
		Rule:     domain.FraudRuleSelfBooking,
		Severity: domain.FraudSeverityHigh,
		Status:   domain.FraudFlagStatusReviewed,
		Action:   domain.FraudActionBlock,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	filter := domain.FraudFlagFilter{Page: 1, PageSize: 10}
	result, err := repo.ListPending(ctx, filter)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 pending flag, got %d", len(result.Items))
	}
	if result.Items[0].Rule != domain.FraudRuleRapidBookings {
		t.Errorf("expected rule %s, got %s", domain.FraudRuleRapidBookings, result.Items[0].Rule)
	}
}

func TestFraudFlagRepo_UpdateStatus(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()
	userID := uuid.New()
	adminID := uuid.New()

	flag := &domain.FraudFlag{
		UserID:   userID,
		Rule:     domain.FraudRuleStructuring,
		Severity: domain.FraudSeverityHigh,
		Action:   domain.FraudActionFreezeWallet,
	}
	_ = repo.Create(ctx, flag)

	err := repo.UpdateStatus(ctx, flag.ID, domain.FraudFlagStatusDismissed, adminID)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	result, _ := repo.ListByUser(ctx, userID, 1, 10)
	if result.Items[0].Status != domain.FraudFlagStatusDismissed {
		t.Errorf("expected status dismissed, got %s", result.Items[0].Status)
	}
	if result.Items[0].ReviewedBy == nil || *result.Items[0].ReviewedBy != adminID {
		t.Error("expected ReviewedBy to be set")
	}
}

func TestFraudFlagRepo_UpdateStatus_NotFound(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, uuid.New(), domain.FraudFlagStatusDismissed, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFraudFlagRepo_CountByUserAndRule(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()
	userID := uuid.New()

	_ = repo.Create(ctx, &domain.FraudFlag{
		UserID: userID,
		Rule:   domain.FraudRuleMultiCardTopUp,
		Action: domain.FraudActionFlag,
	})
	_ = repo.Create(ctx, &domain.FraudFlag{
		UserID: userID,
		Rule:   domain.FraudRuleMultiCardTopUp,
		Action: domain.FraudActionFlag,
	})
	_ = repo.Create(ctx, &domain.FraudFlag{
		UserID: userID,
		Rule:   domain.FraudRuleRapidBookings,
		Action: domain.FraudActionBlock,
	})

	count, err := repo.CountByUserAndRule(ctx, userID, domain.FraudRuleMultiCardTopUp, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("CountByUserAndRule: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestFraudFlagRepo_ListPending_WithFilters(t *testing.T) {
	repo := mock.NewFraudFlagRepo()
	ctx := context.Background()
	user1 := uuid.New()
	user2 := uuid.New()

	_ = repo.Create(ctx, &domain.FraudFlag{UserID: user1, Rule: domain.FraudRuleMultiCardTopUp, Status: domain.FraudFlagStatusPending, Action: domain.FraudActionFlag})
	_ = repo.Create(ctx, &domain.FraudFlag{UserID: user2, Rule: domain.FraudRuleRapidBookings, Status: domain.FraudFlagStatusPending, Action: domain.FraudActionBlock})

	// Filter by user
	filter := domain.FraudFlagFilter{UserID: &user1, Page: 1, PageSize: 10}
	result, err := repo.ListPending(ctx, filter)
	if err != nil {
		t.Fatalf("ListPending with user filter: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 flag for user1, got %d", len(result.Items))
	}

	// Filter by rule
	rule := domain.FraudRuleRapidBookings
	filter2 := domain.FraudFlagFilter{Rule: &rule, Page: 1, PageSize: 10}
	result2, err := repo.ListPending(ctx, filter2)
	if err != nil {
		t.Fatalf("ListPending with rule filter: %v", err)
	}
	if len(result2.Items) != 1 {
		t.Fatalf("expected 1 flag for rule, got %d", len(result2.Items))
	}
}
