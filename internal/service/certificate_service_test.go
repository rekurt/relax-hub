package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type certTestEnv struct {
	svc      service.CertificateService
	certRepo *mock.CertificateRepo
}

func newCertTestEnv() *certTestEnv {
	certRepo := mock.NewCertificateRepo().(*mock.CertificateRepo)
	log := logger.New(logger.LevelWarn)
	svc := service.NewCertificateService(certRepo, log)
	return &certTestEnv{
		svc:      svc,
		certRepo: certRepo,
	}
}

func TestCertificateService_Purchase_Success(t *testing.T) {
	env := newCertTestEnv()
	userID := uuid.New()

	cert, err := env.svc.Purchase(context.Background(), 100000, &userID, "buyer@test.com", "recipient@test.com", "Иван", "С днём рождения!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cert.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if !strings.HasPrefix(cert.Code, "BANI-") {
		t.Errorf("code = %q, want prefix BANI-", cert.Code)
	}
	if len(cert.Code) != 14 { // BANI-XXXX-XXXX
		t.Errorf("code length = %d, want 14", len(cert.Code))
	}
	if cert.Amount != 100000 {
		t.Errorf("amount = %d, want 100000", cert.Amount)
	}
	if cert.Balance != 100000 {
		t.Errorf("balance = %d, want 100000", cert.Balance)
	}
	if cert.Status != domain.CertificateStatusActive {
		t.Errorf("status = %q, want %q", cert.Status, domain.CertificateStatusActive)
	}
	if cert.PurchaserID == nil || *cert.PurchaserID != userID {
		t.Error("purchaser ID mismatch")
	}
}

func TestCertificateService_Purchase_WithoutRegistration(t *testing.T) {
	env := newCertTestEnv()

	cert, err := env.svc.Purchase(context.Background(), 50000, nil, "anon@test.com", "recipient@test.com", "Мария", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cert.PurchaserID != nil {
		t.Error("expected nil PurchaserID for anonymous purchase")
	}
}

func TestCertificateService_Purchase_InvalidAmount(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.Purchase(context.Background(), 0, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}

	_, err = env.svc.Purchase(context.Background(), -100, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}

	_, err = env.svc.Purchase(context.Background(), 10_000_001, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for amount exceeding max", err)
	}
}

func TestCertificateService_Purchase_MissingEmail(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.Purchase(context.Background(), 10000, nil, "", "recipient@test.com", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for empty purchaser email", err)
	}

	_, err = env.svc.Purchase(context.Background(), 10000, nil, "buyer@test.com", "", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for empty recipient email", err)
	}
}

func TestCertificateService_Purchase_InvalidEmail(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.Purchase(context.Background(), 10000, nil, "not-an-email", "recipient@test.com", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for invalid purchaser email", err)
	}

	_, err = env.svc.Purchase(context.Background(), 10000, nil, "buyer@test.com", "also-invalid", "Test", "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput for invalid recipient email", err)
	}
}

func TestCertificateService_Redeem_Success(t *testing.T) {
	env := newCertTestEnv()

	cert, err := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Иван", "")
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}

	userID := uuid.New()
	redeemed, err := env.svc.Redeem(context.Background(), cert.Code, userID)
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}

	if redeemed.RedeemedByID == nil || *redeemed.RedeemedByID != userID {
		t.Error("expected certificate to be redeemed by user")
	}
}

func TestCertificateService_Redeem_Idempotent(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	userID := uuid.New()

	_, err := env.svc.Redeem(context.Background(), cert.Code, userID)
	if err != nil {
		t.Fatalf("first redeem: %v", err)
	}

	redeemed, err := env.svc.Redeem(context.Background(), cert.Code, userID)
	if err != nil {
		t.Fatalf("second redeem should be idempotent: %v", err)
	}
	if *redeemed.RedeemedByID != userID {
		t.Error("user ID mismatch on idempotent redeem")
	}
}

