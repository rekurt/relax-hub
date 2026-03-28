package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFMNotificationService implements service.NotificationService for testing
type mockFMNotificationService struct {
	notifications []fmNotifCall
}

type fmNotifCall struct {
	UserID uuid.UUID
	Type   domain.NotificationType
	Title  string
}

func (m *mockFMNotificationService) Send(_ context.Context, userID uuid.UUID, notifType domain.NotificationType, title, _ string, _ map[string]string) error {
	m.notifications = append(m.notifications, fmNotifCall{
		UserID: userID,
		Type:   notifType,
		Title:  title,
	})
	return nil
}

func (m *mockFMNotificationService) List(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Notification], error) {
	return &domain.PaginatedResult[domain.Notification]{}, nil
}

func (m *mockFMNotificationService) MarkAsRead(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockFMNotificationService) MarkAllAsRead(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockFMNotificationService) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockFMNotificationService) GetPreferences(_ context.Context, _ uuid.UUID) (*domain.NotificationPreferences, error) {
	return nil, nil
}

func (m *mockFMNotificationService) UpdatePreferences(_ context.Context, _ uuid.UUID, _ *domain.NotificationPreferences) error {
	return nil
}

func (m *mockFMNotificationService) HasRecentByType(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _ time.Time) (bool, error) {
	return false, nil
}

func (m *mockFMNotificationService) GetEventPreferences(_ context.Context, _ uuid.UUID) ([]domain.NotificationEventPreference, error) {
	return nil, nil
}

func (m *mockFMNotificationService) UpdateEventPreferences(_ context.Context, _ uuid.UUID, _ []domain.NotificationEventPreference) error {
	return nil
}

// mockFMWalletService implements service.WalletService for testing
type mockFMWalletService struct {
	wallets     map[uuid.UUID]*domain.Wallet
	refundCalls []fmRefundCall
}

type fmRefundCall struct {
	WalletID uuid.UUID
	Amount   int64
	RefType  string
}

func newMockFMWalletService() *mockFMWalletService {
	return &mockFMWalletService{
		wallets: make(map[uuid.UUID]*domain.Wallet),
	}
}

func (m *mockFMWalletService) CreateWallet(_ context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
	w := &domain.Wallet{ID: uuid.New(), UserID: userID, Currency: currency}
	m.wallets[w.ID] = w
	return w, nil
}

func (m *mockFMWalletService) GetWallet(_ context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	for _, w := range m.wallets {
		if w.UserID == userID {
			return w, nil
		}
	}
	return nil, domain.ErrWalletNotFound
}

func (m *mockFMWalletService) TopUp(_ context.Context, _ uuid.UUID, _ int64) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockFMWalletService) Spend(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockFMWalletService) Hold(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	return &domain.WalletHold{}, nil
}

func (m *mockFMWalletService) CaptureHold(_ context.Context, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockFMWalletService) ReleaseHold(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockFMWalletService) Refund(_ context.Context, walletID uuid.UUID, amount int64, refType string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	m.refundCalls = append(m.refundCalls, fmRefundCall{
		WalletID: walletID, Amount: amount, RefType: refType,
	})
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}

func (m *mockFMWalletService) AddBonus(_ context.Context, _ uuid.UUID, _ int64, _ domain.WalletTransactionType, _ *time.Time, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockFMWalletService) GetBalance(_ context.Context, _ uuid.UUID) (*service.WalletBalanceSummary, error) {
	return &service.WalletBalanceSummary{}, nil
}

func (m *mockFMWalletService) ListTransactions(_ context.Context, _ uuid.UUID, _ domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	return &domain.PaginatedResult[domain.WalletTransaction]{}, nil
}

func (m *mockFMWalletService) GetActiveHolds(_ context.Context, _ uuid.UUID) ([]domain.WalletHold, error) {
	return nil, nil
}

func (m *mockFMWalletService) ExpireBonuses(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockFMWalletService) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockFMWalletService) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockFMWalletService) AdminCredit(_ context.Context, _ uuid.UUID, _ int64, _ string, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *mockFMWalletService) AdminDebit(_ context.Context, _ uuid.UUID, _ int64, _ string, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *mockFMWalletService) AdminFreeze(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockFMWalletService) AdminUnfreeze(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockFMWalletService) GetWalletByID(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	return nil, nil
}

