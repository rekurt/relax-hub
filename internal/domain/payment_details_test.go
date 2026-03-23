package domain_test

import (
	"testing"

	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestPaymentDetails_Validate_Individual(t *testing.T) {
	tests := []struct {
		name    string
		pd      domain.PaymentDetails
		wantErr bool
	}{
		{
			name: "valid individual",
			pd: domain.PaymentDetails{
				EntityType:     domain.KYCEntityIndividual,
				BankCardNumber: "4111111111111111",
				CardHolderName: "Иванов Иван",
			},
			wantErr: false,
		},
		{
			name: "missing card number",
			pd: domain.PaymentDetails{
				EntityType:     domain.KYCEntityIndividual,
				CardHolderName: "Иванов",
			},
			wantErr: true,
		},
		{
			name: "invalid card length",
			pd: domain.PaymentDetails{
				EntityType:     domain.KYCEntityIndividual,
				BankCardNumber: "411111111111",
				CardHolderName: "Иванов",
			},
			wantErr: true,
		},
		{
			name: "card with non-digits",
			pd: domain.PaymentDetails{
				EntityType:     domain.KYCEntityIndividual,
				BankCardNumber: "4111-1111-1111-1111",
				CardHolderName: "Иванов",
			},
			wantErr: true,
		},
		{
			name: "missing card holder",
			pd: domain.PaymentDetails{
				EntityType:     domain.KYCEntityIndividual,
				BankCardNumber: "4111111111111111",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentDetails_Validate_SoleProprietor(t *testing.T) {
	valid := domain.PaymentDetails{
		EntityType:  domain.KYCEntitySoleProprietor,
		INN:         "123456789012",
		BankAccount: "40802810000000000001",
		BIK:         "044525225",
		BankName:    "Сбербанк",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid sole_proprietor failed: %v", err)
	}

	// Missing INN
	noINN := valid
	noINN.INN = ""
	if err := noINN.Validate(); err == nil {
		t.Error("expected error for missing INN")
	}

	// Wrong INN length (10 instead of 12)
	wrongINN := valid
	wrongINN.INN = "1234567890"
	if err := wrongINN.Validate(); err == nil {
		t.Error("expected error for 10-digit INN for sole_proprietor")
	}

	// Missing BIK
	noBIK := valid
	noBIK.BIK = ""
	if err := noBIK.Validate(); err == nil {
		t.Error("expected error for missing BIK")
	}
}

func TestPaymentDetails_Validate_LegalEntity(t *testing.T) {
	valid := domain.PaymentDetails{
		EntityType:           domain.KYCEntityLegalEntity,
		INN:                  "1234567890",
		BankAccount:          "40702810000000000001",
		BIK:                  "044525225",
		CorrespondentAccount: "30101810400000000225",
		BankName:             "Сбербанк",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid legal_entity failed: %v", err)
	}

	// Missing correspondent account
	noCorrAcc := valid
	noCorrAcc.CorrespondentAccount = ""
	if err := noCorrAcc.Validate(); err == nil {
		t.Error("expected error for missing correspondent account")
	}

	// Wrong INN length (12 instead of 10)
	wrongINN := valid
	wrongINN.INN = "123456789012"
	if err := wrongINN.Validate(); err == nil {
		t.Error("expected error for 12-digit INN for legal_entity")
	}
}

func TestPaymentDetails_Validate_InvalidEntityType(t *testing.T) {
	pd := domain.PaymentDetails{
		EntityType: "invalid",
	}
	if err := pd.Validate(); err == nil {
		t.Error("expected error for invalid entity type")
	}
}

func TestMaskCardNumber(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"4111111111111111", "4111****1111"},
		{"short", "short"},
	}
	for _, tt := range tests {
		got := domain.MaskCardNumber(tt.input)
		if got != tt.want {
			t.Errorf("MaskCardNumber(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskBankAccount(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"40802810000000000001", "4080****0001"},
		{"short", "short"},
	}
	for _, tt := range tests {
		got := domain.MaskBankAccount(tt.input)
		if got != tt.want {
			t.Errorf("MaskBankAccount(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
