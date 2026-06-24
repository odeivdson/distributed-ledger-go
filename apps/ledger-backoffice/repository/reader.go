package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type DashboardData struct {
	TotalAccounts     int64
	TotalVolumeCents  int64
	LastReconcileStatus string
	Alerts            []ReconcileAlert
}

type ReconcileAlert struct {
	AccountID string
	Reason    string
}

type AccountDetail struct {
	ID        string
	Balance   int64
	Version   int64
	UpdatedAt time.Time
}

type LedgerEntry struct {
	ID            string
	TransactionID string
	Type          string
	Amount        int64
	CreatedAt     time.Time
}

type ReaderRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewReaderRepository(db *sql.DB, rdb *redis.Client) *ReaderRepository {
	return &ReaderRepository{
		db:    db,
		redis: rdb,
	}
}

func (r *ReaderRepository) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	data := &DashboardData{}

	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM accounts_balance").Scan(&data.TotalAccounts)
	if err != nil {
		return nil, fmt.Errorf("error counting accounts: %w", err)
	}

	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount_in_cents), 0) FROM ledger_entries WHERE entry_type = 'CREDIT'").Scan(&data.TotalVolumeCents)
	if err != nil {
		return nil, fmt.Errorf("error calculating total volume: %w", err)
	}

	status, err := r.redis.Get(ctx, "reconciler:last_run_status").Result()
	if err == redis.Nil {
		data.LastReconcileStatus = "Nunca executado"
	} else if err != nil {
		data.LastReconcileStatus = "Erro ao buscar status"
	} else {
		data.LastReconcileStatus = status
	}

	query := `
		SELECT ab.account_id
		FROM accounts_balance ab
		JOIN (
			SELECT account_id, 
			       SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_in_cents ELSE -amount_in_cents END) as calculated_balance
			FROM ledger_entries
			GROUP BY account_id
		) le ON ab.account_id = le.account_id
		WHERE ab.balance_in_cents != le.calculated_balance
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error searching for inconsistencies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alert ReconcileAlert
		if err := rows.Scan(&alert.AccountID); err != nil {
			continue
		}
		alert.Reason = "Saldo corrente diverge do rastro de auditoria"
		data.Alerts = append(data.Alerts, alert)
	}

	return data, nil
}

func (r *ReaderRepository) GetAccountDetail(ctx context.Context, id string) (*AccountDetail, error) {
	acc := &AccountDetail{}
	err := r.db.QueryRowContext(ctx, 
		"SELECT account_id, balance_in_cents, version, updated_at FROM accounts_balance WHERE account_id = $1", 
		id).Scan(&acc.ID, &acc.Balance, &acc.Version, &acc.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching account: %w", err)
	}
	return acc, nil
}

func (r *ReaderRepository) GetAuditTrail(ctx context.Context, accountID string) ([]LedgerEntry, error) {
	rows, err := r.db.QueryContext(ctx, 
		"SELECT id, transaction_id, entry_type, amount_in_cents, created_at FROM ledger_entries WHERE account_id = $1 ORDER BY created_at DESC", 
		accountID)
	if err != nil {
		return nil, fmt.Errorf("error fetching audit trail: %w", err)
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.Type, &e.Amount, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *ReaderRepository) GetDLQVolume(ctx context.Context) (int64, error) {
	val, err := r.redis.Get(ctx, "dlq:notification_failures:count").Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}
