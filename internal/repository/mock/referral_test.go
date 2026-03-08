package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestReferralRepo_Create(t *testing.T) {
	repo := mock.NewReferralRepo()

	referral := &domain.Referral{
		ReferrerID:   uuid.New(),
		RefereeID:    uuid.New(),
		ReferralCode: "abc123",
		Status:       domain.ReferralStatusPending,
		BonusAmount:  50000,
	}

	err := repo.Create(context.Background(), referral)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if referral.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestReferralRepo_Create_DuplicateReferee(t *testing.T) {
	repo := mock.NewReferralRepo()
	refereeID := uuid.New()

	_ = repo.Create(context.Background(), &domain.Referral{
		ReferrerID:   uuid.New(),
		RefereeID:    refereeID,
		ReferralCode: "code1",
		BonusAmount:  50000,
	})

	err := repo.Create(context.Background(), &domain.Referral{
		ReferrerID:   uuid.New(),
		RefereeID:    refereeID,
		ReferralCode: "code2",
		BonusAmount:  50000,
	})
	if err != domain.ErrAlreadyReferred {
		t.Errorf("err = %v, want ErrAlreadyReferred", err)
	}
}

func TestReferralRepo_GetByReferee(t *testing.T) {
	repo := mock.NewReferralRepo()
	refereeID := uuid.New()

	_ = repo.Create(context.Background(), &domain.Referral{
		ReferrerID:   uuid.New(),
		RefereeID:    refereeID,
		ReferralCode: "abc123",
		BonusAmount:  50000,
	})

	ref, err := repo.GetByReferee(context.Background(), refereeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.RefereeID != refereeID {
		t.Errorf("referee_id = %v, want %v", ref.RefereeID, refereeID)
	}
}

func TestReferralRepo_GetByReferee_NotFound(t *testing.T) {
	repo := mock.NewReferralRepo()

	_, err := repo.GetByReferee(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestReferralRepo_UpdateStatus(t *testing.T) {
	repo := mock.NewReferralRepo()
	refereeID := uuid.New()

	_ = repo.Create(context.Background(), &domain.Referral{
		ReferrerID:   uuid.New(),
		RefereeID:    refereeID,
		ReferralCode: "code1",
		BonusAmount:  50000,
	})

	ref, _ := repo.GetByReferee(context.Background(), refereeID)
	now := time.Now()
	err := repo.UpdateStatus(context.Background(), ref.ID, domain.ReferralStatusCompleted, &now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := repo.GetByReferee(context.Background(), refereeID)
	if updated.Status != domain.ReferralStatusCompleted {
		t.Errorf("status = %v, want completed", updated.Status)
	}
}

func TestReferralRepo_Balance(t *testing.T) {
	repo := mock.NewReferralRepo()
	userID := uuid.New()

	// Create balance
	err := repo.CreateBalance(context.Background(), &domain.ReferralBalance{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("create balance: %v", err)
	}

	// Add
	err = repo.UpdateBalance(context.Background(), userID, 50000)
	if err != nil {
		t.Fatalf("add balance: %v", err)
	}

	b, _ := repo.GetBalance(context.Background(), userID)
	if b.Balance != 50000 {
		t.Errorf("balance = %d, want 50000", b.Balance)
	}
	if b.TotalEarned != 50000 {
		t.Errorf("total_earned = %d, want 50000", b.TotalEarned)
	}

	// Spend
	err = repo.UpdateBalance(context.Background(), userID, -20000)
	if err != nil {
		t.Fatalf("spend balance: %v", err)
	}

	b, _ = repo.GetBalance(context.Background(), userID)
	if b.Balance != 30000 {
		t.Errorf("balance = %d, want 30000", b.Balance)
	}

	// Insufficient
	err = repo.UpdateBalance(context.Background(), userID, -50000)
	if err != domain.ErrInsufficientReferralBalance {
		t.Errorf("err = %v, want ErrInsufficientReferralBalance", err)
	}
}

func TestReferralRepo_CountByReferrer(t *testing.T) {
	repo := mock.NewReferralRepo()
	referrerID := uuid.New()

	_ = repo.Create(context.Background(), &domain.Referral{
		ReferrerID:   referrerID,
		RefereeID:    uuid.New(),
		ReferralCode: "code1",
		Status:       domain.ReferralStatusPending,
	})
	ref2 := &domain.Referral{
		ReferrerID:   referrerID,
		RefereeID:    uuid.New(),
		ReferralCode: "code1",
		Status:       domain.ReferralStatusPending,
	}
	_ = repo.Create(context.Background(), ref2)
	now := time.Now()
	_ = repo.UpdateStatus(context.Background(), ref2.ID, domain.ReferralStatusCompleted, &now)

	invited, completed, err := repo.CountByReferrer(context.Background(), referrerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invited != 2 {
		t.Errorf("invited = %d, want 2", invited)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1", completed)
	}
}
