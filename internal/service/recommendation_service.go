package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// RecommendationService provides personalized bathhouse recommendations
type RecommendationService interface {
	GetPersonalized(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error)
	GetSimilar(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error)
	GetPopular(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error)
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.UserPreferences) error
	RecordView(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID) error
}

type recommendationService struct {
	recRepo repository.RecommendationRepository
	bhRepo  repository.BathhouseRepository
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(recRepo repository.RecommendationRepository, bhRepo repository.BathhouseRepository) RecommendationService {
	return &recommendationService{
		recRepo: recRepo,
		bhRepo:  bhRepo,
	}
}

// GetPersonalized returns personalized recommendations for a user
// Algorithm:
// 1. Get user's explicit preferences
// 2. Find similar users (collaborative filtering)
// 3. Get bathhouses booked by similar users but not the current user
// 4. Filter by user preferences (city, price, amenities)
// 5. Score by rating * similarity weight * recency bonus
func (s *recommendationService) GetPersonalized(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// Get user's explicit preferences
	userPrefs, err := s.recRepo.GetUserPreferences(ctx, userID)
	if err != nil && err != domain.ErrNotFound {
		return nil, 0, err
	}

	// Get user's booking history (to exclude from recommendations)
	userBookedBaths, err := s.recRepo.GetUserBookedBathhouses(ctx, userID, 1000)
	if err != nil {
		return nil, 0, err
	}
	bookedSet := make(map[uuid.UUID]bool)
	for _, id := range userBookedBaths {
		bookedSet[id] = true
	}

	// Find similar users
	similarUsers, err := s.recRepo.GetSimilarUsers(ctx, userID, 10)
	if err != nil {
		return nil, 0, err
	}

	// Collect bathhouses booked by similar users
	type scoreItem struct {
		bathhouseID uuid.UUID
		score       float64
	}

	bathhouseScores := make(map[uuid.UUID]float64)

	for i, similarUserID := range similarUsers {
		// Weight by similarity rank (first similar user has highest weight)
		similarity := 1.0 / (1.0 + float64(i)*0.1)

		booked, err := s.recRepo.GetUserBookedBathhouses(ctx, similarUserID, 100)
		if err != nil {
			continue
		}

		for _, bathID := range booked {
			// Skip if user already booked this bathhouse
			if bookedSet[bathID] {
				continue
			}

			// Accumulate score
			bathhouseScores[bathID] += similarity
		}
	}

	// Convert scores to items and fetch bathhouse details for filtering
	var candidates []scoreItem
	for bathID, score := range bathhouseScores {
		candidates = append(candidates, scoreItem{bathID, score})
	}

	// Fetch bathhouse details for filtering and scoring
	filtered := make([]scoreItem, 0, len(candidates))
	for _, item := range candidates {
		bh, err := s.bhRepo.GetByID(ctx, item.bathhouseID)
		if err != nil || bh.Status != domain.BathhouseStatusActive {
			continue
		}

		// Apply preference filters
		if !s.matchesPreferences(bh, userPrefs) {
			continue
		}

		// Enhance score with rating
		ratingBonus := bh.Rating / 5.0 // normalize rating 0-1
		finalScore := item.score * (1.0 + ratingBonus)

		filtered = append(filtered, scoreItem{item.bathhouseID, finalScore})
	}

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].score > filtered[j].score
	})

	// Apply pagination
	total := int64(len(filtered))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	results := make([]uuid.UUID, 0, end-start)
	for i := start; i < end; i++ {
		results = append(results, filtered[i].bathhouseID)
	}

	return results, total, nil
}

// GetSimilar returns bathhouses similar to the given bathhouse
func (s *recommendationService) GetSimilar(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit < 1 {
		limit = 10
	}

	return s.recRepo.GetSimilarBathhouses(ctx, bathhouseID, limit)
}

// GetPopular returns popular bathhouses in a city
func (s *recommendationService) GetPopular(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
	if limit < 1 {
		limit = 10
	}

	return s.recRepo.GetPopularBathhouses(ctx, cityID, limit)
}

// GetUserPreferences returns the user's current preferences
func (s *recommendationService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}

	return s.recRepo.GetUserPreferences(ctx, userID)
}

// UpdatePreferences updates user's explicit preferences
func (s *recommendationService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.UserPreferences) error {
	if userID == uuid.Nil {
		return domain.ErrInvalidInput
	}

	if prefs == nil {
		return domain.ErrInvalidInput
	}

	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()

	if err := prefs.Validate(); err != nil {
		return err
	}

	return s.recRepo.SaveUserPreferences(ctx, prefs)
}

// RecordView records a user's view of a bathhouse
func (s *recommendationService) RecordView(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID) error {
	if userID == uuid.Nil || bathhouseID == uuid.Nil {
		return domain.ErrInvalidInput
	}

	activity := &domain.UserActivity{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: bathhouseID,
		Type:        domain.ActivityTypeView,
		CreatedAt:   time.Now(),
	}

	if err := activity.Validate(); err != nil {
		return err
	}

	return s.recRepo.RecordActivity(ctx, activity)
}

// matchesPreferences checks if a bathhouse matches user preferences
func (s *recommendationService) matchesPreferences(bh *domain.Bathhouse, prefs *domain.UserPreferences) bool {
	if prefs == nil {
		return true // No preferences means all bathhouses match
	}

	// Check city preference
	if prefs.PreferredCityID != nil && bh.CityID != *prefs.PreferredCityID {
		return false
	}

	// Check price range
	if prefs.PriceRangeMin != nil && bh.PricePerHour < *prefs.PriceRangeMin {
		return false
	}
	if prefs.PriceRangeMax != nil && bh.PricePerHour > *prefs.PriceRangeMax {
		return false
	}

	// Check amenities preferences
	// Only filter if user explicitly prefers certain amenities
	if (prefs.PreferPool || prefs.PreferSauna || prefs.PreferSteamRoom ||
		prefs.PreferHotTub || prefs.PreferBBQ || prefs.PreferKaraoke) {

		// Count how many preferred amenities the user wants
		preferredCount := 0
		if prefs.PreferPool {
			preferredCount++
		}
		if prefs.PreferSauna {
			preferredCount++
		}
		if prefs.PreferSteamRoom {
			preferredCount++
		}
		if prefs.PreferHotTub {
			preferredCount++
		}
		if prefs.PreferBBQ {
			preferredCount++
		}
		if prefs.PreferKaraoke {
			preferredCount++
		}

		// Count how many preferences the bathhouse has
		matchedCount := 0
		if prefs.PreferPool && bh.HasPool {
			matchedCount++
		}
		if prefs.PreferSauna && bh.HasSauna {
			matchedCount++
		}
		if prefs.PreferSteamRoom && bh.HasSteamRoom {
			matchedCount++
		}
		if prefs.PreferHotTub && bh.HasHotTub {
			matchedCount++
		}
		if prefs.PreferBBQ && bh.HasBBQ {
			matchedCount++
		}
		if prefs.PreferKaraoke && bh.HasKaraoke {
			matchedCount++
		}

		// Require at least one amenity match if user has preferences
		if matchedCount == 0 {
			return false
		}
	}

	return true
}
