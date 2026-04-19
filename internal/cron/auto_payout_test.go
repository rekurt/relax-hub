package cron

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPayoutService implements service.PayoutService for auto-payout cron tests
type mockPayoutService struct {
	availableBalances map[uuid.UUID]int64
	requestedPayouts  []requestedPayout
	processedPayouts  []uuid.UUID
	requestErr        error
	processErr        error
}

type requestedPayout struct {
	UserID      uuid.UUID
	Amount      int64
	BankDetails json.RawMessage
}

func newMockPayoutService() *mockPayoutService {
	return &mockPayoutService{
		availableBalances: make(map[uuid.UUID]int64),
	}
}

func (m *mockPayoutService) RequestPayout(_ context.Context, userID uuid.UUID, amount int64, bankDetails json.RawMessage) (*domain.Payout, error) {
	if m.requestErr != nil {
		return nil, m.requestErr
	}
	m.requestedPayouts = append(m.requestedPayouts, requestedPayout{
		UserID:      userID,
		Amount:      amount,
		BankDetails: bankDetails,
	})
	return &domain.Payout{
		ID:     uuid.New(),
		UserID: userID,
		Amount: amount,
		Status: domain.PayoutStatusPending,
	}, nil
}

func (m *mockPayoutService) GetPayoutHistory(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Payout], error) {
	return nil, nil
}

func (m *mockPayoutService) SetAutoPayoutThreshold(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}

func (m *mockPayoutService) GetAutoPayoutSettings(_ context.Context, _ uuid.UUID) (*domain.AutoPayoutSettings, error) {
	return nil, nil
}

func (m *mockPayoutService) CalculateAvailableBalance(_ context.Context, userID uuid.UUID) (int64, error) {
	balance, ok := m.availableBalances[userID]
	if !ok {
		return 0, nil
	}
	return balance, nil
}

func (m *mockPayoutService) ProcessPayout(_ context.Context, payoutID uuid.UUID) error {
	if m.processErr != nil {
		return m.processErr
	}
	m.processedPayouts = append(m.processedPayouts, payoutID)
	return nil
}

func newTestAutoPayoutScheduler(payoutSvc service.PayoutService, payoutRepo *mock.PayoutRepo, pdRepo *mock.PaymentDetailsRepo) *CronScheduler {
	log := logger.New(logger.LevelInfo)
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()
	return NewCronScheduler(&config.Config{}, log, mockAnalyticsSvc, mockAnalyticsRepo,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		payoutSvc, payoutRepo, pdRepo,
	)
}

func TestAutoPayoutJob_NoActiveSettings(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	err := cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	assert.Empty(t, payoutSvc.requestedPayouts)
}

func TestAutoPayoutJob_BalanceBelowThreshold(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	userID := uuid.New()

	// Set threshold to 100,000 kopecks (1000 RUB)
	err := payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID:    userID,
		Threshold: 100_000,
	})
	require.NoError(t, err)

	// Available balance is only 50,000 kopecks (below threshold)
	payoutSvc.availableBalances[userID] = 50_000

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	err = cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	assert.Empty(t, payoutSvc.requestedPayouts)
}

func TestAutoPayoutJob_BalanceAboveThreshold_TriggersPayout(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	userID := uuid.New()

	// Set threshold to 100,000 kopecks
	err := payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID:    userID,
		Threshold: 100_000,
	})
	require.NoError(t, err)

	// Set payment details
	err = pdRepo.Upsert(context.Background(), &domain.PaymentDetails{
		ID:             uuid.New(),
		UserID:         userID,
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Test User",
	})
	require.NoError(t, err)

	// Available balance exceeds threshold
	payoutSvc.availableBalances[userID] = 200_000

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	err = cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	// Should have triggered one payout
	require.Len(t, payoutSvc.requestedPayouts, 1)
	assert.Equal(t, userID, payoutSvc.requestedPayouts[0].UserID)
	assert.Equal(t, int64(200_000), payoutSvc.requestedPayouts[0].Amount)
	assert.NotNil(t, payoutSvc.requestedPayouts[0].BankDetails)

	// Should have processed the payout
	require.Len(t, payoutSvc.processedPayouts, 1)
}

