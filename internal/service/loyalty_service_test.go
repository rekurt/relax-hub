package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newLoyaltyService() (service.LoyaltyService, *mock.LoyaltyRepo) {
	repo := mock.NewLoyaltyRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewLoyaltyService(repo, log)
	return svc, repo
}

func TestLoyaltyService_GetAccount_CreatesIfNotExists(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()

	account, err := svc.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if account.UserID != userID {
		t.Errorf("user_id = %v, want %v", account.UserID, userID)
	}
	if account.Level != domain.LoyaltyBronze {
		t.Errorf("level = %q, want %q", account.Level, domain.LoyaltyBronze)
	}
	if account.Points != 0 {
		t.Errorf("points = %d, want 0", account.Points)
	}
	if account.VisitCount != 0 {
		t.Errorf("visit_count = %d, want 0", account.VisitCount)
	}
}

func TestLoyaltyService_GetAccount_ReturnsExisting(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()

	// Pre-create account with some points
	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltySilver,
		Points:     500,
		TotalEarned: 1000,
		TotalSpent: 500,
		VisitCount: 10,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	account, err := svc.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if account.Level != domain.LoyaltySilver {
		t.Errorf("level = %q, want %q", account.Level, domain.LoyaltySilver)
	}
	if account.Points != 500 {
		t.Errorf("points = %d, want 500", account.Points)
	}
}

