package usecase

import (
	"context"
	"transaction-gw/internal/domain"
	"transaction-gw/internal/ports"

	"github.com/google/uuid"
)

type GetTransactionUseCase struct {
	repo ports.TransactionRepository
}

func NewGetTransactionUseCase(repo ports.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{repo: repo}
}

type GetTransactionOutput struct {
	ID              string `json:"transaction_id"`
	SourceAccount   string `json:"source_account_id"`
	TargetAccount   string `json:"target_account_id"`
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	IdempotencyKey  string `json:"idempotency_key"`
	CreatedAt       string `json:"created_at"`
}

func (uc *GetTransactionUseCase) Execute(ctx context.Context, transactionID uuid.UUID) (*GetTransactionOutput, error) {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	if tx == nil {
		return nil, domain.ErrTransactionNotFound
	}

	return &GetTransactionOutput{
		ID:             tx.ID.String(),
		SourceAccount:  tx.SourceAccount.String(),
		TargetAccount:  tx.TargetAccount.String(),
		Amount:         tx.AmountInCents,
		Description:    tx.Description,
		IdempotencyKey: tx.IdempotencyKey,
		CreatedAt:      tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
