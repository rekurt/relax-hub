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
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*domain.User, error)
	listFn    func(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	blockFn   func(ctx context.Context, id uuid.UUID) error
	unblockFn func(ctx context.Context, id uuid.UUID) error
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
	searchFn     func(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	createFn     func(ctx context.Context, ownerID uuid.UUID, input service.CreateBathhouseInput) (*domain.Bathhouse, error)
	updateFn     func(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input service.UpdateBathhouseInput) (*domain.Bathhouse, error)
	deleteFn     func(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error
	listByOwnerFn func(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	approveFn    func(ctx context.Context, id uuid.UUID) error
	rejectFn     func(ctx context.Context, id uuid.UUID) error
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
	createFn          func(ctx context.Context, userID uuid.UUID, input service.CreateReviewInput) (*domain.Review, error)
	listByBathhouseFn func(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
}

func (m *mockReviewService) Create(ctx context.Context, userID uuid.UUID, input service.CreateReviewInput) (*domain.Review, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockReviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	if m.listByBathhouseFn != nil {
		return m.listByBathhouseFn(ctx, bathhouseID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.Review]{}, nil
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
	getAllFn   func(ctx context.Context) ([]domain.City, error)
	getBySlugFn func(ctx context.Context, slug string) (*domain.City, error)
	createFn  func(ctx context.Context, input service.CreateCityInput) (*domain.City, error)
	updateFn  func(ctx context.Context, id int64, input service.UpdateCityInput) (*domain.City, error)
	deleteFn  func(ctx context.Context, id int64) error
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

// setupAuthRouter creates a chi router with auth middleware using our mock
func setupAuthRouter(authSvc middleware.AuthService, handler http.HandlerFunc, roles ...domain.UserRole) *chi.Mux {
	r := chi.NewRouter()
	if len(roles) > 0 {
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Post("/test", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Get("/test", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Patch("/test", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Delete("/test", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Put("/test", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Get("/test/{id}", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Patch("/test/{id}", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Delete("/test/{id}", handler)
		r.With(middleware.RequireAuth(authSvc), middleware.RequireRole(roles...)).Put("/test/{id}", handler)
	} else {
		r.With(middleware.RequireAuth(authSvc)).Post("/test", handler)
		r.With(middleware.RequireAuth(authSvc)).Get("/test", handler)
		r.With(middleware.RequireAuth(authSvc)).Patch("/test/{id}", handler)
	}
	return r
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

	h := handler.NewBathhouseHandler(bhSvc, nil, nil)

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

	h := handler.NewBathhouseHandler(bhSvc, nil, nil)

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

	h := handler.NewBathhouseHandler(bhSvc, nil, nil)

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
	h := handler.NewBathhouseHandler(bhSvc, nil, nil)

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
	h := handler.NewBathhouseHandler(nil, nil, nil)

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

// --- Unauthenticated access tests ---

func TestProtectedEndpoints_RequireAuth(t *testing.T) {
	authSvc := &mockAuthService{} // ParseToken returns error by default

	// Build a minimal router with protected routes
	router := chi.NewRouter()
	auth := middleware.RequireAuth(authSvc)

	bookingH := handler.NewBookingHandler(nil)
	bhH := handler.NewBathhouseHandler(nil, nil, nil)

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
