package messaging

type SendMessageInput struct {
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Body           string `json:"body"`
}

type MessageDTO struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Body           string `json:"body"`
	TS             int64  `json:"ts"`
}
