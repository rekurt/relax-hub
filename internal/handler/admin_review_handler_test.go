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
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

type mockAdminNotificationService struct {
	sendFn func(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error
}

func (m *mockAdminNotificationService) Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error {
	if m.sendFn != nil {
		return m.sendFn(ctx, userID, notifType, title, body, data)
	}
	return nil
}

func (m *mockAdminNotificationService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	return &domain.PaginatedResult[domain.Notification]{}, nil
}

func (m *mockAdminNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	return nil
}

func (m *mockAdminNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockAdminNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockAdminNotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	return &domain.NotificationPreferences{}, nil
}

func (m *mockAdminNotificationService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error {
	return nil
}

func TestAdminHandler_ListReviews(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	bathhouseID := uuid.New()
	review1 := &domain.Review{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bathhouseID,
		BookingID:   uuid.New(),
		Rating:      5,
		Text:        "Great place!",
		Status:      domain.ReviewStatusApproved,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	review2 := &domain.Review{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bathhouseID,
		BookingID:   uuid.New(),
		Rating:      2,
		Text:        "Bad experience",
		Status:      domain.ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	reviewRepo := mock.NewReviewRepo()
	if err := reviewRepo.Create(context.Background(), review1); err != nil {
		t.Fatalf("failed to create review1: %v", err)
	}
	if err := reviewRepo.Create(context.Background(), review2); err != nil {
		t.Fatalf("failed to create review2: %v", err)
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Get("/admin/reviews", adminH.ListReviews)

	tests := []struct {
		name        string
		query       string
		expectedLen int
	}{
		{"all reviews", "", 2},
		{"pending reviews", "?status=pending", 1},
		{"approved reviews", "?status=approved", 1},
		{"high rating", "?min_rating=4", 1},
		{"low rating", "?max_rating=3", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/reviews"+tt.query, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			var resp handler.APIResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if !resp.Success {
				t.Errorf("expected success=true, got false")
			}

			if resp.Meta == nil {
				t.Errorf("expected meta to be present")
			} else if resp.Meta.TotalCount != int64(tt.expectedLen) {
				t.Errorf("expected %d reviews, got %d", tt.expectedLen, resp.Meta.TotalCount)
			}
		})
	}
}

func TestAdminHandler_GetPendingCount(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewRepo := mock.NewReviewRepo()
	for i := 0; i < 3; i++ {
		review := &domain.Review{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			BathhouseID: uuid.New(),
			BookingID:   uuid.New(),
			Rating:      3,
			Text:        "Test",
			Status:      domain.ReviewStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := reviewRepo.Create(context.Background(), review); err != nil {
			t.Fatalf("failed to create review: %v", err)
		}
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Get("/admin/reviews/pending-count", adminH.GetPendingCount)

	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/pending-count", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	count, ok := data["pending_count"].(float64)
	if !ok {
		t.Fatalf("expected pending_count to be number, got %T", data["pending_count"])
	}

	if int(count) != 3 {
		t.Errorf("expected 3 pending reviews, got %d", int(count))
	}
}

func TestAdminHandler_ApproveReview(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      3,
		Text:        "Test review",
		Status:      domain.ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	reviewRepo := mock.NewReviewRepo()
	if err := reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/reviews/{id}/approve", adminH.ApproveReview)

	req := httptest.NewRequest(http.MethodPatch, "/admin/reviews/"+review.ID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if status, ok := respData["status"].(string); !ok || status != "approved" {
		t.Errorf("expected status=approved, got %v", respData["status"])
	}

	// Verify in repo
	updated, err := reviewRepo.GetByID(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("failed to get updated review: %v", err)
	}
	if updated.Status != domain.ReviewStatusApproved {
		t.Errorf("expected status=approved in repo, got %s", updated.Status)
	}
}

func TestAdminHandler_RejectReview(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      1,
		Text:        "Offensive content",
		Status:      domain.ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	reviewRepo := mock.NewReviewRepo()
	if err := reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/reviews/{id}/reject", adminH.RejectReview)

	body := bytes.NewBufferString(`{"reason":"offensive language"}`)
	req := httptest.NewRequest(http.MethodPatch, "/admin/reviews/"+review.ID.String()+"/reject", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if status, ok := respData["status"].(string); !ok || status != "rejected" {
		t.Errorf("expected status=rejected, got %v", respData["status"])
	}

	// Verify in repo
	updated, err := reviewRepo.GetByID(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("failed to get updated review: %v", err)
	}
	if updated.Status != domain.ReviewStatusRejected {
		t.Errorf("expected status=rejected in repo, got %s", updated.Status)
	}
	if len(updated.RejectionReasons) == 0 {
		t.Errorf("expected rejection reasons to be set")
	}
}

func TestAdminHandler_BatchApproveReviews(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewRepo := mock.NewReviewRepo()
	var ids []string
	for i := 0; i < 3; i++ {
		review := &domain.Review{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			BathhouseID: uuid.New(),
			BookingID:   uuid.New(),
			Rating:      3,
			Text:        "Test",
			Status:      domain.ReviewStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := reviewRepo.Create(context.Background(), review); err != nil {
			t.Fatalf("failed to create review: %v", err)
		}
		ids = append(ids, review.ID.String())
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/reviews/batch-approve", adminH.BatchApproveReviews)

	body := bytes.NewBufferString(`{"ids":["` + ids[0] + `","` + ids[1] + `"]}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/reviews/batch-approve", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	successful, ok := respData["successful"].(float64)
	if !ok || int(successful) != 2 {
		t.Errorf("expected successful=2, got %v", respData["successful"])
	}
}

func TestAdminHandler_BatchRejectReviews(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewRepo := mock.NewReviewRepo()
	var ids []string
	for i := 0; i < 2; i++ {
		review := &domain.Review{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			BathhouseID: uuid.New(),
			BookingID:   uuid.New(),
			Rating:      2,
			Text:        "Bad",
			Status:      domain.ReviewStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := reviewRepo.Create(context.Background(), review); err != nil {
			t.Fatalf("failed to create review: %v", err)
		}
		ids = append(ids, review.ID.String())
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/reviews/batch-reject", adminH.BatchRejectReviews)

	body := bytes.NewBufferString(`{"ids":["` + ids[0] + `","` + ids[1] + `"],"reason":"spam"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/reviews/batch-reject", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	successful, ok := respData["successful"].(float64)
	if !ok || int(successful) != 2 {
		t.Errorf("expected successful=2, got %v", respData["successful"])
	}
}

func TestAdminHandler_ApproveReview_SendsNotification(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewAuthorID := uuid.New()
	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      reviewAuthorID,
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      5,
		Text:        "Great place!",
		Status:      domain.ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	reviewRepo := mock.NewReviewRepo()
	if err := reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	notificationSent := false
	var notifUserID uuid.UUID
	var notifType domain.NotificationType

	notifService := &mockAdminNotificationService{
		sendFn: func(ctx context.Context, userID uuid.UUID, nt domain.NotificationType, title, body string, data map[string]string) error {
			notificationSent = true
			notifUserID = userID
			notifType = nt
			return nil
		},
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, notifService)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/reviews/{id}/approve", adminH.ApproveReview)

	req := httptest.NewRequest(http.MethodPatch, "/admin/reviews/"+review.ID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !notificationSent {
		t.Error("expected notification to be sent, but it wasn't")
	}

	if notifUserID != reviewAuthorID {
		t.Errorf("expected notification to be sent to review author %s, got %s", reviewAuthorID, notifUserID)
	}

	if notifType != domain.NotifReviewApproved {
		t.Errorf("expected notification type %s, got %s", domain.NotifReviewApproved, notifType)
	}
}

func TestAdminHandler_RejectReview_SendsNotification(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewAuthorID := uuid.New()
	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      reviewAuthorID,
		BathhouseID: uuid.New(),
		BookingID:   uuid.New(),
		Rating:      1,
		Text:        "Offensive content",
		Status:      domain.ReviewStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	reviewRepo := mock.NewReviewRepo()
	if err := reviewRepo.Create(context.Background(), review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	notificationSent := false
	var notifUserID uuid.UUID
	var notifType domain.NotificationType

	notifService := &mockAdminNotificationService{
		sendFn: func(ctx context.Context, userID uuid.UUID, nt domain.NotificationType, title, body string, data map[string]string) error {
			notificationSent = true
			notifUserID = userID
			notifType = nt
			return nil
		},
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, notifService)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/reviews/{id}/reject", adminH.RejectReview)

	body := bytes.NewBufferString(`{"reason":"offensive language"}`)
	req := httptest.NewRequest(http.MethodPatch, "/admin/reviews/"+review.ID.String()+"/reject", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !notificationSent {
		t.Error("expected notification to be sent, but it wasn't")
	}

	if notifUserID != reviewAuthorID {
		t.Errorf("expected notification to be sent to review author %s, got %s", reviewAuthorID, notifUserID)
	}

	if notifType != domain.NotifReviewRejected {
		t.Errorf("expected notification type %s, got %s", domain.NotifReviewRejected, notifType)
	}
}

func TestAdminHandler_BatchApproveReviews_SendsNotifications(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	reviewRepo := mock.NewReviewRepo()
	var ids []string
	var authorIDs []uuid.UUID

	for i := 0; i < 2; i++ {
		authorID := uuid.New()
		review := &domain.Review{
			ID:          uuid.New(),
			UserID:      authorID,
			BathhouseID: uuid.New(),
			BookingID:   uuid.New(),
			Rating:      5,
			Text:        "Good!",
			Status:      domain.ReviewStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := reviewRepo.Create(context.Background(), review); err != nil {
			t.Fatalf("failed to create review: %v", err)
		}
		ids = append(ids, review.ID.String())
		authorIDs = append(authorIDs, authorID)
	}

	notificationCount := 0
	notifService := &mockAdminNotificationService{
		sendFn: func(ctx context.Context, userID uuid.UUID, nt domain.NotificationType, title, body string, data map[string]string) error {
			notificationCount++
			return nil
		},
	}

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, reviewRepo, notifService)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/reviews/batch-approve", adminH.BatchApproveReviews)

	body := bytes.NewBufferString(`{"ids":["` + ids[0] + `","` + ids[1] + `"]}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/reviews/batch-approve", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if notificationCount != 2 {
		t.Errorf("expected 2 notifications to be sent, got %d", notificationCount)
	}
}
