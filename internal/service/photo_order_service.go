package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

// PhotoOrderService manages professional photography orders.
type PhotoOrderService interface {
	// Owner creates a photo order request.
	Create(ctx context.Context, ownerID uuid.UUID, order *domain.PhotoOrder) error
	// GetByID returns an order by ID with access control.
	GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.PhotoOrder, error)
	// ListByOwner lists orders for an owner.
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PhotoOrder], error)
	// AdminList lists all orders with filters.
	AdminList(ctx context.Context, filter domain.PhotoOrderFilter) (*domain.PaginatedResult[domain.PhotoOrder], error)
	// AdminUpdate allows admin to update order (assign photographer, confirm, complete, cancel).
	AdminUpdate(ctx context.Context, id uuid.UUID, update *PhotoOrderUpdate) (*domain.PhotoOrder, error)
	// OwnerCancel allows owner to cancel their own order (only if requested status).
	OwnerCancel(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error
}

// PhotoOrderUpdate holds fields the admin can change.
type PhotoOrderUpdate struct {
	Status           *domain.PhotoOrderStatus
	PhotographerName *string
	Price            *int64
	ScheduledAt      *string // RFC3339 or empty to clear
	AdminNotes       *string
}

type photoOrderService struct {
	repo      repository.PhotoOrderRepository
	walletSvc WalletService
	checker   *AccessChecker
	log       *logger.Logger
}

func NewPhotoOrderService(
	repo repository.PhotoOrderRepository,
	walletSvc WalletService,
	checker *AccessChecker,
	log *logger.Logger,
) PhotoOrderService {
	return &photoOrderService{
		repo:      repo,
		walletSvc: walletSvc,
		checker:   checker,
		log:       log,
	}
}

func (s *photoOrderService) Create(ctx context.Context, ownerID uuid.UUID, order *domain.PhotoOrder) error {
	order.OwnerID = ownerID
	order.Status = domain.PhotoOrderStatusRequested

	if err := order.Validate(); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return fmt.Errorf("create photo order: %w", err)
	}

	s.log.Info("photo order created", "order_id", order.ID, "owner_id", ownerID, "bathhouse_id", order.BathhouseID)
	return nil
}

func (s *photoOrderService) GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.PhotoOrder, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != domain.RoleAdmin && order.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	return order, nil
}

func (s *photoOrderService) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PhotoOrder], error) {
	return s.repo.List(ctx, domain.PhotoOrderFilter{
		OwnerID:  &ownerID,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *photoOrderService) AdminList(ctx context.Context, filter domain.PhotoOrderFilter) (*domain.PaginatedResult[domain.PhotoOrder], error) {
	return s.repo.List(ctx, filter)
}

func (s *photoOrderService) AdminUpdate(ctx context.Context, id uuid.UUID, update *PhotoOrderUpdate) (*domain.PhotoOrder, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if update.PhotographerName != nil {
		order.PhotographerName = *update.PhotographerName
	}
	if update.Price != nil {
		order.Price = *update.Price
	}
	if update.AdminNotes != nil {
		order.AdminNotes = *update.AdminNotes
	}
	if update.ScheduledAt != nil {
		if *update.ScheduledAt == "" {
			order.ScheduledAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *update.ScheduledAt)
			if err != nil {
				return nil, domain.ErrInvalidInput
			}
			order.ScheduledAt = &t
		}
	}

	if update.Status != nil && *update.Status != order.Status {
		if !isValidPhotoOrderTransition(order.Status, *update.Status) {
			return nil, domain.ErrPhotoOrderInvalidStatus
		}

		// On completion, charge owner wallet.
		if *update.Status == domain.PhotoOrderStatusCompleted {
			if order.Price > 0 {
				wallet, walletErr := s.walletSvc.GetWallet(ctx, order.OwnerID)
				if walletErr != nil {
					return nil, fmt.Errorf("get owner wallet for photo order payment: %w", walletErr)
				}

				orderID := order.ID
				_, spendErr := s.walletSvc.Spend(
					ctx,
					wallet.ID,
					order.Price,
					"photo_order",
					&orderID,
					fmt.Sprintf("Оплата за фотосъёмку #%s", orderID.String()[:8]),
				)
				if spendErr != nil {
					return nil, fmt.Errorf("charge owner wallet for photo order: %w", spendErr)
				}
				s.log.Info("photo order payment charged", "order_id", order.ID, "amount", order.Price)
			}
		}

		order.Status = *update.Status
	}

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("update photo order: %w", err)
	}

	s.log.Info("photo order updated by admin", "order_id", order.ID, "status", order.Status)
	return order, nil
}

func (s *photoOrderService) OwnerCancel(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if order.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	if order.Status != domain.PhotoOrderStatusRequested {
		return domain.ErrPhotoOrderInvalidStatus
	}

	order.Status = domain.PhotoOrderStatusCancelled
	if err := s.repo.Update(ctx, order); err != nil {
		return fmt.Errorf("cancel photo order: %w", err)
	}

	s.log.Info("photo order cancelled by owner", "order_id", order.ID, "owner_id", ownerID)
	return nil
}

// isValidPhotoOrderTransition checks allowed status transitions.
func isValidPhotoOrderTransition(from, to domain.PhotoOrderStatus) bool {
	switch from {
	case domain.PhotoOrderStatusRequested:
		return to == domain.PhotoOrderStatusConfirmed || to == domain.PhotoOrderStatusCancelled
	case domain.PhotoOrderStatusConfirmed:
		return to == domain.PhotoOrderStatusCompleted || to == domain.PhotoOrderStatusCancelled
	default:
		return false
	}
}
