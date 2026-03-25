package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

// mockEscrowService implements service.EscrowService for testing
type mockEscrowService struct {
	disputedBookings []uuid.UUID
	refundedBookings []uuid.UUID
}

func (m *mockEscrowService) CreateEscrow(_ context.Context, _ uuid.UUID, _ int64, _ int64) (*domain.Escrow, error) {
	return &domain.Escrow{ID: uuid.New()}, nil
}

func (m *mockEscrowService) ReleaseToOwner(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockEscrowService) MarkDisputed(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockEscrowService) MarkDisputedByBookingID(_ context.Context, bookingID uuid.UUID) error {
	m.disputedBookings = append(m.disputedBookings, bookingID)
	return nil
}

func (m *mockEscrowService) ProcessRefund(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}

func (m *mockEscrowService) ProcessRefundByBookingID(_ context.Context, bookingID uuid.UUID, _ int64) error {
	m.refundedBookings = append(m.refundedBookings, bookingID)
	return nil
}

func (m *mockEscrowService) ProcessMaturedEscrows(_ context.Context) (int, error) {
	return 0, nil
}

// mockDisputeWalletService implements service.WalletService for dispute testing
type mockDisputeWalletService struct {
	wallets map[uuid.UUID]*domain.Wallet
}

func newMockDisputeWalletService() *mockDisputeWalletService {
	return &mockDisputeWalletService{wallets: make(map[uuid.UUID]*domain.Wallet)}
}

func (m *mockDisputeWalletService) CreateWallet(_ context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
	w := &domain.Wallet{ID: uuid.New(), UserID: userID, Currency: currency}
	m.wallets[userID] = w
	return w, nil
}

func (m *mockDisputeWalletService) GetWallet(_ context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	w, ok := m.wallets[userID]
	if !ok {
		return nil, domain.ErrWalletNotFound
	}
	return w, nil
}

func (m *mockDisputeWalletService) TopUp(_ context.Context, _ uuid.UUID, _ int64) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) Spend(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) Hold(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	return &domain.WalletHold{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) CaptureHold(_ context.Context, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) ReleaseHold(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockDisputeWalletService) Refund(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) AddBonus(_ context.Context, _ uuid.UUID, _ int64, _ domain.WalletTransactionType, _ *time.Time, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}
func (m *mockDisputeWalletService) GetBalance(_ context.Context, _ uuid.UUID) (*service.WalletBalanceSummary, error) {
	return &service.WalletBalanceSummary{}, nil
}
func (m *mockDisputeWalletService) ListTransactions(_ context.Context, _ uuid.UUID, _ domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	return &domain.PaginatedResult[domain.WalletTransaction]{}, nil
}
func (m *mockDisputeWalletService) GetActiveHolds(_ context.Context, _ uuid.UUID) ([]domain.WalletHold, error) {
	return nil, nil
}
func (m *mockDisputeWalletService) ExpireBonuses(_ context.Context) (int, error) { return 0, nil }
func (m *mockDisputeWalletService) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockDisputeWalletService) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error {
	return nil
}

func newDisputeTestService() (service.DisputeService, *mock.DisputeRepo, *mock.BookingRepo, *mock.BathhouseRepo, *mockEscrowService) {
	disputeRepo := mock.NewDisputeRepo().(*mock.DisputeRepo)
	bookingRepo := mock.NewBookingRepo()
	bhRepo := mock.NewBathhouseRepo()
	escrowSvc := &mockEscrowService{}
	walletSvc := newMockDisputeWalletService()
	log := logger.New(logger.LevelWarn)
	svc := service.NewDisputeService(disputeRepo, escrowSvc, walletSvc, bookingRepo, bhRepo, log)
	return svc, disputeRepo, bookingRepo, bhRepo, escrowSvc
}

