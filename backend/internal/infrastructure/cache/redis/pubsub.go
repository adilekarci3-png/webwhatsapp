package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type PubSub struct {
	rdb *redis.Client
}

func NewPubSub(rdb *redis.Client) *PubSub {
	return &PubSub{rdb: rdb}
}

// ✅ application/ports.Publisher uyumluluğu
func (p *PubSub) Publish(ctx context.Context, channel string, payload []byte) error {
	return p.rdb.Publish(ctx, channel, payload).Err()
}

// ✅ application/ports.Subscriber uyumluluğu (go-redis tipi dışarı sızmıyor)
func (p *PubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	ps := p.rdb.Subscribe(ctx, channel)

	// go-redis: abonelik hazır mı kontrol etmek iyi olur
	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		return nil, func() {}, err
	}

	// go-redis mesaj kanalını byte[] kanalına dönüştürüyoruz
	out := make(chan []byte)

	go func() {
		defer close(out)
		ch := ps.Channel()
		for msg := range ch {
			out <- []byte(msg.Payload)
		}
	}()

	unsub := func() {
		_ = ps.Close()
	}

	return out, unsub, nil
}
