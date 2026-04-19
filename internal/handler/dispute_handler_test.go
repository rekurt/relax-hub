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

// --- Mock dispute service ---

type mockDisputeService struct {
	openDisputeFn    func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error)
	getDisputeFn     func(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) (*domain.Dispute, error)
	listUserFn       func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error)
	listAllFn        func(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error)
	submitEvidenceFn func(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID, evidence *domain.DisputeEvidence) error
	listEvidenceFn   func(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) ([]domain.DisputeEvidence, error)
	assignFn         func(ctx context.Context, disputeID uuid.UUID, mediatorID uuid.UUID) error
	resolveFn        func(ctx context.Context, disputeID uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string) error
	appealFn         func(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID) error
	closeFn          func(ctx context.Context, disputeID uuid.UUID) error
}

func (m *mockDisputeService) OpenDispute(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error) {
	if m.openDisputeFn != nil {
		return m.openDisputeFn(ctx, userID, bookingID, reason, description)
	}
	return nil, nil
}

func (m *mockDisputeService) GetDispute(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) (*domain.Dispute, error) {
	if m.getDisputeFn != nil {
		return m.getDisputeFn(ctx, userID, role, disputeID)
	}
	return nil, nil
}

func (m *mockDisputeService) ListUserDisputes(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error) {
	if m.listUserFn != nil {
		return m.listUserFn(ctx, userID, page, pageSize)
	}
	return nil, nil
}

func (m *mockDisputeService) ListAllDisputes(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockDisputeService) SubmitEvidence(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID, evidence *domain.DisputeEvidence) error {
	if m.submitEvidenceFn != nil {
		return m.submitEvidenceFn(ctx, userID, disputeID, evidence)
	}
	return nil
}

func (m *mockDisputeService) ListEvidence(ctx context.Context, userID uuid.UUID, role domain.UserRole, disputeID uuid.UUID) ([]domain.DisputeEvidence, error) {
	if m.listEvidenceFn != nil {
		return m.listEvidenceFn(ctx, userID, role, disputeID)
	}
	return nil, nil
}

func (m *mockDisputeService) AssignDispute(ctx context.Context, disputeID uuid.UUID, mediatorID uuid.UUID) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, disputeID, mediatorID)
	}
	return nil
}

func (m *mockDisputeService) ResolveDispute(ctx context.Context, disputeID uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string) error {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, disputeID, resolution, refundAmount, compensationAmount, mediatorNotes)
	}
	return nil
}

func (m *mockDisputeService) AppealDispute(ctx context.Context, userID uuid.UUID, disputeID uuid.UUID) error {
	if m.appealFn != nil {
		return m.appealFn(ctx, userID, disputeID)
	}
	return nil
}

func (m *mockDisputeService) CloseDispute(ctx context.Context, disputeID uuid.UUID) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, disputeID)
	}
	return nil
}

var _ service.DisputeService = (*mockDisputeService)(nil)

// --- Tests ---

