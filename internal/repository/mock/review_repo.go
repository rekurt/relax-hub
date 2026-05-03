package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// ReviewRepo is an in-memory mock implementation of repository.ReviewRepository.
type ReviewRepo struct {
	mu      sync.RWMutex
	reviews map[uuid.UUID]*domain.Review
}

func NewReviewRepo() *ReviewRepo {
	return &ReviewRepo{reviews: make(map[uuid.UUID]*domain.Review)}
}

func (r *ReviewRepo) Create(_ context.Context, review *domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	for _, existing := range r.reviews {
		if existing.UserID == review.UserID && existing.BookingID == review.BookingID {
			return domain.ErrAlreadyExists
		}
	}
	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now
	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ReviewRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rev, ok := r.reviews[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rev
	return &cp, nil
}

func (r *ReviewRepo) Update(_ context.Context, review *domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[review.ID]; !ok {
		return domain.ErrNotFound
	}
	review.UpdatedAt = time.Now()
	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ReviewRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.reviews, id)
	return nil
}

func (r *ReviewRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if rev.BathhouseID == bathhouseID && rev.Status == domain.ReviewStatusApproved {
			items = append(items, *rev)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *ReviewRepo) ListByBathhouseFiltered(_ context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if filter.BathhouseID != nil && rev.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && rev.Status != *filter.Status {
			continue
		}
		if filter.MinRating != nil && rev.Rating < *filter.MinRating {
			continue
		}
		items = append(items, *rev)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *ReviewRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rev := range r.reviews {
		if rev.BookingID == bookingID {
			cp := *rev
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ReviewRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ReviewStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.Status = status
	rev.UpdatedAt = time.Now()
	return nil
}

func (r *ReviewRepo) AddOwnerResponse(_ context.Context, id uuid.UUID, response string, respondedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.OwnerResponse = response
	rev.OwnerResponseAt = &respondedAt
	rev.UpdatedAt = respondedAt
	return nil
}

func (r *ReviewRepo) GetUserReviewStats(_ context.Context, userID uuid.UUID) (*domain.UserReviewStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var stats domain.UserReviewStats
	var totalRating int
	for _, rev := range r.reviews {
		if rev.UserID == userID && rev.Status == domain.ReviewStatusApproved {
			stats.ReviewCount++
			totalRating += rev.Rating
		}
	}
	if stats.ReviewCount > 0 {
		stats.AvgRating = float64(totalRating) / float64(stats.ReviewCount)
	}
	return &stats, nil
}

func (r *ReviewRepo) UpdateStatusWithReasons(_ context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	rev.Status = status
	rev.RejectionReasons = reasons
	rev.UpdatedAt = time.Now()
	return nil
}

func (r *ReviewRepo) CountPendingReviews(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, rev := range r.reviews {
		if rev.Status == domain.ReviewStatusPending {
			count++
		}
	}
	return count, nil
}

func (r *ReviewRepo) ListAllReviews(_ context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Review
	for _, rev := range r.reviews {
		if filter.BathhouseID != nil && rev.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && rev.Status != *filter.Status {
			continue
		}
		if filter.MinRating != nil && rev.Rating < *filter.MinRating {
			continue
		}
		if filter.MaxRating != nil && rev.Rating > *filter.MaxRating {
			continue
		}
		if filter.FromDate != nil && rev.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && rev.CreatedAt.After(*filter.ToDate) {
			continue
		}
		items = append(items, *rev)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *ReviewRepo) GetCriteriaAverages(_ context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var avgs domain.ReviewCriteriaAverages
	var count float64
	for _, rev := range r.reviews {
		if rev.BathhouseID == bathhouseID && rev.Status == domain.ReviewStatusApproved && rev.Cleanliness != nil {
			avgs.AvgCleanliness += *rev.Cleanliness
			avgs.AvgAccuracy += *rev.Accuracy
			avgs.AvgCommunication += *rev.Communication
			avgs.AvgValueForMoney += *rev.ValueForMoney
			count++
		}
	}
	if count > 0 {
		avgs.AvgCleanliness /= count
		avgs.AvgAccuracy /= count
		avgs.AvgCommunication /= count
		avgs.AvgValueForMoney /= count
	}
	return &avgs, nil
}

func (r *ReviewRepo) GetPlatformAverageRating(_ context.Context) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total float64
	var count float64
	for _, rev := range r.reviews {
		if rev.Status == domain.ReviewStatusApproved {
			total += float64(rev.Rating)
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / count, nil
}

func (r *ReviewRepo) ListUnrevealedPastDeadline(_ context.Context, now time.Time) ([]domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Review
	for _, rev := range r.reviews {
		if !rev.IsRevealed && rev.RevealAt != nil && !rev.RevealAt.After(now) {
			result = append(result, *rev)
		}
	}
	return result, nil
}
