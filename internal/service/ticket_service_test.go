package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newTicketTestService() (service.TicketService, *mock.TicketRepo) {
	repo := mock.NewTicketRepo().(*mock.TicketRepo)
	log := logger.New(logger.LevelWarn)
	return service.NewTicketService(repo, log), repo
}

func TestTicketService_CreateTicket(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "How to book?",
	}

	err := svc.CreateTicket(ctx, userID, ticket, "I need help booking")
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}

	if ticket.UserID != userID {
		t.Errorf("expected user_id=%s, got %s", userID, ticket.UserID)
	}
	if ticket.Status != domain.TicketStatusOpen {
		t.Errorf("expected status=open, got %s", ticket.Status)
	}
	if ticket.Priority != domain.TicketPriorityLow {
		t.Errorf("expected priority=low for question, got %s", ticket.Priority)
	}
	if ticket.Level != domain.TicketLevelL1 {
		t.Errorf("expected level=L1, got %s", ticket.Level)
	}

	// Check initial message was added
	msgs, _ := repo.ListMessages(ctx, ticket.ID)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 initial message, got %d", len(msgs))
	}
	if msgs[0].Body != "I need help booking" {
		t.Errorf("expected message body='I need help booking', got %q", msgs[0].Body)
	}
}

func TestTicketService_CreateTicket_AutoPriority(t *testing.T) {
	tests := []struct {
		category domain.TicketCategory
		priority domain.TicketPriority
		level    domain.TicketLevel
	}{
		{domain.TicketCategoryQuestion, domain.TicketPriorityLow, domain.TicketLevelL1},
		{domain.TicketCategoryProblem, domain.TicketPriorityMedium, domain.TicketLevelL1},
		{domain.TicketCategoryComplaint, domain.TicketPriorityHigh, domain.TicketLevelL1},
		{domain.TicketCategoryRefundReq, domain.TicketPriorityHigh, domain.TicketLevelL1},
		{domain.TicketCategoryAccountIssue, domain.TicketPriorityMedium, domain.TicketLevelL1},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			svc, _ := newTicketTestService()
			ctx := context.Background()

			ticket := &domain.Ticket{
				ID:       uuid.New(),
				Category: tt.category,
				Subject:  "Test",
			}

			err := svc.CreateTicket(ctx, uuid.New(), ticket, "")
			if err != nil {
				t.Fatalf("CreateTicket: %v", err)
			}

			if ticket.Priority != tt.priority {
				t.Errorf("expected priority=%s, got %s", tt.priority, ticket.Priority)
			}
			if ticket.Level != tt.level {
				t.Errorf("expected level=%s, got %s", tt.level, ticket.Level)
			}
		})
	}
}

func TestTicketService_CreateTicket_InvalidInput(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()

	// Empty subject
	err := svc.CreateTicket(ctx, uuid.New(), &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "",
	}, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty subject, got %v", err)
	}

	// Invalid category
	err = svc.CreateTicket(ctx, uuid.New(), &domain.Ticket{
		ID:       uuid.New(),
		Category: "invalid",
		Subject:  "Test",
	}, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid category, got %v", err)
	}
}

func TestTicketService_GetTicket(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	// Owner can see their ticket
	got, err := svc.GetTicket(ctx, userID, domain.RoleClient, ticket.ID)
	if err != nil {
		t.Fatalf("GetTicket: %v", err)
	}
	if got.ID != ticket.ID {
		t.Errorf("expected ticket ID=%s, got %s", ticket.ID, got.ID)
	}

	// Admin can see any ticket
	got, err = svc.GetTicket(ctx, uuid.New(), domain.RoleAdmin, ticket.ID)
	if err != nil {
		t.Fatalf("GetTicket as admin: %v", err)
	}
	if got.ID != ticket.ID {
		t.Errorf("expected ticket ID=%s, got %s", ticket.ID, got.ID)
	}

	// Other user cannot see it
	_, err = svc.GetTicket(ctx, uuid.New(), domain.RoleClient, ticket.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for other user, got %v", err)
	}
}

