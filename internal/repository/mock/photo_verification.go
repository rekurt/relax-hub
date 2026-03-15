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

type BathhousePhotoRepo struct {
	mu     sync.RWMutex
	photos map[uuid.UUID]*domain.BathhousePhoto
}

func NewBathhousePhotoRepo() *BathhousePhotoRepo {
	return &BathhousePhotoRepo{
		photos: make(map[uuid.UUID]*domain.BathhousePhoto),
	}
}

// Ensure interface compliance.
var _ repository.BathhousePhotoRepository = (*BathhousePhotoRepo)(nil)

func (r *BathhousePhotoRepo) Create(_ context.Context, photo *domain.BathhousePhoto) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if photo.ID == uuid.Nil {
		photo.ID = uuid.New()
	}
	if photo.Status == "" {
		photo.Status = domain.PhotoStatusPending
	}
	if photo.UploadedAt.IsZero() {
		photo.UploadedAt = time.Now()
	}

	cp := *photo
	r.photos[photo.ID] = &cp
	return nil
}

func (r *BathhousePhotoRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.BathhousePhoto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.photos[id]
	if !ok {
		return nil, domain.ErrPhotoNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *BathhousePhotoRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.photos[id]; !ok {
		return domain.ErrPhotoNotFound
	}
	delete(r.photos, id)
	return nil
}

func (r *BathhousePhotoRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.BathhousePhoto
	for _, p := range r.photos {
		if p.BathhouseID == bathhouseID {
			cp := *p
			result = append(result, cp)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})
	return result, nil
}

func (r *BathhousePhotoRepo) ListVerifiedByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.BathhousePhoto
	for _, p := range r.photos {
		if p.BathhouseID == bathhouseID && p.Status == domain.PhotoStatusVerified {
			cp := *p
			result = append(result, cp)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})
	return result, nil
}

func (r *BathhousePhotoRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.PhotoStatus, verifiedByID *uuid.UUID, rejectionReason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.photos[id]
	if !ok {
		return domain.ErrPhotoNotFound
	}

	if p.Status != domain.PhotoStatusPending {
		return domain.ErrPhotoNotFound
	}

	p.Status = status
	p.VerifiedByID = verifiedByID
	p.RejectionReason = rejectionReason
	if status == domain.PhotoStatusVerified || status == domain.PhotoStatusRejected {
		now := time.Now()
		p.VerifiedAt = &now
	}
	return nil
}

func (r *BathhousePhotoRepo) Reorder(_ context.Context, bathhouseID uuid.UUID, photoIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, photoID := range photoIDs {
		p, ok := r.photos[photoID]
		if !ok || p.BathhouseID != bathhouseID {
			return domain.ErrPhotoNotFound
		}
		p.Position = i
	}
	return nil
}

func (r *BathhousePhotoRepo) ListPending(_ context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pending []domain.BathhousePhoto
	for _, p := range r.photos {
		if p.Status == domain.PhotoStatusPending {
			cp := *p
			pending = append(pending, cp)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].UploadedAt.Before(pending[j].UploadedAt)
	})

	return paginate(pending, page, pageSize), nil
}
