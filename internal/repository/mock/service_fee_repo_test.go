package mock

import (
	"context"
	"testing"

	"github.com/rekurt/relax-hub/internal/domain"
)

func TestServiceFeeRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewServiceFeeRepo()

	// Upsert global default
	cfg := &domain.ServiceFeeConfig{Region: "*", FeePercent: 10.0}
	if err := repo.Upsert(ctx, cfg); err != nil {
		t.Fatalf("Upsert global: %v", err)
	}

	// GetGlobalDefault
	got, err := repo.GetGlobalDefault(ctx)
	if err != nil {
		t.Fatalf("GetGlobalDefault: %v", err)
	}
	if got.FeePercent != 10.0 {
		t.Errorf("expected 10.0, got %v", got.FeePercent)
	}

	// GetByRegion (not found)
	_, err = repo.GetByRegion(ctx, "RU")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Upsert region
	ruCfg := &domain.ServiceFeeConfig{Region: "RU", FeePercent: 8.0}
	if err := repo.Upsert(ctx, ruCfg); err != nil {
		t.Fatalf("Upsert RU: %v", err)
	}

	got, err = repo.GetByRegion(ctx, "RU")
	if err != nil {
		t.Fatalf("GetByRegion RU: %v", err)
	}
	if got.FeePercent != 8.0 {
		t.Errorf("expected 8.0, got %v", got.FeePercent)
	}

	// Upsert region+category
	cat := "premium"
	catCfg := &domain.ServiceFeeConfig{Region: "RU", Category: &cat, FeePercent: 5.0}
	if err := repo.Upsert(ctx, catCfg); err != nil {
		t.Fatalf("Upsert RU+premium: %v", err)
	}

	got, err = repo.GetByRegionAndCategory(ctx, "RU", &cat)
	if err != nil {
		t.Fatalf("GetByRegionAndCategory: %v", err)
	}
	if got.FeePercent != 5.0 {
		t.Errorf("expected 5.0, got %v", got.FeePercent)
	}

	// List
	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 configs, got %d", len(all))
	}

	// Upsert existing (update)
	ruCfg2 := &domain.ServiceFeeConfig{Region: "RU", FeePercent: 12.0}
	if err := repo.Upsert(ctx, ruCfg2); err != nil {
		t.Fatalf("Upsert RU update: %v", err)
	}

	got, err = repo.GetByRegion(ctx, "RU")
	if err != nil {
		t.Fatalf("GetByRegion after update: %v", err)
	}
	if got.FeePercent != 12.0 {
		t.Errorf("expected 12.0, got %v", got.FeePercent)
	}

	// Still 3 configs (no duplicate)
	all, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 configs after update, got %d", len(all))
	}
}
