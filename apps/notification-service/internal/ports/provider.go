package ports

import "context"

type NotificationProvider interface {
	Send(ctx context.Context, target string, payload []byte) error
}
