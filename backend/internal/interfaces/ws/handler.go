package ws

import (
	"context"
	"encoding/json"
	"net/http"
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
	receiver := r.URL.Query().Get("receiver")
	if receiver == "" {
		receiver = r.URL.Query().Get("to")
	}

	if conv == "" {
		http.Error(w, "conversationId (room) required", http.StatusBadRequest)
		return
	}
	if sender == "" {
		http.Error(w, "sender (user) required", http.StatusBadRequest)
		return
	}

	// ✅ Degraded mode guard
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

	// ✅ Redis subscribe
	msgs, unsubscribe, err := h.pubsub.Subscribe(ctx, channel)
	if err != nil {
		http.Error(w, "redis subscribe failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer unsubscribe()

	// === CONNECT: presence online ===
	_ = h.publishPresence(ctx, channel, conv, sender, true, 0)

	// ✅ Redis -> WS
	done := make(chan struct{})
	go func() {
		defer close(done)
		for payload := range msgs {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}()

	// ✅ WS -> Handlers
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 1) JSON event mi?
		var ev Event
		if json.Unmarshal(payload, &ev) == nil && ev.Type != "" {
			// default bind
			if ev.ConversationID == "" {
				ev.ConversationID = conv
			}
			if ev.Sender == "" {
				ev.Sender = sender
			}
			if ev.Receiver == "" {
				ev.Receiver = receiver
			}

			switch ev.Type {
			case "typing.start":
				_ = h.publishTyping(ctx, channel, conv, sender, true)
				continue
			case "typing.stop":
				_ = h.publishTyping(ctx, channel, conv, sender, false)
				continue
			case "message.read":
				// payload: {messageIds:[...], readAt:...}
				var rp ReadPayload
				_ = json.Unmarshal(ev.Payload, &rp)
				if rp.ReadAt == 0 {
					rp.ReadAt = time.Now().Unix()
				}

				// DB update + publish (mesaj service'e eklenecek)
				if err := h.msgSvc.MarkRead(ctx, ev.ConversationID, sender, rp.MessageIDs, rp.ReadAt); err != nil {
					h.writeError(conn, err)
					continue
				}

				// room'a read event yayınla (diğer taraf tik günceller)
				out, _ := json.Marshal(Event{
					Type:           "message.read",
					ConversationID: ev.ConversationID,
					Sender:         sender,
					Receiver:       receiver,
					Payload:        mustJSONRaw(rp),
				})
				_ = h.pubsub.Publish(ctx, channel, out)
				continue
			default:
				// bilinmeyen event
				continue
			}
		}

		// 2) Plain text geldiyse => message.new gibi davran
		in := messaging.SendMessageInput{
			ConversationID: conv,
			Sender:         sender,
			Receiver:       receiver,
			Body:           string(payload),
		}

		msg, err := h.msgSvc.SendMessage(ctx, in)
		if err != nil {
			h.writeError(conn, err)
			continue
		}

		// ACK: sender'a da (room üzerinden) dönebilir; minimalde room'a yayınla
		ack, _ := json.Marshal(map[string]any{
			"type":      "message.ack",
			"messageId": msg.ID,
			"status":    "ACK",
		})
		_ = h.pubsub.Publish(ctx, channel, ack)
	}

	// === DISCONNECT: presence offline ===
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
