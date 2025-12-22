package postgres

import (
	"context"
	"strconv"
	"strings"

	"example.com/webwhatsapp/backend/internal/domain/message"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (r *MessageRepo) Insert(ctx context.Context, m message.Message) error {
	status := strings.TrimSpace(m.Status)
	if status == "" {
		status = "SENT"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO public.messages(id, conversation_id, sender, receiver, body, status, created_at_unix)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, m.ID, m.ConversationID, m.Sender, m.Receiver, m.Body, status, m.CreatedAtUnix)

	return err
}

func (r *MessageRepo) ListByConversation(ctx context.Context, conversationID string, limit int) ([]message.Message, error) {
	rows, err := r.pool.Query(ctx, `
  SELECT
    id,
    conversation_id,
    sender,
    COALESCE(receiver, '') AS receiver,
    body,
    COALESCE(status, 'SENT') AS status,
    created_at_unix,
    read_at_unix
  FROM public.messages
  WHERE conversation_id = $1
  ORDER BY created_at_unix DESC
  LIMIT $2
`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]message.Message, 0, limit)
	for rows.Next() {
		var m message.Message
		var readAt *int64 // ✅ NULL gelirse nil olur

		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.Sender,
			&m.Receiver,
			&m.Body,
			&m.Status,
			&m.CreatedAtUnix,
			&readAt,
		); err != nil {
			return nil, err
		}

		m.ReadAtUnix = readAt // ✅ direkt ata
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MessageRepo) MarkRead(ctx context.Context, convID, receiver string, messageIDs []string, readAt int64) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// $1=readAt, $2=convID, $3=receiver, ids $4..$N
	placeholders := make([]string, 0, len(messageIDs))
	args := make([]any, 0, 3+len(messageIDs))
	args = append(args, readAt, convID, receiver)

	for i, id := range messageIDs {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+4))
		args = append(args, id)
	}

	q := `
		UPDATE public.messages
		SET status='READ', read_at_unix=$1
		WHERE conversation_id=$2
		  AND receiver=$3
		  AND id IN (` + strings.Join(placeholders, ",") + `)
	`

	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
