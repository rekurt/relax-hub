package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newRegionService() (service.RegionService, *mock.UserRepo, *mock.WalletRepo, *mock.BookingRepo, *mock.DisputeRepo, *mock.LoyaltyRepo) {
	userRepo := mock.NewUserRepo()
	walletRepo := mock.NewWalletRepo().(*mock.WalletRepo)
	bookingRepo := mock.NewBookingRepo()
	disputeRepo := mock.NewDisputeRepo().(*mock.DisputeRepo)
	loyaltyRepo := mock.NewLoyaltyRepo()
	certRepo := mock.NewCertificateRepo()
	log := logger.New(logger.LevelError)
	svc := service.NewRegionService(userRepo, walletRepo, bookingRepo, disputeRepo, loyaltyRepo, certRepo, log)
	return svc, userRepo, walletRepo, bookingRepo, disputeRepo, loyaltyRepo
}

func createRegionTestUser(t *testing.T, userRepo *mock.UserRepo, region domain.UserRegion) (*domain.User, uuid.UUID) {
	t.Helper()
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     domain.RoleClient,
		IsActive: true,
		Region:   region,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user, user.ID
}

func TestRegionService_GetRegion(t *testing.T) {
	svc, userRepo, _, _, _, _ := newRegionService()

	t.Run("returns user region", func(t *testing.T) {
		_, userID := createRegionTestUser(t, userRepo, domain.RegionBY)
		region, err := svc.GetRegion(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if region != domain.RegionBY {
			t.Errorf("region = %q, want %q", region, domain.RegionBY)
		}
	})

	t.Run("defaults to RU for empty region", func(t *testing.T) {
		user := &domain.User{
			ID: uuid.New(), Email: "empty@example.com", Name: "Empty",
			Role: domain.RoleClient, IsActive: true, Region: "",
		}
		_ = userRepo.Create(context.Background(), user)

		region, err := svc.GetRegion(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if region != domain.RegionRU {
			t.Errorf("region = %q, want %q", region, domain.RegionRU)
		}
	})

	t.Run("returns error for unknown user", func(t *testing.T) {
		_, err := svc.GetRegion(context.Background(), uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRegionService_SwitchRegion_Success(t *testing.T) {
	svc, userRepo, walletRepo, _, _, loyaltyRepo := newRegionService()
	ctx := context.Background()

	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	// Create an active wallet with zero balance
	wallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  0,
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	}
	_ = walletRepo.Create(ctx, wallet)

	// Create loyalty account
	_ = loyaltyRepo.CreateAccount(ctx, &domain.LoyaltyAccount{
		UserID: userID,
		Level:  domain.LoyaltyGold,
	})

	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user region updated
	user, _ := userRepo.GetByID(ctx, userID)
	if user.Region != domain.RegionBY {
		t.Errorf("user region = %q, want %q", user.Region, domain.RegionBY)
	}

	// Verify old wallet archived
	_, err = walletRepo.GetByUserID(ctx, userID)
	// Should find a new wallet now (BYN)
	if err != nil {
		t.Fatalf("expected new wallet, got error: %v", err)
	}

	// Verify loyalty reset
	account, _ := loyaltyRepo.GetAccount(ctx, userID)
	if account.Level != domain.LoyaltyBronze {
		t.Errorf("loyalty level = %q, want %q", account.Level, domain.LoyaltyBronze)
	}
}

func TestRegionService_SwitchRegion_InvalidRegion(t *testing.T) {
	svc, userRepo, _, _, _, _ := newRegionService()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	err := svc.SwitchRegion(context.Background(), userID, domain.UserRegion("XX"))
	if err != domain.ErrRegionInvalid {
		t.Errorf("expected ErrRegionInvalid, got %v", err)
	}
}

func TestRegionService_SwitchRegion_SameRegion(t *testing.T) {
	svc, userRepo, _, _, _, _ := newRegionService()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	err := svc.SwitchRegion(context.Background(), userID, domain.RegionRU)
	if err != domain.ErrRegionSameAsCurrent {
		t.Errorf("expected ErrRegionSameAsCurrent, got %v", err)
	}
}

func TestRegionService_SwitchRegion_BlockedByWalletBalance(t *testing.T) {
	svc, userRepo, walletRepo, _, _, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	wallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  10000, // 100 RUB
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	}
	_ = walletRepo.Create(ctx, wallet)

	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != domain.ErrRegionSwitchBlocked {
		t.Errorf("expected ErrRegionSwitchBlocked, got %v", err)
	}
}

func TestRegionService_SwitchRegion_BlockedByActiveBookings(t *testing.T) {
	svc, userRepo, _, bookingRepo, _, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	// Create an active booking
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: uuid.New(),
		Status:      domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(ctx, booking)

	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != domain.ErrRegionSwitchBlocked {
		t.Errorf("expected ErrRegionSwitchBlocked, got %v", err)
	}
}

func TestRegionService_SwitchRegion_BlockedByOpenDisputes(t *testing.T) {
	svc, userRepo, _, _, disputeRepo, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	dispute := &domain.Dispute{
		ID:           uuid.New(),
		BookingID:    uuid.New(),
		InitiatorID:  userID,
		RespondentID: uuid.New(),
		Status:       domain.DisputeStatusOpen,
	}
	_ = disputeRepo.Create(ctx, dispute)

	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != domain.ErrRegionSwitchBlocked {
		t.Errorf("expected ErrRegionSwitchBlocked, got %v", err)
	}
}

func TestRegionService_SwitchRegion_BlockedByHeldWalletAmount(t *testing.T) {
	svc, userRepo, walletRepo, _, _, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	wallet := &domain.Wallet{
		ID:         uuid.New(),
		UserID:     userID,
		Balance:    0,
		HeldAmount: 5000, // held funds
		Currency:   domain.WalletCurrencyRUB,
		Status:     domain.WalletStatusActive,
	}
	_ = walletRepo.Create(ctx, wallet)

	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != domain.ErrRegionSwitchBlocked {
		t.Errorf("expected ErrRegionSwitchBlocked, got %v", err)
	}
}

func TestRegionService_SwitchRegion_NoWallet(t *testing.T) {
	svc, userRepo, _, _, _, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionRU)

	// No wallet created, should still switch successfully
	err := svc.SwitchRegion(ctx, userID, domain.RegionBY)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(ctx, userID)
	if user.Region != domain.RegionBY {
		t.Errorf("user region = %q, want %q", user.Region, domain.RegionBY)
	}
}

func TestRegionService_SwitchRegion_BYtoRU(t *testing.T) {
	svc, userRepo, walletRepo, _, _, _ := newRegionService()
	ctx := context.Background()
	_, userID := createRegionTestUser(t, userRepo, domain.RegionBY)

	wallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  0,
		Currency: domain.WalletCurrencyBYN,
		Status:   domain.WalletStatusActive,
	}
	_ = walletRepo.Create(ctx, wallet)

	err := svc.SwitchRegion(ctx, userID, domain.RegionRU)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(ctx, userID)
	if user.Region != domain.RegionRU {
		t.Errorf("user region = %q, want %q", user.Region, domain.RegionRU)
	}
}
