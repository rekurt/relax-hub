package mock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestComplaintRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewComplaintRepo()

	reporterID := uuid.New()
	targetID := uuid.New()

	complaint := &domain.Complaint{
		ReporterID:  reporterID,
		TargetType:  domain.ComplaintTargetReview,
		TargetID:    targetID,
		Reason:      domain.ComplaintReasonSpam,
		Description: "This is spam",
	}

	// Create
	if err := repo.Create(ctx, complaint); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if complaint.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}
	if complaint.Status != domain.ComplaintStatusPending {
		t.Errorf("Status = %q, want %q", complaint.Status, domain.ComplaintStatusPending)
	}

	// GetByID
	got, err := repo.GetByID(ctx, complaint.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ReporterID != reporterID {
		t.Errorf("ReporterID = %v, want %v", got.ReporterID, reporterID)
	}
	if got.Description != "This is spam" {
		t.Errorf("Description = %q, want %q", got.Description, "This is spam")
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, domain.ErrComplaintNotFound) {
		t.Errorf("GetByID not found: want ErrComplaintNotFound, got %v", err)
	}

	// Duplicate complaint (same reporter, same target)
	dup := &domain.Complaint{
		ReporterID: reporterID,
		TargetType: domain.ComplaintTargetReview,
		TargetID:   targetID,
		Reason:     domain.ComplaintReasonOffensive,
	}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyReported) {
		t.Errorf("Create duplicate: want ErrAlreadyReported, got %v", err)
	}

	// CheckExists
	exists, err := repo.CheckExists(ctx, reporterID, domain.ComplaintTargetReview, targetID)
	if err != nil {
		t.Fatalf("CheckExists: %v", err)
	}
	if !exists {
		t.Error("CheckExists should return true for existing complaint")
	}

	exists, err = repo.CheckExists(ctx, uuid.New(), domain.ComplaintTargetReview, targetID)
	if err != nil {
		t.Fatalf("CheckExists: %v", err)
	}
	if exists {
		t.Error("CheckExists should return false for non-existing complaint")
	}

	// CountByTarget
	count, err := repo.CountByTarget(ctx, domain.ComplaintTargetReview, targetID)
	if err != nil {
		t.Fatalf("CountByTarget: %v", err)
	}
	if count != 1 {
		t.Errorf("CountByTarget = %d, want 1", count)
	}

	// UpdateStatus (resolve)
	adminID := uuid.New()
	if err := repo.UpdateStatus(ctx, complaint.ID, domain.ComplaintStatusResolved, &adminID, "Spam confirmed"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, complaint.ID)
	if got.Status != domain.ComplaintStatusResolved {
		t.Errorf("Status = %q, want %q", got.Status, domain.ComplaintStatusResolved)
	}
	if got.Resolution != "Spam confirmed" {
		t.Errorf("Resolution = %q, want %q", got.Resolution, "Spam confirmed")
	}
	if got.ResolvedByID == nil || *got.ResolvedByID != adminID {
		t.Error("ResolvedByID should be set to admin ID")
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}

	// UpdateStatus not found
	if err := repo.UpdateStatus(ctx, uuid.New(), domain.ComplaintStatusDismissed, &adminID, ""); !errors.Is(err, domain.ErrComplaintNotFound) {
		t.Errorf("UpdateStatus not found: want ErrComplaintNotFound, got %v", err)
	}

	// CountByTarget after resolving (resolved complaints not counted)
	count, err = repo.CountByTarget(ctx, domain.ComplaintTargetReview, targetID)
	if err != nil {
		t.Fatalf("CountByTarget after resolve: %v", err)
	}
	if count != 0 {
		t.Errorf("CountByTarget after resolve = %d, want 0", count)
	}
}

func TestComplaintRepo_List(t *testing.T) {
	ctx := context.Background()
	repo := NewComplaintRepo()

	targetID := uuid.New()
	now := time.Now()

	// Create several complaints with different attributes
	complaints := []domain.Complaint{
		{
			ReporterID:  uuid.New(),
			TargetType:  domain.ComplaintTargetReview,
			TargetID:    targetID,
			Reason:      domain.ComplaintReasonSpam,
			Description: "spam 1",
			CreatedAt:   now.Add(-3 * time.Hour),
		},
		{
			ReporterID:  uuid.New(),
			TargetType:  domain.ComplaintTargetBathhouse,
			TargetID:    uuid.New(),
			Reason:      domain.ComplaintReasonFraud,
			Description: "fraud",
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ReporterID:  uuid.New(),
			TargetType:  domain.ComplaintTargetReview,
			TargetID:    uuid.New(),
			Reason:      domain.ComplaintReasonOffensive,
			Description: "offensive",
			CreatedAt:   now.Add(-1 * time.Hour),
		},
	}
	for i := range complaints {
		if err := repo.Create(ctx, &complaints[i]); err != nil {
			t.Fatalf("Create complaint %d: %v", i, err)
		}
	}

	// List all
	result, err := repo.List(ctx, domain.ComplaintFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", result.TotalCount)
	}
	// Verify DESC order (newest first)
	if len(result.Items) >= 2 && result.Items[0].CreatedAt.Before(result.Items[1].CreatedAt) {
		t.Error("List should return items in DESC order by created_at")
	}

	// Filter by target_type
	reviewType := domain.ComplaintTargetReview
	result, err = repo.List(ctx, domain.ComplaintFilter{TargetType: &reviewType, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List by target_type: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for reviews = %d, want 2", result.TotalCount)
	}

	// Filter by reason
	spamReason := domain.ComplaintReasonSpam
	result, err = repo.List(ctx, domain.ComplaintFilter{Reason: &spamReason, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List by reason: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount for spam = %d, want 1", result.TotalCount)
	}

	// Filter by date range
	fromDate := now.Add(-2*time.Hour - 30*time.Minute)
	result, err = repo.List(ctx, domain.ComplaintFilter{FromDate: &fromDate, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List by date: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for date filter = %d, want 2", result.TotalCount)
	}

	// Pagination
	result, err = repo.List(ctx, domain.ComplaintFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items on page 1 = %d, want 2", len(result.Items))
	}
	if result.TotalPages != 2 {
		t.Errorf("TotalPages = %d, want 2", result.TotalPages)
	}

	result, err = repo.List(ctx, domain.ComplaintFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items on page 2 = %d, want 1", len(result.Items))
	}

	// Filter by status
	pendingStatus := domain.ComplaintStatusPending
	result, err = repo.List(ctx, domain.ComplaintFilter{Status: &pendingStatus, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List by status: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount for pending = %d, want 3", result.TotalCount)
	}
}
