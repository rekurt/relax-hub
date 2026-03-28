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
	mockRepo "github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShareHandler_CreateShareLink(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	body := shareBookingRequest{
		BathhouseID: uuid.New().String(),
		StartTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		EndTime:     time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		GuestCount:  4,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/share", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	userID := uuid.New()
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp APIResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	data := resp.Data.(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.Contains(t, data["share_url"].(string), "https://bani.ru/share/booking/")
}

func TestShareHandler_CreateShareLink_WithBookingID(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	bookingID := uuid.New()
	body := shareBookingRequest{
		BathhouseID: uuid.New().String(),
		StartTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		EndTime:     time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		GuestCount:  2,
		BookingID:   bookingID.String(),
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/share", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestShareHandler_CreateShareLink_InvalidBathhouseID(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	body := shareBookingRequest{
		BathhouseID: "not-a-uuid",
		StartTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		EndTime:     time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		GuestCount:  2,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/share", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestShareHandler_ResolveShareLink(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	bathhouseID := uuid.New()
	startTime := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	endTime := time.Now().Add(26 * time.Hour).Truncate(time.Second)

	share := &domain.BookingShare{
		ID:          uuid.New(),
		Token:       "abc123def456",
		CreatedBy:   uuid.New(),
		BathhouseID: bathhouseID,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  3,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, 30),
	}
	repo.Create(context.Background(), share)

	r := chi.NewRouter()
	r.Get("/api/v1/share/booking/{token}", h.ResolveShareLink)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/booking/abc123def456", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp APIResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, bathhouseID.String(), data["bathhouse_id"])
	assert.Equal(t, float64(3), data["guest_count"])
}

func TestShareHandler_ResolveShareLink_NotFound(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	r := chi.NewRouter()
	r.Get("/api/v1/share/booking/{token}", h.ResolveShareLink)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/booking/nonexistent", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestShareHandler_ResolveShareLink_Expired(t *testing.T) {
	repo := mockRepo.NewBookingShareRepo()
	h := NewShareHandler(repo, "https://bani.ru")

	share := &domain.BookingShare{
		ID:          uuid.New(),
		Token:       "expired123",
		CreatedBy:   uuid.New(),
		BathhouseID: uuid.New(),
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(2 * time.Hour),
		GuestCount:  1,
		CreatedAt:   time.Now().AddDate(0, 0, -31),
		ExpiresAt:   time.Now().AddDate(0, 0, -1), // expired yesterday
	}
	repo.Create(context.Background(), share)

	r := chi.NewRouter()
	r.Get("/api/v1/share/booking/{token}", h.ResolveShareLink)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/booking/expired123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusGone, rr.Code)
}

func TestGenerateShareToken(t *testing.T) {
	token1, err := generateShareToken()
	require.NoError(t, err)
	assert.Len(t, token1, 32) // 16 bytes = 32 hex chars

	token2, err := generateShareToken()
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
}
