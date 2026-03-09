package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

func createMultipartFile(fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile(fieldName, fileName)
	part.Write(content)
	w.Close()
	return &buf, w.FormDataContentType()
}

func TestReviewHandler_UploadMedia(t *testing.T) {
	userID := uuid.New()
	reviewID := uuid.New()
	mediaID := uuid.New()

	reviewSvc := &mockReviewService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Review, error) {
			return &domain.Review{
				ID:     reviewID,
				UserID: userID,
			}, nil
		},
	}

	mediaSvc := &mockMediaService{
		uploadFn: func(_ context.Context, uid uuid.UUID, input service.UploadMediaInput) (*domain.Media, error) {
			return &domain.Media{
				ID:           mediaID,
				OwnerType:    input.OwnerType,
				OwnerID:      input.OwnerID,
				UserID:       uid,
				Type:         domain.MediaTypeImage,
				URL:          "https://s3.example.com/media/test.jpg",
				ThumbnailURL: "https://s3.example.com/media/test_thumb.jpg",
				OriginalName: input.OriginalName,
				Size:         input.Size,
				MimeType:     "image/jpeg",
				Width:        800,
				Height:       600,
				Status:       domain.MediaStatusPending,
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc, mediaSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	// Create a fake JPEG file (starts with JPEG magic bytes)
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	jpegData = append(jpegData, make([]byte, 100)...)
	body, contentType := createMultipartFile("file", "photo.jpg", jpegData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/reviews/%s/media", reviewID), body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success, got error: %v", resp.Error)
	}
}

func TestReviewHandler_UploadMedia_InvalidReviewID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleClient)
	h := handler.NewReviewHandler(&mockReviewService{}, &mockMediaService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	body, contentType := createMultipartFile("file", "photo.jpg", []byte("test"))
	req := httptest.NewRequest(http.MethodPost, "/reviews/invalid-id/media", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_UploadMedia_ReviewNotFound(t *testing.T) {
	userID := uuid.New()

	reviewSvc := &mockReviewService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Review, error) {
			return nil, domain.ErrNotFound
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc, &mockMediaService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	body, contentType := createMultipartFile("file", "photo.jpg", []byte("test"))
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/reviews/%s/media", uuid.New()), body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestReviewHandler_UploadMedia_ForbiddenNotAuthor(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Review, error) {
			return &domain.Review{
				ID:     reviewID,
				UserID: otherUserID, // Different user
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc, &mockMediaService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	body, contentType := createMultipartFile("file", "photo.jpg", []byte("test"))
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/reviews/%s/media", reviewID), body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestReviewHandler_UploadMedia_NoFile(t *testing.T) {
	userID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Review, error) {
			return &domain.Review{
				ID:     reviewID,
				UserID: userID,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc, &mockMediaService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	// No file in request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/reviews/%s/media", reviewID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_UploadMedia_LimitReached(t *testing.T) {
	userID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Review, error) {
			return &domain.Review{
				ID:     reviewID,
				UserID: userID,
			}, nil
		},
	}

	mediaSvc := &mockMediaService{
		uploadFn: func(_ context.Context, uid uuid.UUID, input service.UploadMediaInput) (*domain.Media, error) {
			return nil, domain.ErrMediaLimitReached
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc, mediaSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/reviews/{id}/media", h.UploadMedia)

	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	jpegData = append(jpegData, make([]byte, 100)...)
	body, contentType := createMultipartFile("file", "photo.jpg", jpegData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/reviews/%s/media", reviewID), body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if resp.Error == nil || resp.Error.Code != "media_limit_reached" {
		t.Errorf("expected error code media_limit_reached, got %v", resp.Error)
	}
}

func TestReviewHandler_DeleteMedia(t *testing.T) {
	userID := uuid.New()
	mediaID := uuid.New()

	mediaSvc := &mockMediaService{
		deleteFn: func(_ context.Context, mID uuid.UUID, uid uuid.UUID, role domain.UserRole) error {
			if mID != mediaID {
				t.Errorf("expected media ID %s, got %s", mediaID, mID)
			}
			return nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(&mockReviewService{}, mediaSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/media/{id}", h.DeleteMedia)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/media/%s", mediaID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestReviewHandler_DeleteMedia_InvalidID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleClient)
	h := handler.NewReviewHandler(&mockReviewService{}, &mockMediaService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/media/{id}", h.DeleteMedia)

	req := httptest.NewRequest(http.MethodDelete, "/media/invalid-id", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_DeleteMedia_NotFound(t *testing.T) {
	userID := uuid.New()

	mediaSvc := &mockMediaService{
		deleteFn: func(_ context.Context, mID uuid.UUID, uid uuid.UUID, role domain.UserRole) error {
			return domain.ErrMediaNotFound
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(&mockReviewService{}, mediaSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/media/{id}", h.DeleteMedia)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/media/%s", uuid.New()), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestReviewHandler_DeleteMedia_Forbidden(t *testing.T) {
	userID := uuid.New()

	mediaSvc := &mockMediaService{
		deleteFn: func(_ context.Context, mID uuid.UUID, uid uuid.UUID, role domain.UserRole) error {
			return domain.ErrForbidden
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewReviewHandler(&mockReviewService{}, mediaSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/media/{id}", h.DeleteMedia)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/media/%s", uuid.New()), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestReviewHandler_ListByBathhouse_WithMedia(t *testing.T) {
	bhID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		listByBathhouseFn: func(_ context.Context, bID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
			return &domain.PaginatedResult[domain.Review]{
				Items: []domain.Review{
					{
						ID:          reviewID,
						UserID:      uuid.New(),
						BathhouseID: bhID,
						BookingID:   uuid.New(),
						Rating:      5,
						Text:        "Great!",
						Status:      domain.ReviewStatusApproved,
					},
				},
				TotalCount: 1,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 1,
			}, nil
		},
	}

	mediaSvc := &mockMediaService{
		listByReviewFn: func(_ context.Context, revID uuid.UUID) ([]domain.Media, error) {
			if revID != reviewID {
				return nil, nil
			}
			return []domain.Media{
				{
					ID:           uuid.New(),
					OwnerType:    domain.MediaOwnerReview,
					OwnerID:      reviewID,
					Type:         domain.MediaTypeImage,
					URL:          "https://s3.example.com/media/test.jpg",
					ThumbnailURL: "https://s3.example.com/media/test_thumb.jpg",
					Status:       domain.MediaStatusPending,
				},
			}, nil
		},
	}

	h := handler.NewReviewHandler(reviewSvc, mediaSvc)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/reviews", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/reviews", bhID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success, got error: %v", resp.Error)
	}
}
