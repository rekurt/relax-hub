package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/notification"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS is handled by middleware
	},
}

// WSHandler handles WebSocket connections for real-time notifications.
type WSHandler struct {
	hub         *notification.Hub
	authService middleware.AuthService
	logger      *logger.Logger
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *notification.Hub, authService middleware.AuthService, log *logger.Logger) *WSHandler {
	return &WSHandler{
		hub:         hub,
		authService: authService,
		logger:      log,
	}
}

// HandleWS upgrades HTTP to WebSocket and manages the connection.
// Authentication is done via ?token= query parameter since browsers
// cannot set custom headers on WebSocket connections.
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing token query parameter")
		return
	}

	userID, _, err := h.authService.ParseToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", "error", err)
		return
	}

	client := &notification.Client{
		UserID: userID,
		Send:   make(chan []byte, sendBufSize),
	}

	h.hub.Register(client)

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

// readPump pumps messages from the WebSocket connection.
// It only handles control frames (pong) -- we don't expect data from clients.
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
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("ws unexpected close", "user_id", client.UserID, "error", err)
			}
			break
		}
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
				conn.WriteMessage(websocket.CloseMessage, []byte{})
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