func setupForceMajeureTest(t *testing.T) (service.ForceMajeureService, *mock.ForceMajeureRepo, *mock.BookingRepo, *mock.BathhouseRepo, *mock.CityRepo, *mockFMWalletService, *mockFMNotificationService) {
	t.Helper()

	fmRepo := mock.NewForceMajeureRepo()
	bookingRepo := mock.NewBookingRepo()
	bathhouseRepo := mock.NewBathhouseRepo()
	cityRepo := mock.NewCityRepo()
	walletSvc := newMockFMWalletService()
	notifSvc := &mockFMNotificationService{}
	log := logger.New(logger.LevelWarn)

	svc := service.NewForceMajeureService(fmRepo, bookingRepo, bathhouseRepo, cityRepo, walletSvc, notifSvc, log)
	return svc, fmRepo, bookingRepo, bathhouseRepo, cityRepo, walletSvc, notifSvc
}

func TestForceMajeure_Activate_Success(t *testing.T) {
	svc, fmRepo, bookingRepo, bathhouseRepo, cityRepo, walletSvc, notifSvc := setupForceMajeureTest(t)
	ctx := context.Background()

	adminID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	// Setup city with region
	require.NoError(t, cityRepo.Create(ctx, &domain.City{Name: "Москва", Slug: "moscow", Region: "Москва"}))

	// Setup bathhouse with owner
	require.NoError(t, bathhouseRepo.Create(ctx, &domain.Bathhouse{ID: bathhouseID, OwnerID: ownerID, Name: "Test Banya", Slug: "test-banya"}))

	// Setup region mapping
	bookingRepo.SetBathhouseRegion(bathhouseID, "Москва")

	// Create wallets for users
	wallet1, _ := walletSvc.CreateWallet(ctx, userID1, domain.WalletCurrencyRUB)
	wallet2, _ := walletSvc.CreateWallet(ctx, userID2, domain.WalletCurrencyRUB)

	// Create confirmed bookings
	dateFrom := time.Now().Add(24 * time.Hour)
	dateTo := time.Now().Add(48 * time.Hour)

	booking1 := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID1,
		BathhouseID: bathhouseID,
		StartTime:   dateFrom.Add(2 * time.Hour),
		EndTime:     dateFrom.Add(4 * time.Hour),
		TotalPrice:  500000, // 5000 RUB
		GuestCount:  2,
		Status:      domain.BookingConfirmed,
	}
	booking2 := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID2,
		BathhouseID: bathhouseID,
		StartTime:   dateFrom.Add(6 * time.Hour),
		EndTime:     dateFrom.Add(8 * time.Hour),
		TotalPrice:  300000, // 3000 RUB
		GuestCount:  1,
		Status:      domain.BookingConfirmed,
	}

	require.NoError(t, bookingRepo.Create(ctx, booking1))
	require.NoError(t, bookingRepo.Create(ctx, booking2))

	event, err := svc.Activate(ctx, adminID, "Москва", dateFrom, dateTo, "Наводнение в регионе")
	require.NoError(t, err)
	require.NotNil(t, event)

	// Verify event data
	assert.Equal(t, adminID, event.AdminID)
	assert.Equal(t, "Москва", event.Region)
	assert.Equal(t, "Наводнение в регионе", event.Reason)
	assert.Equal(t, 2, event.AffectedCount)
	assert.Equal(t, int64(800000), event.TotalRefund) // 5000 + 3000 = 8000 RUB

	// Verify bookings were cancelled with force_majeure status
	b1, _ := bookingRepo.GetByID(ctx, booking1.ID)
	assert.Equal(t, domain.BookingForceMajeure, b1.Status)
	b2, _ := bookingRepo.GetByID(ctx, booking2.ID)
	assert.Equal(t, domain.BookingForceMajeure, b2.Status)

	// Verify refunds were issued
	assert.Len(t, walletSvc.refundCalls, 2)
	refundAmounts := map[uuid.UUID]int64{}
	for _, r := range walletSvc.refundCalls {
		refundAmounts[r.WalletID] = r.Amount
		assert.Equal(t, "force_majeure", r.RefType)
	}
	assert.Equal(t, int64(500000), refundAmounts[wallet1.ID])
	assert.Equal(t, int64(300000), refundAmounts[wallet2.ID])

	// Verify notifications: 2 client + 1 owner (deduplicated)
	assert.Len(t, notifSvc.notifications, 3)
	ownerNotifs := 0
	for _, n := range notifSvc.notifications {
		assert.Equal(t, domain.NotifSystem, n.Type)
		if n.UserID == ownerID {
			ownerNotifs++
		}
	}
	assert.Equal(t, 1, ownerNotifs, "owner should receive exactly 1 deduplicated notification")

	// Verify event was persisted
	events, err := fmRepo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event.ID, events[0].ID)
}

