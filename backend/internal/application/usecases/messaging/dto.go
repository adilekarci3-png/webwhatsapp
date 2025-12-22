package messaging

type SendMessageInput struct {
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Body           string `json:"body"`
	Receiver       string `json:"receiver"`
}

type MessageDTO struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Receiver       string `json:"receiver,omitempty"`
	Body           string `json:"body"`
	TS             int64  `json:"ts"`

	Status     string `json:"status,omitempty"`     // SENT / ACK / READ
	ReadAtUnix *int64 `json:"readAtUnix,omitempty"` // okundu zamanı
}
