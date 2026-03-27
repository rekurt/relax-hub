package handler_test

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
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

// --- mock ChatFilter ---

type mockChatFilter struct {
	filterFn       func(ctx context.Context, text string) antifraud.FilterResult
	logFilteredFn  func(ctx context.Context, conversationID, senderID uuid.UUID, original string, result antifraud.FilterResult)
	listFilteredFn func(ctx context.Context, page, pageSize int) ([]antifraud.FilteredChatMessage, int64, error)
}

func (m *mockChatFilter) Filter(ctx context.Context, text string) antifraud.FilterResult {
	if m.filterFn != nil {
		return m.filterFn(ctx, text)
	}
	return antifraud.FilterResult{}
}

func (m *mockChatFilter) LogFiltered(ctx context.Context, conversationID, senderID uuid.UUID, original string, result antifraud.FilterResult) {
	if m.logFilteredFn != nil {
		m.logFilteredFn(ctx, conversationID, senderID, original, result)
	}
}

func (m *mockChatFilter) ListFiltered(ctx context.Context, page, pageSize int) ([]antifraud.FilteredChatMessage, int64, error) {
	if m.listFilteredFn != nil {
		return m.listFilteredFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

var _ antifraud.ChatFilter = (*mockChatFilter)(nil)

// --- helpers ---

func seedFraudFlag(t *testing.T, repo *mock.FraudFlagRepo, userID uuid.UUID, rule domain.FraudRuleName, status domain.FraudFlagStatus) *domain.FraudFlag {
	t.Helper()
	flag := &domain.FraudFlag{
		ID:        uuid.New(),
		UserID:    userID,
		Rule:      rule,
		Severity:  domain.FraudSeverityMedium,
		Status:    status,
		Action:    domain.FraudActionFlag,
		CreatedAt: time.Now(),
	}
	if err := repo.Create(context.Background(), flag); err != nil {
		t.Fatal(err)
	}
	return flag
}

func newFraudFlagRepo() *mock.FraudFlagRepo {
	return mock.NewFraudFlagRepo().(*mock.FraudFlagRepo)
}

// --- TestAntiFraudHandler_ListFlags ---

func TestAntiFraudHandler_ListFlags_DefaultPagination(t *testing.T) {
	repo := newFraudFlagRepo()
	userID := uuid.New()
	seedFraudFlag(t, repo, userID, domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)
	seedFraudFlag(t, repo, userID, domain.FraudRuleSelfBooking, domain.FraudFlagStatusPending)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta")
	}
	if resp.Meta.Page != 1 {
		t.Errorf("expected page=1, got %d", resp.Meta.Page)
	}
	if resp.Meta.PageSize != 20 {
		t.Errorf("expected page_size=20, got %d", resp.Meta.PageSize)
	}
	if resp.Meta.TotalCount != 2 {
		t.Errorf("expected total_count=2, got %d", resp.Meta.TotalCount)
	}
}

func TestAntiFraudHandler_ListFlags_WithStatusFilter(t *testing.T) {
	repo := newFraudFlagRepo()
	userID := uuid.New()
	seedFraudFlag(t, repo, userID, domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)
	seedFraudFlag(t, repo, userID, domain.FraudRuleSelfBooking, domain.FraudFlagStatusReviewed)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?status=reviewed", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected 1 reviewed flag, got %d", resp.Meta.TotalCount)
	}
}

func TestAntiFraudHandler_ListFlags_WithRuleFilter(t *testing.T) {
	repo := newFraudFlagRepo()
	userID := uuid.New()
	seedFraudFlag(t, repo, userID, domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)
	seedFraudFlag(t, repo, userID, domain.FraudRuleSelfBooking, domain.FraudFlagStatusPending)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?rule=RULE_SELF_BOOKING", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected 1 flag with rule filter, got %d", resp.Meta.TotalCount)
	}
}

func TestAntiFraudHandler_ListFlags_WithUserIDFilter(t *testing.T) {
	repo := newFraudFlagRepo()
	user1 := uuid.New()
	user2 := uuid.New()
	seedFraudFlag(t, repo, user1, domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)
	seedFraudFlag(t, repo, user2, domain.FraudRuleSelfBooking, domain.FraudFlagStatusPending)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?user_id="+user1.String(), nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected 1 flag for user1, got %d", resp.Meta.TotalCount)
	}
}

