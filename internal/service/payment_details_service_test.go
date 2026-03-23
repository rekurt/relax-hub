package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type paymentDetailsTestEnv struct {
	svc  service.PaymentDetailsService
	repo *mock.PaymentDetailsRepo
}

func newPaymentDetailsTestEnv() *paymentDetailsTestEnv {
	repo := mock.NewPaymentDetailsRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewPaymentDetailsService(repo, log)
	return &paymentDetailsTestEnv{
		svc:  svc,
		repo: repo,
	}
}

func TestPaymentDetailsService_Set_Individual(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	details, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Иванов Иван Иванович",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.UserID != userID {
		t.Errorf("user_id = %v, want %v", details.UserID, userID)
	}
	if details.EntityType != domain.KYCEntityIndividual {
		t.Errorf("entity_type = %v, want individual", details.EntityType)
	}
	if details.BankCardNumber != "4111111111111111" {
		t.Errorf("bank_card_number = %v, want 4111111111111111", details.BankCardNumber)
	}
}

func TestPaymentDetailsService_Set_SelfEmployed(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	details, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntitySelfEmployed,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Петров Пётр",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.EntityType != domain.KYCEntitySelfEmployed {
		t.Errorf("entity_type = %v, want self_employed", details.EntityType)
	}
}

func TestPaymentDetailsService_Set_SoleProprietor(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	details, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:  domain.KYCEntitySoleProprietor,
		INN:         "123456789012",
		BankAccount: "40802810000000000001",
		BIK:         "044525225",
		BankName:    "ПАО Сбербанк",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.EntityType != domain.KYCEntitySoleProprietor {
		t.Errorf("entity_type = %v, want sole_proprietor", details.EntityType)
	}
	if details.BankAccount != "40802810000000000001" {
		t.Errorf("bank_account = %v, want 40802810000000000001", details.BankAccount)
	}
}

func TestPaymentDetailsService_Set_LegalEntity(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	details, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:           domain.KYCEntityLegalEntity,
		INN:                  "1234567890",
		BankAccount:          "40702810000000000001",
		BIK:                  "044525225",
		CorrespondentAccount: "30101810400000000225",
		BankName:             "ПАО Сбербанк",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.EntityType != domain.KYCEntityLegalEntity {
		t.Errorf("entity_type = %v, want legal_entity", details.EntityType)
	}
}

func TestPaymentDetailsService_Set_InvalidEntityType(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType: "invalid",
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestPaymentDetailsService_Set_Individual_MissingCard(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		CardHolderName: "Иванов Иван",
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestPaymentDetailsService_Set_Individual_InvalidCardLength(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "411111111111",
		CardHolderName: "Иванов Иван",
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for short card, got: %v", err)
	}
}

func TestPaymentDetailsService_Set_SoleProprietor_MissingINN(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:  domain.KYCEntitySoleProprietor,
		BankAccount: "40802810000000000001",
		BIK:         "044525225",
		BankName:    "Банк",
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestPaymentDetailsService_Set_LegalEntity_MissingCorrespondentAccount(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:  domain.KYCEntityLegalEntity,
		INN:         "1234567890",
		BankAccount: "40702810000000000001",
		BIK:         "044525225",
		BankName:    "Банк",
	})
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestPaymentDetailsService_Get_Success(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Иванов Иван",
	})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	details, err := env.svc.Get(context.Background(), userID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if details.UserID != userID {
		t.Errorf("user_id = %v, want %v", details.UserID, userID)
	}
}

func TestPaymentDetailsService_Get_NotFound(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Get(context.Background(), userID)
	if err != domain.ErrPaymentDetailsNotFound {
		t.Fatalf("expected ErrPaymentDetailsNotFound, got: %v", err)
	}
}

func TestPaymentDetailsService_Set_Upsert(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Иванов Иван",
	})
	if err != nil {
		t.Fatalf("first set failed: %v", err)
	}

	details2, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "5555555555554444",
		CardHolderName: "Иванов Иван Петрович",
	})
	if err != nil {
		t.Fatalf("second set (upsert) failed: %v", err)
	}
	if details2.BankCardNumber != "5555555555554444" {
		t.Errorf("card not updated: got %v", details2.BankCardNumber)
	}

	got, err := env.svc.Get(context.Background(), userID)
	if err != nil {
		t.Fatalf("get after upsert failed: %v", err)
	}
	if got.BankCardNumber != "5555555555554444" {
		t.Errorf("card not persisted after upsert: got %v", got.BankCardNumber)
	}
}

func TestPaymentDetailsService_Validate_Success(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	_, err := env.svc.Set(context.Background(), userID, service.SetPaymentDetailsInput{
		EntityType:     domain.KYCEntityIndividual,
		BankCardNumber: "4111111111111111",
		CardHolderName: "Иванов",
	})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	err = env.svc.Validate(context.Background(), userID)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

func TestPaymentDetailsService_Validate_NotSet(t *testing.T) {
	env := newPaymentDetailsTestEnv()
	userID := uuid.New()

	err := env.svc.Validate(context.Background(), userID)
	if err != domain.ErrPaymentDetailsNotSet {
		t.Fatalf("expected ErrPaymentDetailsNotSet, got: %v", err)
	}
}
