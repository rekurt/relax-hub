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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBookingModificationService struct {
	requestFunc   func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input service.ModifyBookingInput) (*domain.BookingModificationRequest, error)
	approveFunc   func(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*service.ModifyBookingResult, error)
	rejectFunc    func(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error
	authorizeFunc func(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	listFunc      func(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error)
	getFunc       func(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error)
	expireFunc    func(ctx context.Context) (int, error)
}

func (m *mockBookingModificationService) RequestModification(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, input service.ModifyBookingInput) (*domain.BookingModificationRequest, error) {
	return m.requestFunc(ctx, userID, bookingID, input)
}

func (m *mockBookingModificationService) ApproveModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID) (*service.ModifyBookingResult, error) {
	return m.approveFunc(ctx, ownerID, role, requestID)
}

func (m *mockBookingModificationService) RejectModification(ctx context.Context, ownerID uuid.UUID, role domain.UserRole, requestID uuid.UUID, reason string) error {
	return m.rejectFunc(ctx, ownerID, role, requestID, reason)
}

func (m *mockBookingModificationService) AuthorizeListAccess(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	if m.authorizeFunc != nil {
		return m.authorizeFunc(ctx, userID, role, bookingID)
	}
	return nil
}

func (m *mockBookingModificationService) ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error) {
	return m.listFunc(ctx, bookingID)
}

func (m *mockBookingModificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error) {
	return m.getFunc(ctx, id)
}

func (m *mockBookingModificationService) ExpireTimedOutRequests(ctx context.Context) (int, error) {
	return m.expireFunc(ctx)
}

func TestBookingModificationHandler_RequestModification(t *testing.T) {
	now := time.Now()
	bookingID := uuid.New()
	requestID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	svc := &mockBookingModificationService{
		requestFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ service.ModifyBookingInput) (*domain.BookingModificationRequest, error) {
			return &domain.BookingModificationRequest{
				ID:                 requestID,
				BookingID:          bookingID,
				UserID:             clientID,
				BathhouseID:        bathhouseID,
				Status:             domain.ModReqPending,
				OldStartTime:       now,
				OldEndTime:         now.Add(2 * time.Hour),
				OldGuestCount:      3,
				OldTotalPrice:      10000,
				ProposedStartTime:  now.Add(24 * time.Hour),
				ProposedEndTime:    now.Add(26 * time.Hour),
				ProposedGuestCount: 4,
				ProposedTotalPrice: 15000,
				CreatedAt:          now,
				ExpiresAt:          now.Add(24 * time.Hour),
			}, nil
		},
	}

	h := NewBookingModificationHandler(svc)

	body, _ := json.Marshal(requestModificationRequest{
		StartTime:  now.Add(24 * time.Hour).Format(time.RFC3339),
		EndTime:    now.Add(26 * time.Hour).Format(time.RFC3339),
		GuestCount: 4,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/"+bookingID.String()+"/modification-request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), clientID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bookingID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.RequestModification(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp APIResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestBookingModificationHandler_ApproveModification(t *testing.T) {
	requestID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()
	ownerID := uuid.New()

	svc := &mockBookingModificationService{
		approveFunc: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*service.ModifyBookingResult, error) {
			return &service.ModifyBookingResult{
				Booking: &domain.Booking{
					ID:          bookingID,
					UserID:      uuid.New(),
					BathhouseID: uuid.New(),
					StartTime:   now.Add(24 * time.Hour),
					EndTime:     now.Add(26 * time.Hour),
					GuestCount:  4,
					TotalPrice:  15000,
					Status:      domain.BookingConfirmed,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				OldPrice:  10000,
				NewPrice:  15000,
				PriceDiff: 5000,
			}, nil
		},
	}

	h := NewBookingModificationHandler(svc)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/modification-requests/"+requestID.String()+"/approve", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), ownerID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", requestID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ApproveModification(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp APIResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestBookingModificationHandler_RejectModification(t *testing.T) {
	requestID := uuid.New()
	ownerID := uuid.New()

	svc := &mockBookingModificationService{
		rejectFunc: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) error {
			return nil
		},
	}

	h := NewBookingModificationHandler(svc)

	body, _ := json.Marshal(rejectModificationRequest{Reason: "Нет свободных мест"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/modification-requests/"+requestID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), ownerID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", requestID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.RejectModification(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBookingModificationHandler_ListModificationRequests(t *testing.T) {
	bookingID := uuid.New()
	now := time.Now()

	svc := &mockBookingModificationService{
		listFunc: func(_ context.Context, _ uuid.UUID) ([]domain.BookingModificationRequest, error) {
			return []domain.BookingModificationRequest{
				{
					ID:                 uuid.New(),
					BookingID:          bookingID,
					UserID:             uuid.New(),
					BathhouseID:        uuid.New(),
					Status:             domain.ModReqPending,
					OldStartTime:       now,
					OldEndTime:         now.Add(2 * time.Hour),
					OldGuestCount:      3,
					OldTotalPrice:      10000,
					ProposedStartTime:  now.Add(24 * time.Hour),
					ProposedEndTime:    now.Add(26 * time.Hour),
					ProposedGuestCount: 4,
					ProposedTotalPrice: 15000,
					CreatedAt:          now,
					ExpiresAt:          now.Add(24 * time.Hour),
				},
			}, nil
		},
	}

	h := NewBookingModificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+bookingID.String()+"/modification-requests", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bookingID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ListModificationRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp APIResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestBookingModificationHandler_RequestModification_InvalidBookingID(t *testing.T) {
	h := NewBookingModificationHandler(&mockBookingModificationService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/invalid/modification-request", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.RequestModification(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
