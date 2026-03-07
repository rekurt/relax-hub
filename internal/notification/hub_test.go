package notification_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	userID := uuid.New()
	client := &notification.Client{
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	hub.Register(client)
	if !hub.IsOnline(userID) {
		t.Error("user should be online after register")
	}
	if hub.OnlineCount() != 1 {
		t.Errorf("online count = %d, want 1", hub.OnlineCount())
	}

	hub.Unregister(client)
	if hub.IsOnline(userID) {
		t.Error("user should be offline after unregister")
	}
	if hub.OnlineCount() != 0 {
		t.Errorf("online count = %d, want 0", hub.OnlineCount())
	}
}

func TestHub_MultipleClientsPerUser(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	userID := uuid.New()
	client1 := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}
	client2 := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}

	hub.Register(client1)
	hub.Register(client2)

	if hub.OnlineCount() != 1 {
		t.Errorf("online count = %d, want 1 (same user)", hub.OnlineCount())
	}

	hub.Unregister(client1)
	if !hub.IsOnline(userID) {
		t.Error("user should still be online with second client")
	}

	hub.Unregister(client2)
	if hub.IsOnline(userID) {
		t.Error("user should be offline after both clients unregister")
	}
}

func TestHub_SendToUser_Online(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	userID := uuid.New()
	client := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)

	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.NotifBookingConfirmed,
		Title:     "Test",
		Body:      "Test body",
		CreatedAt: time.Now(),
	}

	sent := hub.SendToUser(userID, notif)
	if !sent {
		t.Error("SendToUser should return true for online user")
	}

	select {
	case msg := <-client.Send:
		if len(msg) == 0 {
			t.Error("received empty message")
		}
	case <-time.After(time.Second):
		t.Error("did not receive message in time")
	}
}

func TestHub_SendToUser_Offline(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	offlineUserID := uuid.New()
	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    offlineUserID,
		Type:      domain.NotifSystem,
		Title:     "Test",
		Body:      "Test body",
		CreatedAt: time.Now(),
	}

	sent := hub.SendToUser(offlineUserID, notif)
	if sent {
		t.Error("SendToUser should return false for offline user")
	}
}

func TestHub_SendToUser_MultipleClients(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	userID := uuid.New()
	client1 := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}
	client2 := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}
	hub.Register(client1)
	hub.Register(client2)
	defer hub.Unregister(client1)
	defer hub.Unregister(client2)

	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.NotifSystem,
		Title:     "Multi",
		Body:      "Multi body",
		CreatedAt: time.Now(),
	}

	sent := hub.SendToUser(userID, notif)
	if !sent {
		t.Error("SendToUser should return true")
	}

	// Both clients should receive the message
	for i, c := range []*notification.Client{client1, client2} {
		select {
		case msg := <-c.Send:
			if len(msg) == 0 {
				t.Errorf("client%d: received empty message", i+1)
			}
		case <-time.After(time.Second):
			t.Errorf("client%d: did not receive message in time", i+1)
		}
	}
}

func TestHub_IsOnline_NotRegistered(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	if hub.IsOnline(uuid.New()) {
		t.Error("unknown user should not be online")
	}
}

func TestHub_UnregisterUnknownClient(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	// Should not panic
	client := &notification.Client{
		UserID: uuid.New(),
		Send:   make(chan []byte, 256),
	}
	hub.Unregister(client)
}

func TestHub_SubscribeToConversation(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)

	hub.SubscribeToConversation(client, convID)

	if hub.ConversationSubscriberCount(convID) != 1 {
		t.Errorf("subscriber count = %d, want 1", hub.ConversationSubscriberCount(convID))
	}
}

func TestHub_UnsubscribeFromConversation(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)

	hub.SubscribeToConversation(client, convID)
	hub.UnsubscribeFromConversation(client, convID)

	if hub.ConversationSubscriberCount(convID) != 0 {
		t.Errorf("subscriber count = %d, want 0", hub.ConversationSubscriberCount(convID))
	}
}

