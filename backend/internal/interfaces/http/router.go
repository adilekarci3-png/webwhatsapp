package http

import (
	"net/http"

	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
	"example.com/webwhatsapp/backend/internal/interfaces/http/handlers"
	"example.com/webwhatsapp/backend/internal/interfaces/ws"
)

func NewRouter(msgSvc *messaging.Service, wsHandler *ws.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.Health)

	// REST
	mux.HandleFunc("/messages", handlers.MessagesList(msgSvc))

	// WS
	mux.HandleFunc("/ws", wsHandler.ServeWS)

	// Basit CORS (aynı origin’de Nginx ile genelde gerek kalmaz, ama local kolaylık)
	return withBasicHeaders(mux)
}
