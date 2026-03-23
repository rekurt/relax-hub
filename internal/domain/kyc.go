package domain

import (
	"time"

	"github.com/google/uuid"
)

// KYCStatus - статус заявки KYC
type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
	KYCStatusExpired  KYCStatus = "expired"
)

func (s KYCStatus) IsValid() bool {
	switch s {
	case KYCStatusPending, KYCStatusApproved, KYCStatusRejected, KYCStatusExpired:
		return true
	}
	return false
}

// KYCEntityType - тип юридического лица
type KYCEntityType string

const (
	KYCEntityIndividual      KYCEntityType = "individual"
	KYCEntitySoleProprietor  KYCEntityType = "sole_proprietor"
	KYCEntitySelfEmployed    KYCEntityType = "self_employed"
	KYCEntityLegalEntity     KYCEntityType = "legal_entity"
)

func (t KYCEntityType) IsValid() bool {
	switch t {
	case KYCEntityIndividual, KYCEntitySoleProprietor, KYCEntitySelfEmployed, KYCEntityLegalEntity:
		return true
	}
	return false
}

// KYCApplication - заявка на KYC верификацию
type KYCApplication struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Status          KYCStatus  `json:"status"`
	EntityType      KYCEntityType `json:"entity_type"`
	FullName        string     `json:"full_name"`
	INN             string     `json:"inn"`
	OGRNIP          string     `json:"ogrnip,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	DocumentURLs    []string   `json:"document_urls"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Validate проверяет обязательные поля по типу юридического лица
func (k *KYCApplication) Validate() error {
	if !k.EntityType.IsValid() {
		return ErrInvalidInput
	}
	if k.FullName == "" {
		return ErrInvalidInput
	}
	if len(k.DocumentURLs) == 0 {
		return ErrInvalidInput
	}

	switch k.EntityType {
	case KYCEntityIndividual:
		// ИНН необязателен для физлиц
	case KYCEntitySelfEmployed:
		if k.INN == "" {
			return ErrInvalidInput
		}
	case KYCEntitySoleProprietor:
		if k.INN == "" || k.OGRNIP == "" {
			return ErrInvalidInput
		}
	case KYCEntityLegalEntity:
		if k.INN == "" || k.CompanyName == "" {
			return ErrInvalidInput
		}
	}

	return nil
}
