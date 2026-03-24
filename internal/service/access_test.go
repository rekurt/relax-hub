package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

// noopNotifService is a no-op NotificationService for tests that don't verify notifications.
type noopNotifService struct{}

func (n *noopNotifService) Send(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _, _ string, _ map[string]string) error {
	return nil
}
func (n *noopNotifService) List(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Notification], error) {
	return &domain.PaginatedResult[domain.Notification]{}, nil
}
func (n *noopNotifService) MarkAsRead(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (n *noopNotifService) MarkAllAsRead(_ context.Context, _ uuid.UUID) error { return nil }
func (n *noopNotifService) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (n *noopNotifService) GetPreferences(_ context.Context, _ uuid.UUID) (*domain.NotificationPreferences, error) {
	prefs := domain.DefaultNotificationPreferences(uuid.Nil)
	return &prefs, nil
}
func (n *noopNotifService) UpdatePreferences(_ context.Context, _ uuid.UUID, _ *domain.NotificationPreferences) error {
	return nil
}

// noopReferralService is a no-op ReferralService for tests that don't verify referrals.
type noopReferralService struct{}

func (n *noopReferralService) GenerateCode(_ context.Context, _ uuid.UUID) (string, error) {
	return "testcode", nil
}
func (n *noopReferralService) RegisterReferral(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}
func (n *noopReferralService) CompleteReferral(_ context.Context, _ uuid.UUID) (*service.ReferralCompletionResult, error) {
	return nil, nil
}
func (n *noopReferralService) GetBalance(_ context.Context, userID uuid.UUID) (*domain.ReferralBalance, error) {
	return &domain.ReferralBalance{UserID: userID}, nil
}
func (n *noopReferralService) UseBalance(_ context.Context, _ uuid.UUID, _ int64, _ uuid.UUID) error {
	return nil
}
func (n *noopReferralService) RefundBalance(_ context.Context, _ uuid.UUID, _ int64, _ uuid.UUID) error {
	return nil
}
func (n *noopReferralService) GetStats(_ context.Context, _ uuid.UUID) (*domain.ReferralStats, error) {
	return &domain.ReferralStats{}, nil
}

type noopOTPService struct{}

func (n *noopOTPService) SendOTP(_ context.Context, _ string) error {
	return nil
}
func (n *noopOTPService) VerifyOTP(_ context.Context, _ string, _ string) (bool, error) {
	return true, nil
}

// noopPromoService is a no-op PromoService for tests that don't verify promo codes.
type noopPromoService struct{}

func (n *noopPromoService) Create(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ *domain.PromoCode) (*domain.PromoCode, error) {
	return &domain.PromoCode{}, nil
}
func (n *noopPromoService) Validate(_ context.Context, _ string, _ uuid.UUID, _ int64) (*domain.PromoCode, int64, error) {
	return &domain.PromoCode{}, 0, nil
}
func (n *noopPromoService) Apply(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ uuid.UUID, _ int64) (int64, error) {
	return 0, nil
}
func (n *noopPromoService) Deactivate(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (n *noopPromoService) ListByBathhouse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.PromoCode], error) {
	return &domain.PaginatedResult[domain.PromoCode]{}, nil
}
func (n *noopPromoService) RefundUsage(_ context.Context, _ uuid.UUID) error {
	return nil
}

// noopPaymentService is a no-op PaymentService for tests that don't verify payments.
type noopPaymentService struct{}

