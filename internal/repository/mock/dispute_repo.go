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

type DisputeRepo struct {
	mu       sync.RWMutex
	disputes map[uuid.UUID]*domain.Dispute
	evidence map[uuid.UUID][]domain.DisputeEvidence
}

func NewDisputeRepo() repository.DisputeRepository {
	return &DisputeRepo{
		disputes: make(map[uuid.UUID]*domain.Dispute),
		evidence: make(map[uuid.UUID][]domain.DisputeEvidence),
	}
}

func (r *DisputeRepo) Create(_ context.Context, dispute *domain.Dispute) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check unique booking_id constraint
	for _, d := range r.disputes {
		if d.BookingID == dispute.BookingID {
			return domain.ErrDisputeAlreadyExists
		}
	}

	now := time.Now()
	if dispute.ID == uuid.Nil {
		dispute.ID = uuid.New()
	}
	if dispute.CreatedAt.IsZero() {
		dispute.CreatedAt = now
	}
	if dispute.UpdatedAt.IsZero() {
		dispute.UpdatedAt = now
	}

	cp := *dispute
	r.disputes[dispute.ID] = &cp
	return nil
}

func (r *DisputeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Dispute, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.disputes[id]
	if !ok {
		return nil, domain.ErrDisputeNotFound
	}
	cp := *d
	return &cp, nil
}

func (r *DisputeRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.Dispute, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, d := range r.disputes {
		if d.BookingID == bookingID {
			cp := *d
			return &cp, nil
		}
	}
	return nil, domain.ErrDisputeNotFound
}

func (r *DisputeRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.DisputeStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.disputes[id]
	if !ok {
		return domain.ErrDisputeNotFound
	}
	d.Status = status
	d.UpdatedAt = time.Now()
	return nil
}

func (r *DisputeRepo) UpdateResolution(_ context.Context, id uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string, resolvedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.disputes[id]
	if !ok {
		return domain.ErrDisputeNotFound
	}
	d.Status = domain.DisputeStatusResolved
	d.Resolution = &resolution
	d.RefundAmount = refundAmount
	d.CompensationAmount = compensationAmount
	d.MediatorNotes = mediatorNotes
	d.ResolvedAt = &resolvedAt
	appealDeadline := resolvedAt.Add(7 * 24 * time.Hour)
	d.AppealDeadline = &appealDeadline
	d.UpdatedAt = time.Now()
	return nil
}

func (r *DisputeRepo) UpdateAppeal(_ context.Context, id uuid.UUID, status domain.DisputeStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.disputes[id]
	if !ok {
		return domain.ErrDisputeNotFound
	}
	d.Status = status
	d.UpdatedAt = time.Now()
	return nil
}

func (r *DisputeRepo) Assign(_ context.Context, id uuid.UUID, mediatorID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.disputes[id]
	if !ok {
		return domain.ErrDisputeNotFound
	}
	d.MediatorID = &mediatorID
	d.Status = domain.DisputeStatusUnderReview
	d.UpdatedAt = time.Now()
	return nil
}

func (r *DisputeRepo) AddEvidence(_ context.Context, evidence *domain.DisputeEvidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.disputes[evidence.DisputeID]; !ok {
		return domain.ErrDisputeNotFound
	}

	now := time.Now()
	if evidence.ID == uuid.Nil {
		evidence.ID = uuid.New()
	}
	if evidence.CreatedAt.IsZero() {
		evidence.CreatedAt = now
	}

	cp := *evidence
	r.evidence[evidence.DisputeID] = append(r.evidence[evidence.DisputeID], cp)
	return nil
}

func (r *DisputeRepo) ListEvidence(_ context.Context, disputeID uuid.UUID) ([]domain.DisputeEvidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.evidence[disputeID]
	result := make([]domain.DisputeEvidence, len(items))
	copy(result, items)

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (r *DisputeRepo) ListAll(_ context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Dispute
	for _, d := range r.disputes {
		if filter.Status != nil && d.Status != *filter.Status {
			continue
		}
		if filter.UserID != nil && d.InitiatorID != *filter.UserID && d.RespondentID != *filter.UserID {
			continue
		}
		cp := *d
		filtered = append(filtered, cp)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *DisputeRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Dispute
	for _, d := range r.disputes {
		if d.InitiatorID == userID || d.RespondentID == userID {
			cp := *d
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *DisputeRepo) CountOpenByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, d := range r.disputes {
		if (d.InitiatorID == userID || d.RespondentID == userID) &&
			d.Status != domain.DisputeStatusClosed && d.Status != domain.DisputeStatusResolved {
			count++
		}
	}
	return count, nil
}
