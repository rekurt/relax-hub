package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// PriceRecommendation holds the smart pricing recommendation for a bathhouse.
type PriceRecommendation struct {
	CurrentPrice        int64   `json:"current_price"`
	RecommendedPrice    int64   `json:"recommended_price"`
	Coefficient         float64 `json:"coefficient"`
	AvgAreaPrice        int64   `json:"avg_area_price"`
	OccupancyRate       float64 `json:"occupancy_rate"`
	DemandTrend         string  `json:"demand_trend"`
	RecommendationBasis string  `json:"recommendation_basis"`
}

// SmartPricingService provides pricing recommendations based on market data.
type SmartPricingService interface {
	GetRecommendation(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*PriceRecommendation, error)
}

type smartPricingService struct {
	bhRepo        repository.BathhouseRepository
	analyticsRepo repository.AnalyticsRepository
	access        *AccessChecker
	logger        *logger.Logger
}

func NewSmartPricingService(
	bhRepo repository.BathhouseRepository,
	analyticsRepo repository.AnalyticsRepository,
	access *AccessChecker,
	log *logger.Logger,
) SmartPricingService {
	return &smartPricingService{
		bhRepo:        bhRepo,
		analyticsRepo: analyticsRepo,
		access:        access,
		logger:        log,
	}
}

