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

func (r *GuestCardRepo) matchesFilter(c *domain.GuestCard, filter domain.GuestCardFilter) bool {
	if filter.NoOwnerFilter {
		return true
	}
	if len(filter.BathhouseIDs) > 0 {
		for _, id := range filter.BathhouseIDs {
			if c.BathhouseID == id {
				return true
			}
		}
		return false
	}
	return c.OwnerID == filter.OwnerID
}

func (r *GuestCardRepo) CountBySegment(_ context.Context, filter domain.GuestCardFilter, segment domain.GuestSegmentSlug) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, c := range r.cards {
		if !r.matchesFilter(c, filter) {
			continue
		}
		if matchesSegment(c, segment) {
			count++
		}
	}
	return count, nil
}

func (r *GuestCardRepo) GetRFMScores(_ context.Context, filter domain.GuestCardFilter) (*domain.RFMResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.GuestCard
	for _, c := range r.cards {
		if !r.matchesFilter(c, filter) {
			continue
		}
		cp := *c
		filtered = append(filtered, cp)
	}

	if len(filtered) == 0 {
		return &domain.RFMResult{
			Guests: []domain.GuestRFM{},
			Matrix: []domain.RFMMatrixCell{},
		}, nil
	}

	n := len(filtered)
	bucketSize := n / 5
	if bucketSize < 1 {
		bucketSize = 1
	}

	// Sort by last_visit_at ascending for recency
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].LastVisitAt.Before(filtered[j].LastVisitAt) })
	rRanks := make(map[uuid.UUID]int, n)
	for i, c := range filtered {
		score := i/bucketSize + 1
		if score > 5 {
			score = 5
		}
		rRanks[c.ID] = score
	}

	// Sort by visit_count ascending
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].VisitCount < filtered[j].VisitCount })
	fRanks := make(map[uuid.UUID]int, n)
	for i, c := range filtered {
		score := i/bucketSize + 1
		if score > 5 {
			score = 5
		}
		fRanks[c.ID] = score
	}

	// Sort by total_spent ascending
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].TotalSpent < filtered[j].TotalSpent })
	mRanks := make(map[uuid.UUID]int, n)
	for i, c := range filtered {
		score := i/bucketSize + 1
		if score > 5 {
			score = 5
		}
		mRanks[c.ID] = score
	}

	var guests []domain.GuestRFM
	matrixMap := make(map[[2]int]int64)
	for _, c := range filtered {
		rfm := domain.RFMScore{
			Recency:   rRanks[c.ID],
			Frequency: fRanks[c.ID],
			Monetary:  mRanks[c.ID],
		}
		guests = append(guests, domain.GuestRFM{GuestCard: c, RFM: rfm})
		key := [2]int{rfm.Recency, rfm.Frequency}
		matrixMap[key]++
	}

	var matrix []domain.RFMMatrixCell
	for key, count := range matrixMap {
		matrix = append(matrix, domain.RFMMatrixCell{Recency: key[0], Frequency: key[1], Count: count})
	}

	return &domain.RFMResult{Guests: guests, Matrix: matrix}, nil
}

func (r *GuestCardRepo) GetStats(_ context.Context, filter domain.GuestCardFilter) (*domain.GuestCardStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var stats domain.GuestCardStats
	var totalVisitCount int64
	var totalSpent int64
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	for _, c := range r.cards {
		if !r.matchesFilter(c, filter) {
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
