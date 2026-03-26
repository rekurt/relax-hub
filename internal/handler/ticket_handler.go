package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type TicketHandler struct {
	ticketService service.TicketService
}

func NewTicketHandler(ticketService service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

type createTicketRequest struct {
	BookingID *string `json:"booking_id"`
	Category  string  `json:"category"`
	Subject   string  `json:"subject"`
	Message   string  `json:"message"`
}

type addTicketMessageRequest struct {
	Body        string   `json:"body"`
	Attachments []string `json:"attachments,omitempty"`
}

type submitCSATRequest struct {
	Score int `json:"score"`
}

type assignTicketRequest struct {
	AssignedTo string `json:"assigned_to"`
}

type ticketResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	BookingID  *string `json:"booking_id,omitempty"`
	Category   string  `json:"category"`
	Status     string  `json:"status"`
	Priority   string  `json:"priority"`
	Level      string  `json:"level"`
	Subject    string  `json:"subject"`
	AssignedTo *string `json:"assigned_to,omitempty"`
	CSATScore  *int    `json:"csat_score,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
}

type ticketMessageResponse struct {
	ID          string   `json:"id"`
	TicketID    string   `json:"ticket_id"`
	SenderID    string   `json:"sender_id"`
	SenderType  string   `json:"sender_type"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
	CreatedAt   string   `json:"created_at"`
}

type ticketStatsResponse struct {
	Open       int64 `json:"open"`
	InProgress int64 `json:"in_progress"`
	Escalated  int64 `json:"escalated"`
	Resolved   int64 `json:"resolved"`
	Closed     int64 `json:"closed"`
}

func toTicketResponse(t *domain.Ticket) ticketResponse {
	resp := ticketResponse{
		ID:        t.ID.String(),
		UserID:    t.UserID.String(),
		Category:  string(t.Category),
		Status:    string(t.Status),
		Priority:  string(t.Priority),
		Level:     string(t.Level),
		Subject:   t.Subject,
		CSATScore: t.CSATScore,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
	if t.BookingID != nil {
		s := t.BookingID.String()
		resp.BookingID = &s
	}
	if t.AssignedTo != nil {
		s := t.AssignedTo.String()
		resp.AssignedTo = &s
	}
	if t.ResolvedAt != nil {
		s := t.ResolvedAt.Format(time.RFC3339)
		resp.ResolvedAt = &s
	}
	return resp
}

func toTicketMessageResponse(m *domain.TicketMessage) ticketMessageResponse {
	attachments := m.Attachments
	if attachments == nil {
		attachments = []string{}
	}
	return ticketMessageResponse{
		ID:          m.ID.String(),
		TicketID:    m.TicketID.String(),
		SenderID:    m.SenderID.String(),
		SenderType:  string(m.SenderType),
		Body:        m.Body,
		Attachments: attachments,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}

// CreateTicket godoc
// @Summary      Create a support ticket
// @Description  Create a new support ticket with an initial message
// @Tags         support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createTicketRequest  true  "Ticket data"
// @Success      201   {object}  APIResponse{data=ticketResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Router       /my/tickets [post]
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createTicketRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	ticket := &domain.Ticket{
		ID:       uuid.New(),
		Category: domain.TicketCategory(req.Category),
		Subject:  req.Subject,
	}

	if req.BookingID != nil {
		bookingID, err := uuid.Parse(*req.BookingID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_booking_id", "invalid booking ID")
			return
		}
		ticket.BookingID = &bookingID
	}

	if err := h.ticketService.CreateTicket(r.Context(), userID, ticket, req.Message); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTicketResponse(ticket))
}

// ListUserTickets godoc
// @Summary      List my support tickets
// @Description  Get all support tickets for the authenticated user
// @Tags         support
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "Page number"  default(1)
// @Param        page_size  query     int  false  "Page size"    default(20)
// @Success      200  {object}  APIResponse{data=[]ticketResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Router       /my/tickets [get]
func (h *TicketHandler) ListUserTickets(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	result, err := h.ticketService.ListUserTickets(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []ticketResponse
	for _, t := range result.Items {
		items = append(items, toTicketResponse(&t))
	}
	if items == nil {
		items = []ticketResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetTicket godoc
// @Summary      Get support ticket details
// @Description  Get a specific support ticket by ID
// @Tags         support
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Ticket ID (UUID)"
// @Success      200  {object}  APIResponse{data=ticketResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/tickets/{id} [get]
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	ticket, err := h.ticketService.GetTicket(r.Context(), userID, role, ticketID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTicketResponse(ticket))
}

// AddUserMessage godoc
// @Summary      Add message to ticket
// @Description  Add a message to a support ticket
// @Tags         support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Ticket ID (UUID)"
// @Param        body  body      addTicketMessageRequest  true  "Message data"
// @Success      201   {object}  APIResponse{data=ticketMessageResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /my/tickets/{id}/messages [post]
func (h *TicketHandler) AddUserMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	var req addTicketMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	msg, err := h.ticketService.AddMessage(r.Context(), userID, role, ticketID, req.Body, req.Attachments)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTicketMessageResponse(msg))
}

