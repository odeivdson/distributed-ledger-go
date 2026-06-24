package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"ledger-core/internal/domain"
	"ledger-core/internal/ports"
	"shared/events"

	"github.com/google/uuid"
)

type LedgerUseCase struct {
	accountRepo ports.AccountRepository
	txRepo      ports.TransactionRepository
	outboxRepo  ports.OutboxRepository
	uow         ports.UnitOfWork
}

func NewLedgerUseCase(ar ports.AccountRepository, tr ports.TransactionRepository, or ports.OutboxRepository, uow ports.UnitOfWork) *LedgerUseCase {
	return &LedgerUseCase{
		accountRepo: ar,
		txRepo:      tr,
		outboxRepo:  or,
		uow:         uow,
	}
}

func (uc *LedgerUseCase) Execute(ctx context.Context, tx *domain.Transaction) error {
	return uc.uow.Execute(ctx, func(ctx context.Context) error {
		// 1. Verificar idempotência dentro da TX
		exists, err := uc.txRepo.ExistsByIdempotencyKey(ctx, tx.IdempotencyKey)
		if err != nil {
			return err
		}
		if exists {
			slog.Info("Transação ignorada por idempotência", "idempotency_key", tx.IdempotencyKey)
			return nil // Já processado
		}

		slog.Info("Processando entradas da transação", "count", len(tx.Entries))
		// 2. Processar entradas
		for _, entry := range tx.Entries {
			account, err := uc.accountRepo.GetByID(ctx, entry.AccountID)
			if err != nil {
				slog.Error("Conta não encontrada", "account_id", entry.AccountID)
				return err
			}

			if !account.IsActive() {
				slog.Error("Conta inativa", "account_id", entry.AccountID)
				return domain.ErrAccountInactive
			}

			if entry.Type == domain.Debit {
				if err := account.Debit(entry.AmountInCents); err != nil {
					return err
				}
			} else {
				if err := account.Credit(entry.AmountInCents); err != nil {
					return err
				}
			}

			// 3. Update com concorrência otimista
			if err := uc.accountRepo.UpdateBalance(ctx, account); err != nil {
				return err
			}
		}

		// 4. Salvar transação e entries
		slog.Info("Persistindo transação e entradas no banco")
		if err := uc.txRepo.Create(ctx, tx); err != nil {
			return err
		}

		// 5. Salvar notificações no outbox
		for _, entry := range tx.Entries {
			msg := fmt.Sprintf("Transação de R$ %.2f processada com sucesso.", float64(entry.AmountInCents)/100.0)
			
			event := events.NotificationEvent{
				ID:      uuid.New().String(),
				Type:    "TRANSACTION_PROCESSED",
				Target:  entry.AccountID.String(),
				Payload: []byte(msg),
			}

			payloadBytes, err := json.Marshal(event)
			if err != nil {
				return fmt.Errorf("failed to marshal outbox notification payload: %w", err)
			}

			outboxItem := domain.NewOutboxItem("notification", entry.AccountID, "transaction_processed", payloadBytes)
			if err := uc.outboxRepo.Save(ctx, outboxItem); err != nil {
				return fmt.Errorf("failed to save outbox item: %w", err)
			}
		}

		return nil
	})
}

