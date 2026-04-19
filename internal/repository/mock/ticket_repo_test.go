package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestTicketRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo()

	userID := uuid.New()
	bookingID := uuid.New()

	ticket := &domain.Ticket{
		UserID:    userID,
		BookingID: &bookingID,
		Category:  domain.TicketCategoryProblem,
		Status:    domain.TicketStatusOpen,
		Priority:  domain.TicketPriorityMedium,
		Level:     domain.TicketLevelL1,
		Subject:   "Test ticket",
	}

	// Create
	if err := repo.Create(ctx, ticket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ticket.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.Subject != "Test ticket" {
		t.Errorf("Subject = %q, want %q", got.Subject, "Test ticket")
	}
	if got.BookingID == nil || *got.BookingID != bookingID {
		t.Error("BookingID should be set")
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, uuid.New())
	if err != domain.ErrTicketNotFound {
		t.Errorf("GetByID not found: want ErrTicketNotFound, got %v", err)
	}

	// UpdateStatus
	if err := repo.UpdateStatus(ctx, ticket.ID, domain.TicketStatusInProgress); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.Status != domain.TicketStatusInProgress {
		t.Errorf("Status = %q, want %q", got.Status, domain.TicketStatusInProgress)
	}

	// UpdateStatus not found
	if err := repo.UpdateStatus(ctx, uuid.New(), domain.TicketStatusClosed); err != domain.ErrTicketNotFound {
		t.Errorf("UpdateStatus not found: want ErrTicketNotFound, got %v", err)
	}

	// UpdateLevel (escalation)
	if err := repo.UpdateLevel(ctx, ticket.ID, domain.TicketLevelL2); err != nil {
		t.Fatalf("UpdateLevel: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.Level != domain.TicketLevelL2 {
		t.Errorf("Level = %q, want %q", got.Level, domain.TicketLevelL2)
	}
	if got.Status != domain.TicketStatusEscalated {
		t.Errorf("Status after escalation = %q, want %q", got.Status, domain.TicketStatusEscalated)
	}

	// Assign
	adminID := uuid.New()
	if err := repo.Assign(ctx, ticket.ID, adminID); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.AssignedTo == nil || *got.AssignedTo != adminID {
		t.Error("AssignedTo should be set to admin ID")
	}
	if got.Status != domain.TicketStatusInProgress {
		t.Errorf("Status after assign = %q, want %q", got.Status, domain.TicketStatusInProgress)
	}

	// Resolve
	now := time.Now()
	if err := repo.Resolve(ctx, ticket.ID, now); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.Status != domain.TicketStatusResolved {
		t.Errorf("Status = %q, want %q", got.Status, domain.TicketStatusResolved)
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}

	// SubmitCSAT
	if err := repo.SubmitCSAT(ctx, ticket.ID, 5); err != nil {
		t.Fatalf("SubmitCSAT: %v", err)
	}
	got, _ = repo.GetByID(ctx, ticket.ID)
	if got.CSATScore == nil || *got.CSATScore != 5 {
		t.Errorf("CSATScore = %v, want 5", got.CSATScore)
	}

	// SubmitCSAT not found
	if err := repo.SubmitCSAT(ctx, uuid.New(), 3); err != domain.ErrTicketNotFound {
		t.Errorf("SubmitCSAT not found: want ErrTicketNotFound, got %v", err)
	}
}

func TestTicketRepo_Messages(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo()

	userID := uuid.New()
	ticket := &domain.Ticket{
		UserID:   userID,
		Category: domain.TicketCategoryQuestion,
		Status:   domain.TicketStatusOpen,
		Priority: domain.TicketPriorityLow,
		Level:    domain.TicketLevelL1,
		Subject:  "Question ticket",
	}
	if err := repo.Create(ctx, ticket); err != nil {
		t.Fatalf("Create ticket: %v", err)
	}

	// AddMessage
	msg := &domain.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   userID,
		SenderType: domain.TicketSenderUser,
		Body:       "Hello, I need help",
	}
	if err := repo.AddMessage(ctx, msg); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if msg.ID == uuid.Nil {
		t.Error("Message ID should be assigned")
	}

	// AddMessage to non-existent ticket
	badMsg := &domain.TicketMessage{
		TicketID:   uuid.New(),
		SenderID:   userID,
		SenderType: domain.TicketSenderUser,
		Body:       "Should fail",
	}
	if err := repo.AddMessage(ctx, badMsg); err != domain.ErrTicketNotFound {
		t.Errorf("AddMessage bad ticket: want ErrTicketNotFound, got %v", err)
	}

	// Add second message
	adminID := uuid.New()
	msg2 := &domain.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   adminID,
		SenderType: domain.TicketSenderAdmin,
		Body:       "We can help you",
	}
	if err := repo.AddMessage(ctx, msg2); err != nil {
		t.Fatalf("AddMessage 2: %v", err)
	}

	// ListMessages
	messages, err := repo.ListMessages(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("ListMessages count = %d, want 2", len(messages))
	}
	// Should be ordered by created_at ASC
	if messages[0].Body != "Hello, I need help" {
		t.Errorf("First message body = %q, want %q", messages[0].Body, "Hello, I need help")
	}
	if messages[1].Body != "We can help you" {
		t.Errorf("Second message body = %q, want %q", messages[1].Body, "We can help you")
	}
}

