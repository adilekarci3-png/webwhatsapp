package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
	"example.com/webwhatsapp/backend/internal/infrastructure/cache/redis"
)

type Handler struct {
	msgSvc *messaging.Service
	pubsub *redis.PubSub
}

func NewHandler(msgSvc *messaging.Service, pubsub *redis.PubSub) *Handler {
	return &Handler{msgSvc: msgSvc, pubsub: pubsub}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conv := r.URL.Query().Get("conversationId")
	if conv == "" {
		conv = r.URL.Query().Get("room")
	}
	sender := r.URL.Query().Get("sender")
	if sender == "" {
		sender = r.URL.Query().Get("user")
	}

	if conv == "" {
		http.Error(w, "conversationId (room) required", http.StatusBadRequest)
		return
	}
	if sender == "" {
		http.Error(w, "sender (user) required", http.StatusBadRequest)
		return
	}

	if h.pubsub == nil || h.msgSvc == nil {
		http.Error(w, "websocket service unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	channel := "room:" + conv

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Redis subscribe
	msgs, unsubscribe, err := h.pubsub.Subscribe(ctx, channel)
	if err != nil {
		http.Error(w, "redis subscribe failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer unsubscribe()

	// CONNECT: presence online
	_ = h.publishPresence(ctx, channel, conv, sender, true, 0)

	// Redis -> WS
	done := make(chan struct{})
	go func() {
		defer close(done)
		for payload := range msgs {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}()

	// WS -> Handlers
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Yalnızca JSON event kabul ediyoruz (legacy text kaldırıldı)
		var ev Event
		if json.Unmarshal(payload, &ev) != nil || ev.Type == "" {
			continue
		}

		// ✅ conversationId ve sender SERVER tarafında sabitlenir
		ev.ConversationID = conv
		ev.Sender = sender

		switch ev.Type {

		case "typing.start":
			_ = h.publishTyping(ctx, channel, conv, sender, true)
			continue

		case "typing.stop":
			_ = h.publishTyping(ctx, channel, conv, sender, false)
			continue

		case "message.read":
			var rp ReadPayload
			_ = json.Unmarshal(ev.Payload, &rp)
			if rp.ReadAt == 0 {
				rp.ReadAt = time.Now().Unix()
			}

			if err := h.msgSvc.MarkRead(ctx, conv, sender, rp.MessageIDs, rp.ReadAt); err != nil {
				h.writeError(conn, err)
				continue
			}

			// ✅ read event’i odaya yayınla (diğer istemciler double tick görsün)
			out, _ := json.Marshal(map[string]any{
				"type":           "message.read",
				"conversationId": conv,
				"sender":         sender,
				"payload": map[string]any{
					"messageIds": rp.MessageIDs,
					"readAt":     rp.ReadAt,
				},
			})
			_ = h.pubsub.Publish(ctx, channel, out)
			continue

		case "message.send":
			body := strings.TrimSpace(ev.Body)
			if body == "" {
				continue
			}

			// ✅ oda sabitle: client event'i farklı conversationId ile gelirse reddet
			// (Bu satır pratikte hep true olur çünkü ev.ConversationID zaten conv'a sabitlendi.
			// Ama kalsın: ileride sabitlemeyi kaldırırsan güvenlik katmanı olur.)
			if ev.ConversationID != "" && ev.ConversationID != conv {
				h.writeError(conn, fmt.Errorf("conversationId mismatch"))
				continue
			}

			in := messaging.SendMessageInput{
				ConversationID: conv,        // ✅ URL'den gelen conv sabit
				Sender:         sender,      // ✅ URL'den gelen sender sabit
				Receiver:       ev.Receiver, // hedef
				Body:           body,
				TS:             ev.TS,
				ClientMsgID:    ev.ClientMsgID,
			}

			msg, err := h.msgSvc.SendMessage(ctx, in)
			if err != nil {
				h.writeError(conn, err)
				continue
			}

			// ✅ 1) Mesajı anlık dağıtım için Redis'e publish et
			// msg MessageDTO ise direkt marshal uygundur
			out, _ := json.Marshal(msg)
			_ = h.pubsub.Publish(ctx, channel, out)

			// ✅ 2) ACK sadece gönderene
			ack, _ := json.Marshal(map[string]any{
				"type":        "message.ack",
				"messageId":   msg.ID,
				"clientMsgId": in.ClientMsgID,
				"status":      "ACK",
			})
			_ = conn.WriteMessage(websocket.TextMessage, ack)
			continue

		default:
			// bilinmeyen event => ignore
			continue
		}
	}

	// DISCONNECT: presence offline
	_ = h.publishPresence(ctx, channel, conv, sender, false, time.Now().Unix())
	<-done
}

func (h *Handler) writeError(conn *websocket.Conn, err error) {
	b, _ := json.Marshal(map[string]any{
		"type":  "error",
		"error": err.Error(),
	})
	_ = conn.WriteMessage(websocket.TextMessage, b)
}

func (h *Handler) publishTyping(ctx context.Context, channel, conv, user string, isTyping bool) error {
	out, _ := json.Marshal(Event{
		Type:           "typing",
		ConversationID: conv,
		Sender:         user,
		Payload:        mustJSONRaw(TypingPayload{IsTyping: isTyping}),
	})
	return h.pubsub.Publish(ctx, channel, out)
}

func (h *Handler) publishPresence(ctx context.Context, channel, conv, user string, online bool, lastSeen int64) error {
	pp := PresencePayload{Online: online}
	if !online && lastSeen > 0 {
		pp.LastSeenAt = lastSeen
	}
	out, _ := json.Marshal(Event{
		Type:           "presence.update",
		ConversationID: conv,
		Sender:         user,
		Payload:        mustJSONRaw(pp),
	})
	return h.pubsub.Publish(ctx, channel, out)
}

func mustJSONRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
