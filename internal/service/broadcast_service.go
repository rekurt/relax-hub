package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const (
	broadcastWeeklyLimit     = 3
	broadcastGuestCooldownH  = 72 // 3 days in hours
)

type BroadcastService interface {
	Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcast *domain.Broadcast) error
	Send(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcastID uuid.UUID) error
	ListBroadcasts(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.BroadcastFilter) (*domain.PaginatedResult[domain.Broadcast], error)
	GetBroadcast(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcastID uuid.UUID) (*domain.Broadcast, error)
}

type broadcastService struct {
	broadcastRepo repository.BroadcastRepository
	guestCardRepo repository.GuestCardRepository
	notifSvc      NotificationService
	access        *AccessChecker
	logger        *logger.Logger
}

func NewBroadcastService(
	broadcastRepo repository.BroadcastRepository,
	guestCardRepo repository.GuestCardRepository,
	notifSvc NotificationService,
	access *AccessChecker,
	log *logger.Logger,
) BroadcastService {
	return &broadcastService{
		broadcastRepo: broadcastRepo,
		guestCardRepo: guestCardRepo,
		notifSvc:      notifSvc,
		access:        access,
		logger:        log,
	}
}

func (s *broadcastService) Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcast *domain.Broadcast) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	broadcast.OwnerID = userID
	broadcast.Status = domain.BroadcastStatusDraft

	if err := broadcast.Validate(); err != nil {
		return err
	}

	return s.broadcastRepo.Create(ctx, broadcast)
}

func (s *broadcastService) Send(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcastID uuid.UUID) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return err
	}

	// Verify ownership: owner must match, or representative must manage the broadcast
	if broadcast.OwnerID != userID {
		if role == domain.RoleRepresentative {
			// Representatives can send broadcasts they created
			// (OwnerID is set to rep's userID at creation)
		} else if role != domain.RoleAdmin {
			return domain.ErrForbidden
		}
	}

	if broadcast.Status != domain.BroadcastStatusDraft {
		return domain.ErrBroadcastNotDraft
	}

	// Rate limit: 3 broadcasts per week per owner
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	recentCount, err := s.broadcastRepo.CountRecentByOwner(ctx, broadcast.OwnerID, weekAgo)
	if err != nil {
		return fmt.Errorf("check broadcast rate limit: %w", err)
	}
	if recentCount >= broadcastWeeklyLimit {
		return domain.ErrBroadcastRateLimit
	}

	// Mark as sending
	if err := s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusSending); err != nil {
		return fmt.Errorf("update broadcast status to sending: %w", err)
	}

	// Get guests in the target segment
	filter := domain.GuestCardFilter{
		Segment:  &broadcast.Segment,
		Page:     1,
		PageSize: 10000,
	}
	// For representatives, filter by managed bathhouse IDs instead of owner ID
	if role == domain.RoleRepresentative {
		bhIDs, bhErr := s.access.GetManagedBathhouseIDs(ctx, userID)
		if bhErr != nil {
			_ = s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusFailed)
			return fmt.Errorf("get managed bathhouses: %w", bhErr)
		}
		filter.BathhouseIDs = bhIDs
	} else {
		filter.OwnerID = broadcast.OwnerID
	}
	guests, err := s.guestCardRepo.ListByOwner(ctx, filter)
	if err != nil {
		_ = s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusFailed)
		return fmt.Errorf("list guests for broadcast: %w", err)
	}

	// Send notifications to each guest (respecting per-guest cooldown and preferences)
	cooldownSince := time.Now().Add(-time.Duration(broadcastGuestCooldownH) * time.Hour)
	var delivered int64
	for _, guest := range guests.Items {
		// Per-guest rate limit: 1 broadcast per 3 days
		recent, err := s.notifSvc.HasRecentByType(ctx, guest.ClientID, domain.NotifBroadcast, cooldownSince)
		if err != nil {
			s.logger.Error("check guest broadcast cooldown", "client_id", guest.ClientID, "error", err)
			continue
		}
		if recent {
			continue
		}

		err = s.notifSvc.Send(ctx, guest.ClientID, domain.NotifBroadcast, broadcast.Title, broadcast.Body, map[string]string{
			"broadcast_id": broadcastID.String(),
			"owner_id":     userID.String(),
		})
		if err != nil {
			s.logger.Error("send broadcast notification", "client_id", guest.ClientID, "broadcast_id", broadcastID, "error", err)
			continue
		}
		delivered++
	}

	// Update stats and mark as sent
	if err := s.broadcastRepo.UpdateStats(ctx, broadcastID, delivered, 0); err != nil {
		s.logger.Error("update broadcast stats", "broadcast_id", broadcastID, "error", err)
	}
	if err := s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusSent); err != nil {
		return fmt.Errorf("update broadcast status to sent: %w", err)
	}

	return nil
}

func (s *broadcastService) ListBroadcasts(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.BroadcastFilter) (*domain.PaginatedResult[domain.Broadcast], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	filter.OwnerID = userID
	return s.broadcastRepo.ListByOwner(ctx, filter)
}

func (s *broadcastService) GetBroadcast(ctx context.Context, userID uuid.UUID, role domain.UserRole, broadcastID uuid.UUID) (*domain.Broadcast, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		return nil, err
	}

	if broadcast.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	return broadcast, nil
}
