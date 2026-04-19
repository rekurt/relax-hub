package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRegionService struct {
	getRegionFn    func(ctx context.Context, userID uuid.UUID) (domain.UserRegion, error)
	switchRegionFn func(ctx context.Context, userID uuid.UUID, newRegion domain.UserRegion) error
}

func (m *mockRegionService) GetRegion(ctx context.Context, userID uuid.UUID) (domain.UserRegion, error) {
	if m.getRegionFn != nil {
		return m.getRegionFn(ctx, userID)
	}
	return domain.RegionRU, nil
}

func (m *mockRegionService) SwitchRegion(ctx context.Context, userID uuid.UUID, newRegion domain.UserRegion) error {
	if m.switchRegionFn != nil {
		return m.switchRegionFn(ctx, userID, newRegion)
	}
	return nil
}

func TestRegionHandler_GetRegion_Success(t *testing.T) {
	userID := uuid.New()

	svc := &mockRegionService{
		getRegionFn: func(_ context.Context, uid uuid.UUID) (domain.UserRegion, error) {
			assert.Equal(t, userID, uid)
			return domain.RegionRU, nil
		},
	}

	h := NewRegionHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/region", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetRegion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "RU", data["region"])
}

func TestRegionHandler_GetRegion_BY(t *testing.T) {
	userID := uuid.New()

	svc := &mockRegionService{
		getRegionFn: func(_ context.Context, _ uuid.UUID) (domain.UserRegion, error) {
			return domain.RegionBY, nil
		},
	}

	h := NewRegionHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/region", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetRegion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "BY", data["region"])
}

func TestRegionHandler_GetRegion_ServiceError(t *testing.T) {
	svc := &mockRegionService{
		getRegionFn: func(_ context.Context, _ uuid.UUID) (domain.UserRegion, error) {
			return "", domain.ErrNotFound
		},
	}

	h := NewRegionHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/region", nil)
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetRegion(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRegionHandler_SwitchRegion_Success(t *testing.T) {
	userID := uuid.New()
	called := false

	svc := &mockRegionService{
		switchRegionFn: func(_ context.Context, uid uuid.UUID, region domain.UserRegion) error {
			called = true
			assert.Equal(t, userID, uid)
			assert.Equal(t, domain.RegionBY, region)
			return nil
		},
	}

	h := NewRegionHandler(svc)
	body, _ := json.Marshal(map[string]string{"region": "BY"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/region", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), userID)
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.SwitchRegion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "BY", data["region"])
}

func TestRegionHandler_SwitchRegion_SameRegionBlocked(t *testing.T) {
	svc := &mockRegionService{
		switchRegionFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRegion) error {
			return domain.ErrRegionSameAsCurrent
		},
	}

	h := NewRegionHandler(svc)
	body, _ := json.Marshal(map[string]string{"region": "RU"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/region", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.SwitchRegion(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegionHandler_SwitchRegion_BlockedWithWalletBalance(t *testing.T) {
	svc := &mockRegionService{
		switchRegionFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRegion) error {
			return domain.ErrRegionSwitchBlocked
		},
	}

	h := NewRegionHandler(svc)
	body, _ := json.Marshal(map[string]string{"region": "BY"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/region", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.SwitchRegion(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRegionHandler_SwitchRegion_InvalidJSON(t *testing.T) {
	svc := &mockRegionService{}

	h := NewRegionHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/region", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDForTesting(req.Context(), uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.SwitchRegion(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
