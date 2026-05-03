package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// NotificationRepo is an in-memory mock implementation of repository.NotificationRepository.
type NotificationRepo struct {
	mu               sync.RWMutex
	notifications    map[uuid.UUID]*domain.Notification
	preferences      map[uuid.UUID]*domain.NotificationPreferences
	eventPrefs       map[string]*domain.NotificationEventPreference // key: "userID:eventType"
	pushDeliveryLogs map[uuid.UUID]*domain.PushDeliveryLog
}

func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{
		notifications:    make(map[uuid.UUID]*domain.Notification),
		preferences:      make(map[uuid.UUID]*domain.NotificationPreferences),
		eventPrefs:       make(map[string]*domain.NotificationEventPreference),
		pushDeliveryLogs: make(map[uuid.UUID]*domain.PushDeliveryLog),
	}
}

func (r *NotificationRepo) Create(_ context.Context, notification *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	if notification.Data == nil {
		notification.Data = make(map[string]string)
	}

	cp := *notification
	cpData := make(map[string]string, len(notification.Data))
	for k, v := range notification.Data {
		cpData[k] = v
	}
	cp.Data = cpData
	r.notifications[notification.ID] = &cp
	return nil
}

func (r *NotificationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.notifications[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	cpData := make(map[string]string, len(n.Data))
	for k, v := range n.Data {
		cpData[k] = v
	}
	cp.Data = cpData
	return &cp, nil
}

func (r *NotificationRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			items = append(items, *n)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *NotificationRepo) MarkAsRead(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	if n.IsRead {
		return domain.ErrNotFound
	}
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
	return nil
}

func (r *NotificationRepo) MarkAllAsRead(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, n := range r.notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			n.ReadAt = &now
		}
	}
	return nil
}

func (r *NotificationRepo) CountUnread(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, n := range r.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *NotificationRepo) GetPreferences(_ context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.preferences[userID]
	if !ok {
		defaults := domain.DefaultNotificationPreferences(userID)
		return &defaults, nil
	}
	cp := *p
	return &cp, nil
}

func (r *NotificationRepo) UpdatePreferences(_ context.Context, prefs *domain.NotificationPreferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *prefs
	r.preferences[prefs.UserID] = &cp
	return nil
}

func (r *NotificationRepo) HasRecentByType(_ context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, n := range r.notifications {
		if n.UserID == userID && n.Type == notifType && !n.CreatedAt.Before(since) {
			return true, nil
		}
	}
	return false, nil
}

func eventPrefKey(userID uuid.UUID, eventType domain.NotificationEventType) string {
	return userID.String() + ":" + string(eventType)
}

func (r *NotificationRepo) GetEventPreferences(_ context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var prefs []domain.NotificationEventPreference
	prefix := userID.String() + ":"
	for k, p := range r.eventPrefs {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			cp := *p
			prefs = append(prefs, cp)
		}
	}
	return prefs, nil
}

func (r *NotificationRepo) GetEventPreference(_ context.Context, userID uuid.UUID, eventType domain.NotificationEventType) (*domain.NotificationEventPreference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.eventPrefs[eventPrefKey(userID, eventType)]
	if !ok {
		def := domain.DefaultEventPreference(userID, eventType)
		return &def, nil
	}
	cp := *p
	return &cp, nil
}

func (r *NotificationRepo) UpsertEventPreference(_ context.Context, pref *domain.NotificationEventPreference) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *pref
	r.eventPrefs[eventPrefKey(pref.UserID, pref.EventType)] = &cp
	return nil
}

func (r *NotificationRepo) UpsertEventPreferences(_ context.Context, prefs []domain.NotificationEventPreference) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range prefs {
		cp := prefs[i]
		r.eventPrefs[eventPrefKey(cp.UserID, cp.EventType)] = &cp
	}
	return nil
}

func (r *NotificationRepo) CreatePushDeliveryLog(_ context.Context, log *domain.PushDeliveryLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	cp := *log
	r.pushDeliveryLogs[log.ID] = &cp
	return nil
}

func (r *NotificationRepo) GetPendingPushDeliveries(_ context.Context, olderThan time.Time) ([]domain.PushDeliveryLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var logs []domain.PushDeliveryLog
	for _, l := range r.pushDeliveryLogs {
		if l.Status == domain.PushStatusSent && !l.FallbackSent && l.SentAt.Before(olderThan) {
			logs = append(logs, *l)
		}
	}
	return logs, nil
}

func (r *NotificationRepo) UpdatePushDeliveryStatus(_ context.Context, id uuid.UUID, status domain.PushDeliveryStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.pushDeliveryLogs[id]
	if !ok {
		return domain.ErrNotFound
	}
	l.Status = status
	return nil
}

func (r *NotificationRepo) MarkPushFallbackSent(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.pushDeliveryLogs[id]
	if !ok {
		return domain.ErrNotFound
	}
	l.FallbackSent = true
	return nil
}
