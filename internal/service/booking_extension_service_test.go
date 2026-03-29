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

type mockExtWalletService struct {
	noopWalletService
	holds    map[uuid.UUID]*domain.WalletHold
	captured map[uuid.UUID]bool
	released map[uuid.UUID]bool
}

func newMockExtWalletService() *mockExtWalletService {
	return &mockExtWalletService{
		holds:    make(map[uuid.UUID]*domain.WalletHold),
		captured: make(map[uuid.UUID]bool),
		released: make(map[uuid.UUID]bool),
	}
}

func (m *mockExtWalletService) GetWallet(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	return &domain.Wallet{ID: uuid.New(), Balance: 100000}, nil
}

func (m *mockExtWalletService) Hold(_ context.Context, walletID uuid.UUID, amount int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	hold := &domain.WalletHold{
		ID:       uuid.New(),
		WalletID: walletID,
		Amount:   amount,
	}
	m.holds[hold.ID] = hold
	return hold, nil
}

func (m *mockExtWalletService) CaptureHold(_ context.Context, holdID uuid.UUID) (*domain.WalletTransaction, error) {
	m.captured[holdID] = true
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}

func (m *mockExtWalletService) ReleaseHold(_ context.Context, holdID uuid.UUID) error {
	m.released[holdID] = true
	return nil
}

func newExtensionService() (service.BookingExtensionService, service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.ExtensionRequestRepo, *mock.RepresentativeRepo, *mockExtWalletService) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	extReqRepo := mock.NewExtensionRequestRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	slotBlockRepo := mock.NewSlotBlockRepo()
	addonRepo := mock.NewAddOnRepo()
	userRepo := mock.NewUserRepo()
	cityRepo := mock.NewCityRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	walletSvc := newMockExtWalletService()
	bookingSvc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, userRepo, cityRepo, pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
	extSvc := service.NewBookingExtensionService(extReqRepo, bookingRepo, bhRepo, bookingSvc, walletSvc, access, &noopNotifService{}, log)
	return extSvc, bookingSvc, bhRepo, bookingRepo, extReqRepo, repRepo, walletSvc
}

func createConfirmedBooking(t *testing.T, _ service.BookingService, bhRepo *mock.BathhouseRepo, bookingRepo *mock.BookingRepo, ownerID, clientID uuid.UUID) (*domain.Booking, *domain.Bathhouse) {
	t.Helper()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	// Booking is happening right now (confirmed, active session)
	start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 0, 0, 0, now.Location())
	end := start.Add(3 * time.Hour)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
		TotalPrice:  bh.PricePerHour * 3,
		Status:      domain.BookingConfirmed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := bookingRepo.Create(context.Background(), booking)
	require.NoError(t, err)

	return booking, bh
}

func TestExtensionService_RequestExtension_Success(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, walletSvc := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, bh := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	req, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	require.NoError(t, err)

	assert.Equal(t, domain.ExtReqPending, req.Status)
	assert.Equal(t, booking.ID, req.BookingID)
	assert.Equal(t, clientID, req.UserID)
	assert.Equal(t, bh.ID, req.BathhouseID)
	assert.Equal(t, 1, req.ExtraHours)
	assert.Equal(t, bh.PricePerHour, req.ExtensionPrice)
	assert.Equal(t, booking.EndTime.Add(time.Hour), req.NewEndTime)
	assert.NotNil(t, req.HoldID)
	assert.False(t, req.ExpiresAt.IsZero())

	// Verify wallet hold was created
	assert.Len(t, walletSvc.holds, 1)
}

func TestExtensionService_RequestExtension_DuplicatePending(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, _ := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	_, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	require.NoError(t, err)

	// Second request should fail
	_, err = extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	assert.ErrorIs(t, err, domain.ErrExtensionRequestPending)
}

