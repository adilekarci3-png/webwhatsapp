package message

type Message struct {
	ID             string
	ConversationID string
	Sender         string
	Body           string
	CreatedAtUnix  int64
}
