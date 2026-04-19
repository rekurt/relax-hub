package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type PMSConnectionRepo struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]*domain.PMSConnection
}

func NewPMSConnectionRepo() repository.PMSConnectionRepository {
	return &PMSConnectionRepo{
		conns: make(map[uuid.UUID]*domain.PMSConnection),
	}
}

func (r *PMSConnectionRepo) Create(_ context.Context, conn *domain.PMSConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now
	c := *conn
	r.conns[c.ID] = &c
	return nil
}

func (r *PMSConnectionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PMSConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.conns[id]
	if !ok {
		return nil, domain.ErrPMSConnectionNotFound
	}
	copy := *c
	return &copy, nil
}

func (r *PMSConnectionRepo) Update(_ context.Context, conn *domain.PMSConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.conns[conn.ID]; !ok {
		return domain.ErrPMSConnectionNotFound
	}
	conn.UpdatedAt = time.Now()
	c := *conn
	r.conns[c.ID] = &c
	return nil
}

func (r *PMSConnectionRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.conns[id]; !ok {
		return domain.ErrPMSConnectionNotFound
	}
	delete(r.conns, id)
	return nil
}

func (r *PMSConnectionRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSConnection], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.PMSConnection
	for _, c := range r.conns {
		if c.OwnerID == ownerID {
			items = append(items, *c)
		}
	}
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= int(total) {
		return &domain.PaginatedResult[domain.PMSConnection]{
			Items:      []domain.PMSConnection{},
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	return &domain.PaginatedResult[domain.PMSConnection]{
		Items:      items[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *PMSConnectionRepo) GetByBathhouseID(_ context.Context, bathhouseID uuid.UUID) (*domain.PMSConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.conns {
		if c.BathhouseID == bathhouseID {
			copy := *c
			return &copy, nil
		}
	}
	return nil, domain.ErrPMSConnectionNotFound
}

func (r *PMSConnectionRepo) ListActive(_ context.Context) ([]domain.PMSConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.PMSConnection
	for _, c := range r.conns {
		if c.Status == domain.PMSConnectionActive {
			items = append(items, *c)
		}
	}
	return items, nil
}

func (r *PMSConnectionRepo) UpdateSyncStatus(_ context.Context, id uuid.UUID, lastSyncAt time.Time, lastSyncError string, status domain.PMSConnectionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.conns[id]
	if !ok {
		return domain.ErrPMSConnectionNotFound
	}
	c.LastSyncAt = &lastSyncAt
	c.LastSyncError = lastSyncError
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}

type PMSSyncLogRepo struct {
	mu   sync.RWMutex
	logs map[uuid.UUID]*domain.PMSSyncLog
}

func NewPMSSyncLogRepo() repository.PMSSyncLogRepository {
	return &PMSSyncLogRepo{
		logs: make(map[uuid.UUID]*domain.PMSSyncLog),
	}
}

func (r *PMSSyncLogRepo) Create(_ context.Context, log *domain.PMSSyncLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	l := *log
	r.logs[l.ID] = &l
	return nil
}

func (r *PMSSyncLogRepo) ListByConnection(_ context.Context, connectionID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSSyncLog], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.PMSSyncLog
	for _, l := range r.logs {
		if l.ConnectionID == connectionID {
			items = append(items, *l)
		}
	}
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= int(total) {
		return &domain.PaginatedResult[domain.PMSSyncLog]{
			Items:      []domain.PMSSyncLog{},
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	return &domain.PaginatedResult[domain.PMSSyncLog]{
		Items:      items[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}
