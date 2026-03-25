package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const defaultEscrowClaimHours = 48

type EscrowService interface {
	CreateEscrow(ctx context.Context, bookingID uuid.UUID, amount int64, serviceFee int64) (*domain.Escrow, error)
	ReleaseToOwner(ctx context.Context, escrowID uuid.UUID) error
	MarkDisputed(ctx context.Context, escrowID uuid.UUID) error
	MarkDisputedByBookingID(ctx context.Context, bookingID uuid.UUID) error
	ProcessRefund(ctx context.Context, escrowID uuid.UUID, refundAmount int64) error
	ProcessMaturedEscrows(ctx context.Context) (int, error)
}

type escrowService struct {
	escrowRepo  repository.EscrowRepository
	bookingRepo repository.BookingRepository
	bhRepo      repository.BathhouseRepository
	walletSvc   WalletService
	logger      *logger.Logger
	claimHours  int
}

func NewEscrowService(
	escrowRepo repository.EscrowRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	walletSvc WalletService,
	log *logger.Logger,
	claimHours int,
) EscrowService {
	if claimHours < 24 || claimHours > 168 {
		claimHours = defaultEscrowClaimHours
	}
	return &escrowService{
		escrowRepo:  escrowRepo,
		bookingRepo: bookingRepo,
		bhRepo:      bhRepo,
		walletSvc:   walletSvc,
		logger:      log,
		claimHours:  claimHours,
	}
}

func (s *escrowService) CreateEscrow(ctx context.Context, bookingID uuid.UUID, amount int64, serviceFee int64) (*domain.Escrow, error) {
	now := time.Now()
	escrow := &domain.Escrow{
		ID:                uuid.New(),
		BookingID:         bookingID,
		Amount:            amount,
		ServiceFee:        serviceFee,
		Status:            domain.EscrowHeld,
		ClaimPeriodEndsAt: now.Add(time.Duration(s.claimHours) * time.Hour),
		CreatedAt:         now,
	}

	if err := escrow.Validate(); err != nil {
		return nil, err
	}

	if err := s.escrowRepo.Create(ctx, escrow); err != nil {
		return nil, fmt.Errorf("create escrow: %w", err)
	}

	s.logger.Info("Escrow created", "escrow_id", escrow.ID, "booking_id", bookingID, "amount", amount, "service_fee", serviceFee)
	return escrow, nil
}

func (s *escrowService) ReleaseToOwner(ctx context.Context, escrowID uuid.UUID) error {
	escrow, err := s.escrowRepo.GetByID(ctx, escrowID)
	if err != nil {
		return err
	}

	if escrow.Status == domain.EscrowReleased {
		return domain.ErrEscrowAlreadyReleased
	}
	if escrow.Status == domain.EscrowDisputed {
		return domain.ErrEscrowDisputed
	}
	if escrow.Status != domain.EscrowHeld {
		return domain.ErrEscrowAlreadyReleased
	}

	if time.Now().Before(escrow.ClaimPeriodEndsAt) {
		return domain.ErrEscrowNotMatured
	}

	// Get booking to find owner
	booking, err := s.bookingRepo.GetByID(ctx, escrow.BookingID)
	if err != nil {
		return fmt.Errorf("get booking for escrow release: %w", err)
	}

	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return fmt.Errorf("get bathhouse for escrow release: %w", err)
	}

	// Credit owner wallet with amount minus service fee.
	// IMPORTANT: Mark escrow as released FIRST, then credit wallet.
	// This prevents double-payout if the wallet credit succeeds but a subsequent
	// step fails -- ProcessMaturedEscrows would otherwise re-process the escrow.
	ownerAmount := escrow.Amount - escrow.ServiceFee

	now := time.Now()
	if err := s.escrowRepo.UpdateStatus(ctx, escrowID, domain.EscrowReleased, &now); err != nil {
		return fmt.Errorf("update escrow status to released: %w", err)
	}

	if ownerAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, bh.OwnerID)
		if err != nil {
			s.logger.Error("escrow marked released but failed to get owner wallet — manual reconciliation needed",
				"escrow_id", escrowID, "owner_id", bh.OwnerID, "amount", ownerAmount, "error", err)
			return fmt.Errorf("get owner wallet for escrow release: %w", err)
		}

		// Check if owner wallet can accommodate the full payout
		maxBalance := domain.MaxBalanceForCurrency(wallet.Currency)
		if wallet.Balance+ownerAmount > maxBalance {
			s.logger.Error("escrow marked released but owner wallet would exceed max balance — manual reconciliation needed",
				"escrow_id", escrowID, "owner_id", bh.OwnerID,
				"owner_amount", ownerAmount, "wallet_balance", wallet.Balance, "max_balance", maxBalance)
			return fmt.Errorf("owner wallet balance would exceed limit: %w", domain.ErrWalletLimitExceeded)
		}

		bookingID := escrow.BookingID
		_, err = s.walletSvc.Refund(ctx, wallet.ID, ownerAmount, "escrow_release", &bookingID, fmt.Sprintf("Выплата за бронирование %s", escrow.BookingID))
		if err != nil {
			s.logger.Error("escrow marked released but wallet credit failed — manual reconciliation needed",
				"escrow_id", escrowID, "owner_id", bh.OwnerID, "amount", ownerAmount, "error", err)
			return fmt.Errorf("credit owner wallet: %w", err)
		}
	}

	s.logger.Info("Escrow released to owner",
		"escrow_id", escrowID, "booking_id", escrow.BookingID,
		"owner_id", bh.OwnerID, "owner_amount", ownerAmount, "service_fee", escrow.ServiceFee)

	return nil
}

