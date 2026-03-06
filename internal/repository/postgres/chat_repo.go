package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// conversationRepo

type conversationRepo struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) repository.ConversationRepository {
	return &conversationRepo{pool: pool}
}

var conversationColumns = `id, bathhouse_id, client_id, booking_id, last_message_at, created_at`

func scanConversation(row pgx.Row) (*domain.Conversation, error) {
	var c domain.Conversation
	err := row.Scan(&c.ID, &c.BathhouseID, &c.ClientID, &c.BookingID, &c.LastMessageAt, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *conversationRepo) Create(ctx context.Context, conv *domain.Conversation) error {
	if conv.ID == uuid.Nil {
		conv.ID = uuid.New()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO conversations (id, bathhouse_id, client_id, booking_id, last_message_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		conv.ID, conv.BathhouseID, conv.ClientID, conv.BookingID, conv.LastMessageAt, conv.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}
	return nil
}

func (r *conversationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	query := `SELECT ` + conversationColumns + ` FROM conversations WHERE id = $1`
	c, err := scanConversation(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get conversation by id: %w", err)
	}
	return c, nil
}

func (r *conversationRepo) GetByParticipants(ctx context.Context, bathhouseID, clientID uuid.UUID) (*domain.Conversation, error) {
	query := `SELECT ` + conversationColumns + ` FROM conversations WHERE bathhouse_id = $1 AND client_id = $2`
	c, err := scanConversation(r.pool.QueryRow(ctx, query, bathhouseID, clientID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get conversation by participants: %w", err)
	}
	return c, nil
}

func (r *conversationRepo) ListByUser(ctx context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var countQuery, listQuery string
	var args []interface{}

	if len(bathhouseIDs) > 0 {
		// Owner/representative: conversations for their bathhouses OR where they are a client
		countQuery = `SELECT COUNT(*) FROM conversations WHERE bathhouse_id = ANY($1) OR client_id = $2`
		args = append(args, bathhouseIDs, userID)
		listQuery = `SELECT ` + conversationColumns + ` FROM conversations WHERE bathhouse_id = ANY($1) OR client_id = $2 ORDER BY last_message_at DESC NULLS LAST, created_at DESC LIMIT $3 OFFSET $4`
	} else {
		// Client: conversations where they are the client
		countQuery = `SELECT COUNT(*) FROM conversations WHERE client_id = $1`
		args = append(args, userID)
		listQuery = `SELECT ` + conversationColumns + ` FROM conversations WHERE client_id = $1 ORDER BY last_message_at DESC NULLS LAST, created_at DESC LIMIT $2 OFFSET $3`
	}

	var totalCount int64
	if len(bathhouseIDs) > 0 {
		if err := r.pool.QueryRow(ctx, countQuery, args[0], args[1]).Scan(&totalCount); err != nil {
			return nil, fmt.Errorf("count conversations: %w", err)
		}
	} else {
		if err := r.pool.QueryRow(ctx, countQuery, args[0]).Scan(&totalCount); err != nil {
			return nil, fmt.Errorf("count conversations: %w", err)
		}
	}

	offset := (page - 1) * pageSize
	var rows pgx.Rows
	var err error
	if len(bathhouseIDs) > 0 {
		rows, err = r.pool.Query(ctx, listQuery, args[0], args[1], pageSize, offset)
	} else {
		rows, err = r.pool.Query(ctx, listQuery, args[0], pageSize, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		if err := rows.Scan(&c.ID, &c.BathhouseID, &c.ClientID, &c.BookingID, &c.LastMessageAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversation rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Conversation]{
		Items:      conversations,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *conversationRepo) ListAll(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	countQuery := `SELECT COUNT(*) FROM conversations`
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count all conversations: %w", err)
	}

	offset := (page - 1) * pageSize
	listQuery := `SELECT ` + conversationColumns + ` FROM conversations ORDER BY last_message_at DESC NULLS LAST, created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, listQuery, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list all conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		if err := rows.Scan(&c.ID, &c.BathhouseID, &c.ClientID, &c.BookingID, &c.LastMessageAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversation rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Conversation]{
		Items:      conversations,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *conversationRepo) GetOrCreate(ctx context.Context, conv *domain.Conversation) (*domain.Conversation, error) {
	if conv.ID == uuid.Nil {
		conv.ID = uuid.New()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}

	// Use INSERT ... ON CONFLICT to atomically get-or-create
	query := `
		WITH inserted AS (
			INSERT INTO conversations (id, bathhouse_id, client_id, booking_id, last_message_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (bathhouse_id, client_id) DO NOTHING
			RETURNING ` + conversationColumns + `
		)
		SELECT ` + conversationColumns + ` FROM inserted
		UNION ALL
		SELECT ` + conversationColumns + ` FROM conversations
		WHERE bathhouse_id = $2 AND client_id = $3
		AND NOT EXISTS (SELECT 1 FROM inserted)
		LIMIT 1`

	c, err := scanConversation(r.pool.QueryRow(ctx, query,
		conv.ID, conv.BathhouseID, conv.ClientID, conv.BookingID, conv.LastMessageAt, conv.CreatedAt,
	))
	if err != nil {
		return nil, fmt.Errorf("get or create conversation: %w", err)
	}
	return c, nil
}

func (r *conversationRepo) UpdateLastMessageAt(ctx context.Context, id uuid.UUID, t time.Time) error {
	query := `UPDATE conversations SET last_message_at = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, t)
	if err != nil {
		return fmt.Errorf("update last_message_at: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// messageRepo

type messageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) repository.MessageRepository {
	return &messageRepo{pool: pool}
}

var messageColumns = `id, conversation_id, sender_id, text, is_read, read_at, created_at`

func (r *messageRepo) Create(ctx context.Context, msg *domain.Message) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, text, is_read, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.ConversationID, msg.SenderID, msg.Text, msg.IsRead, msg.ReadAt, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	return nil
}

func (r *messageRepo) ListByConversation(ctx context.Context, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`, conversationID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + messageColumns + ` FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, conversationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Text, &m.IsRead, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Message]{
		Items:      messages,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *messageRepo) MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE messages SET is_read = true, read_at = $3 WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false`
	_, err := r.pool.Exec(ctx, query, conversationID, userID, now)
	if err != nil {
		return fmt.Errorf("mark messages as read: %w", err)
	}
	return nil
}

func (r *messageRepo) CountUnread(ctx context.Context, userID uuid.UUID, conversationIDs []uuid.UUID) (int64, error) {
	if len(conversationIDs) == 0 {
		return 0, nil
	}

	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = ANY($1) AND sender_id != $2 AND is_read = false`
	var count int64
	err := r.pool.QueryRow(ctx, query, conversationIDs, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread messages: %w", err)
	}
	return count, nil
}
