package ports

import (
	"context"
	"transaction-gw/internal/domain"

	"github.com/google/uuid"
)

type TransactionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TransactionRequest, error)
	ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
	SaveProcessed(ctx context.Context, id uuid.UUID, key string, requestHash string) error
}
