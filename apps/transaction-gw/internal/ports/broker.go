package ports

import (
	"context"
	"transaction-gw/internal/domain"
)

type TransactionBroker interface {
	Publish(ctx context.Context, tx *domain.TransactionRequest) error
}
