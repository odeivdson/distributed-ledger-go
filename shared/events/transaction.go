package events

import "time"

type TransactionEvent struct {
	ID             string    `json:"id"`
	SourceAccount  string    `json:"source_account_id"`
	TargetAccount  string    `json:"target_account_id"`
	Amount         int64     `json:"amount"`
	IdempotencyKey string    `json:"idempotency_key"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}
