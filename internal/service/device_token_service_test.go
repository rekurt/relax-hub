package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeviceTokenTest(t *testing.T) (service.DeviceTokenService, *mock.DeviceTokenRepo) {
	t.Helper()
	repo := mock.NewDeviceTokenRepo().(*mock.DeviceTokenRepo)
	svc := service.NewDeviceTokenService(repo)
	return svc, repo
}

func TestDeviceToken_Register_Success(t *testing.T) {
	svc, _ := setupDeviceTokenTest(t)
	ctx := context.Background()

	token := &domain.DeviceToken{
		UserID:   uuid.New(),
		Token:    "fcm-token-abc123",
		Platform: "android",
	}
	err := svc.Register(ctx, token)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, token.ID)
}

func TestDeviceToken_Register_AllPlatforms(t *testing.T) {
	platforms := []string{"web", "android", "ios"}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			svc, _ := setupDeviceTokenTest(t)
			ctx := context.Background()

			token := &domain.DeviceToken{
				UserID:   uuid.New(),
				Token:    "token-" + platform,
				Platform: platform,
			}
			err := svc.Register(ctx, token)
			require.NoError(t, err)
		})
	}
}

func TestDeviceToken_Register_UpsertSameToken(t *testing.T) {
	svc, repo := setupDeviceTokenTest(t)
	ctx := context.Background()

	userID1 := uuid.New()
	userID2 := uuid.New()

	// Register token for user1
	token1 := &domain.DeviceToken{
		UserID:   userID1,
		Token:    "shared-token",
		Platform: "web",
	}
	require.NoError(t, svc.Register(ctx, token1))

	// Register same token string for user2 (device changed hands)
	token2 := &domain.DeviceToken{
		UserID:   userID2,
		Token:    "shared-token",
		Platform: "web",
	}
	require.NoError(t, svc.Register(ctx, token2))

	// The token should now belong to user2
	tokens, err := repo.ListByUser(ctx, userID2)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "shared-token", tokens[0].Token)
}

func TestDeviceToken_Delete_Success(t *testing.T) {
	svc, repo := setupDeviceTokenTest(t)
	ctx := context.Background()

	userID := uuid.New()
	token := &domain.DeviceToken{
		UserID:   userID,
		Token:    "token-to-delete",
		Platform: "ios",
	}
	require.NoError(t, svc.Register(ctx, token))

	// Verify it exists
	tokens, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)

	// Delete it
	err = svc.Delete(ctx, tokens[0].ID, userID)
	require.NoError(t, err)

	// Verify it's gone
	tokens, err = repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestDeviceToken_Delete_NotFound(t *testing.T) {
	svc, _ := setupDeviceTokenTest(t)
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeviceToken_Delete_WrongUser(t *testing.T) {
	svc, repo := setupDeviceTokenTest(t)
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()

	token := &domain.DeviceToken{
		UserID:   ownerID,
		Token:    "owner-token",
		Platform: "android",
	}
	require.NoError(t, svc.Register(ctx, token))

	tokens, err := repo.ListByUser(ctx, ownerID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)

	// Another user tries to delete -> should fail
	err = svc.Delete(ctx, tokens[0].ID, otherID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Token should still exist
	tokens, err = repo.ListByUser(ctx, ownerID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
}

func TestDeviceToken_Register_MultipleDevices(t *testing.T) {
	svc, repo := setupDeviceTokenTest(t)
	ctx := context.Background()

	userID := uuid.New()

	// Register multiple tokens for the same user
	for i, platform := range []string{"web", "android", "ios"} {
		token := &domain.DeviceToken{
			UserID:   userID,
			Token:    "token-" + platform + "-" + string(rune('0'+i)),
			Platform: platform,
		}
		require.NoError(t, svc.Register(ctx, token))
	}

	tokens, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, tokens, 3)
}
