package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// --- Mock services ---

type mockAuthService struct {
	registerFn   func(ctx context.Context, input service.RegisterInput) (*domain.User, string, error)
	loginFn      func(ctx context.Context, email, password string) (*domain.User, string, error)
	parseTokenFn func(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

func (m *mockAuthService) Register(ctx context.Context, input service.RegisterInput) (*domain.User, string, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, input)
	}
	return nil, "", nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, email, password)
	}
	return nil, "", nil
}

func (m *mockAuthService) ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error) {
	if m.parseTokenFn != nil {
		return m.parseTokenFn(ctx, token)
	}
	return uuid.Nil, "", domain.ErrUnauthorized
}

type mockUserService struct {
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	updateFn          func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*domain.User, error)
	uploadAvatarFn    func(ctx context.Context, userID uuid.UUID, input service.UploadAvatarInput) (*domain.User, error)
	deleteAvatarFn    func(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	getPublicProfileFn func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	listFn            func(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	blockFn           func(ctx context.Context, id uuid.UUID) error
	unblockFn         func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserService) Update(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*domain.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockUserService) UploadAvatar(ctx context.Context, userID uuid.UUID, input service.UploadAvatarInput) (*domain.User, error) {
	if m.uploadAvatarFn != nil {
		return m.uploadAvatarFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockUserService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	if m.deleteAvatarFn != nil {
		return m.deleteAvatarFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserService) GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	if m.getPublicProfileFn != nil {
		return m.getPublicProfileFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserService) List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, pageSize)
	}
	return &domain.PaginatedResult[domain.User]{}, nil
}

func (m *mockUserService) Block(ctx context.Context, id uuid.UUID) error {
	if m.blockFn != nil {
		return m.blockFn(ctx, id)
	}
	return nil
}

func (m *mockUserService) Unblock(ctx context.Context, id uuid.UUID) error {
	if m.unblockFn != nil {
		return m.unblockFn(ctx, id)
	}
	return nil
}

type mockBathhouseService struct {
	searchFn      func(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	getByIDFn     func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	createFn      func(ctx context.Context, ownerID uuid.UUID, input service.CreateBathhouseInput) (*domain.Bathhouse, error)
	updateFn      func(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input service.UpdateBathhouseInput) (*domain.Bathhouse, error)
	deleteFn      func(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error
	listByOwnerFn func(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	approveFn     func(ctx context.Context, id uuid.UUID) error
	rejectFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockBathhouseService) Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.Bathhouse]{}, nil
}

func (m *mockBathhouseService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockBathhouseService) Create(ctx context.Context, ownerID uuid.UUID, input service.CreateBathhouseInput) (*domain.Bathhouse, error) {
	if m.createFn != nil {
		return m.createFn(ctx, ownerID, input)
	}
	return nil, nil
}

func (m *mockBathhouseService) Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input service.UpdateBathhouseInput) (*domain.Bathhouse, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, userID, role, id, input)
	}
	return nil, nil
}

func (m *mockBathhouseService) Delete(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, ownerID, id)
	}
	return nil
}

func (m *mockBathhouseService) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	if m.listByOwnerFn != nil {
		return m.listByOwnerFn(ctx, ownerID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Bathhouse]{}, nil
}

func (m *mockBathhouseService) Approve(ctx context.Context, id uuid.UUID) error {
	if m.approveFn != nil {
		return m.approveFn(ctx, id)
	}
	return nil
}

func (m *mockBathhouseService) Reject(ctx context.Context, id uuid.UUID) error {
	if m.rejectFn != nil {
		return m.rejectFn(ctx, id)
	}
	return nil
}

type mockBookingService struct {
	createFn          func(ctx context.Context, userID uuid.UUID, input service.CreateBookingInput) (*domain.Booking, error)
	cancelFn          func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	confirmFn         func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	rejectFn          func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	completeFn        func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	listByUserFn      func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	listByBathhouseFn func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	getAvailSlotsFn   func(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]service.TimeSlot, error)
}

func (m *mockBookingService) Create(ctx context.Context, userID uuid.UUID, input service.CreateBookingInput) (*domain.Booking, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockBookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingService) Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.confirmFn != nil {
		return m.confirmFn(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingService) Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.rejectFn != nil {
		return m.rejectFn(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingService) Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.completeFn != nil {
		return m.completeFn(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Booking]{}, nil
}

func (m *mockBookingService) ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, userID, role, bathhouseID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Booking]{}, nil
}

func (m *mockBookingService) GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]service.TimeSlot, error) {
	if m.getAvailSlotsFn != nil {
		return m.getAvailSlotsFn(ctx, bathhouseID, date)
	}
	return nil, nil
}

type mockReviewService struct {
	createFn           func(ctx context.Context, userID uuid.UUID, input service.CreateReviewInput) (*domain.Review, error)
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	updateFn           func(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, input service.UpdateReviewInput) (*domain.Review, error)
	deleteFn           func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID) error
	listByBathhouseFn  func(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	addOwnerResponseFn func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID, response string) (*domain.Review, error)
}

func (m *mockReviewService) Create(ctx context.Context, userID uuid.UUID, input service.CreateReviewInput) (*domain.Review, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockReviewService) Update(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, input service.UpdateReviewInput) (*domain.Review, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, userID, reviewID, input)
	}
	return nil, nil
}

func (m *mockReviewService) Delete(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, userRole, reviewID)
	}
	return nil
}

func (m *mockReviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, bathhouseID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Review]{}, nil
}

func (m *mockReviewService) AddOwnerResponse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID, response string) (*domain.Review, error) {
	if m.addOwnerResponseFn != nil {
		return m.addOwnerResponseFn(ctx, userID, userRole, reviewID, response)
	}
	return nil, nil
}

type mockRepService struct {
	inviteFn          func(ctx context.Context, ownerID uuid.UUID, input service.InviteRepresentativeInput) (*domain.Representative, error)
	revokeFn          func(ctx context.Context, ownerID uuid.UUID, repID uuid.UUID) error
	listByBathhouseFn func(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error)
	getMyBathhousesFn func(ctx context.Context, userID uuid.UUID) ([]domain.Bathhouse, error)
}

