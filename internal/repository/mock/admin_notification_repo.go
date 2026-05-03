package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// AdminNotificationRepo is an in-memory mock implementation of repository.AdminNotificationRepository.
type AdminNotificationRepo struct {
	mu     sync.RWMutex
	notifs map[uuid.UUID]*domain.AdminNotification
}

func NewAdminNotificationRepo() *AdminNotificationRepo {
	return &AdminNotificationRepo{notifs: make(map[uuid.UUID]*domain.AdminNotification)}
}

func (r *AdminNotificationRepo) Create(_ context.Context, notif *domain.AdminNotification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if notif.ID == uuid.Nil {
		notif.ID = uuid.New()
	}
	notif.CreatedAt = time.Now()
	cp := *notif
	r.notifs[notif.ID] = &cp
	return nil
}

func (r *AdminNotificationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.AdminNotification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.notifs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	return &cp, nil
}

func (r *AdminNotificationRepo) List(_ context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.AdminNotification
	for _, n := range r.notifs {
		if filter.Role != nil && n.Role != *filter.Role {
			continue
		}
		if filter.Severity != nil && n.Severity != *filter.Severity {
			continue
		}
		if filter.Type != nil && n.Type != *filter.Type {
			continue
		}
		if filter.IsRead != nil && n.IsRead != *filter.IsRead {
			continue
		}
		cp := *n
		items = append(items, cp)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	return paginate(items, page, pageSize), nil
}

func (r *AdminNotificationRepo) MarkAsRead(_ context.Context, id uuid.UUID, readBy uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notifs[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
	n.ReadBy = &readBy
	return nil
}

func (r *AdminNotificationRepo) MarkAllAsReadByRole(_ context.Context, role domain.AdminSubRole, readBy uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for _, n := range r.notifs {
		if n.Role == role && !n.IsRead {
			n.IsRead = true
			n.ReadAt = &now
			n.ReadBy = &readBy
		}
	}
	return nil
}

func (r *AdminNotificationRepo) CountUnreadByRole(_ context.Context, role domain.AdminSubRole) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, n := range r.notifs {
		if n.Role == role && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *AdminNotificationRepo) ListUnreadCriticalByRole(_ context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.AdminNotification
	for _, n := range r.notifs {
		if n.Role == role && !n.IsRead && (n.Severity == domain.AdminNotifSeverityError || n.Severity == domain.AdminNotifSeverityCritical) {
			cp := *n
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *AdminNotificationRepo) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var deleted int64
	for id, n := range r.notifs {
		if n.CreatedAt.Before(before) {
			delete(r.notifs, id)
			deleted++
		}
	}
	return deleted, nil
}

// CreatedCount returns the total number of notifications stored (for testing).
func (r *AdminNotificationRepo) CreatedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.notifs)
}
