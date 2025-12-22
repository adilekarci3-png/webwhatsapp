package message

import "context"

type Repository interface {
	Insert(ctx context.Context, m Message) error
	ListByConversation(ctx context.Context, conversationID string, limit int) ([]Message, error)

	// ✅ Okundu bilgisi
	MarkRead(ctx context.Context, conversationID, receiver string, messageIDs []string, readAtUnix int64) error
}
