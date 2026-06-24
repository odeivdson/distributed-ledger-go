package usecase

import (
	"context"
	"ledger-reconciler/internal/domain"
	"ledger-reconciler/internal/ports"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type ReconcileUseCase struct {
	repo ports.ReconcilerRepository
}

func NewReconcileUseCase(repo ports.ReconcilerRepository) *ReconcileUseCase {
	return &ReconcileUseCase{repo: repo}
}

func (uc *ReconcileUseCase) RunBatch(ctx context.Context, chunkSize int, maxParallel int) error {
	var lastProcessedID uuid.UUID = uuid.Nil
	slog.Info("Iniciando processamento batch de reconciliação", "chunk_size", chunkSize, "max_parallel", maxParallel)

	for {
		accounts, err := uc.repo.GetAccountsByCursor(ctx, lastProcessedID, chunkSize)
		if err != nil {
			return err
		}

		if len(accounts) == 0 {
			break
		}

		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(maxParallel)

		for _, acc := range accounts {
			acc := acc
			g.Go(func() error {
				result, err := uc.ReconcileAccount(gCtx, acc)
				if err != nil {
					slog.Error("Erro ao reconciliar conta", "account_id", acc.ID, "error", err)
					return nil // Continuar com as outras contas
				}

				if !result.IsValid {
					slog.Warn("Divergência de saldo detectada!", 
						"account_id", result.AccountID, 
						"current", result.CurrentBalance, 
						"computed", result.ComputedBalance, 
						"diff", result.Diff)
				} else {
					slog.Debug("Conta reconciliada com sucesso", "account_id", result.AccountID)
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		lastProcessedID = accounts[len(accounts)-1].ID
		slog.Info("Chunk processado", "last_id", lastProcessedID, "count", len(accounts))
	}

	slog.Info("Processamento batch finalizado.")
	return nil
}

func (uc *ReconcileUseCase) ReconcileAccount(ctx context.Context, acc domain.AccountCursor) (domain.ReconciliationResult, error) {
	computed, err := uc.repo.GetComputedBalance(ctx, acc.ID)
	if err != nil {
		return domain.ReconciliationResult{}, err
	}

	return domain.ReconciliationResult{
		AccountID:       acc.ID,
		CurrentBalance:  acc.BalanceInCents,
		ComputedBalance: computed,
		Diff:            computed - acc.BalanceInCents,
		IsValid:         computed == acc.BalanceInCents,
		Timestamp:       time.Now(),
	}, nil
}
