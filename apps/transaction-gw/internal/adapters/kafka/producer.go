package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"shared/events"
	sharedKafka "shared/kafka"
	"shared/metrics"
	"transaction-gw/internal/domain"
)

type Producer struct {
	shared *sharedKafka.Producer
}

func NewProducer(brokers []string, topic string, registry *metrics.MetricsRegistry) *Producer {
	return &Producer{
		shared: sharedKafka.NewProducer(brokers, topic, registry),
	}
}

func (p *Producer) Close() error {
	return p.shared.Close()
}

func (p *Producer) Publish(ctx context.Context, tx *domain.TransactionRequest) error {
	event := events.TransactionEvent{
		ID:             tx.ID.String(),
		SourceAccount:  tx.SourceAccount.String(),
		TargetAccount:  tx.TargetAccount.String(),
		Amount:         tx.AmountInCents,
		IdempotencyKey: tx.IdempotencyKey,
		Description:    tx.Description,
		CreatedAt:      tx.CreatedAt,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction event: %w", err)
	}

	err = p.shared.Publish(ctx, []byte(event.SourceAccount), payload)
	if err != nil {
		slog.Error("Falha ao publicar mensagem no Kafka", 
			"error", err, 
			"transaction_id", tx.ID,
		)
		return err
	}

	slog.Info("Mensagem enviada para o Kafka com sucesso", 
		"transaction_id", tx.ID, 
		"key", event.SourceAccount,
	)
	return nil
}
