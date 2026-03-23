package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type paymentDetailsRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentDetailsRepository(pool *pgxpool.Pool) repository.PaymentDetailsRepository {
	return &paymentDetailsRepo{pool: pool}
}

var paymentDetailsColumns = `id, user_id, entity_type, bank_card_number, card_holder_name, bank_account, bik, inn, correspondent_account, bank_name, is_verified, created_at, updated_at`

func scanPaymentDetails(row pgx.Row) (*domain.PaymentDetails, error) {
	var pd domain.PaymentDetails
	err := row.Scan(
		&pd.ID, &pd.UserID, &pd.EntityType,
		&pd.BankCardNumber, &pd.CardHolderName,
		&pd.BankAccount, &pd.BIK, &pd.INN,
		&pd.CorrespondentAccount, &pd.BankName,
		&pd.IsVerified, &pd.CreatedAt, &pd.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &pd, nil
}

func (r *paymentDetailsRepo) Upsert(ctx context.Context, details *domain.PaymentDetails) error {
	query := `INSERT INTO owner_payment_details (id, user_id, entity_type, bank_card_number, card_holder_name, bank_account, bik, inn, correspondent_account, bank_name, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id) DO UPDATE SET
			entity_type = EXCLUDED.entity_type,
			bank_card_number = EXCLUDED.bank_card_number,
			card_holder_name = EXCLUDED.card_holder_name,
			bank_account = EXCLUDED.bank_account,
			bik = EXCLUDED.bik,
			inn = EXCLUDED.inn,
			correspondent_account = EXCLUDED.correspondent_account,
			bank_name = EXCLUDED.bank_name,
			updated_at = EXCLUDED.updated_at`
	_, err := r.pool.Exec(ctx, query,
		details.ID, details.UserID, details.EntityType,
		details.BankCardNumber, details.CardHolderName,
		details.BankAccount, details.BIK, details.INN,
		details.CorrespondentAccount, details.BankName,
		details.IsVerified, details.CreatedAt, details.UpdatedAt,
	)
	return err
}

func (r *paymentDetailsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error) {
	query := `SELECT ` + paymentDetailsColumns + ` FROM owner_payment_details WHERE user_id = $1`
	pd, err := scanPaymentDetails(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentDetailsNotFound
		}
		return nil, err
	}
	return pd, nil
}
