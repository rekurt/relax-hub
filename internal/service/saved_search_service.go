package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	recentlyViewedKeyPrefix = "recently_viewed:"
	recentlyViewedMaxItems  = 20
	savedSearchMaxPerUser   = 50
)

// SavedSearchService handles recently viewed and saved search operations.
type SavedSearchService interface {
	// Recently viewed
	RecordView(ctx context.Context, userID, bathhouseID uuid.UUID) error
	ListRecentlyViewed(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)

	// Saved searches
	SaveSearch(ctx context.Context, userID uuid.UUID, name string, filters json.RawMessage, notifyOnNew bool) (*domain.SavedSearch, error)
	ListSavedSearches(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error)
	DeleteSavedSearch(ctx context.Context, userID uuid.UUID, searchID uuid.UUID) error
	ListWithNotifications(ctx context.Context) ([]domain.SavedSearch, error)
	CheckNewMatches(ctx context.Context) error
}

type savedSearchService struct {
	savedSearchRepo repository.SavedSearchRepository
	bhRepo          repository.BathhouseRepository
	notifSvc        NotificationService
	redis           *redis.Client
	logger          *logger.Logger
}

func NewSavedSearchService(
	savedSearchRepo repository.SavedSearchRepository,
	bhRepo repository.BathhouseRepository,
	notifSvc NotificationService,
	redisClient *redis.Client,
	log *logger.Logger,
) SavedSearchService {
	return &savedSearchService{
		savedSearchRepo: savedSearchRepo,
		bhRepo:          bhRepo,
		notifSvc:        notifSvc,
		redis:           redisClient,
		logger:          log,
	}
}

func (s *savedSearchService) RecordView(ctx context.Context, userID, bathhouseID uuid.UUID) error {
	key := recentlyViewedKeyPrefix + userID.String()
	score := float64(time.Now().Unix())

	pipe := s.redis.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: bathhouseID.String()})
	pipe.ZRemRangeByRank(ctx, key, 0, int64(-recentlyViewedMaxItems-1))
	pipe.Expire(ctx, key, 30*24*time.Hour)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("record recently viewed: %w", err)
	}
	return nil
}

func (s *savedSearchService) ListRecentlyViewed(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit <= 0 || limit > recentlyViewedMaxItems {
		limit = recentlyViewedMaxItems
	}

	key := recentlyViewedKeyPrefix + userID.String()
	members, err := s.redis.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("list recently viewed: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		id, err := uuid.Parse(m)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *savedSearchService) SaveSearch(ctx context.Context, userID uuid.UUID, name string, filters json.RawMessage, notifyOnNew bool) (*domain.SavedSearch, error) {
	result, err := s.savedSearchRepo.ListByUser(ctx, userID, 1, 1)
	if err != nil {
		return nil, err
	}
	if result.TotalCount >= int64(savedSearchMaxPerUser) {
		return nil, domain.ErrSavedSearchLimitReached
	}

	search := &domain.SavedSearch{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Filters:     filters,
		NotifyOnNew: notifyOnNew,
		CreatedAt:   time.Now(),
	}

	if err := s.savedSearchRepo.Create(ctx, search); err != nil {
		return nil, err
	}
	return search, nil
}

func (s *savedSearchService) ListSavedSearches(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error) {
	return s.savedSearchRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *savedSearchService) DeleteSavedSearch(ctx context.Context, userID uuid.UUID, searchID uuid.UUID) error {
	search, err := s.savedSearchRepo.GetByID(ctx, searchID)
	if err != nil {
		return err
	}
	if search.UserID != userID {
		return domain.ErrForbidden
	}
	return s.savedSearchRepo.Delete(ctx, searchID)
}

func (s *savedSearchService) ListWithNotifications(ctx context.Context) ([]domain.SavedSearch, error) {
	return s.savedSearchRepo.ListWithNotifications(ctx)
}

func (s *savedSearchService) CheckNewMatches(ctx context.Context) error {
	searches, err := s.savedSearchRepo.ListWithNotifications(ctx)
	if err != nil {
		return fmt.Errorf("list saved searches with notifications: %w", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1)

	for _, search := range searches {
		var filter domain.BathhouseFilter
		if err := json.Unmarshal(search.Filters, &filter); err != nil {
			s.logger.Error("failed to unmarshal saved search filters", "search_id", search.ID, "error", err)
			continue
		}

		filter.Page = 1
		filter.PageSize = 5
		activeStatus := domain.BathhouseStatusActive
		filter.Status = &activeStatus

		result, err := s.bhRepo.List(ctx, filter)
		if err != nil {
			s.logger.Error("failed to check saved search matches", "search_id", search.ID, "error", err)
			continue
		}

		newCount := 0
		for _, bh := range result.Items {
			if bh.CreatedAt.After(yesterday) {
				newCount++
			}
		}

		if newCount > 0 {
			searchName := search.Name
			if searchName == "" {
				searchName = "Сохранённый поиск"
			}
			notifMsg := fmt.Sprintf("По запросу «%s» найдено %d новых бань", searchName, newCount)
			if s.notifSvc != nil {
				_ = s.notifSvc.Send(ctx, search.UserID, domain.NotifSavedSearchMatch, "Новые совпадения", notifMsg, nil)
			}
		}
	}

	return nil
}