func (m *mockRepService) Invite(ctx context.Context, ownerID uuid.UUID, input service.InviteRepresentativeInput) (*domain.Representative, error) {
	if m.inviteFn != nil {
		return m.inviteFn(ctx, ownerID, input)
	}
	return nil, nil
}

func (m *mockRepService) Revoke(ctx context.Context, ownerID uuid.UUID, repID uuid.UUID) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, ownerID, repID)
	}
	return nil
}

func (m *mockRepService) ListByBathhouse(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, ownerID, bathhouseID)
	}
	return nil, nil
}

func (m *mockRepService) GetMyBathhouses(ctx context.Context, userID uuid.UUID) ([]domain.Bathhouse, error) {
	if m.getMyBathhousesFn != nil {
		return m.getMyBathhousesFn(ctx, userID)
	}
	return nil, nil
}

type mockCityService struct {
	getAllFn    func(ctx context.Context) ([]domain.City, error)
	getBySlugFn func(ctx context.Context, slug string) (*domain.City, error)
	createFn    func(ctx context.Context, input service.CreateCityInput) (*domain.City, error)
	updateFn    func(ctx context.Context, id int64, input service.UpdateCityInput) (*domain.City, error)
	deleteFn    func(ctx context.Context, id int64) error
}

func (m *mockCityService) GetAll(ctx context.Context) ([]domain.City, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockCityService) GetBySlug(ctx context.Context, slug string) (*domain.City, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	return nil, nil
}

func (m *mockCityService) Create(ctx context.Context, input service.CreateCityInput) (*domain.City, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockCityService) Update(ctx context.Context, id int64, input service.UpdateCityInput) (*domain.City, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockCityService) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockFavoriteService struct {
	toggleFn     func(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
	listFn       func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error)
	isFavoriteFn func(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
}

func (m *mockFavoriteService) Toggle(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, userID, bathhouseID)
	}
	return false, nil
}

func (m *mockFavoriteService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Favorite]{}, nil
}

func (m *mockFavoriteService) IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	if m.isFavoriteFn != nil {
		return m.isFavoriteFn(ctx, userID, bathhouseID)
	}
	return false, nil
}

// --- Helpers ---

func jsonBody(v interface{}) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

type apiResp struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Meta *struct {
		Page       int   `json:"page"`
		PageSize   int   `json:"page_size"`
		TotalCount int64 `json:"total_count"`
		TotalPages int   `json:"total_pages"`
	} `json:"meta,omitempty"`
}

func parseResponse(t *testing.T, rec *httptest.ResponseRecorder) apiResp {
	t.Helper()
	var resp apiResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, rec.Body.String())
	}
	return resp
}

func makeAuthToken(userID uuid.UUID, role domain.UserRole) *mockAuthService {
	return &mockAuthService{
		parseTokenFn: func(_ context.Context, token string) (uuid.UUID, domain.UserRole, error) {
			if token == "valid-token" {
				return userID, role, nil
			}
			return uuid.Nil, "", domain.ErrUnauthorized
		},
	}
}

// --- Auth Handler Tests ---

