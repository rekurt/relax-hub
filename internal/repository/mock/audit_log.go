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

type AuditLogRepo struct {
	mu   sync.RWMutex
	logs map[uuid.UUID]*domain.AuditLog
}

func NewAuditLogRepo() repository.AuditLogRepository {
	return &AuditLogRepo{
		logs: make(map[uuid.UUID]*domain.AuditLog),
	}
}

func (r *AuditLogRepo) Create(_ context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	cp := *log
	r.logs[log.ID] = &cp
	return nil
}

func (r *AuditLogRepo) ListByEntity(_ context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.AuditLog
	for _, l := range r.logs {
		if l.EntityType == entityType && l.EntityID == entityID {
			cp := *l
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *AuditLogRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.AuditLog
	for _, l := range r.logs {
		if l.UserID == userID {
			cp := *l
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *AuditLogRepo) List(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.AuditLog
	for _, l := range r.logs {
		if filter.EntityType != nil && l.EntityType != *filter.EntityType {
			continue
		}
		if filter.EntityID != nil && l.EntityID != *filter.EntityID {
			continue
		}
		if filter.UserID != nil && l.UserID != *filter.UserID {
			continue
		}
		if filter.Action != nil && l.Action != *filter.Action {
			continue
		}
		if filter.FromDate != nil && l.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && l.CreatedAt.After(*filter.ToDate) {
			continue
		}
		cp := *l
		filtered = append(filtered, cp)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}