func TestTicketService_GetTicket_NotFound(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()

	_, err := svc.GetTicket(ctx, uuid.New(), domain.RoleAdmin, uuid.New())
	if !errors.Is(err, domain.ErrTicketNotFound) {
		t.Errorf("expected ErrTicketNotFound, got %v", err)
	}
}

func TestTicketService_AddMessage(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "Initial")

	// User adds message
	msg, err := svc.AddMessage(ctx, userID, domain.RoleClient, ticket.ID, "Follow up", nil)
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if msg.SenderType != domain.TicketSenderUser {
		t.Errorf("expected sender_type=user, got %s", msg.SenderType)
	}

	// Admin adds message
	adminID := uuid.New()
	msg, err = svc.AddMessage(ctx, adminID, domain.RoleAdmin, ticket.ID, "Admin response", nil)
	if err != nil {
		t.Fatalf("AddMessage admin: %v", err)
	}
	if msg.SenderType != domain.TicketSenderAdmin {
		t.Errorf("expected sender_type=admin, got %s", msg.SenderType)
	}

	// Verify messages count (initial + follow up + admin response)
	msgs, _ := repo.ListMessages(ctx, ticket.ID)
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages, got %d", len(msgs))
	}
}

func TestTicketService_AddMessage_Forbidden(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	// Other non-admin user cannot add message
	_, err := svc.AddMessage(ctx, uuid.New(), domain.RoleClient, ticket.ID, "Hijack", nil)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTicketService_AddMessage_ClosedTicket(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.CloseTicket(ctx, ticket.ID)

	_, err := svc.AddMessage(ctx, userID, domain.RoleClient, ticket.ID, "After close", nil)
	if !errors.Is(err, domain.ErrTicketAlreadyClosed) {
		t.Errorf("expected ErrTicketAlreadyClosed, got %v", err)
	}
}

func TestTicketService_EscalateTicket(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	// L1 -> L2
	err := svc.EscalateTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("EscalateTicket L1->L2: %v", err)
	}
	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.Level != domain.TicketLevelL2 {
		t.Errorf("expected level=L2, got %s", got.Level)
	}
	if got.Status != domain.TicketStatusEscalated {
		t.Errorf("expected status=escalated, got %s", got.Status)
	}

	// L2 -> L3
	err = svc.EscalateTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("EscalateTicket L2->L3: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.Level != domain.TicketLevelL3 {
		t.Errorf("expected level=L3, got %s", got.Level)
	}
}

func TestTicketService_ResolveTicket(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	err := svc.ResolveTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("ResolveTicket: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.Status != domain.TicketStatusResolved {
		t.Errorf("expected status=resolved, got %s", got.Status)
	}
	if got.ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
}

func TestTicketService_ResolveTicket_AlreadyResolved(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)

	err := svc.ResolveTicket(ctx, ticket.ID)
	if !errors.Is(err, domain.ErrTicketAlreadyResolved) {
		t.Errorf("expected ErrTicketAlreadyResolved, got %v", err)
	}
}

func TestTicketService_CloseTicket(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	err := svc.CloseTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("CloseTicket: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.Status != domain.TicketStatusClosed {
		t.Errorf("expected status=closed, got %s", got.Status)
	}
}

func TestTicketService_CloseTicket_AlreadyClosed(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.CloseTicket(ctx, ticket.ID)

	err := svc.CloseTicket(ctx, ticket.ID)
	if !errors.Is(err, domain.ErrTicketAlreadyClosed) {
		t.Errorf("expected ErrTicketAlreadyClosed, got %v", err)
	}
}

func TestTicketService_SubmitCSAT(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)

	err := svc.SubmitCSAT(ctx, userID, ticket.ID, 5)
	if err != nil {
		t.Fatalf("SubmitCSAT: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.CSATScore == nil || *got.CSATScore != 5 {
		t.Errorf("expected csat_score=5, got %v", got.CSATScore)
	}
}

