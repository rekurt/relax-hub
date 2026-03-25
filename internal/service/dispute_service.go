package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const (
	evidenceWindowHours = 72
	appealWindowDays    = 7
)

type DisputeService interface {
	OpenDispute(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error)
	GetDispute(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) (*domain.Dispute, error)
	ListUserDisputes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error)
	ListAllDisputes(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error)
	SubmitEvidence(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID, evidence *domain.DisputeEvidence) error
	ListEvidence(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) ([]domain.DisputeEvidence, error)
	AssignDispute(ctx context.Context, disputeID uuid.UUID, mediatorID uuid.UUID) error
	ResolveDispute(ctx context.Context, disputeID uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string) error
	AppealDispute(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID) error
	CloseDispute(ctx context.Context, disputeID uuid.UUID) error
}

type disputeService struct {
	disputeRepo   repository.DisputeRepository
	escrowSvc     EscrowService
	walletSvc     WalletService
	bookingRepo   repository.BookingRepository
	bathhouseRepo repository.BathhouseRepository
	access        *AccessChecker
	logger        *logger.Logger
}

func NewDisputeService(
	disputeRepo repository.DisputeRepository,
	escrowSvc EscrowService,
	walletSvc WalletService,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	access *AccessChecker,
	log *logger.Logger,
) DisputeService {
	return &disputeService{
		disputeRepo:   disputeRepo,
		escrowSvc:     escrowSvc,
		walletSvc:     walletSvc,
		bookingRepo:   bookingRepo,
		bathhouseRepo: bathhouseRepo,
		access:        access,
		logger:        log,
	}
}

func (s *disputeService) OpenDispute(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error) {
	// Verify booking exists
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}

	// Only completed or no-show bookings can be disputed
	if booking.Status != domain.BookingCompleted && booking.Status != domain.BookingNoShow {
		return nil, domain.ErrInvalidInput
	}

	// Look up bathhouse to get owner
	bathhouse, err := s.bathhouseRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		return nil, fmt.Errorf("get bathhouse: %w", err)
	}

	// Determine respondent: if initiator is client -> respondent is owner; vice versa
	var respondentID uuid.UUID
	if booking.UserID == userID {
		respondentID = bathhouse.OwnerID
	} else if bathhouse.OwnerID == userID || s.access.CanManageBathhouse(ctx, userID, domain.RoleRepresentative, booking.BathhouseID) == nil {
		respondentID = booking.UserID
	} else {
		return nil, domain.ErrForbidden
	}

	// Check no existing dispute for this booking
	_, err = s.disputeRepo.GetByBookingID(ctx, bookingID)
	if err == nil {
		return nil, domain.ErrDisputeAlreadyExists
	}
	if !errors.Is(err, domain.ErrDisputeNotFound) {
		return nil, fmt.Errorf("check existing dispute: %w", err)
	}

	now := time.Now()
	evidenceDeadline := now.Add(evidenceWindowHours * time.Hour)

	dispute := &domain.Dispute{
		ID:               uuid.New(),
		BookingID:        bookingID,
		InitiatorID:      userID,
		RespondentID:     respondentID,
		Reason:           reason,
		Description:      description,
		Status:           domain.DisputeStatusEvidenceCollection,
		EvidenceDeadline: &evidenceDeadline,
	}

	if err := dispute.Validate(); err != nil {
		return nil, err
	}

	if err := s.disputeRepo.Create(ctx, dispute); err != nil {
		return nil, fmt.Errorf("create dispute: %w", err)
	}

	// Block escrow release
	if err := s.escrowSvc.MarkDisputedByBookingID(ctx, bookingID); err != nil {
		s.logger.Error("mark escrow disputed", "booking_id", bookingID, "error", err)
	}

	return dispute, nil
}

func (s *disputeService) GetDispute(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) (*domain.Dispute, error) {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return nil, err
	}

	if role != domain.RoleAdmin && dispute.InitiatorID != userID && dispute.RespondentID != userID {
		return nil, domain.ErrForbidden
	}

	return dispute, nil
}

func (s *disputeService) ListUserDisputes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error) {
	return s.disputeRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *disputeService) ListAllDisputes(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error) {
	return s.disputeRepo.ListAll(ctx, filter)
}

