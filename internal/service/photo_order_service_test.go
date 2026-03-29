package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPhotoOrderTest(t *testing.T) (service.PhotoOrderService, *mockWalletService) {
	t.Helper()

	repo := mock.NewPhotoOrderRepo()
	walletSvc := newMockWalletService()
	log := logger.New(logger.LevelWarn)
	checker := service.NewAccessChecker(mock.NewRepresentativeRepo(), mock.NewBathhouseRepo())

	svc := service.NewPhotoOrderService(repo, walletSvc, checker, log)
	return svc, walletSvc
}

func TestPhotoOrderCreate(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: bathhouseID,
		Region:      "RU",
		Notes:       "Снять парную и бассейн",
	}

	err := svc.Create(ctx, ownerID, order)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, order.ID)
	assert.Equal(t, ownerID, order.OwnerID)
	assert.Equal(t, domain.PhotoOrderStatusRequested, order.Status)
}

func TestPhotoOrderCreate_Validation(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	// Missing bathhouse ID
	order := &domain.PhotoOrder{
		Region: "RU",
	}
	err := svc.Create(ctx, ownerID, order)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)

	// Missing region
	order = &domain.PhotoOrder{
		BathhouseID: uuid.New(),
	}
	err = svc.Create(ctx, ownerID, order)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestPhotoOrderGetByID_AccessControl(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	bathhouseID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: bathhouseID,
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// Owner can access
	got, err := svc.GetByID(ctx, ownerID, domain.RoleOwner, order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ID, got.ID)

	// Admin can access
	got, err = svc.GetByID(ctx, uuid.New(), domain.RoleAdmin, order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ID, got.ID)

	// Other user cannot access
	_, err = svc.GetByID(ctx, otherUserID, domain.RoleOwner, order.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestPhotoOrderOwnerCancel(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// Owner can cancel a requested order
	err := svc.OwnerCancel(ctx, ownerID, order.ID)
	require.NoError(t, err)

	got, err := svc.GetByID(ctx, ownerID, domain.RoleOwner, order.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PhotoOrderStatusCancelled, got.Status)
}

func TestPhotoOrderOwnerCancel_WrongOwner(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	err := svc.OwnerCancel(ctx, uuid.New(), order.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestPhotoOrderOwnerCancel_WrongStatus(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// First confirm the order via admin
	confirmedStatus := domain.PhotoOrderStatusConfirmed
	_, err := svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status: &confirmedStatus,
	})
	require.NoError(t, err)

	// Owner cannot cancel a confirmed order
	err = svc.OwnerCancel(ctx, ownerID, order.ID)
	assert.ErrorIs(t, err, domain.ErrPhotoOrderInvalidStatus)
}

func TestPhotoOrderAdminUpdate_StatusTransitions(t *testing.T) {
	svc, walletSvc := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	// Create wallet for owner
	walletSvc.wallets[uuid.New()] = &domain.Wallet{
		ID:      uuid.New(),
		UserID:  ownerID,
		Balance: 1000000, // 10000 rubles
	}

	// Update wallets map to use correct wallet ID
	var walletID uuid.UUID
	for id, w := range walletSvc.wallets {
		if w.UserID == ownerID {
			walletID = id
			w.ID = walletID
			break
		}
	}

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// requested -> confirmed
	confirmedStatus := domain.PhotoOrderStatusConfirmed
	photographer := "Иван Иванов"
	price := int64(500000) // 5000 rubles
	updated, err := svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status:           &confirmedStatus,
		PhotographerName: &photographer,
		Price:            &price,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.PhotoOrderStatusConfirmed, updated.Status)
	assert.Equal(t, "Иван Иванов", updated.PhotographerName)
	assert.Equal(t, int64(500000), updated.Price)

	// confirmed -> completed (should charge wallet)
	completedStatus := domain.PhotoOrderStatusCompleted
	updated, err = svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status: &completedStatus,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.PhotoOrderStatusCompleted, updated.Status)
}

func TestPhotoOrderAdminUpdate_InvalidTransition(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// requested -> completed (invalid, must go through confirmed)
	completedStatus := domain.PhotoOrderStatusCompleted
	_, err := svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status: &completedStatus,
	})
	assert.ErrorIs(t, err, domain.ErrPhotoOrderInvalidStatus)
}

func TestPhotoOrderAdminUpdate_CancelFromConfirmed(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()

	order := &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "RU",
	}
	require.NoError(t, svc.Create(ctx, ownerID, order))

	// requested -> confirmed
	confirmedStatus := domain.PhotoOrderStatusConfirmed
	_, err := svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status: &confirmedStatus,
	})
	require.NoError(t, err)

	// confirmed -> cancelled (admin can cancel)
	cancelledStatus := domain.PhotoOrderStatusCancelled
	updated, err := svc.AdminUpdate(ctx, order.ID, &service.PhotoOrderUpdate{
		Status: &cancelledStatus,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.PhotoOrderStatusCancelled, updated.Status)
}

func TestPhotoOrderListByOwner(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()

	// Create orders for both owners
	for i := 0; i < 3; i++ {
		require.NoError(t, svc.Create(ctx, ownerID, &domain.PhotoOrder{
			BathhouseID: uuid.New(),
			Region:      "RU",
		}))
	}
	require.NoError(t, svc.Create(ctx, otherOwnerID, &domain.PhotoOrder{
		BathhouseID: uuid.New(),
		Region:      "BY",
	}))

	result, err := svc.ListByOwner(ctx, ownerID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.Len(t, result.Items, 3)
}

func TestPhotoOrderAdminList(t *testing.T) {
	svc, _ := setupPhotoOrderTest(t)
	ctx := context.Background()

	// Create some orders
	require.NoError(t, svc.Create(ctx, uuid.New(), &domain.PhotoOrder{
		BathhouseID: uuid.New(), Region: "RU",
	}))
	require.NoError(t, svc.Create(ctx, uuid.New(), &domain.PhotoOrder{
		BathhouseID: uuid.New(), Region: "BY",
	}))

	// List all
	result, err := svc.AdminList(ctx, domain.PhotoOrderFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.TotalCount)

	// Filter by region
	region := "RU"
	result, err = svc.AdminList(ctx, domain.PhotoOrderFilter{Region: &region, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.TotalCount)
}
