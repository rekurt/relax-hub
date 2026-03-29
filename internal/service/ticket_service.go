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

type TicketService interface {
	CreateTicket(ctx context.Context, userID uuid.UUID, ticket *domain.Ticket, initialMessage string) error
	GetTicket(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) (*domain.Ticket, error)
	ListUserTickets(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error)
	ListAllTickets(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error)
	AddMessage(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID, body string, attachments []string) (*domain.TicketMessage, error)
	ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) ([]domain.TicketMessage, error)
	AssignTicket(ctx context.Context, ticketID uuid.UUID, assignedTo uuid.UUID) error
	EscalateTicket(ctx context.Context, ticketID uuid.UUID) error
	ResolveTicket(ctx context.Context, ticketID uuid.UUID) error
	CloseTicket(ctx context.Context, ticketID uuid.UUID) error
	SubmitCSAT(ctx context.Context, userID uuid.UUID, ticketID uuid.UUID, score int) error
	GetStats(ctx context.Context) (*domain.TicketStatusCounts, error)
	GetOperationMetrics(ctx context.Context, filter domain.TicketMetricsFilter) (*domain.TicketOperationMetrics, error)
	AutoEscalateStaleTickets(ctx context.Context) error
	AutoCloseResolvedTickets(ctx context.Context) error
}

type ticketService struct {
	ticketRepo repository.TicketRepository
	logger     *logger.Logger
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	log *logger.Logger,
) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		logger:     log,
	}
}

func (s *ticketService) CreateTicket(ctx context.Context, userID uuid.UUID, ticket *domain.Ticket, initialMessage string) error {
	ticket.UserID = userID
	ticket.Status = domain.TicketStatusOpen
	ticket.Priority = domain.AutoPriority(ticket.Category)
	ticket.Level = domain.TicketLevelL1

	// Critical tickets auto-escalate to L2
	if ticket.Priority == domain.TicketPriorityCritical {
		ticket.Level = domain.TicketLevelL2
	}

	if err := ticket.Validate(); err != nil {
		return err
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}

	// Add initial message
	if initialMessage != "" {
		msg := &domain.TicketMessage{
			ID:         uuid.New(),
			TicketID:   ticket.ID,
			SenderID:   userID,
			SenderType: domain.TicketSenderUser,
			Body:       initialMessage,
		}
		if err := s.ticketRepo.AddMessage(ctx, msg); err != nil {
			return fmt.Errorf("add initial ticket message: %w", err)
		}
	}

	return nil
}

func (s *ticketService) GetTicket(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Non-admin users can only see their own tickets
	if role != domain.RoleAdmin && ticket.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return ticket, nil
}

func (s *ticketService) ListUserTickets(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error) {
	return s.ticketRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *ticketService) ListAllTickets(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
	return s.ticketRepo.ListAll(ctx, filter)
}

func (s *ticketService) AddMessage(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID, body string, attachments []string) (*domain.TicketMessage, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Non-admin users can only message their own tickets
	if role != domain.RoleAdmin && ticket.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if ticket.Status == domain.TicketStatusClosed {
		return nil, domain.ErrTicketAlreadyClosed
	}

	senderType := domain.TicketSenderUser
	if role == domain.RoleAdmin {
		senderType = domain.TicketSenderAdmin
	}

	msg := &domain.TicketMessage{
		ID:          uuid.New(),
		TicketID:    ticketID,
		SenderID:    userID,
		SenderType:  senderType,
		Body:        body,
		Attachments: attachments,
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if err := s.ticketRepo.AddMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	// Auto-transition to in_progress when admin replies to an open ticket
	if senderType == domain.TicketSenderAdmin && ticket.Status == domain.TicketStatusOpen {
		if err := s.ticketRepo.UpdateStatus(ctx, ticketID, domain.TicketStatusInProgress); err != nil {
			s.logger.Error("auto-update ticket status to in_progress", "ticket_id", ticketID, "error", err)
		}
	}

	return msg, nil
}

func (s *ticketService) ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if role != domain.RoleAdmin && ticket.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return s.ticketRepo.ListMessages(ctx, ticketID)
}

func (s *ticketService) AssignTicket(ctx context.Context, ticketID uuid.UUID, assignedTo uuid.UUID) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == domain.TicketStatusClosed {
		return domain.ErrTicketAlreadyClosed
	}

	return s.ticketRepo.Assign(ctx, ticketID, assignedTo)
}

func (s *ticketService) EscalateTicket(ctx context.Context, ticketID uuid.UUID) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == domain.TicketStatusClosed {
		return domain.ErrTicketAlreadyClosed
	}
	if ticket.Status == domain.TicketStatusResolved {
		return domain.ErrTicketAlreadyResolved
	}

	if ticket.Level == domain.TicketLevelL3 {
		return domain.ErrTicketAlreadyEscalated
	}

	nextLevel := domain.TicketLevelL2
	if ticket.Level == domain.TicketLevelL2 {
		nextLevel = domain.TicketLevelL3
	}

	if err := s.ticketRepo.UpdateLevel(ctx, ticketID, nextLevel); err != nil {
		return fmt.Errorf("escalate ticket: %w", err)
	}
	return s.ticketRepo.UpdateStatus(ctx, ticketID, domain.TicketStatusEscalated)
}

