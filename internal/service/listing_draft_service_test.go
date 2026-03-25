package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

// stubKYCService returns configurable KYC approval status
type stubKYCService struct {
	approved bool
	err      error
}

func (s *stubKYCService) Submit(_ context.Context, _ uuid.UUID, _ service.SubmitKYCInput) (*domain.KYCApplication, error) {
	return nil, nil
}
func (s *stubKYCService) GetStatus(_ context.Context, _ uuid.UUID) (*domain.KYCApplication, error) {
	return nil, nil
}
func (s *stubKYCService) Approve(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubKYCService) Reject(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}
func (s *stubKYCService) ListPending(_ context.Context, _, _ int) (*domain.PaginatedResult[domain.KYCApplication], error) {
	return nil, nil
}
func (s *stubKYCService) GetByID(_ context.Context, _ uuid.UUID) (*domain.KYCApplication, error) {
	return nil, nil
}
func (s *stubKYCService) IsApproved(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.approved, s.err
}
func (s *stubKYCService) CheckExpiredApplications(_ context.Context) (int, error) {
	return 0, nil
}

// stubOfferService returns configurable offer acceptance status
type stubOfferService struct {
	accepted bool
	err      error
}

func (s *stubOfferService) Accept(_ context.Context, _ uuid.UUID, _ service.AcceptOfferInput) (*domain.OfferAcceptance, error) {
	return nil, nil
}
func (s *stubOfferService) GetStatus(_ context.Context, _ uuid.UUID) (*service.OfferStatusResponse, error) {
	return nil, nil
}
func (s *stubOfferService) IsAccepted(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.accepted, s.err
}

// stubPaymentDetailsService returns configurable payment details validation
type stubPaymentDetailsService struct {
	err error
}

func (s *stubPaymentDetailsService) Set(_ context.Context, _ uuid.UUID, _ service.SetPaymentDetailsInput) (*domain.PaymentDetails, error) {
	return nil, nil
}
func (s *stubPaymentDetailsService) Get(_ context.Context, _ uuid.UUID) (*domain.PaymentDetails, error) {
	return nil, nil
}
func (s *stubPaymentDetailsService) Validate(_ context.Context, _ uuid.UUID) error {
	return s.err
}

type draftTestEnv struct {
	svc               service.ListingDraftService
	repo              *mock.ListingDraftRepo
	kycStub           *stubKYCService
	offerStub         *stubOfferService
	paymentDetailsStub *stubPaymentDetailsService
}

func newDraftTestEnv() *draftTestEnv {
	repo := mock.NewListingDraftRepo()
	kycStub := &stubKYCService{approved: true}
	offerStub := &stubOfferService{accepted: true}
	paymentDetailsStub := &stubPaymentDetailsService{}
	log := logger.New(logger.LevelWarn)
	svc := service.NewListingDraftService(repo, kycStub, offerStub, paymentDetailsStub, log)
	return &draftTestEnv{
		svc:                svc,
		repo:               repo,
		kycStub:            kycStub,
		offerStub:          offerStub,
		paymentDetailsStub: paymentDetailsStub,
	}
}

func TestListingDraftService_Create_Success(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, err := env.svc.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.UserID != userID {
		t.Errorf("user_id = %v, want %v", draft.UserID, userID)
	}
	if draft.Status != domain.ListingDraftStatusDraft {
		t.Errorf("status = %v, want draft", draft.Status)
	}
	if draft.CurrentStep != domain.ListingStepBasicInfo {
		t.Errorf("current_step = %d, want %d", draft.CurrentStep, domain.ListingStepBasicInfo)
	}
}

func TestListingDraftService_Create_KYCNotApproved(t *testing.T) {
	env := newDraftTestEnv()
	env.kycStub.approved = false

	_, err := env.svc.Create(context.Background(), uuid.New())
	if err != domain.ErrKYCNotApproved {
		t.Fatalf("expected ErrKYCNotApproved, got: %v", err)
	}
}

func TestListingDraftService_Create_OfferNotAccepted(t *testing.T) {
	env := newDraftTestEnv()
	env.offerStub.accepted = false

	_, err := env.svc.Create(context.Background(), uuid.New())
	if err != domain.ErrOfferNotAccepted {
		t.Fatalf("expected ErrOfferNotAccepted, got: %v", err)
	}
}