func createTestBookingAndBathhouse(ctx context.Context, bookingRepo *mock.BookingRepo, bhRepo *mock.BathhouseRepo, clientID, ownerID uuid.UUID) *domain.Booking {
	bh := &domain.Bathhouse{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "Test Banya",
		Slug:    "test-banya-" + uuid.New().String()[:8],
		Status:  domain.BathhouseStatusActive,
		CityID:  1,
	}
	_ = bhRepo.Create(ctx, bh)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-1 * time.Hour),
		GuestCount:  2,
		TotalPrice:  500000,
		Status:      domain.BookingCompleted,
	}
	_ = bookingRepo.Create(ctx, booking)

	return booking
}

func TestDisputeService_OpenDispute(t *testing.T) {
	svc, _, bookingRepo, bhRepo, escrowSvc := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	dispute, err := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonPoorQuality, "The banya was not heated properly")
	if err != nil {
		t.Fatalf("OpenDispute: %v", err)
	}

	if dispute.BookingID != booking.ID {
		t.Errorf("expected booking_id=%s, got %s", booking.ID, dispute.BookingID)
	}
	if dispute.InitiatorID != clientID {
		t.Errorf("expected initiator_id=%s, got %s", clientID, dispute.InitiatorID)
	}
	if dispute.RespondentID != ownerID {
		t.Errorf("expected respondent_id=%s, got %s", ownerID, dispute.RespondentID)
	}
	if dispute.Status != domain.DisputeStatusEvidenceCollection {
		t.Errorf("expected status=evidence_collection, got %s", dispute.Status)
	}
	if dispute.EvidenceDeadline == nil {
		t.Error("expected evidence_deadline to be set")
	}

	// Verify escrow was marked disputed
	if len(escrowSvc.disputedBookings) != 1 || escrowSvc.disputedBookings[0] != booking.ID {
		t.Error("expected escrow to be marked disputed for the booking")
	}
}

func TestDisputeService_OpenDispute_OwnerInitiates(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	dispute, err := svc.OpenDispute(ctx, ownerID, booking.ID, domain.DisputeReasonDamage, "Client damaged the pool")
	if err != nil {
		t.Fatalf("OpenDispute by owner: %v", err)
	}

	if dispute.InitiatorID != ownerID {
		t.Errorf("expected initiator_id=%s, got %s", ownerID, dispute.InitiatorID)
	}
	if dispute.RespondentID != clientID {
		t.Errorf("expected respondent_id=%s, got %s", clientID, dispute.RespondentID)
	}
}

func TestDisputeService_OpenDispute_Forbidden(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	_, err := svc.OpenDispute(ctx, strangerID, booking.ID, domain.DisputeReasonOther, "Random person")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestDisputeService_OpenDispute_Duplicate(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	_, err := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonPoorQuality, "First dispute")
	if err != nil {
		t.Fatalf("first OpenDispute: %v", err)
	}

	_, err = svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Second dispute")
	if !errors.Is(err, domain.ErrDisputeAlreadyExists) {
		t.Errorf("expected ErrDisputeAlreadyExists, got %v", err)
	}
}

func TestDisputeService_SubmitEvidence(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonPoorQuality, "Bad quality")

	evidence := &domain.DisputeEvidence{
		ID:          uuid.New(),
		Type:        domain.DisputeEvidencePhoto,
		URL:         "https://example.com/photo.jpg",
		Description: "Photo of the issue",
	}

	err := svc.SubmitEvidence(ctx, clientID, dispute.ID, evidence)
	if err != nil {
		t.Fatalf("SubmitEvidence: %v", err)
	}

	// Respondent can also submit evidence
	evidence2 := &domain.DisputeEvidence{
		ID:          uuid.New(),
		Type:        domain.DisputeEvidenceScreenshot,
		URL:         "https://example.com/screenshot.png",
		Description: "Screenshot from our side",
	}
	err = svc.SubmitEvidence(ctx, ownerID, dispute.ID, evidence2)
	if err != nil {
		t.Fatalf("SubmitEvidence by respondent: %v", err)
	}

	// List evidence
	items, err := svc.ListEvidence(ctx, clientID, domain.RoleClient, dispute.ID)
	if err != nil {
		t.Fatalf("ListEvidence: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 evidence items, got %d", len(items))
	}
}

