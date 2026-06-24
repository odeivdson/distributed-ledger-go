package postgres

import (
	"context"
	"database/sql"
	"ledger-reconciler/internal/domain"
	"github.com/google/uuid"
)

type ReconcilerRepository struct {
	db *sql.DB
}

func NewReconcilerRepository(db *sql.DB) *ReconcilerRepository {
	return &ReconcilerRepository{db: db}
}

func (r *ReconcilerRepository) GetAccountsByCursor(ctx context.Context, lastID uuid.UUID, limit int) ([]domain.AccountCursor, error) {
	query := `
		SELECT account_id, balance_in_cents 
		FROM accounts_balance 
		WHERE account_id > $1 
		ORDER BY account_id 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, lastID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []domain.AccountCursor
	for rows.Next() {
		var acc domain.AccountCursor
		if err := rows.Scan(&acc.ID, &acc.BalanceInCents); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *ReconcilerRepository) GetComputedBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	// Soma de créditos - soma de débitos
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_in_cents ELSE -amount_in_cents END), 0)
		FROM ledger_entries 
		WHERE account_id = $1`
	
	var balance int64
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&balance)
	return balance, err
}
