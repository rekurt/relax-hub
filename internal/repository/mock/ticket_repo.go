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

type TicketRepo struct {
	mu       sync.RWMutex
	tickets  map[uuid.UUID]*domain.Ticket
	messages map[uuid.UUID][]domain.TicketMessage
}

func NewTicketRepo() repository.TicketRepository {
	return &TicketRepo{
		tickets:  make(map[uuid.UUID]*domain.Ticket),
		messages: make(map[uuid.UUID][]domain.TicketMessage),
	}
}

func (r *TicketRepo) Create(_ context.Context, ticket *domain.Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}
	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = now
	}
	if ticket.UpdatedAt.IsZero() {
		ticket.UpdatedAt = now
	}

	cp := *ticket
	r.tickets[ticket.ID] = &cp
	return nil
}

func (r *TicketRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tickets[id]
	if !ok {
		return nil, domain.ErrTicketNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *TicketRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Ticket
	for _, t := range r.tickets {
		if t.UserID == userID {
			cp := *t
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *TicketRepo) ListAll(_ context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Ticket
	for _, t := range r.tickets {
		if filter.UserID != nil && t.UserID != *filter.UserID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.Priority != nil && t.Priority != *filter.Priority {
			continue
		}
		if filter.Level != nil && t.Level != *filter.Level {
			continue
		}
		if filter.Category != nil && t.Category != *filter.Category {
			continue
		}
		cp := *t
		filtered = append(filtered, cp)
	}

	// Sort by priority then created_at
	priorityOrder := map[domain.TicketPriority]int{
		domain.TicketPriorityCritical: 0,
		domain.TicketPriorityHigh:     1,
		domain.TicketPriorityMedium:   2,
		domain.TicketPriorityLow:      3,
	}
	sort.Slice(filtered, func(i, j int) bool {
		pi := priorityOrder[filtered[i].Priority]
		pj := priorityOrder[filtered[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *TicketRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.TicketStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

func (r *TicketRepo) UpdateLevel(_ context.Context, id uuid.UUID, level domain.TicketLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.Level = level
	t.Status = domain.TicketStatusEscalated
	t.UpdatedAt = time.Now()
	return nil
}

func (r *TicketRepo) Assign(_ context.Context, id uuid.UUID, assignedTo uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.AssignedTo = &assignedTo
	t.Status = domain.TicketStatusInProgress
	t.UpdatedAt = time.Now()
	return nil
}

func (r *TicketRepo) Resolve(_ context.Context, id uuid.UUID, resolvedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.Status = domain.TicketStatusResolved
	t.ResolvedAt = &resolvedAt
	t.UpdatedAt = time.Now()
	return nil
}

func (r *TicketRepo) SubmitCSAT(_ context.Context, id uuid.UUID, score int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.CSATScore = &score
	t.UpdatedAt = time.Now()
	return nil
}

func (r *TicketRepo) AddMessage(_ context.Context, msg *domain.TicketMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tickets[msg.TicketID]; !ok {
		return domain.ErrTicketNotFound
	}

	now := time.Now()
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}

	cp := *msg
	r.messages[msg.TicketID] = append(r.messages[msg.TicketID], cp)

	r.tickets[msg.TicketID].UpdatedAt = now
	return nil
}

func (r *TicketRepo) ListMessages(_ context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := r.messages[ticketID]
	result := make([]domain.TicketMessage, len(msgs))
	copy(result, msgs)

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (r *TicketRepo) CountByStatus(_ context.Context) (*domain.TicketStatusCounts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var counts domain.TicketStatusCounts
	for _, t := range r.tickets {
		switch t.Status {
		case domain.TicketStatusOpen:
			counts.Open++
		case domain.TicketStatusInProgress:
			counts.InProgress++
		case domain.TicketStatusEscalated:
			counts.Escalated++
		case domain.TicketStatusResolved:
			counts.Resolved++
		case domain.TicketStatusClosed:
			counts.Closed++
		}
	}
	return &counts, nil
}

func (r *TicketRepo) ListStaleTickets(_ context.Context, level domain.TicketLevel, olderThan time.Time) ([]domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Ticket
	for _, t := range r.tickets {
		if t.Level == level &&
			(t.Status == domain.TicketStatusOpen || t.Status == domain.TicketStatusInProgress) &&
			t.UpdatedAt.Before(olderThan) {
			cp := *t
			result = append(result, cp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (r *TicketRepo) ListResolvedForAutoClose(_ context.Context, resolvedBefore time.Time) ([]domain.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Ticket
	for _, t := range r.tickets {
		if t.Status == domain.TicketStatusResolved &&
			t.ResolvedAt != nil &&
			t.ResolvedAt.Before(resolvedBefore) {
			cp := *t
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *TicketRepo) GetOperationMetrics(_ context.Context, filter domain.TicketMetricsFilter) (*domain.TicketOperationMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalTickets, totalResolved, fcrCount int64
	var ahtSum float64
	var ahtCount int64
	var csatSum float64
	var csatCount int64
	var slaCompliant int64

	for _, t := range r.tickets {
		if filter.DateFrom != nil && t.CreatedAt.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && !t.CreatedAt.Before(*filter.DateTo) {
			continue
		}

		totalTickets++

		if t.Status == domain.TicketStatusResolved || t.Status == domain.TicketStatusClosed {
			totalResolved++
			if t.Level == domain.TicketLevelL1 {
				fcrCount++
			}
		}

		if t.ResolvedAt != nil {
			ahtSum += t.ResolvedAt.Sub(t.CreatedAt).Seconds()
			ahtCount++
		}

		if t.CSATScore != nil {
			csatSum += float64(*t.CSATScore)
			csatCount++
		}

		// SLA: check if first admin message was within 24h
		msgs := r.messages[t.ID]
		for _, m := range msgs {
			if m.SenderType == domain.TicketSenderAdmin {
				if m.CreatedAt.Sub(t.CreatedAt) <= 24*time.Hour {
					slaCompliant++
				}
				break
			}
		}
	}

	metrics := &domain.TicketOperationMetrics{
		TotalTickets:  totalTickets,
		TotalResolved: totalResolved,
	}

	if totalResolved > 0 {
		metrics.FCRPercent = float64(fcrCount) / float64(totalResolved) * 100
	}
	if ahtCount > 0 {
		metrics.AHTSeconds = ahtSum / float64(ahtCount)
	}
	if csatCount > 0 {
		metrics.AvgCSAT = csatSum / float64(csatCount)
	}
	if totalTickets > 0 {
		metrics.SLACompliancePercent = float64(slaCompliant) / float64(totalTickets) * 100
	}

	return metrics, nil
}

// BackdateUpdatedAt sets the updated_at time for a ticket (test helper).
func (r *TicketRepo) BackdateUpdatedAt(id uuid.UUID, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ticket, ok := r.tickets[id]; ok {
		ticket.UpdatedAt = t
	}
}

// BackdateResolvedAt sets the resolved_at time for a ticket (test helper).
func (r *TicketRepo) BackdateResolvedAt(id uuid.UUID, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ticket, ok := r.tickets[id]; ok {
		ticket.ResolvedAt = &t
	}
}