func testDispute() *domain.Dispute {
	now := time.Now()
	evidenceDeadline := now.Add(72 * time.Hour)
	return &domain.Dispute{
		ID:               uuid.New(),
		BookingID:        uuid.New(),
		InitiatorID:      uuid.New(),
		RespondentID:     uuid.New(),
		Reason:           domain.DisputeReasonPoorQuality,
		Description:      "Bad quality",
		Status:           domain.DisputeStatusEvidenceCollection,
		EvidenceDeadline: &evidenceDeadline,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestDisputeHandler_OpenDispute(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	dispute := testDispute()
	dispute.InitiatorID = userID
	dispute.BookingID = bookingID

	svc := &mockDisputeService{
		openDisputeFn: func(_ context.Context, uid uuid.UUID, bid uuid.UUID, reason domain.DisputeReason, desc string) (*domain.Dispute, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if bid != bookingID {
				t.Errorf("expected bookingID %v, got %v", bookingID, bid)
			}
			if reason != domain.DisputeReasonPoorQuality {
				t.Errorf("expected reason poor_quality, got %v", reason)
			}
			return dispute, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/bookings/{id}/dispute", h.OpenDispute)

	body, _ := json.Marshal(map[string]string{
		"reason":      "poor_quality",
		"description": "Bad quality service",
	})
	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/dispute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_OpenDispute_InvalidBookingID(t *testing.T) {
	h := handler.NewDisputeHandler(&mockDisputeService{})
	router := chi.NewRouter()
	router.Post("/bookings/{id}/dispute", h.OpenDispute)

	body, _ := json.Marshal(map[string]string{
		"reason":      "poor_quality",
		"description": "Bad quality",
	})
	req := httptest.NewRequest(http.MethodPost, "/bookings/not-a-uuid/dispute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDisputeHandler_OpenDispute_AlreadyExists(t *testing.T) {
	bookingID := uuid.New()
	svc := &mockDisputeService{
		openDisputeFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ domain.DisputeReason, _ string) (*domain.Dispute, error) {
			return nil, domain.ErrDisputeAlreadyExists
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/bookings/{id}/dispute", h.OpenDispute)

	body, _ := json.Marshal(map[string]string{
		"reason":      "poor_quality",
		"description": "Duplicate",
	})
	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/dispute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestDisputeHandler_GetDispute(t *testing.T) {
	userID := uuid.New()
	dispute := testDispute()
	dispute.InitiatorID = userID

	svc := &mockDisputeService{
		getDisputeFn: func(_ context.Context, uid uuid.UUID, role domain.UserRole, did uuid.UUID) (*domain.Dispute, error) {
			if did != dispute.ID {
				t.Errorf("expected disputeID %v, got %v", dispute.ID, did)
			}
			return dispute, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/disputes/{id}", h.GetDispute)

	req := httptest.NewRequest(http.MethodGet, "/my/disputes/"+dispute.ID.String(), nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_GetDispute_NotFound(t *testing.T) {
	svc := &mockDisputeService{
		getDisputeFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Dispute, error) {
			return nil, domain.ErrDisputeNotFound
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/disputes/{id}", h.GetDispute)

	req := httptest.NewRequest(http.MethodGet, "/my/disputes/"+uuid.New().String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDisputeHandler_ListUserDisputes(t *testing.T) {
	userID := uuid.New()

	svc := &mockDisputeService{
		listUserFn: func(_ context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return &domain.PaginatedResult[domain.Dispute]{
				Items:      []domain.Dispute{*testDispute()},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/disputes", h.ListUserDisputes)

	req := httptest.NewRequest(http.MethodGet, "/my/disputes?page=1&page_size=20", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDisputeHandler_SubmitEvidence(t *testing.T) {
	userID := uuid.New()
	disputeID := uuid.New()

	svc := &mockDisputeService{
		submitEvidenceFn: func(_ context.Context, uid uuid.UUID, did uuid.UUID, ev *domain.DisputeEvidence) error {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if did != disputeID {
				t.Errorf("expected disputeID %v, got %v", disputeID, did)
			}
			if ev.Type != domain.DisputeEvidencePhoto {
				t.Errorf("expected type photo, got %v", ev.Type)
			}
			return nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/disputes/{id}/evidence", h.SubmitEvidence)

	body, _ := json.Marshal(map[string]string{
		"type":        "photo",
		"url":         "https://example.com/photo.jpg",
		"description": "Photo evidence",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/disputes/"+disputeID.String()+"/evidence", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_SubmitEvidence_WindowExpired(t *testing.T) {
	svc := &mockDisputeService{
		submitEvidenceFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ *domain.DisputeEvidence) error {
			return domain.ErrDisputeEvidenceWindowExpired
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/disputes/{id}/evidence", h.SubmitEvidence)

	body, _ := json.Marshal(map[string]string{
		"type": "photo",
		"url":  "https://example.com/photo.jpg",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/disputes/"+uuid.New().String()+"/evidence", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDisputeHandler_ListEvidence(t *testing.T) {
	userID := uuid.New()
	disputeID := uuid.New()

	svc := &mockDisputeService{
		listEvidenceFn: func(_ context.Context, uid uuid.UUID, role domain.UserRole, did uuid.UUID) ([]domain.DisputeEvidence, error) {
			return []domain.DisputeEvidence{
				{
					ID:        uuid.New(),
					DisputeID: did,
					UserID:    uid,
					Type:      domain.DisputeEvidencePhoto,
					URL:       "https://example.com/photo.jpg",
					CreatedAt: time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/disputes/{id}/evidence", h.ListEvidence)

	req := httptest.NewRequest(http.MethodGet, "/my/disputes/"+disputeID.String()+"/evidence", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDisputeHandler_AppealDispute(t *testing.T) {
	userID := uuid.New()
	disputeID := uuid.New()

	svc := &mockDisputeService{
		appealFn: func(_ context.Context, uid uuid.UUID, did uuid.UUID) error {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			if did != disputeID {
				t.Errorf("expected disputeID %v, got %v", disputeID, did)
			}
			return nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/disputes/{id}/appeal", h.AppealDispute)

	req := httptest.NewRequest(http.MethodPost, "/my/disputes/"+disputeID.String()+"/appeal", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_AppealDispute_NotResolved(t *testing.T) {
	svc := &mockDisputeService{
		appealFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domain.ErrDisputeNotResolved
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/disputes/{id}/appeal", h.AppealDispute)

	req := httptest.NewRequest(http.MethodPost, "/my/disputes/"+uuid.New().String()+"/appeal", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDisputeHandler_AdminGetDispute(t *testing.T) {
	adminID := uuid.New()
	dispute := testDispute()

	svc := &mockDisputeService{
		getDisputeFn: func(_ context.Context, _ uuid.UUID, role domain.UserRole, did uuid.UUID) (*domain.Dispute, error) {
			if role != domain.RoleAdmin {
				t.Errorf("expected admin role, got %v", role)
			}
			if did != dispute.ID {
				t.Errorf("expected disputeID %v, got %v", dispute.ID, did)
			}
			return dispute, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/disputes/{id}", h.AdminGetDispute)

	req := httptest.NewRequest(http.MethodGet, "/admin/disputes/"+dispute.ID.String(), nil)
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_AdminListDisputes(t *testing.T) {
	svc := &mockDisputeService{
		listAllFn: func(_ context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error) {
			if filter.Status != nil && *filter.Status != domain.DisputeStatusOpen {
				t.Errorf("expected status filter open, got %v", *filter.Status)
			}
			return &domain.PaginatedResult[domain.Dispute]{
				Items:      []domain.Dispute{},
				Page:       1,
				PageSize:   20,
				TotalCount: 0,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/disputes", h.AdminListDisputes)

	req := httptest.NewRequest(http.MethodGet, "/admin/disputes?status=open", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDisputeHandler_AdminAssignDispute(t *testing.T) {
	disputeID := uuid.New()
	mediatorID := uuid.New()

	svc := &mockDisputeService{
		assignFn: func(_ context.Context, did uuid.UUID, mid uuid.UUID) error {
			if did != disputeID {
				t.Errorf("expected disputeID %v, got %v", disputeID, did)
			}
			if mid != mediatorID {
				t.Errorf("expected mediatorID %v, got %v", mediatorID, mid)
			}
			return nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/disputes/{id}/assign", h.AdminAssignDispute)

	body, _ := json.Marshal(map[string]string{
		"mediator_id": mediatorID.String(),
	})
	req := httptest.NewRequest(http.MethodPatch, "/admin/disputes/"+disputeID.String()+"/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_AdminResolveDispute(t *testing.T) {
	disputeID := uuid.New()

	svc := &mockDisputeService{
		resolveFn: func(_ context.Context, did uuid.UUID, resolution domain.DisputeResolution, refundAmt, compAmt int64, notes string) error {
			if did != disputeID {
				t.Errorf("expected disputeID %v, got %v", disputeID, did)
			}
			if resolution != domain.DisputeResolutionFullRefund {
				t.Errorf("expected resolution full_refund, got %v", resolution)
			}
			if refundAmt != 500000 {
				t.Errorf("expected refundAmount 500000, got %d", refundAmt)
			}
			return nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/disputes/{id}/resolve", h.AdminResolveDispute)

	body, _ := json.Marshal(map[string]interface{}{
		"resolution":          "full_refund",
		"refund_amount":       500000,
		"compensation_amount": 0,
		"mediator_notes":      "Client was right",
	})
	req := httptest.NewRequest(http.MethodPatch, "/admin/disputes/"+disputeID.String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_AdminCloseDispute(t *testing.T) {
	disputeID := uuid.New()

	svc := &mockDisputeService{
		closeFn: func(_ context.Context, did uuid.UUID) error {
			if did != disputeID {
				t.Errorf("expected disputeID %v, got %v", disputeID, did)
			}
			return nil
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/disputes/{id}/close", h.AdminCloseDispute)

	req := httptest.NewRequest(http.MethodPatch, "/admin/disputes/"+disputeID.String()+"/close", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisputeHandler_AdminCloseDispute_AlreadyClosed(t *testing.T) {
	svc := &mockDisputeService{
		closeFn: func(_ context.Context, _ uuid.UUID) error {
			return domain.ErrDisputeAlreadyClosed
		},
	}

	h := handler.NewDisputeHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/disputes/{id}/close", h.AdminCloseDispute)

	req := httptest.NewRequest(http.MethodPatch, "/admin/disputes/"+uuid.New().String()+"/close", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}