func TestListingDraftService_SaveStep_Success(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	data := json.RawMessage(`{"name":"Баня на реке","type":"russian","description":"Отличная баня"}`)
	err := env.svc.SaveStep(context.Background(), draft.ID, userID, 1, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify step was saved
	updated, _ := env.svc.GetDraft(context.Background(), draft.ID, userID)
	if _, ok := updated.StepData[1]; !ok {
		t.Error("step 1 data should be saved")
	}
}

func TestListingDraftService_SaveStep_InvalidStep(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	data := json.RawMessage(`{"test":"data"}`)
	err := env.svc.SaveStep(context.Background(), draft.ID, userID, 0, data)
	if err != domain.ErrListingDraftInvalidStep {
		t.Fatalf("expected ErrListingDraftInvalidStep, got: %v", err)
	}

	err = env.svc.SaveStep(context.Background(), draft.ID, userID, 8, data)
	if err != domain.ErrListingDraftInvalidStep {
		t.Fatalf("expected ErrListingDraftInvalidStep for step 8, got: %v", err)
	}
}

func TestListingDraftService_SaveStep_WrongUser(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()
	otherUser := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	data := json.RawMessage(`{"test":"data"}`)
	err := env.svc.SaveStep(context.Background(), draft.ID, otherUser, 1, data)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestListingDraftService_SaveStep_AlreadySubmitted(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)
	// Fill all steps and submit
	for step := 1; step <= 6; step++ {
		data := json.RawMessage(`{"test":"data"}`)
		_ = env.svc.SaveStep(context.Background(), draft.ID, userID, step, data)
	}
	_, _ = env.svc.Submit(context.Background(), draft.ID, userID)

	data := json.RawMessage(`{"test":"data"}`)
	err := env.svc.SaveStep(context.Background(), draft.ID, userID, 1, data)
	if err != domain.ErrListingDraftSubmitted {
		t.Fatalf("expected ErrListingDraftSubmitted, got: %v", err)
	}
}

func TestListingDraftService_GetDraft_Success(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)
	got, err := env.svc.GetDraft(context.Background(), draft.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != draft.ID {
		t.Errorf("id = %v, want %v", got.ID, draft.ID)
	}
}

func TestListingDraftService_GetDraft_WrongUser(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)
	_, err := env.svc.GetDraft(context.Background(), draft.ID, uuid.New())
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestListingDraftService_GetDraft_NotFound(t *testing.T) {
	env := newDraftTestEnv()

	_, err := env.svc.GetDraft(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrListingDraftNotFound {
		t.Fatalf("expected ErrListingDraftNotFound, got: %v", err)
	}
}

func TestListingDraftService_ListDrafts(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	// Create 2 drafts
	_, _ = env.svc.Create(context.Background(), userID)
	_, _ = env.svc.Create(context.Background(), userID)

	drafts, err := env.svc.ListDrafts(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drafts) != 2 {
		t.Errorf("drafts count = %d, want 2", len(drafts))
	}
}

func TestListingDraftService_Submit_Success(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	// Fill steps 1-6
	for step := 1; step <= 6; step++ {
		data := json.RawMessage(`{"test":"data"}`)
		_ = env.svc.SaveStep(context.Background(), draft.ID, userID, step, data)
	}

	submitted, err := env.svc.Submit(context.Background(), draft.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if submitted.Status != domain.ListingDraftStatusSubmitted {
		t.Errorf("status = %v, want submitted", submitted.Status)
	}
}

func TestListingDraftService_Submit_Incomplete(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	// Only fill steps 1-3
	for step := 1; step <= 3; step++ {
		data := json.RawMessage(`{"test":"data"}`)
		_ = env.svc.SaveStep(context.Background(), draft.ID, userID, step, data)
	}

	_, err := env.svc.Submit(context.Background(), draft.ID, userID)
	if err != domain.ErrListingDraftIncomplete {
		t.Fatalf("expected ErrListingDraftIncomplete, got: %v", err)
	}
}

func TestListingDraftService_Submit_AlreadySubmitted(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)
	for step := 1; step <= 6; step++ {
		_ = env.svc.SaveStep(context.Background(), draft.ID, userID, step, json.RawMessage(`{"test":"data"}`))
	}
	_, _ = env.svc.Submit(context.Background(), draft.ID, userID)

	_, err := env.svc.Submit(context.Background(), draft.ID, userID)
	if err != domain.ErrListingDraftSubmitted {
		t.Fatalf("expected ErrListingDraftSubmitted, got: %v", err)
	}
}

func TestListingDraftService_Delete_Success(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	err := env.svc.Delete(context.Background(), draft.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deleted
	_, err = env.svc.GetDraft(context.Background(), draft.ID, userID)
	if err != domain.ErrListingDraftNotFound {
		t.Fatalf("expected ErrListingDraftNotFound after delete, got: %v", err)
	}
}

func TestListingDraftService_Delete_WrongUser(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	err := env.svc.Delete(context.Background(), draft.ID, uuid.New())
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestListingDraftService_SaveStep_AdvancesCurrentStep(t *testing.T) {
	env := newDraftTestEnv()
	userID := uuid.New()

	draft, _ := env.svc.Create(context.Background(), userID)

	// Save step 3
	data := json.RawMessage(`{"test":"data"}`)
	_ = env.svc.SaveStep(context.Background(), draft.ID, userID, 3, data)

	got, _ := env.svc.GetDraft(context.Background(), draft.ID, userID)
	if got.CurrentStep != 4 {
		t.Errorf("current_step = %d, want 4 after saving step 3", got.CurrentStep)
	}

	// Save step 7 (last step) should cap at 7
	_ = env.svc.SaveStep(context.Background(), draft.ID, userID, 7, data)
	got, _ = env.svc.GetDraft(context.Background(), draft.ID, userID)
	if got.CurrentStep != 7 {
		t.Errorf("current_step = %d, want 7 after saving step 7", got.CurrentStep)
	}
}
