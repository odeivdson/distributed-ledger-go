package worker

import (
	"context"
	"log/slog"
	"time"

	"ledger-core/internal/domain"
	"ledger-core/internal/ports"

	"github.com/google/uuid"
)

type OutboxWorker struct {
	repo        ports.OutboxRepository
	publisher   ports.NotificationBroker
	interval    time.Duration
	limit       int
	topic       string
	maxAttempts int
}

func NewOutboxWorker(
	repo ports.OutboxRepository,
	publisher ports.NotificationBroker,
	interval time.Duration,
	limit int,
	topic string,
	maxAttempts int,
) *OutboxWorker {
	if interval <= 0 {
		interval = 1 * time.Second
	}
	if limit <= 0 {
		limit = 100
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return &OutboxWorker{
		repo:        repo,
		publisher:   publisher,
		interval:    interval,
		limit:       limit,
		topic:       topic,
		maxAttempts: maxAttempts,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	slog.Info("Trabalhador de Outbox iniciado", 
		"intervalo", w.interval, 
		"limite_lote", w.limit, 
		"topico", w.topic,
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processBatch(ctx)
		case <-ctx.Done():
			slog.Info("Trabalhador de Outbox encerrando por contexto.")
			return
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	// Buscar itens PENDING
	items, err := w.repo.GetPending(ctx, w.limit)
	if err != nil {
		slog.Error("Erro ao buscar pendências do outbox", "error", err)
		return
	}

	if len(items) == 0 {
		return
	}

	slog.Debug("Processando lote do outbox", "itens_encontrados", len(items))

	for _, item := range items {
		// Republicar no Kafka
		err := w.publisher.Publish(ctx, []byte(item.AggregateID.String()), item.Payload)
		if err != nil {
			item.Attempts++
			item.LastError = err.Error()

			slog.Warn("Falha ao publicar evento do outbox", 
				"outbox_id", item.ID, 
				"tentativas", item.Attempts, 
				"error", err,
			)

			if item.Attempts >= w.maxAttempts {
				item.Status = domain.OutboxFailed
				// Salvar em failed_events (DLQ do Banco)
				dlqErr := w.repo.SaveFailedEvent(ctx, uuid.New(), w.topic, item.Payload, item.LastError, item.Attempts)
				if dlqErr != nil {
					slog.Error("Falha crítica ao salvar evento em failed_events", "outbox_id", item.ID, "error", dlqErr)
				}
			}

			// Atualizar o status e contagem de tentativas no banco
			if updateErr := w.repo.Update(ctx, item); updateErr != nil {
				slog.Error("Falha ao atualizar outbox no banco de dados", "outbox_id", item.ID, "error", updateErr)
			}
			continue
		}

		// Sucesso
		now := time.Now()
		item.Status = domain.OutboxPublished
		item.PublishedAt = &now
		item.Attempts++

		if updateErr := w.repo.Update(ctx, item); updateErr != nil {
			slog.Error("Falha ao marcar outbox como publicado no banco de dados", "outbox_id", item.ID, "error", updateErr)
		} else {
			slog.Debug("Mensagem do outbox publicada com sucesso", "outbox_id", item.ID)
		}
	}
}
