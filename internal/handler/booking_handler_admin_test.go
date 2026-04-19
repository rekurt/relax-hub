package handler

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
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bookingServiceAdminMock is a full implementation of service.BookingService with
// configurable admin function overrides for handler-level unit tests.
type bookingServiceAdminMock struct {
	adminCancelFn       func(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, reason string) error
	adminChangeStatusFn func(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, status domain.BookingStatus, reason string) error
	adminListFn         func(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error)
}

// Required interface methods — no-op stubs for methods not under test.
func (m *bookingServiceAdminMock) Create(_ context.Context, _ uuid.UUID, _ service.CreateBookingInput) (*service.BookingResult, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) Modify(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ service.ModifyBookingInput) (*service.ModifyBookingResult, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) Cancel(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) error {
	return nil
}
func (m *bookingServiceAdminMock) Confirm(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *bookingServiceAdminMock) Reject(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) error {
	return nil
}
func (m *bookingServiceAdminMock) Approve(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *bookingServiceAdminMock) Complete(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.BookingResult, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) GetByID(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Booking, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Booking], error) {
	return &domain.PaginatedResult[domain.Booking]{}, nil
}
func (m *bookingServiceAdminMock) ListByBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Booking], error) {
	return &domain.PaginatedResult[domain.Booking]{}, nil
}
func (m *bookingServiceAdminMock) GetAvailableSlots(_ context.Context, _ uuid.UUID, _ time.Time) ([]service.TimeSlot, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) AutoRejectTimedOutRequests(_ context.Context) (int, error) {
	return 0, nil
}
func (m *bookingServiceAdminMock) CheckIn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *bookingServiceAdminMock) CheckOut(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *bookingServiceAdminMock) MarkNoShows(_ context.Context) (int, error) {
	return 0, nil
}
func (m *bookingServiceAdminMock) DisputeNoShow(_ context.Context, _ uuid.UUID, _ uuid.UUID, _, _ float64, _ string) error {
	return nil
}
func (m *bookingServiceAdminMock) ListUpcomingWithBathhouse(_ context.Context, _, _ time.Time) ([]service.UpcomingBookingInfo, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) Extend(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) (*service.ExtendResult, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) GetRebookData(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*service.RebookData, error) {
	return nil, nil
}
func (m *bookingServiceAdminMock) RecalculateResponseRates(_ context.Context) (int, error) {
	return 0, nil
}

func (m *bookingServiceAdminMock) AdminCancel(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, reason string) error {
	if m.adminCancelFn != nil {
		return m.adminCancelFn(ctx, adminID, bookingID, reason)
	}
	return nil
}

func (m *bookingServiceAdminMock) AdminChangeStatus(ctx context.Context, adminID uuid.UUID, bookingID uuid.UUID, status domain.BookingStatus, reason string) error {
	if m.adminChangeStatusFn != nil {
		return m.adminChangeStatusFn(ctx, adminID, bookingID, status, reason)
	}
	return nil
}