func (s *disputeService) SubmitEvidence(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID, evidence *domain.DisputeEvidence) error {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return err
	}

	// Only participants can submit evidence
	if dispute.InitiatorID != userID && dispute.RespondentID != userID {
		return domain.ErrForbidden
	}

	// Check evidence window
	if dispute.EvidenceDeadline != nil && time.Now().After(*dispute.EvidenceDeadline) {
		return domain.ErrDisputeEvidenceWindowExpired
	}

	// Only allow in evidence_collection or open status
	if dispute.Status != domain.DisputeStatusEvidenceCollection && dispute.Status != domain.DisputeStatusOpen {
		return domain.ErrDisputeAlreadyResolved
	}

	evidence.DisputeID = disputeID
	evidence.UserID = userID

	if err := evidence.Validate(); err != nil {
		return err
	}

	return s.disputeRepo.AddEvidence(ctx, evidence)
}

func (s *disputeService) ListEvidence(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) ([]domain.DisputeEvidence, error) {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return nil, err
	}

	if role != domain.RoleAdmin && dispute.InitiatorID != userID && dispute.RespondentID != userID {
		return nil, domain.ErrForbidden
	}

	return s.disputeRepo.ListEvidence(ctx, disputeID)
}

func (s *disputeService) AssignDispute(ctx context.Context, disputeID uuid.UUID, mediatorID uuid.UUID) error {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return err
	}

	if dispute.Status == domain.DisputeStatusResolved || dispute.Status == domain.DisputeStatusClosed {
		return domain.ErrDisputeAlreadyResolved
	}

	return s.disputeRepo.Assign(ctx, disputeID, mediatorID)
}

func (s *disputeService) ResolveDispute(ctx context.Context, disputeID uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string) error {
	if !resolution.IsValid() {
		return domain.ErrInvalidInput
	}

	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return err
	}

	if dispute.Status == domain.DisputeStatusResolved || dispute.Status == domain.DisputeStatusClosed {
		return domain.ErrDisputeAlreadyResolved
	}

	now := time.Now()
	if err := s.disputeRepo.UpdateResolution(ctx, disputeID, resolution, refundAmount, compensationAmount, mediatorNotes, now); err != nil {
		return fmt.Errorf("update resolution: %w", err)
	}

	// Process refund via escrow if applicable
	if resolution == domain.DisputeResolutionFullRefund || resolution == domain.DisputeResolutionPartialRefund {
		if refundAmount > 0 {
			if escrowErr := s.escrowSvc.ProcessRefundByBookingID(ctx, dispute.BookingID, refundAmount); escrowErr != nil {
				s.logger.Error("process dispute refund", "dispute_id", disputeID, "error", escrowErr)
			}
		}
	}

	// Credit compensation to initiator's wallet if applicable
	if compensationAmount > 0 {
		wallet, err := s.walletSvc.GetWallet(ctx, dispute.InitiatorID)
		if err != nil {
			s.logger.Error("get wallet for dispute compensation", "dispute_id", disputeID, "initiator_id", dispute.InitiatorID, "error", err)
		} else {
			_, err = s.walletSvc.AddBonus(ctx, wallet.ID, compensationAmount, domain.WalletTxBonus, nil, fmt.Sprintf("Компенсация по спору %s", disputeID))
			if err != nil {
				s.logger.Error("credit dispute compensation", "dispute_id", disputeID, "amount", compensationAmount, "error", err)
			}
		}
	}

	return nil
}

func (s *disputeService) AppealDispute(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID) error {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return err
	}

	// Only participants can appeal
	if dispute.InitiatorID != userID && dispute.RespondentID != userID {
		return domain.ErrForbidden
	}

	if dispute.Status == domain.DisputeStatusAppealed {
		return domain.ErrDisputeAlreadyAppealed
	}

	if dispute.Status != domain.DisputeStatusResolved {
		return domain.ErrDisputeNotResolved
	}

	// Check appeal deadline
	if dispute.AppealDeadline != nil && time.Now().After(*dispute.AppealDeadline) {
		return domain.ErrDisputeAppealExpired
	}

	return s.disputeRepo.UpdateAppeal(ctx, disputeID, domain.DisputeStatusAppealed)
}

func (s *disputeService) CloseDispute(ctx context.Context, disputeID uuid.UUID) error {
	dispute, err := s.disputeRepo.GetByID(ctx, disputeID)
	if err != nil {
		return err
	}

	if dispute.Status == domain.DisputeStatusClosed {
		return domain.ErrDisputeAlreadyClosed
	}

	return s.disputeRepo.UpdateStatus(ctx, disputeID, domain.DisputeStatusClosed)
}
