package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const defaultKYCExpiryYears = 1

type SubmitKYCInput struct {
	EntityType   domain.KYCEntityType `json:"entity_type"`
	FullName     string               `json:"full_name"`
	INN          string               `json:"inn"`
	OGRNIP       string               `json:"ogrnip"`
	CompanyName  string               `json:"company_name"`
	DocumentURLs []string             `json:"document_urls"`
}

type KYCService interface {
	Submit(ctx context.Context, userID uuid.UUID, input SubmitKYCInput) (*domain.KYCApplication, error)
	GetStatus(ctx context.Context, userID uuid.UUID) (*domain.KYCApplication, error)
	Approve(ctx context.Context, kycID uuid.UUID, adminID uuid.UUID) error
	Reject(ctx context.Context, kycID uuid.UUID, adminID uuid.UUID, reason string) error
	ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.KYCApplication], error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KYCApplication, error)
	IsApproved(ctx context.Context, userID uuid.UUID) (bool, error)
}

type kycService struct {
	kycRepo  repository.KYCRepository
	notifSvc NotificationService
	logger   *logger.Logger
}

func NewKYCService(
	kycRepo repository.KYCRepository,
	notifSvc NotificationService,
	log *logger.Logger,
) KYCService {
	return &kycService{
		kycRepo:  kycRepo,
		notifSvc: notifSvc,
		logger:   log,
	}
}

func (s *kycService) Submit(ctx context.Context, userID uuid.UUID, input SubmitKYCInput) (*domain.KYCApplication, error) {
	// Check if user already has a pending or valid approved application
	existing, err := s.kycRepo.GetByUserID(ctx, userID)
	if err == nil {
		if existing.Status == domain.KYCStatusPending {
			return nil, domain.ErrKYCPending
		}
		// Block re-submission if user has a valid (non-expired) approved KYC
		if existing.Status == domain.KYCStatusApproved {
			if existing.ExpiresAt == nil || existing.ExpiresAt.After(time.Now()) {
				return nil, domain.ErrKYCPending // already approved, no need to re-submit
			}
		}
	}

	now := time.Now()
	app := &domain.KYCApplication{
		ID:           uuid.New(),
		UserID:       userID,
		Status:       domain.KYCStatusPending,
		EntityType:   input.EntityType,
		FullName:     input.FullName,
		INN:          input.INN,
		OGRNIP:       input.OGRNIP,
		CompanyName:  input.CompanyName,
		DocumentURLs: input.DocumentURLs,
		SubmittedAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := app.Validate(); err != nil {
		return nil, err
	}

	if err := s.kycRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	s.logger.Info("KYC application submitted", "user_id", userID, "kyc_id", app.ID)
	return app, nil
}

func (s *kycService) GetStatus(ctx context.Context, userID uuid.UUID) (*domain.KYCApplication, error) {
	return s.kycRepo.GetByUserID(ctx, userID)
}

func (s *kycService) GetByID(ctx context.Context, id uuid.UUID) (*domain.KYCApplication, error) {
	return s.kycRepo.GetByID(ctx, id)
}

func (s *kycService) Approve(ctx context.Context, kycID uuid.UUID, adminID uuid.UUID) error {
	expiresAt := time.Now().AddDate(defaultKYCExpiryYears, 0, 0)
	if err := s.kycRepo.Approve(ctx, kycID, adminID, expiresAt); err != nil {
		return err
	}

	// Fetch the application to notify the user
	app, err := s.kycRepo.GetByID(ctx, kycID)
	if err != nil {
		s.logger.Error("failed to fetch KYC after approval for notification", "kyc_id", kycID, "error", err)
		return nil // approval succeeded, notification is best-effort
	}

	s.notifSvc.Send(ctx, app.UserID, domain.NotifSystem, "KYC верификация одобрена", "Ваша заявка на KYC верификацию одобрена. Теперь вы можете создавать объекты.", nil)
	s.logger.Info("KYC application approved", "kyc_id", kycID, "admin_id", adminID)
	return nil
}

func (s *kycService) Reject(ctx context.Context, kycID uuid.UUID, adminID uuid.UUID, reason string) error {
	if reason == "" {
		return domain.ErrInvalidInput
	}

	if err := s.kycRepo.Reject(ctx, kycID, adminID, reason); err != nil {
		return err
	}

	app, err := s.kycRepo.GetByID(ctx, kycID)
	if err != nil {
		s.logger.Error("failed to fetch KYC after rejection for notification", "kyc_id", kycID, "error", err)
		return nil
	}

	s.notifSvc.Send(ctx, app.UserID, domain.NotifSystem, "KYC верификация отклонена", "Ваша заявка на KYC верификацию отклонена. Причина: "+reason, nil)
	s.logger.Info("KYC application rejected", "kyc_id", kycID, "admin_id", adminID, "reason", reason)
	return nil
}

func (s *kycService) ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.KYCApplication], error) {
	return s.kycRepo.ListPending(ctx, page, pageSize)
}

func (s *kycService) IsApproved(ctx context.Context, userID uuid.UUID) (bool, error) {
	app, err := s.kycRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	if app.Status != domain.KYCStatusApproved {
		return false, nil
	}

	// Check expiry
	if app.ExpiresAt != nil && app.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}