func TestHub_SendToConversation(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	client1 := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	client2 := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client1)
	hub.Register(client2)
	defer hub.Unregister(client1)
	defer hub.Unregister(client2)

	hub.SubscribeToConversation(client1, convID)
	hub.SubscribeToConversation(client2, convID)

	msg := &notification.ChatWSMessage{
		Type:           notification.ChatMsgNewMessage,
		ConversationID: convID.String(),
		MessageID:      uuid.New().String(),
		SenderID:       client1.UserID.String(),
		Text:           "Hello!",
	}
	hub.SendToConversation(convID, msg)

	// Both clients should receive
	for i, c := range []*notification.Client{client1, client2} {
		select {
		case data := <-c.Send:
			if len(data) == 0 {
				t.Errorf("client%d: received empty message", i+1)
			}
		case <-time.After(time.Second):
			t.Errorf("client%d: did not receive chat message in time", i+1)
		}
	}
}

func TestHub_SendToConversation_NoSubscribers(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	msg := &notification.ChatWSMessage{
		Type:           notification.ChatMsgNewMessage,
		ConversationID: uuid.New().String(),
		Text:           "Nobody here",
	}
	// Should not panic
	hub.SendToConversation(uuid.New(), msg)
}

func TestHub_BroadcastNewMessage(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)
	hub.SubscribeToConversation(client, convID)

	msg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		SenderID:       uuid.New(),
		Text:           "Broadcast test",
		CreatedAt:      time.Now(),
	}
	hub.BroadcastNewMessage(convID, msg)

	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("received empty message")
		}
	case <-time.After(time.Second):
		t.Error("did not receive broadcast message in time")
	}
}

func TestHub_BroadcastMessageRead(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	userID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)
	hub.SubscribeToConversation(client, convID)

	hub.BroadcastMessageRead(convID, userID)

	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("received empty message")
		}
	case <-time.After(time.Second):
		t.Error("did not receive read receipt in time")
	}
}

func TestHub_BroadcastTypingIndicator(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	userID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)
	hub.SubscribeToConversation(client, convID)

	hub.BroadcastTypingIndicator(convID, userID)

	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("received empty message")
		}
	case <-time.After(time.Second):
		t.Error("did not receive typing indicator in time")
	}
}

func TestHub_UnregisterCleansUpChatRooms(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	convID := uuid.New()
	client := &notification.Client{UserID: uuid.New(), Send: make(chan []byte, 256)}
	hub.Register(client)
	hub.SubscribeToConversation(client, convID)

	if hub.ConversationSubscriberCount(convID) != 1 {
		t.Fatalf("subscriber count = %d, want 1", hub.ConversationSubscriberCount(convID))
	}

	hub.Unregister(client)

	if hub.ConversationSubscriberCount(convID) != 0 {
		t.Errorf("subscriber count = %d after unregister, want 0", hub.ConversationSubscriberCount(convID))
	}
}

func TestHub_Dispatch_SendsViaWebSocket(t *testing.T) {
	log := logger.New(logger.LevelError)
	hub := notification.NewHub(log)

	userID := uuid.New()
	client := &notification.Client{UserID: userID, Send: make(chan []byte, 256)}
	hub.Register(client)
	defer hub.Unregister(client)

	notifRepo := mock.NewNotificationRepo()
	emailSender := notification.NewNoopEmailSender()
	d := notification.NewDispatcher(notifRepo, emailSender, notification.NewNoopTelegramSender(), mock.NewTelegramLinkRepo(), hub, log)

	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.NotifSystem,
		Title:     "WS Test",
		Body:      "WS Test Body",
		CreatedAt: time.Now(),
	}
	prefs := &domain.NotificationPreferences{
		UserID: userID,
		InApp:  true,
	}

	d.Dispatch(t.Context(), notif, prefs, "")

	// Verify: notification was created in repo
	count, err := notifRepo.CountUnread(t.Context(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("createCount = %d, want 1", count)
	}

	// AND sent via WebSocket
	select {
	case msg := <-client.Send:
		if len(msg) == 0 {
			t.Error("received empty message via WS")
		}
	case <-time.After(time.Second):
		t.Error("did not receive WS message in time")
	}
}
