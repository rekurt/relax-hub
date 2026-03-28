package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/sms"
)

const (
	broadcastWeeklyLimit    = 3
	broadcastGuestCooldownH = 72 // 3 days in hours
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
	userRepo      repository.UserRepository
	promoRepo     repository.PromoCodeRepository
	notifSvc      NotificationService
	smsProvider   sms.Provider
	access        *AccessChecker
	logger        *logger.Logger
}

func NewBroadcastService(
	broadcastRepo repository.BroadcastRepository,
	guestCardRepo repository.GuestCardRepository,
	userRepo repository.UserRepository,
	promoRepo repository.PromoCodeRepository,
	notifSvc NotificationService,
	smsProvider sms.Provider,
	access *AccessChecker,
	log *logger.Logger,
) BroadcastService {
	return &broadcastService{
		broadcastRepo: broadcastRepo,
		guestCardRepo: guestCardRepo,
		userRepo:      userRepo,
		promoRepo:     promoRepo,
		notifSvc:      notifSvc,
		smsProvider:   smsProvider,
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

	// Verify ownership: owner must match, admins can send any broadcast
	if broadcast.OwnerID != userID && role != domain.RoleAdmin {
		return domain.ErrForbidden
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
	// Scope guest card query by role
	switch role {
	case domain.RoleRepresentative:
		bhIDs, bhErr := s.access.GetManagedBathhouseIDs(ctx, userID)
		if bhErr != nil {
			_ = s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusFailed)
			return fmt.Errorf("get managed bathhouses: %w", bhErr)
		}
		filter.BathhouseIDs = bhIDs
	case domain.RoleAdmin:
		filter.NoOwnerFilter = true
	default:
		filter.OwnerID = broadcast.OwnerID
	}
	guests, err := s.guestCardRepo.ListByOwner(ctx, filter)
	if err != nil {
		_ = s.broadcastRepo.UpdateStatus(ctx, broadcastID, domain.BroadcastStatusFailed)
		return fmt.Errorf("list guests for broadcast: %w", err)
	}

	// Resolve promo code string for personalization
	var promoCode string
	if broadcast.PromoCodeID != nil {
		promo, promoErr := s.promoRepo.GetByID(ctx, *broadcast.PromoCodeID)
		if promoErr != nil {
			s.logger.Warn("resolve promo code for broadcast", "promo_id", broadcast.PromoCodeID, "error", promoErr)
		} else {
			promoCode = promo.Code
		}
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

		// Build personalization data from guest card and user profile
		pData := s.buildPersonalizationData(ctx, &guest, promoCode)
		title := domain.PersonalizeMessage(broadcast.Title, pData)
		body := domain.PersonalizeMessage(broadcast.Body, pData)

		// Send via notification service (push, email, telegram, in-app)
		err = s.notifSvc.Send(ctx, guest.ClientID, domain.NotifBroadcast, title, body, map[string]string{
			"broadcast_id": broadcastID.String(),
			"owner_id":     userID.String(),
		})
		if err != nil {
			s.logger.Error("send broadcast notification", "client_id", guest.ClientID, "broadcast_id", broadcastID, "error", err)
			continue
		}

		// Send SMS if channel enabled
		if broadcast.HasChannel(domain.BroadcastChannelSMS) && s.smsProvider != nil {
			s.sendBroadcastSMS(ctx, guest.ClientID, body)
		}

		delivered++
	}

	// Update stats and mark as sent
	if err := s.broadcastRepo.UpdateStats(ctx, broadcastID, delivered, 0, 0); err != nil {
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

	// Admins can see all broadcasts; owners/reps see only their own
	if role != domain.RoleAdmin {
		filter.OwnerID = userID
	}
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

	// Admins can view any broadcast; owners/reps can only view their own
	if broadcast.OwnerID != userID && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	return broadcast, nil
}

// buildPersonalizationData constructs personalization data for a guest from their card and user profile.
func (s *broadcastService) buildPersonalizationData(ctx context.Context, guest *domain.GuestCard, promoCode string) domain.BroadcastPersonalizationData {
	data := domain.BroadcastPersonalizationData{
		VisitCount: guest.VisitCount,
		PromoCode:  promoCode,
	}

	// Format last visit date
	if !guest.LastVisitAt.IsZero() {
		data.LastVisitDate = guest.LastVisitAt.Format("02.01.2006")
	}

	// Get guest name from user profile
	user, err := s.userRepo.GetByID(ctx, guest.ClientID)
	if err != nil {
		s.logger.Warn("get user for broadcast personalization", "client_id", guest.ClientID, "error", err)
		data.GuestName = "Гость"
	} else {
		data.GuestName = user.Name
		if data.GuestName == "" {
			data.GuestName = "Гость"
		}
	}

	return data
}

// sendBroadcastSMS sends a broadcast message via SMS to a guest.
func (s *broadcastService) sendBroadcastSMS(ctx context.Context, clientID uuid.UUID, body string) {
	user, err := s.userRepo.GetByID(ctx, clientID)
	if err != nil {
		s.logger.Warn("get user phone for broadcast SMS", "client_id", clientID, "error", err)
		return
	}
	if user.Phone == "" {
		return
	}
	if err := s.smsProvider.SendSMS(ctx, user.Phone, body); err != nil {
		s.logger.Warn("send broadcast SMS", "client_id", clientID, "error", err)
	}
}
