package cron

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWalletCronScheduler(notifSvc *mockNotificationService, walletSvc service.WalletService, walletRepo *mock.WalletRepo) *CronScheduler {
	log := logger.New(logger.LevelInfo)
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()
	return NewCronScheduler(&config.Config{}, log, mockAnalyticsSvc, mockAnalyticsRepo, nil, nil, notifSvc, nil, walletSvc, walletRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func TestHandleBonusExpiration_NoWallets(t *testing.T) {
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	notifSvc := &mockNotificationService{}
	walletSvc := service.NewWalletService(walletRepo, logger.New(logger.LevelInfo))

	cs := newTestWalletCronScheduler(notifSvc, walletSvc, walletRepo)
	_ = cs.bonusExpiration(context.Background())

	assert.Empty(t, notifSvc.sent)
}

func TestHandleBonusExpiration_ExpiresOldBonuses(t *testing.T) {
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	notifSvc := &mockNotificationService{}
	walletSvc := service.NewWalletService(walletRepo, logger.New(logger.LevelInfo))

	userID := uuid.New()
	walletID := uuid.New()

	err := walletRepo.Create(context.Background(), &domain.Wallet{
		ID:       walletID,
		UserID:   userID,
		Balance:  50000,
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	})
	require.NoError(t, err)

	// Create an expired bonus transaction
	expiredAt := time.Now().Add(-1 * time.Hour)
	err = walletRepo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         domain.WalletTxWelcomeBonus,
		Amount:       50000,
		BalanceAfter: 50000,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Приветственный бонус",
		IsBonus:      true,
		ExpiresAt:    &expiredAt,
	})
	require.NoError(t, err)

	cs := newTestWalletCronScheduler(notifSvc, walletSvc, walletRepo)
	_ = cs.bonusExpiration(context.Background())

	// Should have sent a notification about expired bonuses
	assert.Equal(t, 1, len(notifSvc.sent))
	assert.Equal(t, userID, notifSvc.sent[0].UserID)
	assert.Equal(t, domain.NotifBonusExpired, notifSvc.sent[0].NotifType)

	// Balance should be deducted
	wallet, err := walletRepo.GetByID(context.Background(), walletID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), wallet.Balance)
}

func TestHandleBonusExpiration_SkipsActiveBonuses(t *testing.T) {
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	notifSvc := &mockNotificationService{}
	walletSvc := service.NewWalletService(walletRepo, logger.New(logger.LevelInfo))

	userID := uuid.New()
	walletID := uuid.New()

	err := walletRepo.Create(context.Background(), &domain.Wallet{
		ID:       walletID,
		UserID:   userID,
		Balance:  50000,
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	})
	require.NoError(t, err)

	// Create a non-expired bonus transaction (future expiry)
	futureExpiry := time.Now().Add(30 * 24 * time.Hour)
	err = walletRepo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         domain.WalletTxWelcomeBonus,
		Amount:       50000,
		BalanceAfter: 50000,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Приветственный бонус",
		IsBonus:      true,
		ExpiresAt:    &futureExpiry,
	})
	require.NoError(t, err)

	cs := newTestWalletCronScheduler(notifSvc, walletSvc, walletRepo)
	_ = cs.bonusExpiration(context.Background())

	// Should NOT send notification - bonuses are still active
	assert.Empty(t, notifSvc.sent)

	// Balance should remain unchanged
	wallet, err := walletRepo.GetByID(context.Background(), walletID)
	require.NoError(t, err)
	assert.Equal(t, int64(50000), wallet.Balance)
}

func TestHandleBonusExpiryNotify_NotifiesAt14And3Days(t *testing.T) {
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	notifSvc := &mockNotificationService{}
	walletSvc := service.NewWalletService(walletRepo, logger.New(logger.LevelInfo))

	userID := uuid.New()
	walletID := uuid.New()

	err := walletRepo.Create(context.Background(), &domain.Wallet{
		ID:       walletID,
		UserID:   userID,
		Balance:  50000,
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	})
	require.NoError(t, err)

	// Create a bonus expiring in ~3 days (within the 3-day notification window)
	expiresIn3Days := time.Now().Add(3 * 24 * time.Hour).Add(-1 * time.Hour)
	err = walletRepo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         domain.WalletTxWelcomeBonus,
		Amount:       50000,
		BalanceAfter: 50000,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Приветственный бонус",
		IsBonus:      true,
		ExpiresAt:    &expiresIn3Days,
	})
	require.NoError(t, err)

	cs := newTestWalletCronScheduler(notifSvc, walletSvc, walletRepo)
	_ = cs.bonusExpiryNotify(context.Background())

	// Should send notification for the 3-day window
	assert.GreaterOrEqual(t, len(notifSvc.sent), 1)
	found := false
	for _, n := range notifSvc.sent {
		if n.NotifType == domain.NotifBonusExpiring && n.UserID == userID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected bonus expiring notification")
}

func TestHandleBonusExpiryNotify_NoNotificationForDistantBonuses(t *testing.T) {
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	notifSvc := &mockNotificationService{}
	walletSvc := service.NewWalletService(walletRepo, logger.New(logger.LevelInfo))

	userID := uuid.New()
	walletID := uuid.New()

	err := walletRepo.Create(context.Background(), &domain.Wallet{
		ID:       walletID,
		UserID:   userID,
		Balance:  50000,
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	})
	require.NoError(t, err)

	// Create a bonus expiring in 60 days (not within notification windows)
	expiresIn60Days := time.Now().Add(60 * 24 * time.Hour)
	err = walletRepo.CreateTransaction(context.Background(), &domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         domain.WalletTxWelcomeBonus,
		Amount:       50000,
		BalanceAfter: 50000,
		Status:       domain.WalletTxStatusCompleted,
		Description:  "Приветственный бонус",
		IsBonus:      true,
		ExpiresAt:    &expiresIn60Days,
	})
	require.NoError(t, err)

	cs := newTestWalletCronScheduler(notifSvc, walletSvc, walletRepo)
	_ = cs.bonusExpiryNotify(context.Background())

	// Should NOT send any notification - bonus is too far from expiry
	assert.Empty(t, notifSvc.sent)
}

func TestGetBonusExpiryDays_Default(t *testing.T) {
	cs := newTestWalletCronScheduler(&mockNotificationService{}, nil, nil)
	days := cs.getBonusExpiryDays()
	assert.Equal(t, 180, days)
}
