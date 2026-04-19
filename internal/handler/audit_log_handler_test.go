package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/handler"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

// --- mock AuditLogService ---

type mockAuditLogService struct {
	logChangeFn      func(ctx context.Context, entityType string, entityID, userID uuid.UUID, action domain.AuditAction, changedFields map[string]interface{}) error
	getHistoryFn     func(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	listAllFn        func(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
	listAdminFn      func(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
	isSubstantialFn  func(old, new *domain.Bathhouse) bool
}

func (m *mockAuditLogService) LogChange(ctx context.Context, entityType string, entityID, userID uuid.UUID, action domain.AuditAction, changedFields map[string]interface{}) error {
	if m.logChangeFn != nil {
		return m.logChangeFn(ctx, entityType, entityID, userID, action, changedFields)
	}
	return nil
}

func (m *mockAuditLogService) GetHistory(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	if m.getHistoryFn != nil {
		return m.getHistoryFn(ctx, entityType, entityID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.AuditLog]{}, nil
}

func (m *mockAuditLogService) ListAll(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.AuditLog]{}, nil
}

func (m *mockAuditLogService) ListAdminActions(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	if m.listAdminFn != nil {
		return m.listAdminFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.AuditLog]{}, nil
}

func (m *mockAuditLogService) IsSubstantialChange(old, new *domain.Bathhouse) bool {
	if m.isSubstantialFn != nil {
		return m.isSubstantialFn(old, new)
	}
	return false
}

var _ service.AuditLogService = (*mockAuditLogService)(nil)

// --- mock AccessChecker (via real AccessChecker with mock repos) ---

func newTestAuditLogHandler(auditSvc service.AuditLogService) *handler.AuditLogHandler {
	// AccessChecker with empty repos - for tests that don't need it, access will fail (ErrForbidden).
	// Tests that need access will seed the bathhouse repo.
	repRepo := mock.NewRepresentativeRepo()
	bhRepo := mock.NewBathhouseRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	return handler.NewAuditLogHandler(auditSvc, nil, ac)
}

func newTestAuditLogHandlerWithAccess(auditSvc service.AuditLogService, ownerID, bathhouseID uuid.UUID) *handler.AuditLogHandler {
	repRepo := mock.NewRepresentativeRepo()
	bhRepo := mock.NewBathhouseRepo()
	// Seed a bathhouse owned by ownerID so AccessChecker allows access.
	bh := &domain.Bathhouse{
		ID:      bathhouseID,
		OwnerID: ownerID,
		Name:    "Test Bathhouse",
		Status:  domain.BathhouseStatusActive,
	}
	bhRepo.Create(context.Background(), bh)
	ac := service.NewAccessChecker(repRepo, bhRepo)
	return handler.NewAuditLogHandler(auditSvc, nil, ac)
}

// --- helpers ---

func makeAuditLogs(n int, entityType string, action domain.AuditAction) []domain.AuditLog {
	logs := make([]domain.AuditLog, n)
	for i := 0; i < n; i++ {
		logs[i] = domain.AuditLog{
			ID:         uuid.New(),
			EntityType: entityType,
			EntityID:   uuid.New(),
			UserID:     uuid.New(),
			Action:     action,
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}
	return logs
}

func paginatedAuditResult(items []domain.AuditLog, page, pageSize int) *domain.PaginatedResult[domain.AuditLog] {
	total := int64(len(items))
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return &domain.PaginatedResult[domain.AuditLog]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}
}

// --- TestAuditLogHandler_ListAdmin ---

func TestAuditLogHandler_ListAdmin_DefaultPagination(t *testing.T) {
	logs := makeAuditLogs(3, "bathhouse", domain.AuditActionUpdate)
	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.Page != 1 {
				t.Errorf("expected page=1, got %d", filter.Page)
			}
			if filter.PageSize != 20 {
				t.Errorf("expected page_size=20, got %d", filter.PageSize)
			}
			return paginatedAuditResult(logs, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

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
	if resp.Meta.TotalCount != 3 {
		t.Errorf("expected total_count=3, got %d", resp.Meta.TotalCount)
	}
}

func TestAuditLogHandler_ListAdmin_WithEntityTypeFilter(t *testing.T) {
	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.EntityType == nil || *filter.EntityType != "bathhouse" {
				t.Errorf("expected entity_type=bathhouse, got %v", filter.EntityType)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?entity_type=bathhouse", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_WithEntityIDFilter(t *testing.T) {
	entityID := uuid.New()
	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.EntityID == nil || *filter.EntityID != entityID {
				t.Errorf("expected entity_id=%s, got %v", entityID, filter.EntityID)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?entity_id="+entityID.String(), nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_InvalidEntityID(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?entity_id=not-a-uuid", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_WithUserIDFilter(t *testing.T) {
	userID := uuid.New()
	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.UserID == nil || *filter.UserID != userID {
				t.Errorf("expected user_id=%s, got %v", userID, filter.UserID)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?user_id="+userID.String(), nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_InvalidUserID(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?user_id=invalid", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_WithActionFilter(t *testing.T) {
	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.Action == nil || *filter.Action != domain.AuditActionCreate {
				t.Errorf("expected action=create, got %v", filter.Action)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?action=create", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_InvalidAction(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?action=bogus", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_WithDateFilters(t *testing.T) {
	fromDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	svc := &mockAuditLogService{
		listAllFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.FromDate == nil || !filter.FromDate.Equal(fromDate) {
				t.Errorf("expected from_date=%v, got %v", fromDate, filter.FromDate)
			}
			if filter.ToDate == nil || !filter.ToDate.Equal(toDate) {
				t.Errorf("expected to_date=%v, got %v", toDate, filter.ToDate)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?from_date="+fromDate.Format(time.RFC3339)+"&to_date="+toDate.Format(time.RFC3339), nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_InvalidFromDate(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?from_date=not-a-date", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdmin_InvalidToDate(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log?to_date=not-a-date", nil)
	w := httptest.NewRecorder()
	h.ListAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- TestAuditLogHandler_ListAdminActions ---

func TestAuditLogHandler_ListAdminActions_DefaultPagination(t *testing.T) {
	logs := makeAuditLogs(2, "admin_action", domain.AuditActionCreate)
	svc := &mockAuditLogService{
		listAdminFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.Page != 1 {
				t.Errorf("expected page=1, got %d", filter.Page)
			}
			if filter.PageSize != 20 {
				t.Errorf("expected page_size=20, got %d", filter.PageSize)
			}
			return paginatedAuditResult(logs, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions", nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

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
	if resp.Meta.TotalCount != 2 {
		t.Errorf("expected total_count=2, got %d", resp.Meta.TotalCount)
	}
}

func TestAuditLogHandler_ListAdminActions_WithAdminIDFilter(t *testing.T) {
	adminID := uuid.New()
	svc := &mockAuditLogService{
		listAdminFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.UserID == nil || *filter.UserID != adminID {
				t.Errorf("expected user_id (admin_id)=%s, got %v", adminID, filter.UserID)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions?admin_id="+adminID.String(), nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdminActions_InvalidAdminID(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions?admin_id=invalid", nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdminActions_WithActionFilter(t *testing.T) {
	svc := &mockAuditLogService{
		listAdminFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.Action == nil || *filter.Action != domain.AuditActionDelete {
				t.Errorf("expected action=delete, got %v", filter.Action)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions?action=delete", nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdminActions_InvalidAction(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions?action=bogus", nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListAdminActions_WithDateFilters(t *testing.T) {
	fromDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	svc := &mockAuditLogService{
		listAdminFn: func(_ context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
			if filter.FromDate == nil || !filter.FromDate.Equal(fromDate) {
				t.Errorf("expected from_date=%v, got %v", fromDate, filter.FromDate)
			}
			if filter.ToDate == nil || !filter.ToDate.Equal(toDate) {
				t.Errorf("expected to_date=%v, got %v", toDate, filter.ToDate)
			}
			return paginatedAuditResult(nil, filter.Page, filter.PageSize), nil
		},
	}
	h := newTestAuditLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit-log/actions?from_date="+fromDate.Format(time.RFC3339)+"&to_date="+toDate.Format(time.RFC3339), nil)
	w := httptest.NewRecorder()
	h.ListAdminActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- TestAuditLogHandler_ListByBathhouse (GetEntityHistory) ---

func TestAuditLogHandler_ListByBathhouse_Success(t *testing.T) {
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	logs := makeAuditLogs(2, "bathhouse", domain.AuditActionUpdate)

	svc := &mockAuditLogService{
		getHistoryFn: func(_ context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
			if entityType != "bathhouse" {
				t.Errorf("expected entity_type=bathhouse, got %s", entityType)
			}
			if entityID != bathhouseID {
				t.Errorf("expected entity_id=%s, got %s", bathhouseID, entityID)
			}
			return paginatedAuditResult(logs, page, pageSize), nil
		},
	}

	h := newTestAuditLogHandlerWithAccess(svc, ownerID, bathhouseID)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/history", nil)
	ctx := middleware.SetUserID(req.Context(), ownerID)
	ctx = middleware.SetUserRole(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Get("/my/bathhouses/{id}/history", h.ListByBathhouse)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

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
	if resp.Meta.TotalCount != 2 {
		t.Errorf("expected total_count=2, got %d", resp.Meta.TotalCount)
	}
}

func TestAuditLogHandler_ListByBathhouse_InvalidID(t *testing.T) {
	h := newTestAuditLogHandler(&mockAuditLogService{})

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/not-a-uuid/history", nil)
	ctx := middleware.SetUserID(req.Context(), uuid.New())
	ctx = middleware.SetUserRole(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Get("/my/bathhouses/{id}/history", h.ListByBathhouse)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuditLogHandler_ListByBathhouse_NotFound(t *testing.T) {
	// Use default handler with no seeded bathhouse - access check returns ErrNotFound
	svc := &mockAuditLogService{}
	h := newTestAuditLogHandler(svc)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/history", nil)
	ctx := middleware.SetUserID(req.Context(), uuid.New())
	ctx = middleware.SetUserRole(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	r := chi.NewRouter()
	r.Get("/my/bathhouses/{id}/history", h.ListByBathhouse)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// AccessChecker returns ErrNotFound for unknown bathhouse, which maps to 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
