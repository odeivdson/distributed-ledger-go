package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProcessedTransactionDTO struct {
	ID             uuid.UUID
	IdempotencyKey string
	RequestHash    string
	ResponseStatus string
	ResponseBody   string
}

type IdempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) GetByKey(ctx context.Context, key string) (*ProcessedTransactionDTO, error) {
	query := `
		SELECT id, idempotency_key, request_hash, response_status, response_body
		FROM processed_transactions
		WHERE idempotency_key = $1
	`

	pt := &ProcessedTransactionDTO{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&pt.ID,
		&pt.IdempotencyKey,
		&pt.RequestHash,
		&pt.ResponseStatus,
		&pt.ResponseBody,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return pt, nil
}

func (r *IdempotencyRepository) Create(ctx context.Context, pt *ProcessedTransactionDTO) error {
	query := `
		INSERT INTO processed_transactions (id, idempotency_key, request_hash, response_status, response_body, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		pt.ID,
		pt.IdempotencyKey,
		pt.RequestHash,
		pt.ResponseStatus,
		pt.ResponseBody,
		time.Now(),
		time.Now().Add(30*24*time.Hour),
	)

	if err != nil {
		return fmt.Errorf("failed to create processed transaction: %w", err)
	}

	return nil
}

func (r *IdempotencyRepository) Update(ctx context.Context, pt *ProcessedTransactionDTO) error {
	query := `
		UPDATE processed_transactions
		SET response_status = $1, response_body = $2, processed_at = $3
		WHERE idempotency_key = $4
	`

	_, err := r.db.ExecContext(ctx, query,
		pt.ResponseStatus,
		pt.ResponseBody,
		time.Now(),
		pt.IdempotencyKey,
	)

	if err != nil {
		return fmt.Errorf("failed to update processed transaction: %w", err)
	}

	return nil
}

func (r *IdempotencyRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM processed_transactions WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired transactions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return count, nil
}

func ComputeRequestHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