func (n *noopPaymentService) InitiatePayment(_ context.Context, _, _ uuid.UUID, _ domain.PaymentMethod) (string, error) {
	return "", nil
}
func (n *noopPaymentService) HandleWebhook(_ context.Context, _ service.WebhookEvent) error {
	return nil
}
func (n *noopPaymentService) RefundPayment(_ context.Context, _ uuid.UUID, _ bool, _ string) error {
	return nil
}
func (n *noopPaymentService) AdminRefund(_ context.Context, _ uuid.UUID, _ int64, _ string, _ string) error {
	return nil
}
func (n *noopPaymentService) GetPaymentByBooking(_ context.Context, _, _ uuid.UUID) (*domain.Payment, error) {
	return nil, domain.ErrPaymentNotFound
}
func (n *noopPaymentService) InitiateComboPayment(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ service.ComboPaymentRequest) (string, error) {
	return "", nil
}
func (n *noopPaymentService) CaptureHoldPayment(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (n *noopPaymentService) ReleaseHoldPayment(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (n *noopPaymentService) ListUserPayments(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Payment], error) {
	return &domain.PaginatedResult[domain.Payment]{}, nil
}

// noopWalletService is a no-op WalletService for tests that don't verify wallet operations.
type noopWalletService struct{}

func (n *noopWalletService) CreateWallet(_ context.Context, _ uuid.UUID, _ domain.WalletCurrency) (*domain.Wallet, error) {
	return &domain.Wallet{ID: uuid.New()}, nil
}
func (n *noopWalletService) GetWallet(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	return &domain.Wallet{ID: uuid.New(), Balance: 1000000}, nil
}
func (n *noopWalletService) TopUp(_ context.Context, _ uuid.UUID, _ int64) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}
func (n *noopWalletService) Spend(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}
func (n *noopWalletService) Hold(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string, _ time.Time) (*domain.WalletHold, error) {
	return &domain.WalletHold{}, nil
}
func (n *noopWalletService) CaptureHold(_ context.Context, _ uuid.UUID) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}
func (n *noopWalletService) ReleaseHold(_ context.Context, _ uuid.UUID) error { return nil }
func (n *noopWalletService) Refund(_ context.Context, _ uuid.UUID, _ int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}
func (n *noopWalletService) AddBonus(_ context.Context, _ uuid.UUID, _ int64, _ domain.WalletTransactionType, _ *time.Time, _ string) (*domain.WalletTransaction, error) {
	return &domain.WalletTransaction{}, nil
}
func (n *noopWalletService) GetBalance(_ context.Context, _ uuid.UUID) (*service.WalletBalanceSummary, error) {
	return &service.WalletBalanceSummary{}, nil
}
func (n *noopWalletService) ListTransactions(_ context.Context, _ uuid.UUID, _ domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	return &domain.PaginatedResult[domain.WalletTransaction]{}, nil
}
func (n *noopWalletService) GetActiveHolds(_ context.Context, _ uuid.UUID) ([]domain.WalletHold, error) {
	return nil, nil
}
func (n *noopWalletService) ExpireBonuses(_ context.Context) (int, error) { return 0, nil }
func (n *noopWalletService) ExpireBonusesForWallet(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (n *noopWalletService) FreezeAndZeroBalance(_ context.Context, _ uuid.UUID) error { return nil }

// noopCertificateService is a no-op CertificateService for tests that don't verify certificates.
type noopCertificateService struct{}

func (n *noopCertificateService) Purchase(_ context.Context, _ int64, _ *uuid.UUID, _, _, _, _ string) (*domain.GiftCertificate, error) {
	return nil, nil
}
func (n *noopCertificateService) Redeem(_ context.Context, _ string, _ uuid.UUID) (*domain.GiftCertificate, error) {
	return nil, nil
}
func (n *noopCertificateService) Apply(_ context.Context, _, _ uuid.UUID, _ int64) error {
	return nil
}
func (n *noopCertificateService) GetBalance(_ context.Context, _ string) (*domain.GiftCertificate, error) {
	return nil, nil
}
func (n *noopCertificateService) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	return &domain.PaginatedResult[domain.GiftCertificate]{}, nil
}

// noopAddOnService is a no-op AddOnService for tests that don't verify add-ons.
type noopAddOnService struct{}

