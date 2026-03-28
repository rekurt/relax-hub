package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type disputeRepo struct {
	pool *pgxpool.Pool
}

func NewDisputeRepository(pool *pgxpool.Pool) repository.DisputeRepository {
	return &disputeRepo{pool: pool}
}

var disputeColumns = `id, booking_id, initiator_id, respondent_id, reason, description, status, resolution, refund_amount, compensation_amount, mediator_id, mediator_notes, appeal_deadline, evidence_deadline, created_at, updated_at, resolved_at`

func scanDispute(row pgx.Row) (*domain.Dispute, error) {
	var d domain.Dispute
	err := row.Scan(
		&d.ID, &d.BookingID, &d.InitiatorID, &d.RespondentID,
		&d.Reason, &d.Description, &d.Status, &d.Resolution,
		&d.RefundAmount, &d.CompensationAmount, &d.MediatorID, &d.MediatorNotes,
		&d.AppealDeadline, &d.EvidenceDeadline,
		&d.CreatedAt, &d.UpdatedAt, &d.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *disputeRepo) Create(ctx context.Context, dispute *domain.Dispute) error {
	now := time.Now()
	if dispute.ID == uuid.Nil {
		dispute.ID = uuid.New()
	}
	if dispute.CreatedAt.IsZero() {
		dispute.CreatedAt = now
	}
	if dispute.UpdatedAt.IsZero() {
		dispute.UpdatedAt = now
	}

	query := `INSERT INTO disputes (id, booking_id, initiator_id, respondent_id, reason, description, status, evidence_deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		dispute.ID, dispute.BookingID, dispute.InitiatorID, dispute.RespondentID,
		dispute.Reason, dispute.Description, dispute.Status, dispute.EvidenceDeadline,
		dispute.CreatedAt, dispute.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create dispute: %w", err)
	}
	return nil
}

func (r *disputeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dispute, error) {
	query := fmt.Sprintf(`SELECT %s FROM disputes WHERE id = $1`, disputeColumns)
	d, err := scanDispute(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDisputeNotFound
		}
		return nil, fmt.Errorf("get dispute: %w", err)
	}
	return d, nil
}

func (r *disputeRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Dispute, error) {
	query := fmt.Sprintf(`SELECT %s FROM disputes WHERE booking_id = $1`, disputeColumns)
	d, err := scanDispute(r.pool.QueryRow(ctx, query, bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDisputeNotFound
		}
		return nil, fmt.Errorf("get dispute by booking: %w", err)
	}
	return d, nil
}

func (r *disputeRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DisputeStatus) error {
	query := `UPDATE disputes SET status = $1, updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update dispute status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDisputeNotFound
	}
	return nil
}

func (r *disputeRepo) UpdateResolution(ctx context.Context, id uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string, resolvedAt time.Time) error {
	query := `UPDATE disputes SET status = 'resolved', resolution = $1, refund_amount = $2, compensation_amount = $3, mediator_notes = $4, resolved_at = $5, appeal_deadline = $6, updated_at = now() WHERE id = $7`
	appealDeadline := resolvedAt.Add(7 * 24 * time.Hour)
	ct, err := r.pool.Exec(ctx, query, resolution, refundAmount, compensationAmount, mediatorNotes, resolvedAt, appealDeadline, id)
	if err != nil {
		return fmt.Errorf("update dispute resolution: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDisputeNotFound
	}
	return nil
}

func (r *disputeRepo) UpdateAppeal(ctx context.Context, id uuid.UUID, status domain.DisputeStatus) error {
	query := `UPDATE disputes SET status = $1, updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update dispute appeal: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDisputeNotFound
	}
	return nil
}

func (r *disputeRepo) Assign(ctx context.Context, id uuid.UUID, mediatorID uuid.UUID) error {
	query := `UPDATE disputes SET mediator_id = $1, status = 'under_review', updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, mediatorID, id)
	if err != nil {
		return fmt.Errorf("assign dispute: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDisputeNotFound
	}
	return nil
}

func (r *disputeRepo) AddEvidence(ctx context.Context, evidence *domain.DisputeEvidence) error {
	now := time.Now()
	if evidence.ID == uuid.Nil {
		evidence.ID = uuid.New()
	}
	if evidence.CreatedAt.IsZero() {
		evidence.CreatedAt = now
	}

	query := `INSERT INTO dispute_evidence (id, dispute_id, user_id, type, url, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		evidence.ID, evidence.DisputeID, evidence.UserID,
		evidence.Type, evidence.URL, evidence.Description, evidence.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("add dispute evidence: %w", err)
	}
	return nil
}

func (r *disputeRepo) ListEvidence(ctx context.Context, disputeID uuid.UUID) ([]domain.DisputeEvidence, error) {
	query := `SELECT id, dispute_id, user_id, type, url, description, created_at
		FROM dispute_evidence WHERE dispute_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, disputeID)
	if err != nil {
		return nil, fmt.Errorf("list dispute evidence: %w", err)
	}
	defer rows.Close()

	var evidence []domain.DisputeEvidence
	for rows.Next() {
		var e domain.DisputeEvidence
		if err := rows.Scan(&e.ID, &e.DisputeID, &e.UserID, &e.Type, &e.URL, &e.Description, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan dispute evidence: %w", err)
		}
		evidence = append(evidence, e)
	}
	return evidence, nil
}

func (r *disputeRepo) ListAll(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("(initiator_id = $%d OR respondent_id = $%d)", argIdx, argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM disputes %s`, where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count disputes: %w", err)
	}

	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT %s FROM disputes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, disputeColumns, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list disputes: %w", err)
	}
	defer rows.Close()

	var disputes []domain.Dispute
	for rows.Next() {
		d, err := scanDispute(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dispute: %w", err)
		}
		disputes = append(disputes, *d)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.Dispute]{
		Items:      disputes,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *disputeRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error) {
	countQuery := `SELECT COUNT(*) FROM disputes WHERE initiator_id = $1 OR respondent_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count user disputes: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT %s FROM disputes WHERE initiator_id = $1 OR respondent_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, disputeColumns)
	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list user disputes: %w", err)
	}
	defer rows.Close()

	var disputes []domain.Dispute
	for rows.Next() {
		d, err := scanDispute(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dispute: %w", err)
		}
		disputes = append(disputes, *d)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.Dispute]{
		Items:      disputes,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *disputeRepo) CountOpenByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM disputes
		WHERE (initiator_id = $1 OR respondent_id = $1)
		AND status NOT IN ('closed', 'resolved')`

	var count int64
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count open disputes by user: %w", err)
	}
	return count, nil
}
