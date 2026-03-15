package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// ComplaintRepo is an in-memory mock implementation of repository.ComplaintRepository
type ComplaintRepo struct {
	mu         sync.RWMutex
	complaints map[uuid.UUID]*domain.Complaint
}

func NewComplaintRepo() repository.ComplaintRepository {
	return &ComplaintRepo{
		complaints: make(map[uuid.UUID]*domain.Complaint),
	}
}

func (r *ComplaintRepo) Create(_ context.Context, complaint *domain.Complaint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if complaint.ID == uuid.Nil {
		complaint.ID = uuid.New()
	}

	// Check unique constraint: one complaint per reporter per target
	for _, c := range r.complaints {
		if c.ReporterID == complaint.ReporterID &&
			c.TargetType == complaint.TargetType &&
			c.TargetID == complaint.TargetID {
			return domain.ErrAlreadyReported
		}
	}

	if complaint.Status == "" {
		complaint.Status = domain.ComplaintStatusPending
	}
	if complaint.CreatedAt.IsZero() {
		complaint.CreatedAt = time.Now()
	}

	cp := *complaint
	r.complaints[complaint.ID] = &cp
	return nil
}

func (r *ComplaintRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Complaint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.complaints[id]
	if !ok {
		return nil, domain.ErrComplaintNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *ComplaintRepo) List(_ context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Complaint
	for _, c := range r.complaints {
		if filter.Status != nil && c.Status != *filter.Status {
			continue
		}
		if filter.TargetType != nil && c.TargetType != *filter.TargetType {
			continue
		}
		if filter.Reason != nil && c.Reason != *filter.Reason {
			continue
		}
		if filter.FromDate != nil && c.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && c.CreatedAt.After(*filter.ToDate) {
			continue
		}
		cp := *c
		filtered = append(filtered, cp)
	}

	// Sort by created_at DESC
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *ComplaintRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ComplaintStatus, resolvedByID *uuid.UUID, resolution string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.complaints[id]
	if !ok {
		return domain.ErrComplaintNotFound
	}

	c.Status = status
	c.ResolvedByID = resolvedByID
	c.Resolution = resolution
	if status == domain.ComplaintStatusResolved || status == domain.ComplaintStatusDismissed {
		now := time.Now()
		c.ResolvedAt = &now
	}
	return nil
}

func (r *ComplaintRepo) CountByTarget(_ context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, c := range r.complaints {
		if c.TargetType == targetType && c.TargetID == targetID && c.Status == domain.ComplaintStatusPending {
			count++
		}
	}
	return count, nil
}

func (r *ComplaintRepo) CheckExists(_ context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.complaints {
		if c.ReporterID == reporterID && c.TargetType == targetType && c.TargetID == targetID {
			return true, nil
		}
	}
	return false, nil
}
