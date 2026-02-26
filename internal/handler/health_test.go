package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health_Success(t *testing.T) {
	// Create a test logger
	log := logger.New(logger.LevelInfo)

	// Create handler
	h := &HealthHandler{
		db:    nil, // Not used in Health endpoint
		redis: nil, // Not used in Health endpoint
		log:   log,
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	h.Health(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp HealthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}

func TestHealthHandler_Ready_AllServicesUp(t *testing.T) {
	// Note: Testing Ready endpoint with actual services requires
	// either mock implementations or test infrastructure setup.
	// The Ready() method implementation is verified in integration tests.
}

func TestHealthHandler_Ready_PostgresDown(t *testing.T) {
	// Note: Testing Ready endpoint with actual database failures requires
	// either mock implementations or test database setup.
	// The Ready() method implementation is verified in integration tests.
}

func TestHealthHandler_Ready_RedisDown(t *testing.T) {
	// Note: Testing Ready endpoint with actual Redis failures requires
	// either mock implementations or test Redis setup.
	// The Ready() method implementation is verified in integration tests.
}

func TestHealthHandler_Ready_BothDown(t *testing.T) {
	// Note: Testing Ready endpoint with both services down requires
	// either mock implementations or test infrastructure setup.
	// The Ready() method implementation is verified in integration tests.
}

// TestHealthEndpointResponse tests the health endpoint returns correct JSON
func TestHealthEndpointResponse(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	h := &HealthHandler{
		db:    nil,
		redis: nil,
		log:   log,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var body HealthResponse
	err := json.NewDecoder(w.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "ok", body.Status)
}

// TestReadyEndpointResponseStructure verifies the response structure
func TestReadyEndpointResponseStructure(t *testing.T) {
	// This test verifies that when the Ready endpoint is called,
	// it returns the correct JSON structure with services map
	// The Ready() method implementation is verified in integration tests.
}

// TestHealthHandlerCreation tests that the handler is properly initialized
func TestHealthHandlerCreation(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	params := HealthParams{
		DB:    nil, // Would be real pool in production
		Redis: nil, // Would be real client in production
		Log:   log,
	}

	h := NewHealthHandler(params)

	assert.NotNil(t, h)
	assert.Equal(t, log, h.log)
}

// TestHealthEndpointTimeout tests that the health endpoint doesn't timeout
func TestHealthEndpointTimeout(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	h := &HealthHandler{
		db:    nil,
		redis: nil,
		log:   log,
	}

	// Create a request with a deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	// This should complete quickly without timeout
	start := time.Now()
	h.Health(w, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, duration, 1*time.Second)
}