func TestForceMajeure_Activate_NoBookings(t *testing.T) {
	svc, fmRepo, _, _, cityRepo, _, _ := setupForceMajeureTest(t)
	ctx := context.Background()

	adminID := uuid.New()
	dateFrom := time.Now().Add(24 * time.Hour)
	dateTo := time.Now().Add(48 * time.Hour)

	// Region must exist in cities
	require.NoError(t, cityRepo.Create(ctx, &domain.City{Name: "Новосибирск", Slug: "novosibirsk", Region: "Сибирь"}))

	event, err := svc.Activate(ctx, adminID, "Сибирь", dateFrom, dateTo, "Мороз")
	require.NoError(t, err)
	assert.Equal(t, 0, event.AffectedCount)
	assert.Equal(t, int64(0), event.TotalRefund)

	// Event should still be saved
	events, err := fmRepo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestForceMajeure_Activate_InvalidInput(t *testing.T) {
	svc, _, _, _, cityRepo, _, _ := setupForceMajeureTest(t)
	ctx := context.Background()
	adminID := uuid.New()
	now := time.Now()

	// Setup city for region validation tests
	require.NoError(t, cityRepo.Create(ctx, &domain.City{Name: "Москва", Slug: "moscow", Region: "Москва"}))

	// Empty region
	_, err := svc.Activate(ctx, adminID, "", now, now.Add(time.Hour), "reason")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)

	// Empty reason
	_, err = svc.Activate(ctx, adminID, "Москва", now, now.Add(time.Hour), "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)

	// dateTo before dateFrom
	_, err = svc.Activate(ctx, adminID, "Москва", now.Add(time.Hour), now, "reason")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)

	// Non-existent region
	_, err = svc.Activate(ctx, adminID, "НесуществующийРегион", now, now.Add(time.Hour), "reason")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestForceMajeure_Activate_SkipsNonMatchingRegion(t *testing.T) {
	svc, _, bookingRepo, _, cityRepo, _, _ := setupForceMajeureTest(t)
	ctx := context.Background()

	adminID := uuid.New()
	bathhouseID := uuid.New()
	userID := uuid.New()

	// Setup cities for both regions
	require.NoError(t, cityRepo.Create(ctx, &domain.City{Name: "Москва", Slug: "moscow", Region: "Москва"}))
	require.NoError(t, cityRepo.Create(ctx, &domain.City{Name: "СПб", Slug: "spb", Region: "СПб"}))

	// Bathhouse is in "СПб", not "Москва"
	bookingRepo.SetBathhouseRegion(bathhouseID, "СПб")

	dateFrom := time.Now().Add(24 * time.Hour)
	dateTo := time.Now().Add(48 * time.Hour)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: bathhouseID,
		StartTime:   dateFrom.Add(2 * time.Hour),
		EndTime:     dateFrom.Add(4 * time.Hour),
		TotalPrice:  500000,
		GuestCount:  2,
		Status:      domain.BookingConfirmed,
	}
	require.NoError(t, bookingRepo.Create(ctx, booking))

	event, err := svc.Activate(ctx, adminID, "Москва", dateFrom, dateTo, "Наводнение")
	require.NoError(t, err)
	assert.Equal(t, 0, event.AffectedCount)

	// Booking should remain confirmed
	b, _ := bookingRepo.GetByID(ctx, booking.ID)
	assert.Equal(t, domain.BookingConfirmed, b.Status)
}

func TestForceMajeure_List(t *testing.T) {
	svc, fmRepo, _, _, _, _, _ := setupForceMajeureTest(t)
	ctx := context.Background()

	// Create multiple events
	_ = fmRepo.Create(ctx, &domain.ForceMajeureEvent{
		AdminID: uuid.New(), Region: "Москва", Reason: "Event 1",
		DateFrom: time.Now(), DateTo: time.Now().Add(time.Hour),
	})
	_ = fmRepo.Create(ctx, &domain.ForceMajeureEvent{
		AdminID: uuid.New(), Region: "СПб", Reason: "Event 2",
		DateFrom: time.Now(), DateTo: time.Now().Add(time.Hour),
	})

	events, err := svc.List(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 2)
	// Newest first
	assert.Equal(t, "Event 2", events[0].Reason)
	assert.Equal(t, "Event 1", events[1].Reason)
}
