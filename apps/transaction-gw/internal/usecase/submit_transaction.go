package usecase

import (
	"context"
	"transaction-gw/internal/domain"
	"transaction-gw/internal/ports"
)

type SubmitTransactionUseCase struct {
	broker ports.TransactionBroker
}

func NewSubmitTransactionUseCase(broker ports.TransactionBroker) *SubmitTransactionUseCase {
	return &SubmitTransactionUseCase{
		broker: broker,
	}
}

func (uc *SubmitTransactionUseCase) Execute(ctx context.Context, req *domain.TransactionRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// Aqui poderíamos adicionar lógica de verificação de idempotência prévia 
	// ou persistência em Outbox caso o broker falhe.
	
	return uc.broker.Publish(ctx, req)
}
