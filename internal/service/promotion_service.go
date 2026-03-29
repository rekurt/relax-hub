package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PromotionService interface {
	Create(ctx context.Context, promo *domain.Promotion) error
	Update(ctx context.Context, promo *domain.Promotion) error
	Pause(ctx context.Context, promotionID uuid.UUID) error
	Resume(ctx context.Context, promotionID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Promotion, error)
	GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error)
	RecordImpression(ctx context.Context, bathhouseID uuid.UUID) error
	RecordClick(ctx context.Context, bathhouseID uuid.UUID) error
	DeductDailyBudgets(ctx context.Context) error
}

type promotionService struct {
	promoRepo repository.PromotionRepository
	bhRepo    repository.BathhouseRepository
	walletSvc WalletService
	notifSvc  NotificationService
	log       *logger.Logger
}

func NewPromotionService(
	promoRepo repository.PromotionRepository,
	bhRepo repository.BathhouseRepository,
	walletSvc WalletService,
	notifSvc NotificationService,
	log *logger.Logger,
) PromotionService {
	return &promotionService{
		promoRepo: promoRepo,
		bhRepo:    bhRepo,
		walletSvc: walletSvc,
		notifSvc:  notifSvc,
		log:       log,
	}
}

func (s *promotionService) Create(ctx context.Context, promo *domain.Promotion) error {
	return s.promoRepo.Create(ctx, promo)
}

func (s *promotionService) Update(ctx context.Context, promo *domain.Promotion) error {
	return s.promoRepo.Update(ctx, promo)
}

func (s *promotionService) Pause(ctx context.Context, promotionID uuid.UUID) error {
	promo, err := s.promoRepo.GetByID(ctx, promotionID)
	if err != nil {
		return err
	}
	if promo.Status != domain.PromotionActive {
		return domain.ErrInvalidInput
	}
	promo.Status = domain.PromotionPaused
	promo.UpdatedAt = time.Now()
	return s.promoRepo.Update(ctx, promo)
}

func (s *promotionService) Resume(ctx context.Context, promotionID uuid.UUID) error {
	promo, err := s.promoRepo.GetByID(ctx, promotionID)
	if err != nil {
		return err
	}
	if promo.Status != domain.PromotionPaused {
		return domain.ErrInvalidInput
	}
	if promo.RemainingBudget() < promo.DailyBidKopecks {
		return domain.ErrPromotionBudgetExhausted
	}
	promo.Status = domain.PromotionActive
	promo.UpdatedAt = time.Now()
	return s.promoRepo.Update(ctx, promo)
}

func (s *promotionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Promotion, error) {
	return s.promoRepo.GetByID(ctx, id)
}

func (s *promotionService) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	return s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
}

func (s *promotionService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	return s.promoRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
}

func (s *promotionService) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	return s.promoRepo.ListByOwner(ctx, ownerID, page, pageSize)
}

func (s *promotionService) RecordImpression(ctx context.Context, bathhouseID uuid.UUID) error {
	promo, err := s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil || promo == nil {
		return nil
	}
	if err := s.promoRepo.RecordImpression(ctx, promo.ID); err != nil {
		s.log.Warn("failed to record promotion impression", "promotion_id", promo.ID, "error", err)
	}
	return nil
}

func (s *promotionService) RecordClick(ctx context.Context, bathhouseID uuid.UUID) error {
	promo, err := s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil || promo == nil {
		return nil
	}
	if err := s.promoRepo.RecordClick(ctx, promo.ID); err != nil {
		s.log.Warn("failed to record promotion click", "promotion_id", promo.ID, "error", err)
	}
	return nil
}

// DeductDailyBudgets charges the daily bid for all active promotions from owner wallets.
// Pauses campaigns that have insufficient budget or wallet balance.
// Expires campaigns past their end date.
func (s *promotionService) DeductDailyBudgets(ctx context.Context) error {
	promos, err := s.promoRepo.ListAllActive(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for i := range promos {
		promo := &promos[i]

		// Expire campaigns past end date
		if now.After(promo.EndDate) {
			promo.Status = domain.PromotionExpired
			promo.UpdatedAt = now
			if err := s.promoRepo.Update(ctx, promo); err != nil {
				s.log.Error("failed to expire promotion", "promotion_id", promo.ID, "error", err)
			}
			continue
		}

		// Check remaining budget
		remaining := promo.RemainingBudget()
		if remaining < promo.DailyBidKopecks {
			promo.Status = domain.PromotionExhausted
			promo.UpdatedAt = now
			if err := s.promoRepo.Update(ctx, promo); err != nil {
				s.log.Error("failed to exhaust promotion", "promotion_id", promo.ID, "error", err)
				continue
			}
			s.notifyBudgetExhausted(ctx, promo)
			continue
		}

		// Deduct daily bid from owner wallet
		bh, err := s.bhRepo.GetByID(ctx, promo.BathhouseID)
		if err != nil {
			s.log.Error("failed to get bathhouse for promotion deduction", "promotion_id", promo.ID, "error", err)
			continue
		}

		wallet, err := s.walletSvc.GetWallet(ctx, bh.OwnerID)
		if err != nil {
			s.log.Error("failed to get owner wallet for promotion", "owner_id", bh.OwnerID, "error", err)
			// Pause promotion if we can't charge
			promo.Status = domain.PromotionPaused
			promo.UpdatedAt = now
			_ = s.promoRepo.Update(ctx, promo)
			continue
		}

		promoID := promo.ID
		_, err = s.walletSvc.Spend(ctx, wallet.ID, promo.DailyBidKopecks, "promotion", &promoID, "Ежедневное списание за продвижение")
		if err != nil {
			s.log.Warn("insufficient wallet balance for promotion, pausing", "promotion_id", promo.ID, "owner_id", bh.OwnerID, "error", err)
			promo.Status = domain.PromotionPaused
			promo.UpdatedAt = now
			_ = s.promoRepo.Update(ctx, promo)
			s.notifyInsufficientBalance(ctx, promo, bh.OwnerID)
			continue
		}

		// Record spend
		promo.SpentKopecks += promo.DailyBidKopecks
		promo.UpdatedAt = now
		if err := s.promoRepo.Update(ctx, promo); err != nil {
			s.log.Error("failed to update promotion spend", "promotion_id", promo.ID, "error", err)
		}

		s.log.Info("promotion daily bid deducted", "promotion_id", promo.ID, "bid", promo.DailyBidKopecks, "spent", promo.SpentKopecks, "budget", promo.BudgetKopecks)
	}

	return nil
}

func (s *promotionService) notifyBudgetExhausted(ctx context.Context, promo *domain.Promotion) {
	bh, err := s.bhRepo.GetByID(ctx, promo.BathhouseID)
	if err != nil {
		return
	}
	_ = s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifPromo,
		"Бюджет продвижения исчерпан",
		"Бюджет рекламной кампании для объекта исчерпан. Пополните бюджет или создайте новую кампанию.",
		map[string]string{"bathhouse_id": promo.BathhouseID.String(), "promotion_id": promo.ID.String()},
	)
}

func (s *promotionService) notifyInsufficientBalance(ctx context.Context, promo *domain.Promotion, ownerID uuid.UUID) {
	_ = s.notifSvc.Send(ctx, ownerID, domain.NotifPromo,
		"Продвижение приостановлено",
		"Недостаточно средств на кошельке для продвижения. Пополните кошелёк и возобновите кампанию.",
		map[string]string{"bathhouse_id": promo.BathhouseID.String(), "promotion_id": promo.ID.String()},
	)
}
