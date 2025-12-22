package message

import "context"

type Repository interface {
	Insert(ctx context.Context, m Message) error
	ListByConversation(ctx context.Context, conversationID string, limit int) ([]Message, error)
}
