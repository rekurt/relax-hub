package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/repository"
)

func TestNewUserRepository_ImplementsInterface(t *testing.T) {
	var _ repository.UserRepository = NewUserRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewCityRepository_ImplementsInterface(t *testing.T) {
	var _ repository.CityRepository = NewCityRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewBathhouseRepository_ImplementsInterface(t *testing.T) {
	var _ repository.BathhouseRepository = NewBathhouseRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewBookingRepository_ImplementsInterface(t *testing.T) {
	var _ repository.BookingRepository = NewBookingRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewReviewRepository_ImplementsInterface(t *testing.T) {
	var _ repository.ReviewRepository = NewReviewRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewRepresentativeRepository_ImplementsInterface(t *testing.T) {
	var _ repository.RepresentativeRepository = NewRepresentativeRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewRecommendationRepository_ImplementsInterface(t *testing.T) {
	var _ repository.RecommendationRepository = NewRecommendationRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewSubscriptionRepository_ImplementsInterface(t *testing.T) {
	var _ repository.SubscriptionRepository = NewSubscriptionRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewPromotionRepository_ImplementsInterface(t *testing.T) {
	var _ repository.PromotionRepository = NewPromotionRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}

func TestNewAnalyticsRepository_ImplementsInterface(t *testing.T) {
	var _ repository.AnalyticsRepository = NewAnalyticsRepository(&pgxpool.Pool{}) //nolint:staticcheck // explicit type for interface compliance check
}