func (m *bookingServiceAdminMock) AdminListBookings(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
	if m.adminListFn != nil {
		return m.adminListFn(ctx, filter)
	}
	return &domain.PaginatedResult[domain.Booking]{
		Items:    []domain.Booking{},
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func chiCtxWithID(req *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestBookingHandler_AdminCancel_Success(t *testing.T) {
	bookingID := uuid.New()
	adminID := uuid.New()
	called := false

	svc := &bookingServiceAdminMock{
		adminCancelFn: func(_ context.Context, aid uuid.UUID, bid uuid.UUID, reason string) error {
			called = true
			assert.Equal(t, adminID, aid)
			assert.Equal(t, bookingID, bid)
			assert.Equal(t, "admin cancellation test", reason)
			return nil
		},
	}

	h := NewBookingHandler(svc, nil)
	body, _ := json.Marshal(map[string]string{"reason": "admin cancellation test"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), adminID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminCancel(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)
}

func TestBookingHandler_AdminCancel_MissingReason(t *testing.T) {
	bookingID := uuid.New()
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{"reason": ""})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminCancel(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminCancel_InvalidID(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{"reason": "test"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.AdminCancel(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminCancel_NotFound(t *testing.T) {
	bookingID := uuid.New()
	svc := &bookingServiceAdminMock{
		adminCancelFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
			return domain.ErrNotFound
		},
	}

	h := NewBookingHandler(svc, nil)
	body, _ := json.Marshal(map[string]string{"reason": "admin cancel"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminCancel(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBookingHandler_AdminChangeStatus_Success(t *testing.T) {
	bookingID := uuid.New()
	adminID := uuid.New()
	called := false

	svc := &bookingServiceAdminMock{
		adminChangeStatusFn: func(_ context.Context, aid uuid.UUID, bid uuid.UUID, status domain.BookingStatus, reason string) error {
			called = true
			assert.Equal(t, adminID, aid)
			assert.Equal(t, bookingID, bid)
			assert.Equal(t, domain.BookingCancelled, status)
			assert.Equal(t, "force close", reason)
			return nil
		},
	}

	h := NewBookingHandler(svc, nil)
	body, _ := json.Marshal(map[string]string{"status": "cancelled", "reason": "force close"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), adminID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminChangeStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)
}

func TestBookingHandler_AdminChangeStatus_InvalidStatus(t *testing.T) {
	bookingID := uuid.New()
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{"status": "invalid_status", "reason": "test"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminChangeStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminChangeStatus_MissingReason(t *testing.T) {
	bookingID := uuid.New()
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{"status": "confirmed", "reason": ""})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.AdminChangeStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminChangeStatus_InvalidID(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{"status": "confirmed", "reason": "test"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.AdminChangeStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminListBookings_DefaultPagination(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	svc := &bookingServiceAdminMock{
		adminListFn: func(_ context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 20, filter.PageSize)
			return &domain.PaginatedResult[domain.Booking]{
				Items: []domain.Booking{
					{
						ID:        bookingID,
						StartTime: now,
						EndTime:   now.Add(2 * time.Hour),
						Status:    domain.BookingConfirmed,
					},
				},
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewBookingHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/bookings", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestBookingHandler_AdminListBookings_WithFilters(t *testing.T) {
	userID := uuid.New()
	bathhouseID := uuid.New()
	fromDate := time.Now().UTC().Add(-24 * time.Hour)
	toDate := time.Now().UTC()

	svc := &bookingServiceAdminMock{
		adminListFn: func(_ context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error) {
			require.NotNil(t, filter.UserID)
			assert.Equal(t, userID, *filter.UserID)
			require.NotNil(t, filter.BathhouseID)
			assert.Equal(t, bathhouseID, *filter.BathhouseID)
			require.NotNil(t, filter.Status)
			assert.Equal(t, domain.BookingConfirmed, *filter.Status)
			require.NotNil(t, filter.FromDate)
			require.NotNil(t, filter.ToDate)
			return &domain.PaginatedResult[domain.Booking]{Items: []domain.Booking{}, Page: 1, PageSize: 20}, nil
		},
	}

	h := NewBookingHandler(svc, nil)
	url := "/?user_id=" + userID.String() +
		"&bathhouse_id=" + bathhouseID.String() +
		"&status=confirmed" +
		"&from_date=" + fromDate.Format(time.RFC3339) +
		"&to_date=" + toDate.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBookingHandler_AdminListBookings_InvalidUserID(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/?user_id=not-a-uuid", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminListBookings_InvalidFromDate(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/?from_date=not-a-date", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminListBookings_InvalidToDate(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/?to_date=not-a-date", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminListBookings_InvalidBathhouseID(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/?bathhouse_id=not-a-uuid", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingHandler_AdminListBookings_InvalidStatus(t *testing.T) {
	svc := &bookingServiceAdminMock{}
	h := NewBookingHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/?status=invalid_status", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.AdminListBookings(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
