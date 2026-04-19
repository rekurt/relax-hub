package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/handler"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
)

type mockWSAuthService struct {
	parseTokenFn func(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

func (m *mockWSAuthService) ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error) {
	if m.parseTokenFn != nil {
		return m.parseTokenFn(ctx, token)
	}
	return uuid.Nil, "", domain.ErrUnauthorized
}

func (m *mockWSAuthService) ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	userID, role, err := m.ParseToken(ctx, token)
	return userID, role, uuid.Nil, err
}

type mockChatServiceForWS struct {
	canAccessFn func(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID) bool
}

func (m *mockChatServiceForWS) StartConversation(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ *uuid.UUID) (*domain.Conversation, error) {
	return nil, nil
}
func (m *mockChatServiceForWS) SendMessage(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) (*domain.Message, error) {
	return nil, nil
}
func (m *mockChatServiceForWS) ListConversations(_ context.Context, _ uuid.UUID, _ domain.UserRole, _, _ int) (*domain.PaginatedResult[domain.Conversation], error) {
	return nil, nil
}
func (m *mockChatServiceForWS) ListMessages(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Message], error) {
	return nil, nil
}
func (m *mockChatServiceForWS) MarkAsRead(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockChatServiceForWS) GetUnreadCount(_ context.Context, _ uuid.UUID, _ domain.UserRole) (int64, error) {
	return 0, nil
}
func (m *mockChatServiceForWS) CanAccessConversation(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID) bool {
	if m.canAccessFn != nil {
		return m.canAccessFn(ctx, userID, role, conversationID)
	}
	return true
}

func TestWSHandler_MissingToken(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	auth := &mockWSAuthService{}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/notifications", nil)
	w := httptest.NewRecorder()
	h.HandleWS(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWSHandler_InvalidToken(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	auth := &mockWSAuthService{
		parseTokenFn: func(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/notifications?token=invalid", nil)
	w := httptest.NewRecorder()
	h.HandleWS(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWSHandler_ValidConnection(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	userID := uuid.New()

	auth := &mockWSAuthService{
		parseTokenFn: func(_ context.Context, token string) (uuid.UUID, domain.UserRole, error) {
			if token == "valid-token" {
				return userID, domain.RoleClient, nil
			}
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	server := httptest.NewServer(http.HandlerFunc(h.HandleWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=valid-token"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Give Hub time to register
	time.Sleep(50 * time.Millisecond)

	if !hub.IsOnline(userID) {
		t.Error("user should be online after WS connection")
	}

	// Send a notification via hub and verify it arrives
	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.NotifSystem,
		Title:     "Test Notification",
		Body:      "Test body",
		CreatedAt: time.Now(),
	}
	sent := hub.SendToUser(userID, notif)
	if !sent {
		t.Error("SendToUser should return true")
	}

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}
	if len(msg) == 0 {
		t.Error("received empty message")
	}
	if !strings.Contains(string(msg), "Test Notification") {
		t.Errorf("message should contain notification title, got: %s", string(msg))
	}
}

func TestWSHandler_DisconnectCleansUp(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	userID := uuid.New()

	auth := &mockWSAuthService{
		parseTokenFn: func(_ context.Context, token string) (uuid.UUID, domain.UserRole, error) {
			if token == "valid-token" {
				return userID, domain.RoleClient, nil
			}
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	server := httptest.NewServer(http.HandlerFunc(h.HandleWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=valid-token"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if !hub.IsOnline(userID) {
		t.Error("user should be online")
	}

	conn.Close()

	// Wait for cleanup goroutine
	time.Sleep(100 * time.Millisecond)
	if hub.IsOnline(userID) {
		t.Error("user should be offline after disconnect")
	}
}

func TestWSHandler_ChatSubscribe(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	userID := uuid.New()
	convID := uuid.New()

	auth := &mockWSAuthService{
		parseTokenFn: func(_ context.Context, token string) (uuid.UUID, domain.UserRole, error) {
			if token == "valid-token" {
				return userID, domain.RoleClient, nil
			}
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	srv := httptest.NewServer(http.HandlerFunc(h.HandleWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "?token=valid-token"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Send subscribe message
	subMsg := `{"action":"subscribe","conversation_id":"` + convID.String() + `"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(subMsg)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if hub.ConversationSubscriberCount(convID) != 1 {
		t.Errorf("subscriber count = %d, want 1", hub.ConversationSubscriberCount(convID))
	}

	// Send a chat message to the conversation and verify delivery
	hub.BroadcastNewMessage(convID, &domain.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		SenderID:       uuid.New(),
		Text:           "Hello from WS!",
		CreatedAt:      time.Now(),
	})

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}
	if !strings.Contains(string(msg), "Hello from WS!") {
		t.Errorf("message should contain text, got: %s", string(msg))
	}
}

func TestWSHandler_ChatTypingIndicator(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	user1ID := uuid.New()
	user2ID := uuid.New()
	convID := uuid.New()

	auth := &mockWSAuthService{
		parseTokenFn: func(_ context.Context, token string) (uuid.UUID, domain.UserRole, error) {
			switch token {
			case "user1":
				return user1ID, domain.RoleClient, nil
			case "user2":
				return user2ID, domain.RoleOwner, nil
			}
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
	h := handler.NewWSHandler(hub, auth, &mockChatServiceForWS{}, log, &config.Config{Environment: "dev"})

	srv := httptest.NewServer(http.HandlerFunc(h.HandleWS))
	defer srv.Close()

	// Connect user1
	wsURL1 := "ws" + strings.TrimPrefix(srv.URL, "http") + "?token=user1"
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("dial user1 failed: %v", err)
	}
	defer conn1.Close()

	// Connect user2
	wsURL2 := "ws" + strings.TrimPrefix(srv.URL, "http") + "?token=user2"
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("dial user2 failed: %v", err)
	}
	defer conn2.Close()

	time.Sleep(50 * time.Millisecond)

	// Both subscribe to the conversation
	subMsg := `{"action":"subscribe","conversation_id":"` + convID.String() + `"}`
	_ = conn1.WriteMessage(websocket.TextMessage, []byte(subMsg))
	_ = conn2.WriteMessage(websocket.TextMessage, []byte(subMsg))

	time.Sleep(50 * time.Millisecond)

	// user1 sends typing indicator
	typingMsg := `{"action":"typing","conversation_id":"` + convID.String() + `"}`
	if err := conn1.WriteMessage(websocket.TextMessage, []byte(typingMsg)); err != nil {
		t.Fatalf("write typing failed: %v", err)
	}

	// user2 should receive typing indicator
	if err := conn2.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, msg, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}
	if !strings.Contains(string(msg), "typing_indicator") {
		t.Errorf("expected typing_indicator message, got: %s", string(msg))
	}
	if !strings.Contains(string(msg), user1ID.String()) {
		t.Errorf("typing indicator should contain user1 ID, got: %s", string(msg))
	}
}
