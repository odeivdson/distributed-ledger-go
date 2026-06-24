package ports

import (
	"context"
	"transaction-gw/internal/adapters/postgres"
)

type IdempotencyRepository interface {
	GetByKey(ctx context.Context, key string) (*postgres.ProcessedTransactionDTO, error)
	Create(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error
	Update(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error
	DeleteExpired(ctx context.Context) (int64, error)
}