// ListMessages godoc
// @Summary      List ticket messages
// @Description  Get all messages for a support ticket
// @Tags         support
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Ticket ID (UUID)"
// @Success      200  {object}  APIResponse{data=[]ticketMessageResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/tickets/{id}/messages [get]
func (h *TicketHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	messages, err := h.ticketService.ListMessages(r.Context(), userID, role, ticketID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []ticketMessageResponse
	for _, m := range messages {
		items = append(items, toTicketMessageResponse(&m))
	}
	if items == nil {
		items = []ticketMessageResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// SubmitCSAT godoc
// @Summary      Submit CSAT score
// @Description  Submit customer satisfaction score for a resolved ticket
// @Tags         support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string           true  "Ticket ID (UUID)"
// @Param        body  body      submitCSATRequest true  "CSAT score (1-5)"
// @Success      200   {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /my/tickets/{id}/csat [post]
func (h *TicketHandler) SubmitCSAT(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	var req submitCSATRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.ticketService.SubmitCSAT(r.Context(), userID, ticketID, req.Score); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "CSAT submitted"})
}

// --- Admin endpoints ---

// AdminGetTicket godoc
// @Summary      Get ticket details (admin)
// @Description  Get a specific support ticket by ID as admin
// @Tags         support-admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Ticket ID (UUID)"
// @Success      200  {object}  APIResponse{data=ticketResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/{id} [get]
func (h *TicketHandler) AdminGetTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	ticket, err := h.ticketService.GetTicket(r.Context(), userID, role, ticketID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTicketResponse(ticket))
}

// AdminListTickets godoc
// @Summary      List all tickets (admin)
// @Description  List all support tickets with filters
// @Tags         support-admin
// @Produce      json
// @Security     BearerAuth
// @Param        status    query     string  false  "Filter by status"
// @Param        priority  query     string  false  "Filter by priority"
// @Param        level     query     string  false  "Filter by level"
// @Param        category  query     string  false  "Filter by category"
// @Param        page      query     int     false  "Page number"  default(1)
// @Param        page_size query     int     false  "Page size"    default(20)
// @Success      200  {object}  APIResponse{data=[]ticketResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /admin/tickets [get]
func (h *TicketHandler) AdminListTickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	filter := domain.TicketFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.TicketStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_status", "invalid ticket status")
			return
		}
		filter.Status = &status
	}
	if p := r.URL.Query().Get("priority"); p != "" {
		priority := domain.TicketPriority(p)
		if !priority.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_priority", "invalid ticket priority")
			return
		}
		filter.Priority = &priority
	}
	if l := r.URL.Query().Get("level"); l != "" {
		level := domain.TicketLevel(l)
		if !level.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_level", "invalid ticket level")
			return
		}
		filter.Level = &level
	}
	if c := r.URL.Query().Get("category"); c != "" {
		category := domain.TicketCategory(c)
		if !category.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_category", "invalid ticket category")
			return
		}
		filter.Category = &category
	}

	result, err := h.ticketService.ListAllTickets(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []ticketResponse
	for _, t := range result.Items {
		items = append(items, toTicketResponse(&t))
	}
	if items == nil {
		items = []ticketResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// AdminAssignTicket godoc
// @Summary      Assign ticket to admin
// @Description  Assign a support ticket to an admin user
// @Tags         support-admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string              true  "Ticket ID (UUID)"
// @Param        body  body      assignTicketRequest  true  "Assignment data"
// @Success      200   {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/{id}/assign [patch]
func (h *TicketHandler) AdminAssignTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	var req assignTicketRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	assignedTo, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_assigned_to", "invalid assigned_to ID")
		return
	}

	if err := h.ticketService.AssignTicket(r.Context(), ticketID, assignedTo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "ticket assigned"})
}

// AdminEscalateTicket godoc
// @Summary      Escalate ticket
// @Description  Escalate a ticket to the next support level
// @Tags         support-admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Ticket ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Failure      409  {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/{id}/escalate [patch]
func (h *TicketHandler) AdminEscalateTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	if err := h.ticketService.EscalateTicket(r.Context(), ticketID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "ticket escalated"})
}

// AdminResolveTicket godoc
// @Summary      Resolve ticket
// @Description  Mark a support ticket as resolved
// @Tags         support-admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Ticket ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Failure      409  {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/{id}/resolve [patch]
func (h *TicketHandler) AdminResolveTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	if err := h.ticketService.ResolveTicket(r.Context(), ticketID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "ticket resolved"})
}

// AdminAddMessage godoc
// @Summary      Add admin message to ticket
// @Description  Add an admin response to a support ticket
// @Tags         support-admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Ticket ID (UUID)"
// @Param        body  body      addTicketMessageRequest  true  "Message data"
// @Success      201   {object}  APIResponse{data=ticketMessageResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/{id}/messages [post]
func (h *TicketHandler) AdminAddMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid ticket ID")
		return
	}

	var req addTicketMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	msg, err := h.ticketService.AddMessage(r.Context(), userID, role, ticketID, req.Body, req.Attachments)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTicketMessageResponse(msg))
}

// AdminGetStats godoc
// @Summary      Get ticket statistics
// @Description  Get counts of tickets by status
// @Tags         support-admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  APIResponse{data=ticketStatsResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Router       /admin/tickets/stats [get]
func (h *TicketHandler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.ticketService.GetStats(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ticketStatsResponse{
		Open:       stats.Open,
		InProgress: stats.InProgress,
		Escalated:  stats.Escalated,
		Resolved:   stats.Resolved,
		Closed:     stats.Closed,
	})
}
