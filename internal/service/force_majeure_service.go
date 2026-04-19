package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ForceMajeureService interface {
	Activate(ctx context.Context, adminID uuid.UUID, region string, dateFrom, dateTo time.Time, reason string) (*domain.ForceMajeureEvent, error)
	List(ctx context.Context) ([]domain.ForceMajeureEvent, error)
}

type forceMajeureService struct {
	fmRepo        repository.ForceMajeureRepository
	bookingRepo   repository.BookingRepository
	bathhouseRepo repository.BathhouseRepository
	cityRepo      repository.CityRepository
	walletSvc     WalletService
	notifSvc      NotificationService
	log           *logger.Logger
}

func NewForceMajeureService(
	fmRepo repository.ForceMajeureRepository,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	cityRepo repository.CityRepository,
	walletSvc WalletService,
	notifSvc NotificationService,
	log *logger.Logger,
) ForceMajeureService {
	return &forceMajeureService{
		fmRepo:        fmRepo,
		bookingRepo:   bookingRepo,
		bathhouseRepo: bathhouseRepo,
		cityRepo:      cityRepo,
		walletSvc:     walletSvc,
		notifSvc:      notifSvc,
		log:           log,
	}
}

func (s *forceMajeureService) Activate(ctx context.Context, adminID uuid.UUID, region string, dateFrom, dateTo time.Time, reason string) (*domain.ForceMajeureEvent, error) {
	if region == "" {
		return nil, domain.ErrInvalidInput
	}
	if reason == "" {
		return nil, domain.ErrInvalidInput
	}
	if !dateTo.After(dateFrom) {
		return nil, domain.ErrInvalidInput
	}

	// Validate that region exists in the cities table
	cities, err := s.cityRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("validate region: %w", err)
	}
	regionExists := false
	for _, c := range cities {
		if c.Region == region {
			regionExists = true
			break
		}
	}
	if !regionExists {
		return nil, domain.ErrInvalidInput
	}

	bookings, err := s.bookingRepo.ListConfirmedByRegionAndDateRange(ctx, region, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("list bookings for force majeure: %w", err)
	}

	var totalRefund int64
	affectedCount := 0
	notifiedOwners := make(map[uuid.UUID]bool)

	for _, booking := range bookings {
		if err := s.bookingRepo.UpdateStatus(ctx, booking.ID, domain.BookingForceMajeure); err != nil {
			s.log.Error("force majeure: failed to cancel booking", "booking_id", booking.ID, "error", err)
			continue
		}

		// 100% refund to wallet
		wallet, err := s.walletSvc.GetWallet(ctx, booking.UserID)
		if err != nil {
			s.log.Error("force majeure: failed to get wallet", "user_id", booking.UserID, "error", err)
		} else {
			bookingID := booking.ID
			_, err = s.walletSvc.Refund(ctx, wallet.ID, booking.TotalPrice, "force_majeure", &bookingID,
				fmt.Sprintf("Возврат по форс-мажору: %s", reason))
			if err != nil {
				s.log.Error("force majeure: failed to refund", "user_id", booking.UserID, "amount", booking.TotalPrice, "error", err)
			} else {
				totalRefund += booking.TotalPrice
			}
		}

		// Notify client
		_ = s.notifSvc.Send(ctx, booking.UserID, domain.NotifSystem,
			"Бронирование отменено по форс-мажору",
			fmt.Sprintf("Ваше бронирование отменено по причине: %s. Средства возвращены на кошелёк.", reason),
			map[string]string{"booking_id": booking.ID.String()})

		// Notify bathhouse owner (deduplicated below)
		if bh, bhErr := s.bathhouseRepo.GetByID(ctx, booking.BathhouseID); bhErr == nil {
			notifiedOwners[bh.OwnerID] = true
		}

		affectedCount++
	}

	// Send owner notifications (deduplicated)
	for ownerID := range notifiedOwners {
		_ = s.notifSvc.Send(ctx, ownerID, domain.NotifSystem,
			"Бронирования отменены по форс-мажору",
			fmt.Sprintf("Бронирования в регионе %q отменены по причине: %s", region, reason),
			map[string]string{"region": region})
	}

	event := &domain.ForceMajeureEvent{
		AdminID:       adminID,
		Region:        region,
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		Reason:        reason,
		AffectedCount: affectedCount,
		TotalRefund:   totalRefund,
	}

	if err := s.fmRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("save force majeure event: %w", err)
	}

	s.log.Info("force majeure activated",
		"region", region,
		"date_from", dateFrom,
		"date_to", dateTo,
		"affected_count", affectedCount,
		"total_refund", totalRefund)

	return event, nil
}

func (s *forceMajeureService) List(ctx context.Context) ([]domain.ForceMajeureEvent, error) {
	return s.fmRepo.List(ctx)
}
