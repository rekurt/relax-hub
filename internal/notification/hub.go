package notification

import (
	"encoding/json"
	"sync"

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

// Client represents a connected WebSocket client.
type Client struct {
	UserID uuid.UUID
	Send   chan []byte
}

// Hub manages WebSocket client connections and broadcasts notifications.
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
	logger  *logger.Logger
}

// NewHub creates a new WebSocket Hub.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]struct{}),
		logger:  log,
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
		select {
		case c.Send <- data:
			sent = true
		default:
			// Client's send buffer is full, skip
			h.logger.Warn("ws client send buffer full, skipping", "user_id", userID)
		}
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
