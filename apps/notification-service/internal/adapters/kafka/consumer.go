package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"notification-service/internal/worker"
	"shared/events"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	pool   *worker.WorkerPool
}

func NewConsumer(brokers []string, topic string, groupID string, pool *worker.WorkerPool) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		pool: pool,
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Consume(ctx context.Context) error {
	slog.Info("Iniciando consumo do Kafka para notificações...", "topic", c.reader.Config().Topic)
	
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("Erro ao ler mensagem do Kafka", "error", err)
			continue
		}

		if err := c.HandleMessage(ctx, m.Value); err != nil {
			slog.Error("Erro ao processar mensagem de notificação", "error", err)
			continue
		}
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, data []byte) error {
	var msg events.NotificationEvent
	if err := json.Unmarshal(data, &msg); err != nil {
		slog.Error("Erro ao deserializar mensagem de notificação", "error", err)
		return err
	}

	job := worker.Job{
		NotificationID: msg.ID,
		Payload:        msg.Payload,
		Target:         msg.Target,
	}

	c.pool.Submit(job)
	slog.Debug("Tarefa de notificação submetida ao pool", "notification_id", msg.ID)
	
	return nil
}
