package bootstrap

import (
	"log"
	"net/http"
	"os"

	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
	"example.com/webwhatsapp/backend/internal/infrastructure/cache/redis"
	"example.com/webwhatsapp/backend/internal/infrastructure/config"
	"example.com/webwhatsapp/backend/internal/infrastructure/persistence/postgres"
	ihttp "example.com/webwhatsapp/backend/internal/interfaces/http"
	"example.com/webwhatsapp/backend/internal/interfaces/ws"
)

type App struct {
	Port   string
	Router http.Handler
}

func Build() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	var (
		pg     *postgres.DB
		ps     *redis.PubSub // ✅ somut tip: ws.NewHandler bunu istiyor
		msgSvc *messaging.Service
	)

	// ---------- Postgres ----------
	log.Printf("connecting postgres host=%s db=%s", cfg.Postgres.Host, cfg.Postgres.DB)
	pg, err = postgres.NewDB(cfg.Postgres)
	if err != nil {
		log.Printf("startup warning: postgres unavailable: %v", err)
	} else {
		if err := postgres.Migrate(pg.Pool, cfg.Postgres.MigrationsDir); err != nil {
			log.Printf("startup warning: postgres migrate failed: %v", err)
		}
	}

	// ---------- Redis ----------
	log.Printf("connecting redis addr=%s", cfg.Redis.Addr)
	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Printf("startup warning: redis unavailable: %v", err)
	} else {
		ps = redis.NewPubSub(rdb) // *redis.PubSub
	}

	// ---------- Messaging Service ----------
	// messaging.NewService ikinci parametre ports.Publisher ise,
	// redis.PubSub'niz zaten Publish metoduyla uyumluydu.
	if pg != nil && ps != nil {
		msgRepo := postgres.NewMessageRepo(pg.Pool)
		msgSvc = messaging.NewService(msgRepo, ps)
	} else {
		log.Printf("starting in DEGRADED MODE: messaging service is unavailable (pg=%v redis=%v)", pg != nil, ps != nil)
		msgSvc = nil
	}

	// ---------- WS + HTTP Router ----------
	// ✅ ws handler: redis yoksa ps nil gider -> ws handler içinde 503 dönmeli
	wsHandler := ws.NewHandler(msgSvc, ps)
	router := ihttp.NewRouter(msgSvc, wsHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return &App{Port: port, Router: router}, nil
}
