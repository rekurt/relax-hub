package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSavedCardService() (SavedCardService, *mock.SavedCardRepo) {
	repo := mock.NewSavedCardRepo()
	log := logger.New(logger.LevelError)
	svc := NewSavedCardService(repo, log)
	return svc, repo
}

func TestSavedCardService_CreateSavedCard(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	card, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_123abc",
		Last4:         "4242",
		Brand:         "visa",
		ExpiryMonth:   12,
		ExpiryYear:    2027,
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, card.ID)
	assert.Equal(t, userID, card.UserID)
	assert.Equal(t, "4242", card.Last4)
	assert.Equal(t, "visa", card.Brand)
	assert.True(t, card.IsDefault, "first card should be default")
}

func TestSavedCardService_CreateSavedCard_SecondNotDefault(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	_, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)

	card2, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_2",
		Last4:         "2222",
		Brand:         "mastercard",
		ExpiryMonth:   6,
		ExpiryYear:    2029,
	})
	require.NoError(t, err)
	assert.False(t, card2.IsDefault, "second card should not be default")
}

func TestSavedCardService_CreateSavedCard_ValidationError(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	_, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "",
		Last4:         "42",
		Brand:         "visa",
		ExpiryMonth:   12,
		ExpiryYear:    2027,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestSavedCardService_CreateSavedCard_LimitReached(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	for i := 0; i < maxSavedCardsPerUser; i++ {
		_, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
			ProviderToken: "tok_" + uuid.New().String(),
			Last4:         "4242",
			Brand:         "visa",
			ExpiryMonth:   12,
			ExpiryYear:    2027,
		})
		require.NoError(t, err)
	}

	_, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_overflow",
		Last4:         "4242",
		Brand:         "visa",
		ExpiryMonth:   12,
		ExpiryYear:    2027,
	})
	assert.ErrorIs(t, err, domain.ErrSavedCardLimitReached)
}

func TestSavedCardService_ListSavedCards(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	_, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)

	_, err = svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_2",
		Last4:         "2222",
		Brand:         "mastercard",
		ExpiryMonth:   6,
		ExpiryYear:    2029,
	})
	require.NoError(t, err)

	result, err := svc.ListSavedCards(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Items, 2)
	// Default card should be first
	assert.True(t, result.Items[0].IsDefault)
}

func TestSavedCardService_DeleteSavedCard(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	card, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)

	err = svc.DeleteSavedCard(context.Background(), userID, card.ID)
	require.NoError(t, err)

	result, err := svc.ListSavedCards(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.TotalCount)
}

func TestSavedCardService_DeleteSavedCard_Forbidden(t *testing.T) {
	svc, _ := newSavedCardService()
	ownerID := uuid.New()
	otherID := uuid.New()

	card, err := svc.CreateSavedCard(context.Background(), ownerID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)

	err = svc.DeleteSavedCard(context.Background(), otherID, card.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSavedCardService_DeleteSavedCard_NotFound(t *testing.T) {
	svc, _ := newSavedCardService()

	err := svc.DeleteSavedCard(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrSavedCardNotFound)
}

func TestSavedCardService_SetDefaultCard(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	card1, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)
	assert.True(t, card1.IsDefault)

	card2, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_2",
		Last4:         "2222",
		Brand:         "mastercard",
		ExpiryMonth:   6,
		ExpiryYear:    2029,
	})
	require.NoError(t, err)

	err = svc.SetDefaultCard(context.Background(), userID, card2.ID)
	require.NoError(t, err)

	// Verify card2 is now default
	updated, err := svc.GetSavedCard(context.Background(), userID, card2.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsDefault)

	// Verify card1 is no longer default
	updated1, err := svc.GetSavedCard(context.Background(), userID, card1.ID)
	require.NoError(t, err)
	assert.False(t, updated1.IsDefault)
}

func TestSavedCardService_SetDefaultCard_Forbidden(t *testing.T) {
	svc, _ := newSavedCardService()
	ownerID := uuid.New()
	otherID := uuid.New()

	card, err := svc.CreateSavedCard(context.Background(), ownerID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "1111",
		Brand:         "visa",
		ExpiryMonth:   1,
		ExpiryYear:    2028,
	})
	require.NoError(t, err)

	err = svc.SetDefaultCard(context.Background(), otherID, card.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSavedCardService_GetSavedCard(t *testing.T) {
	svc, _ := newSavedCardService()
	userID := uuid.New()

	card, err := svc.CreateSavedCard(context.Background(), userID, &domain.SavedCard{
		ProviderToken: "tok_1",
		Last4:         "4242",
		Brand:         "visa",
		ExpiryMonth:   12,
		ExpiryYear:    2027,
	})
	require.NoError(t, err)

	got, err := svc.GetSavedCard(context.Background(), userID, card.ID)
	require.NoError(t, err)
	assert.Equal(t, card.ID, got.ID)
	assert.Equal(t, "4242", got.Last4)
}

func TestSavedCardService_GetSavedCard_NotFound(t *testing.T) {
	svc, _ := newSavedCardService()

	_, err := svc.GetSavedCard(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrSavedCardNotFound)
}
