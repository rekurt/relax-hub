package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type TemplateRepo struct {
	mu        sync.RWMutex
	templates map[uuid.UUID]*domain.ResponseTemplate
}

func NewTemplateRepo() repository.ResponseTemplateRepository {
	return &TemplateRepo{
		templates: make(map[uuid.UUID]*domain.ResponseTemplate),
	}
}

func (r *TemplateRepo) Create(_ context.Context, template *domain.ResponseTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}

	cp := *template
	r.templates[template.ID] = &cp
	return nil
}

func (r *TemplateRepo) Update(_ context.Context, template *domain.ResponseTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.templates[template.ID]
	if !ok {
		return domain.ErrTemplateNotFound
	}
	existing.Title = template.Title
	existing.Body = template.Body
	existing.SortOrder = template.SortOrder
	return nil
}

func (r *TemplateRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.templates[id]; !ok {
		return domain.ErrTemplateNotFound
	}
	delete(r.templates, id)
	return nil
}

func (r *TemplateRepo) ListByOwner(_ context.Context, ownerID uuid.UUID) ([]domain.ResponseTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ResponseTemplate
	for _, t := range r.templates {
		if t.OwnerID == ownerID {
			cp := *t
			result = append(result, cp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder != result[j].SortOrder {
			return result[i].SortOrder < result[j].SortOrder
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (r *TemplateRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ResponseTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[id]
	if !ok {
		return nil, domain.ErrTemplateNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *TemplateRepo) CountByOwner(_ context.Context, ownerID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, t := range r.templates {
		if t.OwnerID == ownerID {
			count++
		}
	}
	return count, nil
}