func TestLoyaltyService_EarnPoints_BronzeLevel(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// totalPrice = 10000 kopecks (100 rubles), bronze multiplier = 1.0
	// points = 10000 / 100 * 1.0 = 100
	earned, err := svc.EarnPoints(context.Background(), userID, bookingID, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if earned != 100 {
		t.Errorf("earned = %d, want 100", earned)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 100 {
		t.Errorf("points = %d, want 100", account.Points)
	}
	if account.TotalEarned != 100 {
		t.Errorf("total_earned = %d, want 100", account.TotalEarned)
	}
	if account.VisitCount != 1 {
		t.Errorf("visit_count = %d, want 1", account.VisitCount)
	}

	// Verify transaction was created
	txs, err := repo.ListTransactions(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}
	if txs.TotalCount != 1 {
		t.Fatalf("transaction count = %d, want 1", txs.TotalCount)
	}
	if txs.Items[0].Type != domain.LoyaltyTransactionEarn {
		t.Errorf("transaction type = %q, want %q", txs.Items[0].Type, domain.LoyaltyTransactionEarn)
	}
	if txs.Items[0].Amount != 100 {
		t.Errorf("transaction amount = %d, want 100", txs.Items[0].Amount)
	}
}

func TestLoyaltyService_EarnPoints_GoldLevel(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// Pre-create gold-level account
	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltyGold,
		Points:     0,
		VisitCount: 20,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// totalPrice = 10000 kopecks, gold multiplier = 1.5
	// points = round(10000 / 100 * 1.5) = 150
	earned, err := svc.EarnPoints(context.Background(), userID, bookingID, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if earned != 150 {
		t.Errorf("earned = %d, want 150", earned)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 150 {
		t.Errorf("points = %d, want 150", account.Points)
	}
}

func TestLoyaltyService_EarnPoints_ZeroPrice(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	earned, err := svc.EarnPoints(context.Background(), userID, bookingID, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if earned != 0 {
		t.Errorf("earned = %d, want 0", earned)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 0 {
		t.Errorf("points = %d, want 0", account.Points)
	}
}

func TestLoyaltyService_SpendPoints_Success(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// Pre-create account with points
	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:      userID,
		Level:       domain.LoyaltySilver,
		Points:      500,
		TotalEarned: 500,
		VisitCount:  10,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	err = svc.SpendPoints(context.Background(), userID, 200, bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 300 {
		t.Errorf("points = %d, want 300", account.Points)
	}
	if account.TotalSpent != 200 {
		t.Errorf("total_spent = %d, want 200", account.TotalSpent)
	}

	// Verify transaction
	txs, err := repo.ListTransactions(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}
	if txs.TotalCount != 1 {
		t.Fatalf("transaction count = %d, want 1", txs.TotalCount)
	}
	if txs.Items[0].Type != domain.LoyaltyTransactionSpend {
		t.Errorf("transaction type = %q, want %q", txs.Items[0].Type, domain.LoyaltyTransactionSpend)
	}
}

func TestLoyaltyService_SpendPoints_InsufficientBalance(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltyBronze,
		Points:     100,
		VisitCount: 0,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	err = svc.SpendPoints(context.Background(), userID, 200, bookingID)
	if !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Errorf("error = %v, want ErrInsufficientPoints", err)
	}
}

func TestLoyaltyService_SpendPoints_ZeroAmount(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := svc.SpendPoints(context.Background(), userID, 0, bookingID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoyaltyService_SpendPoints_NegativeAmount(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := svc.SpendPoints(context.Background(), userID, -50, bookingID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoyaltyService_GetDiscount(t *testing.T) {
	tests := []struct {
		name     string
		level    domain.LoyaltyLevel
		want     int
	}{
		{"bronze", domain.LoyaltyBronze, 0},
		{"silver", domain.LoyaltySilver, 3},
		{"gold", domain.LoyaltyGold, 5},
		{"platinum", domain.LoyaltyPlatinum, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newLoyaltyService()
			userID := uuid.New()

			err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
				UserID:     userID,
				Level:      tt.level,
				VisitCount: 0,
			})
			if err != nil {
				t.Fatalf("failed to create account: %v", err)
			}

			discount, err := svc.GetDiscount(context.Background(), userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if discount != tt.want {
				t.Errorf("discount = %d, want %d", discount, tt.want)
			}
		})
	}
}

func TestLoyaltyService_RecalculateLevel(t *testing.T) {
	tests := []struct {
		name       string
		visitCount int
		wantLevel  domain.LoyaltyLevel
	}{
		{"stays bronze", 3, domain.LoyaltyBronze},
		{"upgrade to silver", 5, domain.LoyaltySilver},
		{"upgrade to gold", 15, domain.LoyaltyGold},
		{"upgrade to platinum", 30, domain.LoyaltyPlatinum},
		{"platinum at 50", 50, domain.LoyaltyPlatinum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newLoyaltyService()
			userID := uuid.New()

			err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
				UserID:     userID,
				Level:      domain.LoyaltyBronze,
				VisitCount: tt.visitCount,
			})
			if err != nil {
				t.Fatalf("failed to create account: %v", err)
			}

			_, err = svc.RecalculateLevel(context.Background(), userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			account, err := repo.GetAccount(context.Background(), userID)
			if err != nil {
				t.Fatalf("failed to get account: %v", err)
			}
			if account.Level != tt.wantLevel {
				t.Errorf("level = %q, want %q", account.Level, tt.wantLevel)
			}
		})
	}
}

func TestLoyaltyService_RecalculateLevel_NoChangeIfSameLevel(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()

	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltySilver,
		VisitCount: 10,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, err = svc.RecalculateLevel(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Level != domain.LoyaltySilver {
		t.Errorf("level = %q, want %q", account.Level, domain.LoyaltySilver)
	}
}

func TestLoyaltyService_ListTransactions(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// Create account and some transactions
	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID: userID,
		Level:  domain.LoyaltyBronze,
		Points: 300,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	for i := 0; i < 3; i++ {
		err := repo.CreateTransaction(context.Background(), &domain.LoyaltyTransaction{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        domain.LoyaltyTransactionEarn,
			Amount:      100,
			BookingID:   &bookingID,
			Description: "test earn",
		})
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
	}

	result, err := svc.ListTransactions(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}
}

func TestLoyaltyService_RefundPoints_Success(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// Pre-create account with spent points
	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:      userID,
		Level:       domain.LoyaltySilver,
		Points:      300,
		TotalEarned: 500,
		TotalSpent:  200,
		VisitCount:  10,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	err = svc.RefundPoints(context.Background(), userID, 150, bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 450 {
		t.Errorf("points = %d, want 450", account.Points)
	}
	if account.TotalSpent != 50 {
		t.Errorf("total_spent = %d, want 50", account.TotalSpent)
	}

	// Verify refund transaction was created
	txs, err := repo.ListTransactions(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}
	if txs.TotalCount != 1 {
		t.Fatalf("transaction count = %d, want 1", txs.TotalCount)
	}
	if txs.Items[0].Type != domain.LoyaltyTransactionRefund {
		t.Errorf("transaction type = %q, want %q", txs.Items[0].Type, domain.LoyaltyTransactionRefund)
	}
	if txs.Items[0].Amount != 150 {
		t.Errorf("transaction amount = %d, want 150", txs.Items[0].Amount)
	}
}

func TestLoyaltyService_RefundPoints_ExceedsTotalSpent(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltyBronze,
		Points:     100,
		TotalSpent: 50,
		VisitCount: 2,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	err = svc.RefundPoints(context.Background(), userID, 100, bookingID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoyaltyService_RefundPoints_ZeroAmount(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := svc.RefundPoints(context.Background(), userID, 0, bookingID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoyaltyService_RefundPoints_NegativeAmount(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := svc.RefundPoints(context.Background(), userID, -50, bookingID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoyaltyService_RefundPoints_AccountNotFound(t *testing.T) {
	svc, _ := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	// RefundPoints on a non-existent account - GetAccount auto-creates,
	// but TotalSpent=0, so refund of any positive amount should fail
	err := svc.RefundPoints(context.Background(), userID, 100, bookingID)
	if err == nil {
		t.Error("expected error for refund on new account with no spent points")
	}
}

func TestLoyaltyService_EarnPoints_PlatinumMultiplier(t *testing.T) {
	svc, repo := newLoyaltyService()
	userID := uuid.New()
	bookingID := uuid.New()

	err := repo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     userID,
		Level:      domain.LoyaltyPlatinum,
		Points:     0,
		VisitCount: 35,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// totalPrice = 5000 kopecks, platinum multiplier = 2.0
	// points = round(5000 / 100 * 2.0) = 100
	earned, err := svc.EarnPoints(context.Background(), userID, bookingID, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if earned != 100 {
		t.Errorf("earned = %d, want 100", earned)
	}

	account, err := repo.GetAccount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	if account.Points != 100 {
		t.Errorf("points = %d, want 100", account.Points)
	}
}
