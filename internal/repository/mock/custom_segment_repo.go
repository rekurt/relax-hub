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

type CustomSegmentRepo struct {
	mu       sync.RWMutex
	segments map[uuid.UUID]*domain.CustomSegment
	gcRepo   *GuestCardRepo
}

func NewCustomSegmentRepo(gcRepo *GuestCardRepo) repository.CustomSegmentRepository {
	return &CustomSegmentRepo{
		segments: make(map[uuid.UUID]*domain.CustomSegment),
		gcRepo:   gcRepo,
	}
}

func (r *CustomSegmentRepo) Create(_ context.Context, segment *domain.CustomSegment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if segment.ID == uuid.Nil {
		segment.ID = uuid.New()
	}
	segment.CreatedAt = now
	segment.UpdatedAt = now

	cp := *segment
	r.segments[segment.ID] = &cp
	return nil
}

func (r *CustomSegmentRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.CustomSegment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.segments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *CustomSegmentRepo) Update(_ context.Context, segment *domain.CustomSegment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.segments[segment.ID]
	if !ok {
		return domain.ErrNotFound
	}
	segment.UpdatedAt = time.Now()
	cp := *segment
	r.segments[segment.ID] = &cp
	return nil
}

func (r *CustomSegmentRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.segments[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.segments, id)
	return nil
}

func (r *CustomSegmentRepo) ListByOwner(_ context.Context, filter domain.CustomSegmentFilter) ([]domain.CustomSegment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.CustomSegment
	for _, s := range r.segments {
		if s.OwnerID != filter.OwnerID {
			continue
		}
		if filter.BathhouseID != nil && s.BathhouseID != nil && *s.BathhouseID != *filter.BathhouseID {
			continue
		}
		cp := *s
		result = append(result, cp)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (r *CustomSegmentRepo) EvaluateSegment(_ context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter, page, pageSize int) (*domain.PaginatedResult[domain.GuestCard], error) {
	r.gcRepo.mu.RLock()
	defer r.gcRepo.mu.RUnlock()

	var matched []domain.GuestCard
	for _, c := range r.gcRepo.cards {
		if !r.gcRepo.matchesFilter(c, ownerFilter) {
			continue
		}
		if matchesCustomConditions(c, segment) {
			cp := *c
			matched = append(matched, cp)
		}
	}

	sort.Slice(matched, func(i, j int) bool { return matched[i].LastVisitAt.After(matched[j].LastVisitAt) })
	return paginate(matched, page, pageSize), nil
}

func (r *CustomSegmentRepo) CountSegmentGuests(_ context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter) (int64, error) {
	r.gcRepo.mu.RLock()
	defer r.gcRepo.mu.RUnlock()

	var count int64
	for _, c := range r.gcRepo.cards {
		if !r.gcRepo.matchesFilter(c, ownerFilter) {
			continue
		}
		if matchesCustomConditions(c, segment) {
			count++
		}
	}
	return count, nil
}

func matchesCustomConditions(c *domain.GuestCard, segment *domain.CustomSegment) bool {
	cond := segment.Conditions

	if segment.BathhouseID != nil && c.BathhouseID != *segment.BathhouseID {
		return false
	}
	if cond.VisitCountMin != nil && c.VisitCount < *cond.VisitCountMin {
		return false
	}
	if cond.VisitCountMax != nil && c.VisitCount > *cond.VisitCountMax {
		return false
	}
	if cond.AvgCheckMin != nil && c.AvgCheck < *cond.AvgCheckMin {
		return false
	}
	if cond.AvgCheckMax != nil && c.AvgCheck > *cond.AvgCheckMax {
		return false
	}
	if cond.TotalSpentMin != nil && c.TotalSpent < *cond.TotalSpentMin {
		return false
	}
	if cond.TotalSpentMax != nil && c.TotalSpent > *cond.TotalSpentMax {
		return false
	}
	if cond.LastVisitDaysMin != nil {
		daysSince := int(time.Since(c.LastVisitAt).Hours() / 24)
		if daysSince < *cond.LastVisitDaysMin {
			return false
		}
	}
	if cond.LastVisitDaysMax != nil {
		daysSince := int(time.Since(c.LastVisitAt).Hours() / 24)
		if daysSince > *cond.LastVisitDaysMax {
			return false
		}
	}
	if len(cond.TagsInclude) > 0 {
		tagSet := make(map[string]bool, len(c.Tags))
		for _, t := range c.Tags {
			tagSet[t] = true
		}
		for _, t := range cond.TagsInclude {
			if !tagSet[t] {
				return false
			}
		}
	}
	if len(cond.TagsExclude) > 0 {
		tagSet := make(map[string]bool, len(c.Tags))
		for _, t := range c.Tags {
			tagSet[t] = true
		}
		for _, t := range cond.TagsExclude {
			if tagSet[t] {
				return false
			}
		}
	}
	return true
}
