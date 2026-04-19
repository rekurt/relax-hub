package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

// BathhouseRepo is an in-memory mock implementation of repository.BathhouseRepository.
type BathhouseRepo struct {
	mu         sync.RWMutex
	bathhouses map[uuid.UUID]*domain.Bathhouse
}

func NewBathhouseRepo() *BathhouseRepo {
	return &BathhouseRepo{bathhouses: make(map[uuid.UUID]*domain.Bathhouse)}
}

func (r *BathhouseRepo) Create(_ context.Context, bh *domain.Bathhouse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bh.ID == uuid.Nil {
		bh.ID = uuid.New()
	}
	now := time.Now()
	bh.CreatedAt = now
	bh.UpdatedAt = now
	cp := *bh
	r.bathhouses[bh.ID] = &cp
	return nil
}

func (r *BathhouseRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *bh
	return &cp, nil
}

func (r *BathhouseRepo) GetByAPIKey(_ context.Context, apiKey string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.ApiKey == apiKey {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) GetBySlug(_ context.Context, slug string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.Slug == slug {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) SlugExists(_ context.Context, slug string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (r *BathhouseRepo) Update(_ context.Context, bh *domain.Bathhouse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bathhouses[bh.ID]; !ok {
		return domain.ErrNotFound
	}
	bh.UpdatedAt = time.Now()
	cp := *bh
	r.bathhouses[bh.ID] = &cp
	return nil
}

func (r *BathhouseRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bathhouses[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.bathhouses, id)
	return nil
}

func (r *BathhouseRepo) List(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if filter.Status != nil {
			if bh.Status != *filter.Status {
				continue
			}
		} else if !filter.ShowAllStatuses && bh.Status != domain.BathhouseStatusActive {
			continue
		}
		if filter.CityID != nil && bh.CityID != *filter.CityID {
			continue
		}
		if filter.PriceMin != nil && bh.PricePerHour < *filter.PriceMin {
			continue
		}
		if filter.PriceMax != nil && bh.PricePerHour > *filter.PriceMax {
			continue
		}
		if filter.GuestCount != nil && bh.MaxGuests < *filter.GuestCount {
			continue
		}
		if filter.SearchQuery != nil && *filter.SearchQuery != "" {
			q := strings.ToLower(*filter.SearchQuery)
			name := strings.ToLower(bh.Name)
			desc := strings.ToLower(bh.Description)
			addr := strings.ToLower(bh.Address)
			// Exact substring match
			matched := strings.Contains(name, q) || strings.Contains(desc, q) || strings.Contains(addr, q)
			// Prefix match: check if any word in name/description starts with any query word
			if !matched {
				queryWords := strings.Fields(q)
				for _, qw := range queryWords {
					for _, field := range []string{name, desc, addr} {
						for _, fw := range strings.Fields(field) {
							if strings.HasPrefix(fw, qw) {
								matched = true
								break
							}
						}
						if matched {
							break
						}
					}
					if matched {
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		if filter.OpenNow != nil && *filter.OpenNow {
			if !isBathhouseOpenNow(bh) {
				continue
			}
		}
		items = append(items, *bh)
	}

	return paginate(items, filter.Page, filter.PageSize), nil
}

func (r *BathhouseRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if bh.OwnerID == ownerID {
			items = append(items, *bh)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *BathhouseRepo) UpdateRating(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *BathhouseRepo) UpdateBayesianRating(_ context.Context, id uuid.UUID, bayesianRating float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.BayesianRating = bayesianRating
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.BathhouseStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.Status = status
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) UpdatePhotoVerified(_ context.Context, id uuid.UUID, verified bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.IsPhotoVerified = verified
	bh.UpdatedAt = time.Now()
	return nil
}

func (r *BathhouseRepo) GetCalendarToken(_ context.Context, bathhouseID uuid.UUID) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bh, ok := r.bathhouses[bathhouseID]
	if !ok {
		return "", domain.ErrNotFound
	}
	return bh.CalendarToken, nil
}

func (r *BathhouseRepo) SetCalendarToken(_ context.Context, bathhouseID uuid.UUID, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[bathhouseID]
	if !ok {
		return domain.ErrNotFound
	}
	bh.CalendarToken = token
	return nil
}

func (r *BathhouseRepo) GetByCalendarToken(_ context.Context, token string) (*domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bh := range r.bathhouses {
		if bh.CalendarToken == token {
			cp := *bh
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *BathhouseRepo) SuggestNames(_ context.Context, filter repository.SuggestionFilter) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(filter.Query)
	var names []string
	for _, bh := range r.bathhouses {
		if bh.Status == domain.BathhouseStatusActive && strings.Contains(strings.ToLower(bh.Name), q) {
			names = append(names, bh.Name)
			if len(names) >= filter.Limit {
				break
			}
		}
	}
	return names, nil
}

func (r *BathhouseRepo) IncrementViewCount(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ViewCount++
	return nil
}

func (r *BathhouseRepo) UpdateRankingFields(_ context.Context, id uuid.UUID, conversionRate, occupancyRate float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ConversionRate = conversionRate
	bh.OccupancyRate = occupancyRate
	return nil
}

func (r *BathhouseRepo) ListIDsByOwner(_ context.Context, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ids []uuid.UUID
	for _, bh := range r.bathhouses {
		if bh.OwnerID == ownerID {
			ids = append(ids, bh.ID)
		}
	}
	return ids, nil
}

func (r *BathhouseRepo) UpdateResponseRate(_ context.Context, id uuid.UUID, responseRate float64, avgResponseMinutes int, lowResponseRateSince *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh, ok := r.bathhouses[id]
	if !ok {
		return domain.ErrNotFound
	}
	bh.ResponseRate = responseRate
	bh.AvgResponseTimeMinutes = avgResponseMinutes
	bh.LowResponseRateSince = lowResponseRateSince
	return nil
}

func (r *BathhouseRepo) ListRequestModeBathhouses(_ context.Context) ([]domain.Bathhouse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Bathhouse
	for _, bh := range r.bathhouses {
		if bh.BookingMode == domain.BookingModeRequest {
			cp := *bh
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *BathhouseRepo) GetAreaAvgPrice(_ context.Context, cityID int64, _, _ float64) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var total int64
	var count int64
	for _, bh := range r.bathhouses {
		if bh.CityID == cityID && bh.Status == domain.BathhouseStatusActive {
			total += bh.PricePerHour
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / count, nil
}

func isBathhouseOpenNow(bh *domain.Bathhouse) bool {
	now := time.Now()
	d := now.Weekday()
	dayOfWeek := int(d) - 1
	if d == time.Sunday {
		dayOfWeek = 6
	}
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	for _, wh := range bh.WorkingHours {
		if wh.DayOfWeek != dayOfWeek {
			continue
		}
		if wh.OpenTime <= wh.CloseTime {
			// Normal schedule: e.g. 09:00 - 22:00
			if wh.OpenTime <= currentTime && wh.CloseTime > currentTime {
				return true
			}
		} else {
			// Overnight schedule: e.g. 20:00 - 06:00
			if wh.OpenTime <= currentTime || wh.CloseTime > currentTime {
				return true
			}
		}
	}
	return false
}


