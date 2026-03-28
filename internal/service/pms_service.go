package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/pms"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PMSService interface {
	Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, conn *domain.PMSConnection) error
	GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.PMSConnection, error)
	Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, conn *domain.PMSConnection) error
	Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.PMSConnection], error)
	TestConnection(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	TriggerSync(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	ListSyncLogs(ctx context.Context, userID uuid.UUID, role domain.UserRole, connID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSSyncLog], error)
	SyncAllActive(ctx context.Context) error
}

type pmsService struct {
	connRepo      repository.PMSConnectionRepository
	syncLogRepo   repository.PMSSyncLogRepository
	accessChecker *AccessChecker
	registry      *pms.ProviderRegistry
	logger        *logger.Logger
}

func NewPMSService(
	connRepo repository.PMSConnectionRepository,
	syncLogRepo repository.PMSSyncLogRepository,
	accessChecker *AccessChecker,
	registry *pms.ProviderRegistry,
	log *logger.Logger,
) PMSService {
	return &pmsService{
		connRepo:      connRepo,
		syncLogRepo:   syncLogRepo,
		accessChecker: accessChecker,
		registry:      registry,
		logger:        log,
	}
}

func (s *pmsService) Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, conn *domain.PMSConnection) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return err
	}

	conn.OwnerID = userID
	if err := conn.Validate(); err != nil {
		return err
	}

	// Check if connection already exists for this bathhouse
	existing, err := s.connRepo.GetByBathhouseID(ctx, conn.BathhouseID)
	if err == nil && existing != nil {
		return domain.ErrPMSConnectionAlreadyExists
	}

	// Verify provider exists
	provider := s.registry.Get(conn.Provider)
	if provider == nil {
		return domain.ErrInvalidInput
	}

	conn.Status = domain.PMSConnectionActive
	return s.connRepo.Create(ctx, conn)
}

func (s *pmsService) GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.PMSConnection, error) {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.accessChecker.CanViewBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *pmsService) Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, conn *domain.PMSConnection) error {
	existing, err := s.connRepo.GetByID(ctx, conn.ID)
	if err != nil {
		return err
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, role, existing.BathhouseID); err != nil {
		return err
	}

	existing.Provider = conn.Provider
	existing.CredentialsEncrypted = conn.CredentialsEncrypted
	existing.SyncDirection = conn.SyncDirection
	existing.SyncIntervalMinutes = conn.SyncIntervalMinutes
	existing.ExternalID = conn.ExternalID
	existing.Status = conn.Status

	if err := existing.Validate(); err != nil {
		return err
	}

	provider := s.registry.Get(existing.Provider)
	if provider == nil {
		return domain.ErrInvalidInput
	}

	return s.connRepo.Update(ctx, existing)
}

func (s *pmsService) Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return err
	}

	return s.connRepo.Delete(ctx, id)
}

func (s *pmsService) List(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.PMSConnection], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	return s.connRepo.ListByOwner(ctx, userID, page, pageSize)
}

func (s *pmsService) TestConnection(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return err
	}

	provider := s.registry.Get(conn.Provider)
	if provider == nil {
		return domain.ErrInvalidInput
	}

	if err := provider.TestConnection(ctx, conn.CredentialsEncrypted); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrPMSSyncFailed, err.Error())
	}

	return nil
}

func (s *pmsService) TriggerSync(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	conn, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return err
	}

	return s.syncConnection(ctx, conn)
}

func (s *pmsService) ListSyncLogs(ctx context.Context, userID uuid.UUID, role domain.UserRole, connID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSSyncLog], error) {
	conn, err := s.connRepo.GetByID(ctx, connID)
	if err != nil {
		return nil, err
	}

	if err := s.accessChecker.CanViewBathhouse(ctx, userID, role, conn.BathhouseID); err != nil {
		return nil, err
	}

	return s.syncLogRepo.ListByConnection(ctx, connID, page, pageSize)
}

func (s *pmsService) SyncAllActive(ctx context.Context) error {
	connections, err := s.connRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active pms connections: %w", err)
	}

	for _, conn := range connections {
		if err := s.syncConnection(ctx, &conn); err != nil {
			s.logger.Error("pms sync failed",
				"connection_id", conn.ID,
				"provider", conn.Provider,
				"bathhouse_id", conn.BathhouseID,
				"error", err)
		}
	}

	return nil
}

func (s *pmsService) syncConnection(ctx context.Context, conn *domain.PMSConnection) error {
	provider := s.registry.Get(conn.Provider)
	if provider == nil {
		return fmt.Errorf("unknown pms provider: %s", conn.Provider)
	}

	syncLog := &domain.PMSSyncLog{
		ConnectionID: conn.ID,
		Direction:    conn.SyncDirection,
		StartedAt:    time.Now(),
	}

	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now.Add(30 * 24 * time.Hour)

	var totalSynced int
	var syncErr error

	// Inbound: pull bookings from PMS
	if conn.SyncDirection == domain.PMSSyncDirectionInbound || conn.SyncDirection == domain.PMSSyncDirectionBoth {
		bookings, err := provider.PullBookings(ctx, conn.CredentialsEncrypted, conn.ExternalID, from, to)
		if err != nil {
			syncErr = fmt.Errorf("pull bookings: %w", err)
		} else {
			totalSynced += len(bookings)
			s.logger.Info("pulled bookings from pms",
				"connection_id", conn.ID,
				"count", len(bookings))
		}
	}

	// Outbound: push is handled per-booking at booking creation time, not in bulk sync
	// The cron job only pulls inbound data

	syncLog.CompletedAt = time.Now()
	syncLog.ItemsSynced = totalSynced

	if syncErr != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = syncErr.Error()
		_ = s.syncLogRepo.Create(ctx, syncLog)
		_ = s.connRepo.UpdateSyncStatus(ctx, conn.ID, now, syncErr.Error(), domain.PMSConnectionError)
		return syncErr
	}

	syncLog.Status = "success"
	_ = s.syncLogRepo.Create(ctx, syncLog)
	_ = s.connRepo.UpdateSyncStatus(ctx, conn.ID, now, "", domain.PMSConnectionActive)

	return nil
}