func (s *smartPricingService) GetRecommendation(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (*PriceRecommendation, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	avgAreaPrice := s.calculateAvgAreaPrice(ctx, bh)
	demandTrend, demandCoefficient := s.calculateDemandTrend(ctx, bathhouseID)
	occupancyCoefficient := s.calculateOccupancyCoefficient(bh.OccupancyRate)
	pricePositionCoefficient := s.calculatePricePositionCoefficient(bh.PricePerHour, avgAreaPrice)

	// Weighted combination: occupancy 40%, demand 30%, price position 30%
	rawCoefficient := occupancyCoefficient*0.4 + demandCoefficient*0.3 + pricePositionCoefficient*0.3

	// Clamp to [0.8, 1.5] range
	coefficient := math.Max(0.8, math.Min(1.5, rawCoefficient))
	// Round to 2 decimal places
	coefficient = math.Round(coefficient*100) / 100

	recommendedPrice := int64(math.Round(float64(bh.PricePerHour) * coefficient))

	basis := s.describeBasis(bh.OccupancyRate, demandTrend, bh.PricePerHour, avgAreaPrice)

	return &PriceRecommendation{
		CurrentPrice:        bh.PricePerHour,
		RecommendedPrice:    recommendedPrice,
		Coefficient:         coefficient,
		AvgAreaPrice:        avgAreaPrice,
		OccupancyRate:       bh.OccupancyRate,
		DemandTrend:         demandTrend,
		RecommendationBasis: basis,
	}, nil
}

// calculateAvgAreaPrice computes the average price per hour for active bathhouses in the same city.
func (s *smartPricingService) calculateAvgAreaPrice(ctx context.Context, bh *domain.Bathhouse) int64 {
	cityID := bh.CityID
	filter := domain.BathhouseFilter{
		CityID:   &cityID,
		Page:     1,
		PageSize: 1000,
	}

	result, err := s.bhRepo.List(ctx, filter)
	if err != nil {
		s.logger.Warn("failed to fetch area bathhouses for pricing", "error", err)
		return bh.PricePerHour
	}

	if len(result.Items) == 0 {
		return bh.PricePerHour
	}

	var totalPrice int64
	var count int64
	for _, item := range result.Items {
		if item.Status == domain.BathhouseStatusActive && item.ID != bh.ID {
			totalPrice += item.PricePerHour
			count++
		}
	}

	if count == 0 {
		return bh.PricePerHour
	}

	return totalPrice / count
}

// calculateDemandTrend compares recent 7-day bookings to previous 7-day bookings.
func (s *smartPricingService) calculateDemandTrend(ctx context.Context, bathhouseID uuid.UUID) (string, float64) {
	now := time.Now()
	recentFrom := now.AddDate(0, 0, -7)
	previousFrom := now.AddDate(0, 0, -14)

	recentStats, err := s.analyticsRepo.GetBathhouseStats(ctx, bathhouseID, recentFrom, now)
	if err != nil {
		s.logger.Warn("failed to get recent stats for pricing", "error", err)
		return "stable", 1.0
	}

	prevStats, err := s.analyticsRepo.GetBathhouseStats(ctx, bathhouseID, previousFrom, recentFrom)
	if err != nil {
		s.logger.Warn("failed to get previous stats for pricing", "error", err)
		return "stable", 1.0
	}

	if prevStats.Bookings == 0 && recentStats.Bookings == 0 {
		return "stable", 1.0
	}

	if prevStats.Bookings == 0 {
		return "growing", 1.15
	}

	ratio := float64(recentStats.Bookings) / float64(prevStats.Bookings)

	switch {
	case ratio > 1.3:
		return "growing", 1.15
	case ratio > 1.1:
		return "growing", 1.08
	case ratio < 0.7:
		return "declining", 0.88
	case ratio < 0.9:
		return "declining", 0.94
	default:
		return "stable", 1.0
	}
}

// calculateOccupancyCoefficient suggests price adjustments based on occupancy.
// High occupancy (>80%) suggests room to raise prices; low (<30%) suggests lowering.
func (s *smartPricingService) calculateOccupancyCoefficient(occupancyRate float64) float64 {
	switch {
	case occupancyRate >= 0.9:
		return 1.3
	case occupancyRate >= 0.8:
		return 1.2
	case occupancyRate >= 0.6:
		return 1.1
	case occupancyRate >= 0.4:
		return 1.0
	case occupancyRate >= 0.2:
		return 0.9
	default:
		return 0.85
	}
}

// calculatePricePositionCoefficient compares this bathhouse's price to area average.
// If significantly above average, suggests lowering; if below, suggests raising.
func (s *smartPricingService) calculatePricePositionCoefficient(currentPrice, avgAreaPrice int64) float64 {
	if avgAreaPrice == 0 {
		return 1.0
	}

	ratio := float64(currentPrice) / float64(avgAreaPrice)

	switch {
	case ratio > 1.5:
		return 0.85
	case ratio > 1.2:
		return 0.92
	case ratio < 0.7:
		return 1.12
	case ratio < 0.85:
		return 1.06
	default:
		return 1.0
	}
}

func (s *smartPricingService) describeBasis(occupancy float64, demandTrend string, currentPrice, avgAreaPrice int64) string {
	switch {
	case occupancy >= 0.8 && demandTrend == "growing":
		return "Высокая загрузка и растущий спрос позволяют повысить цену"
	case occupancy >= 0.8:
		return "Высокая загрузка — можно рассмотреть повышение цены"
	case occupancy < 0.3 && demandTrend == "declining":
		return "Низкая загрузка и снижающийся спрос — рекомендуем снизить цену для привлечения клиентов"
	case occupancy < 0.3:
		return "Низкая загрузка — снижение цены может привлечь больше клиентов"
	case currentPrice > 0 && avgAreaPrice > 0 && float64(currentPrice)/float64(avgAreaPrice) > 1.3:
		return "Цена значительно выше среднерыночной — рекомендуем скорректировать"
	case currentPrice > 0 && avgAreaPrice > 0 && float64(currentPrice)/float64(avgAreaPrice) < 0.75:
		return "Цена ниже среднерыночной — есть потенциал для повышения"
	case demandTrend == "growing":
		return "Растущий спрос позволяет скорректировать цену вверх"
	case demandTrend == "declining":
		return "Спрос снижается — рассмотрите привлекательную цену"
	default:
		return "Текущая цена соответствует рыночным условиям"
	}
}