func (s *escrowService) MarkDisputed(ctx context.Context, escrowID uuid.UUID) error {
	escrow, err := s.escrowRepo.GetByID(ctx, escrowID)
	if err != nil {
		return err
	}

	if escrow.Status != domain.EscrowHeld {
		return domain.ErrEscrowAlreadyReleased
	}

	if err := s.escrowRepo.UpdateStatus(ctx, escrowID, domain.EscrowDisputed, nil); err != nil {
		return fmt.Errorf("mark escrow as disputed: %w", err)
	}

	s.logger.Info("Escrow marked as disputed", "escrow_id", escrowID, "booking_id", escrow.BookingID)
	return nil
}

func (s *escrowService) MarkDisputedByBookingID(ctx context.Context, bookingID uuid.UUID) error {
	escrow, err := s.escrowRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}
	return s.MarkDisputed(ctx, escrow.ID)
}

func (s *escrowService) ProcessRefund(ctx context.Context, escrowID uuid.UUID, refundAmount int64) error {
	escrow, err := s.escrowRepo.GetByID(ctx, escrowID)
	if err != nil {
		return err
	}

	if escrow.Status == domain.EscrowReleased {
		return domain.ErrEscrowAlreadyReleased
	}

	if refundAmount > escrow.Amount {
		return domain.ErrRefundExceedsAmount
	}

	// Get booking to find client
	booking, err := s.bookingRepo.GetByID(ctx, escrow.BookingID)
	if err != nil {
		return fmt.Errorf("get booking for escrow refund: %w", err)
	}

	// Refund client wallet
	if refundAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, booking.UserID)
		if err != nil {
			return fmt.Errorf("get client wallet for escrow refund: %w", err)
		}

		bookingID := escrow.BookingID
		_, err = s.walletSvc.Refund(ctx, wallet.ID, refundAmount, "escrow_refund", &bookingID, fmt.Sprintf("Возврат по бронированию %s", escrow.BookingID))
		if err != nil {
			return fmt.Errorf("refund client wallet: %w", err)
		}
	}

	now := time.Now()
	if err := s.escrowRepo.UpdateStatus(ctx, escrowID, domain.EscrowRefunded, &now); err != nil {
		return fmt.Errorf("update escrow status to refunded: %w", err)
	}

	s.logger.Info("Escrow refunded", "escrow_id", escrowID, "booking_id", escrow.BookingID, "refund_amount", refundAmount)
	return nil
}

func (s *escrowService) ProcessMaturedEscrows(ctx context.Context) (int, error) {
	escrows, err := s.escrowRepo.ListMatured(ctx)
	if err != nil {
		return 0, fmt.Errorf("list matured escrows: %w", err)
	}

	released := 0
	for _, escrow := range escrows {
		if err := s.ReleaseToOwner(ctx, escrow.ID); err != nil {
			s.logger.Error("Failed to release matured escrow",
				"escrow_id", escrow.ID, "booking_id", escrow.BookingID, "error", err)
			continue
		}
		released++
	}

	return released, nil
}
