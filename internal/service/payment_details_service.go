package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type SetPaymentDetailsInput struct {
	EntityType           domain.KYCEntityType `json:"entity_type"`
	BankCardNumber       string               `json:"bank_card_number,omitempty"`
	CardHolderName       string               `json:"card_holder_name,omitempty"`
	BankAccount          string               `json:"bank_account,omitempty"`
	BIK                  string               `json:"bik,omitempty"`
	INN                  string               `json:"inn,omitempty"`
	CorrespondentAccount string               `json:"correspondent_account,omitempty"`
	BankName             string               `json:"bank_name,omitempty"`
}

type PaymentDetailsService interface {
	Set(ctx context.Context, userID uuid.UUID, input SetPaymentDetailsInput) (*domain.PaymentDetails, error)
	Get(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error)
	Validate(ctx context.Context, userID uuid.UUID) error
}

type paymentDetailsService struct {
	repo   repository.PaymentDetailsRepository
	logger *logger.Logger
}

func NewPaymentDetailsService(
	repo repository.PaymentDetailsRepository,
	log *logger.Logger,
) PaymentDetailsService {
	return &paymentDetailsService{
		repo:   repo,
		logger: log,
	}
}

func (s *paymentDetailsService) Set(ctx context.Context, userID uuid.UUID, input SetPaymentDetailsInput) (*domain.PaymentDetails, error) {
	now := time.Now()

	details := &domain.PaymentDetails{
		ID:                   uuid.New(),
		UserID:               userID,
		EntityType:           input.EntityType,
		BankCardNumber:       input.BankCardNumber,
		CardHolderName:       input.CardHolderName,
		BankAccount:          input.BankAccount,
		BIK:                  input.BIK,
		INN:                  input.INN,
		CorrespondentAccount: input.CorrespondentAccount,
		BankName:             input.BankName,
		IsVerified:           false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := details.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Upsert(ctx, details); err != nil {
		return nil, err
	}

	s.logger.Info("payment details set", "user_id", userID, "entity_type", input.EntityType)
	return details, nil
}

func (s *paymentDetailsService) Get(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *paymentDetailsService) Validate(ctx context.Context, userID uuid.UUID) error {
	details, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentDetailsNotFound) {
			return domain.ErrPaymentDetailsNotSet
		}
		return err
	}

	return details.Validate()
}