func TestCertificateService_Redeem_AlreadyRedeemedByOther(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	user1 := uuid.New()
	user2 := uuid.New()

	_, err := env.svc.Redeem(context.Background(), cert.Code, user1)
	if err != nil {
		t.Fatalf("first redeem: %v", err)
	}

	_, err = env.svc.Redeem(context.Background(), cert.Code, user2)
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateService_Redeem_InvalidCode(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.Redeem(context.Background(), "", uuid.New())
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}

	_, err = env.svc.Redeem(context.Background(), "NONEXISTENT", uuid.New())
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateService_Redeem_Expired(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	// Manually expire the certificate via repo
	expired, _ := env.certRepo.GetByID(context.Background(), cert.ID)
	expired.ValidUntil = time.Now().Add(-24 * time.Hour)
	env.certRepo.ForceUpdate(expired)

	_, err := env.svc.Redeem(context.Background(), cert.Code, uuid.New())
	if err != domain.ErrCertificateExpired {
		t.Errorf("err = %v, want ErrCertificateExpired", err)
	}
}

func TestCertificateService_Apply_Success(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")
	bookingID := uuid.New()

	err := env.svc.Apply(context.Background(), cert.ID, bookingID, 30000)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	updated, _ := env.certRepo.GetByID(context.Background(), cert.ID)
	if updated.Balance != 70000 {
		t.Errorf("balance = %d, want 70000", updated.Balance)
	}
}

func TestCertificateService_Apply_FullAmount(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 50000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	err := env.svc.Apply(context.Background(), cert.ID, uuid.New(), 50000)
	if err != nil {
		t.Fatalf("apply full amount: %v", err)
	}

	updated, _ := env.certRepo.GetByID(context.Background(), cert.ID)
	if updated.Balance != 0 {
		t.Errorf("balance = %d, want 0", updated.Balance)
	}
	if updated.Status != domain.CertificateStatusUsed {
		t.Errorf("status = %q, want %q", updated.Status, domain.CertificateStatusUsed)
	}
}

func TestCertificateService_Apply_InsufficientBalance(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 50000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	err := env.svc.Apply(context.Background(), cert.ID, uuid.New(), 60000)
	if err != domain.ErrCertificateInsufficientBalance {
		t.Errorf("err = %v, want ErrCertificateInsufficientBalance", err)
	}
}

func TestCertificateService_Apply_InvalidAmount(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 50000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	err := env.svc.Apply(context.Background(), cert.ID, uuid.New(), 0)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestCertificateService_Apply_Expired(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	expired, _ := env.certRepo.GetByID(context.Background(), cert.ID)
	expired.ValidUntil = time.Now().Add(-24 * time.Hour)
	env.certRepo.ForceUpdate(expired)

	err := env.svc.Apply(context.Background(), cert.ID, uuid.New(), 10000)
	if err != domain.ErrCertificateExpired {
		t.Errorf("err = %v, want ErrCertificateExpired", err)
	}
}

func TestCertificateService_GetBalance_Success(t *testing.T) {
	env := newCertTestEnv()

	cert, _ := env.svc.Purchase(context.Background(), 100000, nil, "buyer@test.com", "recipient@test.com", "Test", "")

	found, err := env.svc.GetBalance(context.Background(), cert.Code)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if found.Balance != 100000 {
		t.Errorf("balance = %d, want 100000", found.Balance)
	}
}

func TestCertificateService_GetBalance_NotFound(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.GetBalance(context.Background(), "BANI-0000-0000")
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateService_GetBalance_EmptyCode(t *testing.T) {
	env := newCertTestEnv()

	_, err := env.svc.GetBalance(context.Background(), "")
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestCertificateService_ListByUser_Success(t *testing.T) {
	env := newCertTestEnv()
	userID := uuid.New()

	_, _ = env.svc.Purchase(context.Background(), 100000, &userID, "buyer@test.com", "r1@test.com", "Test1", "")
	_, _ = env.svc.Purchase(context.Background(), 200000, &userID, "buyer@test.com", "r2@test.com", "Test2", "")

	result, err := env.svc.ListByUser(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
}

func TestCertificateService_ListByUser_Empty(t *testing.T) {
	env := newCertTestEnv()

	result, err := env.svc.ListByUser(context.Background(), uuid.New(), 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("total_count = %d, want 0", result.TotalCount)
	}
}
