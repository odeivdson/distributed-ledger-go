package kafka

import (
	"context"
	"log"
	"os"
	"shared/metrics"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Writer   *kafka.Writer
	registry *metrics.MetricsRegistry
}

func NewProducer(brokers []string, topic string, registry *metrics.MetricsRegistry) *Producer {
	return &Producer{
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			MaxAttempts:  3,
			RequiredAcks: kafka.RequireAll,
			Logger:       log.New(os.Stdout, "KAFKA-PRODUCER: ", log.LstdFlags),
			ErrorLogger:  log.New(os.Stderr, "KAFKA-PRODUCER-ERR: ", log.LstdFlags),
		},
		registry: registry,
	}
}

func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err == nil && p.registry != nil {
		p.registry.KafkaMessagesProduced.WithLabelValues(p.Writer.Topic).Inc()
	}
	return err
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}
