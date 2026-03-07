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
			handler := NewModerationHandler(tt.provider, testLogger(), "/admin-panel/pages")
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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/reject?id="+id.String(), nil)
	handler.HandleReject(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(provider.rejectedIDs) != 1 {
		t.Error("expected review to be rejected even without reasons")
	}
}

func TestModerationHandler_HandleBatchApprove(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/moderation/api/batch-approve", bytes.NewReader([]byte("not json")))
	handler.HandleBatchApprove(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestModerationHandler_HandleBatchApprove_InvalidIDs(t *testing.T) {
	provider := &mockModerationProvider{}
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

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
	handler := NewModerationHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/moderation", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Нет отзывов для отображения") {
		t.Error("expected empty state message")
	}
}
