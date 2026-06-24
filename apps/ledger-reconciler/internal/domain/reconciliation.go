package domain

import (
	"github.com/google/uuid"
	"time"
)

type ReconciliationResult struct {
	AccountID      uuid.UUID
	CurrentBalance int64
	ComputedBalance int64
	Diff           int64
	IsValid        bool
	Timestamp      time.Time
}

type AccountCursor struct {
	ID             uuid.UUID
	BalanceInCents int64
}
