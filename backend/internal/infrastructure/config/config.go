package config

import "os"

type PostgresConfig struct {
	Host          string
	Port          string
	DB            string
	User          string
	Password      string
	MigrationsDir string
}

type RedisConfig struct {
	Addr     string
	Password string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
}

func Load() (Config, error) {
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	return Config{
		Postgres: PostgresConfig{
			Host:          getenv("POSTGRES_HOST", "postgres"),
			Port:          getenv("POSTGRES_PORT", "5432"),
			DB:            getenv("POSTGRES_DB", "example.com/webwhatsapp"),
			User:          getenv("POSTGRES_USER", "webwa"),
			Password:      getenv("POSTGRES_PASSWORD", "webwa123"),
			MigrationsDir: "internal/infrastructure/persistence/postgres/migrations",
		},
		Redis: RedisConfig{
			Addr:     getenv("REDIS_ADDR", "redis:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getenv("MINIO_ENDPOINT", "minio:9000"),
			AccessKey: getenv("MINIO_ACCESS_KEY", ""),
			SecretKey: getenv("MINIO_SECRET_KEY", ""),
			Bucket:    getenv("MINIO_BUCKET", "media"),
			UseSSL:    useSSL,
		},
	}, nil
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
