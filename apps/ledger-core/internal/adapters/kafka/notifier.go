package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"shared/events"
	sharedKafka "shared/kafka"
	"shared/metrics"

	"github.com/google/uuid"
)

type Notifier struct {
	shared *sharedKafka.Producer
}

func NewNotifier(brokers []string, topic string, registry *metrics.MetricsRegistry) *Notifier {
	return &Notifier{
		shared: sharedKafka.NewProducer(brokers, topic, registry),
	}
}

func (n *Notifier) Close() error {
	return n.shared.Close()
}

func (n *Notifier) PublishNotification(ctx context.Context, accountID string, message string) error {
	event := events.NotificationEvent{
		ID:      uuid.New().String(),
		Type:    "TRANSACTION_PROCESSED",
		Target:  accountID,
		Payload: []byte(message),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %w", err)
	}

	err = n.shared.Publish(ctx, []byte(accountID), payload)
	if err != nil {
		slog.Error("Falha ao publicar notificação no Kafka", "error", err, "account_id", accountID)
		return err
	}

	slog.Info("Notificação enviada para o Kafka", "account_id", accountID)
	return nil
}

func (n *Notifier) Publish(ctx context.Context, key []byte, payload []byte) error {
	return n.shared.Publish(ctx, key, payload)
}