func TestTicketService_SubmitCSAT_InvalidScore(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)

	err := svc.SubmitCSAT(ctx, userID, ticket.ID, 0)
	if !errors.Is(err, domain.ErrCSATInvalidScore) {
		t.Errorf("expected ErrCSATInvalidScore for score=0, got %v", err)
	}

	err = svc.SubmitCSAT(ctx, userID, ticket.ID, 6)
	if !errors.Is(err, domain.ErrCSATInvalidScore) {
		t.Errorf("expected ErrCSATInvalidScore for score=6, got %v", err)
	}
}

func TestTicketService_SubmitCSAT_NotResolved(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	err := svc.SubmitCSAT(ctx, userID, ticket.ID, 3)
	if !errors.Is(err, domain.ErrCSATNotResolved) {
		t.Errorf("expected ErrCSATNotResolved, got %v", err)
	}
}

func TestTicketService_SubmitCSAT_AlreadySubmitted(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)
	_ = svc.SubmitCSAT(ctx, userID, ticket.ID, 4)

	err := svc.SubmitCSAT(ctx, userID, ticket.ID, 5)
	if !errors.Is(err, domain.ErrCSATAlreadySubmitted) {
		t.Errorf("expected ErrCSATAlreadySubmitted, got %v", err)
	}
}

func TestTicketService_SubmitCSAT_Forbidden(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)

	err := svc.SubmitCSAT(ctx, uuid.New(), ticket.ID, 3)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTicketService_AssignTicket(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()
	adminID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	err := svc.AssignTicket(ctx, ticket.ID, adminID)
	if err != nil {
		t.Fatalf("AssignTicket: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.AssignedTo == nil || *got.AssignedTo != adminID {
		t.Errorf("expected assigned_to=%s, got %v", adminID, got.AssignedTo)
	}
	if got.Status != domain.TicketStatusInProgress {
		t.Errorf("expected status=in_progress after assign, got %s", got.Status)
	}
}

func TestTicketService_AssignTicket_Closed(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.CloseTicket(ctx, ticket.ID)

	err := svc.AssignTicket(ctx, ticket.ID, uuid.New())
	if !errors.Is(err, domain.ErrTicketAlreadyClosed) {
		t.Errorf("expected ErrTicketAlreadyClosed, got %v", err)
	}
}

func TestTicketService_ListUserTickets(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_ = svc.CreateTicket(ctx, userID, &domain.Ticket{
			ID:       uuid.New(),
			Category: domain.TicketCategoryQuestion,
			Subject:  "Test",
		}, "")
	}

	// Another user's ticket
	_ = svc.CreateTicket(ctx, uuid.New(), &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Other",
	}, "")

	result, err := svc.ListUserTickets(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListUserTickets: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("expected 3 tickets, got %d", result.TotalCount)
	}
}

func TestTicketService_GetStats(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	// Create tickets in different states
	t1 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryQuestion, Subject: "Open"}
	t2 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryProblem, Subject: "Resolved"}
	t3 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryComplaint, Subject: "Closed"}

	_ = svc.CreateTicket(ctx, userID, t1, "")
	_ = svc.CreateTicket(ctx, userID, t2, "")
	_ = svc.CreateTicket(ctx, userID, t3, "")

	_ = svc.ResolveTicket(ctx, t2.ID)
	_ = svc.CloseTicket(ctx, t3.ID)

	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.Open != 1 {
		t.Errorf("expected 1 open, got %d", stats.Open)
	}
	if stats.Resolved != 1 {
		t.Errorf("expected 1 resolved, got %d", stats.Resolved)
	}
	if stats.Closed != 1 {
		t.Errorf("expected 1 closed, got %d", stats.Closed)
	}
}

