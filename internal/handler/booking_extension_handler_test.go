package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBookingExtensionService struct {
	requestFn       func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*domain.BookingExtensionRequest, error)
	approveFn       func(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*service.ExtendResult, error)
	rejectFn        func(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error
	authorizeFn     func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	listByBookingFn func(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error)
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error)
	expireFn        func(ctx context.Context) (int, error)
}

func (m *mockBookingExtensionService) RequestExtension(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*domain.BookingExtensionRequest, error) {
	if m.requestFn != nil {
		return m.requestFn(ctx, userID, bookingID, extraHours)
	}
	return nil, nil
}

func (m *mockBookingExtensionService) ApproveExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*service.ExtendResult, error) {
	if m.approveFn != nil {
		return m.approveFn(ctx, ownerID, role, requestID)
	}
	return nil, nil
}

func (m *mockBookingExtensionService) RejectExtension(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error {
	if m.rejectFn != nil {
		return m.rejectFn(ctx, ownerID, role, requestID, reason)
	}
	return nil
}

func (m *mockBookingExtensionService) AuthorizeListAccess(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.authorizeFn != nil {
		return m.authorizeFn(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingExtensionService) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error) {
	if m.listByBookingFn != nil {
		return m.listByBookingFn(ctx, bookingID)
	}
	return nil, nil
}

func (m *mockBookingExtensionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockBookingExtensionService) ExpireTimedOutRequests(ctx context.Context) (int, error) {
	if m.expireFn != nil {
		return m.expireFn(ctx)
	}
	return 0, nil
}

func TestBookingExtensionHandler_RequestExtension_Success(t *testing.T) {
	bookingID := uuid.New()
	requestID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()
	now := time.Now()

	svc := &mockBookingExtensionService{
		requestFn: func(_ context.Context, userID uuid.UUID, bid uuid.UUID, extraHours int) (*domain.BookingExtensionRequest, error) {
			assert.Equal(t, clientID, userID)
			assert.Equal(t, bookingID, bid)
			assert.Equal(t, 2, extraHours)
			return &domain.BookingExtensionRequest{
				ID:             requestID,
				BookingID:      bookingID,
				UserID:         clientID,
				BathhouseID:    bathhouseID,
				Status:         domain.ExtReqPending,
				ExtraHours:     extraHours,
				ExtensionPrice: 10000,
				NewEndTime:     now.Add(4 * time.Hour),
				CreatedAt:      now,
				ExpiresAt:      now.Add(30 * time.Minute),
			}, nil
		},
	}

	h := NewBookingExtensionHandler(svc)
	body, _ := json.Marshal(map[string]int{"extra_hours": 2})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), clientID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.RequestExtension(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, requestID.String(), data["id"])
	assert.Equal(t, bookingID.String(), data["booking_id"])
	assert.Equal(t, "pending", data["status"])
}

func TestBookingExtensionHandler_RequestExtension_InvalidBookingID(t *testing.T) {
	svc := &mockBookingExtensionService{}
	h := NewBookingExtensionHandler(svc)

	body, _ := json.Marshal(map[string]int{"extra_hours": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.RequestExtension(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingExtensionHandler_RequestExtension_AlreadyPending(t *testing.T) {
	bookingID := uuid.New()
	svc := &mockBookingExtensionService{
		requestFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) (*domain.BookingExtensionRequest, error) {
			return nil, domain.ErrExtensionRequestPending
		},
	}

	h := NewBookingExtensionHandler(svc)
	body, _ := json.Marshal(map[string]int{"extra_hours": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.RequestExtension(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestBookingExtensionHandler_ApproveExtension_Success(t *testing.T) {
	requestID := uuid.New()
	bookingID := uuid.New()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	now := time.Now()

	svc := &mockBookingExtensionService{
		approveFn: func(_ context.Context, uid uuid.UUID, role domain.UserRole, rid uuid.UUID) (*service.ExtendResult, error) {
			assert.Equal(t, ownerID, uid)
			assert.Equal(t, domain.RoleOwner, role)
			assert.Equal(t, requestID, rid)
			return &service.ExtendResult{
				Booking: &domain.Booking{
					ID:          bookingID,
					BathhouseID: bathhouseID,
					StartTime:   now,
					EndTime:     now.Add(4 * time.Hour),
					Status:      domain.BookingConfirmed,
					TotalPrice:  20000,
				},
				ExtensionPrice: 10000,
			}, nil
		},
	}

	h := NewBookingExtensionHandler(svc)
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), ownerID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", requestID.String())

	rr := httptest.NewRecorder()
	h.ApproveExtension(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestBookingExtensionHandler_ApproveExtension_InvalidID(t *testing.T) {
	svc := &mockBookingExtensionService{}
	h := NewBookingExtensionHandler(svc)

	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.ApproveExtension(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingExtensionHandler_ApproveExtension_NotFound(t *testing.T) {
	requestID := uuid.New()
	svc := &mockBookingExtensionService{
		approveFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.ExtendResult, error) {
			return nil, domain.ErrExtensionRequestNotFound
		},
	}

	h := NewBookingExtensionHandler(svc)
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", requestID.String())

	rr := httptest.NewRecorder()
	h.ApproveExtension(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBookingExtensionHandler_ApproveExtension_Expired(t *testing.T) {
	requestID := uuid.New()
	svc := &mockBookingExtensionService{
		approveFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.ExtendResult, error) {
			return nil, domain.ErrExtensionRequestExpired
		},
	}

	h := NewBookingExtensionHandler(svc)
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", requestID.String())

	rr := httptest.NewRecorder()
	h.ApproveExtension(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingExtensionHandler_RejectExtension_Success(t *testing.T) {
	requestID := uuid.New()
	ownerID := uuid.New()
	called := false

	svc := &mockBookingExtensionService{
		rejectFn: func(_ context.Context, uid uuid.UUID, role domain.UserRole, rid uuid.UUID, reason string) error {
			called = true
			assert.Equal(t, ownerID, uid)
			assert.Equal(t, domain.RoleOwner, role)
			assert.Equal(t, requestID, rid)
			assert.Equal(t, "time slot not available", reason)
			return nil
		},
	}

	h := NewBookingExtensionHandler(svc)
	body, _ := json.Marshal(map[string]string{"reason": "time slot not available"})
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), ownerID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", requestID.String())

	rr := httptest.NewRecorder()
	h.RejectExtension(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)
}

func TestBookingExtensionHandler_RejectExtension_InvalidID(t *testing.T) {
	svc := &mockBookingExtensionService{}
	h := NewBookingExtensionHandler(svc)

	body, _ := json.Marshal(map[string]string{"reason": "no"})
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.RejectExtension(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBookingExtensionHandler_ListExtensionRequests_Success(t *testing.T) {
	bookingID := uuid.New()
	requestID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()
	now := time.Now()

	svc := &mockBookingExtensionService{
		authorizeFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, bid uuid.UUID) error {
			assert.Equal(t, bookingID, bid)
			return nil
		},
		listByBookingFn: func(_ context.Context, bid uuid.UUID) ([]domain.BookingExtensionRequest, error) {
			assert.Equal(t, bookingID, bid)
			return []domain.BookingExtensionRequest{
				{
					ID:          requestID,
					BookingID:   bookingID,
					UserID:      clientID,
					BathhouseID: bathhouseID,
					Status:      domain.ExtReqApproved,
					ExtraHours:  1,
					NewEndTime:  now.Add(3 * time.Hour),
					CreatedAt:   now,
					ExpiresAt:   now.Add(30 * time.Minute),
				},
			}, nil
		},
	}

	h := NewBookingExtensionHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), clientID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.ListExtensionRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, requestID.String(), first["id"])
	assert.Equal(t, "approved", first["status"])
}

func TestBookingExtensionHandler_ListExtensionRequests_Forbidden(t *testing.T) {
	bookingID := uuid.New()
	svc := &mockBookingExtensionService{
		authorizeFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}

	h := NewBookingExtensionHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", bookingID.String())

	rr := httptest.NewRecorder()
	h.ListExtensionRequests(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestBookingExtensionHandler_ListExtensionRequests_InvalidID(t *testing.T) {
	svc := &mockBookingExtensionService{}
	h := NewBookingExtensionHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)
	req = chiCtxWithID(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.ListExtensionRequests(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
