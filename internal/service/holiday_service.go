package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

// HolidayService manages platform-level holidays and per-bathhouse holiday multipliers.
type HolidayService interface {
	Create(ctx context.Context, holiday *domain.Holiday) error
	Update(ctx context.Context, holiday *domain.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error)
	ListAll(ctx context.Context) ([]domain.Holiday, error)
	ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error)
	IsHolidayDate(ctx context.Context, date time.Time, region string, bathhouseID uuid.UUID) (*domain.Holiday, float64, error)
	SetBathhouseMultiplier(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, multiplier float64) error
	GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error)
}

type holidayService struct {
	repo   repository.HolidayRepository
	access *AccessChecker
	logger *logger.Logger
}

func NewHolidayService(
	repo repository.HolidayRepository,
	access *AccessChecker,
	log *logger.Logger,
) HolidayService {
	return &holidayService{
		repo:   repo,
		access: access,
		logger: log,
	}
}

func (s *holidayService) Create(ctx context.Context, holiday *domain.Holiday) error {
	if err := holiday.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, holiday)
}

func (s *holidayService) Update(ctx context.Context, holiday *domain.Holiday) error {
	if err := holiday.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, holiday)
}

func (s *holidayService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *holidayService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *holidayService) ListAll(ctx context.Context) ([]domain.Holiday, error) {
	return s.repo.ListAll(ctx)
}

func (s *holidayService) ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error) {
	return s.repo.ListByRegion(ctx, region)
}

// IsHolidayDate checks if a date is a holiday and returns the holiday info and applicable multiplier.
// Returns (nil, 0, nil) if the date is not a holiday.
func (s *holidayService) IsHolidayDate(ctx context.Context, date time.Time, region string, bathhouseID uuid.UUID) (*domain.Holiday, float64, error) {
	holiday, err := s.repo.IsHoliday(ctx, date, region)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	multiplier, err := s.repo.GetBathhouseMultiplier(ctx, bathhouseID)
	if err != nil {
		return nil, 0, err
	}

	return holiday, multiplier, nil
}

func (s *holidayService) SetBathhouseMultiplier(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, multiplier float64) error {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}

	price := &domain.BathhouseHolidayPrice{
		BathhouseID: bathhouseID,
		Multiplier:  multiplier,
	}
	if err := price.Validate(); err != nil {
		return err
	}

	return s.repo.SetBathhouseMultiplier(ctx, bathhouseID, multiplier)
}

func (s *holidayService) GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error) {
	return s.repo.GetBathhouseMultiplier(ctx, bathhouseID)
}