func TestAntiFraudHandler_ListFlags_InvalidStatus(t *testing.T) {
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?status=bogus", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAntiFraudHandler_ListFlags_InvalidRule(t *testing.T) {
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?rule=INVALID_RULE", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAntiFraudHandler_ListFlags_InvalidUserID(t *testing.T) {
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), &mockChatFilter{})

	req := httptest.NewRequest(http.MethodGet, "/admin/antifraud/flags?user_id=not-a-uuid", nil)
	w := httptest.NewRecorder()
	h.ListFlags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- TestAntiFraudHandler_UpdateFlag ---

func TestAntiFraudHandler_UpdateFlag_Success(t *testing.T) {
	repo := newFraudFlagRepo()
	userID := uuid.New()
	flag := seedFraudFlag(t, repo, userID, domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})
	adminID := uuid.New()

	body, _ := json.Marshal(map[string]string{"status": "reviewed"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/antifraud/flags/"+flag.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), adminID)
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Patch("/admin/antifraud/flags/{id}", h.UpdateFlag)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAntiFraudHandler_UpdateFlag_InvalidID(t *testing.T) {
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), &mockChatFilter{})

	body, _ := json.Marshal(map[string]string{"status": "reviewed"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/antifraud/flags/not-a-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r := chi.NewRouter()
	r.Patch("/admin/antifraud/flags/{id}", h.UpdateFlag)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAntiFraudHandler_UpdateFlag_InvalidStatus(t *testing.T) {
	repo := newFraudFlagRepo()
	flag := seedFraudFlag(t, repo, uuid.New(), domain.FraudRuleMultiCardTopUp, domain.FraudFlagStatusPending)

	h := handler.NewAntiFraudHandler(repo, &mockChatFilter{})

	body, _ := json.Marshal(map[string]string{"status": "bogus"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/antifraud/flags/"+flag.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), uuid.New())
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Patch("/admin/antifraud/flags/{id}", h.UpdateFlag)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAntiFraudHandler_UpdateFlag_NotFound(t *testing.T) {
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), &mockChatFilter{})

	body, _ := json.Marshal(map[string]string{"status": "reviewed"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/antifraud/flags/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), uuid.New())
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Patch("/admin/antifraud/flags/{id}", h.UpdateFlag)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- TestAntiFraudHandler_ListFilteredMessages ---

func TestAntiFraudHandler_ListFilteredMessages_DefaultPagination(t *testing.T) {
	cf := &mockChatFilter{
		listFilteredFn: func(_ context.Context, page, pageSize int) ([]antifraud.FilteredChatMessage, int64, error) {
			if page != 1 || pageSize != 20 {
				t.Errorf("expected page=1,pageSize=20; got page=%d,pageSize=%d", page, pageSize)
			}
			return []antifraud.FilteredChatMessage{
				{ID: uuid.New(), ConversationID: uuid.New(), SenderID: uuid.New(), OriginalText: "test", CreatedAt: time.Now()},
			}, 1, nil
		},
	}
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), cf)

	req := httptest.NewRequest(http.MethodGet, "/admin/chat/filtered", nil)
	w := httptest.NewRecorder()
	h.ListFilteredMessages(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp handler.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Meta == nil {
		t.Fatal("expected meta")
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count=1, got %d", resp.Meta.TotalCount)
	}
}

func TestAntiFraudHandler_ListFilteredMessages_CustomPagination(t *testing.T) {
	cf := &mockChatFilter{
		listFilteredFn: func(_ context.Context, page, pageSize int) ([]antifraud.FilteredChatMessage, int64, error) {
			if page != 2 || pageSize != 5 {
				t.Errorf("expected page=2,pageSize=5; got page=%d,pageSize=%d", page, pageSize)
			}
			return nil, 0, nil
		},
	}
	h := handler.NewAntiFraudHandler(newFraudFlagRepo(), cf)

	req := httptest.NewRequest(http.MethodGet, "/admin/chat/filtered?page=2&page_size=5", nil)
	w := httptest.NewRecorder()
	h.ListFilteredMessages(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
