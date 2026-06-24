package postgres

import (
	"context"
	"database/sql"
	"ledger-core/internal/domain"
	"log/slog"
	"shared/metrics"

	"github.com/google/uuid"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type LedgerRepository struct {
	db       *sql.DB
	registry *metrics.MetricsRegistry
}

func NewLedgerRepository(db *sql.DB, registry *metrics.MetricsRegistry) *LedgerRepository {
	return &LedgerRepository{db: db, registry: registry}
}

func (r *LedgerRepository) getExecutor(ctx context.Context) DBExecutor {
	if tx, ok := GetTx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *LedgerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	query := `SELECT account_id, status, balance_in_cents, version, updated_at FROM accounts_balance WHERE account_id = $1`
	var acc domain.Account
	var status string
	var err error
	metrics.MeasureDBQuery("GetAccountByID", r.registry, func() error {
		err = r.getExecutor(ctx).QueryRowContext(ctx, query, id).Scan(&acc.ID, &status, &acc.BalanceInCents, &acc.Version, &acc.UpdatedAt)
		return err
	})
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	acc.Status = domain.AccountStatus(status)
	return &acc, nil
}

func (r *LedgerRepository) UpdateBalance(ctx context.Context, account *domain.Account) error {
	query := `UPDATE accounts_balance SET balance_in_cents = $1, version = version + 1, updated_at = NOW() 
              WHERE account_id = $2 AND version = $3`
	var result sql.Result
	var err error
	metrics.MeasureDBQuery("UpdateAccountBalance", r.registry, func() error {
		result, err = r.getExecutor(ctx).ExecContext(ctx, query, account.BalanceInCents, account.ID, account.Version)
		return err
	})
	if err != nil {
		slog.Error("Erro ao atualizar saldo", "error", err, "account_id", account.ID)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		slog.Warn("Conflito de concorrência detectado no repositório", "account_id", account.ID, "version", account.Version)
		return domain.ErrConcurrentUpdate
	}
	slog.Debug("Saldo atualizado com sucesso", "account_id", account.ID, "new_balance", account.BalanceInCents)
	return nil
}

func (r *LedgerRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE idempotency_key = $1)`
	var exists bool
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, key).Scan(&exists)
	return exists, err
}

func (r *LedgerRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	executor := r.getExecutor(ctx)

	// Inserir Transação
	queryTx := `INSERT INTO transactions (id, idempotency_key, description, created_at) VALUES ($1, $2, $3, $4)`
	_, err := executor.ExecContext(ctx, queryTx, tx.ID, tx.IdempotencyKey, tx.Description, tx.CreatedAt)
	if err != nil {
		slog.Error("Erro ao inserir transação", "error", err, "tx_id", tx.ID)
		return err
	}

	// Inserir Entries
	queryEntry := `INSERT INTO ledger_entries (id, transaction_id, account_id, entry_type, amount_in_cents, created_at) 
                   VALUES ($1, $2, $3, $4, $5, $6)`
	for i, entry := range tx.Entries {
		_, err := executor.ExecContext(ctx, queryEntry, entry.ID, entry.TransactionID, entry.AccountID, entry.Type, entry.AmountInCents, entry.CreatedAt)
		if err != nil {
			slog.Error("Erro ao inserir entry", "error", err, "entry_index", i, "tx_id", tx.ID)
			return err
		}
	}

	slog.Info("Transação e entries persistidas com sucesso", "tx_id", tx.ID, "entries_count", len(tx.Entries))
	return nil
}

func (r *LedgerRepository) Save(ctx context.Context, item *domain.OutboxItem) error {
	executor := r.getExecutor(ctx)
	query := `INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload, status, attempts, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := executor.ExecContext(ctx, query, item.ID, item.AggregateType, item.AggregateID, item.EventType, item.Payload, item.Status, item.Attempts, item.CreatedAt)
	if err != nil {
		slog.Error("Erro ao salvar item no outbox", "error", err, "outbox_id", item.ID)
	}
	return err
}

func (r *LedgerRepository) GetPending(ctx context.Context, limit int) ([]*domain.OutboxItem, error) {
	executor := r.getExecutor(ctx)
	query := `SELECT id, aggregate_type, aggregate_id, event_type, payload, status, attempts, last_error, created_at, published_at 
              FROM outbox 
              WHERE status = 'PENDING' 
              ORDER BY created_at ASC 
              LIMIT $1`
	rows, err := executor.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.OutboxItem
	for rows.Next() {
		var item domain.OutboxItem
		var lastError sql.NullString
		var publishedAt sql.NullTime
		err := rows.Scan(
			&item.ID,
			&item.AggregateType,
			&item.AggregateID,
			&item.EventType,
			&item.Payload,
			&item.Status,
			&item.Attempts,
			&lastError,
			&item.CreatedAt,
			&publishedAt,
		)
		if err != nil {
			return nil, err
		}
		if lastError.Valid {
			item.LastError = lastError.String
		}
		if publishedAt.Valid {
			item.PublishedAt = &publishedAt.Time
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *LedgerRepository) Update(ctx context.Context, item *domain.OutboxItem) error {
	executor := r.getExecutor(ctx)
	query := `UPDATE outbox 
              SET status = $1, attempts = $2, last_error = $3, published_at = $4 
              WHERE id = $5`
	
	var lastError *string
	if item.LastError != "" {
		lastError = &item.LastError
	}
	
	_, err := executor.ExecContext(ctx, query, item.Status, item.Attempts, lastError, item.PublishedAt, item.ID)
	return err
}

func (r *LedgerRepository) SaveFailedEvent(ctx context.Context, id uuid.UUID, topic string, payload []byte, errMsg string, attempts int) error {
	executor := r.getExecutor(ctx)
	query := `INSERT INTO failed_events (id, source_topic, payload, error, attempts, first_error_at, last_error_at)
              VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := executor.ExecContext(ctx, query, id, topic, payload, errMsg, attempts)
	if err != nil {
		slog.Error("Erro ao salvar evento com falha (DLQ)", "error", err, "failed_event_id", id)
	}
	return err
}


