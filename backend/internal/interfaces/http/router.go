package http

import (
	"net/http"

	_ "example.com/webwhatsapp/backend/docs"

	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
	"example.com/webwhatsapp/backend/internal/interfaces/http/auth"
	"example.com/webwhatsapp/backend/internal/interfaces/http/handlers"
	"example.com/webwhatsapp/backend/internal/interfaces/http/middleware"
	"example.com/webwhatsapp/backend/internal/interfaces/ws"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(msgSvc *messaging.Service, wsHandler *ws.Handler, jwtCfg auth.Config) http.Handler {
	mux := http.NewServeMux()

	// Health (public)
	mux.HandleFunc("/health", handlers.Health)

	// Swagger (public)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// AUTH (public)
	mux.HandleFunc("/auth/login", handlers.Login(jwtCfg))
	mux.HandleFunc("/auth/refresh", handlers.Refresh(jwtCfg))
	mux.HandleFunc("/auth/logout", handlers.Logout(jwtCfg))

	// REST (protected)
	mux.Handle("/messages", middleware.RequireJWT(jwtCfg, http.HandlerFunc(handlers.MessagesList(msgSvc))))

	// WS (aşağıda ayrı anlatacağım)
	mux.HandleFunc("/ws", wsHandler.ServeWS)

	return withBasicHeaders(mux)
}
