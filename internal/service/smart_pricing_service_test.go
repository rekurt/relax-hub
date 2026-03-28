package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

var testSmartPricingLogger = logger.New(logger.LevelError)

func newSmartPricingService() (SmartPricingService, *mock.BathhouseRepo, *mock.AnalyticsRepo) {
	bhRepo := mock.NewBathhouseRepo()
	analyticsRepo := mock.NewAnalyticsRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := NewAccessChecker(repRepo, bhRepo)
	svc := NewSmartPricingService(bhRepo, analyticsRepo, access, testSmartPricingLogger)
	return svc, bhRepo, analyticsRepo
}

func TestSmartPricing_HighOccupancy_GrowingDemand(t *testing.T) {
	svc, bhRepo, analyticsRepo := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	cityID := int64(1)

	bh := &domain.Bathhouse{
		ID:            bathhouseID,
		OwnerID:       ownerID,
		Name:          "Test Banya",
		Address:       "Test St",
		CityID:        cityID,
		PricePerHour:  200000, // 2000 RUB
		MaxGuests:     10,
		MinDuration:   1,
		OccupancyRate: 0.85,
		Status:        domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	// Add a competitor in the same city
	competitor := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      uuid.New(),
		Name:         "Competitor",
		Address:      "Comp St",
		CityID:       cityID,
		PricePerHour: 180000,
		MaxGuests:    8,
		MinDuration:  1,
		Status:       domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, competitor)

	now := time.Now()

	// Seed growing demand: more bookings in recent week
	for i := 0; i < 7; i++ {
		_ = analyticsRepo.CreateSnapshot(ctx, &domain.AnalyticsSnapshot{
			BathhouseID: bathhouseID,
			Date:        now.AddDate(0, 0, -14+i),
			Bookings:    2,
		})
	}
	for i := 0; i < 7; i++ {
		_ = analyticsRepo.CreateSnapshot(ctx, &domain.AnalyticsSnapshot{
			BathhouseID: bathhouseID,
			Date:        now.AddDate(0, 0, -7+i),
			Bookings:    5,
		})
	}

	rec, err := svc.GetRecommendation(ctx, ownerID, domain.RoleOwner, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Coefficient < 1.0 {
		t.Errorf("expected coefficient > 1.0 for high occupancy + growing demand, got %f", rec.Coefficient)
	}
	if rec.Coefficient > 1.5 {
		t.Errorf("expected coefficient <= 1.5, got %f", rec.Coefficient)
	}
	if rec.DemandTrend != "growing" {
		t.Errorf("expected demand trend 'growing', got %s", rec.DemandTrend)
	}
	if rec.RecommendedPrice <= rec.CurrentPrice && rec.Coefficient > 1.0 {
		t.Errorf("expected recommended price > current when coefficient > 1.0")
	}
}

func TestSmartPricing_LowOccupancy_DecliningDemand(t *testing.T) {
	svc, bhRepo, analyticsRepo := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	cityID := int64(2)

	bh := &domain.Bathhouse{
		ID:            bathhouseID,
		OwnerID:       ownerID,
		Name:          "Quiet Banya",
		Address:       "Test St",
		CityID:        cityID,
		PricePerHour:  300000, // 3000 RUB
		MaxGuests:     6,
		MinDuration:   1,
		OccupancyRate: 0.15,
		Status:        domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	// Add cheaper competitor
	competitor := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      uuid.New(),
		Name:         "Cheap Banya",
		Address:      "Comp St",
		CityID:       cityID,
		PricePerHour: 150000,
		MaxGuests:    6,
		MinDuration:  1,
		Status:       domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, competitor)

	now := time.Now()

	// Declining demand: fewer bookings recently
	for i := 0; i < 7; i++ {
		_ = analyticsRepo.CreateSnapshot(ctx, &domain.AnalyticsSnapshot{
			BathhouseID: bathhouseID,
			Date:        now.AddDate(0, 0, -14+i),
			Bookings:    5,
		})
	}
	for i := 0; i < 7; i++ {
		_ = analyticsRepo.CreateSnapshot(ctx, &domain.AnalyticsSnapshot{
			BathhouseID: bathhouseID,
			Date:        now.AddDate(0, 0, -7+i),
			Bookings:    1,
		})
	}

	rec, err := svc.GetRecommendation(ctx, ownerID, domain.RoleOwner, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Coefficient > 1.0 {
		t.Errorf("expected coefficient <= 1.0 for low occupancy + declining demand, got %f", rec.Coefficient)
	}
	if rec.Coefficient < 0.8 {
		t.Errorf("expected coefficient >= 0.8, got %f", rec.Coefficient)
	}
	if rec.DemandTrend != "declining" {
		t.Errorf("expected demand trend 'declining', got %s", rec.DemandTrend)
	}
}

func TestSmartPricing_StableDemand_FairPrice(t *testing.T) {
	svc, bhRepo, _ := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	cityID := int64(3)

	bh := &domain.Bathhouse{
		ID:            bathhouseID,
		OwnerID:       ownerID,
		Name:          "Fair Banya",
		Address:       "Test St",
		CityID:        cityID,
		PricePerHour:  200000,
		MaxGuests:     8,
		MinDuration:   1,
		OccupancyRate: 0.5,
		Status:        domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	// Add competitor with similar price
	competitor := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      uuid.New(),
		Name:         "Similar Banya",
		Address:      "Comp St",
		CityID:       cityID,
		PricePerHour: 200000,
		MaxGuests:    8,
		MinDuration:  1,
		Status:       domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, competitor)

	rec, err := svc.GetRecommendation(ctx, ownerID, domain.RoleOwner, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With medium occupancy, stable demand, and fair price, coefficient should be near 1.0
	if rec.Coefficient < 0.9 || rec.Coefficient > 1.1 {
		t.Errorf("expected coefficient near 1.0 for stable conditions, got %f", rec.Coefficient)
	}
	if rec.DemandTrend != "stable" {
		t.Errorf("expected demand trend 'stable', got %s", rec.DemandTrend)
	}
}

func TestSmartPricing_CoefficientClamped(t *testing.T) {
	svc, bhRepo, _ := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID:            bathhouseID,
		OwnerID:       ownerID,
		Name:          "Edge Banya",
		Address:       "Test St",
		CityID:        99,
		PricePerHour:  100000,
		MaxGuests:     4,
		MinDuration:   1,
		OccupancyRate: 0.0, // extremely low
		Status:        domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	rec, err := svc.GetRecommendation(ctx, ownerID, domain.RoleOwner, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Coefficient < 0.8 {
		t.Errorf("coefficient should be clamped at minimum 0.8, got %f", rec.Coefficient)
	}
	if rec.Coefficient > 1.5 {
		t.Errorf("coefficient should be clamped at maximum 1.5, got %f", rec.Coefficient)
	}
}

func TestSmartPricing_Forbidden(t *testing.T) {
	svc, bhRepo, _ := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	otherID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID:           bathhouseID,
		OwnerID:      ownerID,
		Name:         "Private Banya",
		Address:      "Test St",
		CityID:       1,
		PricePerHour: 200000,
		MaxGuests:    6,
		MinDuration:  1,
		Status:       domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	_, err := svc.GetRecommendation(ctx, otherID, domain.RoleClient, bathhouseID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden for non-owner, got %v", err)
	}
}

func TestSmartPricing_NotFound(t *testing.T) {
	svc, _, _ := newSmartPricingService()
	ctx := context.Background()

	_, err := svc.GetRecommendation(ctx, uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSmartPricing_PriceAboveAverage_SuggestsLower(t *testing.T) {
	svc, bhRepo, _ := newSmartPricingService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	cityID := int64(10)

	// Expensive bathhouse
	bh := &domain.Bathhouse{
		ID:            bathhouseID,
		OwnerID:       ownerID,
		Name:          "Expensive Banya",
		Address:       "Test St",
		CityID:        cityID,
		PricePerHour:  500000, // 5000 RUB
		MaxGuests:     6,
		MinDuration:   1,
		OccupancyRate: 0.5,
		Status:        domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	// Several cheap competitors
	for i := 0; i < 3; i++ {
		comp := &domain.Bathhouse{
			ID:           uuid.New(),
			OwnerID:      uuid.New(),
			Name:         "Cheap Banya",
			Address:      "Comp St",
			CityID:       cityID,
			PricePerHour: 200000, // 2000 RUB
			MaxGuests:    6,
			MinDuration:  1,
			Status:       domain.BathhouseStatusActive,
		}
		_ = bhRepo.Create(ctx, comp)
	}

	rec, err := svc.GetRecommendation(ctx, ownerID, domain.RoleOwner, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Coefficient >= 1.0 {
		t.Errorf("expected coefficient < 1.0 when price is 2.5x area average, got %f", rec.Coefficient)
	}
	if rec.AvgAreaPrice >= rec.CurrentPrice {
		t.Errorf("expected avg area price < current price")
	}
}