func TestDisputeService_SubmitEvidence_Forbidden(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonPoorQuality, "Issue")

	evidence := &domain.DisputeEvidence{
		ID:   uuid.New(),
		Type: domain.DisputeEvidencePhoto,
		URL:  "https://example.com/photo.jpg",
	}
	err := svc.SubmitEvidence(ctx, strangerID, dispute.ID, evidence)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestDisputeService_ResolveDispute(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonBillingError, "Wrong amount charged")

	err := svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionFullRefund, 500000, 0, "Client was overcharged")
	if err != nil {
		t.Fatalf("ResolveDispute: %v", err)
	}

	// Verify dispute is resolved
	resolved, err := svc.GetDispute(ctx, clientID, domain.RoleClient, dispute.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if resolved.Status != domain.DisputeStatusResolved {
		t.Errorf("expected status=resolved, got %s", resolved.Status)
	}
	if resolved.Resolution == nil || *resolved.Resolution != domain.DisputeResolutionFullRefund {
		t.Error("expected resolution=full_refund")
	}
	if resolved.RefundAmount != 500000 {
		t.Errorf("expected refund_amount=500000, got %d", resolved.RefundAmount)
	}
	if resolved.AppealDeadline == nil {
		t.Error("expected appeal_deadline to be set after resolution")
	}
}

func TestDisputeService_ResolveDispute_AlreadyResolved(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Issue")

	_ = svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionNoRefund, 0, 0, "No issue found")

	err := svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionFullRefund, 500000, 0, "Changed my mind")
	if !errors.Is(err, domain.ErrDisputeAlreadyResolved) {
		t.Errorf("expected ErrDisputeAlreadyResolved, got %v", err)
	}
}

func TestDisputeService_AppealDispute(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonPoorQuality, "Bad service")
	_ = svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionNoRefund, 0, 0, "No issues found")

	err := svc.AppealDispute(ctx, clientID, dispute.ID)
	if err != nil {
		t.Fatalf("AppealDispute: %v", err)
	}

	// Verify dispute is appealed
	appealed, _ := svc.GetDispute(ctx, clientID, domain.RoleClient, dispute.ID)
	if appealed.Status != domain.DisputeStatusAppealed {
		t.Errorf("expected status=appealed, got %s", appealed.Status)
	}
}

func TestDisputeService_AppealDispute_NotResolved(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Issue")

	err := svc.AppealDispute(ctx, clientID, dispute.ID)
	if !errors.Is(err, domain.ErrDisputeNotResolved) {
		t.Errorf("expected ErrDisputeNotResolved, got %v", err)
	}
}

func TestDisputeService_AppealDispute_Forbidden(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Issue")
	_ = svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionNoRefund, 0, 0, "No issues")

	err := svc.AppealDispute(ctx, strangerID, dispute.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestDisputeService_AssignDispute(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	mediatorID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonSafetyIssue, "Safety concern")

	err := svc.AssignDispute(ctx, dispute.ID, mediatorID)
	if err != nil {
		t.Fatalf("AssignDispute: %v", err)
	}

	// Verify dispute is under review
	assigned, _ := svc.GetDispute(ctx, clientID, domain.RoleAdmin, dispute.ID)
	if assigned.Status != domain.DisputeStatusUnderReview {
		t.Errorf("expected status=under_review, got %s", assigned.Status)
	}
	if assigned.MediatorID == nil || *assigned.MediatorID != mediatorID {
		t.Error("expected mediator_id to be set")
	}
}

func TestDisputeService_CloseDispute(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Issue")

	err := svc.CloseDispute(ctx, dispute.ID)
	if err != nil {
		t.Fatalf("CloseDispute: %v", err)
	}

	closed, _ := svc.GetDispute(ctx, clientID, domain.RoleAdmin, dispute.ID)
	if closed.Status != domain.DisputeStatusClosed {
		t.Errorf("expected status=closed, got %s", closed.Status)
	}

	// Close again should fail
	err = svc.CloseDispute(ctx, dispute.ID)
	if !errors.Is(err, domain.ErrDisputeAlreadyClosed) {
		t.Errorf("expected ErrDisputeAlreadyClosed, got %v", err)
	}
}

