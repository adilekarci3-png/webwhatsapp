package postgres

import (
	"context"

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
	_, err := r.pool.Exec(ctx, `
		INSERT INTO messages(id, conversation_id, sender, body, created_at_unix)
		VALUES ($1,$2,$3,$4,$5)
	`, m.ID, m.ConversationID, m.Sender, m.Body, m.CreatedAtUnix)
	return err
}

func (r *MessageRepo) ListByConversation(ctx context.Context, conversationID string, limit int) ([]message.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, sender, body, created_at_unix
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at_unix DESC
		LIMIT $2
	`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []message.Message
	for rows.Next() {
		var m message.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Sender, &m.Body, &m.CreatedAtUnix); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
