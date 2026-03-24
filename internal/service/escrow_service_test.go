package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWalletService implements service.WalletService for testing
type mockWalletService struct {
	wallets      map[uuid.UUID]*domain.Wallet
	refundCalls  []refundCall
}

type refundCall struct {
	WalletID    uuid.UUID
	Amount      int64
	RefType     string
	RefID       *uuid.UUID
	Description string
}

func newMockWalletService() *mockWalletService {
	return &mockWalletService{
		wallets: make(map[uuid.UUID]*domain.Wallet),
	}
}

func (m *mockWalletService) CreateWallet(_ context.Context, userID uuid.UUID, currency domain.WalletCurrency) (*domain.Wallet, error) {
	w := &domain.Wallet{ID: uuid.New(), UserID: userID, Currency: currency, Balance: 0}
	m.wallets[w.ID] = w
	return w, nil
}

func (m *mockWalletService) GetWallet(_ context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	for _, w := range m.wallets {
		if w.UserID == userID {
			return w, nil
		}
	}
	return nil, domain.ErrWalletNotFound
}

func (m *mockWalletService) TopUp(_ context.Context, _ uuid.UUID, _ int64) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockWalletService) Spend(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockWalletService) Hold(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	return &domain.WalletHold{}, nil
}

func (m *mockWalletService) CaptureHold(_ context.Context, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockWalletService) ReleaseHold(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockWalletService) Refund(_ context.Context, walletID uuid.UUID, amount int64, refType string, refID *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	m.refundCalls = append(m.refundCalls, refundCall{
		WalletID: walletID, Amount: amount, RefType: refType, RefID: refID, Description: description,
	})
	return &domain.WalletTransaction{ID: uuid.New()}, nil
}

func (m *mockWalletService) AddBonus(_ context.Context, _ uuid.UUID, _ int64, _ domain.WalletTransactionType, _ *time.Time, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}

func (m *mockWalletService) GetBalance(_ context.Context, _ uuid.UUID) (*service.WalletBalanceSummary, error) {
	return &service.WalletBalanceSummary{}, nil
}

func (m *mockWalletService) ListTransactions(_ context.Context, _ uuid.UUID, _ domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	return &domain.PaginatedResult[domain.WalletTransaction]{}, nil
}

func (m *mockWalletService) GetActiveHolds(_ context.Context, _ uuid.UUID) ([]domain.WalletHold, error) {
	return nil, nil
}

func (m *mockWalletService) ExpireBonuses(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockWalletService) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockWalletService) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error {
	return nil
}

func setupEscrowTest(t *testing.T) (service.EscrowService, *mock.EscrowRepo, *mock.BookingRepo, *mock.BathhouseRepo, *mockWalletService) {
	t.Helper()

	escrowRepo := mock.NewEscrowRepo().(*mock.EscrowRepo)
	bookingRepo := mock.NewBookingRepo()
	bhRepo := mock.NewBathhouseRepo()
	walletSvc := newMockWalletService()
	log := logger.New(logger.LevelWarn)

	svc := service.NewEscrowService(escrowRepo, bookingRepo, bhRepo, walletSvc, log)
	return svc, escrowRepo, bookingRepo, bhRepo, walletSvc
}

