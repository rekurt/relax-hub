package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/service"
)

// --- Mock ticket service ---

type mockTicketService struct {
	createTicketFn   func(ctx context.Context, userID uuid.UUID, ticket *domain.Ticket, initialMessage string) error
	getTicketFn      func(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) (*domain.Ticket, error)
	listUserFn       func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error)
	listAllFn        func(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error)
	addMessageFn     func(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID, body string, attachments []string) (*domain.TicketMessage, error)
	listMessagesFn   func(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) ([]domain.TicketMessage, error)
	assignTicketFn   func(ctx context.Context, ticketID uuid.UUID, assignedTo uuid.UUID) error
	escalateTicketFn func(ctx context.Context, ticketID uuid.UUID) error
	resolveTicketFn  func(ctx context.Context, ticketID uuid.UUID) error
	closeTicketFn    func(ctx context.Context, ticketID uuid.UUID) error
	submitCSATFn     func(ctx context.Context, userID uuid.UUID, ticketID uuid.UUID, score int) error
	getStatsFn       func(ctx context.Context) (*domain.TicketStatusCounts, error)
}

func (m *mockTicketService) CreateTicket(ctx context.Context, userID uuid.UUID, ticket *domain.Ticket, initialMessage string) error {
	if m.createTicketFn != nil {
		return m.createTicketFn(ctx, userID, ticket, initialMessage)
	}
	return nil
}

func (m *mockTicketService) GetTicket(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) (*domain.Ticket, error) {
	if m.getTicketFn != nil {
		return m.getTicketFn(ctx, userID, role, ticketID)
	}
	return nil, nil
}

func (m *mockTicketService) ListUserTickets(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error) {
	if m.listUserFn != nil {
		return m.listUserFn(ctx, userID, page, pageSize)
	}
	return nil, nil
}

func (m *mockTicketService) ListAllTickets(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockTicketService) AddMessage(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID, body string, attachments []string) (*domain.TicketMessage, error) {
	if m.addMessageFn != nil {
		return m.addMessageFn(ctx, userID, role, ticketID, body, attachments)
	}
	return nil, nil
}

func (m *mockTicketService) ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	if m.listMessagesFn != nil {
		return m.listMessagesFn(ctx, userID, role, ticketID)
	}
	return nil, nil
}

func (m *mockTicketService) AssignTicket(ctx context.Context, ticketID uuid.UUID, assignedTo uuid.UUID) error {
	if m.assignTicketFn != nil {
		return m.assignTicketFn(ctx, ticketID, assignedTo)
	}
	return nil
}

func (m *mockTicketService) EscalateTicket(ctx context.Context, ticketID uuid.UUID) error {
	if m.escalateTicketFn != nil {
		return m.escalateTicketFn(ctx, ticketID)
	}
	return nil
}

func (m *mockTicketService) ResolveTicket(ctx context.Context, ticketID uuid.UUID) error {
	if m.resolveTicketFn != nil {
		return m.resolveTicketFn(ctx, ticketID)
	}
	return nil
}

func (m *mockTicketService) CloseTicket(ctx context.Context, ticketID uuid.UUID) error {
	if m.closeTicketFn != nil {
		return m.closeTicketFn(ctx, ticketID)
	}
	return nil
}

func (m *mockTicketService) SubmitCSAT(ctx context.Context, userID uuid.UUID, ticketID uuid.UUID, score int) error {
	if m.submitCSATFn != nil {
		return m.submitCSATFn(ctx, userID, ticketID, score)
	}
	return nil
}

func (m *mockTicketService) GetStats(ctx context.Context) (*domain.TicketStatusCounts, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx)
	}
	return nil, nil
}

func (m *mockTicketService) AutoEscalateStaleTickets(_ context.Context) error { return nil }
func (m *mockTicketService) AutoCloseResolvedTickets(_ context.Context) error  { return nil }

var _ service.TicketService = (*mockTicketService)(nil)

// --- Helper ---

