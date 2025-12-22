package messaging

import (
	"context"
	"encoding/json"
	"strings"

	"example.com/webwhatsapp/backend/internal/application/ports"
	"example.com/webwhatsapp/backend/internal/domain/common"
	"example.com/webwhatsapp/backend/internal/domain/message"
)

type Service struct {
	repo message.Repository
	pub  ports.Publisher
}

func NewService(repo message.Repository, pub ports.Publisher) *Service {
	return &Service{repo: repo, pub: pub}
}

func roomChannel(conversationID string) string {
	return "room:" + conversationID
}

func (s *Service) SendMessage(ctx context.Context, in SendMessageInput) (MessageDTO, error) {
	in.ConversationID = strings.TrimSpace(in.ConversationID)
	in.Sender = strings.TrimSpace(in.Sender)
	in.Receiver = strings.TrimSpace(in.Receiver) // ✅
	in.Body = strings.TrimSpace(in.Body)

	if in.ConversationID == "" || in.Sender == "" || in.Body == "" {
		return MessageDTO{}, common.ErrInvalidInput
	}

	// 1-1 için receiver zorunlu yapmak istersen bunu aç:
	// if in.Receiver == "" { return MessageDTO{}, common.ErrInvalidInput }

	m := message.Message{
		ID:             common.NewID(),
		ConversationID: in.ConversationID,
		Sender:         in.Sender,
		Receiver:       in.Receiver, // ✅ Message struct'ına eklenecek
		Body:           in.Body,
		Status:         "SENT", // ✅ Message struct'ına eklenecek (istersen default DB)
		CreatedAtUnix:  common.NowUnix(),
	}

	if err := s.repo.Insert(ctx, m); err != nil {
		return MessageDTO{}, err
	}

	dto := MessageDTO{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Sender:         m.Sender,
		Body:           m.Body,
		TS:             m.CreatedAtUnix,
		Receiver:       m.Receiver, // ✅
		Status:         m.Status,   //

		// Receiver/Status DTO'da yoksa şimdilik eklemeyebilirsin
	}

	// Pub/Sub yayını
	b, _ := json.Marshal(dto)
	_ = s.pub.Publish(ctx, roomChannel(in.ConversationID), b)

	return dto, nil
}

func (s *Service) ListMessages(ctx context.Context, conversationID string, limit int) ([]MessageDTO, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return []MessageDTO{}, common.ErrInvalidInput
	}

	msgs, err := s.repo.ListByConversation(ctx, conversationID, limit)
	if err != nil {
		return nil, err
	}

	out := make([]MessageDTO, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MessageDTO{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Sender:         m.Sender,
			Body:           m.Body,
			TS:             m.CreatedAtUnix,
			Receiver:       m.Receiver,
			Status:         m.Status,
			ReadAtUnix:     m.ReadAtUnix,
		})
	}
	return out, nil
}

// ✅ Okundu işaretleme
func (s *Service) MarkRead(ctx context.Context, conversationID, receiver string, messageIDs []string, readAtUnix int64) error {
	conversationID = strings.TrimSpace(conversationID)
	receiver = strings.TrimSpace(receiver)
	if conversationID == "" || receiver == "" {
		return common.ErrInvalidInput
	}
	if len(messageIDs) == 0 {
		return nil
	}
	if readAtUnix <= 0 {
		readAtUnix = common.NowUnix()
	}

	if err := s.repo.MarkRead(ctx, conversationID, receiver, messageIDs, readAtUnix); err != nil {
		return err
	}

	// room'a read event yayınlamak istersen (frontend tik güncellesin):
	ev, _ := json.Marshal(map[string]any{
		"type":           "message.read",
		"conversationId": conversationID,
		"receiver":       receiver,
		"payload": map[string]any{
			"messageIds": messageIDs,
			"readAt":     readAtUnix,
		},
	})
	_ = s.pub.Publish(ctx, roomChannel(conversationID), ev)

	return nil
}
