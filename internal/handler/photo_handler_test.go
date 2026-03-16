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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/service"
)

// --- Mock photo verification service ---

type mockPhotoVerificationService struct {
	uploadPhotoFn              func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, input service.UploadPhotoInput) (*domain.BathhousePhoto, error)
	deletePhotoFn              func(ctx context.Context, photoID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error
	reorderPhotosFn            func(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole, photoIDs []uuid.UUID) error
	verifyPhotoFn              func(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID) (*domain.BathhousePhoto, error)
	rejectPhotoFn              func(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID, reason string) (*domain.BathhousePhoto, error)
	getPendingFn               func(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error)
	listByBathhouseFn          func(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	listByBathhouseForOwnerFn  func(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) ([]domain.BathhousePhoto, error)
	listVerifiedByBathhouseFn  func(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
}

func (m *mockPhotoVerificationService) UploadPhoto(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, input service.UploadPhotoInput) (*domain.BathhousePhoto, error) {
	if m.uploadPhotoFn != nil {
		return m.uploadPhotoFn(ctx, userID, userRole, input)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) DeletePhoto(ctx context.Context, photoID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error {
	if m.deletePhotoFn != nil {
		return m.deletePhotoFn(ctx, photoID, userID, userRole)
	}
	return nil
}

func (m *mockPhotoVerificationService) ReorderPhotos(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole, photoIDs []uuid.UUID) error {
	if m.reorderPhotosFn != nil {
		return m.reorderPhotosFn(ctx, bathhouseID, userID, userRole, photoIDs)
	}
	return nil
}

func (m *mockPhotoVerificationService) VerifyPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID) (*domain.BathhousePhoto, error) {
	if m.verifyPhotoFn != nil {
		return m.verifyPhotoFn(ctx, photoID, adminID)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) RejectPhoto(ctx context.Context, photoID uuid.UUID, adminID uuid.UUID, reason string) (*domain.BathhousePhoto, error) {
	if m.rejectPhotoFn != nil {
		return m.rejectPhotoFn(ctx, photoID, adminID, reason)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) GetPendingPhotos(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
	if m.getPendingFn != nil {
		return m.getPendingFn(ctx, page, pageSize)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, bathhouseID)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) ListByBathhouseForOwner(ctx context.Context, bathhouseID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) ([]domain.BathhousePhoto, error) {
	if m.listByBathhouseForOwnerFn != nil {
		return m.listByBathhouseForOwnerFn(ctx, bathhouseID, userID, userRole)
	}
	return nil, nil
}

func (m *mockPhotoVerificationService) ListVerifiedByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	if m.listVerifiedByBathhouseFn != nil {
		return m.listVerifiedByBathhouseFn(ctx, bathhouseID)
	}
	return nil, nil
}

var _ service.PhotoVerificationService = (*mockPhotoVerificationService)(nil)

// --- Tests ---

func TestPhotoHandler_Upload(t *testing.T) {
	photoID := uuid.New()
	bathhouseID := uuid.New()
	userID := uuid.New()

	svc := &mockPhotoVerificationService{
		uploadPhotoFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, input service.UploadPhotoInput) (*domain.BathhousePhoto, error) {
			if uid != userID {
				t.Errorf("expected user %v, got %v", userID, uid)
			}
			if input.BathhouseID != bathhouseID {
				t.Errorf("expected bathhouse %v, got %v", bathhouseID, input.BathhouseID)
			}
			if input.URL != "https://example.com/photo.jpg" {
				t.Errorf("expected URL https://example.com/photo.jpg, got %v", input.URL)
			}
			return &domain.BathhousePhoto{
				ID:          photoID,
				BathhouseID: bathhouseID,
				URL:         input.URL,
				Position:    0,
				Status:      domain.PhotoStatusPending,
				UploadedAt:  time.Now(),
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/bathhouses/{id}/photos", h.Upload)

	body, _ := json.Marshal(map[string]string{
		"url":           "https://example.com/photo.jpg",
		"thumbnail_url": "https://example.com/photo_thumb.jpg",
	})
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+bathhouseID.String()+"/photos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_Upload_InvalidBathhouseID(t *testing.T) {
	h := handler.NewPhotoHandler(&mockPhotoVerificationService{})
	router := chi.NewRouter()
	router.Post("/my/bathhouses/{id}/photos", h.Upload)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/photo.jpg"})
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/invalid-id/photos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPhotoHandler_Upload_Forbidden(t *testing.T) {
	svc := &mockPhotoVerificationService{
		uploadPhotoFn: func(ctx context.Context, uid uuid.UUID, role domain.UserRole, input service.UploadPhotoInput) (*domain.BathhousePhoto, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Post("/my/bathhouses/{id}/photos", h.Upload)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/photo.jpg"})
	req := httptest.NewRequest(http.MethodPost, "/my/bathhouses/"+uuid.New().String()+"/photos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestPhotoHandler_Delete(t *testing.T) {
	photoID := uuid.New()
	userID := uuid.New()

	svc := &mockPhotoVerificationService{
		deletePhotoFn: func(ctx context.Context, pid uuid.UUID, uid uuid.UUID, role domain.UserRole) error {
			if pid != photoID {
				t.Errorf("expected photo %v, got %v", photoID, pid)
			}
			if uid != userID {
				t.Errorf("expected user %v, got %v", userID, uid)
			}
			return nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Delete("/photos/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/photos/"+photoID.String(), nil)
	req = req.WithContext(createTestContext(userID, domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_Delete_NotFound(t *testing.T) {
	svc := &mockPhotoVerificationService{
		deletePhotoFn: func(ctx context.Context, pid uuid.UUID, uid uuid.UUID, role domain.UserRole) error {
			return domain.ErrNotFound
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Delete("/photos/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/photos/"+uuid.New().String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestPhotoHandler_Reorder(t *testing.T) {
	bathhouseID := uuid.New()
	userID := uuid.New()
	photoID1 := uuid.New()
	photoID2 := uuid.New()

	svc := &mockPhotoVerificationService{
		reorderPhotosFn: func(ctx context.Context, bhID uuid.UUID, uid uuid.UUID, role domain.UserRole, photoIDs []uuid.UUID) error {
			if bhID != bathhouseID {
				t.Errorf("expected bathhouse %v, got %v", bathhouseID, bhID)
			}
			if len(photoIDs) != 2 {
				t.Errorf("expected 2 photo IDs, got %d", len(photoIDs))
			}
			return nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Put("/my/bathhouses/{id}/photos/reorder", h.Reorder)

	body, _ := json.Marshal(map[string][]string{
		"photo_ids": {photoID1.String(), photoID2.String()},
	})
	req := httptest.NewRequest(http.MethodPut, "/my/bathhouses/"+bathhouseID.String()+"/photos/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(userID, domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_Reorder_InvalidPhotoID(t *testing.T) {
	h := handler.NewPhotoHandler(&mockPhotoVerificationService{})
	router := chi.NewRouter()
	router.Put("/my/bathhouses/{id}/photos/reorder", h.Reorder)

	body, _ := json.Marshal(map[string][]string{
		"photo_ids": {"not-a-uuid"},
	})
	req := httptest.NewRequest(http.MethodPut, "/my/bathhouses/"+uuid.New().String()+"/photos/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPhotoHandler_GetPending(t *testing.T) {
	adminID := uuid.New()
	photoID := uuid.New()

	svc := &mockPhotoVerificationService{
		getPendingFn: func(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
			if page != 1 {
				t.Errorf("expected page 1, got %d", page)
			}
			if pageSize != 20 {
				t.Errorf("expected page_size 20, got %d", pageSize)
			}
			return &domain.PaginatedResult[domain.BathhousePhoto]{
				Items: []domain.BathhousePhoto{
					{
						ID:          photoID,
						BathhouseID: uuid.New(),
						URL:         "https://example.com/photo.jpg",
						Status:      domain.PhotoStatusPending,
						UploadedAt:  time.Now(),
					},
				},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Get("/admin/photos/pending", h.GetPending)

	req := httptest.NewRequest(http.MethodGet, "/admin/photos/pending", nil)
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["meta"] == nil {
		t.Error("expected meta in response")
	}
}

func TestPhotoHandler_Verify(t *testing.T) {
	photoID := uuid.New()
	adminID := uuid.New()

	svc := &mockPhotoVerificationService{
		verifyPhotoFn: func(ctx context.Context, pid uuid.UUID, aid uuid.UUID) (*domain.BathhousePhoto, error) {
			if pid != photoID {
				t.Errorf("expected photo %v, got %v", photoID, pid)
			}
			if aid != adminID {
				t.Errorf("expected admin %v, got %v", adminID, aid)
			}
			now := time.Now()
			return &domain.BathhousePhoto{
				ID:           photoID,
				BathhouseID:  uuid.New(),
				URL:          "https://example.com/photo.jpg",
				Status:       domain.PhotoStatusVerified,
				VerifiedByID: &aid,
				VerifiedAt:   &now,
				UploadedAt:   time.Now(),
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/photos/{id}/verify", h.Verify)

	req := httptest.NewRequest(http.MethodPatch, "/admin/photos/"+photoID.String()+"/verify", nil)
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_Verify_InvalidInput(t *testing.T) {
	svc := &mockPhotoVerificationService{
		verifyPhotoFn: func(ctx context.Context, pid uuid.UUID, aid uuid.UUID) (*domain.BathhousePhoto, error) {
			return nil, domain.ErrInvalidInput
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/photos/{id}/verify", h.Verify)

	req := httptest.NewRequest(http.MethodPatch, "/admin/photos/"+uuid.New().String()+"/verify", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPhotoHandler_Reject(t *testing.T) {
	photoID := uuid.New()
	adminID := uuid.New()

	svc := &mockPhotoVerificationService{
		rejectPhotoFn: func(ctx context.Context, pid uuid.UUID, aid uuid.UUID, reason string) (*domain.BathhousePhoto, error) {
			if pid != photoID {
				t.Errorf("expected photo %v, got %v", photoID, pid)
			}
			if reason != "blurry image" {
				t.Errorf("expected reason 'blurry image', got %v", reason)
			}
			return &domain.BathhousePhoto{
				ID:              photoID,
				BathhouseID:     uuid.New(),
				URL:             "https://example.com/photo.jpg",
				Status:          domain.PhotoStatusRejected,
				RejectionReason: reason,
				UploadedAt:      time.Now(),
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Patch("/admin/photos/{id}/reject", h.Reject)

	body, _ := json.Marshal(map[string]string{"reason": "blurry image"})
	req := httptest.NewRequest(http.MethodPatch, "/admin/photos/"+photoID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(adminID, domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_ListByBathhouse(t *testing.T) {
	bathhouseID := uuid.New()

	svc := &mockPhotoVerificationService{
		listVerifiedByBathhouseFn: func(ctx context.Context, bhID uuid.UUID) ([]domain.BathhousePhoto, error) {
			if bhID != bathhouseID {
				t.Errorf("expected bathhouse %v, got %v", bathhouseID, bhID)
			}
			return []domain.BathhousePhoto{
				{
					ID:          uuid.New(),
					BathhouseID: bathhouseID,
					URL:         "https://example.com/photo1.jpg",
					Position:    0,
					Status:      domain.PhotoStatusVerified,
					UploadedAt:  time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/photos", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bathhouseID.String()+"/photos", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPhotoHandler_ListByBathhouseOwner(t *testing.T) {
	bathhouseID := uuid.New()
	userID := uuid.New()

	svc := &mockPhotoVerificationService{
		listByBathhouseForOwnerFn: func(ctx context.Context, bhID uuid.UUID, uid uuid.UUID, role domain.UserRole) ([]domain.BathhousePhoto, error) {
			if bhID != bathhouseID {
				t.Errorf("expected bathhouse %v, got %v", bathhouseID, bhID)
			}
			if uid != userID {
				t.Errorf("expected user %v, got %v", userID, uid)
			}
			return []domain.BathhousePhoto{
				{
					ID:          uuid.New(),
					BathhouseID: bathhouseID,
					URL:         "https://example.com/photo1.jpg",
					Position:    0,
					Status:      domain.PhotoStatusVerified,
					UploadedAt:  time.Now(),
				},
				{
					ID:          uuid.New(),
					BathhouseID: bathhouseID,
					URL:         "https://example.com/photo2.jpg",
					Position:    1,
					Status:      domain.PhotoStatusPending,
					UploadedAt:  time.Now(),
				},
				{
					ID:              uuid.New(),
					BathhouseID:     bathhouseID,
					URL:             "https://example.com/photo3.jpg",
					Position:        2,
					Status:          domain.PhotoStatusRejected,
					RejectionReason: "blurry",
					UploadedAt:      time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/bathhouses/{id}/photos", h.ListByBathhouseOwner)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+bathhouseID.String()+"/photos", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 3 {
		t.Errorf("expected 3 photos, got %d", len(data))
	}
}

func TestPhotoHandler_ListByBathhouseOwner_InvalidID(t *testing.T) {
	h := handler.NewPhotoHandler(&mockPhotoVerificationService{})
	router := chi.NewRouter()
	router.Get("/my/bathhouses/{id}/photos", h.ListByBathhouseOwner)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/invalid-id/photos", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPhotoHandler_ListByBathhouseOwner_Forbidden(t *testing.T) {
	svc := &mockPhotoVerificationService{
		listByBathhouseForOwnerFn: func(ctx context.Context, bhID uuid.UUID, uid uuid.UUID, role domain.UserRole) ([]domain.BathhousePhoto, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := handler.NewPhotoHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/bathhouses/{id}/photos", h.ListByBathhouseOwner)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/"+uuid.New().String()+"/photos", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}
