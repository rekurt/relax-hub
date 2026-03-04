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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
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

func TestWSHandler_MissingToken(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)
	auth := &mockWSAuthService{}
	h := handler.NewWSHandler(hub, auth, log)

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
	h := handler.NewWSHandler(hub, auth, log)

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
	h := handler.NewWSHandler(hub, auth, log)

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

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
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
	h := handler.NewWSHandler(hub, auth, log)

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
