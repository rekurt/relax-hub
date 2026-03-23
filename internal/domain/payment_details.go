package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// PaymentDetails - платежные реквизиты владельца
type PaymentDetails struct {
	ID                   uuid.UUID     `json:"id"`
	UserID               uuid.UUID     `json:"user_id"`
	EntityType           KYCEntityType `json:"entity_type"`
	BankCardNumber       string        `json:"bank_card_number,omitempty"`
	CardHolderName       string        `json:"card_holder_name,omitempty"`
	BankAccount          string        `json:"bank_account,omitempty"`
	BIK                  string        `json:"bik,omitempty"`
	INN                  string        `json:"inn,omitempty"`
	CorrespondentAccount string        `json:"correspondent_account,omitempty"`
	BankName             string        `json:"bank_name,omitempty"`
	IsVerified           bool          `json:"is_verified"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

var digitsOnly = regexp.MustCompile(`^\d+$`)

// Validate проверяет обязательные поля в зависимости от типа юридического лица
func (pd *PaymentDetails) Validate() error {
	if !pd.EntityType.IsValid() {
		return ErrInvalidInput
	}

	switch pd.EntityType {
	case KYCEntityIndividual, KYCEntitySelfEmployed:
		if !isValidCardNumber(pd.BankCardNumber) {
			return ErrInvalidInput
		}
		if pd.CardHolderName == "" {
			return ErrInvalidInput
		}
	case KYCEntitySoleProprietor:
		if !isValidINN(pd.INN, 12) {
			return ErrInvalidInput
		}
		if !isValidBankAccount(pd.BankAccount) {
			return ErrInvalidInput
		}
		if !isValidBIK(pd.BIK) {
			return ErrInvalidInput
		}
		if pd.BankName == "" {
			return ErrInvalidInput
		}
	case KYCEntityLegalEntity:
		if !isValidINN(pd.INN, 10) {
			return ErrInvalidInput
		}
		if !isValidBankAccount(pd.BankAccount) {
			return ErrInvalidInput
		}
		if !isValidBIK(pd.BIK) {
			return ErrInvalidInput
		}
		if !isValidBankAccount(pd.CorrespondentAccount) {
			return ErrInvalidInput
		}
		if pd.BankName == "" {
			return ErrInvalidInput
		}
	}

	return nil
}

func isValidCardNumber(card string) bool {
	return len(card) == 16 && digitsOnly.MatchString(card)
}

func isValidINN(inn string, length int) bool {
	return len(inn) == length && digitsOnly.MatchString(inn)
}

func isValidBankAccount(account string) bool {
	return len(account) == 20 && digitsOnly.MatchString(account)
}

func isValidBIK(bik string) bool {
	return len(bik) == 9 && digitsOnly.MatchString(bik)
}

// MaskCardNumber маскирует номер банковской карты для отображения
func MaskCardNumber(card string) string {
	if len(card) < 8 {
		return card
	}
	return card[:4] + "****" + card[len(card)-4:]
}

// MaskBankAccount маскирует номер банковского счёта для отображения
func MaskBankAccount(account string) string {
	if len(account) < 8 {
		return account
	}
	return account[:4] + "****" + account[len(account)-4:]
}
