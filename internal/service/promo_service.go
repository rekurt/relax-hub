package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PromoService interface {
	Create(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error)
	Validate(ctx context.Context, code string, bathhouseID uuid.UUID, amount int64) (*domain.PromoCode, int64, error)
	Apply(ctx context.Context, userID uuid.UUID, code string, bookingID uuid.UUID, bathhouseID uuid.UUID, amount int64) (int64, error)
	RefundUsage(ctx context.Context, bookingID uuid.UUID) error
	Deactivate(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promoID uuid.UUID) error
	ListByBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error)
}

type promoService struct {
	promoRepo repository.PromoCodeRepository
	access    *AccessChecker
	logger    *logger.Logger
}

func NewPromoService(
	promoRepo repository.PromoCodeRepository,
	access *AccessChecker,
	log *logger.Logger,
) PromoService {
	return &promoService{
		promoRepo: promoRepo,
		access:    access,
		logger:    log,
	}
}

func (s *promoService) Create(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promo *domain.PromoCode) (*domain.PromoCode, error) {
	promo.Code = strings.ToUpper(strings.TrimSpace(promo.Code))
	promo.CreatorID = userID

	if promo.BathhouseID != nil {
		// Owner/representative creates promo for their bathhouse
		if err := s.access.CanManageBathhouse(ctx, userID, userRole, *promo.BathhouseID); err != nil {
			return nil, err
		}
	} else {
		// Only admin can create global promo codes
		if userRole != domain.RoleAdmin {
			return nil, domain.ErrForbidden
		}
	}

	if err := promo.Validate(); err != nil {
		return nil, err
	}

	promo.ID = uuid.New()
	promo.IsActive = true
	promo.CurrentUses = 0
	promo.CreatedAt = time.Now()

	if err := s.promoRepo.Create(ctx, promo); err != nil {
		return nil, err
	}

	s.logger.Info("promo code created", "promo_id", promo.ID, "code", promo.Code, "type", promo.Type)
	return promo, nil
}

func (s *promoService) Validate(ctx context.Context, code string, bathhouseID uuid.UUID, amount int64) (*domain.PromoCode, int64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, 0, domain.ErrPromoInvalid
	}

	promo, err := s.promoRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, err
	}

	if err := s.checkPromoValidity(promo, bathhouseID, amount); err != nil {
		return nil, 0, err
	}

	discount := s.calculateDiscount(promo, amount)
	return promo, discount, nil
}

func (s *promoService) Apply(ctx context.Context, userID uuid.UUID, code string, bookingID uuid.UUID, bathhouseID uuid.UUID, amount int64) (int64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return 0, domain.ErrPromoInvalid
	}

	promo, err := s.promoRepo.GetByCode(ctx, code)
	if err != nil {
		return 0, err
	}

	if err := s.checkPromoValidity(promo, bathhouseID, amount); err != nil {
		return 0, err
	}

	discount := s.calculateDiscount(promo, amount)

	usage := &domain.PromoUsage{
		ID:             uuid.New(),
		PromoCodeID:    promo.ID,
		UserID:         userID,
		BookingID:      bookingID,
		DiscountAmount: discount,
		UsedAt:         time.Now(),
	}
	if err := s.promoRepo.ApplyUsage(ctx, promo.ID, usage); err != nil {
		return 0, err
	}

	s.logger.Info("promo code applied", "promo_id", promo.ID, "booking_id", bookingID, "discount", discount)
	return discount, nil
}

func (s *promoService) RefundUsage(ctx context.Context, bookingID uuid.UUID) error {
	usage, err := s.promoRepo.GetUsageByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}
	if usage == nil {
		return nil
	}

	if err := s.promoRepo.DecrementUses(ctx, usage.PromoCodeID); err != nil {
		s.logger.Error("failed to decrement promo uses on refund", "promo_code_id", usage.PromoCodeID, "booking_id", bookingID, "error", err)
	}

	if err := s.promoRepo.DeleteUsage(ctx, usage.ID); err != nil {
		return err
	}

	s.logger.Info("promo usage refunded", "promo_code_id", usage.PromoCodeID, "booking_id", bookingID)
	return nil
}

func (s *promoService) Deactivate(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, promoID uuid.UUID) error {
	promo, err := s.promoRepo.GetByID(ctx, promoID)
	if err != nil {
		return err
	}

	if promo.BathhouseID != nil {
		if err := s.access.CanManageBathhouse(ctx, userID, userRole, *promo.BathhouseID); err != nil {
			return err
		}
	} else {
		if userRole != domain.RoleAdmin {
			return domain.ErrForbidden
		}
	}

	promo.IsActive = false
	if err := s.promoRepo.Update(ctx, promo); err != nil {
		return err
	}

	s.logger.Info("promo code deactivated", "promo_id", promoID)
	return nil
}

func (s *promoService) ListByBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}

	return s.promoRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
}

func (s *promoService) checkPromoValidity(promo *domain.PromoCode, bathhouseID uuid.UUID, amount int64) error {
	if !promo.IsActive {
		return domain.ErrPromoInvalid
	}

	now := time.Now()
	if now.Before(promo.ValidFrom) || now.After(promo.ValidUntil) {
		return domain.ErrPromoExpired
	}

	if promo.MaxUses > 0 && promo.CurrentUses >= promo.MaxUses {
		return domain.ErrPromoMaxUses
	}

	if promo.MinAmount > 0 && amount < promo.MinAmount {
		return domain.ErrPromoMinAmount
	}

	// If promo is bathhouse-specific, check it matches
	if promo.BathhouseID != nil && *promo.BathhouseID != bathhouseID {
		return domain.ErrPromoInvalid
	}

	return nil
}

func (s *promoService) calculateDiscount(promo *domain.PromoCode, amount int64) int64 {
	switch promo.Type {
	case domain.PromoTypePercentage:
		discount := amount * promo.Value / 100
		if discount > amount {
			discount = amount
		}
		return discount
	case domain.PromoTypeFixedAmount:
		if promo.Value > amount {
			return amount
		}
		return promo.Value
	case domain.PromoTypeFreeHour:
		// Value represents the hourly rate to subtract
		if promo.Value > amount {
			return amount
		}
		return promo.Value
	default:
		return 0
	}
}