func (s *ticketService) ResolveTicket(ctx context.Context, ticketID uuid.UUID) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == domain.TicketStatusClosed {
		return domain.ErrTicketAlreadyClosed
	}
	if ticket.Status == domain.TicketStatusResolved {
		return domain.ErrTicketAlreadyResolved
	}

	return s.ticketRepo.Resolve(ctx, ticketID, time.Now())
}

func (s *ticketService) CloseTicket(ctx context.Context, ticketID uuid.UUID) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == domain.TicketStatusClosed {
		return domain.ErrTicketAlreadyClosed
	}

	return s.ticketRepo.UpdateStatus(ctx, ticketID, domain.TicketStatusClosed)
}

func (s *ticketService) SubmitCSAT(ctx context.Context, userID uuid.UUID, ticketID uuid.UUID, score int) error {
	if score < 1 || score > 5 {
		return domain.ErrCSATInvalidScore
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.UserID != userID {
		return domain.ErrForbidden
	}

	if ticket.Status != domain.TicketStatusResolved && ticket.Status != domain.TicketStatusClosed {
		return domain.ErrCSATNotResolved
	}

	if ticket.CSATScore != nil {
		return domain.ErrCSATAlreadySubmitted
	}

	return s.ticketRepo.SubmitCSAT(ctx, ticketID, score)
}

func (s *ticketService) GetStats(ctx context.Context) (*domain.TicketStatusCounts, error) {
	return s.ticketRepo.CountByStatus(ctx)
}

func (s *ticketService) GetOperationMetrics(ctx context.Context, filter domain.TicketMetricsFilter) (*domain.TicketOperationMetrics, error) {
	return s.ticketRepo.GetOperationMetrics(ctx, filter)
}

// AutoEscalateStaleTickets escalates tickets that have had no response:
// L1 with no update for 24h -> L2
// L2 with no update for 48h -> L3
func (s *ticketService) AutoEscalateStaleTickets(ctx context.Context) error {
	now := time.Now()

	// L1 -> L2 after 24h
	staleL1, err := s.ticketRepo.ListStaleTickets(ctx, domain.TicketLevelL1, now.Add(-24*time.Hour))
	if err != nil {
		return fmt.Errorf("list stale L1 tickets: %w", err)
	}
	for _, t := range staleL1 {
		if err := s.ticketRepo.UpdateLevel(ctx, t.ID, domain.TicketLevelL2); err != nil {
			s.logger.Error("auto-escalate L1->L2", "ticket_id", t.ID, "error", err)
			continue
		}
		if err := s.ticketRepo.UpdateStatus(ctx, t.ID, domain.TicketStatusEscalated); err != nil {
			s.logger.Error("auto-escalate L1->L2 status update", "ticket_id", t.ID, "error", err)
		}
	}

	// L2 -> L3 after 48h
	staleL2, err := s.ticketRepo.ListStaleTickets(ctx, domain.TicketLevelL2, now.Add(-48*time.Hour))
	if err != nil {
		return fmt.Errorf("list stale L2 tickets: %w", err)
	}
	for _, t := range staleL2 {
		if err := s.ticketRepo.UpdateLevel(ctx, t.ID, domain.TicketLevelL3); err != nil {
			s.logger.Error("auto-escalate L2->L3", "ticket_id", t.ID, "error", err)
			continue
		}
		if err := s.ticketRepo.UpdateStatus(ctx, t.ID, domain.TicketStatusEscalated); err != nil {
			s.logger.Error("auto-escalate L2->L3 status update", "ticket_id", t.ID, "error", err)
		}
	}

	if len(staleL1)+len(staleL2) > 0 {
		s.logger.Info("auto-escalated tickets", "l1_to_l2", len(staleL1), "l2_to_l3", len(staleL2))
	}

	return nil
}

// AutoCloseResolvedTickets closes tickets that have been resolved for 7 days.
func (s *ticketService) AutoCloseResolvedTickets(ctx context.Context) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	tickets, err := s.ticketRepo.ListResolvedForAutoClose(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("list resolved for auto close: %w", err)
	}

	for _, t := range tickets {
		if err := s.ticketRepo.UpdateStatus(ctx, t.ID, domain.TicketStatusClosed); err != nil {
			s.logger.Error("auto-close ticket", "ticket_id", t.ID, "error", err)
		}
	}

	if len(tickets) > 0 {
		s.logger.Info("auto-closed resolved tickets", "count", len(tickets))
	}

	return nil
}
