package domain

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrConcurrentUpdate    = errors.New("concurrent update detected")
	ErrAccountNotFound     = errors.New("account not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrAccountInactive     = errors.New("account is not active")
)
