package ports

import "context"

// Subscriber interface'i daha önce eklediyseniz:
// type Subscriber interface {
//   Subscribe(ctx context.Context, channel string) (msgs <-chan []byte, unsubscribe func(), err error)
// }

type NopSubscriber struct{}

func (NopSubscriber) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	ch := make(chan []byte)
	unsub := func() { close(ch) }
	return ch, unsub, nil
}
