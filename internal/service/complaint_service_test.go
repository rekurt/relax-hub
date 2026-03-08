package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type complaintTestEnv struct {
	svc        service.ComplaintService
	reviewRepo *mock.ReviewRepo
}

func newComplaintTestEnv() *complaintTestEnv {
	complaintRepo := mock.NewComplaintRepo()
	reviewRepo := mock.NewReviewRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewComplaintService(complaintRepo, reviewRepo, log)
	return &complaintTestEnv{
		svc:        svc,
		reviewRepo: reviewRepo,
	}
}

func TestComplaintService_Report_Success(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	targetID := uuid.New()

	complaint, err := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType:  domain.ComplaintTargetReview,
		TargetID:    targetID,
		Reason:      domain.ComplaintReasonSpam,
		Description: "This review is spam",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if complaint.ReporterID != reporterID {
		t.Errorf("reporter_id = %v, want %v", complaint.ReporterID, reporterID)
	}
	if complaint.TargetType != domain.ComplaintTargetReview {
		t.Errorf("target_type = %v, want review", complaint.TargetType)
	}
	if complaint.Status != domain.ComplaintStatusPending {
		t.Errorf("status = %v, want pending", complaint.Status)
	}
}

func TestComplaintService_Report_DuplicateError(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	targetID := uuid.New()

	input := service.CreateComplaintInput{
		TargetType:  domain.ComplaintTargetReview,
		TargetID:    targetID,
		Reason:      domain.ComplaintReasonSpam,
		Description: "Spam",
	}

	_, err := env.svc.Report(context.Background(), reporterID, input)
	if err != nil {
		t.Fatalf("first report: %v", err)
	}

	_, err = env.svc.Report(context.Background(), reporterID, input)
	if err != domain.ErrAlreadyReported {
		t.Errorf("err = %v, want ErrAlreadyReported", err)
	}
}

func TestComplaintService_Report_InvalidInput(t *testing.T) {
	env := newComplaintTestEnv()

	_, err := env.svc.Report(context.Background(), uuid.Nil, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonSpam,
	})
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestComplaintService_Resolve_Success(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	adminID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonFake,
	})

	resolved, err := env.svc.Resolve(context.Background(), complaint.ID, adminID, "Review removed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Status != domain.ComplaintStatusResolved {
		t.Errorf("status = %v, want resolved", resolved.Status)
	}
	if resolved.Resolution != "Review removed" {
		t.Errorf("resolution = %q, want %q", resolved.Resolution, "Review removed")
	}
	if resolved.ResolvedByID == nil || *resolved.ResolvedByID != adminID {
		t.Errorf("resolved_by_id = %v, want %v", resolved.ResolvedByID, adminID)
	}
}

func TestComplaintService_Resolve_NotPending(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	adminID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonSpam,
	})

	_, _ = env.svc.Resolve(context.Background(), complaint.ID, adminID, "Done")

	_, err := env.svc.Resolve(context.Background(), complaint.ID, adminID, "Again")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestComplaintService_Resolve_EmptyResolution(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	adminID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonSpam,
	})

	_, err := env.svc.Resolve(context.Background(), complaint.ID, adminID, "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for empty resolution", err)
	}

	_, err = env.svc.Resolve(context.Background(), complaint.ID, adminID, "   ")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for whitespace-only resolution", err)
	}
}

func TestComplaintService_Resolve_NotFound(t *testing.T) {
	env := newComplaintTestEnv()

	_, err := env.svc.Resolve(context.Background(), uuid.New(), uuid.New(), "Done")
	if err != domain.ErrComplaintNotFound {
		t.Errorf("err = %v, want ErrComplaintNotFound", err)
	}
}

func TestComplaintService_Dismiss_Success(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	adminID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetUser,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonOther,
	})

	dismissed, err := env.svc.Dismiss(context.Background(), complaint.ID, adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dismissed.Status != domain.ComplaintStatusDismissed {
		t.Errorf("status = %v, want dismissed", dismissed.Status)
	}
}

