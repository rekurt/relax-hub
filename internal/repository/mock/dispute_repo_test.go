package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestDisputeRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewDisputeRepo()

	initiatorID := uuid.New()
	respondentID := uuid.New()
	bookingID := uuid.New()

	dispute := &domain.Dispute{
		BookingID:    bookingID,
		InitiatorID:  initiatorID,
		RespondentID: respondentID,
		Reason:       domain.DisputeReasonPoorQuality,
		Description:  "The room was dirty",
		Status:       domain.DisputeStatusOpen,
	}

	// Create
	if err := repo.Create(ctx, dispute); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if dispute.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// Create duplicate booking_id should fail
	dup := &domain.Dispute{
		BookingID:    bookingID,
		InitiatorID:  uuid.New(),
		RespondentID: uuid.New(),
		Reason:       domain.DisputeReasonOther,
		Description:  "Duplicate",
		Status:       domain.DisputeStatusOpen,
	}
	if err := repo.Create(ctx, dup); err != domain.ErrDisputeAlreadyExists {
		t.Errorf("Create duplicate: want ErrDisputeAlreadyExists, got %v", err)
	}

	// GetByID
	got, err := repo.GetByID(ctx, dispute.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.InitiatorID != initiatorID {
		t.Errorf("InitiatorID = %v, want %v", got.InitiatorID, initiatorID)
	}
	if got.Reason != domain.DisputeReasonPoorQuality {
		t.Errorf("Reason = %q, want %q", got.Reason, domain.DisputeReasonPoorQuality)
	}
	if got.Description != "The room was dirty" {
		t.Errorf("Description = %q, want %q", got.Description, "The room was dirty")
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, uuid.New())
	if err != domain.ErrDisputeNotFound {
		t.Errorf("GetByID not found: want ErrDisputeNotFound, got %v", err)
	}

	// GetByBookingID
	got, err = repo.GetByBookingID(ctx, bookingID)
	if err != nil {
		t.Fatalf("GetByBookingID: %v", err)
	}
	if got.ID != dispute.ID {
		t.Errorf("GetByBookingID ID = %v, want %v", got.ID, dispute.ID)
	}

	// GetByBookingID not found
	_, err = repo.GetByBookingID(ctx, uuid.New())
	if err != domain.ErrDisputeNotFound {
		t.Errorf("GetByBookingID not found: want ErrDisputeNotFound, got %v", err)
	}

	// UpdateStatus
	if err := repo.UpdateStatus(ctx, dispute.ID, domain.DisputeStatusEvidenceCollection); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, dispute.ID)
	if got.Status != domain.DisputeStatusEvidenceCollection {
		t.Errorf("Status = %q, want %q", got.Status, domain.DisputeStatusEvidenceCollection)
	}

	// UpdateStatus not found
	if err := repo.UpdateStatus(ctx, uuid.New(), domain.DisputeStatusClosed); err != domain.ErrDisputeNotFound {
		t.Errorf("UpdateStatus not found: want ErrDisputeNotFound, got %v", err)
	}

	// Assign
	mediatorID := uuid.New()
	if err := repo.Assign(ctx, dispute.ID, mediatorID); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	got, _ = repo.GetByID(ctx, dispute.ID)
	if got.MediatorID == nil || *got.MediatorID != mediatorID {
		t.Error("MediatorID should be set")
	}
	if got.Status != domain.DisputeStatusUnderReview {
		t.Errorf("Status after assign = %q, want %q", got.Status, domain.DisputeStatusUnderReview)
	}

	// Assign not found
	if err := repo.Assign(ctx, uuid.New(), mediatorID); err != domain.ErrDisputeNotFound {
		t.Errorf("Assign not found: want ErrDisputeNotFound, got %v", err)
	}

	// UpdateResolution
	resolvedAt := time.Now()
	if err := repo.UpdateResolution(ctx, dispute.ID, domain.DisputeResolutionPartialRefund, 500000, 100000, "Partial refund issued", resolvedAt); err != nil {
		t.Fatalf("UpdateResolution: %v", err)
	}
	got, _ = repo.GetByID(ctx, dispute.ID)
	if got.Status != domain.DisputeStatusResolved {
		t.Errorf("Status = %q, want %q", got.Status, domain.DisputeStatusResolved)
	}
	if got.Resolution == nil || *got.Resolution != domain.DisputeResolutionPartialRefund {
		t.Error("Resolution should be partial_refund")
	}
	if got.RefundAmount != 500000 {
		t.Errorf("RefundAmount = %d, want 500000", got.RefundAmount)
	}
	if got.CompensationAmount != 100000 {
		t.Errorf("CompensationAmount = %d, want 100000", got.CompensationAmount)
	}
	if got.MediatorNotes != "Partial refund issued" {
		t.Errorf("MediatorNotes = %q, want %q", got.MediatorNotes, "Partial refund issued")
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}
	if got.AppealDeadline == nil {
		t.Error("AppealDeadline should be set (7 days after resolution)")
	}
	expectedDeadline := resolvedAt.Add(7 * 24 * time.Hour)
	if got.AppealDeadline != nil && got.AppealDeadline.Sub(expectedDeadline) > time.Second {
		t.Errorf("AppealDeadline = %v, want ~%v", got.AppealDeadline, expectedDeadline)
	}

	// UpdateResolution not found
	if err := repo.UpdateResolution(ctx, uuid.New(), domain.DisputeResolutionFullRefund, 0, 0, "", resolvedAt); err != domain.ErrDisputeNotFound {
		t.Errorf("UpdateResolution not found: want ErrDisputeNotFound, got %v", err)
	}

	// UpdateAppeal
	if err := repo.UpdateAppeal(ctx, dispute.ID, domain.DisputeStatusAppealed); err != nil {
		t.Fatalf("UpdateAppeal: %v", err)
	}
	got, _ = repo.GetByID(ctx, dispute.ID)
	if got.Status != domain.DisputeStatusAppealed {
		t.Errorf("Status = %q, want %q", got.Status, domain.DisputeStatusAppealed)
	}

	// UpdateAppeal not found
	if err := repo.UpdateAppeal(ctx, uuid.New(), domain.DisputeStatusAppealed); err != domain.ErrDisputeNotFound {
		t.Errorf("UpdateAppeal not found: want ErrDisputeNotFound, got %v", err)
	}
}

