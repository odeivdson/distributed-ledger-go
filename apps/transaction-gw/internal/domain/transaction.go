package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInvalidAccount      = errors.New("source and destination accounts must be different")
	ErrAccountIDRequired   = errors.New("account ID is required")
	ErrIdempotencyRequired = errors.New("idempotency key is required")
)

type TransactionRequest struct {
	ID             uuid.UUID
	SourceAccount  uuid.UUID
	TargetAccount  uuid.UUID
	AmountInCents  int64
	IdempotencyKey string
	Description    string
	CreatedAt      time.Time
}

func (t *TransactionRequest) Validate() error {
	if t.SourceAccount == uuid.Nil || t.TargetAccount == uuid.Nil {
		return ErrAccountIDRequired
	}
	if t.SourceAccount == t.TargetAccount {
		return ErrInvalidAccount
	}
	if t.AmountInCents <= 0 {
		return ErrInvalidAmount
	}
	if t.IdempotencyKey == "" {
		return ErrIdempotencyRequired
	}
	return nil
}

func NewTransactionRequest(source, target uuid.UUID, amount int64, idempotencyKey, description string) *TransactionRequest {
	return &TransactionRequest{
		ID:             uuid.New(),
		SourceAccount:  source,
		TargetAccount:  target,
		AmountInCents:  amount,
		IdempotencyKey: idempotencyKey,
		Description:    description,
		CreatedAt:      time.Now(),
	}
}
