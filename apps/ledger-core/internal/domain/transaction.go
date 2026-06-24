package domain

import (
	"time"

	"github.com/google/uuid"
)

type EntryType string

const (
	Credit EntryType = "CREDIT"
	Debit  EntryType = "DEBIT"
)

type Transaction struct {
	ID             uuid.UUID
	IdempotencyKey string
	Description    string
	CreatedAt      time.Time
	Entries        []LedgerEntry
}

type LedgerEntry struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	Type          EntryType
	AmountInCents int64
	CreatedAt     time.Time
}

func NewTransaction(idempotencyKey, description string) *Transaction {
	return &Transaction{
		ID:             uuid.New(),
		IdempotencyKey: idempotencyKey,
		Description:    description,
		CreatedAt:      time.Now(),
		Entries:        []LedgerEntry{},
	}
}

func (t *Transaction) AddEntry(accountID uuid.UUID, entryType EntryType, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	t.Entries = append(t.Entries, LedgerEntry{
		ID:            uuid.New(),
		TransactionID: t.ID,
		AccountID:     accountID,
		Type:          entryType,
		AmountInCents: amount,
		CreatedAt:     time.Now(),
	})
	return nil
}