func TestDisputeService_GetDispute_Forbidden(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, _ := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonOther, "Issue")

	_, err := svc.GetDispute(ctx, strangerID, domain.RoleClient, dispute.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}

	// Admin can see any dispute
	_, err = svc.GetDispute(ctx, strangerID, domain.RoleAdmin, dispute.ID)
	if err != nil {
		t.Errorf("admin should be able to see dispute, got %v", err)
	}
}

func TestDisputeService_ListUserDisputes(t *testing.T) {
	svc, _, bookingRepo, bhRepo, _ := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()

	booking1 := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	booking2 := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)

	_, _ = svc.OpenDispute(ctx, clientID, booking1.ID, domain.DisputeReasonPoorQuality, "Issue 1")
	_, _ = svc.OpenDispute(ctx, clientID, booking2.ID, domain.DisputeReasonBillingError, "Issue 2")

	result, err := svc.ListUserDisputes(ctx, clientID, 1, 20)
	if err != nil {
		t.Fatalf("ListUserDisputes: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected 2 disputes, got %d", result.TotalCount)
	}
}

func TestDisputeService_FullLifecycle(t *testing.T) {
	svc, _, bookingRepo, bhRepo, escrowSvc := newDisputeTestService()
	ctx := context.Background()
	clientID := uuid.New()
	ownerID := uuid.New()
	mediatorID := uuid.New()

	// 1. Open dispute
	booking := createTestBookingAndBathhouse(ctx, bookingRepo, bhRepo, clientID, ownerID)
	dispute, err := svc.OpenDispute(ctx, clientID, booking.ID, domain.DisputeReasonServiceNotProvided, "Service was not provided as described")
	if err != nil {
		t.Fatalf("OpenDispute: %v", err)
	}
	if len(escrowSvc.disputedBookings) != 1 {
		t.Fatal("expected escrow to be marked disputed")
	}

	// 2. Submit evidence from both sides
	err = svc.SubmitEvidence(ctx, clientID, dispute.ID, &domain.DisputeEvidence{
		ID: uuid.New(), Type: domain.DisputeEvidencePhoto, URL: "https://example.com/photo1.jpg",
	})
	if err != nil {
		t.Fatalf("SubmitEvidence (client): %v", err)
	}

	err = svc.SubmitEvidence(ctx, ownerID, dispute.ID, &domain.DisputeEvidence{
		ID: uuid.New(), Type: domain.DisputeEvidenceMessage, URL: "https://example.com/chat-log.txt", Description: "Chat log showing agreement",
	})
	if err != nil {
		t.Fatalf("SubmitEvidence (owner): %v", err)
	}

	// 3. Admin assigns mediator
	err = svc.AssignDispute(ctx, dispute.ID, mediatorID)
	if err != nil {
		t.Fatalf("AssignDispute: %v", err)
	}

	// 4. Admin resolves with partial refund
	err = svc.ResolveDispute(ctx, dispute.ID, domain.DisputeResolutionPartialRefund, 250000, 50000, "Partial service provided, partial refund issued")
	if err != nil {
		t.Fatalf("ResolveDispute: %v", err)
	}

	// 5. Client appeals
	err = svc.AppealDispute(ctx, clientID, dispute.ID)
	if err != nil {
		t.Fatalf("AppealDispute: %v", err)
	}

	// 6. Verify final state
	final, _ := svc.GetDispute(ctx, clientID, domain.RoleClient, dispute.ID)
	if final.Status != domain.DisputeStatusAppealed {
		t.Errorf("expected status=appealed, got %s", final.Status)
	}

	// 7. List evidence
	evidence, _ := svc.ListEvidence(ctx, clientID, domain.RoleClient, dispute.ID)
	if len(evidence) != 2 {
		t.Errorf("expected 2 evidence items, got %d", len(evidence))
	}
}
