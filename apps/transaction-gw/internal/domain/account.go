package domain

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidAccountID = errors.New("invalid account id")
	ErrInvalidBalance   = errors.New("balance cannot be negative")
)

type Account struct {
	ID             uuid.UUID
	BalanceInCents int64
	Version        int
}

func NewAccount(id uuid.UUID, balance int64) *Account {
	return &Account{
		ID:             id,
		BalanceInCents: balance,
		Version:        0,
	}
}

func (a *Account) Validate() error {
	if a.ID == uuid.Nil {
		return ErrInvalidAccountID
	}
	if a.BalanceInCents < 0 {
		return ErrInvalidBalance
	}
	return nil
}