func TestTicketService_AutoEscalateStaleTickets(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Stale ticket",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")

	// Backdate updated_at to 25 hours ago
	repo.BackdateUpdatedAt(ticket.ID, time.Now().Add(-25*time.Hour))

	err := svc.AutoEscalateStaleTickets(ctx)
	if err != nil {
		t.Fatalf("AutoEscalateStaleTickets: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.Level != domain.TicketLevelL2 {
		t.Errorf("expected stale L1 ticket to be escalated to L2, got %s", got.Level)
	}
}

func TestTicketService_AutoCloseResolvedTickets(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "To auto-close",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.ResolveTicket(ctx, ticket.ID)

	// Backdate resolved_at to 8 days ago
	repo.BackdateResolvedAt(ticket.ID, time.Now().Add(-8*24*time.Hour))

	err := svc.AutoCloseResolvedTickets(ctx)
	if err != nil {
		t.Fatalf("AutoCloseResolvedTickets: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.Status != domain.TicketStatusClosed {
		t.Errorf("expected auto-closed ticket status=closed, got %s", got.Status)
	}
}

func TestTicketService_ListMessages(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "Initial msg")

	// Add a few more messages
	_, _ = svc.AddMessage(ctx, userID, domain.RoleClient, ticket.ID, "Second msg", nil)
	adminID := uuid.New()
	_, _ = svc.AddMessage(ctx, adminID, domain.RoleAdmin, ticket.ID, "Admin reply", nil)

	// User can list their messages
	msgs, err := svc.ListMessages(ctx, userID, domain.RoleClient, ticket.ID)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages, got %d", len(msgs))
	}

	// Other user cannot see messages
	_, err = svc.ListMessages(ctx, uuid.New(), domain.RoleClient, ticket.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for other user, got %v", err)
	}

	// Admin can see messages
	msgs, err = svc.ListMessages(ctx, uuid.New(), domain.RoleAdmin, ticket.ID)
	if err != nil {
		t.Fatalf("ListMessages as admin: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages as admin, got %d", len(msgs))
	}
}

func TestTicketService_EscalateTicket_Closed(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Subject:  "Test",
	}
	_ = svc.CreateTicket(ctx, userID, ticket, "")
	_ = svc.CloseTicket(ctx, ticket.ID)

	err := svc.EscalateTicket(ctx, ticket.ID)
	if !errors.Is(err, domain.ErrTicketAlreadyClosed) {
		t.Errorf("expected ErrTicketAlreadyClosed, got %v", err)
	}
}