func TestAutoPayoutJob_MultipleUsers(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	// User1: above threshold
	err := payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID: user1, Threshold: 100_000,
	})
	require.NoError(t, err)
	payoutSvc.availableBalances[user1] = 150_000
	err = pdRepo.Upsert(context.Background(), &domain.PaymentDetails{
		ID: uuid.New(), UserID: user1, EntityType: domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111", CardHolderName: "User 1",
	})
	require.NoError(t, err)

	// User2: below threshold
	err = payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID: user2, Threshold: 200_000,
	})
	require.NoError(t, err)
	payoutSvc.availableBalances[user2] = 50_000

	// User3: above threshold
	err = payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID: user3, Threshold: 80_000,
	})
	require.NoError(t, err)
	payoutSvc.availableBalances[user3] = 300_000
	err = pdRepo.Upsert(context.Background(), &domain.PaymentDetails{
		ID: uuid.New(), UserID: user3, EntityType: domain.KYCEntityIndividual,
		BankCardNumber: "4222222222222222", CardHolderName: "User 3",
	})
	require.NoError(t, err)

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	err = cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	// Should have triggered payouts for user1 and user3 only
	assert.Len(t, payoutSvc.requestedPayouts, 2)
	assert.Len(t, payoutSvc.processedPayouts, 2)

	// Verify the right users got payouts
	payoutUsers := make(map[uuid.UUID]bool)
	for _, p := range payoutSvc.requestedPayouts {
		payoutUsers[p.UserID] = true
	}
	assert.True(t, payoutUsers[user1])
	assert.False(t, payoutUsers[user2])
	assert.True(t, payoutUsers[user3])
}

func TestAutoPayoutJob_NilServices_NoError(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	mockAnalyticsSvc := &MockAnalyticsService{}
	mockAnalyticsRepo := mock.NewAnalyticsRepo()
	cs := NewCronScheduler(&config.Config{}, log, mockAnalyticsSvc, mockAnalyticsRepo,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil,
	)

	err := cs.autoPayoutJob(context.Background())
	require.NoError(t, err)
}

func TestAutoPayoutJob_RequestPayoutError_Continues(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutSvc.requestErr = domain.ErrPayoutDailyLimitExceeded
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	userID := uuid.New()
	err := payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID: userID, Threshold: 100_000,
	})
	require.NoError(t, err)
	payoutSvc.availableBalances[userID] = 200_000
	err = pdRepo.Upsert(context.Background(), &domain.PaymentDetails{
		ID: uuid.New(), UserID: userID, EntityType: domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111", CardHolderName: "Test",
	})
	require.NoError(t, err)

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	// Should not return error - individual failures are logged, not propagated
	err = cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	assert.Empty(t, payoutSvc.processedPayouts)
}

func TestAutoPayoutJob_NoPaymentDetails_Continues(t *testing.T) {
	payoutSvc := newMockPayoutService()
	payoutRepo := mock.NewPayoutRepo().(*mock.PayoutRepo)
	pdRepo := mock.NewPaymentDetailsRepo()

	userID := uuid.New()
	err := payoutRepo.UpsertAutoPayoutSettings(context.Background(), &domain.AutoPayoutSettings{
		UserID: userID, Threshold: 100_000,
	})
	require.NoError(t, err)
	payoutSvc.availableBalances[userID] = 200_000
	// No payment details set - should fail gracefully

	cs := newTestAutoPayoutScheduler(payoutSvc, payoutRepo, pdRepo)
	err = cs.autoPayoutJob(context.Background())
	require.NoError(t, err)

	// Should not have triggered payout due to missing payment details
	assert.Empty(t, payoutSvc.requestedPayouts)
}
