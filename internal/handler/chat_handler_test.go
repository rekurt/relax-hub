package handler

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
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockChatService struct {
	startConversationFn func(ctx context.Context, clientID, bathhouseID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error)
	sendMessageFn       func(ctx context.Context, senderID uuid.UUID, role domain.UserRole, convID uuid.UUID, text string) (*domain.Message, error)
	listConversationsFn func(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	listMessagesFn      func(ctx context.Context, userID uuid.UUID, role domain.UserRole, convID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error)
	markAsReadFn        func(ctx context.Context, userID uuid.UUID, role domain.UserRole, convID uuid.UUID) error
	getUnreadCountFn    func(ctx context.Context, userID uuid.UUID, role domain.UserRole) (int64, error)
}

func (m *mockChatService) StartConversation(ctx context.Context, clientID, bathhouseID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error) {
	if m.startConversationFn != nil {
		return m.startConversationFn(ctx, clientID, bathhouseID, bookingID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockChatService) SendMessage(ctx context.Context, senderID uuid.UUID, role domain.UserRole, convID uuid.UUID, text string) (*domain.Message, error) {
	if m.sendMessageFn != nil {
		return m.sendMessageFn(ctx, senderID, role, convID, text)
	}
	return nil, domain.ErrNotFound
}

func (m *mockChatService) ListConversations(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
	if m.listConversationsFn != nil {
		return m.listConversationsFn(ctx, userID, role, page, pageSize)
	}
	return nil, domain.ErrNotFound
}

func (m *mockChatService) ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, convID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error) {
	if m.listMessagesFn != nil {
		return m.listMessagesFn(ctx, userID, role, convID, page, pageSize)
	}
	return nil, domain.ErrNotFound
}

func (m *mockChatService) MarkAsRead(ctx context.Context, userID uuid.UUID, role domain.UserRole, convID uuid.UUID) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(ctx, userID, role, convID)
	}
	return domain.ErrNotFound
}

func (m *mockChatService) GetUnreadCount(ctx context.Context, userID uuid.UUID, role domain.UserRole) (int64, error) {
	if m.getUnreadCountFn != nil {
		return m.getUnreadCountFn(ctx, userID, role)
	}
	return 0, nil
}

