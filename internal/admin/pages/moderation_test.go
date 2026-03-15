package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockModerationProvider is a test implementation of ModerationDataProvider.
type mockModerationProvider struct {
	data             *ModerationData
	err              error
	approveErr       error
	rejectErr        error
	approvedIDs      []uuid.UUID
	rejectedIDs      []uuid.UUID
	rejectedReasons  []string
	batchApproveOK   int
	batchApproveFail int
	batchRejectOK    int
	batchRejectFail  int
}

func (m *mockModerationProvider) GetModerationData(_ context.Context, _ ModerationFilter, _, _ int) (*ModerationData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func (m *mockModerationProvider) ApproveReview(_ context.Context, id uuid.UUID) error {
	m.approvedIDs = append(m.approvedIDs, id)
	return m.approveErr
}

func (m *mockModerationProvider) RejectReview(_ context.Context, id uuid.UUID, reasons []string) error {
	m.rejectedIDs = append(m.rejectedIDs, id)
	m.rejectedReasons = reasons
	return m.rejectErr
}

func (m *mockModerationProvider) BatchApproveReviews(_ context.Context, ids []uuid.UUID) (int, int) {
	m.approvedIDs = append(m.approvedIDs, ids...)
	if m.batchApproveOK > 0 || m.batchApproveFail > 0 {
		return m.batchApproveOK, m.batchApproveFail
	}
	return len(ids), 0
}

func (m *mockModerationProvider) BatchRejectReviews(_ context.Context, ids []uuid.UUID, reasons []string) (int, int) {
	m.rejectedIDs = append(m.rejectedIDs, ids...)
	m.rejectedReasons = reasons
	if m.batchRejectOK > 0 || m.batchRejectFail > 0 {
		return m.batchRejectOK, m.batchRejectFail
	}
	return len(ids), 0
}

func sampleModerationData() *ModerationData {
	return &ModerationData{
		Reviews: []ModerationReview{
			{
				ID:            "11111111-1111-1111-1111-111111111111",
				UserName:      "Иван Иванов",
				BathhouseID:   "aaaa-bbbb",
				BathhouseName: "Русская баня на дровах",
				Rating:        5,
				Text:          "Отличная баня! Рекомендую всем.",
				Status:        "pending",
				Images:        []string{"/img/1.jpg", "/img/2.jpg"},
				CreatedAt:     time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:               "22222222-2222-2222-2222-222222222222",
				UserName:         "Петр Петров",
				BathhouseID:      "cccc-dddd",
				BathhouseName:    "Финская сауна Релакс",
				Rating:           2,
				Text:             "Грязно и холодно.",
				Status:           "rejected",
				RejectionReasons: []string{"Фейковый отзыв"},
				CreatedAt:        time.Date(2026, 3, 6, 18, 0, 0, 0, time.UTC),
			},
		},
		Stats: ModerationStats{
			PendingTotal:  5,
			ApprovedToday: 3,
			RejectedToday: 1,
			PendingWeek:   12,
		},
		Filter: ModerationFilter{},
		Bathhouses: []BathhouseOption{
			{ID: "aaaa-bbbb", Name: "Русская баня на дровах"},
			{ID: "cccc-dddd", Name: "Финская сауна Релакс"},
		},
		TotalCount: 2,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}
}

func TestModerationHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		provider   *mockModerationProvider
		wantStatus int
	}{
		{
			name:       "success with data",
			provider:   &mockModerationProvider{data: sampleModerationData()},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with empty data",
			provider: &mockModerationProvider{
				data: &ModerationData{
					Page:       1,
					PageSize:   20,
					TotalPages: 1,
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "provider error",
			provider:   &mockModerationProvider{err: context.DeadlineExceeded},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandler(tt.provider, testLogger(), "/admin-panel/pages", "/admin-panel")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/moderation", nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if ct != "text/html; charset=utf-8" {
					t.Errorf("content-type = %q, want text/html", ct)
				}
			}
		})
	}
}

func TestModerationHandler_RendersReviews(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"Иван Иванов",
		"Русская баня на дровах",
		"Отличная баня! Рекомендую всем.",
		"Петр Петров",
		"Финская сауна Релакс",
		"Грязно и холодно.",
		"Фейковый отзыв",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestModerationHandler_RendersStats(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Stats: PendingTotal=5, ApprovedToday=3, RejectedToday=1, PendingWeek=12
	if !strings.Contains(body, "warn") {
		t.Error("expected 'warn' class for non-zero pending counts")
	}
}

func TestModerationHandler_RendersImages(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "/img/1.jpg") {
		t.Error("body missing image URL /img/1.jpg")
	}
	if !strings.Contains(body, "/img/2.jpg") {
		t.Error("body missing image URL /img/2.jpg")
	}
}