func (n *noopAddOnService) CreateAddOn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ *domain.AddOn) (*domain.AddOn, error) {
	return &domain.AddOn{}, nil
}
func (n *noopAddOnService) UpdateAddOn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ *domain.AddOn) (*domain.AddOn, error) {
	return &domain.AddOn{}, nil
}
func (n *noopAddOnService) DeleteAddOn(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (n *noopAddOnService) ListAddOns(_ context.Context, _ uuid.UUID) ([]domain.AddOn, error) {
	return nil, nil
}
func (n *noopAddOnService) GetAddOn(_ context.Context, _ uuid.UUID) (*domain.AddOn, error) {
	return &domain.AddOn{}, nil
}
func (n *noopAddOnService) CalculateAddOnTotal(_ context.Context, _ []service.AddOnSelection, _ uuid.UUID, _ int, _ int) (int64, []service.AddOnLineItem, error) {
	return 0, nil, nil
}

// noopServiceFeeService is a no-op ServiceFeeService for tests that don't verify service fees.
type noopServiceFeeService struct{}

func (n *noopServiceFeeService) GetFeePercent(_ context.Context, _ string, _ *string) (float64, error) {
	return 0, nil
}
func (n *noopServiceFeeService) CalculateFee(_ context.Context, _ int64, _ string, _ *string) (int64, error) {
	return 0, nil
}
func (n *noopServiceFeeService) ListConfigs(_ context.Context) ([]domain.ServiceFeeConfig, error) {
	return nil, nil
}
func (n *noopServiceFeeService) UpsertConfig(_ context.Context, _ *domain.ServiceFeeConfig) error {
	return nil
}

type noopComplaintService struct{}

func (n *noopComplaintService) Report(_ context.Context, _ uuid.UUID, _ service.CreateComplaintInput) (*domain.Complaint, error) {
	return &domain.Complaint{}, nil
}
func (n *noopComplaintService) Resolve(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (*domain.Complaint, error) {
	return &domain.Complaint{}, nil
}
func (n *noopComplaintService) Dismiss(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.Complaint, error) {
	return &domain.Complaint{}, nil
}
func (n *noopComplaintService) List(_ context.Context, _ domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
	return &domain.PaginatedResult[domain.Complaint]{}, nil
}
func (n *noopComplaintService) GetByID(_ context.Context, _ uuid.UUID) (*domain.Complaint, error) {
	return &domain.Complaint{}, nil
}

func createBathhouse(t *testing.T, bhRepo *mock.BathhouseRepo, ownerID uuid.UUID) *domain.Bathhouse {
	t.Helper()
	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID:                        uuid.New(),
		OwnerID:                   ownerID,
		Name:                      "Test Bathhouse",
		Address:                   "123 Street",
		CityID:                    1,
		PricePerHour:              5000,
		MinDuration:               1,
		MaxGuests:                 10,
		LongSessionThresholdHours: 4,
		BaseCapacity:              10,
		WorkingHours:              wh,
		Status:                    domain.BathhouseStatusActive,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatal(err)
	}
	return bh
}

func TestAccessChecker_CanManageBathhouse_AdminAlwaysAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	adminID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), adminID, domain.RoleAdmin, bh.ID)
	if err != nil {
		t.Errorf("admin should always be allowed, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_OwnerAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Errorf("owner should be allowed for own bathhouse, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_OwnerForbiddenForOthersBathhouse(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), otherOwnerID, domain.RoleOwner, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("other owner should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_RepresentativeAllowed(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      repUserID,
		BathhouseID: bh.ID,
		OwnerID:     ownerID,
	}
	if err := repRepo.Create(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	err := checker.CanManageBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != nil {
		t.Errorf("representative assigned to bathhouse should be allowed, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_RepresentativeForbiddenForOtherBathhouse(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), repUserID, domain.RoleRepresentative, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("representative not assigned to bathhouse should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_ClientForbidden(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	err := checker.CanManageBathhouse(context.Background(), clientID, domain.RoleClient, bh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("client should be forbidden, got: %v", err)
	}
}

func TestAccessChecker_CanManageBathhouse_NotFound(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	checker := service.NewAccessChecker(repRepo, bhRepo)

	err := checker.CanManageBathhouse(context.Background(), uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("should return not found for non-existent bathhouse, got: %v", err)
	}
}
