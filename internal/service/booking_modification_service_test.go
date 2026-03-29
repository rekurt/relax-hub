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

func newModificationService() (service.BookingModificationService, service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.BookingModificationRequestRepo, *mock.RepresentativeRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	modReqRepo := mock.NewBookingModificationRequestRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	slotBlockRepo := mock.NewSlotBlockRepo()
	addonRepo := mock.NewAddOnRepo()
	userRepo := mock.NewUserRepo()
	cityRepo := mock.NewCityRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	bookingSvc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, userRepo, cityRepo, pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
	modSvc := service.NewBookingModificationService(modReqRepo, bookingRepo, bhRepo, bookingSvc, access, &noopNotifService{}, log)
	return modSvc, bookingSvc, bhRepo, bookingRepo, modReqRepo, repRepo
}

func TestModificationService_RequestModification_Success(t *testing.T) {
	modSvc, bookingSvc, bhRepo, _, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	newStart := time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location())
	newEnd := newStart.Add(3 * time.Hour)

	req, err := modSvc.RequestModification(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 4,
	})
	require.NoError(t, err)

	assert.Equal(t, domain.ModReqPending, req.Status)
	assert.Equal(t, result.Booking.ID, req.BookingID)
	assert.Equal(t, clientID, req.UserID)
	assert.Equal(t, start, req.OldStartTime)
	assert.Equal(t, end, req.OldEndTime)
	assert.Equal(t, 3, req.OldGuestCount)
	assert.Equal(t, newStart, req.ProposedStartTime)
	assert.Equal(t, newEnd, req.ProposedEndTime)
	assert.Equal(t, 4, req.ProposedGuestCount)
	assert.False(t, req.ExpiresAt.IsZero())
}

func TestModificationService_RequestModification_DuplicatePending(t *testing.T) {
	modSvc, bookingSvc, bhRepo, _, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	newStart := time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location())
	newEnd := newStart.Add(2 * time.Hour)

	_, err = modSvc.RequestModification(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 4,
	})
	require.NoError(t, err)

	// Second request should fail
	_, err = modSvc.RequestModification(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 5,
	})
	assert.ErrorIs(t, err, domain.ErrModificationRequestPending)
}

func TestModificationService_RequestModification_WrongUser(t *testing.T) {
	modSvc, bookingSvc, bhRepo, _, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherUser := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	_, err = modSvc.RequestModification(context.Background(), otherUser, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  start.Add(24 * time.Hour),
		EndTime:    end.Add(24 * time.Hour),
		GuestCount: 3,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestModificationService_ApproveModification_Success(t *testing.T) {
	modSvc, bookingSvc, bhRepo, bookingRepo, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	createResult, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	newStart := time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location())
	newEnd := newStart.Add(3 * time.Hour)

	req, err := modSvc.RequestModification(context.Background(), clientID, createResult.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 4,
	})
	require.NoError(t, err)

	// Owner approves
	modResult, err := modSvc.ApproveModification(context.Background(), ownerID, domain.RoleOwner, req.ID)
	require.NoError(t, err)

	assert.NotNil(t, modResult)
	assert.Equal(t, newStart, modResult.Booking.StartTime)
	assert.Equal(t, newEnd, modResult.Booking.EndTime)
	assert.Equal(t, 4, modResult.Booking.GuestCount)

	// Verify booking was updated
	updated, _ := bookingRepo.GetByID(context.Background(), createResult.Booking.ID)
	assert.Equal(t, 1, updated.ModificationCount)
}

