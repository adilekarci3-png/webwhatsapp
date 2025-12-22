package redis

import (
	"context"
	"time"

	"example.com/webwhatsapp/backend/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

func NewClient(c config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
