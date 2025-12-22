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
	in.Body = strings.TrimSpace(in.Body)

	if in.ConversationID == "" || in.Sender == "" || in.Body == "" {
		return MessageDTO{}, common.ErrInvalidInput
	}

	m := message.Message{
		ID:             common.NewID(),
		ConversationID: in.ConversationID,
		Sender:         in.Sender,
		Body:           in.Body,
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
	}

	// Pub/Sub yayını (WS consumer'lar bunu client'a iter)
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
	// repo DESC döndürür; UI genelde ASC ister -> burada ters çevirmiyoruz (frontend ters çevirebilir)
	for _, m := range msgs {
		out = append(out, MessageDTO{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Sender:         m.Sender,
			Body:           m.Body,
			TS:             m.CreatedAtUnix,
		})
	}
	return out, nil
}
