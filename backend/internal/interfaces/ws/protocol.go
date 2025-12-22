package ws

import "encoding/json"

type Event struct {
	Type           string          `json:"type"`
	ConversationID string          `json:"conversationId,omitempty"`
	Sender         string          `json:"sender,omitempty"`
	Receiver       string          `json:"receiver,omitempty"`
	MessageID      string          `json:"messageId,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
}

type PresencePayload struct {
	Online     bool  `json:"online"`
	LastSeenAt int64 `json:"lastSeenAt,omitempty"`
}

type TypingPayload struct {
	IsTyping bool `json:"isTyping"`
}

type ReadPayload struct {
	MessageIDs []string `json:"messageIds"`
	ReadAt     int64    `json:"readAt"`
}