func TestModerationHandler_HandleApprove(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/approve?id="+id.String(), nil)
	handler.HandleApprove(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp actionResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if len(provider.approvedIDs) != 1 || provider.approvedIDs[0] != id {
		t.Errorf("expected approved ID %s, got %v", id, provider.approvedIDs)
	}
}

func TestModerationHandler_HandleApprove_InvalidID(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/approve?id=invalid", nil)
	handler.HandleApprove(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleApprove_ProviderError(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{approveErr: errors.New("db error")}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/approve?id="+id.String(), nil)
	handler.HandleApprove(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestModerationHandler_HandleReject(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(approveRejectRequest{
		Reasons: []string{"Спам или реклама", "Нецензурная лексика"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if len(provider.rejectedIDs) != 1 || provider.rejectedIDs[0] != id {
		t.Errorf("expected rejected ID %s, got %v", id, provider.rejectedIDs)
	}
	if len(provider.rejectedReasons) != 2 {
		t.Errorf("expected 2 reasons, got %d", len(provider.rejectedReasons))
	}
}

func TestModerationHandler_HandleReject_NoReasons(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), nil)
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (reject without reasons/comment should fail)", rec.Code, http.StatusBadRequest)
	}

	var resp actionResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Success {
		t.Error("expected success=false when no reasons or comment provided")
	}
}

func TestModerationHandler_HandleBatchApprove(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{id1.String(), id2.String()}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchApprove(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp batchResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Successful != 2 {
		t.Errorf("expected 2 successful, got %d", resp.Successful)
	}
	if len(provider.approvedIDs) != 2 {
		t.Errorf("expected 2 approved IDs, got %d", len(provider.approvedIDs))
	}
}

func TestModerationHandler_HandleBatchApprove_InvalidBody(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-approve", bytes.NewReader([]byte("not json")))
	handler.HandleBatchApprove(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchApprove_InvalidIDs(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{"not-a-uuid"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchApprove(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject(t *testing.T) {
	id1 := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{
		IDs:     []string{id1.String()},
		Reasons: []string{"Спам или реклама"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp batchResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Successful != 1 {
		t.Errorf("expected 1 successful, got %d", resp.Successful)
	}
	if len(provider.rejectedReasons) != 1 || provider.rejectedReasons[0] != "Спам или реклама" {
		t.Errorf("expected reason 'Спам или реклама', got %v", provider.rejectedReasons)
	}
}

func TestModerationHandler_WithFilters(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation?status=pending&min_rating=3&bathhouse_id=abc", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestModerationHandler_EmptyState(t *testing.T) {
	provider := &mockModerationProvider{
		data: &ModerationData{
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		},
	}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Нет отзывов для отображения") {
		t.Error("expected empty state message")
	}
}

func TestModerationHandler_RendersBaseLayout(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"sidebar",
		"breadcrumb",
		"<title>Модерация отзывов",
		"Последнее обновление:",
		"/admin-panel/pages/dashboard",
		"/admin-panel/pages/moderation",
		"Назад в GoAdmin",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing layout element %q", c)
		}
	}
}

func TestModerationHandler_RendersLocalizedStatuses(t *testing.T) {
	data := sampleModerationData()
	// First review is "pending", second is "rejected"
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation?status=all", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Localized statuses should appear in badges
	if !strings.Contains(body, "Ожидает") {
		t.Error("body missing localized status 'Ожидает' for pending")
	}
	if !strings.Contains(body, "Отклонено") {
		t.Error("body missing localized status 'Отклонено' for rejected")
	}
}

func TestModerationHandler_PagesPrefixEscapedInJS(t *testing.T) {
	data := sampleModerationData()
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// html/template escapes "/" to "\/" in JS string context, which is valid JS
	if !strings.Contains(body, `var pagesPrefix = "\/admin-panel\/pages"`) {
		t.Error("body missing pagesPrefix JS variable (with html/template \\/ escaping)")
	}
}

func TestModerationHandler_HandleReject_WithCommentOnly(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(approveRejectRequest{
		Comment: "Некачественный контент",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(provider.rejectedIDs) != 1 {
		t.Error("expected review to be rejected with comment only")
	}
	if len(provider.rejectedReasons) != 1 || provider.rejectedReasons[0] != "Некачественный контент" {
		t.Errorf("expected comment as reason, got %v", provider.rejectedReasons)
	}
}

func TestModerationHandler_HandleReject_WithReasonsAndComment(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(approveRejectRequest{
		Reasons: []string{"Спам или реклама"},
		Comment: "Дополнительная причина",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(provider.rejectedReasons) != 2 {
		t.Errorf("expected 2 reasons (1 checkbox + comment), got %d", len(provider.rejectedReasons))
	}
}

func TestModerationHandler_HandleReject_EmptyBody(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(approveRejectRequest{Reasons: []string{}, Comment: ""})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (empty reasons and comment should fail)", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject_NoReasons(t *testing.T) {
	id1 := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{id1.String()}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (batch reject without reasons should fail)", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject_WithComment(t *testing.T) {
	id1 := uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{
		IDs:     []string{id1.String()},
		Comment: "Массовый спам",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(provider.rejectedReasons) != 1 || provider.rejectedReasons[0] != "Массовый спам" {
		t.Errorf("expected comment in reasons, got %v", provider.rejectedReasons)
	}
}

func TestModerationHandler_RendersLightbox(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"lightbox-overlay",
		"lightbox-img",
		"lightbox-close",
		"lightbox-prev",
		"lightbox-next",
		"openLightbox(this)",
		`data-lightbox-group=`,
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing lightbox element %q", c)
		}
	}
}

func TestModerationHandler_RendersConfirmModal(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "confirm-modal") {
		t.Error("body missing confirm modal for batch approve")
	}
	if !strings.Contains(body, "confirmBatchApprove()") {
		t.Error("body missing confirmBatchApprove function call")
	}
}

func TestModerationHandler_RendersRejectCommentField(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "reject-comment") {
		t.Error("body missing reject comment textarea")
	}
	if !strings.Contains(body, "Дополнительный комментарий") {
		t.Error("body missing reject comment label")
	}
	if !strings.Contains(body, "validation-error") {
		t.Error("body missing reject validation error element")
	}
}

func TestModerationHandler_RendersHiddenFilterOption(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `value="hidden"`) {
		t.Error("body missing 'hidden' option in status filter")
	}
	if !strings.Contains(body, "Скрытые") {
		t.Error("body missing 'Скрытые' label for hidden filter option")
	}
}

func TestModerationHandler_RendersTextTruncation(t *testing.T) {
	data := sampleModerationData()
	// Make a review with text longer than 200 characters
	data.Reviews[0].Text = strings.Repeat("А", 250)
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "truncated") {
		t.Error("body missing 'truncated' class for long review text")
	}
	if !strings.Contains(body, "Показать полностью") {
		t.Error("body missing 'Показать полностью' toggle button")
	}
}

func TestModerationHandler_NoTruncationForShortText(t *testing.T) {
	data := sampleModerationData()
	data.Reviews = []ModerationReview{data.Reviews[0]}
	data.Reviews[0].Text = "Короткий текст"
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Short text should not have the truncated class applied
	if strings.Contains(body, `class="review-text truncated"`) {
		t.Error("short text should not have truncated class")
	}
	// Short text should not have a toggle button with the review's ID
	if strings.Contains(body, `toggleReviewText('`+data.Reviews[0].ID+`')`) {
		t.Error("short text should not have truncation toggle button")
	}
}

func TestModerationHandler_RendersInlineAJAX(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Should use showToast instead of alert
	if strings.Contains(body, "alert(") {
		t.Error("body should not contain alert() calls - should use showToast instead")
	}
	// location.reload() is allowed only as a fallback when batch operations have partial failures
	// Should have removeReviewCard function
	if !strings.Contains(body, "removeReviewCard") {
		t.Error("body missing removeReviewCard function for inline DOM updates")
	}
	// Should have updateStat function
	if !strings.Contains(body, "updateStat") {
		t.Error("body missing updateStat function for dynamic counter updates")
	}
	// Should have loading state support
	if !strings.Contains(body, "setButtonLoading") {
		t.Error("body missing setButtonLoading function")
	}
	if !strings.Contains(body, "btn-spinner") {
		t.Error("body missing btn-spinner CSS class for loading states")
	}
	// Stats cards should have IDs for JS updates
	if !strings.Contains(body, `id="stat-pending"`) {
		t.Error("body missing stat-pending ID on stats card")
	}
	if !strings.Contains(body, `id="stat-approved"`) {
		t.Error("body missing stat-approved ID on stats card")
	}
}

func TestModerationHandler_RendersKeyboardShortcuts(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Keyboard event listener should be present
	if !strings.Contains(body, "addEventListener('keydown'") {
		t.Error("body missing keydown event listener for keyboard shortcuts")
	}

	// Shortcut keys should be handled
	shortcuts := []string{
		"e.key === 'a'",
		"e.key === 'r'",
		"e.key === 'Escape'",
		"e.key === 'Enter'",
		"e.key === 'ArrowLeft'",
		"e.key === 'ArrowRight'",
	}
	for _, s := range shortcuts {
		if !strings.Contains(body, s) {
			t.Errorf("body missing keyboard shortcut handler for %q", s)
		}
	}

	// Ctrl+A support
	if !strings.Contains(body, "e.ctrlKey") {
		t.Error("body missing Ctrl key check for Ctrl+A shortcut")
	}
}

func TestModerationHandler_RendersShortcutsHelp(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Shortcuts help button and popup
	if !strings.Contains(body, "shortcuts-help") {
		t.Error("body missing shortcuts-help container")
	}
	if !strings.Contains(body, "shortcuts-popup") {
		t.Error("body missing shortcuts-popup element")
	}
	if !strings.Contains(body, "shortcuts-help-btn") {
		t.Error("body missing shortcuts help button")
	}
	if !strings.Contains(body, "toggleShortcutsPopup") {
		t.Error("body missing toggleShortcutsPopup function")
	}

	// Help content should list all shortcuts
	helpLabels := []string{
		"Горячие клавиши",
		"Одобрить выбранные",
		"Отклонить выбранные",
		"Закрыть окно",
		"Подтвердить отклонение",
		"Выбрать все",
		"Предыдущая страница",
		"Следующая страница",
	}
	for _, label := range helpLabels {
		if !strings.Contains(body, label) {
			t.Errorf("shortcuts help missing label %q", label)
		}
	}

	// Shortcut keys displayed in popup
	keyLabels := []string{
		"shortcut-key",
		">A<",
		">R<",
		"Esc",
		"Enter",
		"Ctrl+A",
	}
	for _, key := range keyLabels {
		if !strings.Contains(body, key) {
			t.Errorf("shortcuts help missing key display %q", key)
		}
	}
}

func TestModerationHandler_RendersFocusManagement(t *testing.T) {
	provider := &mockModerationProvider{data: sampleModerationData()}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Review cards should have tabindex for Tab navigation
	if !strings.Contains(body, `tabindex="0"`) {
		t.Error("body missing tabindex attribute on review cards for Tab navigation")
	}

	// CSS should include focus styles for review cards
	if !strings.Contains(body, ".review-card:focus") {
		t.Error("body missing CSS focus styles for review cards")
	}
}

func TestModerationHandler_RendersPaginationIDs(t *testing.T) {
	data := sampleModerationData()
	data.Page = 2
	data.TotalPages = 5
	data.TotalCount = 100
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation?page=2", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Pagination links should have IDs for keyboard navigation
	if !strings.Contains(body, `id="pagination-prev"`) {
		t.Error("body missing pagination-prev ID for arrow key navigation")
	}
	if !strings.Contains(body, `id="pagination-next"`) {
		t.Error("body missing pagination-next ID for arrow key navigation")
	}
}

func TestModerationHandler_HandleReject_ProviderError(t *testing.T) {
	id := uuid.New()
	provider := &mockModerationProvider{rejectErr: errors.New("db error")}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(approveRejectRequest{
		Reasons: []string{"Спам или реклама"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestModerationHandler_HandleReject_InvalidID(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id=not-a-uuid", nil)
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchApprove_EmptyIDs(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchApprove(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject_InvalidBody(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader([]byte("not json")))
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject_InvalidIDs(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{"not-a-uuid"}, Reasons: []string{"spam"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchReject_EmptyIDs(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	body, _ := json.Marshal(batchRequest{IDs: []string{}, Reasons: []string{"spam"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.HandleBatchReject(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_RendersResponsiveDesign(t *testing.T) {
	data := sampleModerationData()
	provider := &mockModerationProvider{data: data}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Should have media queries for responsive layout
	if !strings.Contains(body, "@media (max-width: 768px)") {
		t.Error("body missing mobile media query (768px)")
	}
	if !strings.Contains(body, "@media (max-width: 1024px)") {
		t.Error("body missing tablet media query (1024px)")
	}

	// Mobile thumbnail size (60x60)
	if !strings.Contains(body, "60px") {
		t.Error("body missing reduced thumbnail size for mobile (60px)")
	}
}
