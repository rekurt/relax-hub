package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/repository"
)

func TestNewUserRepository_ImplementsInterface(t *testing.T) {
	var _ repository.UserRepository = NewUserRepository(&pgxpool.Pool{})
}

func TestNewCityRepository_ImplementsInterface(t *testing.T) {
	var _ repository.CityRepository = NewCityRepository(&pgxpool.Pool{})
}

func TestNewBathhouseRepository_ImplementsInterface(t *testing.T) {
	var _ repository.BathhouseRepository = NewBathhouseRepository(&pgxpool.Pool{})
}

func TestNewBookingRepository_ImplementsInterface(t *testing.T) {
	var _ repository.BookingRepository = NewBookingRepository(&pgxpool.Pool{})
}

func TestNewReviewRepository_ImplementsInterface(t *testing.T) {
	var _ repository.ReviewRepository = NewReviewRepository(&pgxpool.Pool{})
}

func TestNewRepresentativeRepository_ImplementsInterface(t *testing.T) {
	var _ repository.RepresentativeRepository = NewRepresentativeRepository(&pgxpool.Pool{})
}

func TestNewRecommendationRepository_ImplementsInterface(t *testing.T) {
	var _ repository.RecommendationRepository = NewRecommendationRepository(&pgxpool.Pool{})
}

func TestNewSubscriptionRepository_ImplementsInterface(t *testing.T) {
	var _ repository.SubscriptionRepository = NewSubscriptionRepository(&pgxpool.Pool{})
}

func TestNewPromotionRepository_ImplementsInterface(t *testing.T) {
	var _ repository.PromotionRepository = NewPromotionRepository(&pgxpool.Pool{})
}

