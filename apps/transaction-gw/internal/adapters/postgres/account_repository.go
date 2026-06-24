package postgres

import (
	"context"
	"database/sql"
	"transaction-gw/internal/domain"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `INSERT INTO accounts_balance (account_id, balance_in_cents, version)
              VALUES ($1, $2, $3)
              ON CONFLICT (account_id) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query, account.ID, account.BalanceInCents, account.Version)
	return err
}
