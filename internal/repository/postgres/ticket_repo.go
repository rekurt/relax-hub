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

type ticketRepo struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) repository.TicketRepository {
	return &ticketRepo{pool: pool}
}

var ticketColumns = `id, user_id, booking_id, category, status, priority, level, subject, assigned_to, csat_score, created_at, updated_at, resolved_at`

func scanTicket(row pgx.Row) (*domain.Ticket, error) {
	var t domain.Ticket
	err := row.Scan(
		&t.ID, &t.UserID, &t.BookingID, &t.Category, &t.Status, &t.Priority,
		&t.Level, &t.Subject, &t.AssignedTo, &t.CSATScore,
		&t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *ticketRepo) Create(ctx context.Context, ticket *domain.Ticket) error {
	now := time.Now()
	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}
	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = now
	}
	if ticket.UpdatedAt.IsZero() {
		ticket.UpdatedAt = now
	}

	query := `INSERT INTO support_tickets (id, user_id, booking_id, category, status, priority, level, subject, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		ticket.ID, ticket.UserID, ticket.BookingID, ticket.Category,
		ticket.Status, ticket.Priority, ticket.Level, ticket.Subject,
		ticket.AssignedTo, ticket.CreatedAt, ticket.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}
	return nil
}

func (r *ticketRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_tickets WHERE id = $1`, ticketColumns)
	t, err := scanTicket(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return t, nil
}

func (r *ticketRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error) {
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE user_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count user tickets: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM support_tickets WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, ticketColumns)
	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list user tickets: %w", err)
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, *t)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.Ticket]{
		Items:      tickets,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ticketRepo) ListAll(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *filter.Priority)
		argIdx++
	}
	if filter.Level != nil {
		conditions = append(conditions, fmt.Sprintf("level = $%d", argIdx))
		args = append(args, *filter.Level)
		argIdx++
	}
	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *filter.Category)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM support_tickets %s`, where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count tickets: %w", err)
	}

	page := filter.Page
	pageSize := filter.PageSize
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT %s FROM support_tickets %s ORDER BY
		CASE priority
			WHEN 'critical' THEN 0
			WHEN 'high' THEN 1
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 3
		END ASC,
		created_at ASC
		LIMIT $%d OFFSET $%d`, ticketColumns, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, *t)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.Ticket]{
		Items:      tickets,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ticketRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TicketStatus) error {
	query := `UPDATE support_tickets SET status = $1, updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update ticket status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepo) UpdateLevel(ctx context.Context, id uuid.UUID, level domain.TicketLevel) error {
	query := `UPDATE support_tickets SET level = $1, status = 'escalated', updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, level, id)
	if err != nil {
		return fmt.Errorf("update ticket level: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepo) Assign(ctx context.Context, id uuid.UUID, assignedTo uuid.UUID) error {
	query := `UPDATE support_tickets SET assigned_to = $1, status = 'in_progress', updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, assignedTo, id)
	if err != nil {
		return fmt.Errorf("assign ticket: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepo) Resolve(ctx context.Context, id uuid.UUID, resolvedAt time.Time) error {
	query := `UPDATE support_tickets SET status = 'resolved', resolved_at = $1, updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, resolvedAt, id)
	if err != nil {
		return fmt.Errorf("resolve ticket: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepo) SubmitCSAT(ctx context.Context, id uuid.UUID, score int) error {
	query := `UPDATE support_tickets SET csat_score = $1, updated_at = now() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, score, id)
	if err != nil {
		return fmt.Errorf("submit CSAT: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepo) AddMessage(ctx context.Context, msg *domain.TicketMessage) error {
	now := time.Now()
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}

	query := `INSERT INTO ticket_messages (id, ticket_id, sender_id, sender_type, body, attachments, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	attachments := msg.Attachments
	if attachments == nil {
		attachments = []string{}
	}
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.TicketID, msg.SenderID, msg.SenderType, msg.Body, attachments, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("add ticket message: %w", err)
	}

	// Update ticket's updated_at
	_, _ = r.pool.Exec(ctx, `UPDATE support_tickets SET updated_at = now() WHERE id = $1`, msg.TicketID)

	return nil
}

func (r *ticketRepo) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	query := `SELECT id, ticket_id, sender_id, sender_type, body, attachments, created_at
		FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list ticket messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.TicketMessage
	for rows.Next() {
		var m domain.TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.SenderType, &m.Body, &m.Attachments, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ticket message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *ticketRepo) CountByStatus(ctx context.Context) (*domain.TicketStatusCounts, error) {
	query := `SELECT
		COALESCE(SUM(CASE WHEN status = 'open' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'escalated' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'closed' THEN 1 ELSE 0 END), 0)
		FROM support_tickets`

	var counts domain.TicketStatusCounts
	err := r.pool.QueryRow(ctx, query).Scan(
		&counts.Open, &counts.InProgress, &counts.Escalated, &counts.Resolved, &counts.Closed,
	)
	if err != nil {
		return nil, fmt.Errorf("count tickets by status: %w", err)
	}
	return &counts, nil
}

func (r *ticketRepo) ListStaleTickets(ctx context.Context, level domain.TicketLevel, olderThan time.Time) ([]domain.Ticket, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_tickets
		WHERE level = $1 AND status IN ('open', 'in_progress', 'escalated') AND updated_at < $2
		ORDER BY created_at ASC`, ticketColumns)

	rows, err := r.pool.Query(ctx, query, level, olderThan)
	if err != nil {
		return nil, fmt.Errorf("list stale tickets: %w", err)
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan stale ticket: %w", err)
		}
		tickets = append(tickets, *t)
	}
	return tickets, nil
}

func (r *ticketRepo) ListResolvedForAutoClose(ctx context.Context, resolvedBefore time.Time) ([]domain.Ticket, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_tickets
		WHERE status = 'resolved' AND resolved_at IS NOT NULL AND resolved_at < $1
		ORDER BY resolved_at ASC`, ticketColumns)

	rows, err := r.pool.Query(ctx, query, resolvedBefore)
	if err != nil {
		return nil, fmt.Errorf("list resolved for auto close: %w", err)
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan resolved ticket: %w", err)
		}
		tickets = append(tickets, *t)
	}
	return tickets, nil
}