func TestCreateEscrow(t *testing.T) {
	svc, _, _, _, _ := setupEscrowTest(t)
	ctx := context.Background()

	bookingID := uuid.New()

	t.Run("successful creation", func(t *testing.T) {
		escrow, err := svc.CreateEscrow(ctx, bookingID, 100000, 10000)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, escrow.ID)
		assert.Equal(t, bookingID, escrow.BookingID)
		assert.Equal(t, int64(100000), escrow.Amount)
		assert.Equal(t, int64(10000), escrow.ServiceFee)
		assert.Equal(t, domain.EscrowHeld, escrow.Status)
		assert.True(t, escrow.ClaimPeriodEndsAt.After(time.Now()))
		assert.Nil(t, escrow.ReleasedAt)
	})

	t.Run("correct claim period", func(t *testing.T) {
		escrow, err := svc.CreateEscrow(ctx, uuid.New(), 50000, 5000)
		require.NoError(t, err)
		// Default 48 hours claim period
		expectedEnd := time.Now().Add(48 * time.Hour)
		assert.WithinDuration(t, expectedEnd, escrow.ClaimPeriodEndsAt, 5*time.Second)
	})

	t.Run("invalid amount", func(t *testing.T) {
		_, err := svc.CreateEscrow(ctx, uuid.New(), 0, 0)
		assert.Error(t, err)
	})

	t.Run("service fee exceeds amount", func(t *testing.T) {
		_, err := svc.CreateEscrow(ctx, uuid.New(), 10000, 20000)
		assert.Error(t, err)
	})
}

