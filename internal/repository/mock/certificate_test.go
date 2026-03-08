package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestCertificateRepo_Create(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}

	err := repo.Create(context.Background(), cert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cert.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if cert.Status != domain.CertificateStatusActive {
		t.Errorf("status = %v, want active", cert.Status)
	}
}

func TestCertificateRepo_Create_DuplicateCode(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert)

	cert2 := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "other@test.com",
		RecipientEmail: "other@test.com",
		Amount:         100000,
		Balance:        100000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	err := repo.Create(context.Background(), cert2)
	if err != domain.ErrAlreadyExists {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestCertificateRepo_GetByID(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert)

	found, err := repo.GetByID(context.Background(), cert.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Code != cert.Code {
		t.Errorf("code = %v, want %v", found.Code, cert.Code)
	}
}

func TestCertificateRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewCertificateRepo()

	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateRepo_GetByCode(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-XXXX-YYYY",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         300000,
		Balance:        300000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert)

	found, err := repo.GetByCode(context.Background(), "BANI-XXXX-YYYY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != cert.ID {
		t.Errorf("id = %v, want %v", found.ID, cert.ID)
	}
}

func TestCertificateRepo_GetByCode_NotFound(t *testing.T) {
	repo := mock.NewCertificateRepo()

	_, err := repo.GetByCode(context.Background(), "BANI-NOPE-NOPE")
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateRepo_Redeem(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		Status:         domain.CertificateStatusActive,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert)

	userID := uuid.New()
	err := repo.Redeem(context.Background(), cert.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), cert.ID)
	if found.RedeemedByID == nil || *found.RedeemedByID != userID {
		t.Errorf("redeemed_by_id = %v, want %v", found.RedeemedByID, userID)
	}
}

func TestCertificateRepo_Redeem_AlreadyRedeemed(t *testing.T) {
	repo := mock.NewCertificateRepo()

	cert := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		Status:         domain.CertificateStatusActive,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert)

	_ = repo.Redeem(context.Background(), cert.ID, uuid.New())

	err := repo.Redeem(context.Background(), cert.ID, uuid.New())
	if err != domain.ErrCertificateNotFound {
		t.Errorf("err = %v, want ErrCertificateNotFound", err)
	}
}

func TestCertificateRepo_ListByUser(t *testing.T) {
	repo := mock.NewCertificateRepo()
	userID := uuid.New()

	// Certificate purchased by user
	cert1 := &domain.GiftCertificate{
		Code:           "BANI-AAAA-BBBB",
		PurchaserID:    &userID,
		PurchaserEmail: "buyer@test.com",
		RecipientEmail: "recipient@test.com",
		Amount:         500000,
		Balance:        500000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert1)

	// Certificate redeemed by user
	cert2 := &domain.GiftCertificate{
		Code:           "BANI-CCCC-DDDD",
		PurchaserEmail: "other@test.com",
		RecipientEmail: "user@test.com",
		Amount:         300000,
		Balance:        300000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	}
	_ = repo.Create(context.Background(), cert2)
	_ = repo.Redeem(context.Background(), cert2.ID, userID)

	// Unrelated certificate
	_ = repo.Create(context.Background(), &domain.GiftCertificate{
		Code:           "BANI-EEEE-FFFF",
		PurchaserEmail: "stranger@test.com",
		RecipientEmail: "nobody@test.com",
		Amount:         100000,
		Balance:        100000,
		ValidUntil:     time.Now().Add(365 * 24 * time.Hour),
	})

	result, err := repo.ListByUser(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}
