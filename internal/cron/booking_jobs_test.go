package cron

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// mockBookingServiceForReminders implements service.BookingService for cron tests
type mockBookingServiceForReminders struct {
	upcomingBookings []service.UpcomingBookingInfo
	upcomingErr      error
}

func (m *mockBookingServiceForReminders) Create(_ context.Context, _ uuid.UUID, _ service.CreateBookingInput) (*service.BookingResult, error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) Cancel(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockBookingServiceForReminders) Confirm(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBookingServiceForReminders) Reject(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockBookingServiceForReminders) Approve(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBookingServiceForReminders) Complete(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.BookingResult, error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Booking], error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) ListByBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Booking], error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) GetAvailableSlots(_ context.Context, _ uuid.UUID, _ time.Time) ([]service.TimeSlot, error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) AutoRejectTimedOutRequests(_ context.Context) (int, error) {
	return 0, nil
}
func (m *mockBookingServiceForReminders) CheckIn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBookingServiceForReminders) CheckOut(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockBookingServiceForReminders) MarkNoShows(_ context.Context) (int, error) {
	return 0, nil
}
func (m *mockBookingServiceForReminders) DisputeNoShow(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _ float64, _ string) error {
	return nil
}
func (m *mockBookingServiceForReminders) ListUpcomingWithBathhouse(_ context.Context, _, _ time.Time) ([]service.UpcomingBookingInfo, error) {
	return m.upcomingBookings, m.upcomingErr
}
func (m *mockBookingServiceForReminders) Extend(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) (*service.ExtendResult, error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) GetRebookData(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*service.RebookData, error) {
	return nil, nil
}
func (m *mockBookingServiceForReminders) RecalculateResponseRates(_ context.Context) (int, error) {
	return 0, nil
}

func newTestReminderScheduler(notifSvc *mockNotificationService, bookingSvc service.BookingService, redisClient *redis.Client) *CronScheduler {
	log := logger.New(logger.LevelInfo)
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()
	return NewCronScheduler(log, mockAnalyticsSvc, mockAnalyticsRepo, nil, nil, notifSvc, nil, nil, nil, nil, nil, nil, bookingSvc, nil, nil, nil, nil, redisClient)
}

func TestHandleBookingReminders_24hReminder(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bookingSvc := &mockBookingServiceForReminders{
		upcomingBookings: []service.UpcomingBookingInfo{
			{
				Booking: domain.Booking{
					ID:          bookingID,
					UserID:      userID,
					BathhouseID: bathhouseID,
					StartTime:   now.Add(24 * time.Hour),
					EndTime:     now.Add(26 * time.Hour),
					Status:      domain.BookingConfirmed,
				},
				BathhouseName: "Баня у Петра",
				Address:       "ул. Ленина 10",
				Latitude:      55.7,
				Longitude:     37.6,
				OwnerID:       ownerID,
			},
		},
	}
	notifSvc := &mockNotificationService{}

	cs := newTestReminderScheduler(notifSvc, bookingSvc, nil)
	cs.handleBookingReminders()

	// Should have sent at least the 24h reminder
	found := false
	for _, n := range notifSvc.sent {
		if n.NotifType == domain.NotifBookingReminder24h && n.UserID == userID {
			found = true
			assert.Equal(t, "Напоминание о бронировании", n.Title)
		}
	}
	assert.True(t, found, "Expected 24h reminder to be sent to client")
}

func TestHandleBookingReminders_2hReminder(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bookingSvc := &mockBookingServiceForReminders{
		upcomingBookings: []service.UpcomingBookingInfo{
			{
				Booking: domain.Booking{
					ID:          bookingID,
					UserID:      userID,
					BathhouseID: bathhouseID,
					StartTime:   now.Add(2 * time.Hour),
					EndTime:     now.Add(4 * time.Hour),
					Status:      domain.BookingConfirmed,
				},
				BathhouseName: "Баня Люкс",
				Address:       "ул. Пушкина 5",
				Latitude:      55.8,
				Longitude:     37.7,
				OwnerID:       ownerID,
			},
		},
	}
	notifSvc := &mockNotificationService{}

	cs := newTestReminderScheduler(notifSvc, bookingSvc, nil)
	cs.handleBookingReminders()

	found := false
	for _, n := range notifSvc.sent {
		if n.NotifType == domain.NotifBookingReminder2h && n.UserID == userID {
			found = true
			assert.Equal(t, "Скоро бронирование", n.Title)
		}
	}
	assert.True(t, found, "Expected 2h reminder to be sent to client")
}

func TestHandleBookingReminders_Owner5minReminder(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bookingSvc := &mockBookingServiceForReminders{
		upcomingBookings: []service.UpcomingBookingInfo{
			{
				Booking: domain.Booking{
					ID:          bookingID,
					UserID:      userID,
					BathhouseID: bathhouseID,
					StartTime:   now.Add(3 * time.Minute),
					EndTime:     now.Add(2 * time.Hour),
					Status:      domain.BookingConfirmed,
				},
				BathhouseName: "Баня Релакс",
				Address:       "ул. Мира 1",
				Latitude:      55.9,
				Longitude:     37.5,
				OwnerID:       ownerID,
			},
		},
	}
	notifSvc := &mockNotificationService{}

	cs := newTestReminderScheduler(notifSvc, bookingSvc, nil)
	cs.handleBookingReminders()

	found := false
	for _, n := range notifSvc.sent {
		if n.NotifType == domain.NotifBookingReminderOwner5min && n.UserID == ownerID {
			found = true
			assert.Equal(t, "Скоро прибытие гостя", n.Title)
		}
	}
	assert.True(t, found, "Expected owner 5min reminder to be sent")
}

func TestHandleBookingReminders_Deduplication(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	now := time.Now()
	bookingID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bookingSvc := &mockBookingServiceForReminders{
		upcomingBookings: []service.UpcomingBookingInfo{
			{
				Booking: domain.Booking{
					ID:          bookingID,
					UserID:      userID,
					BathhouseID: bathhouseID,
					StartTime:   now.Add(24 * time.Hour),
					EndTime:     now.Add(26 * time.Hour),
					Status:      domain.BookingConfirmed,
				},
				BathhouseName: "Баня Тест",
				Address:       "ул. Тестовая 1",
				Latitude:      55.7,
				Longitude:     37.6,
				OwnerID:       ownerID,
			},
		},
	}
	notifSvc := &mockNotificationService{}

	cs := newTestReminderScheduler(notifSvc, bookingSvc, redisClient)

	// First run — should send
	cs.handleBookingReminders()
	sentCount := len(notifSvc.sent)
	assert.Greater(t, sentCount, 0, "First run should send reminders")

	// Second run — same booking, should be deduplicated
	cs.handleBookingReminders()
	assert.Equal(t, sentCount, len(notifSvc.sent), "Second run should not send duplicate reminders")
}

func TestHandleBookingReminders_NoCancelledBookings(t *testing.T) {
	// Empty booking list — no reminders sent
	bookingSvc := &mockBookingServiceForReminders{
		upcomingBookings: nil,
	}
	notifSvc := &mockNotificationService{}

	cs := newTestReminderScheduler(notifSvc, bookingSvc, nil)
	cs.handleBookingReminders()

	assert.Empty(t, notifSvc.sent, "No reminders should be sent for empty booking list")
}
