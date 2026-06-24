package ports

import "context"

type MessageConsumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, msg []byte) error) error
}
