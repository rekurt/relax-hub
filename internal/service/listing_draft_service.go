package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ListingDraftService interface {
	Create(ctx context.Context, userID uuid.UUID) (*domain.ListingDraft, error)
	SaveStep(ctx context.Context, draftID, userID uuid.UUID, step int, data json.RawMessage) error
	GetDraft(ctx context.Context, draftID, userID uuid.UUID) (*domain.ListingDraft, error)
	ListDrafts(ctx context.Context, userID uuid.UUID) ([]domain.ListingDraft, error)
	Submit(ctx context.Context, draftID, userID uuid.UUID) (*domain.ListingDraft, error)
	Delete(ctx context.Context, draftID, userID uuid.UUID) error
}

type listingDraftService struct {
	draftRepo         repository.ListingDraftRepository
	kycSvc            KYCService
	offerSvc          OfferService
	paymentDetailsSvc PaymentDetailsService
	logger            *logger.Logger
}

func NewListingDraftService(
	draftRepo repository.ListingDraftRepository,
	kycSvc KYCService,
	offerSvc OfferService,
	paymentDetailsSvc PaymentDetailsService,
	log *logger.Logger,
) ListingDraftService {
	return &listingDraftService{
		draftRepo:         draftRepo,
		kycSvc:            kycSvc,
		offerSvc:          offerSvc,
		paymentDetailsSvc: paymentDetailsSvc,
		logger:            log,
	}
}

func (s *listingDraftService) Create(ctx context.Context, userID uuid.UUID) (*domain.ListingDraft, error) {
	// Check KYC approved
	approved, err := s.kycSvc.IsApproved(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !approved {
		return nil, domain.ErrKYCNotApproved
	}

	// Check offer accepted
	accepted, err := s.offerSvc.IsAccepted(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, domain.ErrOfferNotAccepted
	}

	// Check payment details set
	if err := s.paymentDetailsSvc.Validate(ctx, userID); err != nil {
		return nil, err
	}

	now := time.Now()
	draft := &domain.ListingDraft{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      domain.ListingDraftStatusDraft,
		CurrentStep: domain.ListingStepBasicInfo,
		StepData:    make(map[int]json.RawMessage),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.draftRepo.Create(ctx, draft); err != nil {
		return nil, err
	}

	s.logger.Info("listing draft created", "draft_id", draft.ID, "user_id", userID)
	return draft, nil
}

func (s *listingDraftService) SaveStep(ctx context.Context, draftID, userID uuid.UUID, step int, data json.RawMessage) error {
	if !domain.IsStepValid(step) {
		return domain.ErrListingDraftInvalidStep
	}

	draft, err := s.draftRepo.GetByID(ctx, draftID)
	if err != nil {
		return err
	}
	if draft.UserID != userID {
		return domain.ErrForbidden
	}
	if draft.Status != domain.ListingDraftStatusDraft {
		return domain.ErrListingDraftSubmitted
	}

	// Advance current step if saving a step beyond current
	currentStep := draft.CurrentStep
	if step >= currentStep {
		currentStep = step + 1
		if currentStep > domain.ListingTotalSteps {
			currentStep = domain.ListingTotalSteps
		}
	}

	if err := s.draftRepo.UpdateStep(ctx, draftID, step, data, currentStep); err != nil {
		return err
	}

	s.logger.Debug("listing draft step saved", "draft_id", draftID, "step", step)
	return nil
}

func (s *listingDraftService) GetDraft(ctx context.Context, draftID, userID uuid.UUID) (*domain.ListingDraft, error) {
	draft, err := s.draftRepo.GetByID(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if draft.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return draft, nil
}

func (s *listingDraftService) ListDrafts(ctx context.Context, userID uuid.UUID) ([]domain.ListingDraft, error) {
	return s.draftRepo.ListByUserID(ctx, userID)
}

func (s *listingDraftService) Submit(ctx context.Context, draftID, userID uuid.UUID) (*domain.ListingDraft, error) {
	draft, err := s.draftRepo.GetByID(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if draft.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if draft.Status != domain.ListingDraftStatusDraft {
		return nil, domain.ErrListingDraftSubmitted
	}
	if !draft.AllStepsComplete() {
		return nil, domain.ErrListingDraftIncomplete
	}

	if err := s.draftRepo.UpdateStatus(ctx, draftID, domain.ListingDraftStatusSubmitted); err != nil {
		return nil, err
	}

	s.logger.Info("listing draft submitted", "draft_id", draftID, "user_id", userID)

	// Re-fetch to return updated state
	return s.draftRepo.GetByID(ctx, draftID)
}

func (s *listingDraftService) Delete(ctx context.Context, draftID, userID uuid.UUID) error {
	draft, err := s.draftRepo.GetByID(ctx, draftID)
	if err != nil {
		return err
	}
	if draft.UserID != userID {
		return domain.ErrForbidden
	}

	if err := s.draftRepo.Delete(ctx, draftID); err != nil {
		return err
	}

	s.logger.Info("listing draft deleted", "draft_id", draftID, "user_id", userID)
	return nil
}
