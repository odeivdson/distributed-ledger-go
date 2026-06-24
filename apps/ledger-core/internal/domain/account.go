package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountStatus string

const (
	StatusActive   AccountStatus = "ACTIVE"
	StatusInactive AccountStatus = "INACTIVE"
)

type Account struct {
	ID             uuid.UUID
	Status         AccountStatus
	BalanceInCents int64
	Version        int64
	UpdatedAt      time.Time
}

func (a *Account) IsActive() bool {
	return a.Status == StatusActive
}

func (a *Account) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.BalanceInCents += amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.BalanceInCents < amount {
		return ErrInsufficientBalance
	}
	a.BalanceInCents -= amount
	a.UpdatedAt = time.Now()
	return nil
}
