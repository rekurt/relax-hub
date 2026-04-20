package domain

import (
	"time"

	"github.com/google/uuid"
)

type CertificateOrderStatus string

const (
	CertificateOrderStatusDraft          CertificateOrderStatus = "draft"
	CertificateOrderStatusPendingPayment CertificateOrderStatus = "pending_payment"
	CertificateOrderStatusPaid           CertificateOrderStatus = "paid"
	CertificateOrderStatusFailed         CertificateOrderStatus = "failed"
	CertificateOrderStatusCanceled       CertificateOrderStatus = "canceled"
)

func (s CertificateOrderStatus) IsValid() bool {
	switch s {
	case CertificateOrderStatusDraft, CertificateOrderStatusPendingPayment, CertificateOrderStatusPaid, CertificateOrderStatusFailed, CertificateOrderStatusCanceled:
		return true
	}
	return false
}

type CertificateOrder struct {
	ID             uuid.UUID
	PurchaserID    *uuid.UUID
	PurchaserEmail string
	RecipientEmail string
	RecipientName  string
	Message        string
	Amount         int64
	Status         CertificateOrderStatus
	PaymentMethod  PaymentMethod
	Provider       string
	ExternalID     string
	CertificateID  *uuid.UUID
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Certificate    *GiftCertificate
}

func (o *CertificateOrder) Validate() error {
	if o.PurchaserEmail == "" {
		return ErrInvalidInput
	}
	if o.Amount <= 0 {
		return ErrInvalidInput
	}
	if o.Status != "" && !o.Status.IsValid() {
		return ErrInvalidInput
	}
	if o.PaymentMethod != "" && !o.PaymentMethod.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
