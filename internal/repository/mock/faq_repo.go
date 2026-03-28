package mock

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type FAQRepo struct {
	mu   sync.RWMutex
	faqs map[uuid.UUID]*domain.FAQ
}

func NewFAQRepo() *FAQRepo {
	return &FAQRepo{
		faqs: make(map[uuid.UUID]*domain.FAQ),
	}
}

var _ repository.FAQRepository = (*FAQRepo)(nil)

func (r *FAQRepo) Create(_ context.Context, faq *domain.FAQ) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if faq.ID == uuid.Nil {
		faq.ID = uuid.New()
	}
	now := time.Now()
	if faq.CreatedAt.IsZero() {
		faq.CreatedAt = now
	}
	if faq.UpdatedAt.IsZero() {
		faq.UpdatedAt = now
	}
	cp := *faq
	r.faqs[faq.ID] = &cp
	return nil
}

func (r *FAQRepo) Update(_ context.Context, faq *domain.FAQ) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.faqs[faq.ID]; !ok {
		return domain.ErrFAQNotFound
	}
	cp := *faq
	r.faqs[faq.ID] = &cp
	return nil
}

func (r *FAQRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.faqs[id]; !ok {
		return domain.ErrFAQNotFound
	}
	delete(r.faqs, id)
	return nil
}

func (r *FAQRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.FAQ, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	faq, ok := r.faqs[id]
	if !ok {
		return nil, domain.ErrFAQNotFound
	}
	cp := *faq
	return &cp, nil
}

func (r *FAQRepo) List(_ context.Context, filter domain.FAQFilter) (*domain.PaginatedResult[domain.FAQ], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var items []domain.FAQ
	for _, faq := range r.faqs {
		if filter.Category != nil && faq.Category != *filter.Category {
			continue
		}
		if filter.Active != nil && faq.Active != *filter.Active {
			continue
		}
		items = append(items, *faq)
	}

	totalCount := int64(len(items))
	offset := (filter.Page - 1) * filter.PageSize
	end := offset + filter.PageSize
	if offset > len(items) {
		offset = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	page := items[offset:end]

	return &domain.PaginatedResult[domain.FAQ]{
		Items:      page,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int((totalCount + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
	}, nil
}

func (r *FAQRepo) SearchByKeywords(_ context.Context, query string, limit int) ([]domain.FAQMatch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 3
	}
	queryLower := strings.ToLower(query)

	var matches []domain.FAQMatch
	for _, faq := range r.faqs {
		if !faq.Active {
			continue
		}

		var score float64
		// Simple keyword matching for mock
		questionLower := strings.ToLower(faq.Question)
		if strings.Contains(questionLower, queryLower) {
			score = 0.8
		}
		for _, kw := range faq.Keywords {
			if strings.Contains(strings.ToLower(kw), queryLower) || strings.Contains(queryLower, strings.ToLower(kw)) {
				if score < 0.6 {
					score = 0.6
				}
			}
		}

		if score > 0 {
			cp := *faq
			matches = append(matches, domain.FAQMatch{FAQ: cp, Score: score})
		}
	}

	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}

func (r *FAQRepo) ListActiveByCategory(_ context.Context, category domain.FAQCategory) ([]domain.FAQ, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.FAQ
	for _, faq := range r.faqs {
		if faq.Active && faq.Category == category {
			items = append(items, *faq)
		}
	}
	return items, nil
}