func TestTicketRepo_ListAll(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo()

	userID := uuid.New()

	// Create tickets with different attributes
	tickets := []domain.Ticket{
		{
			UserID:   userID,
			Category: domain.TicketCategoryProblem,
			Status:   domain.TicketStatusOpen,
			Priority: domain.TicketPriorityCritical,
			Level:    domain.TicketLevelL1,
			Subject:  "Critical issue",
		},
		{
			UserID:   userID,
			Category: domain.TicketCategoryQuestion,
			Status:   domain.TicketStatusOpen,
			Priority: domain.TicketPriorityLow,
			Level:    domain.TicketLevelL1,
			Subject:  "Simple question",
		},
		{
			UserID:   uuid.New(),
			Category: domain.TicketCategoryRefundReq,
			Status:   domain.TicketStatusResolved,
			Priority: domain.TicketPriorityHigh,
			Level:    domain.TicketLevelL2,
			Subject:  "Refund request",
		},
	}
	for i := range tickets {
		if err := repo.Create(ctx, &tickets[i]); err != nil {
			t.Fatalf("Create ticket %d: %v", i, err)
		}
	}

	// List all
	result, err := repo.ListAll(ctx, domain.TicketFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", result.TotalCount)
	}
	// First item should be critical (priority sorting)
	if result.Items[0].Priority != domain.TicketPriorityCritical {
		t.Errorf("First item priority = %q, want %q", result.Items[0].Priority, domain.TicketPriorityCritical)
	}

	// Filter by status
	openStatus := domain.TicketStatusOpen
	result, err = repo.ListAll(ctx, domain.TicketFilter{Status: &openStatus, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by status: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for open = %d, want 2", result.TotalCount)
	}

	// Filter by user
	result, err = repo.ListAll(ctx, domain.TicketFilter{UserID: &userID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by user: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount for user = %d, want 2", result.TotalCount)
	}

	// Filter by priority
	highPriority := domain.TicketPriorityHigh
	result, err = repo.ListAll(ctx, domain.TicketFilter{Priority: &highPriority, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by priority: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount for high priority = %d, want 1", result.TotalCount)
	}

	// Filter by level
	l2Level := domain.TicketLevelL2
	result, err = repo.ListAll(ctx, domain.TicketFilter{Level: &l2Level, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by level: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount for L2 = %d, want 1", result.TotalCount)
	}

	// Filter by category
	refundCat := domain.TicketCategoryRefundReq
	result, err = repo.ListAll(ctx, domain.TicketFilter{Category: &refundCat, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAll by category: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount for refund = %d, want 1", result.TotalCount)
	}

	// Pagination
	result, err = repo.ListAll(ctx, domain.TicketFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListAll page 1: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items on page 1 = %d, want 2", len(result.Items))
	}
	if result.TotalPages != 2 {
		t.Errorf("TotalPages = %d, want 2", result.TotalPages)
	}
}

func TestTicketRepo_ListByUser(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo()

	userID := uuid.New()
	otherID := uuid.New()

	for _, uid := range []uuid.UUID{userID, userID, otherID} {
		ticket := &domain.Ticket{
			UserID:   uid,
			Category: domain.TicketCategoryQuestion,
			Status:   domain.TicketStatusOpen,
			Priority: domain.TicketPriorityLow,
			Level:    domain.TicketLevelL1,
			Subject:  "Ticket",
		}
		if err := repo.Create(ctx, ticket); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	result, err := repo.ListByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestTicketRepo_CountByStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo()

	statuses := []domain.TicketStatus{
		domain.TicketStatusOpen,
		domain.TicketStatusOpen,
		domain.TicketStatusInProgress,
		domain.TicketStatusEscalated,
		domain.TicketStatusResolved,
	}
	for _, s := range statuses {
		ticket := &domain.Ticket{
			UserID:   uuid.New(),
			Category: domain.TicketCategoryQuestion,
			Status:   s,
			Priority: domain.TicketPriorityLow,
			Level:    domain.TicketLevelL1,
			Subject:  "Ticket",
		}
		if err := repo.Create(ctx, ticket); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	counts, err := repo.CountByStatus(ctx)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts.Open != 2 {
		t.Errorf("Open = %d, want 2", counts.Open)
	}
	if counts.InProgress != 1 {
		t.Errorf("InProgress = %d, want 1", counts.InProgress)
	}
	if counts.Escalated != 1 {
		t.Errorf("Escalated = %d, want 1", counts.Escalated)
	}
	if counts.Resolved != 1 {
		t.Errorf("Resolved = %d, want 1", counts.Resolved)
	}
	if counts.Closed != 0 {
		t.Errorf("Closed = %d, want 0", counts.Closed)
	}
}

func TestTicketRepo_ListStaleTickets(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo().(*TicketRepo)

	now := time.Now()

	// Create a stale L1 ticket (updated 25 hours ago)
	staleTicket := &domain.Ticket{
		UserID:   uuid.New(),
		Category: domain.TicketCategoryProblem,
		Status:   domain.TicketStatusOpen,
		Priority: domain.TicketPriorityMedium,
		Level:    domain.TicketLevelL1,
		Subject:  "Stale ticket",
	}
	if err := repo.Create(ctx, staleTicket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	repo.BackdateUpdatedAt(staleTicket.ID, now.Add(-25*time.Hour))

	// Create a fresh L1 ticket (updated 1 hour ago)
	freshTicket := &domain.Ticket{
		UserID:   uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Status:   domain.TicketStatusOpen,
		Priority: domain.TicketPriorityLow,
		Level:    domain.TicketLevelL1,
		Subject:  "Fresh ticket",
	}
	if err := repo.Create(ctx, freshTicket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	repo.BackdateUpdatedAt(freshTicket.ID, now.Add(-1*time.Hour))

	// Query stale L1 tickets older than 24h
	stale, err := repo.ListStaleTickets(ctx, domain.TicketLevelL1, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("ListStaleTickets: %v", err)
	}
	if len(stale) != 1 {
		t.Fatalf("ListStaleTickets count = %d, want 1", len(stale))
	}
	if stale[0].ID != staleTicket.ID {
		t.Errorf("Stale ticket ID = %v, want %v", stale[0].ID, staleTicket.ID)
	}
}

func TestTicketRepo_ListResolvedForAutoClose(t *testing.T) {
	ctx := context.Background()
	repo := NewTicketRepo().(*TicketRepo)

	now := time.Now()

	// Create a resolved ticket with old resolved_at
	oldTicket := &domain.Ticket{
		UserID:   uuid.New(),
		Category: domain.TicketCategoryProblem,
		Status:   domain.TicketStatusResolved,
		Priority: domain.TicketPriorityMedium,
		Level:    domain.TicketLevelL1,
		Subject:  "Old resolved",
	}
	if err := repo.Create(ctx, oldTicket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	repo.BackdateResolvedAt(oldTicket.ID, now.Add(-8*24*time.Hour))

	// Create a recently resolved ticket
	recentTicket := &domain.Ticket{
		UserID:   uuid.New(),
		Category: domain.TicketCategoryQuestion,
		Status:   domain.TicketStatusResolved,
		Priority: domain.TicketPriorityLow,
		Level:    domain.TicketLevelL1,
		Subject:  "Recent resolved",
	}
	if err := repo.Create(ctx, recentTicket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	repo.BackdateResolvedAt(recentTicket.ID, now.Add(-1*24*time.Hour))

	// Query tickets resolved more than 7 days ago
	result, err := repo.ListResolvedForAutoClose(ctx, now.Add(-7*24*time.Hour))
	if err != nil {
		t.Fatalf("ListResolvedForAutoClose: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ListResolvedForAutoClose count = %d, want 1", len(result))
	}
	if result[0].ID != oldTicket.ID {
		t.Errorf("Result ticket ID = %v, want %v", result[0].ID, oldTicket.ID)
	}
}
