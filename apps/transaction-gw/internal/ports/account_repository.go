package ports

import (
	"context"
	"transaction-gw/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
}
