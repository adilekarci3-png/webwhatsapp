package postgres

import (
	"context"
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

func normalizeStatus(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "SENT", "ACK", "READ":
		return s
	default:
		return "SENT"
	}
}

func (r *MessageRepo) Insert(ctx context.Context, m message.Message) error {
	status := normalizeStatus(m.Status)

	// receiver boşsa DB'de NULL tutmak daha sağlıklı
	var receiver *string
	if strings.TrimSpace(m.Receiver) != "" {
		rcv := strings.TrimSpace(m.Receiver)
		receiver = &rcv
	}

	// created_at_unix 0 gelirse server set etsin diye burada zorlamayalım;
	// ama 0 istemiyorsanız burada kontrol edip set edebilirsiniz.
	_, err := r.pool.Exec(ctx, `
		INSERT INTO public.messages(
			id, conversation_id, sender, receiver, body, status, created_at_unix
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, m.ID, m.ConversationID, m.Sender, receiver, m.Body, status, m.CreatedAtUnix)

	return err
}

func (r *MessageRepo) ListByConversation(ctx context.Context, conversationID string, limit int) ([]message.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			conversation_id,
			sender,
			receiver,                 -- NULL gelebilir
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
		var receiver *string
		var readAt *int64

		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.Sender,
			&receiver,
			&m.Body,
			&m.Status,
			&m.CreatedAtUnix,
			&readAt,
		); err != nil {
			return nil, err
		}

		if receiver != nil {
			m.Receiver = *receiver
		} else {
			m.Receiver = ""
		}
		m.ReadAtUnix = readAt

		out = append(out, m)
	}

	return out, rows.Err()
}

// receiver = "bu cihazın/oturumun kullanıcısı" olacak şekilde çağrılmalı.
// Yani: adile tarafı okuduysa receiver="adile" geçilmeli.
func (r *MessageRepo) MarkRead(ctx context.Context, convID, receiver string, messageIDs []string, readAt int64) error {
	if len(messageIDs) == 0 {
		return nil
	}

	receiver = strings.TrimSpace(receiver)
	if receiver == "" {
		// receiver boşsa kimin okuduğu belli değil; bu durumda update yapmayın.
		return nil
	}

	// Daha temiz: id = ANY($4::text[])
	_, err := r.pool.Exec(ctx, `
		UPDATE public.messages
		SET status = 'READ', read_at_unix = $1
		WHERE conversation_id = $2
		  AND receiver = $3
		  AND status <> 'READ'
		  AND id = ANY($4::text[])
	`, readAt, convID, receiver, messageIDs)

	return err
}

func (r *MessageRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE public.messages
		SET status = $2
		WHERE id = $1
	`, id, status)
	return err
}
