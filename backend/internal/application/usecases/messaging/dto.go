package messaging

// REST input (send endpoint veya WS send için kullanılabilir)
type SendMessageInput struct {
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Body           string `json:"body"`
	Receiver       string `json:"receiver,omitempty"`
	TS             int64  `json:"ts,omitempty"`

	ClientMsgID string `json:"clientMsgId,omitempty"` // ✅
}

type MessageDTO struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Sender         string `json:"sender"`
	Receiver       string `json:"receiver,omitempty"`
	Body           string `json:"body"`
	TS             int64  `json:"ts"`

	Status     string `json:"status,omitempty"`
	ReadAtUnix *int64 `json:"readAtUnix,omitempty"`

	ClientMsgID string `json:"clientMsgId,omitempty"` // ✅
}

// WS Envelope (type + metadata + opsiyonel payload)
type WsEvent struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversationId,omitempty"`
	Sender         string `json:"sender,omitempty"`
	Receiver       string `json:"receiver,omitempty"`
	Payload        any    `json:"payload,omitempty"`
}

// Read event payload (Vue: type:"message.read", payload:{messageIds, readAt})
type ReadPayload struct {
	MessageIDs []string `json:"messageIds"`
	ReadAt     int64    `json:"readAt"`
}

// Typing payload (Vue tarafını buna çekmenizi öneririm)
type TypingPayload struct {
	IsTyping bool `json:"isTyping"`
}