func TestModificationService_RejectModification_Success(t *testing.T) {
	modSvc, bookingSvc, bhRepo, bookingRepo, modReqRepo, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	createResult, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	req, err := modSvc.RequestModification(context.Background(), clientID, createResult.Booking.ID, service.ModifyBookingInput{
		StartTime:  time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location()),
		EndTime:    time.Date(now.Year(), now.Month(), now.Day()+3, 16, 0, 0, 0, now.Location()),
		GuestCount: 4,
	})
	require.NoError(t, err)

	err = modSvc.RejectModification(context.Background(), ownerID, domain.RoleOwner, req.ID, "Нет свободных мест")
	require.NoError(t, err)

	// Verify request is rejected
	rejected, _ := modReqRepo.GetByID(context.Background(), req.ID)
	assert.Equal(t, domain.ModReqRejected, rejected.Status)
	assert.Equal(t, "Нет свободных мест", rejected.RejectionReason)

	// Booking should remain unchanged
	unchanged, _ := bookingRepo.GetByID(context.Background(), createResult.Booking.ID)
	assert.Equal(t, start, unchanged.StartTime)
	assert.Equal(t, 0, unchanged.ModificationCount)
}

func TestModificationService_ApproveModification_WrongRole(t *testing.T) {
	modSvc, bookingSvc, bhRepo, _, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherOwner := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	createResult, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	req, err := modSvc.RequestModification(context.Background(), clientID, createResult.Booking.ID, service.ModifyBookingInput{
		StartTime:  time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location()),
		EndTime:    time.Date(now.Year(), now.Month(), now.Day()+3, 16, 0, 0, 0, now.Location()),
		GuestCount: 4,
	})
	require.NoError(t, err)

	// Different owner tries to approve - should be forbidden
	_, err = modSvc.ApproveModification(context.Background(), otherOwner, domain.RoleOwner, req.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestModificationService_ExpireTimedOutRequests(t *testing.T) {
	modSvc, _, bhRepo, bookingRepo, modReqRepo, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+5, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
		TotalPrice:  10000,
		Status:      domain.BookingConfirmed,
	}
	require.NoError(t, bookingRepo.Create(context.Background(), booking))

	// Create an already-expired request directly
	expiredReq := &domain.BookingModificationRequest{
		ID:                 uuid.New(),
		BookingID:          booking.ID,
		UserID:             clientID,
		BathhouseID:        bh.ID,
		Status:             domain.ModReqPending,
		OldStartTime:       start,
		OldEndTime:         end,
		OldGuestCount:      3,
		OldTotalPrice:      10000,
		ProposedStartTime:  start.Add(24 * time.Hour),
		ProposedEndTime:    end.Add(24 * time.Hour),
		ProposedGuestCount: 4,
		ProposedTotalPrice: 15000,
		CreatedAt:          now.Add(-25 * time.Hour),
		ExpiresAt:          now.Add(-1 * time.Hour), // Already expired
	}
	require.NoError(t, modReqRepo.Create(context.Background(), expiredReq))

	count, err := modSvc.ExpireTimedOutRequests(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify it's now expired
	updated, err := modReqRepo.GetByID(context.Background(), expiredReq.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ModReqExpired, updated.Status)
}

func TestModificationService_ListByBooking(t *testing.T) {
	modSvc, bookingSvc, bhRepo, _, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	createResult, err := bookingSvc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	require.NoError(t, err)

	_, err = modSvc.RequestModification(context.Background(), clientID, createResult.Booking.ID, service.ModifyBookingInput{
		StartTime:  time.Date(now.Year(), now.Month(), now.Day()+3, 14, 0, 0, 0, now.Location()),
		EndTime:    time.Date(now.Year(), now.Month(), now.Day()+3, 16, 0, 0, 0, now.Location()),
		GuestCount: 4,
	})
	require.NoError(t, err)

	list, err := modSvc.ListByBooking(context.Background(), createResult.Booking.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestModificationService_CompletedBookingNotModifiable(t *testing.T) {
	modSvc, _, bhRepo, bookingRepo, _, _ := newModificationService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
		TotalPrice:  10000,
		Status:      domain.BookingCompleted,
	}
	require.NoError(t, bookingRepo.Create(context.Background(), booking))

	_, err := modSvc.RequestModification(context.Background(), clientID, booking.ID, service.ModifyBookingInput{
		StartTime:  start.Add(24 * time.Hour),
		EndTime:    end.Add(24 * time.Hour),
		GuestCount: 3,
	})
	assert.ErrorIs(t, err, domain.ErrBookingNotModifiable)
}
