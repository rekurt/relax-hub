package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

const (
	autoHideReviewThreshold       = 3
	notifyAdminBathhouseThreshold = 5
)

type CreateComplaintInput struct {
	TargetType  domain.ComplaintTargetType
	TargetID    uuid.UUID
	Reason      domain.ComplaintReason
	Description string
}

type ComplaintService interface {
	Report(ctx context.Context, reporterID uuid.UUID, input CreateComplaintInput) (*domain.Complaint, error)
	Resolve(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID, resolution string) (*domain.Complaint, error)
	Dismiss(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID) (*domain.Complaint, error)
	List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error)
}

type complaintService struct {
	complaintRepo repository.ComplaintRepository
	reviewRepo    repository.ReviewRepository
	notifSvc      NotificationService
	logger        *logger.Logger
}

func NewComplaintService(
	complaintRepo repository.ComplaintRepository,
	reviewRepo repository.ReviewRepository,
	notifSvc NotificationService,
	log *logger.Logger,
) ComplaintService {
	return &complaintService{
		complaintRepo: complaintRepo,
		reviewRepo:    reviewRepo,
		notifSvc:      notifSvc,
		logger:        log,
	}
}

func (s *complaintService) Report(ctx context.Context, reporterID uuid.UUID, input CreateComplaintInput) (*domain.Complaint, error) {
	complaint := &domain.Complaint{
		ID:          uuid.New(),
		ReporterID:  reporterID,
		TargetType:  input.TargetType,
		TargetID:    input.TargetID,
		Reason:      input.Reason,
		Description: input.Description,
		Status:      domain.ComplaintStatusPending,
		CreatedAt:   time.Now(),
	}

	if err := complaint.Validate(); err != nil {
		return nil, err
	}

	// Prevent self-reporting
	if input.TargetType == domain.ComplaintTargetUser && input.TargetID == reporterID {
		return nil, domain.ErrInvalidInput
	}

	// Check if user already reported this target
	exists, err := s.complaintRepo.CheckExists(ctx, reporterID, input.TargetType, input.TargetID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrAlreadyReported
	}

	if err := s.complaintRepo.Create(ctx, complaint); err != nil {
		return nil, err
	}

	// Check auto-actions after creating the complaint
	s.checkAutoActions(ctx, input.TargetType, input.TargetID)

	return complaint, nil
}

func (s *complaintService) Resolve(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID, resolution string) (*domain.Complaint, error) {
	if strings.TrimSpace(resolution) == "" {
		return nil, domain.ErrInvalidInput
	}

	complaint, err := s.complaintRepo.GetByID(ctx, complaintID)
	if err != nil {
		return nil, err
	}

	if complaint.Status != domain.ComplaintStatusPending {
		return nil, domain.ErrInvalidInput
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaintID, domain.ComplaintStatusResolved, &adminID, resolution); err != nil {
		return nil, err
	}

	return s.complaintRepo.GetByID(ctx, complaintID)
}

func (s *complaintService) Dismiss(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID) (*domain.Complaint, error) {
	complaint, err := s.complaintRepo.GetByID(ctx, complaintID)
	if err != nil {
		return nil, err
	}

	if complaint.Status != domain.ComplaintStatusPending {
		return nil, domain.ErrInvalidInput
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaintID, domain.ComplaintStatusDismissed, &adminID, ""); err != nil {
		return nil, err
	}

	return s.complaintRepo.GetByID(ctx, complaintID)
}

func (s *complaintService) List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
	return s.complaintRepo.List(ctx, filter)
}

func (s *complaintService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
	return s.complaintRepo.GetByID(ctx, id)
}

func (s *complaintService) checkAutoActions(ctx context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) {
	count, err := s.complaintRepo.CountByTarget(ctx, targetType, targetID)
	if err != nil {
		s.logger.Error("failed to count complaints for auto-action", "target_type", targetType, "target_id", targetID, "error", err)
		return
	}

	switch targetType {
	case domain.ComplaintTargetReview:
		if count >= autoHideReviewThreshold {
			if err := s.reviewRepo.UpdateStatus(ctx, targetID, domain.ReviewStatusHidden); err != nil {
				s.logger.Error("failed to auto-hide review", "review_id", targetID, "error", err)
			} else {
				s.logger.Info("review auto-hidden due to complaints", "review_id", targetID, "complaint_count", count)
				// Notify the review author
				review, err := s.reviewRepo.GetByID(ctx, targetID)
				if err != nil {
					s.logger.Warn("failed to get review for hidden notification", "review_id", targetID, "error", err)
				} else {
					if err := s.notifSvc.Send(ctx, review.UserID, domain.NotifReviewHidden,
						"Отзыв скрыт",
						"Ваш отзыв был скрыт после проверки модераторами",
						map[string]string{"review_id": targetID.String()},
					); err != nil {
						s.logger.Warn("failed to send review hidden notification", "review_id", targetID, "error", err)
					}
				}
			}
		}
	case domain.ComplaintTargetBathhouse:
		if count == notifyAdminBathhouseThreshold {
			s.logger.Warn("bathhouse flagged for admin review due to complaints", "bathhouse_id", targetID, "complaint_count", count)
		}
	}
}
