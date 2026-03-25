package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AuditLogService interface {
	LogChange(ctx context.Context, entityType string, entityID, userID uuid.UUID, action domain.AuditAction, changedFields map[string]interface{}) error
	GetHistory(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	ListAll(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
	ListAdminActions(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
	IsSubstantialChange(old, new *domain.Bathhouse) bool
}

type auditLogService struct {
	repo   repository.AuditLogRepository
	logger *logger.Logger
}

func NewAuditLogService(repo repository.AuditLogRepository, log *logger.Logger) AuditLogService {
	return &auditLogService{repo: repo, logger: log}
}

func (s *auditLogService) LogChange(ctx context.Context, entityType string, entityID, userID uuid.UUID, action domain.AuditAction, changedFields map[string]interface{}) error {
	var changedJSON json.RawMessage
	if changedFields != nil {
		data, err := json.Marshal(changedFields)
		if err != nil {
			s.logger.Error("failed to marshal changed fields", "error", err)
			return err
		}
		changedJSON = data
	}

	log := &domain.AuditLog{
		ID:            uuid.New(),
		EntityType:    entityType,
		EntityID:      entityID,
		UserID:        userID,
		Action:        action,
		ChangedFields: changedJSON,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.Error("failed to create audit log", "entity_type", entityType, "entity_id", entityID, "error", err)
		return err
	}

	return nil
}

func (s *auditLogService) GetHistory(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	return s.repo.ListByEntity(ctx, entityType, entityID, page, pageSize)
}

func (s *auditLogService) ListAll(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	return s.repo.List(ctx, filter)
}

// ListAdminActions returns audit log entries filtered to admin_action entity type.
// The filter.UserID is treated as admin_id for convenience.
func (s *auditLogService) ListAdminActions(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	entityType := "admin_action"
	filter.EntityType = &entityType
	return s.repo.List(ctx, filter)
}

// IsSubstantialChange checks if the bathhouse change is substantial enough
// to trigger automatic re-moderation. Substantial changes include:
// - Address change
// - City change
// - More than 50% of photos replaced
func (s *auditLogService) IsSubstantialChange(old, new *domain.Bathhouse) bool {
	if old.Address != new.Address {
		return true
	}
	if old.CityID != new.CityID {
		return true
	}

	// Check if >50% of photos were replaced
	if len(old.Images) > 0 {
		oldSet := make(map[string]bool, len(old.Images))
		for _, img := range old.Images {
			oldSet[img] = true
		}
		kept := 0
		for _, img := range new.Images {
			if oldSet[img] {
				kept++
			}
		}
		replacedRatio := 1.0 - float64(kept)/float64(len(old.Images))
		if replacedRatio > 0.5 {
			return true
		}
	}

	return false
}
