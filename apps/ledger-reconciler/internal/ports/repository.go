package ports

import (
	"context"
	"ledger-reconciler/internal/domain"
	"github.com/google/uuid"
)

type ReconcilerRepository interface {
	GetAccountsByCursor(ctx context.Context, lastID uuid.UUID, limit int) ([]domain.AccountCursor, error)
	GetComputedBalance(ctx context.Context, accountID uuid.UUID) (int64, error)
}
