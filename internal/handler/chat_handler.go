package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

type startConversationRequest struct {
	BookingID *string `json:"booking_id,omitempty"`
}

type sendMessageRequest struct {
	Text string `json:"text"`
}

type conversationResponse struct {
	ID            string     `json:"id"`
	BathhouseID   string     `json:"bathhouse_id"`
	ClientID      string     `json:"client_id"`
	BookingID     *string    `json:"booking_id,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type messageResponse struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversation_id"`
	SenderID       string     `json:"sender_id"`
	Text           string     `json:"text"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func toConversationResponse(c *domain.Conversation) conversationResponse {
	resp := conversationResponse{
		ID:            c.ID.String(),
		BathhouseID:   c.BathhouseID.String(),
		ClientID:      c.ClientID.String(),
		LastMessageAt: c.LastMessageAt,
		CreatedAt:     c.CreatedAt,
	}
	if c.BookingID != nil {
		s := c.BookingID.String()
		resp.BookingID = &s
	}
	return resp
}

func toMessageResponse(m *domain.Message) messageResponse {
	return messageResponse{
		ID:             m.ID.String(),
		ConversationID: m.ConversationID.String(),
		SenderID:       m.SenderID.String(),
		Text:           m.Text,
		IsRead:         m.IsRead,
		ReadAt:         m.ReadAt,
		CreatedAt:      m.CreatedAt,
	}
}

// StartConversation handles POST /api/v1/bathhouses/{id}/chat
func (h *ChatHandler) StartConversation(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var bookingID *uuid.UUID
	var req startConversationRequest
	if r.ContentLength > 0 {
		if err := readJSON(w, r, &req); err != nil {
			handleServiceError(w, err)
			return
		}
		if req.BookingID != nil {
			id, err := uuid.Parse(*req.BookingID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking_id")
				return
			}
			bookingID = &id
		}
	}

	userID := middleware.GetUserID(r.Context())

	conv, err := h.chatService.StartConversation(r.Context(), userID, bathhouseID, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toConversationResponse(conv))
}

// ListConversations handles GET /api/v1/my/conversations
func (h *ChatHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.chatService.ListConversations(r.Context(), userID, userRole, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]conversationResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toConversationResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// ListMessages handles GET /api/v1/conversations/{id}/messages
func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid conversation id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.chatService.ListMessages(r.Context(), userID, userRole, convID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]messageResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toMessageResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// SendMessage handles POST /api/v1/conversations/{id}/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid conversation id")
		return
	}

	var req sendMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	msg, err := h.chatService.SendMessage(r.Context(), userID, userRole, convID, req.Text)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toMessageResponse(msg))
}

// MarkAsRead handles PATCH /api/v1/conversations/{id}/read
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid conversation id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.chatService.MarkAsRead(r.Context(), userID, userRole, convID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

// GetUnreadCount handles GET /api/v1/my/unread-messages-count
func (h *ChatHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	count, err := h.chatService.GetUnreadCount(r.Context(), userID, userRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}
