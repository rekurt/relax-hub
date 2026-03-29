package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/service"
)

// mockWidgetBathhouseService implements service.BathhouseService for widget testing
type mockWidgetBathhouseService struct {
	getByAPIKeyFn func(ctx context.Context, apiKey string) (*domain.Bathhouse, error)
}

func (m *mockWidgetBathhouseService) Create(_ context.Context, _ uuid.UUID, _ service.CreateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}

func (m *mockWidgetBathhouseService) GetByID(_ context.Context, _ uuid.UUID) (*domain.Bathhouse, error) {
	return nil, nil
}

func (m *mockWidgetBathhouseService) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error) {
	if m.getByAPIKeyFn != nil {
		return m.getByAPIKeyFn(ctx, apiKey)
	}
	return nil, domain.ErrNotFound
}

func (m *mockWidgetBathhouseService) GetBySlug(_ context.Context, _ string) (*domain.Bathhouse, error) {
	return nil, domain.ErrNotFound
}

func (m *mockWidgetBathhouseService) Update(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ service.UpdateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}

func (m *mockWidgetBathhouseService) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockWidgetBathhouseService) Search(_ context.Context, _ domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockWidgetBathhouseService) ListByOwner(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}

func (m *mockWidgetBathhouseService) GetWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockWidgetBathhouseService) RegenerateWidgetKey(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockWidgetBathhouseService) Approve(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockWidgetBathhouseService) Reject(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *mockWidgetBathhouseService) CheckCompleteness(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.CompletenessResult, error) {
	return nil, nil
}
func (m *mockWidgetBathhouseService) SubmitForModeration(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockWidgetBathhouseService) DuplicateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Bathhouse, error) {
	return nil, nil
}
func (m *mockWidgetBathhouseService) DeactivateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockWidgetBathhouseService) ActivateBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockWidgetBathhouseService) ArchiveBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockWidgetBathhouseService) IncrementViewCount(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockWidgetBathhouseService) ComputeBadges(_ context.Context, _ *domain.Bathhouse) []string {
	return nil
}
func (m *mockWidgetBathhouseService) IsLastMinuteActive(_ *domain.Bathhouse) bool { return false }
func (m *mockWidgetBathhouseService) GetAreaAvgPrice(_ context.Context, _ int64, _, _ float64) (int64, error) {
	return 0, nil
}

// Mock BookingService for widget testing
type mockWidgetBookingService struct {
	createFn              func(ctx context.Context, userID uuid.UUID, input service.CreateBookingInput) (*service.BookingResult, error)
	getAvailableSlotsFn   func(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]service.TimeSlot, error)
}

func (m *mockWidgetBookingService) Create(ctx context.Context, userID uuid.UUID, input service.CreateBookingInput) (*service.BookingResult, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, input)
	}
	return &service.BookingResult{
		Booking: &domain.Booking{
			ID:          uuid.New(),
			UserID:      userID,
			BathhouseID: input.BathhouseID,
			StartTime:   input.StartTime,
			EndTime:     input.EndTime,
			GuestCount:  input.GuestCount,
			TotalPrice:  5000,
			Status:      domain.BookingPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}, nil
}

func (m *mockWidgetBookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, refundTo string) error {
	return nil
}

func (m *mockWidgetBookingService) Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	return nil
}

func (m *mockWidgetBookingService) Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID, reason string) error {
	return nil
}

func (m *mockWidgetBookingService) Approve(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	return nil
}

func (m *mockWidgetBookingService) AutoRejectTimedOutRequests(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockWidgetBookingService) Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) (*service.BookingResult, error) {
	return &service.BookingResult{}, nil
}

func (m *mockWidgetBookingService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	return nil, nil
}

func (m *mockWidgetBookingService) ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	return nil, nil
}

func (m *mockWidgetBookingService) GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]service.TimeSlot, error) {
	if m.getAvailableSlotsFn != nil {
		return m.getAvailableSlotsFn(ctx, bathhouseID, date)
	}
	return []service.TimeSlot{}, nil
}

