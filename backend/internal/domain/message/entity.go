package message

type Message struct {
	ID             string
	ConversationID string
	Sender         string
	Receiver       string
	Body           string
	Status         string
	CreatedAtUnix  int64
	ReadAtUnix     *int64
}