func TestTicketService_GetOperationMetrics(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()
	adminID := uuid.New()

	// Create 3 tickets, resolve 2 at L1, escalate 1

	// Ticket 1: resolved at L1 with admin message within 24h (FCR=yes, SLA=yes)
	t1 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryQuestion, Subject: "Q1"}
	_ = svc.CreateTicket(ctx, userID, t1, "help")
	_, _ = svc.AddMessage(ctx, adminID, domain.RoleAdmin, t1.ID, "answer", nil)
	_ = svc.ResolveTicket(ctx, t1.ID)
	_ = svc.SubmitCSAT(ctx, userID, t1.ID, 5)

	// Ticket 2: resolved at L1, no admin message (FCR=yes, SLA=no)
	t2 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryProblem, Subject: "Q2"}
	_ = svc.CreateTicket(ctx, userID, t2, "issue")
	_ = svc.ResolveTicket(ctx, t2.ID)
	_ = svc.SubmitCSAT(ctx, userID, t2.ID, 3)

	// Ticket 3: escalated to L2, then resolved (FCR=no, SLA=yes because admin replied)
	t3 := &domain.Ticket{ID: uuid.New(), Category: domain.TicketCategoryComplaint, Subject: "Q3"}
	_ = svc.CreateTicket(ctx, userID, t3, "complaint")
	_, _ = svc.AddMessage(ctx, adminID, domain.RoleAdmin, t3.ID, "looking into it", nil)
	_ = svc.EscalateTicket(ctx, t3.ID)
	_ = svc.ResolveTicket(ctx, t3.ID)

	metrics, err := svc.GetOperationMetrics(ctx, domain.TicketMetricsFilter{})
	if err != nil {
		t.Fatalf("GetOperationMetrics: %v", err)
	}

	if metrics.TotalTickets != 3 {
		t.Errorf("expected total_tickets=3, got %d", metrics.TotalTickets)
	}
	if metrics.TotalResolved != 3 {
		t.Errorf("expected total_resolved=3, got %d", metrics.TotalResolved)
	}

	// FCR: 2 resolved at L1 out of 3 = 66.67%
	expectedFCR := float64(2) / float64(3) * 100
	if diff := metrics.FCRPercent - expectedFCR; diff > 0.1 || diff < -0.1 {
		t.Errorf("expected FCR~%.1f%%, got %.1f%%", expectedFCR, metrics.FCRPercent)
	}

	// AHT should be > 0 (all 3 are resolved)
	if metrics.AHTSeconds <= 0 {
		t.Errorf("expected AHT > 0, got %f", metrics.AHTSeconds)
	}

	// Avg CSAT: (5 + 3) / 2 = 4.0 (only 2 tickets have scores)
	if diff := metrics.AvgCSAT - 4.0; diff > 0.1 || diff < -0.1 {
		t.Errorf("expected avg_csat~4.0, got %.1f", metrics.AvgCSAT)
	}

	// SLA: 2 out of 3 have admin response within 24h (t1, t3)
	expectedSLA := float64(2) / float64(3) * 100
	if diff := metrics.SLACompliancePercent - expectedSLA; diff > 0.1 || diff < -0.1 {
		t.Errorf("expected SLA~%.1f%%, got %.1f%%", expectedSLA, metrics.SLACompliancePercent)
	}

	// Test with date filter - future range should return zeros
	futureFrom := time.Now().Add(24 * time.Hour)
	metricsFiltered, err := svc.GetOperationMetrics(ctx, domain.TicketMetricsFilter{DateFrom: &futureFrom})
	if err != nil {
		t.Fatalf("GetOperationMetrics with filter: %v", err)
	}
	if metricsFiltered.TotalTickets != 0 {
		t.Errorf("expected 0 tickets with future date filter, got %d", metricsFiltered.TotalTickets)
	}

	_ = repo // keep reference
}

func TestTicketService_GetOperationMetrics_Empty(t *testing.T) {
	svc, _ := newTicketTestService()
	ctx := context.Background()

	metrics, err := svc.GetOperationMetrics(ctx, domain.TicketMetricsFilter{})
	if err != nil {
		t.Fatalf("GetOperationMetrics empty: %v", err)
	}

	if metrics.TotalTickets != 0 {
		t.Errorf("expected total_tickets=0, got %d", metrics.TotalTickets)
	}
	if metrics.FCRPercent != 0 {
		t.Errorf("expected FCR=0 for no tickets, got %f", metrics.FCRPercent)
	}
	if metrics.AHTSeconds != 0 {
		t.Errorf("expected AHT=0 for no tickets, got %f", metrics.AHTSeconds)
	}
	if metrics.AvgCSAT != 0 {
		t.Errorf("expected CSAT=0 for no tickets, got %f", metrics.AvgCSAT)
	}
}

func TestTicketService_CreateTicket_WithBookingID(t *testing.T) {
	svc, repo := newTicketTestService()
	ctx := context.Background()
	userID := uuid.New()
	bookingID := uuid.New()

	ticket := &domain.Ticket{
		ID:        uuid.New(),
		Category:  domain.TicketCategoryRefundReq,
		Subject:   "Refund request",
		BookingID: &bookingID,
	}

	err := svc.CreateTicket(ctx, userID, ticket, "Please refund")
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}

	got, _ := repo.GetByID(ctx, ticket.ID)
	if got.BookingID == nil || *got.BookingID != bookingID {
		t.Errorf("expected booking_id=%s, got %v", bookingID, got.BookingID)
	}
	if got.Priority != domain.TicketPriorityHigh {
		t.Errorf("expected priority=high for refund_request, got %s", got.Priority)
	}
}
