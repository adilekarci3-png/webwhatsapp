package ports

import "context"

// Publisher interface'iniz muhtemelen messaging tarafından zaten kullanılıyor:
// type Publisher interface { Publish(ctx context.Context, channel string, payload []byte) error }

type NopPublisher struct{}

func (NopPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	return nil
}
