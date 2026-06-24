package postgres

import (
	"context"
	"database/sql"
	"shared/metrics"
	"transaction-gw/internal/domain"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db       *sql.DB
	registry *metrics.MetricsRegistry
}

func NewTransactionRepository(db *sql.DB, registry *metrics.MetricsRegistry) *TransactionRepository {
	return &TransactionRepository{db: db, registry: registry}
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TransactionRequest, error) {
	query := `
		SELECT id, source_account_id, target_account_id, amount, idempotency_key, description, created_at
		FROM processed_transactions
		WHERE id = $1
	`

	var tx domain.TransactionRequest
	var sourceID, targetID string

	var err error
	metrics.MeasureDBQuery("GetTransactionByID", r.registry, func() error {
		err = r.db.QueryRowContext(ctx, query, id.String()).Scan(
			&tx.ID,
			&sourceID,
			&targetID,
			&tx.AmountInCents,
			&tx.IdempotencyKey,
			&tx.Description,
			&tx.CreatedAt,
		)
		return err
	})

	if err == sql.ErrNoRows {
		return nil, domain.ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}

	tx.SourceAccount, _ = uuid.Parse(sourceID)
	tx.TargetAccount, _ = uuid.Parse(targetID)

	return &tx, nil
}

func (r *TransactionRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM processed_transactions
		WHERE idempotency_key = $1
	`

	var count int
	var err error
	metrics.MeasureDBQuery("ExistsByIdempotencyKey", r.registry, func() error {
		err = r.db.QueryRowContext(ctx, query, key).Scan(&count)
		return err
	})
	
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *TransactionRepository) SaveProcessed(ctx context.Context, id uuid.UUID, key string, requestHash string) error {
	query := `
		INSERT INTO processed_transactions (id, idempotency_key, request_hash, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (idempotency_key) DO NOTHING
	`

	var err error
	metrics.MeasureDBQuery("SaveProcessedTransaction", r.registry, func() error {
		_, err = r.db.ExecContext(ctx, query, id.String(), key, requestHash)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
