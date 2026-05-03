package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// RecommendationRepo is an in-memory mock implementation of repository.RecommendationRepository.
type RecommendationRepo struct {
	mu          sync.RWMutex
	preferences map[uuid.UUID]*domain.UserPreferences
	activities  []domain.UserActivity
	bookings    map[uuid.UUID][]uuid.UUID // userID -> list of bathhouse IDs
}

func NewRecommendationRepo() *RecommendationRepo {
	return &RecommendationRepo{
		preferences: make(map[uuid.UUID]*domain.UserPreferences),
		activities:  make([]domain.UserActivity, 0),
		bookings:    make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *RecommendationRepo) GetUserPreferences(_ context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefs, ok := r.preferences[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *prefs
	return &cp, nil
}

func (r *RecommendationRepo) SaveUserPreferences(_ context.Context, prefs *domain.UserPreferences) error {
	if err := prefs.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if prefs.UpdatedAt.IsZero() {
		prefs.UpdatedAt = time.Now()
	}
	cp := *prefs
	r.preferences[prefs.UserID] = &cp
	return nil
}

func (r *RecommendationRepo) RecordActivity(_ context.Context, activity *domain.UserActivity) error {
	if err := activity.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	r.activities = append(r.activities, *activity)
	return nil
}

func (r *RecommendationRepo) GetUserBookedBathhouses(_ context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	booked, ok := r.bookings[userID]
	if !ok {
		return []uuid.UUID{}, nil
	}

	if limit <= 0 {
		limit = 50
	}
	if len(booked) > limit {
		return booked[:limit], nil
	}
	return booked, nil
}

func (r *RecommendationRepo) GetUserBookedBathhousesWithDates(_ context.Context, userID uuid.UUID, limit int) ([]domain.BookedBathhouseWithDate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	booked, ok := r.bookings[userID]
	if !ok {
		return []domain.BookedBathhouseWithDate{}, nil
	}

	if limit <= 0 {
		limit = 50
	}

	// For mock, we return with current time as booking date
	// In reality, this would come from booking timestamps
	var results []domain.BookedBathhouseWithDate
	end := limit
	if len(booked) < limit {
		end = len(booked)
	}
	for i := 0; i < end; i++ {
		results = append(results, domain.BookedBathhouseWithDate{
			BathhouseID: booked[i],
			BookedAt:    time.Now().AddDate(0, 0, -i), // Mock data: each booking is 1 day older
		})
	}
	return results, nil
}

func (r *RecommendationRepo) GetSimilarUsers(_ context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	r.mu.RLock()

	if limit <= 0 {
		limit = 20
	}

	// Find users with overlapping bookings
	userBookings := make(map[uuid.UUID]bool)
	if booked, ok := r.bookings[userID]; ok {
		for _, bhID := range booked {
			userBookings[bhID] = true
		}
	}

	type userMatch struct {
		userID uuid.UUID
		score  int
	}
	matches := make([]userMatch, 0)

	for otherUserID, otherBookings := range r.bookings {
		if otherUserID == userID {
			continue
		}
		score := 0
		for _, bhID := range otherBookings {
			if userBookings[bhID] {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, userMatch{otherUserID, score})
		}
	}

	r.mu.RUnlock()

	// Sort by score descending (outside the lock)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	var result []uuid.UUID
	for i := 0; i < len(matches) && i < limit; i++ {
		result = append(result, matches[i].userID)
	}
	return result, nil
}

func (r *RecommendationRepo) GetPopularBathhouses(_ context.Context, cityID int64, limit int) ([]uuid.UUID, error) {
	// Mock implementation returns empty list
	// In real implementation would require bathhouse repo access
	return []uuid.UUID{}, nil
}

func (r *RecommendationRepo) GetSimilarBathhouses(_ context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error) {
	// Mock implementation returns empty list
	// In real implementation would require bathhouse repo access
	return []uuid.UUID{}, nil
}
