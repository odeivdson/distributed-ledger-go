package ports

import (
	"context"
)

type MessageConsumer interface {
	Consume(ctx context.Context) error
}

type NotificationBroker interface {
	PublishNotification(ctx context.Context, accountID string, message string) error
	Publish(ctx context.Context, key []byte, payload []byte) error
}

