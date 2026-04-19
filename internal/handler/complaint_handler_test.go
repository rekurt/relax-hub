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
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/handler"
	"github.com/rekurt/relax-hub/internal/service"
)

// --- Mock complaint service ---

type mockComplaintService struct {
	reportFn  func(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error)
	resolveFn func(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID, resolution string) (*domain.Complaint, error)
	dismissFn func(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID) (*domain.Complaint, error)
	listFn    func(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Complaint, error)
}

func (m *mockComplaintService) Report(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error) {
	if m.reportFn != nil {
		return m.reportFn(ctx, reporterID, input)
	}
	return nil, nil
}

func (m *mockComplaintService) Resolve(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID, resolution string) (*domain.Complaint, error) {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, complaintID, adminID, resolution)
	}
	return nil, nil
}

func (m *mockComplaintService) Dismiss(ctx context.Context, complaintID uuid.UUID, adminID uuid.UUID) (*domain.Complaint, error) {
	if m.dismissFn != nil {
		return m.dismissFn(ctx, complaintID, adminID)
	}
	return nil, nil
}

func (m *mockComplaintService) List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockComplaintService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

var _ service.ComplaintService = (*mockComplaintService)(nil)

// --- Tests ---

func TestComplaintHandler_ReportReview(t *testing.T) {
	complaintID := uuid.New()
	userID := uuid.New()
	targetID := uuid.New()

	svc := &mockComplaintService{
		reportFn: func(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error) {
			if reporterID != userID {
				t.Errorf("expected reporter %v, got %v", userID, reporterID)
			}
			if input.TargetType != domain.ComplaintTargetReview {
				t.Errorf("expected target type review, got %v", input.TargetType)
			}
			if input.TargetID != targetID {
				t.Errorf("expected target ID %v, got %v", targetID, input.TargetID)
			}
			return &domain.Complaint{
				ID:         complaintID,
				ReporterID: reporterID,
				TargetType: input.TargetType,
				TargetID:   input.TargetID,
				Reason:     input.Reason,
				Status:     domain.ComplaintStatusPending,
				CreatedAt:  time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Post("/reviews/{id}/report", h.ReportReview)

	body, _ := json.Marshal(map[string]string{
		"reason":      "spam",
		"description": "This is spam",
	})
	req := httptest.NewRequest(http.MethodPost, "/reviews/"+targetID.String()+"/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestComplaintHandler_ReportBathhouse(t *testing.T) {
	targetID := uuid.New()
	userID := uuid.New()

	svc := &mockComplaintService{
		reportFn: func(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error) {
			if input.TargetType != domain.ComplaintTargetBathhouse {
				t.Errorf("expected target type bathhouse, got %v", input.TargetType)
			}
			return &domain.Complaint{
				ID:         uuid.New(),
				ReporterID: reporterID,
				TargetType: input.TargetType,
				TargetID:   input.TargetID,
				Reason:     input.Reason,
				Status:     domain.ComplaintStatusPending,
				CreatedAt:  time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Post("/bathhouses/{id}/report", h.ReportBathhouse)

	body, _ := json.Marshal(map[string]string{"reason": "fraud", "description": "Scam"})
	req := httptest.NewRequest(http.MethodPost, "/bathhouses/"+targetID.String()+"/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestComplaintHandler_ReportUser(t *testing.T) {
	targetID := uuid.New()
	userID := uuid.New()

	svc := &mockComplaintService{
		reportFn: func(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error) {
			if input.TargetType != domain.ComplaintTargetUser {
				t.Errorf("expected target type user, got %v", input.TargetType)
			}
			return &domain.Complaint{
				ID:         uuid.New(),
				ReporterID: reporterID,
				TargetType: input.TargetType,
				TargetID:   input.TargetID,
				Reason:     input.Reason,
				Status:     domain.ComplaintStatusPending,
				CreatedAt:  time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Post("/users/{id}/report", h.ReportUser)

	body, _ := json.Marshal(map[string]string{"reason": "offensive", "description": "Rude"})
	req := httptest.NewRequest(http.MethodPost, "/users/"+targetID.String()+"/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestComplaintHandler_ReportInvalidID(t *testing.T) {
	svc := &mockComplaintService{}
	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Post("/reviews/{id}/report", h.ReportReview)

	body, _ := json.Marshal(map[string]string{"reason": "spam"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/not-a-uuid/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestComplaintHandler_ReportAlreadyReported(t *testing.T) {
	targetID := uuid.New()

	svc := &mockComplaintService{
		reportFn: func(ctx context.Context, reporterID uuid.UUID, input service.CreateComplaintInput) (*domain.Complaint, error) {
			return nil, domain.ErrAlreadyReported
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Post("/reviews/{id}/report", h.ReportReview)

	body, _ := json.Marshal(map[string]string{"reason": "spam"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/"+targetID.String()+"/report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestComplaintHandler_List(t *testing.T) {
	svc := &mockComplaintService{
		listFn: func(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
			if filter.Page != 1 || filter.PageSize != 20 {
				t.Errorf("expected page 1 size 20, got page %d size %d", filter.Page, filter.PageSize)
			}
			return &domain.PaginatedResult[domain.Complaint]{
				Items: []domain.Complaint{
					{
						ID:         uuid.New(),
						ReporterID: uuid.New(),
						TargetType: domain.ComplaintTargetReview,
						TargetID:   uuid.New(),
						Reason:     domain.ComplaintReasonSpam,
						Status:     domain.ComplaintStatusPending,
						CreatedAt:  time.Now(),
					},
				},
				TotalCount: 1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestComplaintHandler_ListWithFilters(t *testing.T) {
	svc := &mockComplaintService{
		listFn: func(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
			if filter.Status == nil || *filter.Status != domain.ComplaintStatusPending {
				t.Error("expected status filter pending")
			}
			if filter.TargetType == nil || *filter.TargetType != domain.ComplaintTargetReview {
				t.Error("expected target_type filter review")
			}
			if filter.Reason == nil || *filter.Reason != domain.ComplaintReasonSpam {
				t.Error("expected reason filter spam")
			}
			return &domain.PaginatedResult[domain.Complaint]{
				Items:      []domain.Complaint{},
				TotalCount: 0,
				Page:       1,
				PageSize:   10,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints?status=pending&target_type=review&reason=spam&page_size=10", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestComplaintHandler_GetByID(t *testing.T) {
	complaintID := uuid.New()

	svc := &mockComplaintService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
			if id != complaintID {
				t.Errorf("expected id %v, got %v", complaintID, id)
			}
			return &domain.Complaint{
				ID:         complaintID,
				ReporterID: uuid.New(),
				TargetType: domain.ComplaintTargetReview,
				TargetID:   uuid.New(),
				Reason:     domain.ComplaintReasonSpam,
				Status:     domain.ComplaintStatusPending,
				CreatedAt:  time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints/"+complaintID.String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestComplaintHandler_GetByIDNotFound(t *testing.T) {
	svc := &mockComplaintService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
			return nil, domain.ErrComplaintNotFound
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints/"+uuid.New().String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestComplaintHandler_Resolve(t *testing.T) {
	complaintID := uuid.New()
	adminID := uuid.New()

	svc := &mockComplaintService{
		resolveFn: func(ctx context.Context, cID uuid.UUID, aID uuid.UUID, resolution string) (*domain.Complaint, error) {
			if cID != complaintID {
				t.Errorf("expected complaint id %v, got %v", complaintID, cID)
			}
			if aID != adminID {
				t.Errorf("expected admin id %v, got %v", adminID, aID)
			}
			if resolution != "Confirmed spam, review hidden" {
				t.Errorf("unexpected resolution: %s", resolution)
			}
			return &domain.Complaint{
				ID:           complaintID,
				ReporterID:   uuid.New(),
				TargetType:   domain.ComplaintTargetReview,
				TargetID:     uuid.New(),
				Reason:       domain.ComplaintReasonSpam,
				Status:       domain.ComplaintStatusResolved,
				ResolvedByID: &aID,
				Resolution:   resolution,
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/complaints/{id}/resolve", h.Resolve)

	body, _ := json.Marshal(map[string]string{"resolution": "Confirmed spam, review hidden"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/complaints/"+complaintID.String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestComplaintHandler_Dismiss(t *testing.T) {
	complaintID := uuid.New()
	adminID := uuid.New()

	svc := &mockComplaintService{
		dismissFn: func(ctx context.Context, cID uuid.UUID, aID uuid.UUID) (*domain.Complaint, error) {
			if cID != complaintID {
				t.Errorf("expected complaint id %v, got %v", complaintID, cID)
			}
			return &domain.Complaint{
				ID:           complaintID,
				ReporterID:   uuid.New(),
				TargetType:   domain.ComplaintTargetReview,
				TargetID:     uuid.New(),
				Reason:       domain.ComplaintReasonSpam,
				Status:       domain.ComplaintStatusDismissed,
				ResolvedByID: &aID,
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/complaints/{id}/dismiss", h.Dismiss)

	req := httptest.NewRequest(http.MethodPatch, "/admin/complaints/"+complaintID.String()+"/dismiss", nil)
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestComplaintHandler_ListInvalidStatusFilter(t *testing.T) {
	svc := &mockComplaintService{}
	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints?status=invalid_status", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid status filter, got %d", rec.Code)
	}
}

func TestComplaintHandler_ListInvalidDateFilter(t *testing.T) {
	svc := &mockComplaintService{}
	h := handler.NewComplaintHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/complaints", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/complaints?from_date=2024-01-01", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid date format, got %d", rec.Code)
	}
}
