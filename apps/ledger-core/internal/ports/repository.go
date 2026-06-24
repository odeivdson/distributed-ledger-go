package ports

import (
	"context"
	"ledger-core/internal/domain"

	"github.com/google/uuid"
)

type AccountRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	UpdateBalance(ctx context.Context, account *domain.Account) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type OutboxRepository interface {
	Save(ctx context.Context, item *domain.OutboxItem) error
	GetPending(ctx context.Context, limit int) ([]*domain.OutboxItem, error)
	Update(ctx context.Context, item *domain.OutboxItem) error
	SaveFailedEvent(ctx context.Context, id uuid.UUID, topic string, payload []byte, errMsg string, attempts int) error
}