func TestDisputeRepo_Evidence(t *testing.T) {
	ctx := context.Background()
	repo := NewDisputeRepo()

	dispute := &domain.Dispute{
		BookingID:    uuid.New(),
		InitiatorID:  uuid.New(),
		RespondentID: uuid.New(),
		Reason:       domain.DisputeReasonDamage,
		Description:  "Damage to property",
		Status:       domain.DisputeStatusEvidenceCollection,
	}
	if err := repo.Create(ctx, dispute); err != nil {
		t.Fatalf("Create dispute: %v", err)
	}

	// AddEvidence
	ev1 := &domain.DisputeEvidence{
		DisputeID:   dispute.ID,
		UserID:      dispute.InitiatorID,
		Type:        domain.DisputeEvidencePhoto,
		URL:         "https://example.com/photo1.jpg",
		Description: "Photo of damage",
	}
	if err := repo.AddEvidence(ctx, ev1); err != nil {
		t.Fatalf("AddEvidence: %v", err)
	}
	if ev1.ID == uuid.Nil {
		t.Error("Evidence ID should be assigned")
	}

	// AddEvidence to non-existent dispute
	badEv := &domain.DisputeEvidence{
		DisputeID: uuid.New(),
		UserID:    uuid.New(),
		Type:      domain.DisputeEvidenceScreenshot,
		URL:       "https://example.com/screen.png",
	}
	if err := repo.AddEvidence(ctx, badEv); err != domain.ErrDisputeNotFound {
		t.Errorf("AddEvidence bad dispute: want ErrDisputeNotFound, got %v", err)
	}

	// Add second evidence
	ev2 := &domain.DisputeEvidence{
		DisputeID:   dispute.ID,
		UserID:      dispute.RespondentID,
		Type:        domain.DisputeEvidenceReceipt,
		URL:         "https://example.com/receipt.pdf",
		Description: "Payment receipt",
	}
	if err := repo.AddEvidence(ctx, ev2); err != nil {
		t.Fatalf("AddEvidence 2: %v", err)
	}

	// ListEvidence
	evidence, err := repo.ListEvidence(ctx, dispute.ID)
	if err != nil {
		t.Fatalf("ListEvidence: %v", err)
	}
	if len(evidence) != 2 {
		t.Fatalf("ListEvidence count = %d, want 2", len(evidence))
	}
	// Should be ordered by created_at ASC
	if evidence[0].Type != domain.DisputeEvidencePhoto {
		t.Errorf("First evidence type = %q, want %q", evidence[0].Type, domain.DisputeEvidencePhoto)
	}
	if evidence[1].Type != domain.DisputeEvidenceReceipt {
		t.Errorf("Second evidence type = %q, want %q", evidence[1].Type, domain.DisputeEvidenceReceipt)
	}

	// ListEvidence for non-existent dispute returns empty
	empty, err := repo.ListEvidence(ctx, uuid.New())
	if err != nil {
		t.Fatalf("ListEvidence empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("ListEvidence empty count = %d, want 0", len(empty))
	}
}

func TestDisputeRepo_ListAll(t *testing.T) {
	ctx := context.Background()
	repo := NewDisputeRepo()

	userID := uuid.New()
	otherUserID := uuid.New()

	disputes := []domain.Dispute{
		{
			BookingID:    uuid.New(),
			InitiatorID:  userID,
			RespondentID: otherUserID,
			Reason:       domain.DisputeReasonPoorQuality,
			Description:  "Poor quality",
			Status:       domain.DisputeStatusOpen,
		},
		{
			BookingID:    uuid.New(),
			InitiatorID:  otherUserID,
			RespondentID: userID,
			Reason:       domain.DisputeReasonBillingError,
			Description:  "Billing error",
			Status:       domain.DisputeStatusResolved,
		},
		{
			BookingID:    uuid.New(),
			InitiatorID:  uuid.New(),
			RespondentID: uuid.New(),
			Reason:       domain.DisputeReasonDamage,
			Description:  "Damage",
			Status:       domain.DisputeStatusOpen,
		},
	}
	for i := range disputes {
		if err := repo.Create(ctx, &disputes[i]); err != nil {
			t.Fatalf("Create dispute %d: %v", i, err)
		}
	}

	// List all without filters
	result, err := repo.ListAll(ctx, domain.DisputeFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", result.TotalCount)
	}

	// Filter by status
	openStatus := domain.DisputeStatusOpen
	result, err = repo.ListAll(ctx, domain.DisputeFilter{Status: &openStatus, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by status: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for open = %d, want 2", result.TotalCount)
	}

	// Filter by user (as initiator or respondent)
	result, err = repo.ListAll(ctx, domain.DisputeFilter{UserID: &userID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by user: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for user = %d, want 2", result.TotalCount)
	}

	// Pagination
	result, err = repo.ListAll(ctx, domain.DisputeFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListAll page 1: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items on page 1 = %d, want 2", len(result.Items))
	}
	if result.TotalPages != 2 {
		t.Errorf("TotalPages = %d, want 2", result.TotalPages)
	}

	result, err = repo.ListAll(ctx, domain.DisputeFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListAll page 2: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items on page 2 = %d, want 1", len(result.Items))
	}
}

func TestDisputeRepo_ListByUser(t *testing.T) {
	ctx := context.Background()
	repo := NewDisputeRepo()

	userID := uuid.New()
	otherID := uuid.New()

	// User as initiator
	d1 := &domain.Dispute{
		BookingID:    uuid.New(),
		InitiatorID:  userID,
		RespondentID: otherID,
		Reason:       domain.DisputeReasonOther,
		Description:  "Dispute 1",
		Status:       domain.DisputeStatusOpen,
	}
	// User as respondent
	d2 := &domain.Dispute{
		BookingID:    uuid.New(),
		InitiatorID:  otherID,
		RespondentID: userID,
		Reason:       domain.DisputeReasonSafetyIssue,
		Description:  "Dispute 2",
		Status:       domain.DisputeStatusOpen,
	}
	// Unrelated dispute
	d3 := &domain.Dispute{
		BookingID:    uuid.New(),
		InitiatorID:  uuid.New(),
		RespondentID: uuid.New(),
		Reason:       domain.DisputeReasonDamage,
		Description:  "Dispute 3",
		Status:       domain.DisputeStatusOpen,
	}
	for _, d := range []*domain.Dispute{d1, d2, d3} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	result, err := repo.ListByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}