func TestComplaintService_Dismiss_NotPending(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()
	adminID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonSpam,
	})

	_, _ = env.svc.Dismiss(context.Background(), complaint.ID, adminID)

	_, err := env.svc.Dismiss(context.Background(), complaint.ID, adminID)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestComplaintService_Dismiss_NotFound(t *testing.T) {
	env := newComplaintTestEnv()

	_, err := env.svc.Dismiss(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrComplaintNotFound {
		t.Errorf("err = %v, want ErrComplaintNotFound", err)
	}
}

func TestComplaintService_GetByID(t *testing.T) {
	env := newComplaintTestEnv()
	reporterID := uuid.New()

	complaint, _ := env.svc.Report(context.Background(), reporterID, service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetBathhouse,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonFraud,
	})

	found, err := env.svc.GetByID(context.Background(), complaint.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != complaint.ID {
		t.Errorf("id = %v, want %v", found.ID, complaint.ID)
	}
}

func TestComplaintService_GetByID_NotFound(t *testing.T) {
	env := newComplaintTestEnv()

	_, err := env.svc.GetByID(context.Background(), uuid.New())
	if err != domain.ErrComplaintNotFound {
		t.Errorf("err = %v, want ErrComplaintNotFound", err)
	}
}

func TestComplaintService_List(t *testing.T) {
	env := newComplaintTestEnv()

	for i := 0; i < 3; i++ {
		_, _ = env.svc.Report(context.Background(), uuid.New(), service.CreateComplaintInput{
			TargetType: domain.ComplaintTargetReview,
			TargetID:   uuid.New(),
			Reason:     domain.ComplaintReasonSpam,
		})
	}

	result, err := env.svc.List(context.Background(), domain.ComplaintFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
}

func TestComplaintService_List_WithFilter(t *testing.T) {
	env := newComplaintTestEnv()
	adminID := uuid.New()

	// Create 2 pending and 1 resolved
	for i := 0; i < 2; i++ {
		_, _ = env.svc.Report(context.Background(), uuid.New(), service.CreateComplaintInput{
			TargetType: domain.ComplaintTargetReview,
			TargetID:   uuid.New(),
			Reason:     domain.ComplaintReasonSpam,
		})
	}
	complaint, _ := env.svc.Report(context.Background(), uuid.New(), service.CreateComplaintInput{
		TargetType: domain.ComplaintTargetBathhouse,
		TargetID:   uuid.New(),
		Reason:     domain.ComplaintReasonFraud,
	})
	_, _ = env.svc.Resolve(context.Background(), complaint.ID, adminID, "Done")

	pending := domain.ComplaintStatusPending
	result, err := env.svc.List(context.Background(), domain.ComplaintFilter{
		Status:   &pending,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestComplaintService_AutoHideReview(t *testing.T) {
	env := newComplaintTestEnv()
	reviewID := uuid.New()

	// Create a review in the mock repo so auto-hide can update it
	review := &domain.Review{
		ID:          reviewID,
		UserID:      uuid.New(),
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      5,
		Text:        "Test review",
		Status:      domain.ReviewStatusApproved,
	}
	if err := env.reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	// File 3 complaints from different users on the same review
	for i := 0; i < 3; i++ {
		_, err := env.svc.Report(context.Background(), uuid.New(), service.CreateComplaintInput{
			TargetType: domain.ComplaintTargetReview,
			TargetID:   reviewID,
			Reason:     domain.ComplaintReasonFake,
		})
		if err != nil {
			t.Fatalf("report %d: %v", i, err)
		}
	}

	// Verify the review was auto-hidden
	updated, err := env.reviewRepo.GetByID(context.Background(), reviewID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if updated.Status != domain.ReviewStatusHidden {
		t.Errorf("review status = %v, want hidden", updated.Status)
	}
}

func TestComplaintService_NoAutoHideBeforeThreshold(t *testing.T) {
	env := newComplaintTestEnv()
	reviewID := uuid.New()

	// Create a review
	review := &domain.Review{
		ID:          reviewID,
		UserID:      uuid.New(),
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      4,
		Text:        "Good place",
		Status:      domain.ReviewStatusApproved,
	}
	if err := env.reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	// File only 2 complaints (below threshold)
	for i := 0; i < 2; i++ {
		_, err := env.svc.Report(context.Background(), uuid.New(), service.CreateComplaintInput{
			TargetType: domain.ComplaintTargetReview,
			TargetID:   reviewID,
			Reason:     domain.ComplaintReasonSpam,
		})
		if err != nil {
			t.Fatalf("report %d: %v", i, err)
		}
	}

	// Review should still be approved
	updated, err := env.reviewRepo.GetByID(context.Background(), reviewID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if updated.Status != domain.ReviewStatusApproved {
		t.Errorf("review status = %v, want approved (should not be hidden yet)", updated.Status)
	}
}
