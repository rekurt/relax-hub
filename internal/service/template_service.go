package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type TemplateService interface {
	Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, template *domain.ResponseTemplate) error
	Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, template *domain.ResponseTemplate) error
	Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.ResponseTemplate, error)
}

type templateService struct {
	templateRepo repository.ResponseTemplateRepository
	logger       *logger.Logger
}

func NewTemplateService(
	templateRepo repository.ResponseTemplateRepository,
	log *logger.Logger,
) TemplateService {
	return &templateService{
		templateRepo: templateRepo,
		logger:       log,
	}
}

func (s *templateService) Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, template *domain.ResponseTemplate) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	template.OwnerID = userID

	if err := template.Validate(); err != nil {
		return err
	}

	count, err := s.templateRepo.CountByOwner(ctx, userID)
	if err != nil {
		return fmt.Errorf("count templates: %w", err)
	}
	if count >= int64(domain.MaxTemplatesPerOwner) {
		return domain.ErrTemplateLimitReached
	}

	return s.templateRepo.Create(ctx, template)
}

func (s *templateService) Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, template *domain.ResponseTemplate) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.templateRepo.GetByID(ctx, template.ID)
	if err != nil {
		return err
	}

	if existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	existing.Title = template.Title
	existing.Body = template.Body
	existing.SortOrder = template.SortOrder

	if err := existing.Validate(); err != nil {
		return err
	}

	return s.templateRepo.Update(ctx, existing)
}

func (s *templateService) Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	return s.templateRepo.Delete(ctx, id)
}

func (s *templateService) List(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.ResponseTemplate, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	templates, err := s.templateRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Seed defaults on first use
	if len(templates) == 0 {
		if err := s.seedDefaults(ctx, userID); err != nil {
			s.logger.Error("seed default templates", "owner_id", userID, "error", err)
			return []domain.ResponseTemplate{}, nil
		}
		templates, err = s.templateRepo.ListByOwner(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

func (s *templateService) seedDefaults(ctx context.Context, ownerID uuid.UUID) error {
	for i, dt := range domain.DefaultTemplates {
		t := &domain.ResponseTemplate{
			ID:        uuid.New(),
			OwnerID:   ownerID,
			Title:     dt.Title,
			Body:      dt.Body,
			IsDefault: true,
			SortOrder: i,
		}
		if err := s.templateRepo.Create(ctx, t); err != nil {
			return fmt.Errorf("seed template %d: %w", i, err)
		}
	}
	return nil
}
