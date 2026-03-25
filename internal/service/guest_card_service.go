package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type GuestCardService interface {
	RecordVisit(ctx context.Context, ownerID, clientID, bathhouseID uuid.UUID, amount int64) error
	ListGuests(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.GuestCardFilter) (*domain.PaginatedResult[domain.GuestCard], error)
	GetGuestDetail(ctx context.Context, userID uuid.UUID, role domain.UserRole, ownerID, clientID, bathhouseID uuid.UUID) (*domain.GuestCard, error)
	UpdateGuestNotes(ctx context.Context, userID uuid.UUID, role domain.UserRole, cardID uuid.UUID, notes string, tags []string) error
	ExportCSV(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.GuestCardFilter, w io.Writer) error
	GetStats(ctx context.Context, userID uuid.UUID, role domain.UserRole) (*domain.GuestCardStats, error)
}

type guestCardService struct {
	guestCardRepo repository.GuestCardRepository
	access        *AccessChecker
	logger        *logger.Logger
}

func NewGuestCardService(
	guestCardRepo repository.GuestCardRepository,
	access *AccessChecker,
	log *logger.Logger,
) GuestCardService {
	return &guestCardService{
		guestCardRepo: guestCardRepo,
		access:        access,
		logger:        log,
	}
}

func (s *guestCardService) RecordVisit(ctx context.Context, ownerID, clientID, bathhouseID uuid.UUID, amount int64) error {
	now := time.Now()
	card := &domain.GuestCard{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		ClientID:     clientID,
		BathhouseID:  bathhouseID,
		FirstVisitAt: now,
		LastVisitAt:  now,
		VisitCount:   1,
		TotalSpent:   amount,
		AvgCheck:     amount,
		Tags:         []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.guestCardRepo.Upsert(ctx, card); err != nil {
		return fmt.Errorf("record guest visit: %w", err)
	}
	return nil
}

func (s *guestCardService) ListGuests(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.GuestCardFilter) (*domain.PaginatedResult[domain.GuestCard], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	filter.OwnerID = userID
	return s.guestCardRepo.ListByOwner(ctx, filter)
}

func (s *guestCardService) GetGuestDetail(ctx context.Context, userID uuid.UUID, role domain.UserRole, ownerID, clientID, bathhouseID uuid.UUID) (*domain.GuestCard, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	return s.guestCardRepo.GetByOwnerAndClient(ctx, ownerID, clientID, bathhouseID)
}

func (s *guestCardService) UpdateGuestNotes(ctx context.Context, userID uuid.UUID, role domain.UserRole, cardID uuid.UUID, notes string, tags []string) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if tags == nil {
		tags = []string{}
	}

	return s.guestCardRepo.UpdateNotes(ctx, cardID, notes, tags)
}

func (s *guestCardService) ExportCSV(ctx context.Context, userID uuid.UUID, role domain.UserRole, filter domain.GuestCardFilter, w io.Writer) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	// Fetch all guests (large page to get all)
	filter.OwnerID = userID
	filter.Page = 1
	filter.PageSize = 10000

	result, err := s.guestCardRepo.ListByOwner(ctx, filter)
	if err != nil {
		return fmt.Errorf("export guest cards: %w", err)
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Header
	if err := csvWriter.Write([]string{
		"client_id", "bathhouse_id", "first_visit", "last_visit",
		"visit_count", "total_spent_rub", "avg_check_rub", "tags", "notes",
	}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, card := range result.Items {
		tags := ""
		for i, t := range card.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += t
		}
		if err := csvWriter.Write([]string{
			card.ClientID.String(),
			card.BathhouseID.String(),
			card.FirstVisitAt.Format("2006-01-02"),
			card.LastVisitAt.Format("2006-01-02"),
			fmt.Sprintf("%d", card.VisitCount),
			fmt.Sprintf("%.2f", float64(card.TotalSpent)/100.0),
			fmt.Sprintf("%.2f", float64(card.AvgCheck)/100.0),
			tags,
			card.Notes,
		}); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	return nil
}

func (s *guestCardService) GetStats(ctx context.Context, userID uuid.UUID, role domain.UserRole) (*domain.GuestCardStats, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	return s.guestCardRepo.GetStats(ctx, userID)
}