func TestAuthHandler_Register(t *testing.T) {
	userID := uuid.New()
	authSvc := &mockAuthService{
		registerFn: func(_ context.Context, input service.RegisterInput) (*domain.User, string, error) {
			return &domain.User{
				ID:       userID,
				Email:    input.Email,
				Name:     input.Name,
				Phone:    input.Phone,
				Role:     input.Role,
				IsActive: true,
			}, "test-token", nil
		},
	}

	h := handler.NewAuthHandler(authSvc, nil)

	body := jsonBody(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
		"phone":    "+71234567890",
		"role":     "client",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	authSvc := &mockAuthService{
		registerFn: func(_ context.Context, input service.RegisterInput) (*domain.User, string, error) {
			return nil, "", domain.ErrInvalidInput
		},
	}

	h := handler.NewAuthHandler(authSvc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	userID := uuid.New()
	authSvc := &mockAuthService{
		loginFn: func(_ context.Context, email, password string) (*domain.User, string, error) {
			return &domain.User{
				ID:       userID,
				Email:    email,
				Name:     "Test",
				Role:     domain.RoleClient,
				IsActive: true,
			}, "jwt-token", nil
		},
	}

	h := handler.NewAuthHandler(authSvc, nil)

	body := jsonBody(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	authSvc := &mockAuthService{
		loginFn: func(_ context.Context, email, password string) (*domain.User, string, error) {
			return nil, "", domain.ErrUnauthorized
		},
	}

	h := handler.NewAuthHandler(authSvc, nil)

	body := jsonBody(map[string]string{
		"email":    "test@example.com",
		"password": "wrong",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Me(t *testing.T) {
	userID := uuid.New()
	userSvc := &mockUserService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID:       id,
				Email:    "test@example.com",
				Name:     "Test",
				Role:     domain.RoleClient,
				IsActive: true,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewAuthHandler(authSvc, userSvc)

	r := chi.NewRouter()
	r.With(middleware.RequireAuth(authSvc)).Get("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, body: %s", rec.Body.String())
	}
}

func TestAuthHandler_Me_Unauthenticated(t *testing.T) {
	authSvc := &mockAuthService{}
	h := handler.NewAuthHandler(authSvc, nil)

	r := chi.NewRouter()
	r.With(middleware.RequireAuth(authSvc)).Get("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_UpdateProfile(t *testing.T) {
	userID := uuid.New()
	userSvc := &mockUserService{
		updateFn: func(_ context.Context, id uuid.UUID, input service.UpdateUserInput) (*domain.User, error) {
			name := "Updated"
			if input.Name != nil {
				name = *input.Name
			}
			return &domain.User{
				ID: id, Email: "test@example.com", Name: name,
				Role: domain.RoleClient, IsActive: true,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewAuthHandler(authSvc, userSvc)

	r := chi.NewRouter()
	r.With(middleware.RequireAuth(authSvc)).Put("/auth/me", h.UpdateProfile)

	body := `{"name":"New Name","bio":"Hello world"}`
	req := httptest.NewRequest(http.MethodPut, "/auth/me", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, body: %s", rec.Body.String())
	}
}

func TestAuthHandler_DeleteAvatar(t *testing.T) {
	userID := uuid.New()
	userSvc := &mockUserService{
		deleteAvatarFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID: id, Email: "test@example.com", Name: "Test",
				Role: domain.RoleClient, IsActive: true, AvatarURL: "",
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewAuthHandler(authSvc, userSvc)

	r := chi.NewRouter()
	r.With(middleware.RequireAuth(authSvc)).Delete("/auth/me/avatar", h.DeleteAvatar)

	req := httptest.NewRequest(http.MethodDelete, "/auth/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, body: %s", rec.Body.String())
	}
}

func TestAuthHandler_GetPublicProfile(t *testing.T) {
	profileID := uuid.New()
	userSvc := &mockUserService{
		getPublicProfileFn: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:          id,
				Name:        "Test User",
				AvatarURL:   "http://example.com/avatar.jpg",
				Bio:         "Bio text",
				CityName:    "Moscow",
				MemberSince: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ReviewCount: 5,
				VisitCount:  10,
				AvgRating:   4.5,
			}, nil
		},
	}

	h := handler.NewAuthHandler(&mockAuthService{}, userSvc)

	r := chi.NewRouter()
	r.Get("/users/{id}/profile", h.GetPublicProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/"+profileID.String()+"/profile", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, body: %s", rec.Body.String())
	}
}

func TestAuthHandler_GetPublicProfile_NotFound(t *testing.T) {
	userSvc := &mockUserService{
		getPublicProfileFn: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewAuthHandler(&mockAuthService{}, userSvc)

	r := chi.NewRouter()
	r.Get("/users/{id}/profile", h.GetPublicProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String()+"/profile", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestAuthHandler_GetPublicProfile_InvalidID(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{}, &mockUserService{})

	r := chi.NewRouter()
	r.Get("/users/{id}/profile", h.GetPublicProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid/profile", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// --- Bathhouse Handler Tests ---

func TestBathhouseHandler_Search(t *testing.T) {
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{
					{ID: uuid.New(), Name: "Test Bath", Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test"},
				},
				TotalCount: 1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?city_id=1&has_pool=true", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Error("expected meta to be present")
	}
	if resp.Meta != nil && resp.Meta.TotalCount != 1 {
		t.Errorf("expected total_count=1, got %d", resp.Meta.TotalCount)
	}
}

func TestBathhouseHandler_GetByID(t *testing.T) {
	bhID := uuid.New()
	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Test Bath", Status: domain.BathhouseStatusActive,
				CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test",
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bhID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBathhouseHandler_GetByID_NotFound(t *testing.T) {
	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestBathhouseHandler_Create_RequiresOwnerRole(t *testing.T) {
	ownerID := uuid.New()
	bhSvc := &mockBathhouseService{
		createFn: func(_ context.Context, oID uuid.UUID, input service.CreateBathhouseInput) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: uuid.New(), OwnerID: oID, Name: input.Name,
				Status: domain.BathhouseStatusPending, CityID: input.CityID,
				PricePerHour: input.PricePerHour, MaxGuests: input.MaxGuests,
				MinDuration: input.MinDuration, Address: input.Address,
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses", h.Create)

	body := jsonBody(map[string]interface{}{
		"name": "My Bath", "address": "123 Street", "city_id": 1,
		"price_per_hour": 5000, "min_duration": 1, "max_guests": 10,
	})

	req := httptest.NewRequest(http.MethodPost, "/bathhouses", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBathhouseHandler_Create_ForbiddenForClient(t *testing.T) {
	clientID := uuid.New()
	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses", h.Create)

	body := jsonBody(map[string]interface{}{"name": "My Bath"})
	req := httptest.NewRequest(http.MethodPost, "/bathhouses", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for client, got %d", rec.Code)
	}
}

// --- Booking Handler Tests ---

func TestBookingHandler_Create(t *testing.T) {
	clientID := uuid.New()
	bhID := uuid.New()
	bookingSvc := &mockBookingService{
		createFn: func(_ context.Context, userID uuid.UUID, input service.CreateBookingInput) (*domain.Booking, error) {
			return &domain.Booking{
				ID: uuid.New(), UserID: userID, BathhouseID: input.BathhouseID,
				StartTime: input.StartTime, EndTime: input.EndTime,
				GuestCount: input.GuestCount, TotalPrice: 10000, Status: domain.BookingPending,
			}, nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleClient)).Post("/bookings", h.Create)

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	body := jsonBody(map[string]interface{}{
		"bathhouse_id": bhID.String(),
		"start_time":   startTime.Format(time.RFC3339),
		"end_time":     endTime.Format(time.RFC3339),
		"guest_count":  5,
		"comment":      "test",
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_Create_ForbiddenForOwner(t *testing.T) {
	ownerID := uuid.New()
	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBookingHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleClient)).Post("/bookings", h.Create)

	body := jsonBody(map[string]interface{}{"bathhouse_id": uuid.New().String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestBookingHandler_Cancel(t *testing.T) {
	clientID := uuid.New()
	bookingID := uuid.New()
	bookingSvc := &mockBookingService{
		cancelFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, bID uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/cancel", h.Cancel)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/"+bookingID.String()+"/cancel", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_Confirm_RequiresOwnerOrRep(t *testing.T) {
	clientID := uuid.New()
	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewBookingHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/"+uuid.New().String()+"/confirm", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for client, got %d", rec.Code)
	}
}

// --- Admin Handler Tests ---

func TestAdminHandler_ListUsers_RequiresAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     domain.UserRole
		expected int
	}{
		{"admin allowed", domain.RoleAdmin, http.StatusOK},
		{"client forbidden", domain.RoleClient, http.StatusForbidden},
		{"owner forbidden", domain.RoleOwner, http.StatusForbidden},
		{"representative forbidden", domain.RoleRepresentative, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			authSvc := makeAuthToken(userID, tt.role)

			userSvc := &mockUserService{
				listFn: func(_ context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error) {
					return &domain.PaginatedResult[domain.User]{
						Items:      []domain.User{},
						TotalCount: 0,
						Page:       page,
						PageSize:   pageSize,
					}, nil
				},
			}
			adminH := handler.NewAdminHandler(userSvc, nil, nil)

			router := chi.NewRouter()
			router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Get("/admin/users", adminH.ListUsers)

			req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expected {
				t.Errorf("expected status %d, got %d", tt.expected, rec.Code)
			}
		})
	}
}

func TestAdminHandler_BlockUser(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	userSvc := &mockUserService{
		blockFn: func(_ context.Context, id uuid.UUID) error {
			if id != targetID {
				t.Errorf("expected block of %s, got %s", targetID, id)
			}
			return nil
		},
	}
	adminH := handler.NewAdminHandler(userSvc, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/users/{id}/block", adminH.BlockUser)

	req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+targetID.String()+"/block", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminHandler_ApproveBathhouse(t *testing.T) {
	adminID := uuid.New()
	bhID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	bhSvc := &mockBathhouseService{
		approveFn: func(_ context.Context, id uuid.UUID) error {
			if id != bhID {
				t.Errorf("expected approve of %s, got %s", bhID, id)
			}
			return nil
		},
	}
	adminH := handler.NewAdminHandler(nil, bhSvc, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/bathhouses/{id}/approve", adminH.ApproveBathhouse)

	req := httptest.NewRequest(http.MethodPatch, "/admin/bathhouses/"+bhID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_CreateCity(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	citySvc := &mockCityService{
		createFn: func(_ context.Context, input service.CreateCityInput) (*domain.City, error) {
			return &domain.City{ID: 1, Name: input.Name, Slug: input.Slug}, nil
		},
	}
	adminH := handler.NewAdminHandler(nil, nil, citySvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/cities", adminH.CreateCity)

	body := jsonBody(map[string]interface{}{
		"name": "Moscow", "slug": "moscow", "latitude": 55.75, "longitude": 37.62,
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/cities", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

// --- Review Handler Tests ---

func TestReviewHandler_Create(t *testing.T) {
	clientID := uuid.New()
	bookingID := uuid.New()
	bhID := uuid.New()

	reviewSvc := &mockReviewService{
		createFn: func(_ context.Context, userID uuid.UUID, input service.CreateReviewInput) (*domain.Review, error) {
			return &domain.Review{
				ID: uuid.New(), UserID: userID, BathhouseID: bhID,
				BookingID: input.BookingID, Rating: input.Rating, Text: input.Text,
			}, nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleClient)).Post("/bathhouses/{id}/reviews", h.Create)

	body := jsonBody(map[string]interface{}{
		"booking_id": bookingID.String(),
		"rating":     5,
		"text":       "Great bath!",
	})

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/reviews", bhID), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestReviewHandler_ListByBathhouse(t *testing.T) {
	bhID := uuid.New()

	reviewSvc := &mockReviewService{
		listByBathhouseFn: func(_ context.Context, bID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
			return &domain.PaginatedResult[domain.Review]{
				Items:      []domain.Review{},
				TotalCount: 0,
				Page:       page,
				PageSize:   pageSize,
			}, nil
		},
	}

	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/reviews", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/reviews", bhID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- City Handler Tests ---

func TestCityHandler_List(t *testing.T) {
	citySvc := &mockCityService{
		getAllFn: func(_ context.Context) ([]domain.City, error) {
			return []domain.City{
				{ID: 1, Name: "Moscow", Slug: "moscow"},
				{ID: 2, Name: "SPb", Slug: "spb"},
			}, nil
		},
	}

	h := handler.NewCityHandler(citySvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cities", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

// --- Representative Handler Tests ---

func TestRepresentativeHandler_Invite(t *testing.T) {
	ownerID := uuid.New()
	bhID := uuid.New()

	repSvc := &mockRepService{
		inviteFn: func(_ context.Context, oID uuid.UUID, input service.InviteRepresentativeInput) (*domain.Representative, error) {
			return &domain.Representative{
				ID: uuid.New(), UserID: uuid.New(), BathhouseID: input.BathhouseID, OwnerID: oID,
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewRepresentativeHandler(repSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses/{id}/representatives", h.Invite)

	body := jsonBody(map[string]string{"user_email": "rep@example.com"})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/representatives", bhID), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestRepresentativeHandler_Invite_ForbiddenForClient(t *testing.T) {
	clientID := uuid.New()
	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewRepresentativeHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses/{id}/representatives", h.Invite)

	body := jsonBody(map[string]string{"user_email": "rep@example.com"})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/representatives", uuid.New()), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// --- Response format test ---

func TestHandleServiceError_MapsCorrectly(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"already exists", domain.ErrAlreadyExists, http.StatusConflict, "already_exists"},
		{"invalid input", domain.ErrInvalidInput, http.StatusBadRequest, "invalid_input"},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"slot unavailable", domain.ErrSlotUnavailable, http.StatusConflict, "slot_unavailable"},
		{"cancel too late", domain.ErrBookingCancelLate, http.StatusBadRequest, "cancel_too_late"},
		{"user blocked", domain.ErrUserBlocked, http.StatusForbidden, "user_blocked"},
		{"bathhouse not active", domain.ErrBathhouseNotActive, http.StatusBadRequest, "bathhouse_not_active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through auth handler login, since it calls handleServiceError
			authSvc := &mockAuthService{
				loginFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
					return nil, "", tt.err
				},
			}
			h := handler.NewAuthHandler(authSvc, nil)

			body := jsonBody(map[string]string{"email": "a@b.com", "password": "x"})
			req := httptest.NewRequest(http.MethodPost, "/login", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			resp := parseResponse(t, rec)
			if resp.Error == nil {
				t.Fatal("expected error in response")
			}
			if resp.Error.Code != tt.wantCode {
				t.Errorf("expected error code %q, got %q", tt.wantCode, resp.Error.Code)
			}
		})
	}
}

// --- Bathhouse Handler Update/Delete/AvailableSlots/MyBathhouses Tests ---

func TestBathhouseHandler_Update(t *testing.T) {
	ownerID := uuid.New()
	bhID := uuid.New()
	bhSvc := &mockBathhouseService{
		updateFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input service.UpdateBathhouseInput) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, OwnerID: userID, Name: *input.Name,
				Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000,
				MaxGuests: 10, MinDuration: 1, Address: "Test",
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner, domain.RoleRepresentative)).Put("/bathhouses/{id}", h.Update)

	body := jsonBody(map[string]interface{}{"name": "Updated Name"})
	req := httptest.NewRequest(http.MethodPut, "/bathhouses/"+bhID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBathhouseHandler_Delete(t *testing.T) {
	ownerID := uuid.New()
	bhID := uuid.New()
	bhSvc := &mockBathhouseService{
		deleteFn: func(_ context.Context, oID uuid.UUID, id uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Delete("/bathhouses/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/bathhouses/"+bhID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBathhouseHandler_GetAvailableSlots(t *testing.T) {
	bhID := uuid.New()
	bookingSvc := &mockBookingService{
		getAvailSlotsFn: func(_ context.Context, id uuid.UUID, date time.Time) ([]service.TimeSlot, error) {
			return []service.TimeSlot{
				{StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Available: true},
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(nil, bookingSvc, nil, nil)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/available-slots", h.GetAvailableSlots)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/available-slots?date=2026-03-01", bhID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBathhouseHandler_GetAvailableSlots_NoDate(t *testing.T) {
	bhID := uuid.New()
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/available-slots", h.GetAvailableSlots)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/available-slots", bhID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestBathhouseHandler_MyBathhouses_Owner(t *testing.T) {
	ownerID := uuid.New()
	bhSvc := &mockBathhouseService{
		listByOwnerFn: func(_ context.Context, oID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items:      []domain.Bathhouse{{ID: uuid.New(), Name: "My Bath", OwnerID: oID, Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test"}},
				TotalCount: 1, Page: 1, PageSize: 20, TotalPages: 1,
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner, domain.RoleRepresentative)).Get("/my/bathhouses", h.MyBathhouses)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBathhouseHandler_MyBathhouses_Representative(t *testing.T) {
	repID := uuid.New()
	repSvc := &mockRepService{
		getMyBathhousesFn: func(_ context.Context, userID uuid.UUID) ([]domain.Bathhouse, error) {
			return []domain.Bathhouse{
				{ID: uuid.New(), Name: "Rep Bath", Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test"},
			}, nil
		},
	}

	authSvc := makeAuthToken(repID, domain.RoleRepresentative)
	h := handler.NewBathhouseHandler(nil, nil, repSvc, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner, domain.RoleRepresentative)).Get("/my/bathhouses", h.MyBathhouses)

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

// --- Booking Handler Additional Tests ---

func TestBookingHandler_ListByUser(t *testing.T) {
	clientID := uuid.New()
	bookingSvc := &mockBookingService{
		listByUserFn: func(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
			return &domain.PaginatedResult[domain.Booking]{
				Items: []domain.Booking{}, TotalCount: 0, Page: 1, PageSize: 20,
			}, nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/bookings", h.ListByUser)

	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBookingHandler_Confirm(t *testing.T) {
	ownerID := uuid.New()
	bookingID := uuid.New()
	bookingSvc := &mockBookingService{
		confirmFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, bID uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/"+bookingID.String()+"/confirm", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_Reject(t *testing.T) {
	ownerID := uuid.New()
	bookingID := uuid.New()
	bookingSvc := &mockBookingService{
		rejectFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, bID uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/reject", h.Reject)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/"+bookingID.String()+"/reject", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_ListByBathhouse(t *testing.T) {
	ownerID := uuid.New()
	bhID := uuid.New()
	bookingSvc := &mockBookingService{
		listByBathhouseFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
			return &domain.PaginatedResult[domain.Booking]{
				Items: []domain.Booking{}, TotalCount: 0, Page: 1, PageSize: 20,
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewBookingHandler(bookingSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/bathhouses/{id}/bookings", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bhID.String()+"/bookings", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

// --- Admin Handler Additional Tests ---

func TestAdminHandler_UnblockUser(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	userSvc := &mockUserService{
		unblockFn: func(_ context.Context, id uuid.UUID) error {
			return nil
		},
	}
	adminH := handler.NewAdminHandler(userSvc, nil, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/users/{id}/unblock", adminH.UnblockUser)

	req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+targetID.String()+"/unblock", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminHandler_RejectBathhouse(t *testing.T) {
	adminID := uuid.New()
	bhID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	bhSvc := &mockBathhouseService{
		rejectFn: func(_ context.Context, id uuid.UUID) error {
			return nil
		},
	}
	adminH := handler.NewAdminHandler(nil, bhSvc, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/bathhouses/{id}/reject", adminH.RejectBathhouse)

	req := httptest.NewRequest(http.MethodPatch, "/admin/bathhouses/"+bhID.String()+"/reject", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminHandler_ListBathhouses(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20,
			}, nil
		},
	}
	adminH := handler.NewAdminHandler(nil, bhSvc, nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Get("/admin/bathhouses", adminH.ListBathhouses)

	req := httptest.NewRequest(http.MethodGet, "/admin/bathhouses?status=pending", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminHandler_UpdateCity(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	citySvc := &mockCityService{
		updateFn: func(_ context.Context, id int64, input service.UpdateCityInput) (*domain.City, error) {
			name := "Updated"
			if input.Name != nil {
				name = *input.Name
			}
			return &domain.City{ID: id, Name: name, Slug: "moscow"}, nil
		},
	}
	adminH := handler.NewAdminHandler(nil, nil, citySvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Put("/admin/cities/{id}", adminH.UpdateCity)

	body := jsonBody(map[string]interface{}{"name": "Updated Moscow"})
	req := httptest.NewRequest(http.MethodPut, "/admin/cities/1", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_DeleteCity(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	citySvc := &mockCityService{
		deleteFn: func(_ context.Context, id int64) error {
			return nil
		},
	}
	adminH := handler.NewAdminHandler(nil, nil, citySvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Delete("/admin/cities/{id}", adminH.DeleteCity)

	req := httptest.NewRequest(http.MethodDelete, "/admin/cities/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- Representative Handler Additional Tests ---

func TestRepresentativeHandler_ListByBathhouse(t *testing.T) {
	ownerID := uuid.New()
	bhID := uuid.New()
	repSvc := &mockRepService{
		listByBathhouseFn: func(_ context.Context, oID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error) {
			return []domain.Representative{}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewRepresentativeHandler(repSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Get("/bathhouses/{id}/representatives", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/representatives", bhID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRepresentativeHandler_Revoke(t *testing.T) {
	ownerID := uuid.New()
	repID := uuid.New()
	repSvc := &mockRepService{
		revokeFn: func(_ context.Context, oID uuid.UUID, rID uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewRepresentativeHandler(repSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Delete("/representatives/{id}", h.Revoke)

	req := httptest.NewRequest(http.MethodDelete, "/representatives/"+repID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- Search with various filters ---

func TestBathhouseHandler_Search_WithAllFilters(t *testing.T) {
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 10, TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?page=2&page_size=10&city_slug=moscow&price_min=1000&price_max=10000&min_guests=5&has_sauna=true&has_steam_room=true&has_hot_tub=true&has_bbq=true&has_karaoke=true&min_rating=4.0&lat=55.75&lng=37.62&radius_km=10&sort_by=price&sort_order=asc", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBathhouseHandler_Search_WithExtendedFilters(t *testing.T) {
	var capturedFilter domain.BathhouseFilter
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			capturedFilter = filter
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20, TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?guest_count=8&available_date=2026-03-15&available_time_from=10:00&available_time_to=14:00&open_now=true&q=русская+баня", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if capturedFilter.GuestCount == nil || *capturedFilter.GuestCount != 8 {
		t.Error("expected GuestCount=8")
	}
	if capturedFilter.AvailableDate == nil {
		t.Error("expected AvailableDate to be set")
	} else if capturedFilter.AvailableDate.Format("2006-01-02") != "2026-03-15" {
		t.Errorf("expected AvailableDate=2026-03-15, got %s", capturedFilter.AvailableDate.Format("2006-01-02"))
	}
	if capturedFilter.AvailableTimeFrom == nil || *capturedFilter.AvailableTimeFrom != "10:00" {
		t.Error("expected AvailableTimeFrom=10:00")
	}
	if capturedFilter.AvailableTimeTo == nil || *capturedFilter.AvailableTimeTo != "14:00" {
		t.Error("expected AvailableTimeTo=14:00")
	}
	if capturedFilter.OpenNow == nil || !*capturedFilter.OpenNow {
		t.Error("expected OpenNow=true")
	}
	if capturedFilter.SearchQuery == nil || *capturedFilter.SearchQuery != "русская баня" {
		t.Errorf("expected SearchQuery='русская баня', got %v", capturedFilter.SearchQuery)
	}
}

func TestBathhouseHandler_Search_InvalidTimeParams(t *testing.T) {
	var capturedFilter domain.BathhouseFilter
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			capturedFilter = filter
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20, TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?available_time_from=invalid&available_time_to=25:00&available_date=bad-date", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 (invalid params ignored), got %d", rec.Code)
	}

	if capturedFilter.AvailableTimeFrom != nil {
		t.Error("expected AvailableTimeFrom to be nil for invalid value")
	}
	if capturedFilter.AvailableTimeTo != nil {
		t.Error("expected AvailableTimeTo to be nil for invalid value")
	}
	if capturedFilter.AvailableDate != nil {
		t.Error("expected AvailableDate to be nil for invalid value")
	}
}

func TestBathhouseHandler_Search_GuestCountOnly(t *testing.T) {
	var capturedFilter domain.BathhouseFilter
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			capturedFilter = filter
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20, TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?guest_count=5", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedFilter.GuestCount == nil || *capturedFilter.GuestCount != 5 {
		t.Error("expected GuestCount=5")
	}
}

func TestBathhouseHandler_Search_SearchQueryOnly(t *testing.T) {
	var capturedFilter domain.BathhouseFilter
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			capturedFilter = filter
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20, TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?q=sauna", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedFilter.SearchQuery == nil || *capturedFilter.SearchQuery != "sauna" {
		t.Errorf("expected SearchQuery='sauna', got %v", capturedFilter.SearchQuery)
	}
}

func TestBathhouseHandler_Search_InvalidPage(t *testing.T) {
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{}, TotalCount: 0, Page: 1, PageSize: 20,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bathhouses?page=invalid&page_size=-1", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 (defaults should apply), got %d", rec.Code)
	}
}

// --- Invalid ID error paths ---

func TestBathhouseHandler_GetByID_InvalidUUID(t *testing.T) {
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)
	router := chi.NewRouter()
	router.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestBookingHandler_Cancel_InvalidUUID(t *testing.T) {
	h := handler.NewBookingHandler(nil)
	router := chi.NewRouter()
	authSvc := makeAuthToken(uuid.New(), domain.RoleClient)
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/cancel", h.Cancel)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/not-a-uuid/cancel", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestBookingHandler_Confirm_InvalidUUID(t *testing.T) {
	h := handler.NewBookingHandler(nil)
	router := chi.NewRouter()
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/not-a-uuid/confirm", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestBookingHandler_Reject_InvalidUUID(t *testing.T) {
	h := handler.NewBookingHandler(nil)
	router := chi.NewRouter()
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	router.With(middleware.RequireAuth(authSvc)).Patch("/bookings/{id}/reject", h.Reject)

	req := httptest.NewRequest(http.MethodPatch, "/bookings/not-a-uuid/reject", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestBookingHandler_ListByBathhouse_InvalidUUID(t *testing.T) {
	h := handler.NewBookingHandler(nil)
	router := chi.NewRouter()
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	router.With(middleware.RequireAuth(authSvc)).Get("/bathhouses/{id}/bookings", h.ListByBathhouse)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/not-a-uuid/bookings", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestAdminHandler_BlockUser_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleAdmin)
	adminH := handler.NewAdminHandler(nil, nil, nil)
	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Patch("/admin/users/{id}/block", adminH.BlockUser)

	req := httptest.NewRequest(http.MethodPatch, "/admin/users/not-a-uuid/block", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestAdminHandler_UpdateCity_InvalidID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleAdmin)
	adminH := handler.NewAdminHandler(nil, nil, nil)
	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Put("/admin/cities/{id}", adminH.UpdateCity)

	body := jsonBody(map[string]interface{}{"name": "Test"})
	req := httptest.NewRequest(http.MethodPut, "/admin/cities/abc", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestBathhouseHandler_GetAvailableSlots_InvalidDate(t *testing.T) {
	bhID := uuid.New()
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)
	router := chi.NewRouter()
	router.Get("/bathhouses/{id}/available-slots", h.GetAvailableSlots)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bathhouses/%s/available-slots?date=not-a-date", bhID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestBathhouseHandler_Update_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)
	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Put("/bathhouses/{id}", h.Update)

	body := jsonBody(map[string]interface{}{"name": "Test"})
	req := httptest.NewRequest(http.MethodPut, "/bathhouses/not-a-uuid", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestBathhouseHandler_Delete_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	h := handler.NewBathhouseHandler(nil, nil, nil, nil)
	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleOwner)).Delete("/bathhouses/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/bathhouses/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid UUID, got %d", rec.Code)
	}
}

// --- Unauthenticated access tests ---

// --- Review Handler Extended Tests ---

func TestReviewHandler_Update(t *testing.T) {
	clientID := uuid.New()
	reviewID := uuid.New()
	bhID := uuid.New()
	bookingID := uuid.New()

	reviewSvc := &mockReviewService{
		updateFn: func(_ context.Context, userID uuid.UUID, rID uuid.UUID, input service.UpdateReviewInput) (*domain.Review, error) {
			rating := 4
			if input.Rating != nil {
				rating = *input.Rating
			}
			return &domain.Review{
				ID: rID, UserID: userID, BathhouseID: bhID,
				BookingID: bookingID, Rating: rating, Text: "Updated text",
				Status: domain.ReviewStatusPending,
			}, nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Put("/reviews/{id}", h.Update)

	newRating := 4
	newText := "Updated text"
	body := jsonBody(map[string]interface{}{
		"rating": newRating,
		"text":   newText,
	})

	req := httptest.NewRequest(http.MethodPut, "/reviews/"+reviewID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestReviewHandler_Update_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleClient)
	h := handler.NewReviewHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Put("/reviews/{id}", h.Update)

	body := jsonBody(map[string]interface{}{"rating": 3})
	req := httptest.NewRequest(http.MethodPut, "/reviews/not-a-uuid", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_Update_Forbidden(t *testing.T) {
	clientID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		updateFn: func(_ context.Context, userID uuid.UUID, rID uuid.UUID, input service.UpdateReviewInput) (*domain.Review, error) {
			return nil, domain.ErrForbidden
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Put("/reviews/{id}", h.Update)

	body := jsonBody(map[string]interface{}{"rating": 3})
	req := httptest.NewRequest(http.MethodPut, "/reviews/"+reviewID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestReviewHandler_Delete(t *testing.T) {
	clientID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		deleteFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, rID uuid.UUID) error {
			return nil
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/reviews/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+reviewID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestReviewHandler_Delete_AdminCanDelete(t *testing.T) {
	adminID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		deleteFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, rID uuid.UUID) error {
			if role != domain.RoleAdmin {
				t.Error("expected admin role")
			}
			return nil
		},
	}

	authSvc := makeAuthToken(adminID, domain.RoleAdmin)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/reviews/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+reviewID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestReviewHandler_Delete_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleClient)
	h := handler.NewReviewHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/reviews/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_Delete_Forbidden(t *testing.T) {
	clientID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		deleteFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, rID uuid.UUID) error {
			return domain.ErrForbidden
		},
	}

	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Delete("/reviews/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/"+reviewID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestReviewHandler_AddOwnerResponse(t *testing.T) {
	ownerID := uuid.New()
	reviewID := uuid.New()
	bhID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	reviewSvc := &mockReviewService{
		addOwnerResponseFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, rID uuid.UUID, response string) (*domain.Review, error) {
			return &domain.Review{
				ID: rID, UserID: uuid.New(), BathhouseID: bhID,
				BookingID: bookingID, Rating: 5, Text: "Great!",
				Status:        domain.ReviewStatusApproved,
				OwnerResponse: response, OwnerResponseAt: &now,
			}, nil
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", h.AddOwnerResponse)

	body := jsonBody(map[string]string{"response": "Thank you for your review!"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/"+reviewID.String()+"/response", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestReviewHandler_AddOwnerResponse_InvalidUUID(t *testing.T) {
	authSvc := makeAuthToken(uuid.New(), domain.RoleOwner)
	h := handler.NewReviewHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", h.AddOwnerResponse)

	body := jsonBody(map[string]string{"response": "test"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/not-a-uuid/response", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReviewHandler_AddOwnerResponse_ForbiddenForClient(t *testing.T) {
	clientID := uuid.New()
	authSvc := makeAuthToken(clientID, domain.RoleClient)
	h := handler.NewReviewHandler(nil)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", h.AddOwnerResponse)

	body := jsonBody(map[string]string{"response": "test"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/"+uuid.New().String()+"/response", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestReviewHandler_AddOwnerResponse_AlreadyResponded(t *testing.T) {
	ownerID := uuid.New()
	reviewID := uuid.New()

	reviewSvc := &mockReviewService{
		addOwnerResponseFn: func(_ context.Context, userID uuid.UUID, role domain.UserRole, rID uuid.UUID, response string) (*domain.Review, error) {
			return nil, domain.ErrReviewAlreadyResponded
		},
	}

	authSvc := makeAuthToken(ownerID, domain.RoleOwner)
	h := handler.NewReviewHandler(reviewSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", h.AddOwnerResponse)

	body := jsonBody(map[string]string{"response": "test"})
	req := httptest.NewRequest(http.MethodPost, "/reviews/"+reviewID.String()+"/response", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	if resp.Error == nil || resp.Error.Code != "review_already_responded" {
		t.Errorf("expected error code review_already_responded, got %v", resp.Error)
	}
}

func TestProtectedEndpoints_RequireAuth(t *testing.T) {
	authSvc := &mockAuthService{} // ParseToken returns error by default

	// Build a minimal router with protected routes
	router := chi.NewRouter()
	auth := middleware.RequireAuth(authSvc)

	bookingH := handler.NewBookingHandler(nil)
	bhH := handler.NewBathhouseHandler(nil, nil, nil, nil)

	router.With(auth).Get("/bookings", bookingH.ListByUser)
	router.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings", bookingH.Create)
	router.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses", bhH.Create)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/bookings"},
		{http.MethodPost, "/bookings"},
		{http.MethodPost, "/bathhouses"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.method == http.MethodPost {
				req = httptest.NewRequest(ep.method, ep.path, jsonBody(map[string]string{}))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401 for unauthenticated %s %s, got %d", ep.method, ep.path, rec.Code)
			}
		})
	}
}

// --- Favorite Handler Tests ---

func TestFavoriteHandler_Toggle_Add(t *testing.T) {
	userID := uuid.New()
	bhID := uuid.New()

	favSvc := &mockFavoriteService{
		toggleFn: func(_ context.Context, uID, bID uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/bathhouses/{id}/favorite", h.Toggle)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/favorite", bhID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestFavoriteHandler_Toggle_Remove(t *testing.T) {
	userID := uuid.New()
	bhID := uuid.New()

	favSvc := &mockFavoriteService{
		toggleFn: func(_ context.Context, uID, bID uuid.UUID) (bool, error) {
			return false, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/bathhouses/{id}/favorite", h.Toggle)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/favorite", bhID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestFavoriteHandler_Toggle_InvalidUUID(t *testing.T) {
	userID := uuid.New()

	favSvc := &mockFavoriteService{}
	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/bathhouses/{id}/favorite", h.Toggle)

	req := httptest.NewRequest(http.MethodPost, "/bathhouses/invalid-uuid/favorite", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestFavoriteHandler_Toggle_BathhouseNotFound(t *testing.T) {
	userID := uuid.New()
	bhID := uuid.New()

	favSvc := &mockFavoriteService{
		toggleFn: func(_ context.Context, uID, bID uuid.UUID) (bool, error) {
			return false, domain.ErrNotFound
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Post("/bathhouses/{id}/favorite", h.Toggle)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/bathhouses/%s/favorite", bhID), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestFavoriteHandler_List(t *testing.T) {
	userID := uuid.New()

	favSvc := &mockFavoriteService{
		listFn: func(_ context.Context, uID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
			return &domain.PaginatedResult[domain.Favorite]{
				Items:      []domain.Favorite{},
				TotalCount: 0,
				Page:       page,
				PageSize:   pageSize,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/favorites", h.List)

	req := httptest.NewRequest(http.MethodGet, "/my/favorites", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestFavoriteHandler_List_WithPagination(t *testing.T) {
	userID := uuid.New()

	favSvc := &mockFavoriteService{
		listFn: func(_ context.Context, uID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
			if page != 2 {
				t.Errorf("expected page=2, got %d", page)
			}
			if pageSize != 5 {
				t.Errorf("expected pageSize=5, got %d", pageSize)
			}
			return &domain.PaginatedResult[domain.Favorite]{
				Items:      []domain.Favorite{},
				TotalCount: 10,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: 2,
			}, nil
		},
	}

	authSvc := makeAuthToken(userID, domain.RoleClient)
	h := handler.NewFavoriteHandler(favSvc)

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc)).Get("/my/favorites", h.List)

	req := httptest.NewRequest(http.MethodGet, "/my/favorites?page=2&page_size=5", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- IsFavorite in Bathhouse responses ---

func TestBathhouseHandler_Search_IsFavorite_Authenticated(t *testing.T) {
	bhID := uuid.New()
	userID := uuid.New()

	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{
					{ID: bhID, Name: "Fav Bath", Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test"},
				},
				TotalCount: 1, Page: 1, PageSize: 20, TotalPages: 1,
			}, nil
		},
	}
	favSvc := &mockFavoriteService{
		isFavoriteFn: func(_ context.Context, uid, bid uuid.UUID) (bool, error) {
			if uid == userID && bid == bhID {
				return true, nil
			}
			return false, nil
		},
	}
	authSvc := makeAuthToken(userID, domain.RoleClient)

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, favSvc)

	router := chi.NewRouter()
	router.With(middleware.OptionalAuth(authSvc)).Get("/bathhouses", h.Search)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	var items []map[string]interface{}
	if err := json.Unmarshal(resp.Data, &items); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if isFav, ok := items[0]["is_favorite"].(bool); !ok || !isFav {
		t.Errorf("expected is_favorite=true, got %v", items[0]["is_favorite"])
	}
}

func TestBathhouseHandler_Search_IsFavorite_Unauthenticated(t *testing.T) {
	bhSvc := &mockBathhouseService{
		searchFn: func(_ context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
			return &domain.PaginatedResult[domain.Bathhouse]{
				Items: []domain.Bathhouse{
					{ID: uuid.New(), Name: "Bath", Status: domain.BathhouseStatusActive, CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test"},
				},
				TotalCount: 1, Page: 1, PageSize: 20, TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses", nil)
	rec := httptest.NewRecorder()

	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	var items []map[string]interface{}
	if err := json.Unmarshal(resp.Data, &items); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if isFav, ok := items[0]["is_favorite"].(bool); !ok || isFav {
		t.Errorf("expected is_favorite=false, got %v", items[0]["is_favorite"])
	}
}

func TestBathhouseHandler_GetByID_IsFavorite_Authenticated(t *testing.T) {
	bhID := uuid.New()
	userID := uuid.New()

	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Fav Bath", Status: domain.BathhouseStatusActive,
				CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test",
			}, nil
		},
	}
	favSvc := &mockFavoriteService{
		isFavoriteFn: func(_ context.Context, uid, bid uuid.UUID) (bool, error) {
			if uid == userID && bid == bhID {
				return true, nil
			}
			return false, nil
		},
	}
	authSvc := makeAuthToken(userID, domain.RoleClient)

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, favSvc)

	router := chi.NewRouter()
	router.With(middleware.OptionalAuth(authSvc)).Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bhID.String(), nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	resp := parseResponse(t, rec)
	var item map[string]interface{}
	if err := json.Unmarshal(resp.Data, &item); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if isFav, ok := item["is_favorite"].(bool); !ok || !isFav {
		t.Errorf("expected is_favorite=true, got %v", item["is_favorite"])
	}
}

func TestBathhouseHandler_GetByID_IsFavorite_Unauthenticated(t *testing.T) {
	bhID := uuid.New()

	bhSvc := &mockBathhouseService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Bathhouse, error) {
			return &domain.Bathhouse{
				ID: id, Name: "Bath", Status: domain.BathhouseStatusActive,
				CityID: 1, PricePerHour: 5000, MaxGuests: 10, MinDuration: 1, Address: "Test",
			}, nil
		},
	}

	h := handler.NewBathhouseHandler(bhSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/bathhouses/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bathhouses/"+bhID.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse(t, rec)
	var item map[string]interface{}
	if err := json.Unmarshal(resp.Data, &item); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if isFav, ok := item["is_favorite"].(bool); !ok || isFav {
		t.Errorf("expected is_favorite=false, got %v", item["is_favorite"])
	}
}
