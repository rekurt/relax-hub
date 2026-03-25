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

type GuestCardRepo struct {
	mu    sync.RWMutex
	cards map[uuid.UUID]*domain.GuestCard
}

func NewGuestCardRepo() repository.GuestCardRepository {
	return &GuestCardRepo{
		cards: make(map[uuid.UUID]*domain.GuestCard),
	}
}

func (r *GuestCardRepo) Upsert(_ context.Context, card *domain.GuestCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}

	// Check for existing card with same owner+client+bathhouse
	for _, existing := range r.cards {
		if existing.OwnerID == card.OwnerID &&
			existing.ClientID == card.ClientID &&
			existing.BathhouseID == card.BathhouseID {
			// Update existing
			existing.LastVisitAt = card.LastVisitAt
			existing.VisitCount++
			existing.TotalSpent += card.TotalSpent
			if existing.VisitCount > 0 {
				existing.AvgCheck = existing.TotalSpent / int64(existing.VisitCount)
			}
			existing.UpdatedAt = now
			return nil
		}
	}

	// Create new
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	if card.UpdatedAt.IsZero() {
		card.UpdatedAt = now
	}
	if card.FirstVisitAt.IsZero() {
		card.FirstVisitAt = now
	}
	if card.LastVisitAt.IsZero() {
		card.LastVisitAt = now
	}
	if card.Tags == nil {
		card.Tags = []string{}
	}

	cp := *card
	r.cards[card.ID] = &cp
	return nil
}

func (r *GuestCardRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.GuestCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.cards[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *GuestCardRepo) GetByOwnerAndClient(_ context.Context, ownerID, clientID, bathhouseID uuid.UUID) (*domain.GuestCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.cards {
		if c.OwnerID == ownerID && c.ClientID == clientID && c.BathhouseID == bathhouseID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func matchesSegment(c *domain.GuestCard, seg domain.GuestSegmentSlug) bool {
	switch seg {
	case domain.SegmentNew:
		return c.VisitCount == 1
	case domain.SegmentRegular:
		return c.VisitCount >= 3
	case domain.SegmentLost:
		return time.Since(c.LastVisitAt) > 90*24*time.Hour
	case domain.SegmentVIP:
		return c.TotalSpent > 5000000 // 50,000 RUB in kopecks
	case domain.SegmentBirthdaySoon:
		return false // no birthday data available
	default:
		return false
	}
}

func (r *GuestCardRepo) ListByOwner(_ context.Context, filter domain.GuestCardFilter) (*domain.PaginatedResult[domain.GuestCard], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.GuestCard
	for _, c := range r.cards {
		if c.OwnerID != filter.OwnerID {
			continue
		}
		if filter.BathhouseID != nil && c.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Segment != nil && !matchesSegment(c, *filter.Segment) {
			continue
		}
		if filter.Tag != nil && *filter.Tag != "" {
			found := false
			for _, t := range c.Tags {
				if t == *filter.Tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if filter.DateFrom != nil && c.LastVisitAt.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && c.LastVisitAt.After(*filter.DateTo) {
			continue
		}
		cp := *c
		filtered = append(filtered, cp)
	}

	// Sort
	switch filter.SortBy {
	case "total_spent":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].TotalSpent > filtered[j].TotalSpent })
	case "visit_count":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].VisitCount > filtered[j].VisitCount })
	case "avg_check":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].AvgCheck > filtered[j].AvgCheck })
	default:
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].LastVisitAt.After(filtered[j].LastVisitAt) })
	}

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *GuestCardRepo) UpdateNotes(_ context.Context, id uuid.UUID, notes string, tags []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cards[id]
	if !ok {
		return domain.ErrNotFound
	}
	c.Notes = notes
	c.Tags = tags
	c.UpdatedAt = time.Now()
	return nil
}

func (r *GuestCardRepo) CountBySegment(_ context.Context, ownerID uuid.UUID, segment domain.GuestSegmentSlug) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, c := range r.cards {
		if c.OwnerID != ownerID {
			continue
		}
		if matchesSegment(c, segment) {
			count++
		}
	}
	return count, nil
}

func (r *GuestCardRepo) GetStats(_ context.Context, ownerID uuid.UUID) (*domain.GuestCardStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var stats domain.GuestCardStats
	var totalVisitCount int64
	var totalSpent int64
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	for _, c := range r.cards {
		if c.OwnerID != ownerID {
			continue
		}
		stats.TotalGuests++
		totalVisitCount += int64(c.VisitCount)
		totalSpent += c.TotalSpent
		if c.CreatedAt.After(monthStart) || c.CreatedAt.Equal(monthStart) {
			stats.NewThisMonth++
		}
	}

	if stats.TotalGuests > 0 {
		stats.AvgVisitCount = totalVisitCount / stats.TotalGuests
		stats.AvgSpent = totalSpent / stats.TotalGuests
	}

	return &stats, nil
}

