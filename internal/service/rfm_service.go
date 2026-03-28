package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type RFMService interface {
	GetRFMAnalysis(ctx context.Context, userID uuid.UUID, role domain.UserRole) (*domain.RFMResult, error)
	ListCustomSegments(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.CustomSegment, error)
	CreateCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segment *domain.CustomSegment) error
	UpdateCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segment *domain.CustomSegment) error
	DeleteCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segmentID uuid.UUID) error
	GetCustomSegmentGuests(ctx context.Context, userID uuid.UUID, role domain.UserRole, segmentID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GuestCard], error)
}

type rfmService struct {
	guestCardRepo     repository.GuestCardRepository
	customSegmentRepo repository.CustomSegmentRepository
	access            *AccessChecker
	logger            *logger.Logger
}

func NewRFMService(
	guestCardRepo repository.GuestCardRepository,
	customSegmentRepo repository.CustomSegmentRepository,
	access *AccessChecker,
	log *logger.Logger,
) RFMService {
	return &rfmService{
		guestCardRepo:     guestCardRepo,
		customSegmentRepo: customSegmentRepo,
		access:            access,
		logger:            log,
	}
}

func (s *rfmService) GetRFMAnalysis(ctx context.Context, userID uuid.UUID, role domain.UserRole) (*domain.RFMResult, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	var filter domain.GuestCardFilter
	if err := setCRMFilter(ctx, s.access, userID, role, &filter); err != nil {
		return nil, err
	}

	result, err := s.guestCardRepo.GetRFMScores(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get rfm scores: %w", err)
	}
	return result, nil
}

func (s *rfmService) ListCustomSegments(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]domain.CustomSegment, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	filter := domain.CustomSegmentFilter{OwnerID: userID}
	segments, err := s.customSegmentRepo.ListByOwner(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list custom segments: %w", err)
	}

	// Enrich with guest counts
	var gcFilter domain.GuestCardFilter
	if err := setCRMFilter(ctx, s.access, userID, role, &gcFilter); err != nil {
		return nil, err
	}
	for i := range segments {
		count, err := s.customSegmentRepo.CountSegmentGuests(ctx, &segments[i], gcFilter)
		if err != nil {
			s.logger.Error("count custom segment guests", "segment_id", segments[i].ID, "error", err)
			continue
		}
		segments[i].GuestCount = count
	}

	return segments, nil
}

func (s *rfmService) CreateCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segment *domain.CustomSegment) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	segment.OwnerID = userID
	if err := segment.Validate(); err != nil {
		return err
	}

	return s.customSegmentRepo.Create(ctx, segment)
}

func (s *rfmService) UpdateCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segment *domain.CustomSegment) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.customSegmentRepo.GetByID(ctx, segment.ID)
	if err != nil {
		return err
	}
	if role != domain.RoleAdmin && existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	segment.OwnerID = existing.OwnerID
	if err := segment.Validate(); err != nil {
		return err
	}

	return s.customSegmentRepo.Update(ctx, segment)
}

func (s *rfmService) DeleteCustomSegment(ctx context.Context, userID uuid.UUID, role domain.UserRole, segmentID uuid.UUID) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.customSegmentRepo.GetByID(ctx, segmentID)
	if err != nil {
		return err
	}
	if role != domain.RoleAdmin && existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	return s.customSegmentRepo.Delete(ctx, segmentID)
}

func (s *rfmService) GetCustomSegmentGuests(ctx context.Context, userID uuid.UUID, role domain.UserRole, segmentID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GuestCard], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	segment, err := s.customSegmentRepo.GetByID(ctx, segmentID)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleAdmin && segment.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	var gcFilter domain.GuestCardFilter
	if err := setCRMFilter(ctx, s.access, userID, role, &gcFilter); err != nil {
		return nil, err
	}

	return s.customSegmentRepo.EvaluateSegment(ctx, segment, gcFilter, page, pageSize)
}

// setCRMFilter populates the GuestCardFilter based on user role (shared with guest card service).
func setCRMFilter(ctx context.Context, access *AccessChecker, userID uuid.UUID, role domain.UserRole, filter *domain.GuestCardFilter) error {
	switch role {
	case domain.RoleRepresentative:
		bhIDs, err := access.GetManagedBathhouseIDs(ctx, userID)
		if err != nil {
			return fmt.Errorf("get managed bathhouses: %w", err)
		}
		if len(bhIDs) == 0 {
			return domain.ErrForbidden
		}
		filter.BathhouseIDs = bhIDs
	case domain.RoleAdmin:
		filter.NoOwnerFilter = true
	default:
		filter.OwnerID = userID
	}
	return nil
}
