package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavedSearchService_SaveSearch(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	userID := uuid.New()
	filters := json.RawMessage(`{"city_id":1,"has_pool":true}`)

	search, err := svc.SaveSearch(context.Background(), userID, "Москва с бассейном", filters, true)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, search.ID)
	assert.Equal(t, userID, search.UserID)
	assert.Equal(t, "Москва с бассейном", search.Name)
	assert.True(t, search.NotifyOnNew)
	assert.JSONEq(t, `{"city_id":1,"has_pool":true}`, string(search.Filters))
}

func TestSavedSearchService_ListSavedSearches(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	userID := uuid.New()
	filters := json.RawMessage(`{"city_id":1}`)

	_, err := svc.SaveSearch(context.Background(), userID, "Search 1", filters, false)
	require.NoError(t, err)
	_, err = svc.SaveSearch(context.Background(), userID, "Search 2", filters, true)
	require.NoError(t, err)

	result, err := svc.ListSavedSearches(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Items, 2)
}

func TestSavedSearchService_DeleteSavedSearch(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	userID := uuid.New()
	filters := json.RawMessage(`{"city_id":1}`)

	search, err := svc.SaveSearch(context.Background(), userID, "Test", filters, false)
	require.NoError(t, err)

	err = svc.DeleteSavedSearch(context.Background(), userID, search.ID)
	require.NoError(t, err)

	result, err := svc.ListSavedSearches(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.TotalCount)
}

func TestSavedSearchService_DeleteSavedSearch_Forbidden(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	filters := json.RawMessage(`{"city_id":1}`)

	search, err := svc.SaveSearch(context.Background(), ownerID, "Test", filters, false)
	require.NoError(t, err)

	err = svc.DeleteSavedSearch(context.Background(), otherUserID, search.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSavedSearchService_DeleteSavedSearch_NotFound(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	err := svc.DeleteSavedSearch(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrSavedSearchNotFound)
}

func TestSavedSearchService_ListWithNotifications(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	userID := uuid.New()
	filters := json.RawMessage(`{"city_id":1}`)

	_, err := svc.SaveSearch(context.Background(), userID, "With notif", filters, true)
	require.NoError(t, err)
	_, err = svc.SaveSearch(context.Background(), userID, "Without notif", filters, false)
	require.NoError(t, err)

	items, err := svc.ListWithNotifications(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "With notif", items[0].Name)
}

func TestSavedSearchService_SaveSearchLimit(t *testing.T) {
	repo := mock.NewSavedSearchRepo()
	bhRepo := mock.NewBathhouseRepo()
	log := logger.New(logger.LevelError)

	svc := NewSavedSearchService(repo, bhRepo, nil, nil, log)

	userID := uuid.New()
	filters := json.RawMessage(`{"city_id":1}`)

	for i := 0; i < savedSearchMaxPerUser; i++ {
		_, err := svc.SaveSearch(context.Background(), userID, "Search", filters, false)
		require.NoError(t, err)
	}

	_, err := svc.SaveSearch(context.Background(), userID, "One too many", filters, false)
	assert.ErrorIs(t, err, domain.ErrSavedSearchLimitReached)
}