func TestExtensionService_RequestExtension_WrongUser(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, _ := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherUser := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	_, err := extSvc.RequestExtension(context.Background(), otherUser, booking.ID, 1)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestExtensionService_RequestExtension_InvalidHours(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, _ := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	_, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)

	_, err = extSvc.RequestExtension(context.Background(), clientID, booking.ID, 3)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestExtensionService_ApproveExtension_Success(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, walletSvc := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	req, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 2)
	require.NoError(t, err)

	result, err := extSvc.ApproveExtension(context.Background(), ownerID, domain.RoleOwner, req.ID)
	require.NoError(t, err)

	assert.Equal(t, booking.EndTime.Add(2*time.Hour), result.NewEndTime)
	assert.Equal(t, booking.TotalPrice+req.ExtensionPrice, result.NewTotalPrice)
	assert.Equal(t, req.ExtensionPrice, result.ExtensionPrice)

	// Verify hold was captured
	assert.True(t, walletSvc.captured[*req.HoldID])
}

func TestExtensionService_RejectExtension_Success(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, walletSvc := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	req, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	require.NoError(t, err)

	err = extSvc.RejectExtension(context.Background(), ownerID, domain.RoleOwner, req.ID, "слишком поздно")
	require.NoError(t, err)

	// Verify hold was released
	assert.True(t, walletSvc.released[*req.HoldID])

	// Verify request status
	updated, err := extSvc.GetByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ExtReqRejected, updated.Status)
	assert.Equal(t, "слишком поздно", updated.RejectionReason)
}

func TestExtensionService_RejectExtension_WrongOwner(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, _ := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherOwner := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	req, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	require.NoError(t, err)

	err = extSvc.RejectExtension(context.Background(), otherOwner, domain.RoleOwner, req.ID, "no")
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestExtensionService_ExpireTimedOutRequests(t *testing.T) {
	extSvc, _, bhRepo, bookingRepo, extReqRepo, _, walletSvc := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	// Create a confirmed booking
	bh := createBathhouse(t, bhRepo, ownerID)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 0, 0, 0, now.Location())
	end := start.Add(3 * time.Hour)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
		TotalPrice:  15000,
		Status:      domain.BookingConfirmed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := bookingRepo.Create(context.Background(), booking)
	require.NoError(t, err)

	// Create an already-expired extension request directly in the repo
	holdID := uuid.New()
	walletSvc.holds[holdID] = &domain.WalletHold{ID: holdID}
	expiredReq := &domain.BookingExtensionRequest{
		ID:             uuid.New(),
		BookingID:      booking.ID,
		UserID:         clientID,
		BathhouseID:    bh.ID,
		Status:         domain.ExtReqPending,
		ExtraHours:     1,
		ExtensionPrice: 5000,
		NewEndTime:     end.Add(time.Hour),
		HoldID:         &holdID,
		CreatedAt:      now.Add(-1 * time.Hour),
		ExpiresAt:      now.Add(-30 * time.Minute), // Already expired
	}
	err = extReqRepo.Create(context.Background(), expiredReq)
	require.NoError(t, err)

	count, err := extSvc.ExpireTimedOutRequests(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify hold was released
	assert.True(t, walletSvc.released[holdID])

	// Verify status
	updated, err := extSvc.GetByID(context.Background(), expiredReq.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ExtReqExpired, updated.Status)
}

func TestExtensionService_ListByBooking(t *testing.T) {
	extSvc, bookingSvc, bhRepo, bookingRepo, _, _, _ := newExtensionService()
	ownerID := uuid.New()
	clientID := uuid.New()

	booking, _ := createConfirmedBooking(t, bookingSvc, bhRepo, bookingRepo, ownerID, clientID)

	_, err := extSvc.RequestExtension(context.Background(), clientID, booking.ID, 1)
	require.NoError(t, err)

	requests, err := extSvc.ListByBooking(context.Background(), booking.ID)
	require.NoError(t, err)
	assert.Len(t, requests, 1)
}