func TestChatHandler_StartConversation(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	convID := uuid.New()
	now := time.Now()

	chatSvc := &mockChatService{
		startConversationFn: func(ctx context.Context, clientID, bhID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error) {
			if clientID != userID || bhID != bathhouseID {
				t.Errorf("unexpected args: clientID=%v, bhID=%v", clientID, bhID)
			}
			return &domain.Conversation{
				ID:          convID,
				BathhouseID: bathhouseID,
				ClientID:    userID,
				CreatedAt:   now,
			}, nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/bathhouses/{id}/chat", h.StartConversation)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/bathhouses/"+bathhouseID.String()+"/chat", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["id"] != convID.String() {
		t.Errorf("expected id %s, got %v", convID.String(), data["id"])
	}
	if data["bathhouse_id"] != bathhouseID.String() {
		t.Errorf("expected bathhouse_id %s, got %v", bathhouseID.String(), data["bathhouse_id"])
	}
}

func TestChatHandler_StartConversation_InvalidBathhouseID(t *testing.T) {
	h := NewChatHandler(&mockChatService{})
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/bathhouses/{id}/chat", h.StartConversation)

	req := httptest.NewRequest(http.MethodPost, "/bathhouses/invalid-uuid/chat", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_StartConversation_ServiceError(t *testing.T) {
	chatSvc := &mockChatService{
		startConversationFn: func(ctx context.Context, clientID, bhID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/bathhouses/{id}/chat", h.StartConversation)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/bathhouses/"+bathhouseID.String()+"/chat", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestChatHandler_ListConversations(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	now := time.Now()

	chatSvc := &mockChatService{
		listConversationsFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
			return &domain.PaginatedResult[domain.Conversation]{
				Items: []domain.Conversation{
					{
						ID:          convID,
						BathhouseID: uuid.New(),
						ClientID:    userID,
						CreatedAt:   now,
					},
				},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/conversations", h.ListConversations)

	req := httptest.NewRequest(http.MethodGet, "/my/conversations?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count 1, got %d", resp.Meta.TotalCount)
	}

	items, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(items))
	}
}

func TestChatHandler_ListMessages(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	msgID := uuid.New()
	now := time.Now()

	chatSvc := &mockChatService{
		listMessagesFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, cID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error) {
			if cID != convID {
				t.Errorf("expected convID %s, got %s", convID, cID)
			}
			return &domain.PaginatedResult[domain.Message]{
				Items: []domain.Message{
					{
						ID:             msgID,
						ConversationID: convID,
						SenderID:       userID,
						Text:           "Hello",
						IsRead:         false,
						CreatedAt:      now,
					},
				},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/conversations/{id}/messages", h.ListMessages)

	req := httptest.NewRequest(http.MethodGet, "/conversations/"+convID.String()+"/messages?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	items, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 message, got %d", len(items))
	}

	msg := items[0].(map[string]interface{})
	if msg["text"] != "Hello" {
		t.Errorf("expected text 'Hello', got %v", msg["text"])
	}
}

func TestChatHandler_ListMessages_InvalidConvID(t *testing.T) {
	h := NewChatHandler(&mockChatService{})
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/conversations/{id}/messages", h.ListMessages)

	req := httptest.NewRequest(http.MethodGet, "/conversations/bad-id/messages", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_SendMessage(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	msgID := uuid.New()
	now := time.Now()

	chatSvc := &mockChatService{
		sendMessageFn: func(ctx context.Context, senderID uuid.UUID, role domain.UserRole, cID uuid.UUID, text string) (*domain.Message, error) {
			if senderID != userID || cID != convID || text != "Hi there" {
				t.Errorf("unexpected args: sender=%v, conv=%v, text=%v", senderID, cID, text)
			}
			return &domain.Message{
				ID:             msgID,
				ConversationID: convID,
				SenderID:       userID,
				Text:           text,
				CreatedAt:      now,
			}, nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/conversations/{id}/messages", h.SendMessage)

	body, _ := json.Marshal(sendMessageRequest{Text: "Hi there"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/"+convID.String()+"/messages", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data := resp.Data.(map[string]interface{})
	if data["text"] != "Hi there" {
		t.Errorf("expected text 'Hi there', got %v", data["text"])
	}
}

func TestChatHandler_SendMessage_InvalidBody(t *testing.T) {
	h := NewChatHandler(&mockChatService{})
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/conversations/{id}/messages", h.SendMessage)

	convID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/conversations/"+convID.String()+"/messages", bytes.NewReader([]byte("not json")))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_SendMessage_Forbidden(t *testing.T) {
	chatSvc := &mockChatService{
		sendMessageFn: func(ctx context.Context, senderID uuid.UUID, role domain.UserRole, cID uuid.UUID, text string) (*domain.Message, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/conversations/{id}/messages", h.SendMessage)

	convID := uuid.New()
	body, _ := json.Marshal(sendMessageRequest{Text: "test"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/"+convID.String()+"/messages", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestChatHandler_MarkAsRead(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()

	chatSvc := &mockChatService{
		markAsReadFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, cID uuid.UUID) error {
			if uid != userID || cID != convID {
				t.Errorf("unexpected args: uid=%v, cID=%v", uid, cID)
			}
			return nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Patch("/conversations/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/conversations/"+convID.String()+"/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestChatHandler_MarkAsRead_InvalidConvID(t *testing.T) {
	h := NewChatHandler(&mockChatService{})
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Patch("/conversations/{id}/read", h.MarkAsRead)

	req := httptest.NewRequest(http.MethodPatch, "/conversations/bad-id/read", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_GetUnreadCount(t *testing.T) {
	userID := uuid.New()

	chatSvc := &mockChatService{
		getUnreadCountFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole) (int64, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return 5, nil
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/unread-messages-count", h.GetUnreadCount)

	req := httptest.NewRequest(http.MethodGet, "/my/unread-messages-count", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data := resp.Data.(map[string]interface{})
	if data["unread_count"].(float64) != 5 {
		t.Errorf("expected unread_count 5, got %v", data["unread_count"])
	}
}

func TestChatHandler_GetUnreadCount_ServiceError(t *testing.T) {
	chatSvc := &mockChatService{
		getUnreadCountFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole) (int64, error) {
			return 0, domain.ErrNotFound
		},
	}

	h := NewChatHandler(chatSvc)
	authService := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/unread-messages-count", h.GetUnreadCount)

	req := httptest.NewRequest(http.MethodGet, "/my/unread-messages-count", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestChatHandler_Unauthorized(t *testing.T) {
	h := NewChatHandler(&mockChatService{})
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/conversations", h.ListConversations)

	req := httptest.NewRequest(http.MethodGet, "/my/conversations", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
