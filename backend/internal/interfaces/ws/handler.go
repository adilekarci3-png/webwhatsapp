package ws

import (
	"context"
	"encoding/json"
	"net/http"

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

	// ✅ Redis subscribe (yeni imza: 3 dönüş)
	msgs, unsubscribe, err := h.pubsub.Subscribe(ctx, channel)
	if err != nil {
		http.Error(w, "redis subscribe failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer unsubscribe()

	// ✅ Redis -> WS
	done := make(chan struct{})
	go func() {
		defer close(done)
		for payload := range msgs {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}()

	// ✅ WS -> Usecase (Insert DB + Publish Redis)
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		in := messaging.SendMessageInput{
			ConversationID: conv,
			Sender:         sender,
			Body:           string(payload),
		}

		_, err = h.msgSvc.SendMessage(ctx, in)
		if err != nil {
			b, _ := json.Marshal(map[string]any{
				"type":  "error",
				"error": err.Error(),
			})
			_ = conn.WriteMessage(websocket.TextMessage, b)
		}
		// publish zaten usecase içinde
	}

	<-done
}
