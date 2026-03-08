package domain

import (
	"time"

	"github.com/google/uuid"
)

type CertificateStatus string

const (
	CertificateStatusActive  CertificateStatus = "active"
	CertificateStatusUsed    CertificateStatus = "used"
	CertificateStatusExpired CertificateStatus = "expired"
)

func (s CertificateStatus) IsValid() bool {
	switch s {
	case CertificateStatusActive, CertificateStatusUsed, CertificateStatusExpired:
		return true
	}
	return false
}

type GiftCertificate struct {
	ID             uuid.UUID
	Code           string            // уникальный код формата BANI-XXXX-XXXX
	PurchaserID    *uuid.UUID        // nil если куплен без регистрации
	PurchaserEmail string
	RecipientEmail string
	RecipientName  string
	Amount         int64             // номинал в копейках
	Balance        int64             // остаток в копейках
	Message        string            // поздравительное сообщение
	Status         CertificateStatus // active, used, expired
	ValidUntil     time.Time
	RedeemedByID   *uuid.UUID        // кто привязал к аккаунту
	CreatedAt      time.Time
}

func (c *GiftCertificate) Validate() error {
	if c.Code == "" {
		return ErrInvalidInput
	}
	if c.PurchaserEmail == "" {
		return ErrInvalidInput
	}
	if c.RecipientEmail == "" {
		return ErrInvalidInput
	}
	if c.Amount <= 0 {
		return ErrInvalidInput
	}
	if c.Balance < 0 {
		return ErrInvalidInput
	}
	if c.Balance > c.Amount {
		return ErrInvalidInput
	}
	if c.Status != "" && !c.Status.IsValid() {
		return ErrInvalidInput
	}
	if c.ValidUntil.IsZero() {
		return ErrInvalidInput
	}
	return nil
}

func (c *GiftCertificate) IsExpired() bool {
	return time.Now().After(c.ValidUntil)
}

func (c *GiftCertificate) IsUsable() bool {
	return c.Status == CertificateStatusActive && !c.IsExpired() && c.Balance > 0
}

type CertificateUsage struct {
	ID            uuid.UUID
	CertificateID uuid.UUID
	BookingID     uuid.UUID
	Amount        int64 // сумма использования в копейках
	UsedAt        time.Time
}

func (u *CertificateUsage) Validate() error {
	if u.CertificateID == uuid.Nil {
		return ErrInvalidInput
	}
	if u.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if u.Amount <= 0 {
		return ErrInvalidInput
	}
	return nil
}

