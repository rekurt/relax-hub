package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

const maxAddOnsPerBathhouse = 20

type AddOnService interface {
	CreateAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addon *domain.AddOn) (*domain.AddOn, error)
	UpdateAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addon *domain.AddOn) (*domain.AddOn, error)
	DeleteAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addonID uuid.UUID) error
	ListAddOns(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	ListActiveAddOns(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	ListAddOnsManaged(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	GetAddOn(ctx context.Context, id uuid.UUID) (*domain.AddOn, error)
	CalculateAddOnTotal(ctx context.Context, selections []AddOnSelection, bathhouseID uuid.UUID, durationHours int, guests int) (int64, []AddOnLineItem, error)
}

type AddOnSelection struct {
	AddOnID  uuid.UUID
	Quantity int
}

type AddOnLineItem struct {
	AddOnID    uuid.UUID
	Name       string
	Quantity   int
	UnitPrice  int64
	TotalPrice int64
}

type addOnService struct {
	addonRepo repository.AddOnRepository
	access    *AccessChecker
	logger    *logger.Logger
}

func NewAddOnService(
	addonRepo repository.AddOnRepository,
	access *AccessChecker,
	log *logger.Logger,
) AddOnService {
	return &addOnService{
		addonRepo: addonRepo,
		access:    access,
		logger:    log,
	}
}

func (s *addOnService) CreateAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addon *domain.AddOn) (*domain.AddOn, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, addon.BathhouseID); err != nil {
		return nil, err
	}

	if err := addon.Validate(); err != nil {
		return nil, err
	}

	count, err := s.addonRepo.CountByBathhouse(ctx, addon.BathhouseID)
	if err != nil {
		return nil, err
	}
	if count >= maxAddOnsPerBathhouse {
		return nil, domain.ErrAddOnLimitReached
	}

	addon.ID = uuid.New()
	addon.IsActive = true
	now := time.Now()
	addon.CreatedAt = now
	addon.UpdatedAt = now

	if err := s.addonRepo.Create(ctx, addon); err != nil {
		return nil, err
	}

	s.logger.Info("add-on created", "addon_id", addon.ID, "bathhouse_id", addon.BathhouseID)
	return addon, nil
}

func (s *addOnService) UpdateAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addon *domain.AddOn) (*domain.AddOn, error) {
	existing, err := s.addonRepo.GetByID(ctx, addon.ID)
	if err != nil {
		return nil, err
	}

	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return nil, err
	}

	existing.Name = addon.Name
	existing.Description = addon.Description
	existing.Price = addon.Price
	existing.Unit = addon.Unit
	existing.IsActive = addon.IsActive
	existing.SortOrder = addon.SortOrder
	existing.UpdatedAt = time.Now()

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := s.addonRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.logger.Info("add-on updated", "addon_id", existing.ID)
	return existing, nil
}

func (s *addOnService) DeleteAddOn(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, addonID uuid.UUID) error {
	existing, err := s.addonRepo.GetByID(ctx, addonID)
	if err != nil {
		return err
	}

	if err := s.access.CanManageBathhouse(ctx, userID, userRole, existing.BathhouseID); err != nil {
		return err
	}

	if err := s.addonRepo.Delete(ctx, addonID); err != nil {
		return err
	}

	s.logger.Info("add-on deleted", "addon_id", addonID)
	return nil
}

func (s *addOnService) ListAddOns(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	return s.addonRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *addOnService) ListActiveAddOns(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	return s.addonRepo.ListActiveByBathhouse(ctx, bathhouseID)
}

func (s *addOnService) ListAddOnsManaged(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}
	return s.addonRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *addOnService) GetAddOn(ctx context.Context, id uuid.UUID) (*domain.AddOn, error) {
	return s.addonRepo.GetByID(ctx, id)
}

func (s *addOnService) CalculateAddOnTotal(ctx context.Context, selections []AddOnSelection, bathhouseID uuid.UUID, durationHours int, guests int) (int64, []AddOnLineItem, error) {
	var total int64
	items := make([]AddOnLineItem, 0, len(selections))

	for _, sel := range selections {
		if sel.Quantity <= 0 {
			return 0, nil, domain.ErrInvalidInput
		}

		addon, err := s.addonRepo.GetByID(ctx, sel.AddOnID)
		if err != nil {
			return 0, nil, err
		}

		if addon.BathhouseID != bathhouseID {
			return 0, nil, domain.ErrAddOnNotFound
		}

		if !addon.IsActive {
			return 0, nil, domain.ErrAddOnNotFound
		}

		var lineTotal int64
		switch addon.Unit {
		case domain.AddOnUnitPerItem:
			lineTotal = addon.Price * int64(sel.Quantity)
		case domain.AddOnUnitPerHour:
			lineTotal = addon.Price * int64(sel.Quantity) * int64(durationHours)
		case domain.AddOnUnitPerPerson:
			lineTotal = addon.Price * int64(sel.Quantity) * int64(guests)
		}

		items = append(items, AddOnLineItem{
			AddOnID:    addon.ID,
			Name:       addon.Name,
			Quantity:   sel.Quantity,
			UnitPrice:  addon.Price,
			TotalPrice: lineTotal,
		})
		total += lineTotal
	}

	return total, items, nil
}