func testTicket() *domain.Ticket {
	now := time.Now()
	return &domain.Ticket{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Category:  domain.TicketCategoryProblem,
		Status:    domain.TicketStatusOpen,
		Priority:  domain.TicketPriorityMedium,
		Level:     domain.TicketLevelL1,
		Subject:   "Test ticket",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testTicketMessage(ticketID, senderID uuid.UUID) *domain.TicketMessage {
	return &domain.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: domain.TicketSenderUser,
		Body:       "Test message",
		CreatedAt:  time.Now(),
	}
}

// --- User endpoint tests ---

func TestTicketHandler_CreateTicket(t *testing.T) {
	userID := uuid.New()
	ticket := testTicket()
	ticket.UserID = userID

	svc := &mockTicketService{
		createTicketFn: func(_ context.Context, uid uuid.UUID, tk *domain.Ticket, msg string) error {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if msg != "Help me please" {
				t.Errorf("expected message 'Help me please', got %q", msg)
			}
			tk.UserID = uid
			tk.Status = domain.TicketStatusOpen
			tk.Priority = domain.TicketPriorityMedium
			tk.Level = domain.TicketLevelL1
			tk.CreatedAt = time.Now()
			tk.UpdatedAt = time.Now()
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets", h.CreateTicket)

	body, _ := json.Marshal(map[string]string{
		"category": "problem",
		"subject":  "Test ticket",
		"message":  "Help me please",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_CreateTicket_WithBookingID(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	svc := &mockTicketService{
		createTicketFn: func(_ context.Context, _ uuid.UUID, tk *domain.Ticket, _ string) error {
			if tk.BookingID == nil {
				t.Error("expected bookingID to be set")
			} else if *tk.BookingID != bookingID {
				t.Errorf("expected bookingID %v, got %v", bookingID, *tk.BookingID)
			}
			tk.UserID = userID
			tk.CreatedAt = time.Now()
			tk.UpdatedAt = time.Now()
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets", h.CreateTicket)

	bid := bookingID.String()
	body, _ := json.Marshal(map[string]string{
		"booking_id": bid,
		"category":   "refund_request",
		"subject":    "Refund needed",
		"message":    "I need a refund",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_CreateTicket_InvalidBookingID(t *testing.T) {
	svc := &mockTicketService{}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets", h.CreateTicket)

	body, _ := json.Marshal(map[string]string{
		"booking_id": "not-a-uuid",
		"category":   "problem",
		"subject":    "Test",
		"message":    "Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_GetTicket(t *testing.T) {
	userID := uuid.New()
	ticket := testTicket()
	ticket.UserID = userID

	svc := &mockTicketService{
		getTicketFn: func(_ context.Context, uid uuid.UUID, _ domain.UserRole, tid uuid.UUID) (*domain.Ticket, error) {
			if tid != ticket.ID {
				t.Errorf("expected ticketID %v, got %v", ticket.ID, tid)
			}
			return ticket, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/tickets/{id}", h.GetTicket)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets/"+ticket.ID.String(), nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_GetTicket_NotFound(t *testing.T) {
	svc := &mockTicketService{
		getTicketFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Ticket, error) {
			return nil, domain.ErrTicketNotFound
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/tickets/{id}", h.GetTicket)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets/"+uuid.New().String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestTicketHandler_GetTicket_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Get("/my/tickets/{id}", h.GetTicket)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets/not-a-uuid", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_ListUserTickets(t *testing.T) {
	userID := uuid.New()

	svc := &mockTicketService{
		listUserFn: func(_ context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return &domain.PaginatedResult[domain.Ticket]{
				Items:      []domain.Ticket{*testTicket()},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/tickets", h.ListUserTickets)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets?page=1&page_size=20", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTicketHandler_AddUserMessage(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()
	msg := testTicketMessage(ticketID, userID)

	svc := &mockTicketService{
		addMessageFn: func(_ context.Context, uid uuid.UUID, _ domain.UserRole, tid uuid.UUID, body string, _ []string) (*domain.TicketMessage, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if tid != ticketID {
				t.Errorf("expected ticketID %v, got %v", ticketID, tid)
			}
			if body != "Hello support" {
				t.Errorf("expected body 'Hello support', got %q", body)
			}
			return msg, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/messages", h.AddUserMessage)

	body, _ := json.Marshal(map[string]string{
		"body": "Hello support",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/"+ticketID.String()+"/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_AddUserMessage_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/messages", h.AddUserMessage)

	body, _ := json.Marshal(map[string]string{"body": "test"})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/bad-id/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_ListMessages(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	svc := &mockTicketService{
		listMessagesFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, tid uuid.UUID) ([]domain.TicketMessage, error) {
			return []domain.TicketMessage{
				*testTicketMessage(tid, userID),
			}, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/tickets/{id}/messages", h.ListMessages)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets/"+ticketID.String()+"/messages", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTicketHandler_ListMessages_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Get("/my/tickets/{id}/messages", h.ListMessages)

	req := httptest.NewRequest(http.MethodGet, "/my/tickets/bad-id/messages", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_SubmitCSAT(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	svc := &mockTicketService{
		submitCSATFn: func(_ context.Context, uid uuid.UUID, tid uuid.UUID, score int) error {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if tid != ticketID {
				t.Errorf("expected ticketID %v, got %v", ticketID, tid)
			}
			if score != 5 {
				t.Errorf("expected score 5, got %d", score)
			}
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/csat", h.SubmitCSAT)

	body, _ := json.Marshal(map[string]int{"score": 5})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/"+ticketID.String()+"/csat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_SubmitCSAT_AlreadySubmitted(t *testing.T) {
	svc := &mockTicketService{
		submitCSATFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) error {
			return domain.ErrCSATAlreadySubmitted
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/csat", h.SubmitCSAT)

	body, _ := json.Marshal(map[string]int{"score": 4})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/"+uuid.New().String()+"/csat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestTicketHandler_SubmitCSAT_NotResolved(t *testing.T) {
	svc := &mockTicketService{
		submitCSATFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) error {
			return domain.ErrCSATNotResolved
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/csat", h.SubmitCSAT)

	body, _ := json.Marshal(map[string]int{"score": 3})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/"+uuid.New().String()+"/csat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_SubmitCSAT_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Post("/my/tickets/{id}/csat", h.SubmitCSAT)

	body, _ := json.Marshal(map[string]int{"score": 5})
	req := httptest.NewRequest(http.MethodPost, "/my/tickets/bad-id/csat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// --- Admin endpoint tests ---

func TestTicketHandler_AdminListTickets(t *testing.T) {
	svc := &mockTicketService{
		listAllFn: func(_ context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
			if filter.Status != nil && *filter.Status != domain.TicketStatusOpen {
				t.Errorf("expected status filter open, got %v", *filter.Status)
			}
			return &domain.PaginatedResult[domain.Ticket]{
				Items:      []domain.Ticket{},
				Page:       1,
				PageSize:   20,
				TotalCount: 0,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/tickets", h.AdminListTickets)

	req := httptest.NewRequest(http.MethodGet, "/admin/tickets?status=open", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminListTickets_WithAllFilters(t *testing.T) {
	svc := &mockTicketService{
		listAllFn: func(_ context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
			if filter.Priority == nil || *filter.Priority != domain.TicketPriorityHigh {
				t.Errorf("expected priority filter high")
			}
			if filter.Level == nil || *filter.Level != domain.TicketLevelL2 {
				t.Errorf("expected level filter L2")
			}
			if filter.Category == nil || *filter.Category != domain.TicketCategoryComplaint {
				t.Errorf("expected category filter complaint")
			}
			return &domain.PaginatedResult[domain.Ticket]{
				Items:      []domain.Ticket{},
				Page:       1,
				PageSize:   20,
				TotalCount: 0,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/tickets", h.AdminListTickets)

	req := httptest.NewRequest(http.MethodGet, "/admin/tickets?priority=high&level=L2&category=complaint", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminAssignTicket(t *testing.T) {
	ticketID := uuid.New()
	assignedTo := uuid.New()

	svc := &mockTicketService{
		assignTicketFn: func(_ context.Context, tid uuid.UUID, aid uuid.UUID) error {
			if tid != ticketID {
				t.Errorf("expected ticketID %v, got %v", ticketID, tid)
			}
			if aid != assignedTo {
				t.Errorf("expected assignedTo %v, got %v", assignedTo, aid)
			}
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/assign", h.AdminAssignTicket)

	body, _ := json.Marshal(map[string]string{
		"assigned_to": assignedTo.String(),
	})
	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+ticketID.String()+"/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_AdminAssignTicket_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/assign", h.AdminAssignTicket)

	body, _ := json.Marshal(map[string]string{"assigned_to": uuid.New().String()})
	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/bad-id/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminAssignTicket_InvalidAssignedTo(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/assign", h.AdminAssignTicket)

	body, _ := json.Marshal(map[string]string{"assigned_to": "not-a-uuid"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+uuid.New().String()+"/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminAssignTicket_NotFound(t *testing.T) {
	svc := &mockTicketService{
		assignTicketFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domain.ErrTicketNotFound
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/assign", h.AdminAssignTicket)

	body, _ := json.Marshal(map[string]string{"assigned_to": uuid.New().String()})
	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+uuid.New().String()+"/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminEscalateTicket(t *testing.T) {
	ticketID := uuid.New()

	svc := &mockTicketService{
		escalateTicketFn: func(_ context.Context, tid uuid.UUID) error {
			if tid != ticketID {
				t.Errorf("expected ticketID %v, got %v", ticketID, tid)
			}
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/escalate", h.AdminEscalateTicket)

	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+ticketID.String()+"/escalate", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_AdminEscalateTicket_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/escalate", h.AdminEscalateTicket)

	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/bad-id/escalate", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminResolveTicket(t *testing.T) {
	ticketID := uuid.New()

	svc := &mockTicketService{
		resolveTicketFn: func(_ context.Context, tid uuid.UUID) error {
			if tid != ticketID {
				t.Errorf("expected ticketID %v, got %v", ticketID, tid)
			}
			return nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/resolve", h.AdminResolveTicket)

	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+ticketID.String()+"/resolve", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_AdminResolveTicket_NotFound(t *testing.T) {
	svc := &mockTicketService{
		resolveTicketFn: func(_ context.Context, _ uuid.UUID) error {
			return domain.ErrTicketNotFound
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/tickets/{id}/resolve", h.AdminResolveTicket)

	req := httptest.NewRequest(http.MethodPatch, "/admin/tickets/"+uuid.New().String()+"/resolve", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminAddMessage(t *testing.T) {
	adminID := uuid.New()
	ticketID := uuid.New()
	msg := testTicketMessage(ticketID, adminID)
	msg.SenderType = domain.TicketSenderAdmin

	svc := &mockTicketService{
		addMessageFn: func(_ context.Context, uid uuid.UUID, role domain.UserRole, tid uuid.UUID, body string, _ []string) (*domain.TicketMessage, error) {
			if uid != adminID {
				t.Errorf("expected adminID %v, got %v", adminID, uid)
			}
			if role != domain.RoleAdmin {
				t.Errorf("expected admin role, got %v", role)
			}
			return msg, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Post("/admin/tickets/{id}/messages", h.AdminAddMessage)

	body, _ := json.Marshal(map[string]string{"body": "Admin response"})
	req := httptest.NewRequest(http.MethodPost, "/admin/tickets/"+ticketID.String()+"/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTicketHandler_AdminAddMessage_InvalidID(t *testing.T) {
	h := handler.NewTicketHandler(&mockTicketService{})
	router := chi.NewRouter()
	router.Post("/admin/tickets/{id}/messages", h.AdminAddMessage)

	body, _ := json.Marshal(map[string]string{"body": "test"})
	req := httptest.NewRequest(http.MethodPost, "/admin/tickets/bad-id/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTicketHandler_AdminGetStats(t *testing.T) {
	svc := &mockTicketService{
		getStatsFn: func(_ context.Context) (*domain.TicketStatusCounts, error) {
			return &domain.TicketStatusCounts{
				Open:       5,
				InProgress: 3,
				Escalated:  1,
				Resolved:   10,
				Closed:     20,
			}, nil
		},
	}

	h := handler.NewTicketHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/tickets/stats", h.AdminGetStats)

	req := httptest.NewRequest(http.MethodGet, "/admin/tickets/stats", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
