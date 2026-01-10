package bootstrap

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	authuc "example.com/webwhatsapp/backend/internal/application/usecases/auth"
	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
	"example.com/webwhatsapp/backend/internal/infrastructure/cache/redis"
	"example.com/webwhatsapp/backend/internal/infrastructure/config"
	"example.com/webwhatsapp/backend/internal/infrastructure/persistence/postgres"
	ihttp "example.com/webwhatsapp/backend/internal/interfaces/http"
	httpauth "example.com/webwhatsapp/backend/internal/interfaces/http/auth"
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
		pg      *postgres.DB
		ps      *redis.PubSub
		msgSvc  *messaging.Service
		userSvc *authuc.UserService
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
		ps = redis.NewPubSub(rdb)
	}

	// ---------- Messaging Service ----------
	if pg != nil && ps != nil {
		msgRepo := postgres.NewMessageRepo(pg.Pool)
		msgSvc = messaging.NewService(msgRepo, ps)
	} else {
		log.Printf("starting in DEGRADED MODE: messaging service is unavailable (pg=%v redis=%v)", pg != nil, ps != nil)
		msgSvc = nil
	}

	// ---------- Auth (UserService) ----------
	// Postgres varsa auth aktif olur. Yoksa userSvc nil kalır.
	if pg != nil {
		// Burada repo’nun pg.Pool (pgxpool) ile uyumlu olması gerekir.
		// Eğer sen repo’yu *sql.DB ile yazdıysan, bunu pgxpool’a çevirmeliyiz.
		userRepo := postgres.NewUserRepository(pg.Pool)
		userSvc = authuc.NewUserService(userRepo)
	} else {
		log.Printf("starting in DEGRADED MODE: auth service is unavailable (postgres is nil)")
		userSvc = nil
	}

	// ---------- WS + HTTP Router ----------
	wsHandler := ws.NewHandler(msgSvc, ps)

	// Prod tespiti: APP_ENV=prod|production ise CookieSecure=true
	isProd := false
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))); env == "prod" || env == "production" {
		isProd = true
	}

	// JWT config: cfg.JWT yerine ENV'den oku (cfg.Config'ta JWT alanı yoksa bile derlenir)
	jwtCfg := httpauth.Config{
		AccessSecret:  []byte(getEnv("JWT_ACCESS_SECRET", "change-me-access")),
		RefreshSecret: []byte(getEnv("JWT_REFRESH_SECRET", "change-me-refresh")),
		AccessTTL:     mustParseDuration(getEnv("JWT_ACCESS_TTL", "15m")),
		RefreshTTL:    mustParseDuration(getEnv("JWT_REFRESH_TTL", "720h")), // 30 gün
		CookieSecure:  isProd,
		CookieDomain:  getEnv("JWT_COOKIE_DOMAIN", ""),
	}

	// Router’a userSvc enjekte et (NewRouter imzası bunu kabul etmeli)
	router := ihttp.NewRouter(msgSvc, wsHandler, jwtCfg, userSvc)

	// Port
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return &App{Port: port, Router: router}, nil
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(s))
	if err != nil {
		// Güvenli fallback: parse edilemezse 15m
		return 15 * time.Minute
	}
	return d
}
