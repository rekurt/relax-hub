package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func newTestPromo(code string, creatorID uuid.UUID, bathhouseID *uuid.UUID) *domain.PromoCode {
	return &domain.PromoCode{
		Code:        code,
		Type:        domain.PromoTypePercentage,
		Value:       10,
		BathhouseID: bathhouseID,
		CreatorID:   creatorID,
		MaxUses:     100,
		MinAmount:   100000,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
		IsActive:    true,
	}
}

func TestPromoCodeRepo_Create(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := newTestPromo("SUMMER20", uuid.New(), nil)

	err := repo.Create(context.Background(), promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promo.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestPromoCodeRepo_Create_DuplicateCode(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo1 := newTestPromo("SUMMER20", uuid.New(), nil)
	_ = repo.Create(context.Background(), promo1)

	promo2 := newTestPromo("SUMMER20", uuid.New(), nil)
	err := repo.Create(context.Background(), promo2)
	if err != domain.ErrAlreadyExists {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestPromoCodeRepo_GetByID(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := newTestPromo("WINTER30", uuid.New(), nil)
	_ = repo.Create(context.Background(), promo)

	found, err := repo.GetByID(context.Background(), promo.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Code != "WINTER30" {
		t.Errorf("code = %v, want WINTER30", found.Code)
	}
}

func TestPromoCodeRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrPromoNotFound {
		t.Errorf("err = %v, want ErrPromoNotFound", err)
	}
}

func TestPromoCodeRepo_GetByCode(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := newTestPromo("SPRING15", uuid.New(), nil)
	_ = repo.Create(context.Background(), promo)

	found, err := repo.GetByCode(context.Background(), "SPRING15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != promo.ID {
		t.Errorf("id = %v, want %v", found.ID, promo.ID)
	}
}

func TestPromoCodeRepo_GetByCode_NotFound(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	_, err := repo.GetByCode(context.Background(), "NONEXISTENT")
	if err != domain.ErrPromoNotFound {
		t.Errorf("err = %v, want ErrPromoNotFound", err)
	}
}

func TestPromoCodeRepo_Update(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := newTestPromo("UPDATE10", uuid.New(), nil)
	_ = repo.Create(context.Background(), promo)

	promo.Value = 20
	promo.IsActive = false
	err := repo.Update(context.Background(), promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), promo.ID)
	if found.Value != 20 {
		t.Errorf("value = %d, want 20", found.Value)
	}
	if found.IsActive {
		t.Error("expected IsActive=false")
	}
}

func TestPromoCodeRepo_Update_NotFound(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := &domain.PromoCode{ID: uuid.New(), Code: "NOPE"}
	err := repo.Update(context.Background(), promo)
	if err != domain.ErrPromoNotFound {
		t.Errorf("err = %v, want ErrPromoNotFound", err)
	}
}

func TestPromoCodeRepo_ListByBathhouse(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	bhID := uuid.New()
	creatorID := uuid.New()

	p1 := newTestPromo("BH1", creatorID, &bhID)
	p2 := newTestPromo("BH2", creatorID, &bhID)
	p3 := newTestPromo("GLOBAL", creatorID, nil)
	_ = repo.Create(context.Background(), p1)
	_ = repo.Create(context.Background(), p2)
	_ = repo.Create(context.Background(), p3)

	result, err := repo.ListByBathhouse(context.Background(), bhID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestPromoCodeRepo_IncrementUses(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	promo := newTestPromo("INC1", uuid.New(), nil)
	_ = repo.Create(context.Background(), promo)

	err := repo.IncrementUses(context.Background(), promo.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), promo.ID)
	if found.CurrentUses != 1 {
		t.Errorf("current_uses = %d, want 1", found.CurrentUses)
	}
}

func TestPromoCodeRepo_IncrementUses_NotFound(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	err := repo.IncrementUses(context.Background(), uuid.New())
	if err != domain.ErrPromoNotFound {
		t.Errorf("err = %v, want ErrPromoNotFound", err)
	}
}

func TestPromoCodeRepo_RecordUsage(t *testing.T) {
	repo := mock.NewPromoCodeRepo()
	usage := &domain.PromoUsage{
		PromoCodeID:    uuid.New(),
		UserID:         uuid.New(),
		BookingID:      uuid.New(),
		DiscountAmount: 50000,
	}

	err := repo.RecordUsage(context.Background(), usage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if usage.UsedAt.IsZero() {
		t.Error("expected UsedAt to be set")
	}
}
