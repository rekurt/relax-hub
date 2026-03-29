package notification

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// WSMessage represents a notification message sent over WebSocket.
type WSMessage struct {
	Type      string            `json:"type"`
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ChatWSMessage represents a chat-related message sent over WebSocket.
type ChatWSMessage struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id,omitempty"`
	SenderID       string `json:"sender_id,omitempty"`
	Text           string `json:"text,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

// Chat WebSocket message types.
const (
	ChatMsgNewMessage      = "new_message"
	ChatMsgMessageRead     = "message_read"
	ChatMsgTypingIndicator = "typing_indicator"
)

// ChatClientMessage represents an incoming message from a WebSocket client.
type ChatClientMessage struct {
	Action         string `json:"action"`
	ConversationID string `json:"conversation_id"`
}

// Chat client actions.
const (
	ChatActionSubscribe   = "subscribe"
	ChatActionUnsubscribe = "unsubscribe"
	ChatActionTyping      = "typing"
)

// Client represents a connected WebSocket client.
type Client struct {
	UserID uuid.UUID
	Role   string
	Send   chan []byte
}

// Hub manages WebSocket client connections and broadcasts notifications.
type Hub struct {
	mu        sync.RWMutex
	clients   map[uuid.UUID]map[*Client]struct{}
	chatRooms map[uuid.UUID]map[*Client]struct{} // conversationID -> clients
	logger    *logger.Logger
}

// NewHub creates a new WebSocket Hub.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:   make(map[uuid.UUID]map[*Client]struct{}),
		chatRooms: make(map[uuid.UUID]map[*Client]struct{}),
		logger:    log,
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
	h.logger.Debug("ws client registered", "user_id", client.UserID)
}

// Unregister removes a client from the hub and closes its send channel.
// Also removes the client from any chat rooms it was subscribed to.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[client.UserID]; ok {
		if _, exists := conns[client]; exists {
			delete(conns, client)
			close(client.Send)
			if len(conns) == 0 {
				delete(h.clients, client.UserID)
			}
		}
	}

	// Remove client from all chat rooms
	for roomID, members := range h.chatRooms {
		if _, exists := members[client]; exists {
			delete(members, client)
			if len(members) == 0 {
				delete(h.chatRooms, roomID)
			}
		}
	}

	h.logger.Debug("ws client unregistered", "user_id", client.UserID)
}

// SendToUser sends a notification to all connected clients of a specific user.
// Returns true if the user had at least one active connection.
func (h *Hub) SendToUser(userID uuid.UUID, notif *domain.Notification) bool {
	msg := WSMessage{
		Type:      string(notif.Type),
		ID:        notif.ID.String(),
		Title:     notif.Title,
		Body:      notif.Body,
		Data:      notif.Data,
		CreatedAt: notif.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal ws message", "error", err)
		return false
	}

	h.mu.RLock()
	conns, ok := h.clients[userID]
	if !ok || len(conns) == 0 {
		h.mu.RUnlock()
		return false
	}

	// Copy client list under read lock
	clients := make([]*Client, 0, len(conns))
	for c := range conns {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	sent := false
	for _, c := range clients {
		func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("panic recovered while sending to ws client", "user_id", userID, "panic", r)
				}
			}()
			select {
			case c.Send <- data:
				sent = true
			default:
				// Client's send buffer is full, skip
				h.logger.Warn("ws client send buffer full, skipping", "user_id", userID)
			}
		}()
	}

	return sent
}

// IsOnline checks if a user has any active WebSocket connections.
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userID]
	return ok && len(conns) > 0
}

// OnlineCount returns the number of users with active connections.
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SubscribeToConversation adds a client to a chat room.
func (h *Hub) SubscribeToConversation(client *Client, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chatRooms[conversationID] == nil {
		h.chatRooms[conversationID] = make(map[*Client]struct{})
	}
	h.chatRooms[conversationID][client] = struct{}{}
	h.logger.Debug("client subscribed to conversation", "user_id", client.UserID, "conversation_id", conversationID)
}

// IsClientSubscribed checks if a client is subscribed to a conversation room.
func (h *Hub) IsClientSubscribed(client *Client, conversationID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if members, ok := h.chatRooms[conversationID]; ok {
		_, subscribed := members[client]
		return subscribed
	}
	return false
}

// UnsubscribeFromConversation removes a client from a chat room.
func (h *Hub) UnsubscribeFromConversation(client *Client, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members, ok := h.chatRooms[conversationID]; ok {
		delete(members, client)
		if len(members) == 0 {
			delete(h.chatRooms, conversationID)
		}
	}
	h.logger.Debug("client unsubscribed from conversation", "user_id", client.UserID, "conversation_id", conversationID)
}

// SendToConversation sends a chat message to all clients subscribed to a conversation.
func (h *Hub) SendToConversation(conversationID uuid.UUID, msg *ChatWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal chat ws message", "error", err)
		return
	}

	h.mu.RLock()
	members, ok := h.chatRooms[conversationID]
	if !ok || len(members) == 0 {
		h.mu.RUnlock()
		return
	}

	clients := make([]*Client, 0, len(members))
	for c := range members {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("panic recovered while sending to chat ws client", "user_id", c.UserID, "conversation_id", conversationID, "panic", r)
				}
			}()
			select {
			case c.Send <- data:
			default:
				h.logger.Warn("chat ws client send buffer full, skipping", "user_id", c.UserID)
			}
		}()
	}
}

// BroadcastNewMessage broadcasts a new chat message to conversation subscribers.
func (h *Hub) BroadcastNewMessage(conversationID uuid.UUID, msg *domain.Message) {
	chatMsg := &ChatWSMessage{
		Type:           ChatMsgNewMessage,
		ConversationID: conversationID.String(),
		MessageID:      msg.ID.String(),
		SenderID:       msg.SenderID.String(),
		Text:           msg.Text,
		CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
	}
	h.SendToConversation(conversationID, chatMsg)
}

// BroadcastMessageRead broadcasts a message_read event to conversation subscribers.
func (h *Hub) BroadcastMessageRead(conversationID uuid.UUID, userID uuid.UUID) {
	chatMsg := &ChatWSMessage{
		Type:           ChatMsgMessageRead,
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
	}
	h.SendToConversation(conversationID, chatMsg)
}

// BroadcastTypingIndicator broadcasts a typing indicator to conversation subscribers.
func (h *Hub) BroadcastTypingIndicator(conversationID uuid.UUID, userID uuid.UUID) {
	chatMsg := &ChatWSMessage{
		Type:           ChatMsgTypingIndicator,
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
	}
	h.SendToConversation(conversationID, chatMsg)
}

// ConversationSubscriberCount returns the number of clients subscribed to a conversation.
func (h *Hub) ConversationSubscriberCount(conversationID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.chatRooms[conversationID])
}
