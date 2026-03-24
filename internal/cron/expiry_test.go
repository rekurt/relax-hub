package cron

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotificationService implements service.NotificationService for testing
type mockNotificationService struct {
	sent []sentNotification
}

type sentNotification struct {
	UserID    uuid.UUID
	NotifType domain.NotificationType
	Title     string
}

func (m *mockNotificationService) Send(_ context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error {
	m.sent = append(m.sent, sentNotification{UserID: userID, NotifType: notifType, Title: title})
	return nil
}

func (m *mockNotificationService) List(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Notification], error) {
	return nil, nil
}

func (m *mockNotificationService) MarkAsRead(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockNotificationService) MarkAllAsRead(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockNotificationService) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockNotificationService) GetPreferences(_ context.Context, _ uuid.UUID) (*domain.NotificationPreferences, error) {
	return nil, nil
}

func (m *mockNotificationService) UpdatePreferences(_ context.Context, _ uuid.UUID, _ *domain.NotificationPreferences) error {
	return nil
}

func TestHandleSubscriptionExpiryNotify(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	subRepo := mock.NewSubscriptionRepo()
	promoRepo := mock.NewPromoCodeRepo()
	notifSvc := &mockNotificationService{}
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()

	ownerID := uuid.New()
	bhID := uuid.New()

	// Create a subscription expiring in 2 days (should be notified)
	expiringDate := time.Now().Add(2 * 24 * time.Hour)
	err := subRepo.Create(context.Background(), &domain.Subscription{
		ID:          uuid.New(),
		BathhouseID: bhID,
		OwnerID:     ownerID,
		Plan:        domain.PlanPremium,
		Status:      domain.SubscriptionActive,
		StartDate:   time.Now().Add(-28 * 24 * time.Hour),
		EndDate:     &expiringDate,
	})
	require.NoError(t, err)

	// Create an already expired subscription (should NOT be notified)
	expiredDate := time.Now().Add(-1 * 24 * time.Hour)
	err = subRepo.Create(context.Background(), &domain.Subscription{
		ID:          uuid.New(),
		BathhouseID: uuid.New(),
		OwnerID:     uuid.New(),
		Plan:        domain.PlanPremium,
		Status:      domain.SubscriptionActive,
		StartDate:   time.Now().Add(-30 * 24 * time.Hour),
		EndDate:     &expiredDate,
	})
	require.NoError(t, err)

	cs := NewCronScheduler(log, mockAnalyticsSvc, mockAnalyticsRepo, subRepo, promoRepo, notifSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	cs.handleSubscriptionExpiryNotify()

	// Only the expiring (not expired) subscription should be notified
	assert.Equal(t, 1, len(notifSvc.sent))
	assert.Equal(t, ownerID, notifSvc.sent[0].UserID)
	assert.Equal(t, domain.NotifSubscriptionExpiring, notifSvc.sent[0].NotifType)
}

func TestHandleExpiredSubscriptionUpdate(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	subRepo := mock.NewSubscriptionRepo()
	promoRepo := mock.NewPromoCodeRepo()
	notifSvc := &mockNotificationService{}
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()

	ownerID := uuid.New()
	bhID := uuid.New()
	subID := uuid.New()

	expiredDate := time.Now().Add(-1 * 24 * time.Hour)
	err := subRepo.Create(context.Background(), &domain.Subscription{
		ID:          subID,
		BathhouseID: bhID,
		OwnerID:     ownerID,
		Plan:        domain.PlanPremium,
		Status:      domain.SubscriptionActive,
		StartDate:   time.Now().Add(-30 * 24 * time.Hour),
		EndDate:     &expiredDate,
	})
	require.NoError(t, err)

	cs := NewCronScheduler(log, mockAnalyticsSvc, mockAnalyticsRepo, subRepo, promoRepo, notifSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	cs.handleExpiredSubscriptionUpdate()

	// Verify status was updated
	sub, err := subRepo.GetByID(context.Background(), subID)
	require.NoError(t, err)
	assert.Equal(t, domain.SubscriptionExpired, sub.Status)

	// Verify notification was sent
	assert.Equal(t, 1, len(notifSvc.sent))
	assert.Equal(t, domain.NotifSubscriptionExpired, notifSvc.sent[0].NotifType)
}

func TestHandlePromoDeactivation(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	subRepo := mock.NewSubscriptionRepo()
	promoRepo := mock.NewPromoCodeRepo()
	notifSvc := &mockNotificationService{}
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()

	// Create an expired promo code
	expiredTime := time.Now().Add(-24 * time.Hour)
	err := promoRepo.Create(context.Background(), &domain.PromoCode{
		ID:         uuid.New(),
		Code:       "EXPIRED1",
		Type:       domain.PromoTypePercentage,
		Value:      10,
		CreatorID:  uuid.New(),
		ValidFrom:  time.Now().Add(-30 * 24 * time.Hour),
		ValidUntil: expiredTime,
		IsActive:   true,
	})
	require.NoError(t, err)

	// Create an active promo code (not expired)
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	activePromoID := uuid.New()
	err = promoRepo.Create(context.Background(), &domain.PromoCode{
		ID:         activePromoID,
		Code:       "ACTIVE1",
		Type:       domain.PromoTypePercentage,
		Value:      15,
		CreatorID:  uuid.New(),
		ValidFrom:  time.Now().Add(-10 * 24 * time.Hour),
		ValidUntil: futureTime,
		IsActive:   true,
	})
	require.NoError(t, err)

	cs := NewCronScheduler(log, mockAnalyticsSvc, mockAnalyticsRepo, subRepo, promoRepo, notifSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	cs.handlePromoDeactivation()

	// Active promo should still be active
	activePromo, err := promoRepo.GetByID(context.Background(), activePromoID)
	require.NoError(t, err)
	assert.True(t, activePromo.IsActive)
}

func TestFormatDays(t *testing.T) {
	assert.Equal(t, "менее суток", formatDays(0))
	assert.Equal(t, "1 день", formatDays(1))
	assert.Equal(t, "2 дня", formatDays(2))
	assert.Equal(t, "3 дня", formatDays(3))
	assert.Equal(t, "5 дней", formatDays(5))
}
