package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ClientReviewRepo struct {
	mu      sync.RWMutex
	reviews map[uuid.UUID]*domain.ClientReview
}

func NewClientReviewRepo() *ClientReviewRepo {
	return &ClientReviewRepo{
		reviews: make(map[uuid.UUID]*domain.ClientReview),
	}
}

var _ repository.ClientReviewRepository = (*ClientReviewRepo)(nil)

func (r *ClientReviewRepo) Create(_ context.Context, review *domain.ClientReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}

	// Check unique constraint on booking_id
	for _, existing := range r.reviews {
		if existing.BookingID == review.BookingID {
			return domain.ErrAlreadyExists
		}
	}

	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ClientReviewRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ClientReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rev, ok := r.reviews[id]
	if !ok {
		return nil, domain.ErrClientReviewNotFound
	}
	cp := *rev
	return &cp, nil
}

func (r *ClientReviewRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.ClientReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rev := range r.reviews {
		if rev.BookingID == bookingID {
			cp := *rev
			return &cp, nil
		}
	}
	return nil, domain.ErrClientReviewNotFound
}

func (r *ClientReviewRepo) Update(_ context.Context, review *domain.ClientReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.reviews[review.ID]; !ok {
		return domain.ErrClientReviewNotFound
	}
	cp := *review
	r.reviews[review.ID] = &cp
	return nil
}

func (r *ClientReviewRepo) ListByClient(_ context.Context, clientID uuid.UUID, onlyRevealed bool, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.ClientReview
	for _, rev := range r.reviews {
		if rev.ClientID == clientID {
			if onlyRevealed && !rev.IsRevealed {
				continue
			}
			cp := *rev
			filtered = append(filtered, cp)
		}
	}

	totalCount := int64(len(filtered))
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.ClientReview]{
		Items:      filtered[start:end],
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func (r *ClientReviewRepo) ListUnrevealedPastDeadline(_ context.Context, now time.Time) ([]domain.ClientReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.ClientReview
	for _, rev := range r.reviews {
		if !rev.IsRevealed && !rev.RevealAt.After(now) {
			cp := *rev
			items = append(items, cp)
		}
	}
	return items, nil
}

func (r *ClientReviewRepo) RevealByID(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rev, ok := r.reviews[id]
	if !ok {
		return domain.ErrClientReviewNotFound
	}
	rev.IsRevealed = true
	rev.UpdatedAt = time.Now()
	return nil
}
