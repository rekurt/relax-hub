package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCertificateStatus_IsValid(t *testing.T) {
	tests := []struct {
		status CertificateStatus
		valid  bool
	}{
		{CertificateStatusActive, true},
		{CertificateStatusUsed, true},
		{CertificateStatusExpired, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("CertificateStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestGiftCertificate_Validate(t *testing.T) {
	validUntil := time.Now().Add(24 * time.Hour)

	valid := &GiftCertificate{
		Code:           "BANI-ABCD-EFGH",
		PurchaserEmail: "buyer@example.com",
		RecipientEmail: "friend@example.com",
		Amount:         500000,
		Balance:        500000,
		Status:         CertificateStatusActive,
		ValidUntil:     validUntil,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid certificate returned error: %v", err)
	}

	// Valid without status (empty status is ok, skip validation)
	noStatus := &GiftCertificate{
		Code:           "BANI-ABCD-EFGH",
		PurchaserEmail: "buyer@example.com",
		RecipientEmail: "friend@example.com",
		Amount:         500000,
		Balance:        500000,
		ValidUntil:     validUntil,
	}
	if err := noStatus.Validate(); err != nil {
		t.Errorf("certificate without status returned error: %v", err)
	}

	tests := []struct {
		name    string
		cert    GiftCertificate
		wantErr error
	}{
		{
			"empty code",
			GiftCertificate{Code: "", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 1000, Balance: 1000, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"empty purchaser email",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "", RecipientEmail: "c@d.com", Amount: 1000, Balance: 1000, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"empty recipient email",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "", Amount: 1000, Balance: 1000, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"zero amount",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 0, Balance: 0, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"negative amount",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: -1000, Balance: 0, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"negative balance",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 1000, Balance: -1, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"balance exceeds amount",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 1000, Balance: 2000, ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"invalid status",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 1000, Balance: 1000, Status: "bad", ValidUntil: validUntil},
			ErrInvalidInput,
		},
		{
			"zero valid until",
			GiftCertificate{Code: "BANI-1234-5678", PurchaserEmail: "a@b.com", RecipientEmail: "c@d.com", Amount: 1000, Balance: 1000},
			ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cert.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGiftCertificate_IsExpired(t *testing.T) {
	expired := &GiftCertificate{ValidUntil: time.Now().Add(-1 * time.Hour)}
	if !expired.IsExpired() {
		t.Error("expected expired certificate to be expired")
	}

	notExpired := &GiftCertificate{ValidUntil: time.Now().Add(24 * time.Hour)}
	if notExpired.IsExpired() {
		t.Error("expected non-expired certificate to not be expired")
	}
}

func TestGiftCertificate_IsUsable(t *testing.T) {
	// Usable: active, not expired, has balance
	usable := &GiftCertificate{
		Status:     CertificateStatusActive,
		ValidUntil: time.Now().Add(24 * time.Hour),
		Balance:    10000,
	}
	if !usable.IsUsable() {
		t.Error("expected usable certificate")
	}

	// Not usable: expired
	expiredCert := &GiftCertificate{
		Status:     CertificateStatusActive,
		ValidUntil: time.Now().Add(-1 * time.Hour),
		Balance:    10000,
	}
	if expiredCert.IsUsable() {
		t.Error("expected expired certificate to not be usable")
	}

	// Not usable: used status
	usedCert := &GiftCertificate{
		Status:     CertificateStatusUsed,
		ValidUntil: time.Now().Add(24 * time.Hour),
		Balance:    0,
	}
	if usedCert.IsUsable() {
		t.Error("expected used certificate to not be usable")
	}

	// Not usable: zero balance
	zeroBal := &GiftCertificate{
		Status:     CertificateStatusActive,
		ValidUntil: time.Now().Add(24 * time.Hour),
		Balance:    0,
	}
	if zeroBal.IsUsable() {
		t.Error("expected zero-balance certificate to not be usable")
	}
}

func TestCertificateUsage_Validate(t *testing.T) {
	valid := &CertificateUsage{
		CertificateID: uuid.New(),
		BookingID:     uuid.New(),
		Amount:        5000,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid usage returned error: %v", err)
	}

	tests := []struct {
		name    string
		usage   CertificateUsage
		wantErr error
	}{
		{
			"nil certificate id",
			CertificateUsage{CertificateID: uuid.Nil, BookingID: uuid.New(), Amount: 1000},
			ErrInvalidInput,
		},
		{
			"nil booking id",
			CertificateUsage{CertificateID: uuid.New(), BookingID: uuid.Nil, Amount: 1000},
			ErrInvalidInput,
		},
		{
			"zero amount",
			CertificateUsage{CertificateID: uuid.New(), BookingID: uuid.New(), Amount: 0},
			ErrInvalidInput,
		},
		{
			"negative amount",
			CertificateUsage{CertificateID: uuid.New(), BookingID: uuid.New(), Amount: -500},
			ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.usage.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
