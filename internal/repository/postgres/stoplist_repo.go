package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type stoplistRepo struct {
	pool *pgxpool.Pool
}

func NewStoplistRepository(pool *pgxpool.Pool) repository.StoplistRepository {
	return &stoplistRepo{pool: pool}
}

func (r *stoplistRepo) Create(ctx context.Context, entry *domain.StoplistEntry) error {
	query := `INSERT INTO antifraud_stoplist (phone, email, inn, bank_card_number, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, blocked_at`

	var phone, email, inn, bankCard *string
	if entry.Phone != "" {
		phone = &entry.Phone
	}
	if entry.Email != "" {
		email = &entry.Email
	}
	if entry.INN != "" {
		inn = &entry.INN
	}
	if entry.BankCardNumber != "" {
		bankCard = &entry.BankCardNumber
	}

	return r.pool.QueryRow(ctx, query,
		phone, email, inn, bankCard, entry.Reason, entry.CreatedBy,
	).Scan(&entry.ID, &entry.BlockedAt)
}

func (r *stoplistRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM antifraud_stoplist WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *stoplistRepo) List(ctx context.Context, filter domain.StoplistFilter) (*domain.PaginatedResult[domain.StoplistEntry], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, filter.Phone)
		argIdx++
	}
	if filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, filter.Email)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM antifraud_stoplist %s`, where)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, phone, email, inn, bank_card_number, reason, blocked_at, created_by
		FROM antifraud_stoplist %s
		ORDER BY blocked_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.StoplistEntry
	for rows.Next() {
		var e domain.StoplistEntry
		var phone, email, inn, bankCard *string
		if err := rows.Scan(&e.ID, &phone, &email, &inn, &bankCard, &e.Reason, &e.BlockedAt, &e.CreatedBy); err != nil {
			return nil, err
		}
		if phone != nil {
			e.Phone = *phone
		}
		if email != nil {
			e.Email = *email
		}
		if inn != nil {
			e.INN = *inn
		}
		if bankCard != nil {
			e.BankCardNumber = *bankCard
		}
		items = append(items, e)
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	}

	return &domain.PaginatedResult[domain.StoplistEntry]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func (r *stoplistRepo) IsBlocked(ctx context.Context, phone, email, inn, bankCardNumber string) (bool, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, phone)
		argIdx++
	}
	if email != "" {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, email)
		argIdx++
	}
	if inn != "" {
		conditions = append(conditions, fmt.Sprintf("inn = $%d", argIdx))
		args = append(args, inn)
		argIdx++
	}
	if bankCardNumber != "" {
		conditions = append(conditions, fmt.Sprintf("bank_card_number = $%d", argIdx))
		args = append(args, bankCardNumber)
		argIdx++
	}

	if len(conditions) == 0 {
		return false, nil
	}

	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM antifraud_stoplist WHERE %s)`,
		strings.Join(conditions, " OR "))

	var exists bool
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *stoplistRepo) CountDuplicateOwners(ctx context.Context, phone, email, inn string, excludeUserID uuid.UUID) (int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 2 // $1 is excludeUserID

	args = append(args, excludeUserID)

	if phone != "" {
		conditions = append(conditions, fmt.Sprintf("u.phone = $%d", argIdx))
		args = append(args, phone)
		argIdx++
	}
	if email != "" {
		conditions = append(conditions, fmt.Sprintf("u.email = $%d", argIdx))
		args = append(args, email)
		argIdx++
	}
	if inn != "" {
		conditions = append(conditions, fmt.Sprintf("pd.inn = $%d", argIdx))
		args = append(args, inn)
		argIdx++
	}

	if len(conditions) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf(`SELECT COUNT(DISTINCT u.id) FROM users u
		LEFT JOIN payment_details pd ON pd.user_id = u.id
		WHERE u.id != $1
		AND u.role = 'owner'
		AND (%s)`, strings.Join(conditions, " OR "))

	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