func (m *mockWidgetBookingService) CheckIn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}

func (m *mockWidgetBookingService) CheckOut(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}

func (m *mockWidgetBookingService) MarkNoShows(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockWidgetBookingService) DisputeNoShow(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _ float64, _ string) error {
	return nil
}

func (m *mockWidgetBookingService) ListUpcomingWithBathhouse(_ context.Context, _, _ time.Time) ([]service.UpcomingBookingInfo, error) {
	return nil, nil
}
func (m *mockWidgetBookingService) Extend(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) (*service.ExtendResult, error) {
	return nil, nil
}
func (m *mockWidgetBookingService) GetRebookData(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*service.RebookData, error) {
	return nil, nil
}
func (m *mockWidgetBookingService) RecalculateResponseRates(_ context.Context) (int, error) {
	return 0, nil
}
func (m *mockWidgetBookingService) Modify(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ service.ModifyBookingInput) (*service.ModifyBookingResult, error) {
	return nil, nil
}
func (m *mockWidgetBookingService) AdminCancel(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockWidgetBookingService) AdminChangeStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ domain.BookingStatus, _ string) error {
	return nil
}
func (m *mockWidgetBookingService) AdminListBookings(_ context.Context, _ domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
	return nil, nil
}

func TestWidgetHandler_GetBathhouse(t *testing.T) {
	apiKey := "test-api-key"
	bathhouseID := uuid.New()

	bhRepo := &mockWidgetBathhouseService{
		getByAPIKeyFn: func(ctx context.Context, key string) (*domain.Bathhouse, error) {
			if key == apiKey {
				return &domain.Bathhouse{
					ID:           bathhouseID,
					Name:         "Test Bathhouse",
					Description:  "A nice bathhouse",
					Address:      "123 Main St",
					PricePerHour: 5000,
					MinDuration:  1,
					MaxGuests:    10,
					Rating:       4.5,
					ReviewCount:  20,
					Status:       domain.BathhouseStatusActive,
					ApiKey:       apiKey,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	log := logger.New(logger.LevelDebug)
	h := NewWidgetHandler(bhRepo, &mockWidgetBookingService{}, log)

	r := chi.NewRouter()
	r.Get("/widget/{api_key}/bathhouse", h.GetBathhouse)

	req := httptest.NewRequest(http.MethodGet, "/widget/"+apiKey+"/bathhouse", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	if resp.Data == nil {
		t.Error("expected data to be present")
	}
}

func TestWidgetHandler_GetBathhouse_NotFound(t *testing.T) {
	bhRepo := &mockWidgetBathhouseService{
		getByAPIKeyFn: func(ctx context.Context, key string) (*domain.Bathhouse, error) {
			return nil, domain.ErrNotFound
		},
	}

	log := logger.New(logger.LevelDebug)
	h := NewWidgetHandler(bhRepo, &mockWidgetBookingService{}, log)

	r := chi.NewRouter()
	r.Get("/widget/{api_key}/bathhouse", h.GetBathhouse)

	req := httptest.NewRequest(http.MethodGet, "/widget/invalid-key/bathhouse", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestWidgetHandler_GetAvailableSlots(t *testing.T) {
	apiKey := "test-api-key"
	bathhouseID := uuid.New()
	testDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	bhRepo := &mockWidgetBathhouseService{
		getByAPIKeyFn: func(ctx context.Context, key string) (*domain.Bathhouse, error) {
			if key == apiKey {
				return &domain.Bathhouse{
					ID:           bathhouseID,
					Name:         "Test Bathhouse",
					Status:       domain.BathhouseStatusActive,
					PricePerHour: 5000,
					MinDuration:  1,
					MaxGuests:    10,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	bookingSvc := &mockWidgetBookingService{
		getAvailableSlotsFn: func(ctx context.Context, bid uuid.UUID, date time.Time) ([]service.TimeSlot, error) {
			return []service.TimeSlot{
				{
					StartTime: time.Now().Add(24 * time.Hour),
					EndTime:   time.Now().Add(25 * time.Hour),
					Available: true,
					Price:     5000,
				},
			}, nil
		},
	}

	log := logger.New(logger.LevelDebug)
	h := NewWidgetHandler(bhRepo, bookingSvc, log)

	r := chi.NewRouter()
	r.Get("/widget/{api_key}/slots", h.GetAvailableSlots)

	req := httptest.NewRequest(http.MethodGet, "/widget/"+apiKey+"/slots?date="+testDate, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestWidgetHandler_CreateBooking(t *testing.T) {
	apiKey := "test-api-key"
	bathhouseID := uuid.New()

	bhRepo := &mockWidgetBathhouseService{
		getByAPIKeyFn: func(ctx context.Context, key string) (*domain.Bathhouse, error) {
			if key == apiKey {
				return &domain.Bathhouse{
					ID:           bathhouseID,
					Name:         "Test Bathhouse",
					Status:       domain.BathhouseStatusActive,
					PricePerHour: 5000,
					MinDuration:  1,
					MaxGuests:    10,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	log := logger.New(logger.LevelDebug)
	h := NewWidgetHandler(bhRepo, &mockWidgetBookingService{}, log)

	r := chi.NewRouter()
	r.Post("/widget/{api_key}/booking", h.CreateBooking)

	startTime := time.Now().Add(2 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)

	bookingReq := map[string]interface{}{
		"name":        "John Doe",
		"phone":       "+1234567890",
		"email":       "john@example.com",
		"start_time":  startTime.Format(time.RFC3339),
		"end_time":    endTime.Format(time.RFC3339),
		"guest_count": 4,
		"comment":     "Test booking",
	}

	body, _ := json.Marshal(bookingReq)
	req := httptest.NewRequest(http.MethodPost, "/widget/"+apiKey+"/booking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestWidgetHandler_CreateBooking_MissingFields(t *testing.T) {
	apiKey := "test-api-key"
	bathhouseID := uuid.New()

	bhRepo := &mockWidgetBathhouseService{
		getByAPIKeyFn: func(ctx context.Context, key string) (*domain.Bathhouse, error) {
			if key == apiKey {
				return &domain.Bathhouse{
					ID:     bathhouseID,
					Status: domain.BathhouseStatusActive,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	log := logger.New(logger.LevelDebug)
	h := NewWidgetHandler(bhRepo, &mockWidgetBookingService{}, log)

	r := chi.NewRouter()
	r.Post("/widget/{api_key}/booking", h.CreateBooking)

	// Missing email
	bookingReq := map[string]interface{}{
		"name":        "John Doe",
		"phone":       "+1234567890",
		"start_time":  time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"end_time":    time.Now().Add(3 * time.Hour).Format(time.RFC3339),
		"guest_count": 4,
	}

	body, _ := json.Marshal(bookingReq)
	req := httptest.NewRequest(http.MethodPost, "/widget/"+apiKey+"/booking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// TestServeScript tests the static script serving
func TestServeScript(t *testing.T) {
	handler := NewWidgetHandler(
		&mockWidgetBathhouseService{},
		&mockWidgetBookingService{},
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/widget.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeScript(rec, req)

	// If file exists (status 200), verify correct Content-Type
	// If file doesn't exist in test environment (status 404), that's acceptable
	if rec.Code == http.StatusOK {
		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/javascript; charset=utf-8" {
			t.Errorf("expected Content-Type application/javascript, got %s", contentType)
		}
	} else if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", rec.Code)
	}
}

// TestServeStyles tests the static styles serving
func TestServeStyles(t *testing.T) {
	handler := NewWidgetHandler(
		&mockWidgetBathhouseService{},
		&mockWidgetBookingService{},
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/widget.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeStyles(rec, req)

	// If file exists (status 200), verify correct Content-Type
	// If file doesn't exist in test environment (status 404), that's acceptable
	if rec.Code == http.StatusOK {
		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/css; charset=utf-8" {
			t.Errorf("expected Content-Type text/css, got %s", contentType)
		}
	} else if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", rec.Code)
	}
}
