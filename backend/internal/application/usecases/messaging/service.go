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
	in.Receiver = strings.TrimSpace(in.Receiver)
	in.Body = strings.TrimSpace(in.Body)

	if in.ConversationID == "" || in.Sender == "" || in.Body == "" {
		return MessageDTO{}, common.ErrInvalidInput
	}

	createdAt := in.TS
	if createdAt <= 0 {
		createdAt = common.NowUnix()
	}

	m := message.Message{
		ID:             common.NewID(),
		ConversationID: in.ConversationID,
		Sender:         in.Sender,
		Receiver:       in.Receiver,
		Body:           in.Body,
		Status:         "SENT",
		CreatedAtUnix:  createdAt,
	}

	// 1) Önce DB'ye yaz
	if err := s.repo.Insert(ctx, m); err != nil {
		return MessageDTO{}, err
	}

	// 2) ✅ DB insert başarılı -> ACK ver (kalıcı olsun istiyorsanız DB'de güncelle)
	if err := s.repo.UpdateStatus(ctx, m.ID, "ACK"); err == nil {
		m.Status = "ACK"
	} else {
		// UpdateStatus başarısız olsa bile mesaj kaydedildi; en azından DTO'da SENT kalsın
		// İsterseniz burada log atın.
	}

	dto := MessageDTO{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Sender:         m.Sender,
		Receiver:       m.Receiver,
		Body:           m.Body,
		TS:             m.CreatedAtUnix,
		Status:         m.Status, // ACK veya SENT
		ReadAtUnix:     m.ReadAtUnix,
		// ClientMsgID: in.ClientMsgID, // varsa echo edebilirsiniz
	}

	// 3) Odaya saf MessageDTO yayınla (Vue bunu direkt mesaj olarak push ediyor)
	if b, err := json.Marshal(dto); err == nil {
		_ = s.pub.Publish(ctx, roomChannel(in.ConversationID), b)
	}

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
			Receiver:       m.Receiver,
			Body:           m.Body,
			TS:             m.CreatedAtUnix,
			Status:         m.Status,
			ReadAtUnix:     m.ReadAtUnix,
		})
	}
	return out, nil
}

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

	// ✅ Vue bekliyor: type:"message.read" + payload:{messageIds, readAt}
	ev := WsEvent{
		Type:           "message.read",
		ConversationID: conversationID,
		Receiver:       receiver,
		Payload: map[string]any{
			"messageIds": messageIDs,
			"readAt":     readAtUnix,
		},
	}

	if b, err := json.Marshal(ev); err == nil {
		_ = s.pub.Publish(ctx, roomChannel(conversationID), b)
	}

	return nil
}
