package service_test

import (
	"context"
	"testing"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func setupServiceFee(t *testing.T) (service.ServiceFeeService, *mock.ServiceFeeRepo) {
	t.Helper()
	repo := mock.NewServiceFeeRepo().(*mock.ServiceFeeRepo)
	svc := service.NewServiceFeeService(repo)
	return svc, repo
}

func seedGlobalDefault(t *testing.T, repo *mock.ServiceFeeRepo, pct float64) {
	t.Helper()
	ctx := context.Background()
	cfg := &domain.ServiceFeeConfig{Region: "*", FeePercent: pct}
	if err := repo.Upsert(ctx, cfg); err != nil {
		t.Fatalf("seed global default: %v", err)
	}
}

func TestServiceFeeService_GetFeePercent_GlobalDefault(t *testing.T) {
	svc, repo := setupServiceFee(t)
	seedGlobalDefault(t, repo, 10.0)

	pct, err := svc.GetFeePercent(context.Background(), "RU", nil)
	if err != nil {
		t.Fatalf("GetFeePercent: %v", err)
	}
	if pct != 10.0 {
		t.Errorf("expected 10.0, got %v", pct)
	}
}

func TestServiceFeeService_GetFeePercent_RegionOverride(t *testing.T) {
	svc, repo := setupServiceFee(t)
	ctx := context.Background()
	seedGlobalDefault(t, repo, 10.0)

	// Add region-specific config
	ruCfg := &domain.ServiceFeeConfig{Region: "RU", FeePercent: 8.0}
	if err := repo.Upsert(ctx, ruCfg); err != nil {
		t.Fatalf("seed RU config: %v", err)
	}

	pct, err := svc.GetFeePercent(ctx, "RU", nil)
	if err != nil {
		t.Fatalf("GetFeePercent: %v", err)
	}
	if pct != 8.0 {
		t.Errorf("expected 8.0, got %v", pct)
	}

	// BY should fall back to global
	pct, err = svc.GetFeePercent(ctx, "BY", nil)
	if err != nil {
		t.Fatalf("GetFeePercent BY: %v", err)
	}
	if pct != 10.0 {
		t.Errorf("expected 10.0 for BY, got %v", pct)
	}
}

func TestServiceFeeService_GetFeePercent_RegionAndCategory(t *testing.T) {
	svc, repo := setupServiceFee(t)
	ctx := context.Background()
	seedGlobalDefault(t, repo, 10.0)

	ruCfg := &domain.ServiceFeeConfig{Region: "RU", FeePercent: 8.0}
	if err := repo.Upsert(ctx, ruCfg); err != nil {
		t.Fatalf("seed RU: %v", err)
	}

	cat := "premium"
	ruPremium := &domain.ServiceFeeConfig{Region: "RU", Category: &cat, FeePercent: 5.0}
	if err := repo.Upsert(ctx, ruPremium); err != nil {
		t.Fatalf("seed RU+premium: %v", err)
	}

	// RU + premium -> 5%
	pct, err := svc.GetFeePercent(ctx, "RU", &cat)
	if err != nil {
		t.Fatalf("GetFeePercent RU+premium: %v", err)
	}
	if pct != 5.0 {
		t.Errorf("expected 5.0, got %v", pct)
	}

	// RU + unknown category -> falls back to RU (8%)
	other := "basic"
	pct, err = svc.GetFeePercent(ctx, "RU", &other)
	if err != nil {
		t.Fatalf("GetFeePercent RU+basic: %v", err)
	}
	if pct != 8.0 {
		t.Errorf("expected 8.0, got %v", pct)
	}
}

func TestServiceFeeService_CalculateFee(t *testing.T) {
	svc, repo := setupServiceFee(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		pct       float64
		basePrice int64
		expected  int64
	}{
		{"10% of 100000", 10.0, 100000, 10000},
		{"0%", 0.0, 100000, 0},
		{"25% of 100000", 25.0, 100000, 25000},
		{"10% of 1 (rounding)", 10.0, 1, 0},   // 0.1 -> rounds to 0
		{"10% of 15 (rounding)", 10.0, 15, 2},  // 1.5 -> rounds to 2
		{"10% of 999", 10.0, 999, 100},          // 99.9 -> rounds to 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset repo with the desired percentage
			seedGlobalDefault(t, repo, tt.pct)

			fee, err := svc.CalculateFee(ctx, tt.basePrice, "*", nil)
			if err != nil {
				t.Fatalf("CalculateFee: %v", err)
			}
			if fee != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, fee)
			}
		})
	}
}

func TestServiceFeeService_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.ServiceFeeConfig
		wantErr bool
	}{
		{"valid", domain.ServiceFeeConfig{Region: "*", FeePercent: 10.0}, false},
		{"empty region", domain.ServiceFeeConfig{Region: "", FeePercent: 10.0}, true},
		{"negative percent", domain.ServiceFeeConfig{Region: "*", FeePercent: -1.0}, true},
		{"too high percent", domain.ServiceFeeConfig{Region: "*", FeePercent: 26.0}, true},
		{"max percent", domain.ServiceFeeConfig{Region: "*", FeePercent: 25.0}, false},
		{"zero percent", domain.ServiceFeeConfig{Region: "*", FeePercent: 0.0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