func TestReleaseToOwner(t *testing.T) {
	ctx := context.Background()

	t.Run("successful release after claim period", func(t *testing.T) {
		svc, escrowRepo, bookingRepo, bhRepo, walletSvc := setupEscrowTest(t)

		ownerID := uuid.New()
		bathhouseID := uuid.New()
		bookingID := uuid.New()

		// Setup owner wallet
		ownerWallet := &domain.Wallet{ID: uuid.New(), UserID: ownerID, Balance: 0, Currency: "RUB"}
		walletSvc.wallets[ownerWallet.ID] = ownerWallet

		// Setup bathhouse
		bh := &domain.Bathhouse{
			ID:      bathhouseID,
			OwnerID: ownerID,
			Name:    "Test Bathhouse",
			Status:  domain.BathhouseStatusActive,
			PricePerHour: 100000,
			MaxGuests: 10,
			BaseCapacity: 5,
			LongSessionThresholdHours: 4,
			BufferMinutes: 30,
			LeadTimeHours: 2,
			MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant,
			RequestTimeout: 24,
		}
		require.NoError(t, bhRepo.Create(ctx, bh))

		// Setup booking
		booking := &domain.Booking{
			ID:          bookingID,
			UserID:      uuid.New(),
			BathhouseID: bathhouseID,
			StartTime:   time.Now().Add(-2 * time.Hour),
			EndTime:     time.Now().Add(-1 * time.Hour),
			GuestCount:  2,
			TotalPrice:  100000,
			Status:      domain.BookingCompleted,
		}
		require.NoError(t, bookingRepo.Create(ctx, booking))

		// Create escrow with past claim period
		escrow := &domain.Escrow{
			ID:                uuid.New(),
			BookingID:         bookingID,
			Amount:            100000,
			ServiceFee:        10000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(-1 * time.Hour), // already matured
			CreatedAt:         time.Now().Add(-49 * time.Hour),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ReleaseToOwner(ctx, escrow.ID)
		require.NoError(t, err)

		// Verify escrow status updated
		updated, err := escrowRepo.GetByID(ctx, escrow.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.EscrowReleased, updated.Status)
		assert.NotNil(t, updated.ReleasedAt)

		// Verify wallet was credited with amount - service fee
		require.Len(t, walletSvc.refundCalls, 1)
		assert.Equal(t, ownerWallet.ID, walletSvc.refundCalls[0].WalletID)
		assert.Equal(t, int64(90000), walletSvc.refundCalls[0].Amount) // 100000 - 10000
		assert.Equal(t, "escrow_release", walletSvc.refundCalls[0].RefType)
	})

	t.Run("release before claim period fails", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		escrow := &domain.Escrow{
			ID:                uuid.New(),
			BookingID:         uuid.New(),
			Amount:            100000,
			ServiceFee:        10000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(24 * time.Hour), // not yet matured
			CreatedAt:         time.Now(),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ReleaseToOwner(ctx, escrow.ID)
		assert.ErrorIs(t, err, domain.ErrEscrowNotMatured)
	})

	t.Run("release already released fails", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		now := time.Now()
		escrow := &domain.Escrow{
			ID:                uuid.New(),
			BookingID:         uuid.New(),
			Amount:            100000,
			ServiceFee:        10000,
			Status:            domain.EscrowReleased,
			ClaimPeriodEndsAt: now.Add(-1 * time.Hour),
			ReleasedAt:        &now,
			CreatedAt:         now.Add(-49 * time.Hour),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ReleaseToOwner(ctx, escrow.ID)
		assert.ErrorIs(t, err, domain.ErrEscrowAlreadyReleased)
	})
}

func TestMarkDisputed(t *testing.T) {
	ctx := context.Background()

	t.Run("successful dispute", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		escrow := &domain.Escrow{
			ID:                uuid.New(),
			BookingID:         uuid.New(),
			Amount:            100000,
			ServiceFee:        10000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now(),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.MarkDisputed(ctx, escrow.ID)
		require.NoError(t, err)

		updated, err := escrowRepo.GetByID(ctx, escrow.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.EscrowDisputed, updated.Status)
	})

	t.Run("dispute prevents release", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		escrow := &domain.Escrow{
			ID:                uuid.New(),
			BookingID:         uuid.New(),
			Amount:            100000,
			ServiceFee:        10000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(-1 * time.Hour), // matured
			CreatedAt:         time.Now().Add(-49 * time.Hour),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		// Mark as disputed
		err := svc.MarkDisputed(ctx, escrow.ID)
		require.NoError(t, err)

		// Try to release — should fail
		err = svc.ReleaseToOwner(ctx, escrow.ID)
		assert.ErrorIs(t, err, domain.ErrEscrowDisputed)
	})
}

func TestProcessMaturedEscrows(t *testing.T) {
	ctx := context.Background()

	t.Run("releases all matured escrows", func(t *testing.T) {
		svc, escrowRepo, bookingRepo, bhRepo, walletSvc := setupEscrowTest(t)

		ownerID := uuid.New()
		bathhouseID := uuid.New()

		// Setup owner wallet
		ownerWallet := &domain.Wallet{ID: uuid.New(), UserID: ownerID, Balance: 0, Currency: "RUB"}
		walletSvc.wallets[ownerWallet.ID] = ownerWallet

		// Setup bathhouse
		bh := &domain.Bathhouse{
			ID:      bathhouseID,
			OwnerID: ownerID,
			Name:    "Test Bathhouse",
			Status:  domain.BathhouseStatusActive,
			PricePerHour: 100000,
			MaxGuests: 10,
			BaseCapacity: 5,
			LongSessionThresholdHours: 4,
			BufferMinutes: 30,
			LeadTimeHours: 2,
			MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant,
			RequestTimeout: 24,
		}
		require.NoError(t, bhRepo.Create(ctx, bh))

		// Create 2 matured escrows and 1 not matured
		for i := 0; i < 2; i++ {
			bookingID := uuid.New()
			booking := &domain.Booking{
				ID: bookingID, UserID: uuid.New(), BathhouseID: bathhouseID,
				StartTime: time.Now().Add(-3 * time.Hour), EndTime: time.Now().Add(-2 * time.Hour),
				GuestCount: 1, TotalPrice: 50000, Status: domain.BookingCompleted,
			}
			require.NoError(t, bookingRepo.Create(ctx, booking))

			escrow := &domain.Escrow{
				ID: uuid.New(), BookingID: bookingID,
				Amount: 50000, ServiceFee: 5000,
				Status:            domain.EscrowHeld,
				ClaimPeriodEndsAt: time.Now().Add(-1 * time.Hour),
				CreatedAt:         time.Now().Add(-49 * time.Hour),
			}
			require.NoError(t, escrowRepo.Create(ctx, escrow))
		}

		// Not matured escrow
		notMaturedEscrow := &domain.Escrow{
			ID: uuid.New(), BookingID: uuid.New(),
			Amount: 50000, ServiceFee: 5000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now(),
		}
		require.NoError(t, escrowRepo.Create(ctx, notMaturedEscrow))

		released, err := svc.ProcessMaturedEscrows(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, released)

		// Verify not-matured is still held
		stillHeld, err := escrowRepo.GetByID(ctx, notMaturedEscrow.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.EscrowHeld, stillHeld.Status)
	})

	t.Run("skips disputed escrows", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		// Create a disputed escrow with past claim period
		escrow := &domain.Escrow{
			ID: uuid.New(), BookingID: uuid.New(),
			Amount: 50000, ServiceFee: 5000,
			Status:            domain.EscrowDisputed,
			ClaimPeriodEndsAt: time.Now().Add(-1 * time.Hour),
			CreatedAt:         time.Now().Add(-49 * time.Hour),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		released, err := svc.ProcessMaturedEscrows(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, released) // disputed is not in ListMatured

		// Verify still disputed
		updated, err := escrowRepo.GetByID(ctx, escrow.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.EscrowDisputed, updated.Status)
	})
}

func TestProcessRefund(t *testing.T) {
	ctx := context.Background()

	t.Run("successful refund", func(t *testing.T) {
		svc, escrowRepo, bookingRepo, _, walletSvc := setupEscrowTest(t)

		clientID := uuid.New()
		bookingID := uuid.New()

		// Setup client wallet
		clientWallet := &domain.Wallet{ID: uuid.New(), UserID: clientID, Balance: 0, Currency: "RUB"}
		walletSvc.wallets[clientWallet.ID] = clientWallet

		// Setup booking
		booking := &domain.Booking{
			ID: bookingID, UserID: clientID, BathhouseID: uuid.New(),
			StartTime: time.Now().Add(-2 * time.Hour), EndTime: time.Now().Add(-1 * time.Hour),
			GuestCount: 1, TotalPrice: 100000, Status: domain.BookingCompleted,
		}
		require.NoError(t, bookingRepo.Create(ctx, booking))

		escrow := &domain.Escrow{
			ID: uuid.New(), BookingID: bookingID,
			Amount: 100000, ServiceFee: 10000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now(),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ProcessRefund(ctx, escrow.ID, 100000)
		require.NoError(t, err)

		// Verify escrow is refunded
		updated, err := escrowRepo.GetByID(ctx, escrow.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.EscrowRefunded, updated.Status)

		// Verify client wallet was credited
		require.Len(t, walletSvc.refundCalls, 1)
		assert.Equal(t, clientWallet.ID, walletSvc.refundCalls[0].WalletID)
		assert.Equal(t, int64(100000), walletSvc.refundCalls[0].Amount)
		assert.Equal(t, "escrow_refund", walletSvc.refundCalls[0].RefType)
	})

	t.Run("refund exceeds amount", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		escrow := &domain.Escrow{
			ID: uuid.New(), BookingID: uuid.New(),
			Amount: 50000, ServiceFee: 5000,
			Status:            domain.EscrowHeld,
			ClaimPeriodEndsAt: time.Now().Add(24 * time.Hour),
			CreatedAt:         time.Now(),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ProcessRefund(ctx, escrow.ID, 60000)
		assert.ErrorIs(t, err, domain.ErrRefundExceedsAmount)
	})

	t.Run("refund already released fails", func(t *testing.T) {
		svc, escrowRepo, _, _, _ := setupEscrowTest(t)

		now := time.Now()
		escrow := &domain.Escrow{
			ID: uuid.New(), BookingID: uuid.New(),
			Amount: 50000, ServiceFee: 5000,
			Status:            domain.EscrowReleased,
			ClaimPeriodEndsAt: now.Add(-1 * time.Hour),
			ReleasedAt:        &now,
			CreatedAt:         now.Add(-49 * time.Hour),
		}
		require.NoError(t, escrowRepo.Create(ctx, escrow))

		err := svc.ProcessRefund(ctx, escrow.ID, 50000)
		assert.ErrorIs(t, err, domain.ErrEscrowAlreadyReleased)
	})
}
