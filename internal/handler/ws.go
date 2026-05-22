package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/service"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Send channel buffer size per client.
	sendBufSize = 256
)

// WSHandler handles WebSocket connections for real-time notifications.
type WSHandler struct {
	hub         *notification.Hub
	authService middleware.AuthService
	chatService service.ChatService
	logger      *logger.Logger
	upgrader    websocket.Upgrader
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *notification.Hub, authService middleware.AuthService, chatService service.ChatService, log *logger.Logger, cfg *config.Config) *WSHandler {
	isDev := middleware.IsDevEnvironment(cfg.Environment)
	allowedOrigins := cfg.CORS.AllowedOrigins
	return &WSHandler{
		hub:         hub,
		authService: authService,
		chatService: chatService,
		logger:      log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if isDev {
					return true
				}

				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}

				for _, allowed := range allowedOrigins {
					if allowed == "*" || allowed == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// HandleWS godoc
//
//	@Summary		Connect notifications WebSocket
//	@Description	Upgrade HTTP to WebSocket for real-time notifications. Authentication uses the token query parameter because browsers cannot set custom headers on WebSocket connections.
//	@Tags			websocket
//	@Param			token	query		string	true	"JWT access token"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/ws/notifications [get]
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing token query parameter")
		return
	}

	userID, role, err := h.authService.ParseToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", "error", err)
		return
	}

	client := &notification.Client{
		UserID: userID,
		Role:   string(role),
		Send:   make(chan []byte, sendBufSize),
	}

	h.hub.Register(client)

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

// readPump pumps messages from the WebSocket connection.
// It handles chat client messages (subscribe, unsubscribe, typing) and control frames.
func (h *WSHandler) readPump(conn *websocket.Conn, client *notification.Client) {
	defer func() {
		h.hub.Unregister(client)
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		h.logger.Error("ws set read deadline failed", "error", err)
		return
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("ws unexpected close", "user_id", client.UserID, "error", err)
			}
			break
		}

		h.handleClientMessage(client, message)
	}
}

// handleClientMessage processes an incoming chat message from a WebSocket client.
func (h *WSHandler) handleClientMessage(client *notification.Client, raw []byte) {
	var msg notification.ChatClientMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		h.logger.Debug("ws invalid client message", "user_id", client.UserID, "error", err)
		return
	}

	convID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		h.logger.Debug("ws invalid conversation_id", "user_id", client.UserID, "value", msg.ConversationID)
		return
	}

	switch msg.Action {
	case notification.ChatActionSubscribe:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if !h.chatService.CanAccessConversation(ctx, client.UserID, domain.UserRole(client.Role), convID) {
			h.logger.Warn("ws unauthorized conversation subscribe attempt", "user_id", client.UserID, "conversation_id", convID)
			return
		}
		h.hub.SubscribeToConversation(client, convID)
	case notification.ChatActionUnsubscribe:
		h.hub.UnsubscribeFromConversation(client, convID)
	case notification.ChatActionTyping:
		if !h.hub.IsClientSubscribed(client, convID) {
			return
		}
		h.hub.BroadcastTypingIndicator(convID, client.UserID)
	default:
		h.logger.Debug("ws unknown action", "user_id", client.UserID, "action", msg.Action)
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// It also sends periodic pings to keep the connection alive.
func (h *WSHandler) writePump(conn *websocket.Conn, client *notification.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				h.logger.Error("ws set write deadline failed", "error", err)
				return
			}
			if !ok {
				// Hub closed the channel
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Warn("ws write failed", "user_id", client.UserID, "error", err)
				return
			}

		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				h.logger.Error("ws set write deadline failed", "error", err)
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
