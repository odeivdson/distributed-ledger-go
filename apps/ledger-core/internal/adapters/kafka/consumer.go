package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"ledger-core/internal/domain"
	"ledger-core/internal/usecase"
	"log"
	"log/slog"
	"os"
	"shared/events"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	usecase *usecase.LedgerUseCase
}

func NewConsumer(brokers []string, topic string, groupID string, uc *usecase.LedgerUseCase) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     groupID,
			MaxBytes:    10e6, // 10MB
			Logger:      log.New(os.Stdout, "KAFKA-CONSUMER: ", log.LstdFlags),
			ErrorLogger: log.New(os.Stderr, "KAFKA-CONSUMER-ERR: ", log.LstdFlags),
		}),
		usecase: uc,
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Consume(ctx context.Context) error {
	slog.Info("Iniciando consumo do Kafka", 
		"topic", c.reader.Config().Topic, 
		"group", c.reader.Config().GroupID,
		"brokers", c.reader.Config().Brokers,
	)
	
	for {
		slog.Debug("Aguardando próxima mensagem...")
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("Erro ao ler mensagem do Kafka", "error", err)
			continue
		}

		slog.Info("Mensagem recebida do Kafka", "partition", m.Partition, "offset", m.Offset)
		if err := c.HandleMessage(ctx, m.Value); err != nil {
			slog.Error("Erro ao processar mensagem", "error", err, "offset", m.Offset)
			// Em produção, aqui implementaríamos uma DLQ ou retentativas exponenciais
			continue
		}
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, data []byte) error {
	var msg events.TransactionEvent
	if err := json.Unmarshal(data, &msg); err != nil {
		slog.Error("Erro ao deserializar mensagem", "error", err)
		return err
	}

	slog.Info("Processando transação", "id", msg.ID, "idempotency_key", msg.IdempotencyKey)

	// Converter para domínio do Ledger
	txID, err := uuid.Parse(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid transaction id: %w", err)
	}
	sourceID, err := uuid.Parse(msg.SourceAccount)
	if err != nil {
		return fmt.Errorf("invalid source account id: %w", err)
	}
	targetID, err := uuid.Parse(msg.TargetAccount)
	if err != nil {
		return fmt.Errorf("invalid target account id: %w", err)
	}

	ledgerTx := domain.NewTransaction(msg.IdempotencyKey, msg.Description)
	ledgerTx.ID = txID // Preservar ID original se necessário

	// Adicionar entrada de Débito
	if err := ledgerTx.AddEntry(sourceID, domain.Debit, msg.Amount); err != nil {
		return err
	}

	// Adicionar entrada de Crédito
	if err := ledgerTx.AddEntry(targetID, domain.Credit, msg.Amount); err != nil {
		return err
	}

	// Executar caso de uso
	err = c.usecase.Execute(ctx, ledgerTx)
	if err != nil {
		if err == domain.ErrConcurrentUpdate {
			slog.Warn("Conflito de concorrência, retentativa automática via Kafka (NACK)")
			return err
		}
		slog.Error("Erro ao processar transação", "error", err, "id", msg.ID)
		return err
	}

	slog.Info("Transação processada com sucesso", "id", msg.ID)
	return nil
}
