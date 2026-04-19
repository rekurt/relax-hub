package mock

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ListingDraftRepo struct {
	mu     sync.RWMutex
	drafts map[uuid.UUID]*domain.ListingDraft
}

func NewListingDraftRepo() *ListingDraftRepo {
	return &ListingDraftRepo{
		drafts: make(map[uuid.UUID]*domain.ListingDraft),
	}
}

var _ repository.ListingDraftRepository = (*ListingDraftRepo)(nil)

func (r *ListingDraftRepo) Create(_ context.Context, draft *domain.ListingDraft) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if draft.ID == uuid.Nil {
		draft.ID = uuid.New()
	}
	now := time.Now()
	if draft.CreatedAt.IsZero() {
		draft.CreatedAt = now
	}
	if draft.UpdatedAt.IsZero() {
		draft.UpdatedAt = now
	}

	cp := *draft
	cp.StepData = copyStepData(draft.StepData)
	r.drafts[draft.ID] = &cp
	return nil
}

func (r *ListingDraftRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ListingDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.drafts[id]
	if !ok {
		return nil, domain.ErrListingDraftNotFound
	}
	cp := *d
	cp.StepData = copyStepData(d.StepData)
	return &cp, nil
}

func (r *ListingDraftRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]domain.ListingDraft, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ListingDraft
	for _, d := range r.drafts {
		if d.UserID == userID {
			cp := *d
			cp.StepData = copyStepData(d.StepData)
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *ListingDraftRepo) UpdateStep(_ context.Context, id uuid.UUID, step int, data json.RawMessage, currentStep int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.drafts[id]
	if !ok || d.Status != domain.ListingDraftStatusDraft {
		return domain.ErrListingDraftNotFound
	}

	if d.StepData == nil {
		d.StepData = make(map[int]json.RawMessage)
	}
	d.StepData[step] = append(json.RawMessage(nil), data...)
	d.CurrentStep = currentStep
	d.UpdatedAt = time.Now()
	return nil
}

func (r *ListingDraftRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ListingDraftStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.drafts[id]
	if !ok {
		return domain.ErrListingDraftNotFound
	}
	d.Status = status
	d.UpdatedAt = time.Now()
	return nil
}

func (r *ListingDraftRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.drafts[id]; !ok {
		return domain.ErrListingDraftNotFound
	}
	delete(r.drafts, id)
	return nil
}

func copyStepData(src map[int]json.RawMessage) map[int]json.RawMessage {
	if src == nil {
		return nil
	}
	cp := make(map[int]json.RawMessage, len(src))
	for k, v := range src {
		cp[k] = append(json.RawMessage(nil), v...)
	}
	return cp
}
