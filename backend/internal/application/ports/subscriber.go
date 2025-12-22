package ports

import "context"

type Subscriber interface {
	Subscribe(ctx context.Context, channel string) (msgs <-chan []byte, unsubscribe func(), err error)
}
